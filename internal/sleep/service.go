package sleep

import (
	"context"
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
	// TODO: implement
	return nil, nil
}

func (s *service) Get(ctx context.Context, id string) (*Sleep, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) StartSleep(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
	// TODO: implement - create sleep with start time, no end time
	return nil, nil
}

func (s *service) EndSleep(ctx context.Context, id string) (*Sleep, error) {
	// TODO: implement - set end time on existing sleep
	return nil, nil
}

func (s *service) GetActiveSleep(ctx context.Context, childID string) (*Sleep, error) {
	return s.repo.GetActiveSleep(ctx, childID)
}
