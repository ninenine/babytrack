package appointment

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

var appointmentColumns = []string{
	"id", "child_id", "type", "title", "provider", "location", "scheduled_at",
	"duration", "notes", "completed", "cancelled", "created_at", "updated_at",
}

func TestRepository_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(appointmentColumns).
		AddRow("apt-123", "child-456", "well_visit", "Checkup", "Dr. Smith", "Clinic", now, 30, "Notes", false, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("apt-123").
		WillReturnRows(rows)

	apt, err := repo.GetByID(context.Background(), "apt-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if apt == nil {
		t.Fatal("GetByID() returned nil")
	}

	if apt.ID != "apt-123" {
		t.Errorf("GetByID() ID = %v, want apt-123", apt.ID)
	}

	if apt.Provider != "Dr. Smith" {
		t.Errorf("GetByID() Provider = %v, want Dr. Smith", apt.Provider)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	apt, err := repo.GetByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if apt != nil {
		t.Error("GetByID() should return nil for non-existent appointment")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("apt-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetByID(context.Background(), "apt-123")
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
	rows := sqlmock.NewRows(appointmentColumns).
		AddRow("apt-123", "child-456", "well_visit", "Checkup", nil, nil, now, 30, nil, false, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("apt-123").
		WillReturnRows(rows)

	apt, err := repo.GetByID(context.Background(), "apt-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if apt.Provider != "" || apt.Location != "" || apt.Notes != "" {
		t.Error("GetByID() optional fields should be empty for NULL values")
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
	rows := sqlmock.NewRows(appointmentColumns).
		AddRow("apt-1", "child-456", "well_visit", "Checkup 1", "Dr. A", "Clinic A", now, 30, "Notes 1", false, false, now, now).
		AddRow("apt-2", "child-456", "sick_visit", "Checkup 2", "Dr. B", "Clinic B", now, 45, "Notes 2", true, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &AppointmentFilter{ChildID: "child-456"}
	apts, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(apts) != 2 {
		t.Errorf("List() returned %d appointments, want 2", len(apts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(appointmentColumns)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WillReturnRows(rows)

	filter := &AppointmentFilter{}
	apts, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if apts == nil {
		t.Error("List() should return empty slice, not nil")
	}

	if len(apts) != 0 {
		t.Errorf("List() returned %d appointments, want 0", len(apts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WillReturnError(errors.New("database error"))

	filter := &AppointmentFilter{}
	_, err := repo.List(context.Background(), filter)
	if err == nil {
		t.Error("List() should return error on database failure")
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
	apt := &Appointment{
		ID:          "new-apt",
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "New Checkup",
		Provider:    "Dr. New",
		Location:    "New Clinic",
		ScheduledAt: now,
		Duration:    60,
		Notes:       "New notes",
		Completed:   false,
		Cancelled:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO appointments").
		WithArgs(apt.ID, apt.ChildID, apt.Type, apt.Title, &apt.Provider, &apt.Location, apt.ScheduledAt,
			apt.Duration, &apt.Notes, apt.Completed, apt.Cancelled, apt.CreatedAt, apt.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), apt)
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
	apt := &Appointment{
		ID:          "new-apt",
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "New Checkup",
		ScheduledAt: now,
		Duration:    30,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO appointments").
		WithArgs(apt.ID, apt.ChildID, apt.Type, apt.Title, nil, nil, apt.ScheduledAt,
			apt.Duration, nil, apt.Completed, apt.Cancelled, apt.CreatedAt, apt.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), apt)
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
	apt := &Appointment{
		ID:          "error-apt",
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "Error Checkup",
		ScheduledAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO appointments").
		WithArgs(apt.ID, apt.ChildID, apt.Type, apt.Title, nil, nil, apt.ScheduledAt,
			apt.Duration, nil, apt.Completed, apt.Cancelled, apt.CreatedAt, apt.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(context.Background(), apt)
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
	apt := &Appointment{
		ID:          "update-apt",
		Type:        AppointmentTypeSickVisit,
		Title:       "Updated Checkup",
		Provider:    "Dr. Updated",
		Location:    "Updated Clinic",
		ScheduledAt: now,
		Duration:    45,
		Notes:       "Updated notes",
		Completed:   true,
		Cancelled:   false,
		UpdatedAt:   now,
	}

	mock.ExpectExec("UPDATE appointments SET type").
		WithArgs(apt.ID, apt.Type, apt.Title, &apt.Provider, &apt.Location, apt.ScheduledAt,
			apt.Duration, &apt.Notes, apt.Completed, apt.Cancelled, apt.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), apt)
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
	apt := &Appointment{
		ID:          "error-update",
		Type:        AppointmentTypeWellVisit,
		Title:       "Error Update",
		ScheduledAt: now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("UPDATE appointments SET type").
		WithArgs(apt.ID, apt.Type, apt.Title, nil, nil, apt.ScheduledAt,
			apt.Duration, nil, apt.Completed, apt.Cancelled, apt.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.Update(context.Background(), apt)
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

	mock.ExpectExec("DELETE FROM appointments WHERE id").
		WithArgs("delete-apt").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "delete-apt")
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

	mock.ExpectExec("DELETE FROM appointments WHERE id").
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

func TestRepository_GetUpcoming(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(appointmentColumns).
		AddRow("apt-1", "child-123", "well_visit", "Upcoming 1", "Dr. A", "Clinic", now.Add(24*time.Hour), 30, "Notes", false, false, now, now).
		AddRow("apt-2", "child-123", "dental", "Upcoming 2", "Dr. B", "Dental", now.Add(48*time.Hour), 45, nil, false, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("child-123", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	apts, err := repo.GetUpcoming(context.Background(), "child-123", 7)
	if err != nil {
		t.Fatalf("GetUpcoming() error = %v", err)
	}

	if len(apts) != 2 {
		t.Errorf("GetUpcoming() returned %d appointments, want 2", len(apts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUpcoming_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(appointmentColumns)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("child-123", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	apts, err := repo.GetUpcoming(context.Background(), "child-123", 7)
	if err != nil {
		t.Fatalf("GetUpcoming() error = %v", err)
	}

	if apts == nil {
		t.Error("GetUpcoming() should return empty slice, not nil")
	}

	if len(apts) != 0 {
		t.Errorf("GetUpcoming() returned %d appointments, want 0", len(apts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUpcoming_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, type, title, provider, location, scheduled_at").
		WithArgs("child-123", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err := repo.GetUpcoming(context.Background(), "child-123", 7)
	if err == nil {
		t.Error("GetUpcoming() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
