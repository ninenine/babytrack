package medication

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

var medicationColumns = []string{
	"id", "child_id", "name", "dosage", "unit", "frequency", "instructions",
	"start_date", "end_date", "active", "created_at", "updated_at",
}

var medicationLogColumns = []string{
	"id", "medication_id", "child_id", "given_at", "given_by", "dosage", "notes", "created_at", "synced_at",
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestRepository_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endDate := now.Add(30 * 24 * time.Hour)
	rows := sqlmock.NewRows(medicationColumns).
		AddRow("med-123", "child-456", "Ibuprofen", "200mg", "ml", "daily", "Take with food", now, endDate, true, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WithArgs("med-123").
		WillReturnRows(rows)

	med, err := repo.GetByID(context.Background(), "med-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if med == nil {
		t.Fatal("GetByID() returned nil")
	}

	if med.ID != "med-123" {
		t.Errorf("GetByID() ID = %v, want med-123", med.ID)
	}

	if med.Name != "Ibuprofen" {
		t.Errorf("GetByID() Name = %v, want Ibuprofen", med.Name)
	}

	if med.Instructions != "Take with food" {
		t.Errorf("GetByID() Instructions = %v, want Take with food", med.Instructions)
	}

	if med.EndDate == nil {
		t.Error("GetByID() EndDate should not be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	med, err := repo.GetByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if med != nil {
		t.Error("GetByID() should return nil for non-existent medication")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WithArgs("med-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetByID(context.Background(), "med-123")
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
	rows := sqlmock.NewRows(medicationColumns).
		AddRow("med-123", "child-456", "Ibuprofen", "200mg", "ml", "daily", nil, now, nil, true, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WithArgs("med-123").
		WillReturnRows(rows)

	med, err := repo.GetByID(context.Background(), "med-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if med.Instructions != "" {
		t.Errorf("GetByID() Instructions should be empty for NULL, got %v", med.Instructions)
	}

	if med.EndDate != nil {
		t.Errorf("GetByID() EndDate should be nil for NULL, got %v", med.EndDate)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// List Tests (GetActiveMedications via filter)
// =============================================================================

func TestRepository_List(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endDate := now.Add(30 * 24 * time.Hour)
	rows := sqlmock.NewRows(medicationColumns).
		AddRow("med-1", "child-456", "Ibuprofen", "200mg", "ml", "daily", "Take with food", now, endDate, true, now, now).
		AddRow("med-2", "child-456", "Acetaminophen", "500mg", "tablet", "as_needed", nil, now, nil, true, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &MedicationFilter{ChildID: "child-456"}
	meds, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(meds) != 2 {
		t.Errorf("List() returned %d medications, want 2", len(meds))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_ActiveOnly(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(medicationColumns).
		AddRow("med-1", "child-456", "Ibuprofen", "200mg", "ml", "daily", nil, now, nil, true, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WithArgs("child-456", true).
		WillReturnRows(rows)

	filter := &MedicationFilter{ChildID: "child-456", ActiveOnly: true}
	meds, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(meds) != 1 {
		t.Errorf("List() returned %d medications, want 1", len(meds))
	}

	if !meds[0].Active {
		t.Error("List() with ActiveOnly should return only active medications")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(medicationColumns)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WillReturnRows(rows)

	filter := &MedicationFilter{}
	meds, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if meds == nil {
		t.Error("List() should return empty slice, not nil")
	}

	if len(meds) != 0 {
		t.Errorf("List() returned %d medications, want 0", len(meds))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WillReturnError(errors.New("database error"))

	filter := &MedicationFilter{}
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

	// Create rows with wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "child_id"}).
		AddRow("med-1", "child-456")

	mock.ExpectQuery("SELECT id, child_id, name, dosage, unit, frequency, instructions").
		WillReturnRows(rows)

	filter := &MedicationFilter{}
	_, err := repo.List(context.Background(), filter)
	if err == nil {
		t.Error("List() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Create Tests
// =============================================================================

func TestRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endDate := now.Add(30 * 24 * time.Hour)
	med := &Medication{
		ID:           "new-med",
		ChildID:      "child-123",
		Name:         "Amoxicillin",
		Dosage:       "250mg",
		Unit:         "ml",
		Frequency:    "twice_daily",
		Instructions: "Take with food",
		StartDate:    now,
		EndDate:      &endDate,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mock.ExpectExec("INSERT INTO medications").
		WithArgs(med.ID, med.ChildID, med.Name, med.Dosage, med.Unit, med.Frequency,
			&med.Instructions, med.StartDate, med.EndDate, med.Active, med.CreatedAt, med.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), med)
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
	med := &Medication{
		ID:        "new-med",
		ChildID:   "child-123",
		Name:      "Amoxicillin",
		Dosage:    "250mg",
		Unit:      "ml",
		Frequency: "daily",
		StartDate: now,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO medications").
		WithArgs(med.ID, med.ChildID, med.Name, med.Dosage, med.Unit, med.Frequency,
			nil, med.StartDate, nil, med.Active, med.CreatedAt, med.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), med)
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
	med := &Medication{
		ID:        "error-med",
		ChildID:   "child-123",
		Name:      "Error Med",
		Dosage:    "100mg",
		Unit:      "tablet",
		Frequency: "daily",
		StartDate: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO medications").
		WithArgs(med.ID, med.ChildID, med.Name, med.Dosage, med.Unit, med.Frequency,
			nil, med.StartDate, nil, med.Active, med.CreatedAt, med.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(context.Background(), med)
	if err == nil {
		t.Error("Create() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Update Tests (Deactivate is done via Update)
// =============================================================================

func TestRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	endDate := now.Add(14 * 24 * time.Hour)
	med := &Medication{
		ID:           "update-med",
		Name:         "Updated Med",
		Dosage:       "500mg",
		Unit:         "tablet",
		Frequency:    "twice_daily",
		Instructions: "Updated instructions",
		StartDate:    now,
		EndDate:      &endDate,
		Active:       true,
		UpdatedAt:    now,
	}

	mock.ExpectExec("UPDATE medications SET name").
		WithArgs(med.ID, med.Name, med.Dosage, med.Unit, med.Frequency,
			&med.Instructions, med.StartDate, med.EndDate, med.Active, med.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), med)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update_Deactivate(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	med := &Medication{
		ID:        "deactivate-med",
		Name:      "Deactivated Med",
		Dosage:    "100mg",
		Unit:      "ml",
		Frequency: "daily",
		StartDate: now,
		Active:    false, // Deactivated
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE medications SET name").
		WithArgs(med.ID, med.Name, med.Dosage, med.Unit, med.Frequency,
			nil, med.StartDate, nil, med.Active, med.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), med)
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
	med := &Medication{
		ID:        "error-update",
		Name:      "Error Update",
		Dosage:    "100mg",
		Unit:      "tablet",
		Frequency: "daily",
		StartDate: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE medications SET name").
		WithArgs(med.ID, med.Name, med.Dosage, med.Unit, med.Frequency,
			nil, med.StartDate, nil, med.Active, med.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.Update(context.Background(), med)
	if err == nil {
		t.Error("Update() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM medications WHERE id").
		WithArgs("delete-med").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "delete-med")
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

	mock.ExpectExec("DELETE FROM medications WHERE id").
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

// =============================================================================
// GetLogByID Tests
// =============================================================================

func TestRepository_GetLogByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	rows := sqlmock.NewRows(medicationLogColumns).
		AddRow("log-123", "med-456", "child-789", now, "user-abc", "200mg", "Patient felt better", now, syncedAt)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("log-123").
		WillReturnRows(rows)

	log, err := repo.GetLogByID(context.Background(), "log-123")
	if err != nil {
		t.Fatalf("GetLogByID() error = %v", err)
	}

	if log == nil {
		t.Fatal("GetLogByID() returned nil")
	}

	if log.ID != "log-123" {
		t.Errorf("GetLogByID() ID = %v, want log-123", log.ID)
	}

	if log.MedicationID != "med-456" {
		t.Errorf("GetLogByID() MedicationID = %v, want med-456", log.MedicationID)
	}

	if log.Notes != "Patient felt better" {
		t.Errorf("GetLogByID() Notes = %v, want Patient felt better", log.Notes)
	}

	if log.SyncedAt == nil {
		t.Error("GetLogByID() SyncedAt should not be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLogByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	log, err := repo.GetLogByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetLogByID() error = %v", err)
	}

	if log != nil {
		t.Error("GetLogByID() should return nil for non-existent log")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLogByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("log-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetLogByID(context.Background(), "log-123")
	if err == nil {
		t.Error("GetLogByID() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLogByID_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(medicationLogColumns).
		AddRow("log-123", "med-456", "child-789", now, "user-abc", "200mg", nil, now, nil)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("log-123").
		WillReturnRows(rows)

	log, err := repo.GetLogByID(context.Background(), "log-123")
	if err != nil {
		t.Fatalf("GetLogByID() error = %v", err)
	}

	if log.Notes != "" {
		t.Errorf("GetLogByID() Notes should be empty for NULL, got %v", log.Notes)
	}

	if log.SyncedAt != nil {
		t.Errorf("GetLogByID() SyncedAt should be nil for NULL, got %v", log.SyncedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// ListLogs Tests (GetLogs)
// =============================================================================

func TestRepository_ListLogs(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	rows := sqlmock.NewRows(medicationLogColumns).
		AddRow("log-1", "med-456", "child-789", now, "user-abc", "200mg", "Note 1", now, syncedAt).
		AddRow("log-2", "med-456", "child-789", now.Add(-time.Hour), "user-def", "200mg", nil, now, nil)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-456").
		WillReturnRows(rows)

	logs, err := repo.ListLogs(context.Background(), "med-456")
	if err != nil {
		t.Fatalf("ListLogs() error = %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("ListLogs() returned %d logs, want 2", len(logs))
	}

	if logs[0].Notes != "Note 1" {
		t.Errorf("ListLogs() first log Notes = %v, want Note 1", logs[0].Notes)
	}

	if logs[1].Notes != "" {
		t.Errorf("ListLogs() second log Notes should be empty, got %v", logs[1].Notes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_ListLogs_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(medicationLogColumns)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-456").
		WillReturnRows(rows)

	logs, err := repo.ListLogs(context.Background(), "med-456")
	if err != nil {
		t.Fatalf("ListLogs() error = %v", err)
	}

	if logs == nil {
		t.Error("ListLogs() should return empty slice, not nil")
	}

	if len(logs) != 0 {
		t.Errorf("ListLogs() returned %d logs, want 0", len(logs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_ListLogs_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-456").
		WillReturnError(errors.New("database error"))

	_, err := repo.ListLogs(context.Background(), "med-456")
	if err == nil {
		t.Error("ListLogs() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_ListLogs_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create rows with wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "medication_id"}).
		AddRow("log-1", "med-456")

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-456").
		WillReturnRows(rows)

	_, err := repo.ListLogs(context.Background(), "med-456")
	if err == nil {
		t.Error("ListLogs() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// CreateLog Tests (LogMedication)
// =============================================================================

func TestRepository_CreateLog(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	log := &MedicationLog{
		ID:           "new-log",
		MedicationID: "med-123",
		ChildID:      "child-456",
		GivenAt:      now,
		GivenBy:      "user-789",
		Dosage:       "200mg",
		Notes:        "Given with breakfast",
		CreatedAt:    now,
		SyncedAt:     &syncedAt,
	}

	mock.ExpectExec("INSERT INTO medication_logs").
		WithArgs(log.ID, log.MedicationID, log.ChildID, log.GivenAt, log.GivenBy,
			log.Dosage, &log.Notes, log.CreatedAt, log.SyncedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateLog(context.Background(), log)
	if err != nil {
		t.Fatalf("CreateLog() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateLog_NoOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	log := &MedicationLog{
		ID:           "new-log",
		MedicationID: "med-123",
		ChildID:      "child-456",
		GivenAt:      now,
		GivenBy:      "user-789",
		Dosage:       "200mg",
		CreatedAt:    now,
	}

	mock.ExpectExec("INSERT INTO medication_logs").
		WithArgs(log.ID, log.MedicationID, log.ChildID, log.GivenAt, log.GivenBy,
			log.Dosage, nil, log.CreatedAt, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateLog(context.Background(), log)
	if err != nil {
		t.Fatalf("CreateLog() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_CreateLog_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	log := &MedicationLog{
		ID:           "error-log",
		MedicationID: "med-123",
		ChildID:      "child-456",
		GivenAt:      now,
		GivenBy:      "user-789",
		Dosage:       "200mg",
		CreatedAt:    now,
	}

	mock.ExpectExec("INSERT INTO medication_logs").
		WithArgs(log.ID, log.MedicationID, log.ChildID, log.GivenAt, log.GivenBy,
			log.Dosage, nil, log.CreatedAt, nil).
		WillReturnError(errors.New("duplicate key"))

	err := repo.CreateLog(context.Background(), log)
	if err == nil {
		t.Error("CreateLog() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// GetLastLog Tests
// =============================================================================

func TestRepository_GetLastLog(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	rows := sqlmock.NewRows(medicationLogColumns).
		AddRow("log-123", "med-456", "child-789", now, "user-abc", "200mg", "Latest dose", now, syncedAt)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-456").
		WillReturnRows(rows)

	log, err := repo.GetLastLog(context.Background(), "med-456")
	if err != nil {
		t.Fatalf("GetLastLog() error = %v", err)
	}

	if log == nil {
		t.Fatal("GetLastLog() returned nil")
	}

	if log.ID != "log-123" {
		t.Errorf("GetLastLog() ID = %v, want log-123", log.ID)
	}

	if log.Notes != "Latest dose" {
		t.Errorf("GetLastLog() Notes = %v, want Latest dose", log.Notes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLastLog_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-no-logs").
		WillReturnError(sql.ErrNoRows)

	log, err := repo.GetLastLog(context.Background(), "med-no-logs")
	if err != nil {
		t.Fatalf("GetLastLog() error = %v", err)
	}

	if log != nil {
		t.Error("GetLastLog() should return nil when no logs exist")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLastLog_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-456").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetLastLog(context.Background(), "med-456")
	if err == nil {
		t.Error("GetLastLog() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetLastLog_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(medicationLogColumns).
		AddRow("log-123", "med-456", "child-789", now, "user-abc", "200mg", nil, now, nil)

	mock.ExpectQuery("SELECT id, medication_id, child_id, given_at, given_by, dosage, notes, created_at, synced_at").
		WithArgs("med-456").
		WillReturnRows(rows)

	log, err := repo.GetLastLog(context.Background(), "med-456")
	if err != nil {
		t.Fatalf("GetLastLog() error = %v", err)
	}

	if log.Notes != "" {
		t.Errorf("GetLastLog() Notes should be empty for NULL, got %v", log.Notes)
	}

	if log.SyncedAt != nil {
		t.Errorf("GetLastLog() SyncedAt should be nil for NULL, got %v", log.SyncedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
