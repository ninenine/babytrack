package auth

import (
	"context"
	"database/sql"
	"errors"
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
	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	var avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	var avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

func (r *repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, name, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	var avatarURL *string
	if user.AvatarURL != "" {
		avatarURL = &user.AvatarURL
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		avatarURL,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *repository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET name = $2, avatar_url = $3, updated_at = $4
		WHERE id = $1
	`

	var avatarURL *string
	if user.AvatarURL != "" {
		avatarURL = &user.AvatarURL
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		avatarURL,
		user.UpdatedAt,
	)

	return err
}
