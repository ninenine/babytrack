package medication

import (
	"context"
)

type Service interface {
	// Medications
	Create(ctx context.Context, req *CreateMedicationRequest) (*Medication, error)
	Get(ctx context.Context, id string) (*Medication, error)
	List(ctx context.Context, filter *MedicationFilter) ([]Medication, error)
	Update(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error)
	Delete(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error

	// Medication Logs
	LogMedication(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error)
	GetLogs(ctx context.Context, medicationID string) ([]MedicationLog, error)
	GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateMedicationRequest) (*Medication, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Get(ctx context.Context, id string) (*Medication, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Deactivate(ctx context.Context, id string) error {
	// TODO: implement - set active = false
	return nil
}

func (s *service) LogMedication(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) GetLogs(ctx context.Context, medicationID string) ([]MedicationLog, error) {
	return s.repo.ListLogs(ctx, medicationID)
}

func (s *service) GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error) {
	return s.repo.GetLastLog(ctx, medicationID)
}
