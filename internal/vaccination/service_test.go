package vaccination

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	vaccinations map[string]*Vaccination
	schedule     []VaccinationSchedule
	createErr    error
	updateErr    error
	deleteErr    error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		vaccinations: make(map[string]*Vaccination),
		schedule: []VaccinationSchedule{
			{ID: "hep-b-1", Name: "Hepatitis B", Dose: 1, AgeWeeks: 0, AgeLabel: "Birth"},
			{ID: "dtap-1", Name: "DTaP", Dose: 1, AgeWeeks: 8, AgeLabel: "2 months"},
			{ID: "dtap-2", Name: "DTaP", Dose: 2, AgeWeeks: 16, AgeLabel: "4 months"},
		},
	}
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Vaccination, error) {
	vax, ok := m.vaccinations[id]
	if !ok {
		return nil, nil
	}
	return vax, nil
}

func (m *mockRepository) List(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
	var result []Vaccination
	for _, vax := range m.vaccinations {
		if filter.ChildID != "" && vax.ChildID != filter.ChildID {
			continue
		}
		if filter.Completed != nil && vax.Completed != *filter.Completed {
			continue
		}
		result = append(result, *vax)
	}
	return result, nil
}

func (m *mockRepository) Create(ctx context.Context, vax *Vaccination) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.vaccinations[vax.ID] = vax
	return nil
}

func (m *mockRepository) Update(ctx context.Context, vax *Vaccination) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.vaccinations[vax.ID] = vax
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.vaccinations, id)
	return nil
}

func (m *mockRepository) GetUpcoming(ctx context.Context, childID string, days int) ([]Vaccination, error) {
	var result []Vaccination
	now := time.Now()
	cutoff := now.AddDate(0, 0, days)

	for _, vax := range m.vaccinations {
		if vax.ChildID == childID && !vax.Completed {
			if vax.ScheduledAt.After(now) && vax.ScheduledAt.Before(cutoff) {
				result = append(result, *vax)
			}
		}
	}
	return result, nil
}

func (m *mockRepository) GetSchedule() []VaccinationSchedule {
	return m.schedule
}

func TestService_Create(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	scheduledAt := time.Now().AddDate(0, 0, 14)

	req := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Hepatitis B",
		Dose:        1,
		ScheduledAt: scheduledAt,
	}

	vax, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if vax.ID == "" {
		t.Error("Create() should generate an ID")
	}

	if vax.ChildID != req.ChildID {
		t.Errorf("Create() ChildID = %v, want %v", vax.ChildID, req.ChildID)
	}

	if vax.Name != "Hepatitis B" {
		t.Errorf("Create() Name = %v, want Hepatitis B", vax.Name)
	}

	if vax.Dose != 1 {
		t.Errorf("Create() Dose = %v, want 1", vax.Dose)
	}

	if vax.Completed {
		t.Error("Create() should set Completed to false")
	}
}

func TestService_Create_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Test Vax",
		Dose:        1,
		ScheduledAt: time.Now(),
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Error("Create() should return error when repo fails")
	}
}

func TestService_Get(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Test Vax",
		Dose:        1,
		ScheduledAt: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	vax, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if vax == nil {
		t.Fatal("Get() returned nil for existing vaccination")
	}

	if vax.ID != created.ID {
		t.Errorf("Get() ID = %v, want %v", vax.ID, created.ID)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	vax, err := svc.Get(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if vax != nil {
		t.Error("Get() should return nil for non-existent vaccination")
	}
}

func TestService_List(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create multiple vaccinations
	for i := range 3 {
		req := &CreateVaccinationRequest{
			ChildID:     "child-123",
			Name:        "Vax " + string(rune('A'+i)),
			Dose:        i + 1,
			ScheduledAt: time.Now().AddDate(0, 0, i*7),
		}
		svc.Create(context.Background(), req)
	}

	// Create one for different child
	req := &CreateVaccinationRequest{
		ChildID:     "child-456",
		Name:        "Other Vax",
		Dose:        1,
		ScheduledAt: time.Now(),
	}
	svc.Create(context.Background(), req)

	filter := &VaccinationFilter{ChildID: "child-123"}
	vaxs, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(vaxs) != 3 {
		t.Errorf("List() returned %d vaccinations, want 3", len(vaxs))
	}
}

func TestService_List_WithCompletedFilter(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create pending vaccination
	pendingReq := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Pending Vax",
		Dose:        1,
		ScheduledAt: time.Now().AddDate(0, 0, 14),
	}
	svc.Create(context.Background(), pendingReq)

	// Create and complete another
	completedReq := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Completed Vax",
		Dose:        1,
		ScheduledAt: time.Now().AddDate(0, 0, -7),
	}
	completed, _ := svc.Create(context.Background(), completedReq)

	recordReq := &RecordVaccinationRequest{
		AdministeredAt: time.Now(),
		Provider:       "Dr. Smith",
	}
	svc.RecordAdministration(context.Background(), completed.ID, recordReq)

	// Filter by completed only
	isCompleted := true
	filter := &VaccinationFilter{ChildID: "child-123", Completed: &isCompleted}
	vaxs, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(vaxs) != 1 {
		t.Errorf("List() with Completed filter returned %d vaccinations, want 1", len(vaxs))
	}
}

