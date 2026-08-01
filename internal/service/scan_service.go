package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/pkg/hmac"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type logJob struct {
	accessLog *domain.AccessLog
	mealLog   *domain.MealLog
}

type ScanService struct {
	logRepo     domain.ScanLogRepository
	lookupRepo  domain.ScanLookupRepository
	rdb         *redis.Client
	hmacHelper  *hmac.Helper
	wsHub       domain.WSHub
	logChan     chan logJob
	workerWg    sync.WaitGroup
	stopWorker  chan struct{}
}

// NewScanService creates a new high-concurrency ScanService with background logging workers.
func NewScanService(
	logRepo domain.ScanLogRepository,
	lookupRepo domain.ScanLookupRepository,
	rdb *redis.Client,
	hmacHelper *hmac.Helper,
	wsHub domain.WSHub,
) *ScanService {
	s := &ScanService{
		logRepo:    logRepo,
		lookupRepo: lookupRepo,
		rdb:        rdb,
		hmacHelper: hmacHelper,
		wsHub:      wsHub,
		logChan:    make(chan logJob, 10000), // Buffer for high concurrency burst logging
		stopWorker: make(chan struct{}),
	}

	// Start 5 background worker goroutines for async DB logging
	s.workerWg.Add(5)
	for i := 0; i < 5; i++ {
		go s.logWorker(i)
	}

	return s
}

// logWorker processes asynchronous log entries to PostgreSQL without blocking scanner HTTP responses.
func (s *ScanService) logWorker(workerID int) {
	defer s.workerWg.Done()
	for job := range s.logChan {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if job.accessLog != nil && s.logRepo != nil {
			if err := s.logRepo.LogAccess(ctx, job.accessLog); err != nil {
				log.Error().Err(err).Int("worker", workerID).Msg("Async access log insertion failed")
			}
		}
		if job.mealLog != nil && s.logRepo != nil {
			if err := s.logRepo.LogMeal(ctx, job.mealLog); err != nil {
				log.Error().Err(err).Int("worker", workerID).Msg("Async meal log insertion failed")
			}
		}
		cancel()
	}
}

// Close gracefully flushes the async logging pipeline during application shutdown.
func (s *ScanService) Close() {
	close(s.logChan)
	s.workerWg.Wait()
}

