package feeding

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

// GetByID tests

func TestRepository_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(30 * time.Minute)
	amount := 120.0
	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"}).
		AddRow("feeding-123", "child-456", "breast", now, endTime, amount, "ml", "left", "Good feeding", now, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE id = \\$1").
		WithArgs("feeding-123").
		WillReturnRows(rows)

	feeding, err := repo.GetByID(context.Background(), "feeding-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if feeding == nil {
		t.Fatal("GetByID() returned nil feeding")
	}

	if feeding.ID != "feeding-123" {
		t.Errorf("GetByID() ID = %v, want feeding-123", feeding.ID)
	}

	if feeding.ChildID != "child-456" {
		t.Errorf("GetByID() ChildID = %v, want child-456", feeding.ChildID)
	}

	if feeding.Type != FeedingTypeBreast {
		t.Errorf("GetByID() Type = %v, want breast", feeding.Type)
	}

	if feeding.EndTime == nil {
		t.Error("GetByID() EndTime should not be nil")
	}

	if feeding.Amount == nil || *feeding.Amount != 120.0 {
		t.Errorf("GetByID() Amount = %v, want 120.0", feeding.Amount)
	}

	if feeding.Unit != "ml" {
		t.Errorf("GetByID() Unit = %v, want ml", feeding.Unit)
	}

	if feeding.Side != "left" {
		t.Errorf("GetByID() Side = %v, want left", feeding.Side)
	}

	if feeding.Notes != "Good feeding" {
		t.Errorf("GetByID() Notes = %v, want Good feeding", feeding.Notes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE id = \\$1").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	feeding, err := repo.GetByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if feeding != nil {
		t.Error("GetByID() should return nil for non-existent feeding")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE id = \\$1").
		WithArgs("feeding-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetByID(context.Background(), "feeding-123")
	if err == nil {
		t.Error("GetByID() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NullFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"}).
		AddRow("feeding-123", "child-456", "bottle", now, nil, nil, nil, nil, nil, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE id = \\$1").
		WithArgs("feeding-123").
		WillReturnRows(rows)

	feeding, err := repo.GetByID(context.Background(), "feeding-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if feeding.EndTime != nil {
		t.Error("GetByID() EndTime should be nil for NULL")
	}

	if feeding.Amount != nil {
		t.Error("GetByID() Amount should be nil for NULL")
	}

	if feeding.Unit != "" {
		t.Errorf("GetByID() Unit should be empty for NULL, got %v", feeding.Unit)
	}

	if feeding.Side != "" {
		t.Errorf("GetByID() Side should be empty for NULL, got %v", feeding.Side)
	}

	if feeding.Notes != "" {
		t.Errorf("GetByID() Notes should be empty for NULL, got %v", feeding.Notes)
	}

	if feeding.SyncedAt != nil {
		t.Error("GetByID() SyncedAt should be nil for NULL")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// List tests

func TestRepository_List(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(30 * time.Minute)
	amount := 150.0
	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"}).
		AddRow("feeding-1", "child-456", "breast", now, endTime, amount, "ml", "left", "Note 1", now, now, now).
		AddRow("feeding-2", "child-456", "bottle", now, nil, nil, nil, nil, nil, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE 1=1 AND child_id = \\$1 ORDER BY start_time DESC LIMIT 100").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &FeedingFilter{ChildID: "child-456"}
	feedings, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(feedings) != 2 {
		t.Fatalf("List() returned %d feedings, want 2", len(feedings))
	}

	if feedings[0].ID != "feeding-1" {
		t.Errorf("List() first feeding ID = %v, want feeding-1", feedings[0].ID)
	}

	if feedings[1].ID != "feeding-2" {
		t.Errorf("List() second feeding ID = %v, want feeding-2", feedings[1].ID)
	}

	// Check NULL handling for second feeding
	if feedings[1].EndTime != nil {
		t.Error("List() second feeding EndTime should be nil")
	}

	if feedings[1].Amount != nil {
		t.Error("List() second feeding Amount should be nil")
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
	endDate := now
	feedingType := FeedingTypeBreast
	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"}).
		AddRow("feeding-1", "child-456", "breast", now, nil, nil, nil, "left", nil, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE 1=1 AND child_id = \\$1 AND start_time >= \\$2 AND start_time <= \\$3 AND type = \\$4 ORDER BY start_time DESC LIMIT 100").
		WithArgs("child-456", startDate, endDate, feedingType).
		WillReturnRows(rows)

	filter := &FeedingFilter{
		ChildID:   "child-456",
		StartDate: &startDate,
		EndDate:   &endDate,
		Type:      &feedingType,
	}
	feedings, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(feedings) != 1 {
		t.Fatalf("List() returned %d feedings, want 1", len(feedings))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"})

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE 1=1 AND child_id = \\$1 ORDER BY start_time DESC LIMIT 100").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &FeedingFilter{ChildID: "child-456"}
	feedings, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if feedings == nil {
		t.Error("List() should return empty slice, not nil")
	}

	if len(feedings) != 0 {
		t.Errorf("List() returned %d feedings, want 0", len(feedings))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE 1=1 AND child_id = \\$1 ORDER BY start_time DESC LIMIT 100").
		WithArgs("child-456").
		WillReturnError(errors.New("database error"))

	filter := &FeedingFilter{ChildID: "child-456"}
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

	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"}).
		AddRow("feeding-1", "child-456", "breast", "invalid-time", nil, nil, nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE 1=1 AND child_id = \\$1 ORDER BY start_time DESC LIMIT 100").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &FeedingFilter{ChildID: "child-456"}
	_, err := repo.List(context.Background(), filter)
	if err == nil {
		t.Error("List() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// Create tests

func TestRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(30 * time.Minute)
	amount := 120.0
	feeding := &Feeding{
		ID:        "new-feeding-123",
		ChildID:   "child-456",
		Type:      FeedingTypeBreast,
		StartTime: now,
		EndTime:   &endTime,
		Amount:    &amount,
		Unit:      "ml",
		Side:      "left",
		Notes:     "Good feeding",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO feedings").
		WithArgs(feeding.ID, feeding.ChildID, feeding.Type, feeding.StartTime, feeding.EndTime, feeding.Amount, &feeding.Unit, &feeding.Side, &feeding.Notes, feeding.CreatedAt, feeding.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), feeding)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Create_MinimalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	feeding := &Feeding{
		ID:        "new-feeding-456",
		ChildID:   "child-456",
		Type:      FeedingTypeBottle,
		StartTime: now,
		EndTime:   nil,
		Amount:    nil,
		Unit:      "",
		Side:      "",
		Notes:     "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO feedings").
		WithArgs(feeding.ID, feeding.ChildID, feeding.Type, feeding.StartTime, nil, nil, nil, nil, nil, feeding.CreatedAt, feeding.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), feeding)
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
	feeding := &Feeding{
		ID:        "error-feeding",
		ChildID:   "child-456",
		Type:      FeedingTypeBreast,
		StartTime: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO feedings").
		WithArgs(feeding.ID, feeding.ChildID, feeding.Type, feeding.StartTime, nil, nil, nil, nil, nil, feeding.CreatedAt, feeding.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(context.Background(), feeding)
	if err == nil {
		t.Error("Create() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// Update tests

func TestRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(45 * time.Minute)
	amount := 180.0
	feeding := &Feeding{
		ID:        "update-feeding-123",
		Type:      FeedingTypeBottle,
		StartTime: now,
		EndTime:   &endTime,
		Amount:    &amount,
		Unit:      "oz",
		Side:      "both",
		Notes:     "Updated notes",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE feedings SET type = \\$2, start_time = \\$3, end_time = \\$4, amount = \\$5, unit = \\$6, side = \\$7, notes = \\$8, updated_at = \\$9 WHERE id = \\$1").
		WithArgs(feeding.ID, feeding.Type, feeding.StartTime, feeding.EndTime, feeding.Amount, &feeding.Unit, &feeding.Side, &feeding.Notes, feeding.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), feeding)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	feeding := &Feeding{
		ID:        "update-feeding-456",
		Type:      FeedingTypeFormula,
		StartTime: now,
		EndTime:   nil,
		Amount:    nil,
		Unit:      "",
		Side:      "",
		Notes:     "",
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE feedings SET type = \\$2, start_time = \\$3, end_time = \\$4, amount = \\$5, unit = \\$6, side = \\$7, notes = \\$8, updated_at = \\$9 WHERE id = \\$1").
		WithArgs(feeding.ID, feeding.Type, feeding.StartTime, nil, nil, nil, nil, nil, feeding.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), feeding)
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
	feeding := &Feeding{
		ID:        "error-update-feeding",
		Type:      FeedingTypeSolid,
		StartTime: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE feedings SET type = \\$2, start_time = \\$3, end_time = \\$4, amount = \\$5, unit = \\$6, side = \\$7, notes = \\$8, updated_at = \\$9 WHERE id = \\$1").
		WithArgs(feeding.ID, feeding.Type, feeding.StartTime, nil, nil, nil, nil, nil, feeding.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.Update(context.Background(), feeding)
	if err == nil {
		t.Error("Update() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// Delete tests

func TestRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM feedings WHERE id = \\$1").
		WithArgs("delete-feeding-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "delete-feeding-123")
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

	mock.ExpectExec("DELETE FROM feedings WHERE id = \\$1").
		WithArgs("error-delete-feeding").
		WillReturnError(errors.New("database error"))

	err := repo.Delete(context.Background(), "error-delete-feeding")
	if err == nil {
		t.Error("Delete() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// GetLastFeeding tests

func TestRepository_GetLastFeeding(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endTime := now.Add(30 * time.Minute)
	amount := 100.0
	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"}).
		AddRow("last-feeding-123", "child-456", "breast", now, endTime, amount, "ml", "right", "Last feeding notes", now, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE child_id = \\$1 ORDER BY start_time DESC LIMIT 1").
		WithArgs("child-456").
		WillReturnRows(rows)

	feeding, err := repo.GetLastFeeding(context.Background(), "child-456")
	if err != nil {
		t.Fatalf("GetLastFeeding() error = %v", err)
	}

	if feeding == nil {
		t.Fatal("GetLastFeeding() returned nil feeding")
	}

	if feeding.ID != "last-feeding-123" {
		t.Errorf("GetLastFeeding() ID = %v, want last-feeding-123", feeding.ID)
	}

	if feeding.ChildID != "child-456" {
		t.Errorf("GetLastFeeding() ChildID = %v, want child-456", feeding.ChildID)
	}

	if feeding.Side != "right" {
		t.Errorf("GetLastFeeding() Side = %v, want right", feeding.Side)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLastFeeding_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE child_id = \\$1 ORDER BY start_time DESC LIMIT 1").
		WithArgs("child-no-feedings").
		WillReturnError(sql.ErrNoRows)

	feeding, err := repo.GetLastFeeding(context.Background(), "child-no-feedings")
	if err != nil {
		t.Fatalf("GetLastFeeding() error = %v", err)
	}

	if feeding != nil {
		t.Error("GetLastFeeding() should return nil for child with no feedings")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLastFeeding_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE child_id = \\$1 ORDER BY start_time DESC LIMIT 1").
		WithArgs("child-456").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetLastFeeding(context.Background(), "child-456")
	if err == nil {
		t.Error("GetLastFeeding() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLastFeeding_NullFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "child_id", "type", "start_time", "end_time", "amount", "unit", "side", "notes", "created_at", "updated_at", "synced_at"}).
		AddRow("last-feeding-789", "child-456", "formula", now, nil, nil, nil, nil, nil, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, type, start_time, end_time, amount, unit, side, notes, created_at, updated_at, synced_at FROM feedings WHERE child_id = \\$1 ORDER BY start_time DESC LIMIT 1").
		WithArgs("child-456").
		WillReturnRows(rows)

	feeding, err := repo.GetLastFeeding(context.Background(), "child-456")
	if err != nil {
		t.Fatalf("GetLastFeeding() error = %v", err)
	}

	if feeding == nil {
		t.Fatal("GetLastFeeding() returned nil feeding")
	}

	if feeding.EndTime != nil {
		t.Error("GetLastFeeding() EndTime should be nil for NULL")
	}

	if feeding.Amount != nil {
		t.Error("GetLastFeeding() Amount should be nil for NULL")
	}

	if feeding.Unit != "" {
		t.Errorf("GetLastFeeding() Unit should be empty for NULL, got %v", feeding.Unit)
	}

	if feeding.Side != "" {
		t.Errorf("GetLastFeeding() Side should be empty for NULL, got %v", feeding.Side)
	}

	if feeding.Notes != "" {
		t.Errorf("GetLastFeeding() Notes should be empty for NULL, got %v", feeding.Notes)
	}

	if feeding.SyncedAt != nil {
		t.Error("GetLastFeeding() SyncedAt should be nil for NULL")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
