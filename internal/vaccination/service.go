package vaccination

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error)
	Get(ctx context.Context, id string) (*Vaccination, error)
	List(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error)
	Update(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error)
	Delete(ctx context.Context, id string) error
	RecordAdministration(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error)
	GetUpcoming(ctx context.Context, childID string, days int) ([]Vaccination, error)
	GetSchedule() []VaccinationSchedule
	GenerateScheduleForChild(ctx context.Context, childID string, birthDate string) ([]Vaccination, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Get(ctx context.Context, id string) (*Vaccination, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) RecordAdministration(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
	// TODO: implement - mark vaccination as completed with details
	return nil, nil
}

func (s *service) GetUpcoming(ctx context.Context, childID string, days int) ([]Vaccination, error) {
	return s.repo.GetUpcoming(ctx, childID, days)
}

func (s *service) GetSchedule() []VaccinationSchedule {
	return s.repo.GetSchedule()
}

func (s *service) GenerateScheduleForChild(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
	// TODO: implement - create vaccination records based on standard schedule
	return nil, nil
}
