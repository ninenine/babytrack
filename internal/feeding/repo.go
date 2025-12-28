package feeding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Feeding, error)
	List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error)
	Create(ctx context.Context, feeding *Feeding) error
	Update(ctx context.Context, feeding *Feeding) error
	Delete(ctx context.Context, id string) error
	GetLastFeeding(ctx context.Context, childID string) (*Feeding, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Feeding, error) {
	query := `
		SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at
		FROM feedings
		WHERE id = $1
	`

	var f Feeding
	var endTime, syncedAt sql.NullTime
	var amount sql.NullFloat64
	var unit, side, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID, &f.ChildID, &f.Type, &f.StartTime, &endTime,
		&amount, &unit, &side, &notes, &f.CreatedAt, &f.UpdatedAt, &syncedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		f.EndTime = &endTime.Time
	}
	if amount.Valid {
		f.Amount = &amount.Float64
	}
	if unit.Valid {
		f.Unit = unit.String
	}
	if side.Valid {
		f.Side = side.String
	}
	if notes.Valid {
		f.Notes = notes.String
	}
	if syncedAt.Valid {
		f.SyncedAt = &syncedAt.Time
	}

	return &f, nil
}

func (r *repository) List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
	query := `
		SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at
		FROM feedings
		WHERE 1=1
	`
	args := []any{}
	argIndex := 1

	if filter.ChildID != "" {
		query += fmt.Sprintf(` AND child_id = $%d`, argIndex)
		args = append(args, filter.ChildID)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(` AND start_time >= $%d`, argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(` AND start_time <= $%d`, argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.Type != nil {
		query += fmt.Sprintf(` AND type = $%d`, argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}

	query += ` ORDER BY start_time DESC LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedings []Feeding
	for rows.Next() {
		var f Feeding
		var endTime, syncedAt sql.NullTime
		var amount sql.NullFloat64
		var unit, side, notes sql.NullString

		if err := rows.Scan(
			&f.ID, &f.ChildID, &f.Type, &f.StartTime, &endTime,
			&amount, &unit, &side, &notes, &f.CreatedAt, &f.UpdatedAt, &syncedAt,
		); err != nil {
			return nil, err
		}

		if endTime.Valid {
			f.EndTime = &endTime.Time
		}
		if amount.Valid {
			f.Amount = &amount.Float64
		}
		if unit.Valid {
			f.Unit = unit.String
		}
		if side.Valid {
			f.Side = side.String
		}
		if notes.Valid {
			f.Notes = notes.String
		}
		if syncedAt.Valid {
			f.SyncedAt = &syncedAt.Time
		}

		feedings = append(feedings, f)
	}

	if feedings == nil {
		return []Feeding{}, nil
	}

	return feedings, rows.Err()
}

func (r *repository) Create(ctx context.Context, feeding *Feeding) error {
	query := `
		INSERT INTO feedings (id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	var unit, side, notes *string
	if feeding.Unit != "" {
		unit = &feeding.Unit
	}
	if feeding.Side != "" {
		side = &feeding.Side
	}
	if feeding.Notes != "" {
		notes = &feeding.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		feeding.ID,
		feeding.ChildID,
		feeding.Type,
		feeding.StartTime,
		feeding.EndTime,
		feeding.Amount,
		unit,
		side,
		notes,
		feeding.CreatedAt,
		feeding.UpdatedAt,
	)

	return err
}

func (r *repository) Update(ctx context.Context, feeding *Feeding) error {
	query := `
		UPDATE feedings
		SET type = $2, start_time = $3, end_time = $4, amount = $5, unit = $6, side = $7, notes = $8, updated_at = $9
		WHERE id = $1
	`

	var unit, side, notes *string
	if feeding.Unit != "" {
		unit = &feeding.Unit
	}
	if feeding.Side != "" {
		side = &feeding.Side
	}
	if feeding.Notes != "" {
		notes = &feeding.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		feeding.ID,
		feeding.Type,
		feeding.StartTime,
		feeding.EndTime,
		feeding.Amount,
		unit,
		side,
		notes,
		feeding.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM feedings WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) GetLastFeeding(ctx context.Context, childID string) (*Feeding, error) {
	query := `
		SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at
		FROM feedings
		WHERE child_id = $1
		ORDER BY start_time DESC
		LIMIT 1
	`

	var f Feeding
	var endTime, syncedAt sql.NullTime
	var amount sql.NullFloat64
	var unit, side, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, childID).Scan(
		&f.ID, &f.ChildID, &f.Type, &f.StartTime, &endTime,
		&amount, &unit, &side, &notes, &f.CreatedAt, &f.UpdatedAt, &syncedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		f.EndTime = &endTime.Time
	}
	if amount.Valid {
		f.Amount = &amount.Float64
	}
	if unit.Valid {
		f.Unit = unit.String
	}
	if side.Valid {
		f.Side = side.String
	}
	if notes.Valid {
		f.Notes = notes.String
	}
	if syncedAt.Valid {
		f.SyncedAt = &syncedAt.Time
	}

	return &f, nil
}
