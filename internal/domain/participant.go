package domain

import (
	"context"
	"time"
)

// ParticipantStatus defines the operational lifecycle states of a Participant.
type ParticipantStatus string

const (
	ParticipantStatusActive    ParticipantStatus = "ACTIVE"
	ParticipantStatusInactive  ParticipantStatus = "INACTIVE"
	ParticipantStatusSuspended ParticipantStatus = "SUSPENDED"
)

// Participant maps to the participants database entity.
type Participant struct {
	ID         string            `json:"id"`
	FullName   string            `json:"full_name"`
	CategoryID int               `json:"category_id"`
	QRToken    string            `json:"qr_token"`
	Status     ParticipantStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// ParticipantRepository handles database persistence operations for Participants.
type ParticipantRepository interface {
	Create(ctx context.Context, p *Participant) error
	GetByID(ctx context.Context, id string) (*Participant, error)
	List(ctx context.Context, limit, offset int) ([]*Participant, int, error)
	GetByIDs(ctx context.Context, ids []string) ([]*Participant, error)
	GetByCategoryID(ctx context.Context, categoryID int) ([]*Participant, error)
}

// ParticipantService holds use cases / business logic operations for Participants.
type ParticipantService interface {
	Create(ctx context.Context, fullName string, categoryID int) (*Participant, error)
	GetByID(ctx context.Context, id string) (*Participant, error)
	List(ctx context.Context, limit, offset int) ([]*Participant, int, error)
}
