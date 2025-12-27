package vaccination

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
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
	now := time.Now()

	vax := &Vaccination{
		ID:          generateID(),
		ChildID:     req.ChildID,
		Name:        req.Name,
		Dose:        req.Dose,
		ScheduledAt: req.ScheduledAt,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, vax); err != nil {
		return nil, fmt.Errorf("failed to create vaccination: %w", err)
	}

	return vax, nil
}

func (s *service) Get(ctx context.Context, id string) (*Vaccination, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error) {
	vax, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vax == nil {
		return nil, fmt.Errorf("vaccination not found")
	}

	vax.Name = req.Name
	vax.Dose = req.Dose
	vax.ScheduledAt = req.ScheduledAt
	vax.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, vax); err != nil {
		return nil, fmt.Errorf("failed to update vaccination: %w", err)
	}

	return vax, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) RecordAdministration(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
	vax, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vax == nil {
		return nil, fmt.Errorf("vaccination not found")
	}

	vax.AdministeredAt = &req.AdministeredAt
	vax.Provider = req.Provider
	vax.Location = req.Location
	vax.LotNumber = req.LotNumber
	vax.Notes = req.Notes
	vax.Completed = true
	vax.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, vax); err != nil {
		return nil, fmt.Errorf("failed to record vaccination: %w", err)
	}

	return vax, nil
}

func (s *service) GetUpcoming(ctx context.Context, childID string, days int) ([]Vaccination, error) {
	return s.repo.GetUpcoming(ctx, childID, days)
}

func (s *service) GetSchedule() []VaccinationSchedule {
	return s.repo.GetSchedule()
}

func (s *service) GenerateScheduleForChild(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
	// Parse birth date (try multiple formats)
	var birth time.Time
	var err error

	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
	}

	for _, format := range formats {
		birth, err = time.Parse(format, birthDate)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("invalid birth date format: %w", err)
	}

	schedule := s.repo.GetSchedule()
	now := time.Now()
	var vaccinations []Vaccination

	for _, sched := range schedule {
		// Calculate scheduled date based on age in weeks (more accurate for infant schedule)
		scheduledAt := birth.AddDate(0, 0, sched.AgeWeeks*7)

		// Only create future vaccinations or ones due in the past 30 days
		if scheduledAt.After(now.AddDate(0, 0, -30)) {
			vax := &Vaccination{
				ID:          generateID(),
				ChildID:     childID,
				Name:        sched.Name,
				Dose:        sched.Dose,
				ScheduledAt: scheduledAt,
				Completed:   false,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := s.repo.Create(ctx, vax); err != nil {
				return nil, fmt.Errorf("failed to create vaccination %s: %w", sched.Name, err)
			}

			vaccinations = append(vaccinations, *vax)
		}
	}

	return vaccinations, nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
