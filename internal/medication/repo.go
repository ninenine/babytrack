package medication

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	// Medications
	GetByID(ctx context.Context, id string) (*Medication, error)
	List(ctx context.Context, filter *MedicationFilter) ([]Medication, error)
	Create(ctx context.Context, med *Medication) error
	Update(ctx context.Context, med *Medication) error
	Delete(ctx context.Context, id string) error

	// Medication Logs
	GetLogByID(ctx context.Context, id string) (*MedicationLog, error)
	ListLogs(ctx context.Context, medicationID string) ([]MedicationLog, error)
	CreateLog(ctx context.Context, log *MedicationLog) error
	GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Medication, error) {
	query := `
		SELECT id, child_id, name, dosage, unit, frequency, instructions,
		       start_date, end_date, active, created_at, updated_at
		FROM medications
		WHERE id = $1
	`

	var m Medication
	var instructions sql.NullString
	var endDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.ChildID, &m.Name, &m.Dosage, &m.Unit, &m.Frequency,
		&instructions, &m.StartDate, &endDate, &m.Active, &m.CreatedAt, &m.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if instructions.Valid {
		m.Instructions = instructions.String
	}
	if endDate.Valid {
		m.EndDate = &endDate.Time
	}

	return &m, nil
}

func (r *repository) List(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
	query := `
		SELECT id, child_id, name, dosage, unit, frequency, instructions,
		       start_date, end_date, active, created_at, updated_at
		FROM medications
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.ChildID != "" {
		query += fmt.Sprintf(` AND child_id = $%d`, argIndex)
		args = append(args, filter.ChildID)
		argIndex++
	}

	if filter.ActiveOnly {
		query += fmt.Sprintf(` AND active = $%d`, argIndex)
		args = append(args, true)
		argIndex++
	}

	query += ` ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medications []Medication
	for rows.Next() {
		var m Medication
		var instructions sql.NullString
		var endDate sql.NullTime

		if err := rows.Scan(
			&m.ID, &m.ChildID, &m.Name, &m.Dosage, &m.Unit, &m.Frequency,
			&instructions, &m.StartDate, &endDate, &m.Active, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if instructions.Valid {
			m.Instructions = instructions.String
		}
		if endDate.Valid {
			m.EndDate = &endDate.Time
		}

		medications = append(medications, m)
	}

	if medications == nil {
		return []Medication{}, nil
	}

	return medications, rows.Err()
}

func (r *repository) Create(ctx context.Context, med *Medication) error {
	query := `
		INSERT INTO medications (id, child_id, name, dosage, unit, frequency, instructions,
		                         start_date, end_date, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var instructions *string
	if med.Instructions != "" {
		instructions = &med.Instructions
	}

	_, err := r.db.ExecContext(ctx, query,
		med.ID, med.ChildID, med.Name, med.Dosage, med.Unit, med.Frequency,
		instructions, med.StartDate, med.EndDate, med.Active,
		med.CreatedAt, med.UpdatedAt,
	)

	return err
}

func (r *repository) Update(ctx context.Context, med *Medication) error {
	query := `
		UPDATE medications
		SET name = $2, dosage = $3, unit = $4, frequency = $5, instructions = $6,
		    start_date = $7, end_date = $8, active = $9, updated_at = $10
		WHERE id = $1
	`

	var instructions *string
	if med.Instructions != "" {
		instructions = &med.Instructions
	}

	_, err := r.db.ExecContext(ctx, query,
		med.ID, med.Name, med.Dosage, med.Unit, med.Frequency,
		instructions, med.StartDate, med.EndDate, med.Active, med.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM medications WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) GetLogByID(ctx context.Context, id string) (*MedicationLog, error) {
	query := `
		SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at
		FROM medication_logs
		WHERE id = $1
	`

	var log MedicationLog
	var notes sql.NullString
	var syncedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.MedicationID, &log.ChildID, &log.GivenAt, &log.GivenBy,
		&log.Dosage, &notes, &log.CreatedAt, &syncedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if notes.Valid {
		log.Notes = notes.String
	}
	if syncedAt.Valid {
		log.SyncedAt = &syncedAt.Time
	}

	return &log, nil
}

func (r *repository) ListLogs(ctx context.Context, medicationID string) ([]MedicationLog, error) {
	query := `
		SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at
		FROM medication_logs
		WHERE medication_id = $1
		ORDER BY given_at DESC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, medicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []MedicationLog
	for rows.Next() {
		var log MedicationLog
		var notes sql.NullString
		var syncedAt sql.NullTime

		if err := rows.Scan(
			&log.ID, &log.MedicationID, &log.ChildID, &log.GivenAt, &log.GivenBy,
			&log.Dosage, &notes, &log.CreatedAt, &syncedAt,
		); err != nil {
			return nil, err
		}

		if notes.Valid {
			log.Notes = notes.String
		}
		if syncedAt.Valid {
			log.SyncedAt = &syncedAt.Time
		}

		logs = append(logs, log)
	}

	if logs == nil {
		return []MedicationLog{}, nil
	}

	return logs, rows.Err()
}

func (r *repository) CreateLog(ctx context.Context, log *MedicationLog) error {
	query := `
		INSERT INTO medication_logs (id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	var notes *string
	if log.Notes != "" {
		notes = &log.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.MedicationID, log.ChildID, log.GivenAt, log.GivenBy,
		log.Dosage, notes, log.CreatedAt, log.SyncedAt,
	)

	return err
}

func (r *repository) GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error) {
	query := `
		SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at
		FROM medication_logs
		WHERE medication_id = $1
		ORDER BY given_at DESC
		LIMIT 1
	`

	var log MedicationLog
	var notes sql.NullString
	var syncedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, medicationID).Scan(
		&log.ID, &log.MedicationID, &log.ChildID, &log.GivenAt, &log.GivenBy,
		&log.Dosage, &notes, &log.CreatedAt, &syncedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if notes.Valid {
		log.Notes = notes.String
	}
	if syncedAt.Valid {
		log.SyncedAt = &syncedAt.Time
	}

	return &log, nil
}
