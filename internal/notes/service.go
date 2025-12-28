package notes

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Service interface {
	Create(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error)
	Get(ctx context.Context, id string) (*Note, error)
	List(ctx context.Context, filter *NoteFilter) ([]Note, error)
	Update(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error)
	Delete(ctx context.Context, id string) error
	Pin(ctx context.Context, id string, pinned bool) error
	Search(ctx context.Context, childID, query string) ([]Note, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
	now := time.Now()

	note := &Note{
		ID:        generateID(),
		ChildID:   req.ChildID,
		AuthorID:  userID,
		Title:     req.Title,
		Content:   req.Content,
		Tags:      req.Tags,
		Pinned:    req.Pinned,
		CreatedAt: now,
		UpdatedAt: now,
		SyncedAt:  &now,
	}

	if err := s.repo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (s *service) Get(ctx context.Context, id string) (*Note, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *NoteFilter) ([]Note, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, fmt.Errorf("note not found")
	}

	now := time.Now()

	note.Title = req.Title
	note.Content = req.Content
	note.Tags = req.Tags
	note.Pinned = req.Pinned
	note.UpdatedAt = now
	note.SyncedAt = &now

	if err := s.repo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Pin(ctx context.Context, id string, pinned bool) error {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if note == nil {
		return fmt.Errorf("note not found")
	}

	now := time.Now()
	note.Pinned = pinned
	note.UpdatedAt = now
	note.SyncedAt = &now

	if err := s.repo.Update(ctx, note); err != nil {
		return fmt.Errorf("failed to pin note: %w", err)
	}

	return nil
}

func (s *service) Search(ctx context.Context, childID, query string) ([]Note, error) {
	return s.repo.Search(ctx, childID, query)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}
