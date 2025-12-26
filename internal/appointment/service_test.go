package appointment

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	appointments map[string]*Appointment
	createErr    error
	updateErr    error
	deleteErr    error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		appointments: make(map[string]*Appointment),
	}
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Appointment, error) {
	apt, ok := m.appointments[id]
	if !ok {
		return nil, nil
	}
	return apt, nil
}

func (m *mockRepository) List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
	var result []Appointment
	for _, apt := range m.appointments {
		if filter.ChildID != "" && apt.ChildID != filter.ChildID {
			continue
		}
		result = append(result, *apt)
	}
	return result, nil
}

func (m *mockRepository) Create(ctx context.Context, apt *Appointment) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.appointments[apt.ID] = apt
	return nil
}

func (m *mockRepository) Update(ctx context.Context, apt *Appointment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.appointments[apt.ID] = apt
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.appointments, id)
	return nil
}

func (m *mockRepository) GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error) {
	var result []Appointment
	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	for _, apt := range m.appointments {
		if apt.ChildID == childID && !apt.Completed && !apt.Cancelled {
			if apt.ScheduledAt.After(now) && apt.ScheduledAt.Before(endDate) {
				result = append(result, *apt)
			}
		}
	}
	return result, nil
}

func TestService_Create(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateAppointmentRequest{
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "6 Month Checkup",
		Provider:    "Dr. Smith",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	}

	apt, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if apt.ID == "" {
		t.Error("Create() should generate an ID")
	}

	if apt.ChildID != req.ChildID {
		t.Errorf("Create() ChildID = %v, want %v", apt.ChildID, req.ChildID)
	}

	if apt.Title != req.Title {
		t.Errorf("Create() Title = %v, want %v", apt.Title, req.Title)
	}

	if apt.Duration != 30 {
		t.Errorf("Create() should default Duration to 30, got %v", apt.Duration)
	}

	if apt.Completed {
		t.Error("Create() should set Completed = false")
	}
}

func TestService_Create_WithDuration(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateAppointmentRequest{
		ChildID:     "child-123",
		Type:        AppointmentTypeSpecialist,
		Title:       "Specialist Visit",
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Duration:    60,
	}

	apt, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if apt.Duration != 60 {
		t.Errorf("Create() Duration = %v, want %v", apt.Duration, 60)
	}
}

func TestService_Create_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateAppointmentRequest{
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "Checkup",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Error("Create() should return error when repo fails")
	}
}

func TestService_Get(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create an appointment first
	req := &CreateAppointmentRequest{
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "Checkup",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.Create(context.Background(), req)

	// Get it back
	apt, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if apt == nil {
		t.Fatal("Get() returned nil for existing appointment")
	}

	if apt.ID != created.ID {
		t.Errorf("Get() ID = %v, want %v", apt.ID, created.ID)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	apt, err := svc.Get(context.Background(), "non-existent-id")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if apt != nil {
		t.Error("Get() should return nil for non-existent appointment")
	}
}

func TestService_Complete(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create an appointment
	req := &CreateAppointmentRequest{
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "Checkup",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.Create(context.Background(), req)

	// Complete it
	err := svc.Complete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	// Verify it's marked as completed
	apt, _ := svc.Get(context.Background(), created.ID)
	if !apt.Completed {
		t.Error("Complete() should set Completed = true")
	}
}

func TestService_Complete_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	err := svc.Complete(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Complete() should return error for non-existent appointment")
	}
}

func TestService_Cancel(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create an appointment
	req := &CreateAppointmentRequest{
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "Checkup",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.Create(context.Background(), req)

	// Cancel it
	err := svc.Cancel(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}

	// Verify it's marked as cancelled
	apt, _ := svc.Get(context.Background(), created.ID)
	if !apt.Cancelled {
		t.Error("Cancel() should set Cancelled = true")
	}
}

func TestService_Cancel_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	err := svc.Cancel(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Cancel() should return error for non-existent appointment")
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create an appointment
	req := &CreateAppointmentRequest{
		ChildID:     "child-123",
		Type:        AppointmentTypeWellVisit,
		Title:       "Checkup",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.Create(context.Background(), req)

	// Delete it
	err := svc.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	apt, _ := svc.Get(context.Background(), created.ID)
	if apt != nil {
		t.Error("Delete() should remove the appointment")
	}
}
