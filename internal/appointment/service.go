package appointment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
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
	now := time.Now()

	duration := req.Duration
	if duration == 0 {
		duration = 30 // Default 30 minutes
	}

	apt := &Appointment{
		ID:          generateID(),
		ChildID:     req.ChildID,
		Type:        req.Type,
		Title:       req.Title,
		Provider:    req.Provider,
		Location:    req.Location,
		ScheduledAt: req.ScheduledAt,
		Duration:    duration,
		Notes:       req.Notes,
		Completed:   false,
		Cancelled:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, apt); err != nil {
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}

	return apt, nil
}

func (s *service) Get(ctx context.Context, id string) (*Appointment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error) {
	apt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if apt == nil {
		return nil, fmt.Errorf("appointment not found")
	}

	apt.Type = req.Type
	apt.Title = req.Title
	apt.Provider = req.Provider
	apt.Location = req.Location
	apt.ScheduledAt = req.ScheduledAt
	apt.Duration = req.Duration
	apt.Notes = req.Notes
	apt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, apt); err != nil {
		return nil, fmt.Errorf("failed to update appointment: %w", err)
	}

	return apt, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Complete(ctx context.Context, id string) error {
	apt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if apt == nil {
		return fmt.Errorf("appointment not found")
	}

	apt.Completed = true
	apt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, apt); err != nil {
		return fmt.Errorf("failed to complete appointment: %w", err)
	}

	return nil
}

func (s *service) Cancel(ctx context.Context, id string) error {
	apt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if apt == nil {
		return fmt.Errorf("appointment not found")
	}

	apt.Cancelled = true
	apt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, apt); err != nil {
		return fmt.Errorf("failed to cancel appointment: %w", err)
	}

	return nil
}

func (s *service) GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error) {
	return s.repo.GetUpcoming(ctx, childID, days)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}
