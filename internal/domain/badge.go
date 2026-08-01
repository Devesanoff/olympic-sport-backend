package domain

import (
	"context"
)

// BulkBadgeRequest represents the JSON payload for bulk badge generation.
type BulkBadgeRequest struct {
	ParticipantIDs []string `json:"participant_ids,omitempty"`
	CategoryID     *int     `json:"category_id,omitempty"`
}

// BadgeService defines the use cases for generating PDF badges.
type BadgeService interface {
	GenerateSingleBadge(ctx context.Context, participantID string) ([]byte, error)
	GenerateBulkBadges(ctx context.Context, req *BulkBadgeRequest) ([]byte, error)
}
