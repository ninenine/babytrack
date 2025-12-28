package sleep

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Service interface {
	Create(ctx context.Context, req *CreateSleepRequest) (*Sleep, error)
	Get(ctx context.Context, id string) (*Sleep, error)
	List(ctx context.Context, filter *SleepFilter) ([]Sleep, error)
	Update(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error)
	Delete(ctx context.Context, id string) error
	StartSleep(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error)
	EndSleep(ctx context.Context, id string) (*Sleep, error)
	GetActiveSleep(ctx context.Context, childID string) (*Sleep, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
	now := time.Now()

	sleep := &Sleep{
		ID:        generateID(),
		ChildID:   req.ChildID,
		Type:      req.Type,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Quality:   req.Quality,
		Notes:     req.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, sleep); err != nil {
		return nil, fmt.Errorf("failed to create sleep: %w", err)
	}

	return sleep, nil
}

func (s *service) Get(ctx context.Context, id string) (*Sleep, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error) {
	sleep, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sleep == nil {
		return nil, fmt.Errorf("sleep not found")
	}

	sleep.Type = req.Type
	sleep.StartTime = req.StartTime
	sleep.EndTime = req.EndTime
	sleep.Quality = req.Quality
	sleep.Notes = req.Notes
	sleep.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, sleep); err != nil {
		return nil, fmt.Errorf("failed to update sleep: %w", err)
	}

	return sleep, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) StartSleep(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
	now := time.Now()

	sleep := &Sleep{
		ID:        generateID(),
		ChildID:   childID,
		Type:      sleepType,
		StartTime: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, sleep); err != nil {
		return nil, fmt.Errorf("failed to start sleep: %w", err)
	}

	return sleep, nil
}

func (s *service) EndSleep(ctx context.Context, id string) (*Sleep, error) {
	sleep, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sleep == nil {
		return nil, fmt.Errorf("sleep not found")
	}

	now := time.Now()
	sleep.EndTime = &now
	sleep.UpdatedAt = now

	if err := s.repo.Update(ctx, sleep); err != nil {
		return nil, fmt.Errorf("failed to end sleep: %w", err)
	}

	return sleep, nil
}

func (s *service) GetActiveSleep(ctx context.Context, childID string) (*Sleep, error) {
	return s.repo.GetActiveSleep(ctx, childID)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}