// ScanMeal processes meal accreditation checks with strict Redis SETNX deduplication.
func (s *ScanService) ScanMeal(ctx context.Context, req *domain.MealScanRequest) (*domain.MealScanResponse, error) {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	timeStr := now.Format("15:04:05")

	// 1. Offline-Ready HMAC QR Token Signature Validation
	participantID, err := s.hmacHelper.ValidateQRToken(req.QRToken)
	if err != nil {
		s.asyncLogMeal(&domain.MealLog{
			ParticipantID: "00000000-0000-0000-0000-000000000000",
			MealType:      domain.MealTypeLunch,
			Date:          dateStr,
			Status:        domain.MealStatusDenied,
			Reason:        "invalid QR signature",
			CreatedAt:     now,
		})
		return &domain.MealScanResponse{
			Status: domain.MealStatusDenied,
			Reason: "invalid or tampered QR token",
		}, nil
	}

	// 2. Fetch Participant & Category Information
	var categoryID int
	var categoryCanEat bool
	var participantName string

	if s.lookupRepo != nil {
		participant, category, err := s.lookupRepo.GetParticipantWithCategory(ctx, participantID)
		if err != nil {
			s.asyncLogMeal(&domain.MealLog{
				ParticipantID: participantID,
				MealType:      domain.MealTypeLunch,
				Date:          dateStr,
				Status:        domain.MealStatusDenied,
				Reason:        "participant not found",
				CreatedAt:     now,
			})
			return &domain.MealScanResponse{
				Status:        domain.MealStatusDenied,
				ParticipantID: participantID,
				Reason:        "participant record not found",
			}, nil
		}

		if participant.Status != domain.ParticipantStatusActive {
			s.asyncLogMeal(&domain.MealLog{
				ParticipantID: participantID,
				MealType:      domain.MealTypeLunch,
				Date:          dateStr,
				Status:        domain.MealStatusDenied,
				Reason:        "participant status inactive",
				CreatedAt:     now,
			})
			return &domain.MealScanResponse{
				Status:          domain.MealStatusDenied,
				ParticipantID:   participantID,
				ParticipantName: participant.FullName,
				Reason:          fmt.Sprintf("participant status is %s", participant.Status),
			}, nil
		}

		categoryID = category.ID
		categoryCanEat = category.CanEat
		participantName = participant.FullName
	} else {
		// Mock fallback mode when DB is uninitialized for tests
		categoryID = 1
		categoryCanEat = true
		participantName = "Mock Participant"
	}

	if !categoryCanEat {
		s.asyncLogMeal(&domain.MealLog{
			ParticipantID: participantID,
			MealType:      domain.MealTypeLunch,
			Date:          dateStr,
			Status:        domain.MealStatusDenied,
			Reason:        "category not entitled to meals",
			CreatedAt:     now,
		})
		return &domain.MealScanResponse{
			Status:          domain.MealStatusDenied,
			ParticipantID:   participantID,
			ParticipantName: participantName,
			Reason:          "participant category is not allowed meal access",
		}, nil
	}

	// 3. Active Meal Schedule Validation
	var mealType domain.MealType = domain.MealTypeLunch
	var mealScheduleID *int

	if s.lookupRepo != nil {
		schedule, active, err := s.lookupRepo.GetActiveMealSchedule(ctx, dateStr, timeStr)
		if err != nil || !active {
			s.asyncLogMeal(&domain.MealLog{
				ParticipantID: participantID,
				MealType:      mealType,
				Date:          dateStr,
				Status:        domain.MealStatusDenied,
				Reason:        "no active meal schedule",
				CreatedAt:     now,
			})
			return &domain.MealScanResponse{
				Status:          domain.MealStatusDenied,
				ParticipantID:   participantID,
				ParticipantName: participantName,
				Reason:          "no active meal schedule found for current time",
			}, nil
		}

		mealType = schedule.MealType
		mealScheduleID = &schedule.ID

		allowed, err := s.lookupRepo.IsCategoryAllowedForMealSchedule(ctx, schedule.ID, categoryID)
		if err != nil || !allowed {
			s.asyncLogMeal(&domain.MealLog{
				ParticipantID:  participantID,
				MealScheduleID: mealScheduleID,
				MealType:       mealType,
				Date:           dateStr,
				Status:         domain.MealStatusDenied,
				Reason:         "category not allowed for current meal schedule",
				CreatedAt:      now,
			})
			return &domain.MealScanResponse{
				Status:          domain.MealStatusDenied,
				ParticipantID:   participantID,
				ParticipantName: participantName,
				MealType:        mealType,
				Reason:          "participant category is not permitted for current meal schedule",
			}, nil
		}
	}

	// 4. Redis SETNX Concurrency Guard (Strictly One Meal per Meal Type per Day)
	redisKey := fmt.Sprintf("meal:%s:%s:%s", dateStr, mealType, participantID)
	if s.rdb != nil {
		acquired, err := s.rdb.SetNX(ctx, redisKey, "CLAIMED", 24*time.Hour).Result()
		if err != nil {
			log.Error().Err(err).Str("key", redisKey).Msg("Redis SETNX meal check error")
		} else if !acquired {
			// Double consumption attempt! Key already exists!
			s.asyncLogMeal(&domain.MealLog{
				ParticipantID:  participantID,
				MealScheduleID: mealScheduleID,
				MealType:       mealType,
				Date:           dateStr,
				Status:         domain.MealStatusDenied,
				Reason:         "meal already claimed today",
				CreatedAt:      now,
			})
			return &domain.MealScanResponse{
				Status:          domain.MealStatusDenied,
				ParticipantID:   participantID,
				ParticipantName: participantName,
				MealType:        mealType,
				Reason:          "meal has already been consumed today for this schedule",
			}, nil
		}
	}

	// 5. Successful Meal Scan
	s.asyncLogMeal(&domain.MealLog{
		ParticipantID:  participantID,
		MealScheduleID: mealScheduleID,
		MealType:       mealType,
		Date:           dateStr,
		Status:         domain.MealStatusAllowed,
		CreatedAt:      now,
	})

	return &domain.MealScanResponse{
		Status:          domain.MealStatusAllowed,
		ParticipantID:   participantID,
		ParticipantName: participantName,
		MealType:        mealType,
	}, nil
}

