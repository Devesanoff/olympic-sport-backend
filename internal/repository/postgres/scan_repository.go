package postgres

import (
	"context"
	"fmt"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScanRepository struct {
	db *pgxpool.Pool
}

// NewScanRepository creates a new ScanRepository.
func NewScanRepository(db *pgxpool.Pool) *ScanRepository {
	return &ScanRepository{
		db: db,
	}
}

// LogAccess inserts an entry into access_logs.
func (r *ScanRepository) LogAccess(ctx context.Context, log *domain.AccessLog) error {
	query := `
		INSERT INTO access_logs (participant_id, zone_id, direction, status, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.Exec(ctx, query,
		log.ParticipantID,
		log.ZoneID,
		log.Direction,
		log.Status,
		log.Reason,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert access log: %w", err)
	}
	return nil
}

// LogMeal inserts an entry into meal_logs.
func (r *ScanRepository) LogMeal(ctx context.Context, log *domain.MealLog) error {
	query := `
		INSERT INTO meal_logs (participant_id, meal_schedule_id, meal_type, date, status, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err := r.db.Exec(ctx, query,
		log.ParticipantID,
		log.MealScheduleID,
		log.MealType,
		log.Date,
		log.Status,
		log.Reason,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert meal log: %w", err)
	}
	return nil
}

// GetParticipantWithCategory retrieves participant details alongside their category information.
func (r *ScanRepository) GetParticipantWithCategory(ctx context.Context, participantID string) (*domain.Participant, *domain.Category, error) {
	query := `
		SELECT p.id, p.full_name, p.category_id, p.qr_token, p.status, p.created_at, p.updated_at,
		       c.id, c.name, c.color_code, c.can_eat, c.created_at
		FROM participants p
		JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1;
	`
	var p domain.Participant
	var c domain.Category

	err := r.db.QueryRow(ctx, query, participantID).Scan(
		&p.ID,
		&p.FullName,
		&p.CategoryID,
		&p.QRToken,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
		&c.ID,
		&c.Name,
		&c.ColorCode,
		&c.CanEat,
		&c.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query participant with category: %w", err)
	}

	return &p, &c, nil
}

// IsZoneAllowedForCategory checks if a category has access permissions to a given zone.
func (r *ScanRepository) IsZoneAllowedForCategory(ctx context.Context, categoryID, zoneID int) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM category_allowed_zones
			WHERE category_id = $1 AND zone_id = $2
		);
	`
	var allowed bool
	err := r.db.QueryRow(ctx, query, categoryID, zoneID).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("failed to check zone permission: %w", err)
	}
	return allowed, nil
}

// GetActiveMealSchedule checks for an active meal schedule at the specified date and current time.
func (r *ScanRepository) GetActiveMealSchedule(ctx context.Context, date string, currentTime string) (*domain.MealSchedule, bool, error) {
	query := `
		SELECT id, date::text, meal_type, start_time::text, end_time::text
		FROM meal_schedules
		WHERE date = $1::date
		  AND start_time <= $2::time
		  AND end_time >= $2::time
		LIMIT 1;
	`
	var s domain.MealSchedule
	err := r.db.QueryRow(ctx, query, date, currentTime).Scan(
		&s.ID,
		&s.Date,
		&s.MealType,
		&s.StartTime,
		&s.EndTime,
	)
	if err != nil {
		return nil, false, nil // No active schedule found
	}
	return &s, true, nil
}

// IsCategoryAllowedForMealSchedule checks if a participant category is allowed for a meal schedule.
func (r *ScanRepository) IsCategoryAllowedForMealSchedule(ctx context.Context, mealScheduleID, categoryID int) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM meal_schedule_categories
			WHERE meal_schedule_id = $1 AND category_id = $2
		);
	`
	var allowed bool
	err := r.db.QueryRow(ctx, query, mealScheduleID, categoryID).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("failed to check meal category permission: %w", err)
	}
	return allowed, nil
}
