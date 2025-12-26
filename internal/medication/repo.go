package medication

import (
	"context"
	"database/sql"
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
	// TODO: implement
	return nil, nil
}

func (r *repository) List(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) Create(ctx context.Context, med *Medication) error {
	// TODO: implement
	return nil
}

func (r *repository) Update(ctx context.Context, med *Medication) error {
	// TODO: implement
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (r *repository) GetLogByID(ctx context.Context, id string) (*MedicationLog, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) ListLogs(ctx context.Context, medicationID string) ([]MedicationLog, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) CreateLog(ctx context.Context, log *MedicationLog) error {
	// TODO: implement
	return nil
}

func (r *repository) GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error) {
	// TODO: implement
	return nil, nil
}
