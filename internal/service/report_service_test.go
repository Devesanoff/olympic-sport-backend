package service

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/xuri/excelize/v2"
)

type mockReportRepository struct {
	accessLogs     []*domain.AccessReportItem
	mealLogs       []*domain.MealReportItem
	deniedAttempts []*domain.DeniedAttemptGroup
}

func (m *mockReportRepository) GetAccessLogs(ctx context.Context, filter *domain.AccessLogFilter) ([]*domain.AccessReportItem, int, error) {
	var result []*domain.AccessReportItem
	for _, item := range m.accessLogs {
		if filter.Status != nil && string(item.Status) != *filter.Status {
			continue
		}
		if filter.ZoneID != nil && item.ZoneID != *filter.ZoneID {
			continue
		}
		result = append(result, item)
	}
	return result, len(result), nil
}

func (m *mockReportRepository) GetMealLogs(ctx context.Context, filter *domain.MealLogFilter) ([]*domain.MealReportItem, int, error) {
	var result []*domain.MealReportItem
	for _, item := range m.mealLogs {
		if filter.Status != nil && string(item.Status) != *filter.Status {
			continue
		}
		result = append(result, item)
	}
	return result, len(result), nil
}

func (m *mockReportRepository) GetDeniedAttempts(ctx context.Context) ([]*domain.DeniedAttemptGroup, error) {
	return m.deniedAttempts, nil
}

func TestReportService(t *testing.T) {
	t.Run("Test Query and Export Access Logs", func(t *testing.T) {
		repo := &mockReportRepository{
			accessLogs: []*domain.AccessReportItem{
				{
					ID:              "log-1",
					ParticipantID:   "p-1",
					ParticipantName: "John Doe",
					CategoryName:    "Athlete",
					ZoneID:          1,
					ZoneName:        "VIP Lounge",
					ZoneCode:        "VIP",
					Direction:       domain.DirectionIn,
					Status:          domain.AccessStatusAllowed,
					CreatedAt:       time.Now(),
				},
				{
					ID:              "log-2",
					ParticipantID:   "p-2",
					ParticipantName: "Jane Smith",
					CategoryName:    "Press",
					ZoneID:          2,
					ZoneName:        "Press Area",
					ZoneCode:        "PRESS",
					Direction:       domain.DirectionIn,
					Status:          domain.AccessStatusDenied,
					Reason:          "unauthorized category",
					CreatedAt:       time.Now(),
				},
			},
		}

		svc := NewReportService(repo)

		// 1. GetAccessLogs query test
		status := "DENIED"
		logs, total, err := svc.GetAccessLogs(context.Background(), &domain.AccessLogFilter{
			Status: &status,
		})
		if err != nil {
			t.Fatalf("unexpected error getting access logs: %v", err)
		}
		if total != 1 {
			t.Errorf("expected 1 log, got %d", total)
		}
		if logs[0].ParticipantName != "Jane Smith" {
			t.Errorf("expected Jane Smith, got %s", logs[0].ParticipantName)
		}

		// 2. Export Excel test
		excelBytes, err := svc.ExportExcel(context.Background(), "access", &domain.AccessLogFilter{})
		if err != nil {
			t.Fatalf("unexpected error exporting excel: %v", err)
		}
		if len(excelBytes) == 0 {
			t.Error("exported excel bytes is empty")
		}

		// Verify excel sheet structure
		r := bytes.NewReader(excelBytes)
		xlsx, err := excelize.OpenReader(r)
		if err != nil {
			t.Fatalf("failed to parse exported excel: %v", err)
		}
		defer xlsx.Close()

		rows, err := xlsx.GetRows("Report")
		if err != nil {
			t.Fatalf("failed to get Report sheet rows: %v", err)
		}
		if len(rows) != 3 { // Header + 2 data rows
			t.Errorf("expected 3 rows in sheet, got %d", len(rows))
		}
		// Check headers
		if rows[0][0] != "Log ID" || rows[0][2] != "Participant Name" {
			t.Errorf("unexpected headers: %v", rows[0])
		}
		// Check data values
		if rows[1][2] != "John Doe" || rows[2][2] != "Jane Smith" {
			t.Errorf("unexpected names in rows: %s, %s", rows[1][2], rows[2][2])
		}
	})

	t.Run("Test Query and Export Meal Logs", func(t *testing.T) {
		repo := &mockReportRepository{
			mealLogs: []*domain.MealReportItem{
				{
					ID:              "meal-1",
					ParticipantID:   "p-1",
					ParticipantName: "John Doe",
					CategoryName:    "Athlete",
					MealType:        domain.MealTypeLunch,
					Date:            "2026-08-02",
					Status:          domain.MealStatusAllowed,
					CreatedAt:       time.Now(),
				},
			},
		}

		svc := NewReportService(repo)

		logs, total, err := svc.GetMealLogs(context.Background(), &domain.MealLogFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected 1 meal log, got %d", total)
		}
		if logs[0].ParticipantName != "John Doe" {
			t.Errorf("expected John Doe, got %s", logs[0].ParticipantName)
		}

		excelBytes, err := svc.ExportExcel(context.Background(), "meal", &domain.MealLogFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(excelBytes) == 0 {
			t.Error("exported excel bytes is empty")
		}
	})

	t.Run("Test Query and Export Denied Attempts", func(t *testing.T) {
		repo := &mockReportRepository{
			deniedAttempts: []*domain.DeniedAttemptGroup{
				{
					ParticipantID:   "p-2",
					ParticipantName: "Jane Smith",
					DeniedCount:     3,
					LastAttemptAt:   time.Now(),
					Reasons:         []string{"unauthorized zone", "duplicate scan"},
				},
			},
		}

		svc := NewReportService(repo)

		groups, err := svc.GetDeniedAttempts(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(groups) != 1 {
			t.Errorf("expected 1 group, got %d", len(groups))
		}

		excelBytes, err := svc.ExportExcel(context.Background(), "denied", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(excelBytes) == 0 {
			t.Error("exported excel bytes is empty")
		}
	})
}
