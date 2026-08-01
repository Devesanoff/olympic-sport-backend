package service

import (
	"context"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/pkg/hmac"
	"github.com/google/uuid"
)

type ParticipantService struct {
	repo       domain.ParticipantRepository
	hmacHelper *hmac.Helper
}

// NewParticipantService creates a new ParticipantService instance.
func NewParticipantService(repo domain.ParticipantRepository, hmacHelper *hmac.Helper) *ParticipantService {
	return &ParticipantService{
		repo:       repo,
		hmacHelper: hmacHelper,
	}
}

// Create generates a secure UUID, signs it to generate a QR token, and inserts the Participant.
func (s *ParticipantService) Create(ctx context.Context, fullName string, categoryID int) (*domain.Participant, error) {
	participantID := uuid.New().String()
	timestamp := time.Now().Unix()

	qrToken, err := s.hmacHelper.GenerateQRToken(participantID, timestamp)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	p := &domain.Participant{
		ID:         participantID,
		FullName:   fullName,
		CategoryID: categoryID,
		QRToken:    qrToken,
		Status:     domain.ParticipantStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetByID returns a Participant by their UUID.
func (s *ParticipantService) GetByID(ctx context.Context, id string) (*domain.Participant, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a paginated list of Participants.
func (s *ParticipantService) List(ctx context.Context, limit, offset int) ([]*domain.Participant, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}
