package postgres

import (
	"context"
	"fmt"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	db *pgxpool.Pool
}

// NewReportRepository creates a new PostgreSQL ReportRepository.
func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{
		db: db,
	}
}

// GetAccessLogs returns filtered, paginated access logs with total count.
func (r *ReportRepository) GetAccessLogs(ctx context.Context, filter *domain.AccessLogFilter) ([]*domain.AccessReportItem, int, error) {
	baseQuery := `
		FROM access_logs al
		LEFT JOIN participants p ON al.participant_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN zones z ON al.zone_id = z.id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filter.StartDate != nil {
		baseQuery += fmt.Sprintf(" AND al.created_at >= $%d", argCount)
		args = append(args, *filter.StartDate)
		argCount++
	}
	if filter.EndDate != nil {
		baseQuery += fmt.Sprintf(" AND al.created_at <= $%d", argCount)
		args = append(args, *filter.EndDate)
		argCount++
	}
	if filter.ZoneID != nil {
		baseQuery += fmt.Sprintf(" AND al.zone_id = $%d", argCount)
		args = append(args, *filter.ZoneID)
		argCount++
	}
	if filter.CategoryID != nil {
		baseQuery += fmt.Sprintf(" AND p.category_id = $%d", argCount)
		args = append(args, *filter.CategoryID)
		argCount++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND al.status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}

	countQuery := "SELECT COUNT(*)" + baseQuery
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count access logs: %w", err)
	}

	dataQuery := `
		SELECT 
			al.id, 
			al.participant_id, 
			COALESCE(p.full_name, 'Unknown / Invalid QR') as participant_name,
			COALESCE(c.name, '') as category_name,
			al.zone_id, 
			COALESCE(z.name, '') as zone_name,
			COALESCE(z.code, '') as zone_code,
			al.direction, 
			al.status, 
			al.reason, 
			al.created_at
	` + baseQuery + " ORDER BY al.created_at DESC"

	if filter.Limit > 0 {
		dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query access logs: %w", err)
	}
	defer rows.Close()

	list := make([]*domain.AccessReportItem, 0)
	for rows.Next() {
		var item domain.AccessReportItem
		var reason *string
		err := rows.Scan(
			&item.ID,
			&item.ParticipantID,
			&item.ParticipantName,
			&item.CategoryName,
			&item.ZoneID,
			&item.ZoneName,
			&item.ZoneCode,
			&item.Direction,
			&item.Status,
			&reason,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan access log row: %w", err)
		}
		if reason != nil {
			item.Reason = *reason
		}
		list = append(list, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return list, total, nil
}

// GetMealLogs returns filtered, paginated meal logs with total count.
func (r *ReportRepository) GetMealLogs(ctx context.Context, filter *domain.MealLogFilter) ([]*domain.MealReportItem, int, error) {
	baseQuery := `
		FROM meal_logs ml
		LEFT JOIN participants p ON ml.participant_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filter.StartDate != nil {
		baseQuery += fmt.Sprintf(" AND ml.created_at >= $%d", argCount)
		args = append(args, *filter.StartDate)
		argCount++
	}
	if filter.EndDate != nil {
		baseQuery += fmt.Sprintf(" AND ml.created_at <= $%d", argCount)
		args = append(args, *filter.EndDate)
		argCount++
	}
	if filter.CategoryID != nil {
		baseQuery += fmt.Sprintf(" AND p.category_id = $%d", argCount)
		args = append(args, *filter.CategoryID)
		argCount++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND ml.status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}

	countQuery := "SELECT COUNT(*)" + baseQuery
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count meal logs: %w", err)
	}

	dataQuery := `
		SELECT 
			ml.id, 
			ml.participant_id, 
			COALESCE(p.full_name, 'Unknown / Invalid QR') as participant_name,
			COALESCE(c.name, '') as category_name,
			ml.meal_schedule_id, 
			ml.meal_type, 
			ml.date::text, 
			ml.status, 
			ml.reason, 
			ml.created_at
	` + baseQuery + " ORDER BY ml.created_at DESC"

	if filter.Limit > 0 {
		dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query meal logs: %w", err)
	}
	defer rows.Close()

	list := make([]*domain.MealReportItem, 0)
	for rows.Next() {
		var item domain.MealReportItem
		var reason *string
		err := rows.Scan(
			&item.ID,
			&item.ParticipantID,
			&item.ParticipantName,
			&item.CategoryName,
			&item.MealScheduleID,
			&item.MealType,
			&item.Date,
			&item.Status,
			&reason,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan meal log row: %w", err)
		}
		if reason != nil {
			item.Reason = *reason
		}
		list = append(list, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return list, total, nil
}

// GetDeniedAttempts returns grouped repeatedly denied attempts from access and meal logs.
func (r *ReportRepository) GetDeniedAttempts(ctx context.Context) ([]*domain.DeniedAttemptGroup, error) {
	query := `
		WITH denied_combined AS (
			SELECT 
				participant_id,
				reason,
				created_at
			FROM access_logs
			WHERE status = 'DENIED'
			UNION ALL
			SELECT 
				participant_id,
				reason,
				created_at
			FROM meal_logs
			WHERE status = 'DENIED'
		)
		SELECT 
			dc.participant_id,
			COALESCE(p.full_name, 'Unknown / Invalid QR') as participant_name,
			COUNT(*) as denied_count,
			MAX(dc.created_at) as last_attempt_at,
			COALESCE(array_agg(DISTINCT dc.reason) FILTER (WHERE dc.reason IS NOT NULL AND dc.reason <> ''), '{}') as reasons
		FROM denied_combined dc
		LEFT JOIN participants p ON dc.participant_id = p.id
		GROUP BY dc.participant_id, p.full_name
		ORDER BY denied_count DESC;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query denied attempts: %w", err)
	}
	defer rows.Close()

	list := make([]*domain.DeniedAttemptGroup, 0)
	for rows.Next() {
		var item domain.DeniedAttemptGroup
		var reasons []string
		err := rows.Scan(
			&item.ParticipantID,
			&item.ParticipantName,
			&item.DeniedCount,
			&item.LastAttemptAt,
			&reasons,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan denied attempt group: %w", err)
		}
		item.Reasons = reasons
		list = append(list, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return list, nil
}
