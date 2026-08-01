package domain

import (
	"context"
)

// SyncParticipant is a lightweight representation of a Participant for offline use.
type SyncParticipant struct {
	ID         string `json:"id"`
	CategoryID int    `json:"category_id"`
	Status     string `json:"status"` // e.g. ACTIVE
	QRToken    string `json:"qr_token"`
}

// CategoryAllowedZone represents the many-to-many relationship mapping.
type CategoryAllowedZone struct {
	CategoryID int `json:"category_id"`
	ZoneID     int `json:"zone_id"`
}

// OfflinePackage represents the complete payload sent to the mobile client for offline sync.
type OfflinePackage struct {
	Categories           []*Category            `json:"categories"`
	Zones                []*Zone                `json:"zones"`
	CategoryAllowedZones []*CategoryAllowedZone `json:"category_allowed_zones"`
	MealSchedules        []*MealSchedule        `json:"meal_schedules"`
	Participants         []*SyncParticipant     `json:"participants"`
}

// SyncUploadLogsRequest contains batches of logs generated offline by the mobile app.
type SyncUploadLogsRequest struct {
	AccessLogs []AccessLog `json:"access_logs"`
	MealLogs   []MealLog   `json:"meal_logs"`
}

// SyncRepository handles database operations for fetching the offline package and bulk inserting logs.
type SyncRepository interface {
	GetOfflinePackage(ctx context.Context) (*OfflinePackage, error)
	BulkInsertAccessLogs(ctx context.Context, logs []AccessLog) error
	BulkInsertMealLogs(ctx context.Context, logs []MealLog) error
}

// SyncService holds business logic for offline synchronization.
type SyncService interface {
	GetOfflinePackage(ctx context.Context) (*OfflinePackage, error)
	UploadLogs(ctx context.Context, req *SyncUploadLogsRequest) error
}
