package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/ninenine/babytrack/internal/medication"
	"github.com/ninenine/babytrack/internal/notifications"
)

// mockMedicationService is a test double for medication.Service
type mockMedicationService struct {
	medications []medication.Medication
	logs        map[string]*medication.MedicationLog
	listErr     error
	logErr      error
}

func newMockMedicationService() *mockMedicationService {
	return &mockMedicationService{
		logs: make(map[string]*medication.MedicationLog),
	}
}

func (m *mockMedicationService) Create(ctx context.Context, req *medication.CreateMedicationRequest) (*medication.Medication, error) {
	return nil, nil
}

func (m *mockMedicationService) Get(ctx context.Context, id string) (*medication.Medication, error) {
	return nil, nil
}

func (m *mockMedicationService) List(ctx context.Context, filter *medication.MedicationFilter) ([]medication.Medication, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if filter.ActiveOnly {
		var active []medication.Medication
		for _, med := range m.medications {
			if med.Active {
				active = append(active, med)
			}
		}
		return active, nil
	}
	return m.medications, nil
}

func (m *mockMedicationService) Update(ctx context.Context, id string, req *medication.CreateMedicationRequest) (*medication.Medication, error) {
	return nil, nil
}

func (m *mockMedicationService) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockMedicationService) Deactivate(ctx context.Context, id string) error {
	return nil
}

func (m *mockMedicationService) LogMedication(ctx context.Context, userID string, req *medication.LogMedicationRequest) (*medication.MedicationLog, error) {
	return nil, nil
}

func (m *mockMedicationService) GetLogs(ctx context.Context, medicationID string) ([]medication.MedicationLog, error) {
	return nil, nil
}

func (m *mockMedicationService) GetLastLog(ctx context.Context, medicationID string) (*medication.MedicationLog, error) {
	if m.logErr != nil {
		return nil, m.logErr
	}
	return m.logs[medicationID], nil
}

func TestNewMedicationReminderJob(t *testing.T) {
	medSvc := newMockMedicationService()
	hub := notifications.NewHub()

	job := NewMedicationReminderJob(medSvc, hub)

	if job == nil {
		t.Fatal("NewMedicationReminderJob() returned nil")
	}

	if job.medicationService == nil {
		t.Error("NewMedicationReminderJob() medicationService should be set")
	}

	if job.notificationHub == nil {
		t.Error("NewMedicationReminderJob() notificationHub should be set")
	}
}

func TestMedicationReminderJob_Name(t *testing.T) {
	job := NewMedicationReminderJob(nil, nil)

	if job.Name() != "medication-reminder" {
		t.Errorf("Name() = %v, want medication-reminder", job.Name())
	}
}

func TestMedicationReminderJob_Interval(t *testing.T) {
	job := NewMedicationReminderJob(nil, nil)

	expected := 15 * time.Minute
	if job.Interval() != expected {
		t.Errorf("Interval() = %v, want %v", job.Interval(), expected)
	}
}

func TestMedicationReminderJob_Run_NoMedications(t *testing.T) {
	medSvc := newMockMedicationService()
	job := NewMedicationReminderJob(medSvc, nil)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestMedicationReminderJob_Run_WithActiveMedications(t *testing.T) {
	medSvc := newMockMedicationService()
	medSvc.medications = []medication.Medication{
		{ID: "med-1", Name: "Medicine A", ChildID: "child-1", Frequency: "once_daily", Active: true},
		{ID: "med-2", Name: "Medicine B", ChildID: "child-1", Frequency: "twice_daily", Active: true},
		{ID: "med-3", Name: "Inactive Med", ChildID: "child-1", Frequency: "once_daily", Active: false},
	}

	job := NewMedicationReminderJob(medSvc, nil)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestMedicationReminderJob_IsMedicationDue_NeverGiven(t *testing.T) {
	job := NewMedicationReminderJob(nil, nil)
	now := time.Now()

	tests := []struct {
		name      string
		frequency string
		expected  bool
	}{
		{"daily medication never given", "once_daily", true},
		{"twice daily never given", "twice_daily", true},
		{"as needed never given", "as_needed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			med := medication.Medication{Frequency: tt.frequency}
			result := job.isMedicationDue(med, nil, now)
			if result != tt.expected {
				t.Errorf("isMedicationDue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMedicationReminderJob_IsMedicationDue_WithLastLog(t *testing.T) {
	job := NewMedicationReminderJob(nil, nil)
	now := time.Now()

	tests := []struct {
		name         string
		frequency    string
		lastGivenAgo time.Duration
		expected     bool
	}{
		{"daily - given 25 hours ago", "once_daily", 25 * time.Hour, true},
		{"daily - given 12 hours ago", "once_daily", 12 * time.Hour, false},
		{"daily - given 23.5 hours ago (grace period)", "once_daily", 23*time.Hour + 30*time.Minute, true},
		{"twice daily - given 13 hours ago", "twice_daily", 13 * time.Hour, true},
		{"twice daily - given 6 hours ago", "twice_daily", 6 * time.Hour, false},
		{"every 4 hours - given 5 hours ago", "every_4_hours", 5 * time.Hour, true},
		{"every 4 hours - given 2 hours ago", "every_4_hours", 2 * time.Hour, false},
		{"every 6 hours - given 7 hours ago", "every_6_hours", 7 * time.Hour, true},
		{"every 8 hours - given 9 hours ago", "every_8_hours", 9 * time.Hour, true},
		{"three times daily - given 9 hours ago", "three_times_daily", 9 * time.Hour, true},
		{"four times daily - given 7 hours ago", "four_times_daily", 7 * time.Hour, true},
		{"as needed - given 48 hours ago", "as_needed", 48 * time.Hour, false},
		{"unknown frequency (defaults to daily)", "unknown", 25 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			med := medication.Medication{Frequency: tt.frequency}
			lastLog := &medication.MedicationLog{
				GivenAt: now.Add(-tt.lastGivenAgo),
			}
			result := job.isMedicationDue(med, lastLog, now)
			if result != tt.expected {
				t.Errorf("isMedicationDue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMedicationReminderJob_Run_WithNotificationHub(t *testing.T) {
	medSvc := newMockMedicationService()
	medSvc.medications = []medication.Medication{
		{ID: "med-1", Name: "Due Medicine", ChildID: "child-1", Frequency: "once_daily", Active: true},
	}
	// No last log means it's due

	hub := notifications.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	// Add a client to receive notifications
	client := &notifications.Client{
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	job := NewMedicationReminderJob(medSvc, hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Check if notification was sent
	select {
	case data := <-client.Send:
		if len(data) == 0 {
			t.Error("Expected notification data")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive notification")
	}
}
