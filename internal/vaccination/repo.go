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
	args := []interface{}{}
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
		argIndex++
	}

	query += ` ORDER BY scheduled_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	defer rows.Close()

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
	// Standard CDC vaccination schedule for children
	return []VaccinationSchedule{
		// Birth
		{ID: "hepb-1", Name: "Hepatitis B", Description: "First dose at birth", AgeMonths: 0, Dose: 1},
		// 2 months
		{ID: "hepb-2", Name: "Hepatitis B", Description: "Second dose", AgeMonths: 2, Dose: 2},
		{ID: "dtap-1", Name: "DTaP", Description: "Diphtheria, Tetanus, Pertussis - First dose", AgeMonths: 2, Dose: 1},
		{ID: "polio-1", Name: "Polio (IPV)", Description: "First dose", AgeMonths: 2, Dose: 1},
		{ID: "hib-1", Name: "Hib", Description: "Haemophilus influenzae type b - First dose", AgeMonths: 2, Dose: 1},
		{ID: "pcv-1", Name: "PCV13", Description: "Pneumococcal - First dose", AgeMonths: 2, Dose: 1},
		{ID: "rv-1", Name: "Rotavirus", Description: "First dose", AgeMonths: 2, Dose: 1},
		// 4 months
		{ID: "dtap-2", Name: "DTaP", Description: "Second dose", AgeMonths: 4, Dose: 2},
		{ID: "polio-2", Name: "Polio (IPV)", Description: "Second dose", AgeMonths: 4, Dose: 2},
		{ID: "hib-2", Name: "Hib", Description: "Second dose", AgeMonths: 4, Dose: 2},
		{ID: "pcv-2", Name: "PCV13", Description: "Second dose", AgeMonths: 4, Dose: 2},
		{ID: "rv-2", Name: "Rotavirus", Description: "Second dose", AgeMonths: 4, Dose: 2},
		// 6 months
		{ID: "hepb-3", Name: "Hepatitis B", Description: "Third dose", AgeMonths: 6, Dose: 3},
		{ID: "dtap-3", Name: "DTaP", Description: "Third dose", AgeMonths: 6, Dose: 3},
		{ID: "polio-3", Name: "Polio (IPV)", Description: "Third dose", AgeMonths: 6, Dose: 3},
		{ID: "hib-3", Name: "Hib", Description: "Third dose (if needed)", AgeMonths: 6, Dose: 3},
		{ID: "pcv-3", Name: "PCV13", Description: "Third dose", AgeMonths: 6, Dose: 3},
		{ID: "rv-3", Name: "Rotavirus", Description: "Third dose (if needed)", AgeMonths: 6, Dose: 3},
		{ID: "flu-1", Name: "Influenza", Description: "Annual flu shot (6+ months)", AgeMonths: 6, Dose: 1},
		// 12 months
		{ID: "mmr-1", Name: "MMR", Description: "Measles, Mumps, Rubella - First dose", AgeMonths: 12, Dose: 1},
		{ID: "varicella-1", Name: "Varicella", Description: "Chickenpox - First dose", AgeMonths: 12, Dose: 1},
		{ID: "hepa-1", Name: "Hepatitis A", Description: "First dose", AgeMonths: 12, Dose: 1},
		{ID: "pcv-4", Name: "PCV13", Description: "Fourth dose", AgeMonths: 12, Dose: 4},
		{ID: "hib-4", Name: "Hib", Description: "Booster", AgeMonths: 12, Dose: 4},
		// 15-18 months
		{ID: "dtap-4", Name: "DTaP", Description: "Fourth dose", AgeMonths: 15, Dose: 4},
		// 18 months
		{ID: "hepa-2", Name: "Hepatitis A", Description: "Second dose", AgeMonths: 18, Dose: 2},
		// 4-6 years
		{ID: "dtap-5", Name: "DTaP", Description: "Fifth dose", AgeMonths: 48, Dose: 5},
		{ID: "polio-4", Name: "Polio (IPV)", Description: "Fourth dose", AgeMonths: 48, Dose: 4},
		{ID: "mmr-2", Name: "MMR", Description: "Second dose", AgeMonths: 48, Dose: 2},
		{ID: "varicella-2", Name: "Varicella", Description: "Second dose", AgeMonths: 48, Dose: 2},
	}
}
