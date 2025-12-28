package vaccination

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Vaccination, error)
	List(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error)
	Create(ctx context.Context, vax *Vaccination) error
	Update(ctx context.Context, vax *Vaccination) error
	Delete(ctx context.Context, id string) error
	GetUpcoming(ctx context.Context, childID string, days int) ([]Vaccination, error)
	GetSchedule() []VaccinationSchedule
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Vaccination, error) {
	query := `
		SELECT id, child_id, name, dose, scheduled_at, administered_at,
		       provider, location, lot_number, notes, completed, created_at, updated_at
		FROM vaccinations
		WHERE id = $1
	`

	var v Vaccination
	var administeredAt sql.NullTime
	var provider, location, lotNumber, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.ChildID, &v.Name, &v.Dose, &v.ScheduledAt, &administeredAt,
		&provider, &location, &lotNumber, &notes, &v.Completed, &v.CreatedAt, &v.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if administeredAt.Valid {
		v.AdministeredAt = &administeredAt.Time
	}
	if provider.Valid {
		v.Provider = provider.String
	}
	if location.Valid {
		v.Location = location.String
	}
	if lotNumber.Valid {
		v.LotNumber = lotNumber.String
	}
	if notes.Valid {
		v.Notes = notes.String
	}

	return &v, nil
}

