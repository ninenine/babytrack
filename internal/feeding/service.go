package feeding

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
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
	now := time.Now()

	feeding := &Feeding{
		ID:        generateID(),
		ChildID:   req.ChildID,
		Type:      req.Type,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Amount:    req.Amount,
		Unit:      req.Unit,
		Side:      req.Side,
		Notes:     req.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, feeding); err != nil {
		return nil, fmt.Errorf("failed to create feeding: %w", err)
	}

	return feeding, nil
}

func (s *service) Get(ctx context.Context, id string) (*Feeding, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error) {
	feeding, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if feeding == nil {
		return nil, fmt.Errorf("feeding not found")
	}

	feeding.Type = req.Type
	feeding.StartTime = req.StartTime
	feeding.EndTime = req.EndTime
	feeding.Amount = req.Amount
	feeding.Unit = req.Unit
	feeding.Side = req.Side
	feeding.Notes = req.Notes
	feeding.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, feeding); err != nil {
		return nil, fmt.Errorf("failed to update feeding: %w", err)
	}

	return feeding, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetLastFeeding(ctx context.Context, childID string) (*Feeding, error) {
	return s.repo.GetLastFeeding(ctx, childID)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}
