package feeding

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Feeding, error)
	List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error)
	Create(ctx context.Context, feeding *Feeding) error
	Update(ctx context.Context, feeding *Feeding) error
	Delete(ctx context.Context, id string) error
	GetLastFeeding(ctx context.Context, childID string) (*Feeding, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Feeding, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) Create(ctx context.Context, feeding *Feeding) error {
	// TODO: implement
	return nil
}

func (r *repository) Update(ctx context.Context, feeding *Feeding) error {
	// TODO: implement
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (r *repository) GetLastFeeding(ctx context.Context, childID string) (*Feeding, error) {
	// TODO: implement
	return nil, nil
}
