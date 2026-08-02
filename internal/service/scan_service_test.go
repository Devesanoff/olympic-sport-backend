package service

import (
	"context"
	"testing"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/config"
	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/pkg/database"
	"github.com/Devesanoff/olympic-sport-backend/pkg/hmac"
)

func TestScanService_MealAndAccess(t *testing.T) {
	ctx := context.Background()
	secret := "scan_service_test_secret"
	hmacHelper := hmac.NewHelper(secret)

	// Connect to local Redis
	rdb, err := database.NewRedisClient(ctx, &config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       1, // Use DB 1 for testing
		PoolSize: 10,
	})
	if err != nil {
		t.Fatalf("Failed to connect to Redis for test: %v", err)
	}
	rdb.FlushDB(ctx) // Clear any leftover state before starting
	defer rdb.FlushDB(ctx) // Clean up test keys after test
	defer rdb.Close()

	svc := NewScanService(nil, nil, rdb, hmacHelper, nil)
	defer svc.Close()

	participantID := "123e4567-e89b-12d3-a456-426614174000"
	validToken, err := hmacHelper.GenerateQRToken(participantID, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 1. Test Meal Scan (First attempt -> ALLOWED)
	mealReq := &domain.MealScanRequest{
		QRToken: validToken,
	}
	mealRes, err := svc.ScanMeal(ctx, mealReq)
	if err != nil {
		t.Fatalf("ScanMeal returned unexpected error: %v", err)
	}
	if mealRes.Status != domain.MealStatusAllowed {
		t.Fatalf("Expected status ALLOWED, got %v (reason: %s)", mealRes.Status, mealRes.Reason)
	}

	// 2. Test Meal Scan Duplicate (Second attempt -> DENIED due to Redis SETNX)
	mealResDuplicate, err := svc.ScanMeal(ctx, mealReq)
	if err != nil {
		t.Fatalf("ScanMeal duplicate returned unexpected error: %v", err)
	}
	if mealResDuplicate.Status != domain.MealStatusDenied {
		t.Fatalf("Expected status DENIED for duplicate meal scan, got %v", mealResDuplicate.Status)
	}
	if mealResDuplicate.Reason != "meal has already been consumed today for this schedule" {
		t.Errorf("Unexpected reason for duplicate meal scan: %s", mealResDuplicate.Reason)
	}

	// 3. Test Access Scan IN (Headcount should increment to 1)
	accessInReq := &domain.AccessScanRequest{
		QRToken:   validToken,
		ZoneID:    101,
		Direction: domain.DirectionIn,
	}
	accessInRes, err := svc.ScanAccess(ctx, accessInReq)
	if err != nil {
		t.Fatalf("ScanAccess IN returned unexpected error: %v", err)
	}
	if accessInRes.Status != domain.AccessStatusAllowed {
		t.Fatalf("Expected status ALLOWED, got %v (reason: %s)", accessInRes.Status, accessInRes.Reason)
	}
	if accessInRes.OccupancyCount != 1 {
		t.Errorf("Expected occupancy 1 after IN, got %d", accessInRes.OccupancyCount)
	}

	// 4. Test Access Scan OUT (Headcount should decrement back to 0)
	accessOutReq := &domain.AccessScanRequest{
		QRToken:   validToken,
		ZoneID:    101,
		Direction: domain.DirectionOut,
	}
	accessOutRes, err := svc.ScanAccess(ctx, accessOutReq)
	if err != nil {
		t.Fatalf("ScanAccess OUT returned unexpected error: %v", err)
	}
	if accessOutRes.Status != domain.AccessStatusAllowed {
		t.Fatalf("Expected status ALLOWED, got %v", accessOutRes.Status)
	}
	if accessOutRes.OccupancyCount != 0 {
		t.Errorf("Expected occupancy 0 after OUT, got %d", accessOutRes.OccupancyCount)
	}

	// 5. Test Invalid Token (Should return DENIED)
	badTokenReq := &domain.MealScanRequest{
		QRToken: "invalid.token.signature",
	}
	badTokenRes, err := svc.ScanMeal(ctx, badTokenReq)
	if err != nil {
		t.Fatalf("ScanMeal bad token returned error: %v", err)
	}
	if badTokenRes.Status != domain.MealStatusDenied {
		t.Fatalf("Expected status DENIED for invalid token, got %v", badTokenRes.Status)
	}
}
