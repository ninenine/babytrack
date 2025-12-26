package sleep

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Sleep, error)
	List(ctx context.Context, filter *SleepFilter) ([]Sleep, error)
	Create(ctx context.Context, sleep *Sleep) error
	Update(ctx context.Context, sleep *Sleep) error
	Delete(ctx context.Context, id string) error
	GetActiveSleep(ctx context.Context, childID string) (*Sleep, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Sleep, error) {
	query := `
		SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at
		FROM sleep
		WHERE id = $1
	`

	var s Sleep
	var endTime, syncedAt sql.NullTime
	var quality sql.NullInt32
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.ChildID, &s.Type, &s.StartTime, &endTime,
		&quality, &notes, &s.CreatedAt, &s.UpdatedAt, &syncedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		s.EndTime = &endTime.Time
	}
	if quality.Valid {
		q := int(quality.Int32)
		s.Quality = &q
	}
	if notes.Valid {
		s.Notes = notes.String
	}
	if syncedAt.Valid {
		s.SyncedAt = &syncedAt.Time
	}

	return &s, nil
}

func (r *repository) List(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
	query := `
		SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at
		FROM sleep
		WHERE 1=1
	`
	args := []interface{}{}
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

	var sleeps []Sleep
	for rows.Next() {
		var s Sleep
		var endTime, syncedAt sql.NullTime
		var quality sql.NullInt32
		var notes sql.NullString

		if err := rows.Scan(
			&s.ID, &s.ChildID, &s.Type, &s.StartTime, &endTime,
			&quality, &notes, &s.CreatedAt, &s.UpdatedAt, &syncedAt,
		); err != nil {
			return nil, err
		}

		if endTime.Valid {
			s.EndTime = &endTime.Time
		}
		if quality.Valid {
			q := int(quality.Int32)
			s.Quality = &q
		}
		if notes.Valid {
			s.Notes = notes.String
		}
		if syncedAt.Valid {
			s.SyncedAt = &syncedAt.Time
		}

		sleeps = append(sleeps, s)
	}

	if sleeps == nil {
		return []Sleep{}, nil
	}

	return sleeps, rows.Err()
}

func (r *repository) Create(ctx context.Context, sleep *Sleep) error {
	query := `
		INSERT INTO sleep (id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	var notes *string
	if sleep.Notes != "" {
		notes = &sleep.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		sleep.ID,
		sleep.ChildID,
		sleep.Type,
		sleep.StartTime,
		sleep.EndTime,
		sleep.Quality,
		notes,
		sleep.CreatedAt,
		sleep.UpdatedAt,
	)

	return err
}

func (r *repository) Update(ctx context.Context, sleep *Sleep) error {
	query := `
		UPDATE sleep
		SET type = $2, start_time = $3, end_time = $4, quality = $5, notes = $6, updated_at = $7
		WHERE id = $1
	`

	var notes *string
	if sleep.Notes != "" {
		notes = &sleep.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		sleep.ID,
		sleep.Type,
		sleep.StartTime,
		sleep.EndTime,
		sleep.Quality,
		notes,
		sleep.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sleep WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) GetActiveSleep(ctx context.Context, childID string) (*Sleep, error) {
	query := `
		SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at
		FROM sleep
		WHERE child_id = $1 AND end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`

	var s Sleep
	var endTime, syncedAt sql.NullTime
	var quality sql.NullInt32
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, childID).Scan(
		&s.ID, &s.ChildID, &s.Type, &s.StartTime, &endTime,
		&quality, &notes, &s.CreatedAt, &s.UpdatedAt, &syncedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		s.EndTime = &endTime.Time
	}
	if quality.Valid {
		q := int(quality.Int32)
		s.Quality = &q
	}
	if notes.Valid {
		s.Notes = notes.String
	}
	if syncedAt.Valid {
		s.SyncedAt = &syncedAt.Time
	}

	return &s, nil
}
