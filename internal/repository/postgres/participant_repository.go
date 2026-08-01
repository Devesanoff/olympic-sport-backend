package postgres

import (
	"context"
	"fmt"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ParticipantRepository struct {
	db *pgxpool.Pool
}

// NewParticipantRepository creates a new PostgreSQL participant repository.
func NewParticipantRepository(db *pgxpool.Pool) *ParticipantRepository {
	return &ParticipantRepository{
		db: db,
	}
}

// Create inserts a new Participant record into the database.
func (r *ParticipantRepository) Create(ctx context.Context, p *domain.Participant) error {
	query := `
		INSERT INTO participants (id, full_name, category_id, qr_token, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at;
	`
	err := r.db.QueryRow(ctx, query,
		p.ID,
		p.FullName,
		p.CategoryID,
		p.QRToken,
		p.Status,
		p.CreatedAt,
		p.UpdatedAt,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create participant: %w", err)
	}

	return nil
}

// GetByID retrieves a single Participant by their unique UUID.
func (r *ParticipantRepository) GetByID(ctx context.Context, id string) (*domain.Participant, error) {
	query := `
		SELECT id, full_name, category_id, qr_token, status, created_at, updated_at
		FROM participants
		WHERE id = $1;
	`

	var p domain.Participant
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.FullName,
		&p.CategoryID,
		&p.QRToken,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant by id: %w", err)
	}

	return &p, nil
}

// List returns a paginated list of participants and the total count.
func (r *ParticipantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Participant, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM participants;`
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total participants count: %w", err)
	}

	query := `
		SELECT id, full_name, category_id, qr_token, status, created_at, updated_at
		FROM participants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query participants list: %w", err)
	}
	defer rows.Close()

	list := make([]*domain.Participant, 0)
	for rows.Next() {
		var p domain.Participant
		err := rows.Scan(
			&p.ID,
			&p.FullName,
			&p.CategoryID,
			&p.QRToken,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan participant row: %w", err)
		}
		list = append(list, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return list, total, nil
}
