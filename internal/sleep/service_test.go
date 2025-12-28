package sleep

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	sleeps    map[string]*Sleep
	createErr error
	updateErr error
	deleteErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		sleeps: make(map[string]*Sleep),
	}
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Sleep, error) {
	s, ok := m.sleeps[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockRepository) List(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
	var result []Sleep
	for _, s := range m.sleeps {
		if filter.ChildID != "" && s.ChildID != filter.ChildID {
			continue
		}
		if filter.Type != nil && s.Type != *filter.Type {
			continue
		}
		result = append(result, *s)
	}
	return result, nil
}

func (m *mockRepository) Create(ctx context.Context, sleep *Sleep) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.sleeps[sleep.ID] = sleep
	return nil
}

func (m *mockRepository) Update(ctx context.Context, sleep *Sleep) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.sleeps[sleep.ID] = sleep
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.sleeps, id)
	return nil
}

func (m *mockRepository) GetActiveSleep(ctx context.Context, childID string) (*Sleep, error) {
	for _, s := range m.sleeps {
		if s.ChildID == childID && s.EndTime == nil {
			return s, nil
		}
	}
	return nil, nil
}

func TestService_Create(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	startTime := time.Now()
	endTime := startTime.Add(2 * time.Hour)
	quality := 4

	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: startTime,
		EndTime:   &endTime,
		Quality:   &quality,
		Notes:     "Good nap",
	}

	sleep, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if sleep.ID == "" {
		t.Error("Create() should generate an ID")
	}

	if sleep.ChildID != req.ChildID {
		t.Errorf("Create() ChildID = %v, want %v", sleep.ChildID, req.ChildID)
	}

	if sleep.Type != SleepTypeNap {
		t.Errorf("Create() Type = %v, want %v", sleep.Type, SleepTypeNap)
	}

	if sleep.Quality == nil || *sleep.Quality != quality {
		t.Errorf("Create() Quality = %v, want %v", sleep.Quality, quality)
	}
}

func TestService_Create_NightSleep(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNight,
		StartTime: time.Now(),
	}

	sleep, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if sleep.Type != SleepTypeNight {
		t.Errorf("Create() Type = %v, want %v", sleep.Type, SleepTypeNight)
	}
}

func TestService_Create_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: time.Now(),
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Error("Create() should return error when repo fails")
	}
}

func TestService_Get(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	sleep, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if sleep == nil {
		t.Fatal("Get() returned nil for existing sleep")
	}

	if sleep.ID != created.ID {
		t.Errorf("Get() ID = %v, want %v", sleep.ID, created.ID)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	sleep, err := svc.Get(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if sleep != nil {
		t.Error("Get() should return nil for non-existent sleep")
	}
}

func TestService_List(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create multiple sleeps
	for i := range 3 {
		req := &CreateSleepRequest{
			ChildID:   "child-123",
			Type:      SleepTypeNap,
			StartTime: time.Now().Add(time.Duration(i) * time.Hour),
		}
		svc.Create(context.Background(), req)
	}

	// Create one for different child
	req := &CreateSleepRequest{
		ChildID:   "child-456",
		Type:      SleepTypeNight,
		StartTime: time.Now(),
	}
	svc.Create(context.Background(), req)

	filter := &SleepFilter{ChildID: "child-123"}
	sleeps, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sleeps) != 3 {
		t.Errorf("List() returned %d sleeps, want 3", len(sleeps))
	}
}

func TestService_List_WithTypeFilter(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	napReq := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: time.Now(),
	}
	svc.Create(context.Background(), napReq)

	nightReq := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNight,
		StartTime: time.Now(),
	}
	svc.Create(context.Background(), nightReq)

	napType := SleepTypeNap
	filter := &SleepFilter{
		ChildID: "child-123",
		Type:    &napType,
	}
	sleeps, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sleeps) != 1 {
		t.Errorf("List() returned %d sleeps, want 1", len(sleeps))
	}

	if sleeps[0].Type != SleepTypeNap {
		t.Errorf("List() returned type %v, want %v", sleeps[0].Type, SleepTypeNap)
	}
}

