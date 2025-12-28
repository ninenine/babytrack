package medication

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
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
	now := time.Now()

	med := &Medication{
		ID:           generateID(),
		ChildID:      req.ChildID,
		Name:         req.Name,
		Dosage:       req.Dosage,
		Unit:         req.Unit,
		Frequency:    req.Frequency,
		Instructions: req.Instructions,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, med); err != nil {
		return nil, fmt.Errorf("failed to create medication: %w", err)
	}

	return med, nil
}

func (s *service) Get(ctx context.Context, id string) (*Medication, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error) {
	med, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if med == nil {
		return nil, fmt.Errorf("medication not found")
	}

	med.Name = req.Name
	med.Dosage = req.Dosage
	med.Unit = req.Unit
	med.Frequency = req.Frequency
	med.Instructions = req.Instructions
	med.StartDate = req.StartDate
	med.EndDate = req.EndDate
	med.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, med); err != nil {
		return nil, fmt.Errorf("failed to update medication: %w", err)
	}

	return med, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Deactivate(ctx context.Context, id string) error {
	med, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if med == nil {
		return fmt.Errorf("medication not found")
	}

	med.Active = false
	now := time.Now()
	med.EndDate = &now
	med.UpdatedAt = now

	if err := s.repo.Update(ctx, med); err != nil {
		return fmt.Errorf("failed to deactivate medication: %w", err)
	}

	return nil
}

func (s *service) LogMedication(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
	// Get the medication to get the child ID
	med, err := s.repo.GetByID(ctx, req.MedicationID)
	if err != nil {
		return nil, err
	}
	if med == nil {
		return nil, fmt.Errorf("medication not found")
	}

	now := time.Now()

	log := &MedicationLog{
		ID:           generateID(),
		MedicationID: req.MedicationID,
		ChildID:      med.ChildID,
		GivenAt:      req.GivenAt,
		GivenBy:      userID,
		Dosage:       req.Dosage,
		Notes:        req.Notes,
		CreatedAt:    now,
		SyncedAt:     &now,
	}

	if err := s.repo.CreateLog(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to log medication: %w", err)
	}

	return log, nil
}

func (s *service) GetLogs(ctx context.Context, medicationID string) ([]MedicationLog, error) {
	return s.repo.ListLogs(ctx, medicationID)
}

func (s *service) GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error) {
	return s.repo.GetLastLog(ctx, medicationID)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}
