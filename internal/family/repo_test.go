package family

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

// GetFamilyByID tests

func TestRepository_GetFamilyByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
		AddRow("family-123", "Smith Family", now, now)

	mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM families WHERE id = \\$1").
		WithArgs("family-123").
		WillReturnRows(rows)

	family, err := repo.GetFamilyByID(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetFamilyByID() error = %v", err)
	}

	if family == nil {
		t.Fatal("GetFamilyByID() returned nil family")
	}

	if family.ID != "family-123" {
		t.Errorf("GetFamilyByID() ID = %v, want family-123", family.ID)
	}

	if family.Name != "Smith Family" {
		t.Errorf("GetFamilyByID() Name = %v, want Smith Family", family.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM families WHERE id = \\$1").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	family, err := repo.GetFamilyByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetFamilyByID() error = %v", err)
	}

	if family != nil {
		t.Error("GetFamilyByID() should return nil for non-existent family")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM families WHERE id = \\$1").
		WithArgs("family-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetFamilyByID(context.Background(), "family-123")
	if err == nil {
		t.Error("GetFamilyByID() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// CreateFamily tests

func TestRepository_CreateFamily(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	family := &Family{
		ID:        "new-family-123",
		Name:      "Johnson Family",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO families").
		WithArgs(family.ID, family.Name, family.CreatedAt, family.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateFamily(context.Background(), family)
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateFamily_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	family := &Family{
		ID:        "error-family",
		Name:      "Error Family",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO families").
		WithArgs(family.ID, family.Name, family.CreatedAt, family.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.CreateFamily(context.Background(), family)
	if err == nil {
		t.Error("CreateFamily() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// UpdateFamily tests

func TestRepository_UpdateFamily(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	family := &Family{
		ID:        "update-family-123",
		Name:      "Updated Family Name",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE families SET name = \\$2, updated_at = \\$3 WHERE id = \\$1").
		WithArgs(family.ID, family.Name, family.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateFamily(context.Background(), family)
	if err != nil {
		t.Fatalf("UpdateFamily() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_UpdateFamily_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	family := &Family{
		ID:        "error-update-family",
		Name:      "Error Update",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE families SET name = \\$2, updated_at = \\$3 WHERE id = \\$1").
		WithArgs(family.ID, family.Name, family.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.UpdateFamily(context.Background(), family)
	if err == nil {
		t.Error("UpdateFamily() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// GetFamilyMembers tests

func TestRepository_GetFamilyMembers(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "family_id", "user_id", "role", "created_at"}).
		AddRow("member-1", "family-123", "user-1", "admin", now).
		AddRow("member-2", "family-123", "user-2", "member", now)

	mock.ExpectQuery("SELECT id, family_id, user_id, role, created_at FROM family_members WHERE family_id = \\$1").
		WithArgs("family-123").
		WillReturnRows(rows)

	members, err := repo.GetFamilyMembers(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetFamilyMembers() error = %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("GetFamilyMembers() returned %d members, want 2", len(members))
	}

	if members[0].ID != "member-1" {
		t.Errorf("GetFamilyMembers() first member ID = %v, want member-1", members[0].ID)
	}

	if members[0].Role != "admin" {
		t.Errorf("GetFamilyMembers() first member Role = %v, want admin", members[0].Role)
	}

	if members[1].ID != "member-2" {
		t.Errorf("GetFamilyMembers() second member ID = %v, want member-2", members[1].ID)
	}

	if members[1].Role != "member" {
		t.Errorf("GetFamilyMembers() second member Role = %v, want member", members[1].Role)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyMembers_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "family_id", "user_id", "role", "created_at"})

	mock.ExpectQuery("SELECT id, family_id, user_id, role, created_at FROM family_members WHERE family_id = \\$1").
		WithArgs("family-empty").
		WillReturnRows(rows)

	members, err := repo.GetFamilyMembers(context.Background(), "family-empty")
	if err != nil {
		t.Fatalf("GetFamilyMembers() error = %v", err)
	}

	if len(members) != 0 {
		t.Errorf("GetFamilyMembers() returned %d members, want 0", len(members))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyMembers_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, family_id, user_id, role, created_at FROM family_members WHERE family_id = \\$1").
		WithArgs("family-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetFamilyMembers(context.Background(), "family-123")
	if err == nil {
		t.Error("GetFamilyMembers() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyMembers_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "family_id", "user_id", "role", "created_at"}).
		AddRow("member-1", "family-123", "user-1", "admin", "invalid-time")

	mock.ExpectQuery("SELECT id, family_id, user_id, role, created_at FROM family_members WHERE family_id = \\$1").
		WithArgs("family-123").
		WillReturnRows(rows)

	_, err := repo.GetFamilyMembers(context.Background(), "family-123")
	if err == nil {
		t.Error("GetFamilyMembers() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// GetFamilyMembersWithUsers tests

func TestRepository_GetFamilyMembersWithUsers(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "name", "email", "avatar_url", "role", "created_at"}).
		AddRow("member-1", "user-1", "John Smith", "john@example.com", "https://avatar.com/john.jpg", "admin", now).
		AddRow("member-2", "user-2", "Jane Smith", "jane@example.com", "https://avatar.com/jane.jpg", "member", now)

	mock.ExpectQuery("SELECT fm.id, fm.user_id, u.name, u.email, u.avatar_url, fm.role, fm.created_at FROM family_members fm INNER JOIN users u ON fm.user_id = u.id WHERE fm.family_id = \\$1 ORDER BY fm.created_at ASC").
		WithArgs("family-123").
		WillReturnRows(rows)

	members, err := repo.GetFamilyMembersWithUsers(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetFamilyMembersWithUsers() error = %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("GetFamilyMembersWithUsers() returned %d members, want 2", len(members))
	}

	if members[0].Name != "John Smith" {
		t.Errorf("GetFamilyMembersWithUsers() first member Name = %v, want John Smith", members[0].Name)
	}

	if members[0].Email != "john@example.com" {
		t.Errorf("GetFamilyMembersWithUsers() first member Email = %v, want john@example.com", members[0].Email)
	}

	if members[0].AvatarURL != "https://avatar.com/john.jpg" {
		t.Errorf("GetFamilyMembersWithUsers() first member AvatarURL = %v, want https://avatar.com/john.jpg", members[0].AvatarURL)
	}

	if members[1].Name != "Jane Smith" {
		t.Errorf("GetFamilyMembersWithUsers() second member Name = %v, want Jane Smith", members[1].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyMembersWithUsers_NullAvatarURL(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "name", "email", "avatar_url", "role", "created_at"}).
		AddRow("member-1", "user-1", "John Smith", "john@example.com", nil, "admin", now)

	mock.ExpectQuery("SELECT fm.id, fm.user_id, u.name, u.email, u.avatar_url, fm.role, fm.created_at FROM family_members fm INNER JOIN users u ON fm.user_id = u.id WHERE fm.family_id = \\$1 ORDER BY fm.created_at ASC").
		WithArgs("family-123").
		WillReturnRows(rows)

	members, err := repo.GetFamilyMembersWithUsers(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetFamilyMembersWithUsers() error = %v", err)
	}

	if len(members) != 1 {
		t.Fatalf("GetFamilyMembersWithUsers() returned %d members, want 1", len(members))
	}

	if members[0].AvatarURL != "" {
		t.Errorf("GetFamilyMembersWithUsers() AvatarURL should be empty for NULL, got %v", members[0].AvatarURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyMembersWithUsers_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "user_id", "name", "email", "avatar_url", "role", "created_at"})

	mock.ExpectQuery("SELECT fm.id, fm.user_id, u.name, u.email, u.avatar_url, fm.role, fm.created_at FROM family_members fm INNER JOIN users u ON fm.user_id = u.id WHERE fm.family_id = \\$1 ORDER BY fm.created_at ASC").
		WithArgs("family-empty").
		WillReturnRows(rows)

	members, err := repo.GetFamilyMembersWithUsers(context.Background(), "family-empty")
	if err != nil {
		t.Fatalf("GetFamilyMembersWithUsers() error = %v", err)
	}

	if len(members) != 0 {
		t.Errorf("GetFamilyMembersWithUsers() returned %d members, want 0", len(members))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyMembersWithUsers_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT fm.id, fm.user_id, u.name, u.email, u.avatar_url, fm.role, fm.created_at FROM family_members fm INNER JOIN users u ON fm.user_id = u.id WHERE fm.family_id = \\$1 ORDER BY fm.created_at ASC").
		WithArgs("family-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetFamilyMembersWithUsers(context.Background(), "family-123")
	if err == nil {
		t.Error("GetFamilyMembersWithUsers() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetFamilyMembersWithUsers_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "user_id", "name", "email", "avatar_url", "role", "created_at"}).
		AddRow("member-1", "user-1", "John Smith", "john@example.com", nil, "admin", "invalid-time")

	mock.ExpectQuery("SELECT fm.id, fm.user_id, u.name, u.email, u.avatar_url, fm.role, fm.created_at FROM family_members fm INNER JOIN users u ON fm.user_id = u.id WHERE fm.family_id = \\$1 ORDER BY fm.created_at ASC").
		WithArgs("family-123").
		WillReturnRows(rows)

	_, err := repo.GetFamilyMembersWithUsers(context.Background(), "family-123")
	if err == nil {
		t.Error("GetFamilyMembersWithUsers() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// AddFamilyMember tests

func TestRepository_AddFamilyMember(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	member := &FamilyMember{
		ID:        "new-member-123",
		FamilyID:  "family-123",
		UserID:    "user-456",
		Role:      "member",
		CreatedAt: now,
	}

	mock.ExpectExec("INSERT INTO family_members").
		WithArgs(member.ID, member.FamilyID, member.UserID, member.Role, member.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddFamilyMember(context.Background(), member)
	if err != nil {
		t.Fatalf("AddFamilyMember() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_AddFamilyMember_Admin(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	member := &FamilyMember{
		ID:        "admin-member-123",
		FamilyID:  "family-123",
		UserID:    "user-789",
		Role:      "admin",
		CreatedAt: now,
	}

	mock.ExpectExec("INSERT INTO family_members").
		WithArgs(member.ID, member.FamilyID, member.UserID, member.Role, member.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddFamilyMember(context.Background(), member)
	if err != nil {
		t.Fatalf("AddFamilyMember() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_AddFamilyMember_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	member := &FamilyMember{
		ID:        "error-member",
		FamilyID:  "family-123",
		UserID:    "user-456",
		Role:      "member",
		CreatedAt: now,
	}

	mock.ExpectExec("INSERT INTO family_members").
		WithArgs(member.ID, member.FamilyID, member.UserID, member.Role, member.CreatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.AddFamilyMember(context.Background(), member)
	if err == nil {
		t.Error("AddFamilyMember() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// RemoveFamilyMember tests

func TestRepository_RemoveFamilyMember(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM family_members WHERE family_id = \\$1 AND user_id = \\$2").
		WithArgs("family-123", "user-456").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveFamilyMember(context.Background(), "family-123", "user-456")
	if err != nil {
		t.Fatalf("RemoveFamilyMember() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_RemoveFamilyMember_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM family_members WHERE family_id = \\$1 AND user_id = \\$2").
		WithArgs("family-123", "non-existent-user").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.RemoveFamilyMember(context.Background(), "family-123", "non-existent-user")
	if err != nil {
		t.Fatalf("RemoveFamilyMember() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_RemoveFamilyMember_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM family_members WHERE family_id = \\$1 AND user_id = \\$2").
		WithArgs("family-123", "user-456").
		WillReturnError(errors.New("database error"))

	err := repo.RemoveFamilyMember(context.Background(), "family-123", "user-456")
	if err == nil {
		t.Error("RemoveFamilyMember() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// IsMember tests

func TestRepository_IsMember_True(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM family_members WHERE family_id = \\$1 AND user_id = \\$2\\)").
		WithArgs("family-123", "user-456").
		WillReturnRows(rows)

	isMember, err := repo.IsMember(context.Background(), "family-123", "user-456")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}

	if !isMember {
		t.Error("IsMember() should return true for existing member")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_IsMember_False(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM family_members WHERE family_id = \\$1 AND user_id = \\$2\\)").
		WithArgs("family-123", "non-member").
		WillReturnRows(rows)

	isMember, err := repo.IsMember(context.Background(), "family-123", "non-member")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}

	if isMember {
		t.Error("IsMember() should return false for non-member")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_IsMember_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM family_members WHERE family_id = \\$1 AND user_id = \\$2\\)").
		WithArgs("family-123", "user-456").
		WillReturnError(errors.New("database error"))

	_, err := repo.IsMember(context.Background(), "family-123", "user-456")
	if err == nil {
		t.Error("IsMember() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// GetUserFamilies tests

func TestRepository_GetUserFamilies(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
		AddRow("family-1", "Smith Family", now, now).
		AddRow("family-2", "Jones Family", now.Add(-24*time.Hour), now.Add(-24*time.Hour))

	mock.ExpectQuery("SELECT f.id, f.name, f.created_at, f.updated_at FROM families f INNER JOIN family_members fm ON f.id = fm.family_id WHERE fm.user_id = \\$1 ORDER BY f.created_at DESC").
		WithArgs("user-123").
		WillReturnRows(rows)

	families, err := repo.GetUserFamilies(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("GetUserFamilies() error = %v", err)
	}

	if len(families) != 2 {
		t.Fatalf("GetUserFamilies() returned %d families, want 2", len(families))
	}

	if families[0].ID != "family-1" {
		t.Errorf("GetUserFamilies() first family ID = %v, want family-1", families[0].ID)
	}

	if families[0].Name != "Smith Family" {
		t.Errorf("GetUserFamilies() first family Name = %v, want Smith Family", families[0].Name)
	}

	if families[1].ID != "family-2" {
		t.Errorf("GetUserFamilies() second family ID = %v, want family-2", families[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserFamilies_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"})

	mock.ExpectQuery("SELECT f.id, f.name, f.created_at, f.updated_at FROM families f INNER JOIN family_members fm ON f.id = fm.family_id WHERE fm.user_id = \\$1 ORDER BY f.created_at DESC").
		WithArgs("user-no-families").
		WillReturnRows(rows)

	families, err := repo.GetUserFamilies(context.Background(), "user-no-families")
	if err != nil {
		t.Fatalf("GetUserFamilies() error = %v", err)
	}

	if len(families) != 0 {
		t.Errorf("GetUserFamilies() returned %d families, want 0", len(families))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserFamilies_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT f.id, f.name, f.created_at, f.updated_at FROM families f INNER JOIN family_members fm ON f.id = fm.family_id WHERE fm.user_id = \\$1 ORDER BY f.created_at DESC").
		WithArgs("user-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetUserFamilies(context.Background(), "user-123")
	if err == nil {
		t.Error("GetUserFamilies() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUserFamilies_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
		AddRow("family-1", "Smith Family", "invalid-time", "invalid-time")

	mock.ExpectQuery("SELECT f.id, f.name, f.created_at, f.updated_at FROM families f INNER JOIN family_members fm ON f.id = fm.family_id WHERE fm.user_id = \\$1 ORDER BY f.created_at DESC").
		WithArgs("user-123").
		WillReturnRows(rows)

	_, err := repo.GetUserFamilies(context.Background(), "user-123")
	if err == nil {
		t.Error("GetUserFamilies() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// GetChildren tests

func TestRepository_GetChildren(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob1 := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	dob2 := time.Date(2020, 3, 10, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "date_of_birth", "gender", "avatar_url", "created_at", "updated_at"}).
		AddRow("child-1", "family-123", "Emma", dob1, "female", "https://avatar.com/emma.jpg", now, now).
		AddRow("child-2", "family-123", "Liam", dob2, "male", "https://avatar.com/liam.jpg", now, now)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE family_id = \\$1 ORDER BY date_of_birth DESC").
		WithArgs("family-123").
		WillReturnRows(rows)

	children, err := repo.GetChildren(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetChildren() error = %v", err)
	}

	if len(children) != 2 {
		t.Fatalf("GetChildren() returned %d children, want 2", len(children))
	}

	if children[0].ID != "child-1" {
		t.Errorf("GetChildren() first child ID = %v, want child-1", children[0].ID)
	}

	if children[0].Name != "Emma" {
		t.Errorf("GetChildren() first child Name = %v, want Emma", children[0].Name)
	}

	if children[0].Gender != "female" {
		t.Errorf("GetChildren() first child Gender = %v, want female", children[0].Gender)
	}

	if children[0].AvatarURL != "https://avatar.com/emma.jpg" {
		t.Errorf("GetChildren() first child AvatarURL = %v, want https://avatar.com/emma.jpg", children[0].AvatarURL)
	}

	if children[1].ID != "child-2" {
		t.Errorf("GetChildren() second child ID = %v, want child-2", children[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetChildren_NullFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "date_of_birth", "gender", "avatar_url", "created_at", "updated_at"}).
		AddRow("child-1", "family-123", "Alex", dob, nil, nil, now, now)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE family_id = \\$1 ORDER BY date_of_birth DESC").
		WithArgs("family-123").
		WillReturnRows(rows)

	children, err := repo.GetChildren(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetChildren() error = %v", err)
	}

	if len(children) != 1 {
		t.Fatalf("GetChildren() returned %d children, want 1", len(children))
	}

	if children[0].Gender != "" {
		t.Errorf("GetChildren() Gender should be empty for NULL, got %v", children[0].Gender)
	}

	if children[0].AvatarURL != "" {
		t.Errorf("GetChildren() AvatarURL should be empty for NULL, got %v", children[0].AvatarURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetChildren_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "date_of_birth", "gender", "avatar_url", "created_at", "updated_at"})

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE family_id = \\$1 ORDER BY date_of_birth DESC").
		WithArgs("family-no-children").
		WillReturnRows(rows)

	children, err := repo.GetChildren(context.Background(), "family-no-children")
	if err != nil {
		t.Fatalf("GetChildren() error = %v", err)
	}

	if len(children) != 0 {
		t.Errorf("GetChildren() returned %d children, want 0", len(children))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetChildren_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE family_id = \\$1 ORDER BY date_of_birth DESC").
		WithArgs("family-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetChildren(context.Background(), "family-123")
	if err == nil {
		t.Error("GetChildren() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetChildren_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "date_of_birth", "gender", "avatar_url", "created_at", "updated_at"}).
		AddRow("child-1", "family-123", "Emma", "invalid-date", nil, nil, nil, nil)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE family_id = \\$1 ORDER BY date_of_birth DESC").
		WithArgs("family-123").
		WillReturnRows(rows)

	_, err := repo.GetChildren(context.Background(), "family-123")
	if err == nil {
		t.Error("GetChildren() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// GetChildByID tests

func TestRepository_GetChildByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "date_of_birth", "gender", "avatar_url", "created_at", "updated_at"}).
		AddRow("child-123", "family-456", "Emma", dob, "female", "https://avatar.com/emma.jpg", now, now)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE id = \\$1").
		WithArgs("child-123").
		WillReturnRows(rows)

	child, err := repo.GetChildByID(context.Background(), "child-123")
	if err != nil {
		t.Fatalf("GetChildByID() error = %v", err)
	}

	if child == nil {
		t.Fatal("GetChildByID() returned nil child")
	}

	if child.ID != "child-123" {
		t.Errorf("GetChildByID() ID = %v, want child-123", child.ID)
	}

	if child.Name != "Emma" {
		t.Errorf("GetChildByID() Name = %v, want Emma", child.Name)
	}

	if child.FamilyID != "family-456" {
		t.Errorf("GetChildByID() FamilyID = %v, want family-456", child.FamilyID)
	}

	if child.Gender != "female" {
		t.Errorf("GetChildByID() Gender = %v, want female", child.Gender)
	}

	if child.AvatarURL != "https://avatar.com/emma.jpg" {
		t.Errorf("GetChildByID() AvatarURL = %v, want https://avatar.com/emma.jpg", child.AvatarURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetChildByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE id = \\$1").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	child, err := repo.GetChildByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetChildByID() error = %v", err)
	}

	if child != nil {
		t.Error("GetChildByID() should return nil for non-existent child")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetChildByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE id = \\$1").
		WithArgs("child-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetChildByID(context.Background(), "child-123")
	if err == nil {
		t.Error("GetChildByID() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetChildByID_NullFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "family_id", "name", "date_of_birth", "gender", "avatar_url", "created_at", "updated_at"}).
		AddRow("child-123", "family-456", "Alex", dob, nil, nil, now, now)

	mock.ExpectQuery("SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at FROM children WHERE id = \\$1").
		WithArgs("child-123").
		WillReturnRows(rows)

	child, err := repo.GetChildByID(context.Background(), "child-123")
	if err != nil {
		t.Fatalf("GetChildByID() error = %v", err)
	}

	if child == nil {
		t.Fatal("GetChildByID() returned nil child")
	}

	if child.Gender != "" {
		t.Errorf("GetChildByID() Gender should be empty for NULL, got %v", child.Gender)
	}

	if child.AvatarURL != "" {
		t.Errorf("GetChildByID() AvatarURL should be empty for NULL, got %v", child.AvatarURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// CreateChild tests

func TestRepository_CreateChild(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	child := &Child{
		ID:          "new-child-123",
		FamilyID:    "family-456",
		Name:        "Emma",
		DateOfBirth: dob,
		Gender:      "female",
		AvatarURL:   "https://avatar.com/emma.jpg",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gender := child.Gender
	avatarURL := child.AvatarURL

	mock.ExpectExec("INSERT INTO children").
		WithArgs(child.ID, child.FamilyID, child.Name, child.DateOfBirth, &gender, &avatarURL, child.CreatedAt, child.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateChild(context.Background(), child)
	if err != nil {
		t.Fatalf("CreateChild() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateChild_MinimalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	child := &Child{
		ID:          "new-child-456",
		FamilyID:    "family-789",
		Name:        "Alex",
		DateOfBirth: dob,
		Gender:      "",
		AvatarURL:   "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO children").
		WithArgs(child.ID, child.FamilyID, child.Name, child.DateOfBirth, nil, nil, child.CreatedAt, child.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateChild(context.Background(), child)
	if err != nil {
		t.Fatalf("CreateChild() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateChild_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	child := &Child{
		ID:          "error-child",
		FamilyID:    "family-456",
		Name:        "Error Child",
		DateOfBirth: dob,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO children").
		WithArgs(child.ID, child.FamilyID, child.Name, child.DateOfBirth, nil, nil, child.CreatedAt, child.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.CreateChild(context.Background(), child)
	if err == nil {
		t.Error("CreateChild() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// UpdateChild tests

func TestRepository_UpdateChild(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	child := &Child{
		ID:          "update-child-123",
		Name:        "Emma Updated",
		DateOfBirth: dob,
		Gender:      "female",
		AvatarURL:   "https://avatar.com/emma-updated.jpg",
		UpdatedAt:   now,
	}

	gender := child.Gender
	avatarURL := child.AvatarURL

	mock.ExpectExec("UPDATE children SET name = \\$2, date_of_birth = \\$3, gender = \\$4, avatar_url = \\$5, updated_at = \\$6 WHERE id = \\$1").
		WithArgs(child.ID, child.Name, child.DateOfBirth, &gender, &avatarURL, child.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateChild(context.Background(), child)
	if err != nil {
		t.Fatalf("UpdateChild() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_UpdateChild_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	child := &Child{
		ID:          "update-child-456",
		Name:        "Alex Updated",
		DateOfBirth: dob,
		Gender:      "",
		AvatarURL:   "",
		UpdatedAt:   now,
	}

	mock.ExpectExec("UPDATE children SET name = \\$2, date_of_birth = \\$3, gender = \\$4, avatar_url = \\$5, updated_at = \\$6 WHERE id = \\$1").
		WithArgs(child.ID, child.Name, child.DateOfBirth, nil, nil, child.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateChild(context.Background(), child)
	if err != nil {
		t.Fatalf("UpdateChild() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_UpdateChild_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	dob := time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)
	child := &Child{
		ID:          "error-update-child",
		Name:        "Error Update",
		DateOfBirth: dob,
		UpdatedAt:   now,
	}

	mock.ExpectExec("UPDATE children SET name = \\$2, date_of_birth = \\$3, gender = \\$4, avatar_url = \\$5, updated_at = \\$6 WHERE id = \\$1").
		WithArgs(child.ID, child.Name, child.DateOfBirth, nil, nil, child.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.UpdateChild(context.Background(), child)
	if err == nil {
		t.Error("UpdateChild() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// DeleteChild tests

func TestRepository_DeleteChild(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM children WHERE id = \\$1").
		WithArgs("delete-child-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteChild(context.Background(), "delete-child-123")
	if err != nil {
		t.Fatalf("DeleteChild() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_DeleteChild_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM children WHERE id = \\$1").
		WithArgs("non-existent-child").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteChild(context.Background(), "non-existent-child")
	if err != nil {
		t.Fatalf("DeleteChild() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_DeleteChild_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM children WHERE id = \\$1").
		WithArgs("error-delete-child").
		WillReturnError(errors.New("database error"))

	err := repo.DeleteChild(context.Background(), "error-delete-child")
	if err == nil {
		t.Error("DeleteChild() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
