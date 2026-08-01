package domain

import (
	"context"
	"time"
)

// Enum Definitions
type Direction string

const (
	DirectionIn  Direction = "IN"
	DirectionOut Direction = "OUT"
)

type AccessStatus string

const (
	AccessStatusAllowed AccessStatus = "ALLOWED"
	AccessStatusDenied  AccessStatus = "DENIED"
)

type MealType string

const (
	MealTypeBreakfast  MealType = "BREAKFAST"
	MealTypeLunch      MealType = "LUNCH"
	MealTypeDinner     MealType = "DINNER"
	MealTypeNightSnack MealType = "NIGHT_SNACK"
)

type MealStatus string

const (
	MealStatusAllowed MealStatus = "ALLOWED"
	MealStatusDenied  MealStatus = "DENIED"
)

// Domain Models
type AccessLog struct {
	ID            string       `json:"id"`
	ParticipantID string       `json:"participant_id"`
	ZoneID        int          `json:"zone_id"`
	Direction     Direction    `json:"direction"`
	Status        AccessStatus `json:"status"`
	Reason        string       `json:"reason,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
}

type MealLog struct {
	ID             string     `json:"id"`
	ParticipantID  string     `json:"participant_id"`
	MealScheduleID *int       `json:"meal_schedule_id,omitempty"`
	MealType       MealType   `json:"meal_type"`
	Date           string     `json:"date"`
	Status         MealStatus `json:"status"`
	Reason         string     `json:"reason,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type MealSchedule struct {
	ID        int      `json:"id"`
	Date      string   `json:"date"`
	MealType  MealType `json:"meal_type"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
}

type Zone struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	RequiresInOut  bool      `json:"requires_in_out"`
	CreatedAt      time.Time `json:"created_at"`
}

type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ColorCode string    `json:"color_code"`
	CanEat    bool      `json:"can_eat"`
	CreatedAt time.Time `json:"created_at"`
}

// Request and Response DTOs
type MealScanRequest struct {
	QRToken string `json:"qr_token" binding:"required"`
}

type MealScanResponse struct {
	Status          MealStatus `json:"status"`
	ParticipantID   string     `json:"participant_id,omitempty"`
	ParticipantName string     `json:"participant_name,omitempty"`
	MealType        MealType   `json:"meal_type,omitempty"`
	Reason          string     `json:"reason,omitempty"`
}

type AccessScanRequest struct {
	QRToken   string    `json:"qr_token" binding:"required"`
	ZoneID    int       `json:"zone_id" binding:"required"`
	Direction Direction `json:"direction" binding:"required"`
}

type AccessScanResponse struct {
	Status          AccessStatus `json:"status"`
	ParticipantID   string       `json:"participant_id,omitempty"`
	ParticipantName string       `json:"participant_name,omitempty"`
	ZoneID          int          `json:"zone_id,omitempty"`
	Direction       Direction    `json:"direction,omitempty"`
	OccupancyCount  int64        `json:"occupancy_count"`
	Reason          string       `json:"reason,omitempty"`
}

// Repository Interfaces
type ScanLogRepository interface {
	LogAccess(ctx context.Context, log *AccessLog) error
	LogMeal(ctx context.Context, log *MealLog) error
}

type ScanLookupRepository interface {
	GetParticipantWithCategory(ctx context.Context, participantID string) (*Participant, *Category, error)
	IsZoneAllowedForCategory(ctx context.Context, categoryID, zoneID int) (bool, error)
	GetActiveMealSchedule(ctx context.Context, date string, currentTime string) (*MealSchedule, bool, error)
	IsCategoryAllowedForMealSchedule(ctx context.Context, mealScheduleID, categoryID int) (bool, error)
}

// Service Interface
type ScanService interface {
	ScanAccess(ctx context.Context, req *AccessScanRequest) (*AccessScanResponse, error)
	ScanMeal(ctx context.Context, req *MealScanRequest) (*MealScanResponse, error)
	Close()
}
