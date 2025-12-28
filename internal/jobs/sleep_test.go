package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/ninenine/babytrack/internal/notifications"
	"github.com/ninenine/babytrack/internal/sleep"
)

// mockSleepService is a test double for sleep.Service
type mockSleepService struct {
	sleeps  []sleep.Sleep
	listErr error
}

func newMockSleepService() *mockSleepService {
	return &mockSleepService{}
}

func (m *mockSleepService) Create(ctx context.Context, req *sleep.CreateSleepRequest) (*sleep.Sleep, error) {
	return nil, nil
}

func (m *mockSleepService) Get(ctx context.Context, id string) (*sleep.Sleep, error) {
	return nil, nil
}

func (m *mockSleepService) List(ctx context.Context, filter *sleep.SleepFilter) ([]sleep.Sleep, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.sleeps, nil
}

func (m *mockSleepService) Update(ctx context.Context, id string, req *sleep.CreateSleepRequest) (*sleep.Sleep, error) {
	return nil, nil
}

func (m *mockSleepService) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockSleepService) StartSleep(ctx context.Context, childID string, sleepType sleep.SleepType) (*sleep.Sleep, error) {
	return nil, nil
}

func (m *mockSleepService) EndSleep(ctx context.Context, id string) (*sleep.Sleep, error) {
	return nil, nil
}

func (m *mockSleepService) GetActiveSleep(ctx context.Context, childID string) (*sleep.Sleep, error) {
	return nil, nil
}

func TestNewSleepAnalyticsJob(t *testing.T) {
	sleepSvc := newMockSleepService()

	job := NewSleepAnalyticsJob(sleepSvc)

	if job == nil {
		t.Fatal("NewSleepAnalyticsJob() returned nil")
	}

	if job.sleepService == nil {
		t.Error("NewSleepAnalyticsJob() sleepService should be set")
	}
}

func TestSleepAnalyticsJob_WithNotificationHub(t *testing.T) {
	sleepSvc := newMockSleepService()
	hub := notifications.NewHub()

	job := NewSleepAnalyticsJob(sleepSvc).WithNotificationHub(hub)

	if job.notificationHub == nil {
		t.Error("WithNotificationHub() should set notificationHub")
	}
}

func TestSleepAnalyticsJob_Name(t *testing.T) {
	job := NewSleepAnalyticsJob(nil)

	if job.Name() != "sleep-analytics" {
		t.Errorf("Name() = %v, want sleep-analytics", job.Name())
	}
}

func TestSleepAnalyticsJob_Interval(t *testing.T) {
	job := NewSleepAnalyticsJob(nil)

	expected := 1 * time.Hour
	if job.Interval() != expected {
		t.Errorf("Interval() = %v, want %v", job.Interval(), expected)
	}
}

func TestSleepAnalyticsJob_Run_NoSleeps(t *testing.T) {
	sleepSvc := newMockSleepService()
	job := NewSleepAnalyticsJob(sleepSvc)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestSleepAnalyticsJob_Run_CompletedSleeps(t *testing.T) {
	now := time.Now()
	endTime := now.Add(-1 * time.Hour)
	sleepSvc := newMockSleepService()
	sleepSvc.sleeps = []sleep.Sleep{
		{
			ID:        "sleep-1",
			ChildID:   "child-1",
			Type:      sleep.SleepTypeNap,
			StartTime: now.Add(-3 * time.Hour),
			EndTime:   &endTime,
		},
	}

	job := NewSleepAnalyticsJob(sleepSvc)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestSleepAnalyticsJob_Run_LongNapAlert(t *testing.T) {
	now := time.Now()
	sleepSvc := newMockSleepService()
	sleepSvc.sleeps = []sleep.Sleep{
		{
			ID:        "sleep-1",
			ChildID:   "child-1",
			Type:      sleep.SleepTypeNap,
			StartTime: now.Add(-4 * time.Hour), // 4 hours - exceeds 3 hour threshold
			EndTime:   nil,                     // Still ongoing
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

	job := NewSleepAnalyticsJob(sleepSvc).WithNotificationHub(hub)

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
		t.Error("Expected to receive long nap alert")
	}
}

func TestSleepAnalyticsJob_Run_LongNightSleepAlert(t *testing.T) {
	now := time.Now()
	sleepSvc := newMockSleepService()
	sleepSvc.sleeps = []sleep.Sleep{
		{
			ID:        "sleep-1",
			ChildID:   "child-1",
			Type:      sleep.SleepTypeNight,
			StartTime: now.Add(-15 * time.Hour), // 15 hours - exceeds 14 hour threshold
			EndTime:   nil,
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

	job := NewSleepAnalyticsJob(sleepSvc).WithNotificationHub(hub)

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
		t.Error("Expected to receive long night sleep alert")
	}
}

func TestSleepAnalyticsJob_Run_NormalNap_NoAlert(t *testing.T) {
	now := time.Now()
	sleepSvc := newMockSleepService()
	sleepSvc.sleeps = []sleep.Sleep{
		{
			ID:        "sleep-1",
			ChildID:   "child-1",
			Type:      sleep.SleepTypeNap,
			StartTime: now.Add(-2 * time.Hour), // 2 hours - within threshold
			EndTime:   nil,
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

	job := NewSleepAnalyticsJob(sleepSvc).WithNotificationHub(hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case <-client.Send:
		t.Error("Should not receive alert for normal duration nap")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}

func TestSleepAnalyticsJob_Run_NormalNightSleep_NoAlert(t *testing.T) {
	now := time.Now()
	sleepSvc := newMockSleepService()
	sleepSvc.sleeps = []sleep.Sleep{
		{
			ID:        "sleep-1",
			ChildID:   "child-1",
			Type:      sleep.SleepTypeNight,
			StartTime: now.Add(-10 * time.Hour), // 10 hours - within threshold
			EndTime:   nil,
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

	job := NewSleepAnalyticsJob(sleepSvc).WithNotificationHub(hub)

	err := job.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case <-client.Send:
		t.Error("Should not receive alert for normal duration night sleep")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}

func TestSleepRecommendations(t *testing.T) {
	tests := []struct {
		ageGroup  string
		expectMin float64
		expectMax float64
	}{
		{"newborn", 14, 17},
		{"infant", 12, 15},
		{"toddler", 11, 14},
		{"pre-k", 10, 13},
		{"child", 9, 11},
		{"default", 10, 14},
	}

	for _, tt := range tests {
		t.Run(tt.ageGroup, func(t *testing.T) {
			rec, ok := sleepRecommendations[tt.ageGroup]
			if !ok {
				t.Fatalf("sleepRecommendations missing %s", tt.ageGroup)
			}
			if rec.minHours != tt.expectMin {
				t.Errorf("minHours = %v, want %v", rec.minHours, tt.expectMin)
			}
			if rec.maxHours != tt.expectMax {
				t.Errorf("maxHours = %v, want %v", rec.maxHours, tt.expectMax)
			}
		})
	}
}
