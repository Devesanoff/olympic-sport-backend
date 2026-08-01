package service

import (
	"context"
	"fmt"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/pkg/pdf"
)

type badgeService struct {
	participantRepo domain.ParticipantRepository
	lookupRepo      domain.ScanLookupRepository
}

// NewBadgeService creates a new instance of badgeService.
func NewBadgeService(participantRepo domain.ParticipantRepository, lookupRepo domain.ScanLookupRepository) domain.BadgeService {
	return &badgeService{
		participantRepo: participantRepo,
		lookupRepo:      lookupRepo,
	}
}

func (s *badgeService) GenerateSingleBadge(ctx context.Context, participantID string) ([]byte, error) {
	p, category, err := s.lookupRepo.GetParticipantWithCategory(ctx, participantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch participant details: %w", err)
	}

	builder := pdf.NewBadgeBuilder()
	if err := builder.AddBadgePage(p.FullName, category.Name, category.ColorCode, p.QRToken); err != nil {
		return nil, fmt.Errorf("failed to generate badge page: %w", err)
	}

	return builder.Output()
}

func (s *badgeService) GenerateBulkBadges(ctx context.Context, req *domain.BulkBadgeRequest) ([]byte, error) {
	var participants []*domain.Participant
	var err error

	if len(req.ParticipantIDs) > 0 {
		participants, err = s.participantRepo.GetByIDs(ctx, req.ParticipantIDs)
	} else if req.CategoryID != nil {
		participants, err = s.participantRepo.GetByCategoryID(ctx, *req.CategoryID)
	} else {
		return nil, fmt.Errorf("either participant_ids or category_id must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch participants: %w", err)
	}

	builder := pdf.NewBadgeBuilder()

	for _, p := range participants {
		// To get category color/name, we can either hit the DB again or ideally we have it joined.
		// For simplicity, we use lookupRepo to get the category. Since this is bulk, this is N+1.
		// However, in a real scenario we'd do a joined query. Let's use lookupRepo for now as it's straightforward.
		_, category, err := s.lookupRepo.GetParticipantWithCategory(ctx, p.ID)
		if err != nil {
			continue // skip if we can't fetch category
		}

		if err := builder.AddBadgePage(p.FullName, category.Name, category.ColorCode, p.QRToken); err != nil {
			return nil, fmt.Errorf("failed to generate badge for %s: %w", p.ID, err)
		}
	}

	return builder.Output()
}
