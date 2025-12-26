package notes

import (
	"context"
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
	// TODO: implement
	return nil, nil
}

func (s *service) Get(ctx context.Context, id string) (*Note, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, filter *NoteFilter) ([]Note, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) Update(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Pin(ctx context.Context, id string, pinned bool) error {
	// TODO: implement
	return nil
}

func (s *service) Search(ctx context.Context, childID, query string) ([]Note, error) {
	return s.repo.Search(ctx, childID, query)
}
