package notes

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Note, error)
	List(ctx context.Context, filter *NoteFilter) ([]Note, error)
	Create(ctx context.Context, note *Note) error
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, childID, query string) ([]Note, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Note, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) List(ctx context.Context, filter *NoteFilter) ([]Note, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) Create(ctx context.Context, note *Note) error {
	// TODO: implement
	return nil
}

func (r *repository) Update(ctx context.Context, note *Note) error {
	// TODO: implement
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (r *repository) Search(ctx context.Context, childID, query string) ([]Note, error) {
	// TODO: implement full-text search
	return nil, nil
}
