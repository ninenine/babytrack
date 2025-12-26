package appointment

import (
	"context"
	"database/sql"
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
	// TODO: implement
	return nil, nil
}

func (r *repository) List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) Create(ctx context.Context, apt *Appointment) error {
	// TODO: implement
	return nil
}

func (r *repository) Update(ctx context.Context, apt *Appointment) error {
	// TODO: implement
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (r *repository) GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error) {
	// TODO: implement
	return nil, nil
}
