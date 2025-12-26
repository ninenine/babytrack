package feeding

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error)
	Get(ctx context.Context, id string) (*Feeding, error)
	List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error)
	Update(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error)
	Delete(ctx context.Context, id string) error
	GetLastFeeding(ctx context.Context, childID string) (*Feeding, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Get(ctx context.Context, id string) (*Feeding, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetLastFeeding(ctx context.Context, childID string) (*Feeding, error) {
	return s.repo.GetLastFeeding(ctx, childID)
}
