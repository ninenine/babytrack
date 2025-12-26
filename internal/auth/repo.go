package auth

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) CreateUser(ctx context.Context, user *User) error {
	// TODO: implement
	return nil
}

func (r *repository) UpdateUser(ctx context.Context, user *User) error {
	// TODO: implement
	return nil
}
