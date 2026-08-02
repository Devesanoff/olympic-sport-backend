package domain

import (
	"context"
	"time"
)

type AccessLogFilter struct {
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	ZoneID     *int       `json:"zone_id"`
	CategoryID *int       `json:"category_id"`
	Status     *string    `json:"status"` // ALLOWED / DENIED
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type MealLogFilter struct {
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	CategoryID *int       `json:"category_id"`
	Status     *string    `json:"status"` // ALLOWED / DENIED
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type AccessReportItem struct {
	ID              string       `json:"id"`
	ParticipantID   string       `json:"participant_id"`
	ParticipantName string       `json:"participant_name"`
	CategoryName    string       `json:"category_name"`
	ZoneID          int          `json:"zone_id"`
	ZoneName        string       `json:"zone_name"`
	ZoneCode        string       `json:"zone_code"`
	Direction       Direction    `json:"direction"`
	Status          AccessStatus `json:"status"`
	Reason          string       `json:"reason,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
}

type MealReportItem struct {
	ID              string     `json:"id"`
	ParticipantID   string     `json:"participant_id"`
	ParticipantName string     `json:"participant_name"`
	CategoryName    string     `json:"category_name"`
	MealScheduleID  *int       `json:"meal_schedule_id,omitempty"`
	MealType        MealType   `json:"meal_type"`
	Date            string     `json:"date"`
	Status          MealStatus `json:"status"`
	Reason          string     `json:"reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type DeniedAttemptGroup struct {
	ParticipantID   string    `json:"participant_id"`
	ParticipantName string    `json:"participant_name"`
	DeniedCount     int       `json:"denied_count"`
	LastAttemptAt   time.Time `json:"last_attempt_at"`
	Reasons         []string  `json:"reasons"`
}

type ReportRepository interface {
	GetAccessLogs(ctx context.Context, filter *AccessLogFilter) ([]*AccessReportItem, int, error)
	GetMealLogs(ctx context.Context, filter *MealLogFilter) ([]*MealReportItem, int, error)
	GetDeniedAttempts(ctx context.Context) ([]*DeniedAttemptGroup, error)
}

type ReportService interface {
	GetAccessLogs(ctx context.Context, filter *AccessLogFilter) ([]*AccessReportItem, int, error)
	GetMealLogs(ctx context.Context, filter *MealLogFilter) ([]*MealReportItem, int, error)
	GetDeniedAttempts(ctx context.Context) ([]*DeniedAttemptGroup, error)
	ExportExcel(ctx context.Context, reportType string, filter interface{}) ([]byte, error)
}