func TestService_Update(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Original Vax",
		Dose:        1,
		ScheduledAt: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	newSchedule := time.Now().AddDate(0, 0, 7)
	updateReq := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Updated Vax",
		Dose:        2,
		ScheduledAt: newSchedule,
	}

	updated, err := svc.Update(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Name != "Updated Vax" {
		t.Errorf("Update() Name = %v, want Updated Vax", updated.Name)
	}

	if updated.Dose != 2 {
		t.Errorf("Update() Dose = %v, want 2", updated.Dose)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Test Vax",
		Dose:        1,
		ScheduledAt: time.Now(),
	}

	_, err := svc.Update(context.Background(), "non-existent", req)
	if err == nil {
		t.Error("Update() should return error for non-existent vaccination")
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Test Vax",
		Dose:        1,
		ScheduledAt: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	err := svc.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	vax, _ := svc.Get(context.Background(), created.ID)
	if vax != nil {
		t.Error("Delete() should remove the vaccination")
	}
}

func TestService_RecordAdministration(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a vaccination
	createReq := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Test Vax",
		Dose:        1,
		ScheduledAt: time.Now().AddDate(0, 0, -1),
	}
	created, _ := svc.Create(context.Background(), createReq)

	adminTime := time.Now()
	recordReq := &RecordVaccinationRequest{
		AdministeredAt: adminTime,
		Provider:       "Dr. Smith",
		Location:       "City Clinic",
		LotNumber:      "LOT123456",
		Notes:          "No adverse reactions",
	}

	recorded, err := svc.RecordAdministration(context.Background(), created.ID, recordReq)
	if err != nil {
		t.Fatalf("RecordAdministration() error = %v", err)
	}

	if !recorded.Completed {
		t.Error("RecordAdministration() should set Completed to true")
	}

	if recorded.AdministeredAt == nil {
		t.Error("RecordAdministration() should set AdministeredAt")
	}

	if recorded.Provider != "Dr. Smith" {
		t.Errorf("RecordAdministration() Provider = %v, want Dr. Smith", recorded.Provider)
	}

	if recorded.Location != "City Clinic" {
		t.Errorf("RecordAdministration() Location = %v, want City Clinic", recorded.Location)
	}

	if recorded.LotNumber != "LOT123456" {
		t.Errorf("RecordAdministration() LotNumber = %v, want LOT123456", recorded.LotNumber)
	}
}

func TestService_RecordAdministration_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	recordReq := &RecordVaccinationRequest{
		AdministeredAt: time.Now(),
		Provider:       "Dr. Smith",
	}

	_, err := svc.RecordAdministration(context.Background(), "non-existent", recordReq)
	if err == nil {
		t.Error("RecordAdministration() should return error for non-existent vaccination")
	}
}

func TestService_GetUpcoming(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	now := time.Now()

	// Create upcoming vaccination (7 days from now)
	upcomingReq := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Upcoming Vax",
		Dose:        1,
		ScheduledAt: now.AddDate(0, 0, 7),
	}
	svc.Create(context.Background(), upcomingReq)

	// Create vaccination too far in future (60 days from now)
	farReq := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Far Future Vax",
		Dose:        1,
		ScheduledAt: now.AddDate(0, 0, 60),
	}
	svc.Create(context.Background(), farReq)

	// Create past vaccination
	pastReq := &CreateVaccinationRequest{
		ChildID:     "child-123",
		Name:        "Past Vax",
		Dose:        1,
		ScheduledAt: now.AddDate(0, 0, -7),
	}
	svc.Create(context.Background(), pastReq)

	// Get upcoming within 30 days
	upcoming, err := svc.GetUpcoming(context.Background(), "child-123", 30)
	if err != nil {
		t.Fatalf("GetUpcoming() error = %v", err)
	}

	if len(upcoming) != 1 {
		t.Errorf("GetUpcoming() returned %d vaccinations, want 1", len(upcoming))
	}
}

func TestService_GetSchedule(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	schedule := svc.GetSchedule()

	if len(schedule) == 0 {
		t.Error("GetSchedule() returned empty schedule")
	}

	// Verify schedule has expected structure
	for _, s := range schedule {
		if s.ID == "" {
			t.Error("GetSchedule() schedule item missing ID")
		}
		if s.Name == "" {
			t.Error("GetSchedule() schedule item missing Name")
		}
	}
}

func TestService_GenerateScheduleForChild(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Birth date 1 month ago
	birthDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02")

	vaccinations, err := svc.GenerateScheduleForChild(context.Background(), "child-123", birthDate)
	if err != nil {
		t.Fatalf("GenerateScheduleForChild() error = %v", err)
	}

	if len(vaccinations) == 0 {
		t.Error("GenerateScheduleForChild() should generate vaccinations")
	}

	// All vaccinations should be for the correct child
	for _, vax := range vaccinations {
		if vax.ChildID != "child-123" {
			t.Errorf("GenerateScheduleForChild() ChildID = %v, want child-123", vax.ChildID)
		}
	}
}

func TestService_GenerateScheduleForChild_InvalidDate(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.GenerateScheduleForChild(context.Background(), "child-123", "invalid-date")
	if err == nil {
		t.Error("GenerateScheduleForChild() should return error for invalid date")
	}
}

func TestService_GenerateScheduleForChild_RFC3339Format(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Use RFC3339 format
	birthDate := time.Now().AddDate(0, -1, 0).Format(time.RFC3339)

	vaccinations, err := svc.GenerateScheduleForChild(context.Background(), "child-123", birthDate)
	if err != nil {
		t.Fatalf("GenerateScheduleForChild() with RFC3339 error = %v", err)
	}

	if len(vaccinations) == 0 {
		t.Error("GenerateScheduleForChild() should work with RFC3339 format")
	}
}