func (r *repository) List(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
	query := `
		SELECT id, child_id, name, dose, scheduled_at, administered_at,
		       provider, location, lot_number, notes, completed, created_at, updated_at
		FROM vaccinations
		WHERE 1=1
	`
	args := []any{}
	argIndex := 1

	if filter.ChildID != "" {
		query += fmt.Sprintf(` AND child_id = $%d`, argIndex)
		args = append(args, filter.ChildID)
		argIndex++
	}

	if filter.Completed != nil {
		query += fmt.Sprintf(` AND completed = $%d`, argIndex)
		args = append(args, *filter.Completed)
		argIndex++
	}

	if filter.UpcomingOnly {
		query += fmt.Sprintf(` AND completed = false AND scheduled_at >= $%d`, argIndex)
		args = append(args, time.Now().Truncate(24*time.Hour))
	}

	query += ` ORDER BY scheduled_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // Best-effort close

	var vaccinations []Vaccination
	for rows.Next() {
		var v Vaccination
		var administeredAt sql.NullTime
		var provider, location, lotNumber, notes sql.NullString

		if err := rows.Scan(
			&v.ID, &v.ChildID, &v.Name, &v.Dose, &v.ScheduledAt, &administeredAt,
			&provider, &location, &lotNumber, &notes, &v.Completed, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if administeredAt.Valid {
			v.AdministeredAt = &administeredAt.Time
		}
		if provider.Valid {
			v.Provider = provider.String
		}
		if location.Valid {
			v.Location = location.String
		}
		if lotNumber.Valid {
			v.LotNumber = lotNumber.String
		}
		if notes.Valid {
			v.Notes = notes.String
		}

		vaccinations = append(vaccinations, v)
	}

	if vaccinations == nil {
		return []Vaccination{}, nil
	}

	return vaccinations, rows.Err()
}

func (r *repository) Create(ctx context.Context, vax *Vaccination) error {
	query := `
		INSERT INTO vaccinations (id, child_id, name, dose, scheduled_at, administered_at,
		                          provider, location, lot_number, notes, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	var provider, location, lotNumber, notes *string
	if vax.Provider != "" {
		provider = &vax.Provider
	}
	if vax.Location != "" {
		location = &vax.Location
	}
	if vax.LotNumber != "" {
		lotNumber = &vax.LotNumber
	}
	if vax.Notes != "" {
		notes = &vax.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		vax.ID, vax.ChildID, vax.Name, vax.Dose, vax.ScheduledAt, vax.AdministeredAt,
		provider, location, lotNumber, notes, vax.Completed, vax.CreatedAt, vax.UpdatedAt,
	)

	return err
}

func (r *repository) Update(ctx context.Context, vax *Vaccination) error {
	query := `
		UPDATE vaccinations
		SET name = $2, dose = $3, scheduled_at = $4, administered_at = $5,
		    provider = $6, location = $7, lot_number = $8, notes = $9,
		    completed = $10, updated_at = $11
		WHERE id = $1
	`

	var provider, location, lotNumber, notes *string
	if vax.Provider != "" {
		provider = &vax.Provider
	}
	if vax.Location != "" {
		location = &vax.Location
	}
	if vax.LotNumber != "" {
		lotNumber = &vax.LotNumber
	}
	if vax.Notes != "" {
		notes = &vax.Notes
	}

	_, err := r.db.ExecContext(ctx, query,
		vax.ID, vax.Name, vax.Dose, vax.ScheduledAt, vax.AdministeredAt,
		provider, location, lotNumber, notes, vax.Completed, vax.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM vaccinations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) GetUpcoming(ctx context.Context, childID string, days int) ([]Vaccination, error) {
	query := `
		SELECT id, child_id, name, dose, scheduled_at, administered_at,
		       provider, location, lot_number, notes, completed, created_at, updated_at
		FROM vaccinations
		WHERE child_id = $1
		  AND completed = false
		  AND scheduled_at >= $2
		  AND scheduled_at <= $3
		ORDER BY scheduled_at ASC
	`

	now := time.Now().Truncate(24 * time.Hour)
	endDate := now.AddDate(0, 0, days)

	rows, err := r.db.QueryContext(ctx, query, childID, now, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // Best-effort close

	var vaccinations []Vaccination
	for rows.Next() {
		var v Vaccination
		var administeredAt sql.NullTime
		var provider, location, lotNumber, notes sql.NullString

		if err := rows.Scan(
			&v.ID, &v.ChildID, &v.Name, &v.Dose, &v.ScheduledAt, &administeredAt,
			&provider, &location, &lotNumber, &notes, &v.Completed, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if administeredAt.Valid {
			v.AdministeredAt = &administeredAt.Time
		}
		if provider.Valid {
			v.Provider = provider.String
		}
		if location.Valid {
			v.Location = location.String
		}
		if lotNumber.Valid {
			v.LotNumber = lotNumber.String
		}
		if notes.Valid {
			v.Notes = notes.String
		}

		vaccinations = append(vaccinations, v)
	}

	if vaccinations == nil {
		return []Vaccination{}, nil
	}

	return vaccinations, rows.Err()
}

func (r *repository) GetSchedule() []VaccinationSchedule {
	// Kenya Expanded Program on Immunization (EPI) schedule
	return []VaccinationSchedule{
		// Birth
		{ID: "bcg-1", Name: "BCG", Description: "Bacillus Calmette-Guérin (Tuberculosis)", AgeWeeks: 0, AgeMonths: 0, AgeLabel: "Birth", Dose: 1},
		{ID: "opv-0", Name: "OPV", Description: "Oral Polio Vaccine - Birth dose", AgeWeeks: 0, AgeMonths: 0, AgeLabel: "Birth", Dose: 0},
		{ID: "hepb-1", Name: "Hepatitis B", Description: "Birth dose", AgeWeeks: 0, AgeMonths: 0, AgeLabel: "Birth", Dose: 1},

		// 6 weeks
		{ID: "penta-1", Name: "Pentavalent", Description: "DPT-HepB-Hib - First dose", AgeWeeks: 6, AgeMonths: 1, AgeLabel: "6 weeks", Dose: 1},
		{ID: "opv-1", Name: "OPV", Description: "Oral Polio Vaccine - First dose", AgeWeeks: 6, AgeMonths: 1, AgeLabel: "6 weeks", Dose: 1},
		{ID: "pcv-1", Name: "PCV", Description: "Pneumococcal - First dose", AgeWeeks: 6, AgeMonths: 1, AgeLabel: "6 weeks", Dose: 1},
		{ID: "rv-1", Name: "Rotavirus", Description: "First dose", AgeWeeks: 6, AgeMonths: 1, AgeLabel: "6 weeks", Dose: 1},

		// 10 weeks
		{ID: "penta-2", Name: "Pentavalent", Description: "DPT-HepB-Hib - Second dose", AgeWeeks: 10, AgeMonths: 2, AgeLabel: "10 weeks", Dose: 2},
		{ID: "opv-2", Name: "OPV", Description: "Oral Polio Vaccine - Second dose", AgeWeeks: 10, AgeMonths: 2, AgeLabel: "10 weeks", Dose: 2},
		{ID: "pcv-2", Name: "PCV", Description: "Pneumococcal - Second dose", AgeWeeks: 10, AgeMonths: 2, AgeLabel: "10 weeks", Dose: 2},
		{ID: "rv-2", Name: "Rotavirus", Description: "Second dose", AgeWeeks: 10, AgeMonths: 2, AgeLabel: "10 weeks", Dose: 2},

		// 14 weeks
		{ID: "penta-3", Name: "Pentavalent", Description: "DPT-HepB-Hib - Third dose", AgeWeeks: 14, AgeMonths: 3, AgeLabel: "14 weeks", Dose: 3},
		{ID: "opv-3", Name: "OPV", Description: "Oral Polio Vaccine - Third dose", AgeWeeks: 14, AgeMonths: 3, AgeLabel: "14 weeks", Dose: 3},
		{ID: "ipv-1", Name: "IPV", Description: "Inactivated Polio Vaccine", AgeWeeks: 14, AgeMonths: 3, AgeLabel: "14 weeks", Dose: 1},
		{ID: "pcv-3", Name: "PCV", Description: "Pneumococcal - Third dose", AgeWeeks: 14, AgeMonths: 3, AgeLabel: "14 weeks", Dose: 3},

		// 6 months
		{ID: "vita-1", Name: "Vitamin A", Description: "First supplement", AgeWeeks: 26, AgeMonths: 6, AgeLabel: "6 months", Dose: 1},

		// 9 months
		{ID: "mr-1", Name: "Measles-Rubella", Description: "MR - First dose", AgeWeeks: 39, AgeMonths: 9, AgeLabel: "9 months", Dose: 1},
		{ID: "yellow-fever-1", Name: "Yellow Fever", Description: "Single dose (endemic areas)", AgeWeeks: 39, AgeMonths: 9, AgeLabel: "9 months", Dose: 1},

		// 12 months
		{ID: "vita-2", Name: "Vitamin A", Description: "Second supplement", AgeWeeks: 52, AgeMonths: 12, AgeLabel: "12 months", Dose: 2},

		// 18 months
		{ID: "mr-2", Name: "Measles-Rubella", Description: "MR - Second dose", AgeWeeks: 78, AgeMonths: 18, AgeLabel: "18 months", Dose: 2},
		{ID: "vita-3", Name: "Vitamin A", Description: "Third supplement", AgeWeeks: 78, AgeMonths: 18, AgeLabel: "18 months", Dose: 3},
	}
}
