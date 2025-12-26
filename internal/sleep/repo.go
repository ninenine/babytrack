package sleep

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Sleep, error)
	List(ctx context.Context, filter *SleepFilter) ([]Sleep, error)
	Create(ctx context.Context, sleep *Sleep) error
	Update(ctx context.Context, sleep *Sleep) error
	Delete(ctx context.Context, id string) error
	GetActiveSleep(ctx context.Context, childID string) (*Sleep, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Sleep, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) List(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) Create(ctx context.Context, sleep *Sleep) error {
	// TODO: implement
	return nil
}

func (r *repository) Update(ctx context.Context, sleep *Sleep) error {
	// TODO: implement
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

func (r *repository) GetActiveSleep(ctx context.Context, childID string) (*Sleep, error) {
	// TODO: implement - get sleep session with no end time
	return nil, nil
}