func TestService_Update(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	newQuality := 5
	newEndTime := time.Now().Add(3 * time.Hour)
	updateReq := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNight,
		StartTime: created.StartTime,
		EndTime:   &newEndTime,
		Quality:   &newQuality,
		Notes:     "Updated notes",
	}

	updated, err := svc.Update(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Type != SleepTypeNight {
		t.Errorf("Update() Type = %v, want %v", updated.Type, SleepTypeNight)
	}

	if updated.Quality == nil || *updated.Quality != newQuality {
		t.Errorf("Update() Quality = %v, want %v", updated.Quality, newQuality)
	}

	if updated.Notes != "Updated notes" {
		t.Errorf("Update() Notes = %v, want Updated notes", updated.Notes)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: time.Now(),
	}

	_, err := svc.Update(context.Background(), "non-existent", req)
	if err == nil {
		t.Error("Update() should return error for non-existent sleep")
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	err := svc.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	sleep, _ := svc.Get(context.Background(), created.ID)
	if sleep != nil {
		t.Error("Delete() should remove the sleep")
	}
}

func TestService_StartSleep(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	sleep, err := svc.StartSleep(context.Background(), "child-123", SleepTypeNap)
	if err != nil {
		t.Fatalf("StartSleep() error = %v", err)
	}

	if sleep.ID == "" {
		t.Error("StartSleep() should generate an ID")
	}

	if sleep.ChildID != "child-123" {
		t.Errorf("StartSleep() ChildID = %v, want child-123", sleep.ChildID)
	}

	if sleep.Type != SleepTypeNap {
		t.Errorf("StartSleep() Type = %v, want %v", sleep.Type, SleepTypeNap)
	}

	if sleep.EndTime != nil {
		t.Error("StartSleep() EndTime should be nil (sleep in progress)")
	}

	if sleep.StartTime.IsZero() {
		t.Error("StartSleep() should set StartTime")
	}
}

func TestService_StartSleep_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = errors.New("database error")
	svc := NewService(repo)

	_, err := svc.StartSleep(context.Background(), "child-123", SleepTypeNap)
	if err == nil {
		t.Error("StartSleep() should return error when repo fails")
	}
}

func TestService_EndSleep(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Start a sleep
	started, _ := svc.StartSleep(context.Background(), "child-123", SleepTypeNap)

	// End it
	ended, err := svc.EndSleep(context.Background(), started.ID)
	if err != nil {
		t.Fatalf("EndSleep() error = %v", err)
	}

	if ended.EndTime == nil {
		t.Error("EndSleep() should set EndTime")
	}

	if ended.EndTime.Before(started.StartTime) {
		t.Error("EndSleep() EndTime should be after StartTime")
	}
}

func TestService_EndSleep_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.EndSleep(context.Background(), "non-existent")
	if err == nil {
		t.Error("EndSleep() should return error for non-existent sleep")
	}
}

func TestService_EndSleep_RepoError(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	started, _ := svc.StartSleep(context.Background(), "child-123", SleepTypeNap)

	repo.updateErr = errors.New("database error")

	_, err := svc.EndSleep(context.Background(), started.ID)
	if err == nil {
		t.Error("EndSleep() should return error when repo fails")
	}
}

func TestService_GetActiveSleep(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Start a sleep (no end time = active)
	started, _ := svc.StartSleep(context.Background(), "child-123", SleepTypeNap)

	// Get active sleep
	active, err := svc.GetActiveSleep(context.Background(), "child-123")
	if err != nil {
		t.Fatalf("GetActiveSleep() error = %v", err)
	}

	if active == nil {
		t.Fatal("GetActiveSleep() returned nil for active sleep")
	}

	if active.ID != started.ID {
		t.Errorf("GetActiveSleep() ID = %v, want %v", active.ID, started.ID)
	}
}

func TestService_GetActiveSleep_NoActive(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a completed sleep
	endTime := time.Now()
	req := &CreateSleepRequest{
		ChildID:   "child-123",
		Type:      SleepTypeNap,
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   &endTime,
	}
	svc.Create(context.Background(), req)

	// Get active sleep
	active, err := svc.GetActiveSleep(context.Background(), "child-123")
	if err != nil {
		t.Fatalf("GetActiveSleep() error = %v", err)
	}

	if active != nil {
		t.Error("GetActiveSleep() should return nil when no active sleep")
	}
}

func TestService_GetActiveSleep_DifferentChild(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Start sleep for child-123
	svc.StartSleep(context.Background(), "child-123", SleepTypeNap)

	// Get active sleep for child-456
	active, err := svc.GetActiveSleep(context.Background(), "child-456")
	if err != nil {
		t.Fatalf("GetActiveSleep() error = %v", err)
	}

	if active != nil {
		t.Error("GetActiveSleep() should return nil for different child")
	}
}

func TestSleepTypes(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	types := []SleepType{SleepTypeNap, SleepTypeNight}

	for _, sleepType := range types {
		t.Run(string(sleepType), func(t *testing.T) {
			req := &CreateSleepRequest{
				ChildID:   "child-123",
				Type:      sleepType,
				StartTime: time.Now(),
			}

			sleep, err := svc.Create(context.Background(), req)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			if sleep.Type != sleepType {
				t.Errorf("Create() Type = %v, want %v", sleep.Type, sleepType)
			}
		})
	}
}
