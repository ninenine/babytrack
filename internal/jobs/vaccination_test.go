package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/ninenine/babytrack/internal/notifications"
	"github.com/ninenine/babytrack/internal/vaccination"
)

// mockVaccinationService is a test double for vaccination.Service
type mockVaccinationService struct {
	upcoming    []vaccination.Vaccination
	upcomingErr error
}

func newMockVaccinationService() *mockVaccinationService {
	return &mockVaccinationService{}
}

func (m *mockVaccinationService) Create(ctx context.Context, req *vaccination.CreateVaccinationRequest) (*vaccination.Vaccination, error) {
	return nil, nil
}

func (m *mockVaccinationService) Get(ctx context.Context, id string) (*vaccination.Vaccination, error) {
	return nil, nil
}

func (m *mockVaccinationService) List(ctx context.Context, filter *vaccination.VaccinationFilter) ([]vaccination.Vaccination, error) {
	return nil, nil
}

func (m *mockVaccinationService) Update(ctx context.Context, id string, req *vaccination.CreateVaccinationRequest) (*vaccination.Vaccination, error) {
	return nil, nil
}

func (m *mockVaccinationService) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockVaccinationService) RecordAdministration(ctx context.Context, id string, req *vaccination.RecordVaccinationRequest) (*vaccination.Vaccination, error) {
	return nil, nil
}

func (m *mockVaccinationService) GetUpcoming(ctx context.Context, childID string, days int) ([]vaccination.Vaccination, error) {
	if m.upcomingErr != nil {
		return nil, m.upcomingErr
	}
	return m.upcoming, nil
}

func (m *mockVaccinationService) GetSchedule() []vaccination.VaccinationSchedule {
	return nil
}

func (m *mockVaccinationService) GenerateScheduleForChild(ctx context.Context, childID string, birthDate string) ([]vaccination.Vaccination, error) {
	return nil, nil
}

func TestNewVaccinationReminderJob(t *testing.T) {
	vaxSvc := newMockVaccinationService()
	hub := notifications.NewHub()

	job := NewVaccinationReminderJob(vaxSvc, hub)

	if job == nil {
		t.Fatal("NewVaccinationReminderJob() returned nil")
	}

	if job.vaccinationService == nil {
		t.Error("NewVaccinationReminderJob() vaccinationService should be set")
	}

	if job.notificationHub == nil {
		t.Error("NewVaccinationReminderJob() notificationHub should be set")
	}
}

func TestVaccinationReminderJob_Name(t *testing.T) {
	job := NewVaccinationReminderJob(nil, nil)

	if job.Name() != "vaccination-reminder" {
		t.Errorf("Name() = %v, want vaccination-reminder", job.Name())
	}
}

func TestVaccinationReminderJob_Interval(t *testing.T) {
	job := NewVaccinationReminderJob(nil, nil)

	expected := 6 * time.Hour
	if job.Interval() != expected {
		t.Errorf("Interval() = %v, want %v", job.Interval(), expected)
	}
}

func TestVaccinationReminderJob_Run_NoUpcoming(t *testing.T) {
	vaxSvc := newMockVaccinationService()
	job := NewVaccinationReminderJob(vaxSvc, nil)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestVaccinationReminderJob_Run_WithUpcoming(t *testing.T) {
	now := time.Now()
	vaxSvc := newMockVaccinationService()
	vaxSvc.upcoming = []vaccination.Vaccination{
		{ID: "vax-1", Name: "DTaP", Dose: 1, ChildID: "child-1", ScheduledAt: now.AddDate(0, 0, 1), Completed: false},
		{ID: "vax-2", Name: "Polio", Dose: 1, ChildID: "child-1", ScheduledAt: now.AddDate(0, 0, 2), Completed: false},
		{ID: "vax-3", Name: "Completed", Dose: 1, ChildID: "child-1", ScheduledAt: now, Completed: true},
	}

	job := NewVaccinationReminderJob(vaxSvc, nil)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestVaccinationReminderJob_Run_DueToday(t *testing.T) {
	now := time.Now()
	vaxSvc := newMockVaccinationService()
	vaxSvc.upcoming = []vaccination.Vaccination{
		{ID: "vax-1", Name: "DTaP", Dose: 1, ChildID: "child-1", ScheduledAt: now, Completed: false},
	}

	hub := notifications.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &notifications.Client{
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	job := NewVaccinationReminderJob(vaxSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should receive notification for vaccination due today
	select {
	case data := <-client.Send:
		if len(data) == 0 {
			t.Error("Expected notification data")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive notification for vaccination due today")
	}
}

func TestVaccinationReminderJob_Run_DueTomorrow(t *testing.T) {
	now := time.Now()
	vaxSvc := newMockVaccinationService()
	vaxSvc.upcoming = []vaccination.Vaccination{
		{ID: "vax-1", Name: "DTaP", Dose: 2, ChildID: "child-1", ScheduledAt: now.AddDate(0, 0, 1), Completed: false},
	}

	hub := notifications.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &notifications.Client{
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	job := NewVaccinationReminderJob(vaxSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case data := <-client.Send:
		if len(data) == 0 {
			t.Error("Expected notification data")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive notification for vaccination due tomorrow")
	}
}

func TestVaccinationReminderJob_Run_DueInFewDays(t *testing.T) {
	now := time.Now()
	vaxSvc := newMockVaccinationService()
	vaxSvc.upcoming = []vaccination.Vaccination{
		{ID: "vax-1", Name: "DTaP", Dose: 3, ChildID: "child-1", ScheduledAt: now.AddDate(0, 0, 3), Completed: false},
	}

	hub := notifications.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &notifications.Client{
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	job := NewVaccinationReminderJob(vaxSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case data := <-client.Send:
		if len(data) == 0 {
			t.Error("Expected notification data")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive notification for vaccination due in 3 days")
	}
}

func TestVaccinationReminderJob_Run_TooFarAway(t *testing.T) {
	now := time.Now()
	vaxSvc := newMockVaccinationService()
	vaxSvc.upcoming = []vaccination.Vaccination{
		{ID: "vax-1", Name: "DTaP", Dose: 1, ChildID: "child-1", ScheduledAt: now.AddDate(0, 0, 5), Completed: false},
	}

	hub := notifications.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &notifications.Client{
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	job := NewVaccinationReminderJob(vaxSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should NOT receive notification (> 3 days away)
	select {
	case <-client.Send:
		t.Error("Should not receive notification for vaccination > 3 days away")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}

func TestVaccinationReminderJob_Run_SkipsCompleted(t *testing.T) {
	now := time.Now()
	vaxSvc := newMockVaccinationService()
	vaxSvc.upcoming = []vaccination.Vaccination{
		{ID: "vax-1", Name: "DTaP", Dose: 1, ChildID: "child-1", ScheduledAt: now, Completed: true},
	}

	hub := notifications.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &notifications.Client{
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	job := NewVaccinationReminderJob(vaxSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should NOT receive notification (completed)
	select {
	case <-client.Send:
		t.Error("Should not receive notification for completed vaccination")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}
