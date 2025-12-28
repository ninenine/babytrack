package appointment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Appointment, error)
	List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error)
	Create(ctx context.Context, apt *Appointment) error
	Update(ctx context.Context, apt *Appointment) error
	Delete(ctx context.Context, id string) error
	GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Appointment, error) {
	query := `
		SELECT id, child_id, type, title, provider, location, scheduled_at,
		       duration, notes, completed, cancelled, created_at, updated_at
		FROM appointments
		WHERE id = $1
	`

	var a Appointment
	var provider, location, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.ChildID, &a.Type, &a.Title, &provider, &location, &a.ScheduledAt,
		&a.Duration, &notes, &a.Completed, &a.Cancelled, &a.CreatedAt, &a.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if provider.Valid {
		a.Provider = provider.String
	}
	if location.Valid {
		a.Location = location.String
	}
	if notes.Valid {
		a.Notes = notes.String
	}

	return &a, nil
}

func (r *repository) List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
	query := `
		SELECT id, child_id, type, title, provider, location, scheduled_at,
		       duration, notes, completed, cancelled, created_at, updated_at
		FROM appointments
		WHERE 1=1
	`
	args := []any{}
	argIndex := 1

	if filter.ChildID != "" {
		query += fmt.Sprintf(` AND child_id = $%d`, argIndex)
		args = append(args, filter.ChildID)
		argIndex++
	}

	if filter.Type != nil {
		query += fmt.Sprintf(` AND type = $%d`, argIndex)
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.UpcomingOnly {
		query += fmt.Sprintf(` AND completed = false AND cancelled = false AND scheduled_at >= $%d`, argIndex)
		args = append(args, time.Now())
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(` AND scheduled_at >= $%d`, argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(` AND scheduled_at <= $%d`, argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	query += ` ORDER BY scheduled_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var a Appointment
		var provider, location, notes sql.NullString

		if err := rows.Scan(
			&a.ID, &a.ChildID, &a.Type, &a.Title, &provider, &location, &a.ScheduledAt,
			&a.Duration, &notes, &a.Completed, &a.Cancelled, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if provider.Valid {
			a.Provider = provider.String
		}
		if location.Valid {
			a.Location = location.String
		}
		if notes.Valid {
			a.Notes = notes.String
		}

		appointments = append(appointments, a)
	}

	if appointments == nil {
		return []Appointment{}, nil
	}

	return appointments, rows.Err()
}

func (r *repository) Create(ctx context.Context, apt *Appointment) error {
	query := `
		INSERT INTO appointments (id, child_id, type, title, provider, location, scheduled_at,
		                          duration, notes, completed, cancelled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	var provider, location, notes *string
	if apt.Provider != "" {
		provider = &apt.Provider
	}
	if apt.Location != "" {
		location = &apt.Location
	}
	if apt.Notes != "" {
		notes = &apt.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		apt.ID, apt.ChildID, apt.Type, apt.Title, provider, location, apt.ScheduledAt,
		apt.Duration, notes, apt.Completed, apt.Cancelled, apt.CreatedAt, apt.UpdatedAt,
	)

	return err
}

func (r *repository) Update(ctx context.Context, apt *Appointment) error {
	query := `
		UPDATE appointments
		SET type = $2, title = $3, provider = $4, location = $5, scheduled_at = $6,
		    duration = $7, notes = $8, completed = $9, cancelled = $10, updated_at = $11
		WHERE id = $1
	`

	var provider, location, notes *string
	if apt.Provider != "" {
		provider = &apt.Provider
	}
	if apt.Location != "" {
		location = &apt.Location
	}
	if apt.Notes != "" {
		notes = &apt.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		apt.ID, apt.Type, apt.Title, provider, location, apt.ScheduledAt,
		apt.Duration, notes, apt.Completed, apt.Cancelled, apt.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM appointments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error) {
	query := `
		SELECT id, child_id, type, title, provider, location, scheduled_at,
		       duration, notes, completed, cancelled, created_at, updated_at
		FROM appointments
		WHERE child_id = $1
		  AND completed = false
		  AND cancelled = false
		  AND scheduled_at >= $2
		  AND scheduled_at <= $3
		ORDER BY scheduled_at ASC
	`

	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	rows, err := r.db.QueryContext(ctx, query, childID, now, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var a Appointment
		var provider, location, notes sql.NullString

		if err := rows.Scan(
			&a.ID, &a.ChildID, &a.Type, &a.Title, &provider, &location, &a.ScheduledAt,
			&a.Duration, &notes, &a.Completed, &a.Cancelled, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if provider.Valid {
			a.Provider = provider.String
		}
		if location.Valid {
			a.Location = location.String
		}
		if notes.Valid {
			a.Notes = notes.String
		}

		appointments = append(appointments, a)
	}

	if appointments == nil {
		return []Appointment{}, nil
	}

	return appointments, rows.Err()
}
