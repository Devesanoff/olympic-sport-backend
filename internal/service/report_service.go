package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/xuri/excelize/v2"
)

type ReportService struct {
	repo domain.ReportRepository
}

// NewReportService creates a new ReportService.
func NewReportService(repo domain.ReportRepository) *ReportService {
	return &ReportService{
		repo: repo,
	}
}

// GetAccessLogs retrieves access logs.
func (s *ReportService) GetAccessLogs(ctx context.Context, filter *domain.AccessLogFilter) ([]*domain.AccessReportItem, int, error) {
	return s.repo.GetAccessLogs(ctx, filter)
}

// GetMealLogs retrieves meal logs.
func (s *ReportService) GetMealLogs(ctx context.Context, filter *domain.MealLogFilter) ([]*domain.MealReportItem, int, error) {
	return s.repo.GetMealLogs(ctx, filter)
}

// GetDeniedAttempts retrieves denied attempts list.
func (s *ReportService) GetDeniedAttempts(ctx context.Context) ([]*domain.DeniedAttemptGroup, error) {
	return s.repo.GetDeniedAttempts(ctx)
}

// ExportExcel generates an Excel spreadsheet for the specified report type and filter.
func (s *ReportService) ExportExcel(ctx context.Context, reportType string, filter interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Report"
	f.SetSheetName("Sheet1", sheetName)

	// Styles definitions
	headerStyle, err := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1F4E78"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Family: "Segoe UI", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D3D3D3", Style: 1},
			{Type: "right", Color: "#D3D3D3", Style: 1},
			{Type: "top", Color: "#D3D3D3", Style: 1},
			{Type: "bottom", Color: "#D3D3D3", Style: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	allowedStyle, err := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D4EFDF"}, Pattern: 1},
		Font:      &excelize.Font{Color: "#145A32", Family: "Segoe UI", Size: 10, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create allowed style: %w", err)
	}

	deniedStyle, err := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FADBD8"}, Pattern: 1},
		Font:      &excelize.Font{Color: "#78281F", Family: "Segoe UI", Size: 10, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create denied style: %w", err)
	}

	normalStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Segoe UI", Size: 10},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create normal style: %w", err)
	}

	centerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Segoe UI", Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create center style: %w", err)
	}

	var rowCount int
	var colCount int

	switch strings.ToLower(reportType) {
	case "access":
		flt, ok := filter.(*domain.AccessLogFilter)
		if !ok {
			return nil, fmt.Errorf("invalid filter type for access report")
		}
		// Create copy and disable pagination limit for export
		exportFlt := *flt
		exportFlt.Limit = 0
		exportFlt.Offset = 0

		data, _, err := s.repo.GetAccessLogs(ctx, &exportFlt)
		if err != nil {
			return nil, fmt.Errorf("failed to query access logs for export: %w", err)
		}

		headers := []string{"Log ID", "Participant ID", "Participant Name", "Category", "Zone ID", "Zone Name", "Zone Code", "Direction", "Status", "Reason", "Timestamp"}
		colCount = len(headers)

		// Set headers
		for cIdx, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, 1)
			f.SetCellValue(sheetName, cell, header)
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}

		// Set data rows
		for rIdx, item := range data {
			row := rIdx + 2
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.ID)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ParticipantID)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.ParticipantName)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.CategoryName)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.ZoneID)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.ZoneName)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.ZoneCode)
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), string(item.Direction))
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), string(item.Status))
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), item.Reason)
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), item.CreatedAt.Format("2006-01-02 15:04:05 MST"))

			// Apply cell styles
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), normalStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), centerStyle)

			// Highlight Status
			statusCell := fmt.Sprintf("I%d", row)
			if item.Status == domain.AccessStatusAllowed {
				f.SetCellStyle(sheetName, statusCell, statusCell, allowedStyle)
			} else {
				f.SetCellStyle(sheetName, statusCell, statusCell, deniedStyle)
			}

			f.SetCellStyle(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), normalStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), centerStyle)
		}
		rowCount = len(data) + 1

	case "meal":
		flt, ok := filter.(*domain.MealLogFilter)
		if !ok {
			return nil, fmt.Errorf("invalid filter type for meal report")
		}
		// Create copy and disable pagination limit for export
		exportFlt := *flt
		exportFlt.Limit = 0
		exportFlt.Offset = 0

		data, _, err := s.repo.GetMealLogs(ctx, &exportFlt)
		if err != nil {
			return nil, fmt.Errorf("failed to query meal logs for export: %w", err)
		}

		headers := []string{"Log ID", "Participant ID", "Participant Name", "Category", "Meal Schedule ID", "Meal Type", "Schedule Date", "Status", "Reason", "Timestamp"}
		colCount = len(headers)

		// Set headers
		for cIdx, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, 1)
			f.SetCellValue(sheetName, cell, header)
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}

		// Set data rows
		for rIdx, item := range data {
			row := rIdx + 2
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.ID)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ParticipantID)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.ParticipantName)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.CategoryName)
			if item.MealScheduleID != nil {
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), *item.MealScheduleID)
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "N/A")
			}
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), string(item.MealType))
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.Date)
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), string(item.Status))
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), item.Reason)
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), item.CreatedAt.Format("2006-01-02 15:04:05 MST"))

			// Apply cell styles
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), normalStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("G%d", row), centerStyle)

			// Highlight Status
			statusCell := fmt.Sprintf("H%d", row)
			if item.Status == domain.MealStatusAllowed {
				f.SetCellStyle(sheetName, statusCell, statusCell, allowedStyle)
			} else {
				f.SetCellStyle(sheetName, statusCell, statusCell, deniedStyle)
			}

			f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), normalStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), centerStyle)
		}
		rowCount = len(data) + 1

	case "denied":
		data, err := s.repo.GetDeniedAttempts(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query denied attempts for export: %w", err)
		}

		headers := []string{"Participant ID", "Participant Name", "Denied Attempts Count", "Last Attempt Timestamp", "Attempt Reasons"}
		colCount = len(headers)

		// Set headers
		for cIdx, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, 1)
			f.SetCellValue(sheetName, cell, header)
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}

		// Set data rows
		for rIdx, item := range data {
			row := rIdx + 2
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.ParticipantID)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ParticipantName)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.DeniedCount)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.LastAttemptAt.Format("2006-01-02 15:04:05 MST"))
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), strings.Join(item.Reasons, ", "))

			// Apply cell styles
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), normalStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("D%d", row), centerStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), normalStyle)

			// Highlight if attempts > 2 (higher potential fraud indicator)
			if item.DeniedCount > 1 {
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), deniedStyle)
			}
		}
		rowCount = len(data) + 1

	default:
		return nil, fmt.Errorf("unsupported report type: %s", reportType)
	}

	// Auto-fit column widths
	autoFitColumns(f, sheetName, rowCount, colCount)

	// Set row heights for visual breathing room
	f.SetRowHeight(sheetName, 1, 26) // Header row
	for r := 2; r <= rowCount; r++ {
		f.SetRowHeight(sheetName, r, 20)
	}

	// Show grid lines explicitly
	_ = f.SetSheetView(sheetName, 0, &excelize.ViewOptions{
		ShowGridLines: &[]bool{true}[0],
	})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to serialize excel workbook: %w", err)
	}

	return buf.Bytes(), nil
}

// autoFitColumns calculates the maximum text length and adjusts sheet column widths dynamically.
func autoFitColumns(f *excelize.File, sheetName string, maxRows int, numCols int) {
	for col := 1; col <= numCols; col++ {
		colName, _ := excelize.ColumnNumberToName(col)
		maxWidth := 12.0 // default minimum
		for row := 1; row <= maxRows; row++ {
			val, err := f.GetCellValue(sheetName, fmt.Sprintf("%s%d", colName, row))
			if err == nil && len(val) > 0 {
				width := float64(len(val)) + 3.0
				if width > maxWidth {
					maxWidth = width
				}
			}
		}
		// Cap maximum width at 50 to prevent overflow with very long values like concatenated reasons
		if maxWidth > 50.0 {
			maxWidth = 50.0
		}
		_ = f.SetColWidth(sheetName, colName, colName, maxWidth)
	}
}
