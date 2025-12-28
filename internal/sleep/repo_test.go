package sleep

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

var sleepColumns = []string{
	"id", "child_id", "type", "start_time", "end_time", "quality", "notes", "created_at", "updated_at", "synced_at",
}

func TestRepository_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(2 * time.Hour)
	quality := 4
	rows := sqlmock.NewRows(sleepColumns).
		AddRow("sleep-123", "child-456", "nap", now, endTime, quality, "Good nap", now, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("sleep-123").
		WillReturnRows(rows)

	s, err := repo.GetByID(context.Background(), "sleep-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if s == nil {
		t.Fatal("GetByID() returned nil")
	}

	if s.ID != "sleep-123" {
		t.Errorf("GetByID() ID = %v, want sleep-123", s.ID)
	}

	if s.ChildID != "child-456" {
		t.Errorf("GetByID() ChildID = %v, want child-456", s.ChildID)
	}

	if s.Type != SleepTypeNap {
		t.Errorf("GetByID() Type = %v, want nap", s.Type)
	}

	if s.EndTime == nil || !s.EndTime.Equal(endTime) {
		t.Errorf("GetByID() EndTime = %v, want %v", s.EndTime, endTime)
	}

	if s.Quality == nil || *s.Quality != quality {
		t.Errorf("GetByID() Quality = %v, want %v", s.Quality, quality)
	}

	if s.Notes != "Good nap" {
		t.Errorf("GetByID() Notes = %v, want Good nap", s.Notes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	s, err := repo.GetByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if s != nil {
		t.Error("GetByID() should return nil for non-existent sleep record")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("sleep-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetByID(context.Background(), "sleep-123")
	if err == nil {
		t.Error("GetByID() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(sleepColumns).
		AddRow("sleep-123", "child-456", "night", now, nil, nil, nil, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("sleep-123").
		WillReturnRows(rows)

	s, err := repo.GetByID(context.Background(), "sleep-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if s.EndTime != nil {
		t.Errorf("GetByID() EndTime should be nil for NULL value, got %v", s.EndTime)
	}

	if s.Quality != nil {
		t.Errorf("GetByID() Quality should be nil for NULL value, got %v", s.Quality)
	}

	if s.Notes != "" {
		t.Errorf("GetByID() Notes should be empty for NULL value, got %v", s.Notes)
	}

	if s.SyncedAt != nil {
		t.Errorf("GetByID() SyncedAt should be nil for NULL value, got %v", s.SyncedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(2 * time.Hour)
	quality := 5
	rows := sqlmock.NewRows(sleepColumns).
		AddRow("sleep-1", "child-456", "nap", now, endTime, quality, "Nap notes", now, now, now).
		AddRow("sleep-2", "child-456", "night", now, endTime, quality, "Night notes", now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &SleepFilter{ChildID: "child-456"}
	sleeps, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sleeps) != 2 {
		t.Errorf("List() returned %d sleep records, want 2", len(sleeps))
	}

	if sleeps[0].ID != "sleep-1" {
		t.Errorf("List() first sleep ID = %v, want sleep-1", sleeps[0].ID)
	}

	if sleeps[1].ID != "sleep-2" {
		t.Errorf("List() second sleep ID = %v, want sleep-2", sleeps[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_WithAllFilters(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now.Add(24 * time.Hour)
	sleepType := SleepTypeNap

	rows := sqlmock.NewRows(sleepColumns).
		AddRow("sleep-1", "child-456", "nap", now, now.Add(time.Hour), 4, "Filtered nap", now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("child-456", startDate, endDate, sleepType).
		WillReturnRows(rows)

	filter := &SleepFilter{
		ChildID:   "child-456",
		StartDate: &startDate,
		EndDate:   &endDate,
		Type:      &sleepType,
	}
	sleeps, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sleeps) != 1 {
		t.Errorf("List() returned %d sleep records, want 1", len(sleeps))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(sleepColumns)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WillReturnRows(rows)

	filter := &SleepFilter{}
	sleeps, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if sleeps == nil {
		t.Error("List() should return empty slice, not nil")
	}

	if len(sleeps) != 0 {
		t.Errorf("List() returned %d sleep records, want 0", len(sleeps))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WillReturnError(errors.New("database error"))

	filter := &SleepFilter{}
	_, err := repo.List(context.Background(), filter)
	if err == nil {
		t.Error("List() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create rows with invalid data type to trigger scan error
	rows := sqlmock.NewRows(sleepColumns).
		AddRow("sleep-1", "child-456", "nap", "invalid-time", nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WillReturnRows(rows)

	filter := &SleepFilter{}
	_, err := repo.List(context.Background(), filter)
	if err == nil {
		t.Error("List() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(sleepColumns).
		AddRow("sleep-1", "child-456", "nap", now, nil, nil, nil, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WillReturnRows(rows)

	filter := &SleepFilter{}
	sleeps, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sleeps) != 1 {
		t.Fatalf("List() returned %d sleep records, want 1", len(sleeps))
	}

	s := sleeps[0]
	if s.EndTime != nil {
		t.Errorf("List() EndTime should be nil for NULL value, got %v", s.EndTime)
	}

	if s.Quality != nil {
		t.Errorf("List() Quality should be nil for NULL value, got %v", s.Quality)
	}

	if s.Notes != "" {
		t.Errorf("List() Notes should be empty for NULL value, got %v", s.Notes)
	}

	if s.SyncedAt != nil {
		t.Errorf("List() SyncedAt should be nil for NULL value, got %v", s.SyncedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(2 * time.Hour)
	quality := 4
	s := &Sleep{
		ID:        "new-sleep",
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: now,
		EndTime:   &endTime,
		Quality:   &quality,
		Notes:     "New nap notes",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO sleep_records").
		WithArgs(s.ID, s.ChildID, s.Type, s.StartTime, s.EndTime, s.Quality, &s.Notes, s.CreatedAt, s.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Create_NoOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	s := &Sleep{
		ID:        "new-sleep",
		ChildID:   "child-123",
		Type:      SleepTypeNight,
		StartTime: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO sleep_records").
		WithArgs(s.ID, s.ChildID, s.Type, s.StartTime, nil, nil, nil, s.CreatedAt, s.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Create_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	s := &Sleep{
		ID:        "error-sleep",
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO sleep_records").
		WithArgs(s.ID, s.ChildID, s.Type, s.StartTime, nil, nil, nil, s.CreatedAt, s.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(context.Background(), s)
	if err == nil {
		t.Error("Create() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(3 * time.Hour)
	quality := 5
	s := &Sleep{
		ID:        "update-sleep",
		Type:      SleepTypeNight,
		StartTime: now,
		EndTime:   &endTime,
		Quality:   &quality,
		Notes:     "Updated notes",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE sleep_records SET type").
		WithArgs(s.ID, s.Type, s.StartTime, s.EndTime, s.Quality, &s.Notes, s.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), s)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update_NoOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	s := &Sleep{
		ID:        "update-sleep",
		Type:      SleepTypeNap,
		StartTime: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE sleep_records SET type").
		WithArgs(s.ID, s.Type, s.StartTime, nil, nil, nil, s.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), s)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	s := &Sleep{
		ID:        "error-update",
		Type:      SleepTypeNap,
		StartTime: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE sleep_records SET type").
		WithArgs(s.ID, s.Type, s.StartTime, nil, nil, nil, s.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.Update(context.Background(), s)
	if err == nil {
		t.Error("Update() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM sleep_records WHERE id").
		WithArgs("delete-sleep").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "delete-sleep")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Delete_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM sleep_records WHERE id").
		WithArgs("error-delete").
		WillReturnError(errors.New("database error"))

	err := repo.Delete(context.Background(), "error-delete")
	if err == nil {
		t.Error("Delete() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetActiveSleep(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(sleepColumns).
		AddRow("active-sleep", "child-456", "nap", now, nil, nil, "Active nap", now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("child-456").
		WillReturnRows(rows)

	s, err := repo.GetActiveSleep(context.Background(), "child-456")
	if err != nil {
		t.Fatalf("GetActiveSleep() error = %v", err)
	}

	if s == nil {
		t.Fatal("GetActiveSleep() returned nil")
	}

	if s.ID != "active-sleep" {
		t.Errorf("GetActiveSleep() ID = %v, want active-sleep", s.ID)
	}

	if s.ChildID != "child-456" {
		t.Errorf("GetActiveSleep() ChildID = %v, want child-456", s.ChildID)
	}

	if s.EndTime != nil {
		t.Errorf("GetActiveSleep() EndTime should be nil for active sleep, got %v", s.EndTime)
	}

	if s.Notes != "Active nap" {
		t.Errorf("GetActiveSleep() Notes = %v, want Active nap", s.Notes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetActiveSleep_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("child-456").
		WillReturnError(sql.ErrNoRows)

	s, err := repo.GetActiveSleep(context.Background(), "child-456")
	if err != nil {
		t.Fatalf("GetActiveSleep() error = %v", err)
	}

	if s != nil {
		t.Error("GetActiveSleep() should return nil when no active sleep exists")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetActiveSleep_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("child-456").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetActiveSleep(context.Background(), "child-456")
	if err == nil {
		t.Error("GetActiveSleep() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetActiveSleep_WithQuality(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	quality := 3
	rows := sqlmock.NewRows(sleepColumns).
		AddRow("active-sleep", "child-456", "night", now, nil, quality, nil, now, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, quality, notes, created_at, updated_at, synced_at").
		WithArgs("child-456").
		WillReturnRows(rows)

	s, err := repo.GetActiveSleep(context.Background(), "child-456")
	if err != nil {
		t.Fatalf("GetActiveSleep() error = %v", err)
	}

	if s == nil {
		t.Fatal("GetActiveSleep() returned nil")
	}

	if s.Quality == nil || *s.Quality != quality {
		t.Errorf("GetActiveSleep() Quality = %v, want %v", s.Quality, quality)
	}

	if s.SyncedAt == nil {
		t.Error("GetActiveSleep() SyncedAt should not be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