// ScanAccess processes zone access scans and updates zone occupancy count in Redis.
func (s *ScanService) ScanAccess(ctx context.Context, req *domain.AccessScanRequest) (*domain.AccessScanResponse, error) {
	now := time.Now()

	// 1. Offline-Ready HMAC Signature Validation
	participantID, err := s.hmacHelper.ValidateQRToken(req.QRToken)
	if err != nil {
		s.asyncLogAccess(&domain.AccessLog{
			ParticipantID: "00000000-0000-0000-0000-000000000000",
			ZoneID:        req.ZoneID,
			Direction:     req.Direction,
			Status:        domain.AccessStatusDenied,
			Reason:        "invalid QR signature",
			CreatedAt:     now,
		})
		return &domain.AccessScanResponse{
			Status:    domain.AccessStatusDenied,
			ZoneID:    req.ZoneID,
			Direction: req.Direction,
			Reason:    "invalid or tampered QR token",
		}, nil
	}

	// 2. Fetch Participant & Check Zone Permissions
	var categoryID int
	var participantName string

	if s.lookupRepo != nil {
		participant, category, err := s.lookupRepo.GetParticipantWithCategory(ctx, participantID)
		if err != nil {
			s.asyncLogAccess(&domain.AccessLog{
				ParticipantID: participantID,
				ZoneID:        req.ZoneID,
				Direction:     req.Direction,
				Status:        domain.AccessStatusDenied,
				Reason:        "participant not found",
				CreatedAt:     now,
			})
			return &domain.AccessScanResponse{
				Status:        domain.AccessStatusDenied,
				ParticipantID: participantID,
				ZoneID:        req.ZoneID,
				Direction:     req.Direction,
				Reason:        "participant record not found",
			}, nil
		}

		if participant.Status != domain.ParticipantStatusActive {
			s.asyncLogAccess(&domain.AccessLog{
				ParticipantID: participantID,
				ZoneID:        req.ZoneID,
				Direction:     req.Direction,
				Status:        domain.AccessStatusDenied,
				Reason:        "participant status inactive",
				CreatedAt:     now,
			})
			return &domain.AccessScanResponse{
				Status:          domain.AccessStatusDenied,
				ParticipantID:   participantID,
				ParticipantName: participant.FullName,
				ZoneID:          req.ZoneID,
				Direction:       req.Direction,
				Reason:          fmt.Sprintf("participant status is %s", participant.Status),
			}, nil
		}

		categoryID = category.ID
		participantName = participant.FullName

		// Check Category Allowed Zones
		allowed, err := s.lookupRepo.IsZoneAllowedForCategory(ctx, categoryID, req.ZoneID)
		if err != nil || !allowed {
			s.asyncLogAccess(&domain.AccessLog{
				ParticipantID: participantID,
				ZoneID:        req.ZoneID,
				Direction:     req.Direction,
				Status:        domain.AccessStatusDenied,
				Reason:        "zone not allowed for category",
				CreatedAt:     now,
			})
			return &domain.AccessScanResponse{
				Status:          domain.AccessStatusDenied,
				ParticipantID:   participantID,
				ParticipantName: participantName,
				ZoneID:          req.ZoneID,
				Direction:       req.Direction,
				Reason:          "participant category is not authorized for this zone",
			}, nil
		}
	} else {
		// Mock fallback mode when DB is uninitialized for tests
		participantName = "Mock Participant"
	}

	// 3. Update Real-Time Zone Headcount Occupancy in Redis
	var occupancyCount int64 = 0
	zoneCountKey := fmt.Sprintf("zone_count:%d", req.ZoneID)

	if s.rdb != nil {
		if req.Direction == domain.DirectionIn {
			count, err := s.rdb.Incr(ctx, zoneCountKey).Result()
			if err == nil {
				occupancyCount = count
			}
		} else if req.Direction == domain.DirectionOut {
			// Lua script to safely decrement without dropping below zero
			script := `
				local current = redis.call('GET', KEYS[1])
				if not current or tonumber(current) <= 0 then
					redis.call('SET', KEYS[1], 0)
					return 0
				else
					return redis.call('DECR', KEYS[1])
				end
			`
			res, err := s.rdb.Eval(ctx, script, []string{zoneCountKey}).Result()
			if err == nil {
				if count, ok := res.(int64); ok {
					occupancyCount = count
				}
			}
		}
	}

	// Trigger WebSocket Broadcast asynchronously
	if s.wsHub != nil {
		// Use a goroutine to not block the fast response path
		go s.wsHub.BroadcastZoneUpdate(domain.ZoneOccupancyUpdate{
			ZoneID:         req.ZoneID,
			OccupancyCount: occupancyCount,
		})
	}

	// 4. Log Access Attempt Asynchronously
	s.asyncLogAccess(&domain.AccessLog{
		ParticipantID: participantID,
		ZoneID:        req.ZoneID,
		Direction:     req.Direction,
		Status:        domain.AccessStatusAllowed,
		CreatedAt:     now,
	})

	return &domain.AccessScanResponse{
		Status:          domain.AccessStatusAllowed,
		ParticipantID:   participantID,
		ParticipantName: participantName,
		ZoneID:          req.ZoneID,
		Direction:       req.Direction,
		OccupancyCount:  occupancyCount,
	}, nil
}

// Helper methods for non-blocking async logging queue
func (s *ScanService) asyncLogAccess(log *domain.AccessLog) {
	select {
	case s.logChan <- logJob{accessLog: log}:
	default:
		// Queue full under extreme burst; drop log to prevent scanner UI stall
	}
}

func (s *ScanService) asyncLogMeal(log *domain.MealLog) {
	select {
	case s.logChan <- logJob{mealLog: log}:
	default:
		// Queue full under extreme burst; drop log to prevent scanner UI stall
	}
}
