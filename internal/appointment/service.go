package appointment

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error)
	Get(ctx context.Context, id string) (*Appointment, error)
	List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error)
	Update(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error)
	Delete(ctx context.Context, id string) error
	Complete(ctx context.Context, id string) error
	Cancel(ctx context.Context, id string) error
	GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Get(ctx context.Context, id string) (*Appointment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Complete(ctx context.Context, id string) error {
	// TODO: implement - set completed = true
	return nil
}

func (s *service) Cancel(ctx context.Context, id string) error {
	// TODO: implement - set cancelled = true
	return nil
}

func (s *service) GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error) {
	return s.repo.GetUpcoming(ctx, childID, days)
}
