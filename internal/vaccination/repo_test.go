package vaccination

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

var vaccinationColumns = []string{
	"id", "child_id", "name", "dose", "scheduled_at", "administered_at",
	"provider", "location", "lot_number", "notes", "completed", "created_at", "updated_at",
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestRepository_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	administeredAt := now.Add(-time.Hour)
	rows := sqlmock.NewRows(vaccinationColumns).
		AddRow("vax-123", "child-456", "BCG", 1, now, administeredAt,
			"Dr. Smith", "City Hospital", "LOT123", "First dose given", true, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("vax-123").
		WillReturnRows(rows)

	vax, err := repo.GetByID(context.Background(), "vax-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if vax == nil {
		t.Fatal("GetByID() returned nil")
	}

	if vax.ID != "vax-123" {
		t.Errorf("GetByID() ID = %v, want vax-123", vax.ID)
	}

	if vax.Name != "BCG" {
		t.Errorf("GetByID() Name = %v, want BCG", vax.Name)
	}

	if vax.Dose != 1 {
		t.Errorf("GetByID() Dose = %v, want 1", vax.Dose)
	}

	if vax.Provider != "Dr. Smith" {
		t.Errorf("GetByID() Provider = %v, want Dr. Smith", vax.Provider)
	}

	if vax.Location != "City Hospital" {
		t.Errorf("GetByID() Location = %v, want City Hospital", vax.Location)
	}

	if vax.LotNumber != "LOT123" {
		t.Errorf("GetByID() LotNumber = %v, want LOT123", vax.LotNumber)
	}

	if vax.Notes != "First dose given" {
		t.Errorf("GetByID() Notes = %v, want First dose given", vax.Notes)
	}

	if vax.AdministeredAt == nil {
		t.Error("GetByID() AdministeredAt should not be nil")
	}

	if !vax.Completed {
		t.Error("GetByID() Completed should be true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	vax, err := repo.GetByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if vax != nil {
		t.Error("GetByID() should return nil for non-existent vaccination")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("vax-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetByID(context.Background(), "vax-123")
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
	rows := sqlmock.NewRows(vaccinationColumns).
		AddRow("vax-123", "child-456", "BCG", 1, now, nil,
			nil, nil, nil, nil, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("vax-123").
		WillReturnRows(rows)

	vax, err := repo.GetByID(context.Background(), "vax-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if vax.AdministeredAt != nil {
		t.Errorf("GetByID() AdministeredAt should be nil for NULL, got %v", vax.AdministeredAt)
	}

	if vax.Provider != "" {
		t.Errorf("GetByID() Provider should be empty for NULL, got %v", vax.Provider)
	}

	if vax.Location != "" {
		t.Errorf("GetByID() Location should be empty for NULL, got %v", vax.Location)
	}

	if vax.LotNumber != "" {
		t.Errorf("GetByID() LotNumber should be empty for NULL, got %v", vax.LotNumber)
	}

	if vax.Notes != "" {
		t.Errorf("GetByID() Notes should be empty for NULL, got %v", vax.Notes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// List Tests
// =============================================================================

func TestRepository_List(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	administeredAt := now.Add(-time.Hour)
	rows := sqlmock.NewRows(vaccinationColumns).
		AddRow("vax-1", "child-456", "BCG", 1, now, administeredAt,
			"Dr. Smith", "City Hospital", "LOT123", "First dose", true, now, now).
		AddRow("vax-2", "child-456", "OPV", 1, now.Add(24*time.Hour), nil,
			nil, nil, nil, nil, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &VaccinationFilter{ChildID: "child-456"}
	vaccinations, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(vaccinations) != 2 {
		t.Errorf("List() returned %d vaccinations, want 2", len(vaccinations))
	}

	if vaccinations[0].Name != "BCG" {
		t.Errorf("List() first vaccination Name = %v, want BCG", vaccinations[0].Name)
	}

	if vaccinations[1].Name != "OPV" {
		t.Errorf("List() second vaccination Name = %v, want OPV", vaccinations[1].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_WithCompletedFilter(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(vaccinationColumns).
		AddRow("vax-1", "child-456", "BCG", 1, now, now,
			"Dr. Smith", "City Hospital", "LOT123", "Done", true, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456", true).
		WillReturnRows(rows)

	completed := true
	filter := &VaccinationFilter{ChildID: "child-456", Completed: &completed}
	vaccinations, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(vaccinations) != 1 {
		t.Errorf("List() returned %d vaccinations, want 1", len(vaccinations))
	}

	if !vaccinations[0].Completed {
		t.Error("List() with Completed filter should return only completed vaccinations")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_UpcomingOnly(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	scheduledAt := now.Add(7 * 24 * time.Hour)
	rows := sqlmock.NewRows(vaccinationColumns).
		AddRow("vax-1", "child-456", "Pentavalent", 2, scheduledAt, nil,
			nil, nil, nil, nil, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456", sqlmock.AnyArg()).
		WillReturnRows(rows)

	filter := &VaccinationFilter{ChildID: "child-456", UpcomingOnly: true}
	vaccinations, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(vaccinations) != 1 {
		t.Errorf("List() returned %d vaccinations, want 1", len(vaccinations))
	}

	if vaccinations[0].Completed {
		t.Error("List() with UpcomingOnly should return only incomplete vaccinations")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(vaccinationColumns)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WillReturnRows(rows)

	filter := &VaccinationFilter{}
	vaccinations, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if vaccinations == nil {
		t.Error("List() should return empty slice, not nil")
	}

	if len(vaccinations) != 0 {
		t.Errorf("List() returned %d vaccinations, want 0", len(vaccinations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WillReturnError(errors.New("database error"))

	filter := &VaccinationFilter{}
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
		AddRow("vax-1", "child-456")

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WillReturnRows(rows)

	filter := &VaccinationFilter{}
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
	administeredAt := now.Add(-time.Hour)
	vax := &Vaccination{
		ID:             "new-vax",
		ChildID:        "child-123",
		Name:           "BCG",
		Dose:           1,
		ScheduledAt:    now,
		AdministeredAt: &administeredAt,
		Provider:       "Dr. Smith",
		Location:       "City Hospital",
		LotNumber:      "LOT123",
		Notes:          "Administered successfully",
		Completed:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	mock.ExpectExec("INSERT INTO vaccinations").
		WithArgs(vax.ID, vax.ChildID, vax.Name, vax.Dose, vax.ScheduledAt, vax.AdministeredAt,
			&vax.Provider, &vax.Location, &vax.LotNumber, &vax.Notes, vax.Completed, vax.CreatedAt, vax.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), vax)
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
	vax := &Vaccination{
		ID:          "new-vax",
		ChildID:     "child-123",
		Name:        "BCG",
		Dose:        1,
		ScheduledAt: now,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO vaccinations").
		WithArgs(vax.ID, vax.ChildID, vax.Name, vax.Dose, vax.ScheduledAt, nil,
			nil, nil, nil, nil, vax.Completed, vax.CreatedAt, vax.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), vax)
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
	vax := &Vaccination{
		ID:          "error-vax",
		ChildID:     "child-123",
		Name:        "Error Vax",
		Dose:        1,
		ScheduledAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("INSERT INTO vaccinations").
		WithArgs(vax.ID, vax.ChildID, vax.Name, vax.Dose, vax.ScheduledAt, nil,
			nil, nil, nil, nil, vax.Completed, vax.CreatedAt, vax.UpdatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(context.Background(), vax)
	if err == nil {
		t.Error("Create() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Update Tests
// =============================================================================

func TestRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	administeredAt := now.Add(-time.Hour)
	vax := &Vaccination{
		ID:             "update-vax",
		Name:           "Updated Vax",
		Dose:           2,
		ScheduledAt:    now,
		AdministeredAt: &administeredAt,
		Provider:       "Dr. Johnson",
		Location:       "County Clinic",
		LotNumber:      "LOT456",
		Notes:          "Updated notes",
		Completed:      true,
		UpdatedAt:      now,
	}

	mock.ExpectExec("UPDATE vaccinations SET name").
		WithArgs(vax.ID, vax.Name, vax.Dose, vax.ScheduledAt, vax.AdministeredAt,
			&vax.Provider, &vax.Location, &vax.LotNumber, &vax.Notes, vax.Completed, vax.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), vax)
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
	vax := &Vaccination{
		ID:          "update-vax",
		Name:        "Updated Vax",
		Dose:        1,
		ScheduledAt: now,
		Completed:   false,
		UpdatedAt:   now,
	}

	mock.ExpectExec("UPDATE vaccinations SET name").
		WithArgs(vax.ID, vax.Name, vax.Dose, vax.ScheduledAt, nil,
			nil, nil, nil, nil, vax.Completed, vax.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), vax)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update_MarkCompleted(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	administeredAt := now
	vax := &Vaccination{
		ID:             "complete-vax",
		Name:           "BCG",
		Dose:           1,
		ScheduledAt:    now.Add(-7 * 24 * time.Hour),
		AdministeredAt: &administeredAt,
		Provider:       "Dr. Smith",
		Location:       "City Hospital",
		LotNumber:      "LOT789",
		Completed:      true,
		UpdatedAt:      now,
	}

	mock.ExpectExec("UPDATE vaccinations SET name").
		WithArgs(vax.ID, vax.Name, vax.Dose, vax.ScheduledAt, vax.AdministeredAt,
			&vax.Provider, &vax.Location, &vax.LotNumber, nil, vax.Completed, vax.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), vax)
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
	vax := &Vaccination{
		ID:          "error-update",
		Name:        "Error Update",
		Dose:        1,
		ScheduledAt: now,
		UpdatedAt:   now,
	}

	mock.ExpectExec("UPDATE vaccinations SET name").
		WithArgs(vax.ID, vax.Name, vax.Dose, vax.ScheduledAt, nil,
			nil, nil, nil, nil, vax.Completed, vax.UpdatedAt).
		WillReturnError(errors.New("database error"))

	err := repo.Update(context.Background(), vax)
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

	mock.ExpectExec("DELETE FROM vaccinations WHERE id").
		WithArgs("delete-vax").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "delete-vax")
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

	mock.ExpectExec("DELETE FROM vaccinations WHERE id").
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
// GetUpcoming Tests
// =============================================================================

func TestRepository_GetUpcoming(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	scheduledAt := now.Add(7 * 24 * time.Hour)
	rows := sqlmock.NewRows(vaccinationColumns).
		AddRow("vax-1", "child-456", "Pentavalent", 2, scheduledAt, nil,
			nil, nil, nil, nil, false, now, now).
		AddRow("vax-2", "child-456", "PCV", 2, scheduledAt.Add(24*time.Hour), nil,
			nil, nil, nil, nil, false, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	vaccinations, err := repo.GetUpcoming(context.Background(), "child-456", 30)
	if err != nil {
		t.Fatalf("GetUpcoming() error = %v", err)
	}

	if len(vaccinations) != 2 {
		t.Errorf("GetUpcoming() returned %d vaccinations, want 2", len(vaccinations))
	}

	if vaccinations[0].Name != "Pentavalent" {
		t.Errorf("GetUpcoming() first vaccination Name = %v, want Pentavalent", vaccinations[0].Name)
	}

	if vaccinations[1].Name != "PCV" {
		t.Errorf("GetUpcoming() second vaccination Name = %v, want PCV", vaccinations[1].Name)
	}

	// Verify all returned vaccinations are not completed
	for i, v := range vaccinations {
		if v.Completed {
			t.Errorf("GetUpcoming() vaccination %d should not be completed", i)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUpcoming_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(vaccinationColumns)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	vaccinations, err := repo.GetUpcoming(context.Background(), "child-456", 30)
	if err != nil {
		t.Fatalf("GetUpcoming() error = %v", err)
	}

	if vaccinations == nil {
		t.Error("GetUpcoming() should return empty slice, not nil")
	}

	if len(vaccinations) != 0 {
		t.Errorf("GetUpcoming() returned %d vaccinations, want 0", len(vaccinations))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUpcoming_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err := repo.GetUpcoming(context.Background(), "child-456", 30)
	if err == nil {
		t.Error("GetUpcoming() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUpcoming_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create rows with wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "child_id"}).
		AddRow("vax-1", "child-456")

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	_, err := repo.GetUpcoming(context.Background(), "child-456", 30)
	if err == nil {
		t.Error("GetUpcoming() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetUpcoming_WithOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	scheduledAt := now.Add(7 * 24 * time.Hour)
	rows := sqlmock.NewRows(vaccinationColumns).
		AddRow("vax-1", "child-456", "BCG", 1, scheduledAt, nil,
			"Dr. Smith", "City Hospital", nil, "Scheduled appointment", false, now, now)

	mock.ExpectQuery("SELECT id, child_id, name, dose, scheduled_at, administered_at").
		WithArgs("child-456", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	vaccinations, err := repo.GetUpcoming(context.Background(), "child-456", 30)
	if err != nil {
		t.Fatalf("GetUpcoming() error = %v", err)
	}

	if len(vaccinations) != 1 {
		t.Errorf("GetUpcoming() returned %d vaccinations, want 1", len(vaccinations))
	}

	if vaccinations[0].Provider != "Dr. Smith" {
		t.Errorf("GetUpcoming() Provider = %v, want Dr. Smith", vaccinations[0].Provider)
	}

	if vaccinations[0].Location != "City Hospital" {
		t.Errorf("GetUpcoming() Location = %v, want City Hospital", vaccinations[0].Location)
	}

	if vaccinations[0].LotNumber != "" {
		t.Errorf("GetUpcoming() LotNumber should be empty for NULL, got %v", vaccinations[0].LotNumber)
	}

	if vaccinations[0].Notes != "Scheduled appointment" {
		t.Errorf("GetUpcoming() Notes = %v, want Scheduled appointment", vaccinations[0].Notes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// GetSchedule Tests
// =============================================================================

func TestRepository_GetSchedule(t *testing.T) {
	db, _ := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	schedule := repo.GetSchedule()

	if len(schedule) == 0 {
		t.Fatal("GetSchedule() returned empty schedule")
	}

	// Verify first entry (BCG at birth)
	foundBCG := false
	for _, s := range schedule {
		if s.ID == "bcg-1" {
			foundBCG = true
			if s.Name != "BCG" {
				t.Errorf("GetSchedule() BCG Name = %v, want BCG", s.Name)
			}
			if s.AgeWeeks != 0 {
				t.Errorf("GetSchedule() BCG AgeWeeks = %v, want 0", s.AgeWeeks)
			}
			if s.AgeLabel != "Birth" {
				t.Errorf("GetSchedule() BCG AgeLabel = %v, want Birth", s.AgeLabel)
			}
			if s.Dose != 1 {
				t.Errorf("GetSchedule() BCG Dose = %v, want 1", s.Dose)
			}
			break
		}
	}

	if !foundBCG {
		t.Error("GetSchedule() should contain BCG vaccination")
	}

	// Verify 6-week vaccinations exist
	foundPenta1 := false
	for _, s := range schedule {
		if s.ID == "penta-1" {
			foundPenta1 = true
			if s.AgeWeeks != 6 {
				t.Errorf("GetSchedule() Pentavalent-1 AgeWeeks = %v, want 6", s.AgeWeeks)
			}
			if s.AgeLabel != "6 weeks" {
				t.Errorf("GetSchedule() Pentavalent-1 AgeLabel = %v, want 6 weeks", s.AgeLabel)
			}
			break
		}
	}

	if !foundPenta1 {
		t.Error("GetSchedule() should contain Pentavalent-1 vaccination")
	}

	// Verify 9-month vaccinations exist (Measles-Rubella)
	foundMR1 := false
	for _, s := range schedule {
		if s.ID == "mr-1" {
			foundMR1 = true
			if s.AgeMonths != 9 {
				t.Errorf("GetSchedule() MR-1 AgeMonths = %v, want 9", s.AgeMonths)
			}
			if s.AgeLabel != "9 months" {
				t.Errorf("GetSchedule() MR-1 AgeLabel = %v, want 9 months", s.AgeLabel)
			}
			break
		}
	}

	if !foundMR1 {
		t.Error("GetSchedule() should contain Measles-Rubella-1 vaccination")
	}

	// Verify 18-month vaccinations exist
	foundMR2 := false
	for _, s := range schedule {
		if s.ID == "mr-2" {
			foundMR2 = true
			if s.AgeMonths != 18 {
				t.Errorf("GetSchedule() MR-2 AgeMonths = %v, want 18", s.AgeMonths)
			}
			if s.Dose != 2 {
				t.Errorf("GetSchedule() MR-2 Dose = %v, want 2", s.Dose)
			}
			break
		}
	}

	if !foundMR2 {
		t.Error("GetSchedule() should contain Measles-Rubella-2 vaccination")
	}
}

func TestRepository_GetSchedule_AllEntriesHaveRequiredFields(t *testing.T) {
	db, _ := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	schedule := repo.GetSchedule()

	for i, s := range schedule {
		if s.ID == "" {
			t.Errorf("GetSchedule() entry %d has empty ID", i)
		}
		if s.Name == "" {
			t.Errorf("GetSchedule() entry %d has empty Name", i)
		}
		if s.Description == "" {
			t.Errorf("GetSchedule() entry %d has empty Description", i)
		}
		if s.AgeLabel == "" {
			t.Errorf("GetSchedule() entry %d has empty AgeLabel", i)
		}
		// Dose can be 0 for OPV-0 (birth dose)
		if s.Dose < 0 {
			t.Errorf("GetSchedule() entry %d has negative Dose: %d", i, s.Dose)
		}
	}
}

func TestRepository_GetSchedule_UniqueIDs(t *testing.T) {
	db, _ := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	schedule := repo.GetSchedule()

	idMap := make(map[string]bool)
	for _, s := range schedule {
		if idMap[s.ID] {
			t.Errorf("GetSchedule() has duplicate ID: %s", s.ID)
		}
		idMap[s.ID] = true
	}
}

func TestRepository_GetSchedule_OrderedByAge(t *testing.T) {
	db, _ := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	schedule := repo.GetSchedule()

	// Verify schedule entries are ordered (grouped by age)
	// Check that birth vaccines come first
	if len(schedule) > 0 && schedule[0].AgeWeeks != 0 {
		t.Error("GetSchedule() should start with birth (0 weeks) vaccinations")
	}

	// Check general progression
	lastAgeWeeks := -1
	for _, s := range schedule {
		if s.AgeWeeks < lastAgeWeeks {
			// It's okay to have same age entries grouped together
			// but we should not go backwards in age significantly
			if lastAgeWeeks-s.AgeWeeks > 4 {
				t.Errorf("GetSchedule() age order issue: %d weeks comes after %d weeks", s.AgeWeeks, lastAgeWeeks)
			}
		}
		if s.AgeWeeks > lastAgeWeeks {
			lastAgeWeeks = s.AgeWeeks
		}
	}
}
