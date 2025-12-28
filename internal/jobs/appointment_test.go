package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/ninenine/babytrack/internal/appointment"
	"github.com/ninenine/babytrack/internal/notifications"
)

// mockAppointmentService is a test double for appointment.Service
type mockAppointmentService struct {
	upcoming    []appointment.Appointment
	upcomingErr error
}

func newMockAppointmentService() *mockAppointmentService {
	return &mockAppointmentService{}
}

func (m *mockAppointmentService) Create(ctx context.Context, req *appointment.CreateAppointmentRequest) (*appointment.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentService) Get(ctx context.Context, id string) (*appointment.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentService) List(ctx context.Context, filter *appointment.AppointmentFilter) ([]appointment.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentService) Update(ctx context.Context, id string, req *appointment.CreateAppointmentRequest) (*appointment.Appointment, error) {
	return nil, nil
}

func (m *mockAppointmentService) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAppointmentService) Complete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAppointmentService) Cancel(ctx context.Context, id string) error {
	return nil
}

func (m *mockAppointmentService) GetUpcoming(ctx context.Context, childID string, days int) ([]appointment.Appointment, error) {
	if m.upcomingErr != nil {
		return nil, m.upcomingErr
	}
	return m.upcoming, nil
}

func TestNewAppointmentReminderJob(t *testing.T) {
	aptSvc := newMockAppointmentService()
	hub := notifications.NewHub()

	job := NewAppointmentReminderJob(aptSvc, hub)

	if job == nil {
		t.Fatal("NewAppointmentReminderJob() returned nil")
	}

	if job.appointmentService == nil {
		t.Error("NewAppointmentReminderJob() appointmentService should be set")
	}

	if job.notificationHub == nil {
		t.Error("NewAppointmentReminderJob() notificationHub should be set")
	}
}

func TestAppointmentReminderJob_Name(t *testing.T) {
	job := NewAppointmentReminderJob(nil, nil)

	if job.Name() != "appointment-reminder" {
		t.Errorf("Name() = %v, want appointment-reminder", job.Name())
	}
}

func TestAppointmentReminderJob_Interval(t *testing.T) {
	job := NewAppointmentReminderJob(nil, nil)

	expected := 30 * time.Minute
	if job.Interval() != expected {
		t.Errorf("Interval() = %v, want %v", job.Interval(), expected)
	}
}

func TestAppointmentReminderJob_Run_NoUpcoming(t *testing.T) {
	aptSvc := newMockAppointmentService()
	job := NewAppointmentReminderJob(aptSvc, nil)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestAppointmentReminderJob_Run_StartingSoon(t *testing.T) {
	now := time.Now()
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Doctor Visit",
			ChildID:     "child-1",
			ScheduledAt: now.Add(30 * time.Minute), // 30 minutes from now
			Completed:   false,
			Cancelled:   false,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

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
		t.Error("Expected to receive notification for appointment starting soon")
	}
}

func TestAppointmentReminderJob_Run_Today(t *testing.T) {
	now := time.Now()
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Checkup",
			ChildID:     "child-1",
			ScheduledAt: now.Add(5 * time.Hour), // 5 hours from now (today)
			Completed:   false,
			Cancelled:   false,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

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
		t.Error("Expected to receive notification for appointment today")
	}
}

func TestAppointmentReminderJob_Run_Tomorrow(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Vaccination",
			ChildID:     "child-1",
			ScheduledAt: tomorrow,
			Completed:   false,
			Cancelled:   false,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

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
		t.Error("Expected to receive notification for appointment tomorrow")
	}
}

func TestAppointmentReminderJob_Run_TooFarAway(t *testing.T) {
	now := time.Now()
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Future Appointment",
			ChildID:     "child-1",
			ScheduledAt: now.Add(30 * time.Hour), // More than 24 hours away
			Completed:   false,
			Cancelled:   false,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case <-client.Send:
		t.Error("Should not receive notification for appointment > 24 hours away")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}

func TestAppointmentReminderJob_Run_SkipsCompleted(t *testing.T) {
	now := time.Now()
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Completed Appointment",
			ChildID:     "child-1",
			ScheduledAt: now.Add(30 * time.Minute),
			Completed:   true,
			Cancelled:   false,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case <-client.Send:
		t.Error("Should not receive notification for completed appointment")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}

func TestAppointmentReminderJob_Run_SkipsCancelled(t *testing.T) {
	now := time.Now()
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Canceled Appointment",
			ChildID:     "child-1",
			ScheduledAt: now.Add(30 * time.Minute),
			Completed:   false,
			Cancelled:   true,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case <-client.Send:
		t.Error("Should not receive notification for canceled appointment")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}

func TestAppointmentReminderJob_Run_MultipleAppointments(t *testing.T) {
	now := time.Now()
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Doctor",
			ChildID:     "child-1",
			ScheduledAt: now.Add(30 * time.Minute),
			Completed:   false,
			Cancelled:   false,
		},
		{
			ID:          "apt-2",
			Title:       "Dentist",
			ChildID:     "child-1",
			ScheduledAt: now.Add(2 * time.Hour),
			Completed:   false,
			Cancelled:   false,
		},
		{
			ID:          "apt-3",
			Title:       "Completed",
			ChildID:     "child-1",
			ScheduledAt: now.Add(1 * time.Hour),
			Completed:   true,
			Cancelled:   false,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should receive 2 notifications (not the completed one)
	count := 0
loop:
	for range 2 {
		select {
		case <-client.Send:
			count++
		case <-time.After(100 * time.Millisecond):
			break loop
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 notifications, got %d", count)
	}
}

func TestAppointmentReminderJob_Run_PastAppointment(t *testing.T) {
	now := time.Now()
	aptSvc := newMockAppointmentService()
	aptSvc.upcoming = []appointment.Appointment{
		{
			ID:          "apt-1",
			Title:       "Past Appointment",
			ChildID:     "child-1",
			ScheduledAt: now.Add(-1 * time.Hour), // In the past
			Completed:   false,
			Cancelled:   false,
		},
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

	job := NewAppointmentReminderJob(aptSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should not receive notification for past appointment (hoursUntil <= 0)
	select {
	case <-client.Send:
		t.Error("Should not receive notification for past appointment")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}
