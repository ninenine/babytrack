package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	return db, mock
}

func TestRepository_GetUserByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "name", "avatar_url", "created_at", "updated_at"}).
		AddRow("user-123", "test@example.com", "Test User", "https://avatar.com/test.jpg", now, now)

	mock.ExpectQuery("SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE id = \\$1").
		WithArgs("user-123").
		WillReturnRows(rows)

	user, err := repo.GetUserByID(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if user == nil {
		t.Fatal("GetUserByID() returned nil user")
	}

	if user.ID != "user-123" {
		t.Errorf("GetUserByID() ID = %v, want user-123", user.ID)
	}

	if user.Email != "test@example.com" {
		t.Errorf("GetUserByID() Email = %v, want test@example.com", user.Email)
	}

	if user.AvatarURL != "https://avatar.com/test.jpg" {
		t.Errorf("GetUserByID() AvatarURL = %v, want https://avatar.com/test.jpg", user.AvatarURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE id = \\$1").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if user != nil {
		t.Error("GetUserByID() should return nil for non-existent user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE id = \\$1").
		WithArgs("user-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetUserByID(context.Background(), "user-123")
	if err == nil {
		t.Error("GetUserByID() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserByID_NullAvatarURL(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "name", "avatar_url", "created_at", "updated_at"}).
		AddRow("user-123", "test@example.com", "Test User", nil, now, now)

	mock.ExpectQuery("SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE id = \\$1").
		WithArgs("user-123").
		WillReturnRows(rows)

	user, err := repo.GetUserByID(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if user.AvatarURL != "" {
		t.Errorf("GetUserByID() AvatarURL should be empty for NULL, got %v", user.AvatarURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserByEmail(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "name", "avatar_url", "created_at", "updated_at"}).
		AddRow("user-456", "email@example.com", "Email User", "https://avatar.com/email.jpg", now, now)

	mock.ExpectQuery("SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE email = \\$1").
		WithArgs("email@example.com").
		WillReturnRows(rows)

	user, err := repo.GetUserByEmail(context.Background(), "email@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}

	if user == nil {
		t.Fatal("GetUserByEmail() returned nil user")
	}

	if user.Email != "email@example.com" {
		t.Errorf("GetUserByEmail() Email = %v, want email@example.com", user.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserByEmail_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE email = \\$1").
		WithArgs("unknown@example.com").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByEmail(context.Background(), "unknown@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}

	if user != nil {
		t.Error("GetUserByEmail() should return nil for non-existent email")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserByEmail_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE email = \\$1").
		WithArgs("test@example.com").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetUserByEmail(context.Background(), "test@example.com")
	if err == nil {
		t.Error("GetUserByEmail() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	user := &User{
		ID:        "new-user-123",
		Email:     "new@example.com",
		Name:      "New User",
		AvatarURL: "https://avatar.com/new.jpg",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Email, user.Name, &user.AvatarURL, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateUser_NoAvatar(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	user := &User{
		ID:        "new-user-456",
		Email:     "noavatar@example.com",
		Name:      "No Avatar User",
		AvatarURL: "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Email, user.Name, nil, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateUser_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	user := &User{
		ID:        "error-user",
		Email:     "error@example.com",
		Name:      "Error User",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Email, user.Name, nil, user.CreatedAt, user.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.CreateUser(context.Background(), user)
	if err == nil {
		t.Error("CreateUser() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_UpdateUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	user := &User{
		ID:        "update-user-123",
		Email:     "update@example.com",
		Name:      "Updated Name",
		AvatarURL: "https://avatar.com/updated.jpg",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE users SET name = \\$2, avatar_url = \\$3, updated_at = \\$4 WHERE id = \\$1").
		WithArgs(user.ID, user.Name, &user.AvatarURL, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_UpdateUser_NoAvatar(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	user := &User{
		ID:        "update-user-456",
		Name:      "Updated Name",
		AvatarURL: "",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE users SET name = \\$2, avatar_url = \\$3, updated_at = \\$4 WHERE id = \\$1").
		WithArgs(user.ID, user.Name, nil, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_UpdateUser_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	user := &User{
		ID:        "error-update-user",
		Name:      "Error Update",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE users SET name = \\$2, avatar_url = \\$3, updated_at = \\$4 WHERE id = \\$1").
		WithArgs(user.ID, user.Name, nil, user.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.UpdateUser(context.Background(), user)
	if err == nil {
		t.Error("UpdateUser() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
