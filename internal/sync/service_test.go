package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ninenine/babytrack/internal/feeding"
	"github.com/ninenine/babytrack/internal/medication"
	"github.com/ninenine/babytrack/internal/notes"
	"github.com/ninenine/babytrack/internal/sleep"
)

// Mock services for testing

type mockFeedingService struct {
	feedings  map[string]*feeding.Feeding
	createErr error
	updateErr error
	deleteErr error
}

func newMockFeedingService() *mockFeedingService {
	return &mockFeedingService{
		feedings: make(map[string]*feeding.Feeding),
	}
}

func (m *mockFeedingService) Create(ctx context.Context, req *feeding.CreateFeedingRequest) (*feeding.Feeding, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	f := &feeding.Feeding{
		ID:        "feeding-new-id",
		ChildID:   req.ChildID,
		Type:      req.Type,
		StartTime: req.StartTime,
	}
	m.feedings[f.ID] = f
	return f, nil
}

func (m *mockFeedingService) Get(ctx context.Context, id string) (*feeding.Feeding, error) {
	return m.feedings[id], nil
}

func (m *mockFeedingService) List(ctx context.Context, filter *feeding.FeedingFilter) ([]feeding.Feeding, error) {
	return nil, nil
}

func (m *mockFeedingService) Update(ctx context.Context, id string, req *feeding.CreateFeedingRequest) (*feeding.Feeding, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	f, ok := m.feedings[id]
	if !ok {
		return nil, errors.New("not found")
	}
	f.Type = req.Type
	return f, nil
}

func (m *mockFeedingService) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.feedings, id)
	return nil
}

func (m *mockFeedingService) GetLastFeeding(ctx context.Context, childID string) (*feeding.Feeding, error) {
	return nil, nil
}

type mockSleepService struct {
	sleeps    map[string]*sleep.Sleep
	createErr error
	updateErr error
	deleteErr error
}

func newMockSleepService() *mockSleepService {
	return &mockSleepService{
		sleeps: make(map[string]*sleep.Sleep),
	}
}

func (m *mockSleepService) Create(ctx context.Context, req *sleep.CreateSleepRequest) (*sleep.Sleep, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	s := &sleep.Sleep{
		ID:        "sleep-new-id",
		ChildID:   req.ChildID,
		Type:      req.Type,
		StartTime: req.StartTime,
	}
	m.sleeps[s.ID] = s
	return s, nil
}

func (m *mockSleepService) Get(ctx context.Context, id string) (*sleep.Sleep, error) {
	return m.sleeps[id], nil
}

func (m *mockSleepService) List(ctx context.Context, filter *sleep.SleepFilter) ([]sleep.Sleep, error) {
	return nil, nil
}

func (m *mockSleepService) Update(ctx context.Context, id string, req *sleep.CreateSleepRequest) (*sleep.Sleep, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	s, ok := m.sleeps[id]
	if !ok {
		return nil, errors.New("not found")
	}
	s.Type = req.Type
	return s, nil
}

func (m *mockSleepService) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.sleeps, id)
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

type mockMedicationService struct {
	medications map[string]*medication.Medication
	logs        map[string]*medication.MedicationLog
	createErr   error
	updateErr   error
	deleteErr   error
	logErr      error
}

func newMockMedicationService() *mockMedicationService {
	return &mockMedicationService{
		medications: make(map[string]*medication.Medication),
		logs:        make(map[string]*medication.MedicationLog),
	}
}

func (m *mockMedicationService) Create(ctx context.Context, req *medication.CreateMedicationRequest) (*medication.Medication, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	med := &medication.Medication{
		ID:      "medication-new-id",
		ChildID: req.ChildID,
		Name:    req.Name,
	}
	m.medications[med.ID] = med
	return med, nil
}

func (m *mockMedicationService) Get(ctx context.Context, id string) (*medication.Medication, error) {
	return m.medications[id], nil
}

func (m *mockMedicationService) List(ctx context.Context, filter *medication.MedicationFilter) ([]medication.Medication, error) {
	return nil, nil
}

func (m *mockMedicationService) Update(ctx context.Context, id string, req *medication.CreateMedicationRequest) (*medication.Medication, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	med, ok := m.medications[id]
	if !ok {
		return nil, errors.New("not found")
	}
	med.Name = req.Name
	return med, nil
}

func (m *mockMedicationService) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.medications, id)
	return nil
}

func (m *mockMedicationService) Deactivate(ctx context.Context, id string) error {
	med, ok := m.medications[id]
	if !ok {
		return errors.New("not found")
	}
	med.Active = false
	return nil
}

func (m *mockMedicationService) LogMedication(ctx context.Context, userID string, req *medication.LogMedicationRequest) (*medication.MedicationLog, error) {
	if m.logErr != nil {
		return nil, m.logErr
	}
	log := &medication.MedicationLog{
		ID:           "log-new-id",
		MedicationID: req.MedicationID,
		GivenBy:      userID,
	}
	m.logs[log.ID] = log
	return log, nil
}

func (m *mockMedicationService) GetLogs(ctx context.Context, medicationID string) ([]medication.MedicationLog, error) {
	return nil, nil
}

func (m *mockMedicationService) GetLastLog(ctx context.Context, medicationID string) (*medication.MedicationLog, error) {
	return nil, nil
}

type mockNotesService struct {
	notes     map[string]*notes.Note
	createErr error
	updateErr error
	deleteErr error
}

func newMockNotesService() *mockNotesService {
	return &mockNotesService{
		notes: make(map[string]*notes.Note),
	}
}

func (m *mockNotesService) Create(ctx context.Context, userID string, req *notes.CreateNoteRequest) (*notes.Note, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	n := &notes.Note{
		ID:       "note-new-id",
		ChildID:  req.ChildID,
		AuthorID: userID,
		Content:  req.Content,
	}
	m.notes[n.ID] = n
	return n, nil
}

func (m *mockNotesService) Get(ctx context.Context, id string) (*notes.Note, error) {
	return m.notes[id], nil
}

func (m *mockNotesService) List(ctx context.Context, filter *notes.NoteFilter) ([]notes.Note, error) {
	return nil, nil
}

func (m *mockNotesService) Update(ctx context.Context, id string, req *notes.UpdateNoteRequest) (*notes.Note, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	n, ok := m.notes[id]
	if !ok {
		return nil, errors.New("not found")
	}
	n.Content = req.Content
	return n, nil
}

func (m *mockNotesService) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.notes, id)
	return nil
}

func (m *mockNotesService) Pin(ctx context.Context, id string, pinned bool) error {
	return nil
}

func (m *mockNotesService) Search(ctx context.Context, childID, query string) ([]notes.Note, error) {
	return nil, nil
}

// Tests

func TestService_Push_FeedingCreate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "bottle",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if resp.Failed != 0 {
		t.Errorf("Push() Failed = %d, want 0", resp.Failed)
	}

	if resp.Results["event-1"] != "feeding-new-id" {
		t.Errorf("Push() Results[event-1] = %v, want feeding-new-id", resp.Results["event-1"])
	}
}

func TestService_Push_FeedingUpdate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	feedingSvc.feedings["feeding-123"] = &feeding.Feeding{ID: "feeding-123"}
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "update",
				EntityID:  "feeding-123",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "breast",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}
}

func TestService_Push_FeedingDelete(t *testing.T) {
	feedingSvc := newMockFeedingService()
	feedingSvc.feedings["feeding-123"] = &feeding.Feeding{ID: "feeding-123"}
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "delete",
				EntityID:  "feeding-123",
				Timestamp: time.Now(),
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if feedingSvc.feedings["feeding-123"] != nil {
		t.Error("Push() should have deleted the feeding")
	}
}

func TestService_Push_SleepCreate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeSleep,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "nap",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if resp.Results["event-1"] != "sleep-new-id" {
		t.Errorf("Push() Results[event-1] = %v, want sleep-new-id", resp.Results["event-1"])
	}
}

func TestService_Push_MedicationCreate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedication,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"name":       "Acetaminophen",
					"dosage":     "5",
					"unit":       "ml",
					"frequency":  "daily",
					"start_date": time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if resp.Results["event-1"] != "medication-new-id" {
		t.Errorf("Push() Results[event-1] = %v, want medication-new-id", resp.Results["event-1"])
	}
}

func TestService_Push_MedicationDeactivate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	medSvc.medications["med-123"] = &medication.Medication{ID: "med-123", Active: true}
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedication,
				Action:    "deactivate",
				EntityID:  "med-123",
				Timestamp: time.Now(),
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if medSvc.medications["med-123"].Active {
		t.Error("Push() should have deactivated the medication")
	}
}

func TestService_Push_MedicationLog(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedicationLog,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"medication_id": "med-123",
					"given_at":      time.Now().Format(time.RFC3339),
					"dosage":        "5ml",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if resp.Results["event-1"] != "log-new-id" {
		t.Errorf("Push() Results[event-1] = %v, want log-new-id", resp.Results["event-1"])
	}
}

func TestService_Push_NoteCreate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeNote,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id": "child-123",
					"content":  "Baby's first steps!",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if resp.Results["event-1"] != "note-new-id" {
		t.Errorf("Push() Results[event-1] = %v, want note-new-id", resp.Results["event-1"])
	}
}

func TestService_Push_NoteUpdate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()
	notesSvc.notes["note-123"] = &notes.Note{ID: "note-123", Content: "Original"}

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeNote,
				Action:    "update",
				EntityID:  "note-123",
				Timestamp: time.Now(),
				Data: map[string]any{
					"content": "Updated content",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}
}

func TestService_Push_NoteDelete(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()
	notesSvc.notes["note-123"] = &notes.Note{ID: "note-123"}

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeNote,
				Action:    "delete",
				EntityID:  "note-123",
				Timestamp: time.Now(),
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if notesSvc.notes["note-123"] != nil {
		t.Error("Push() should have deleted the note")
	}
}

func TestService_Push_MultipleEvents(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "bottle",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
			{
				ID:        "event-2",
				Type:      EventTypeSleep,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "nap",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
			{
				ID:        "event-3",
				Type:      EventTypeNote,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id": "child-123",
					"content":  "Test note",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 3 {
		t.Errorf("Push() Processed = %d, want 3", resp.Processed)
	}

	if resp.Failed != 0 {
		t.Errorf("Push() Failed = %d, want 0", resp.Failed)
	}
}

func TestService_Push_WithFailures(t *testing.T) {
	feedingSvc := newMockFeedingService()
	feedingSvc.createErr = errors.New("database error")
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "bottle",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
			{
				ID:        "event-2",
				Type:      EventTypeNote,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id": "child-123",
					"content":  "Test note",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1", resp.Failed)
	}

	if len(resp.FailedIDs) != 1 || resp.FailedIDs[0] != "event-1" {
		t.Errorf("Push() FailedIDs = %v, want [event-1]", resp.FailedIDs)
	}
}

func TestService_Push_UnknownEventType(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventType("unknown"),
				Action:    "create",
				Timestamp: time.Now(),
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1 for unknown event type", resp.Failed)
	}
}

func TestService_Push_UnknownAction(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "unknown_action",
				Timestamp: time.Now(),
				Data:      map[string]any{},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1 for unknown action", resp.Failed)
	}
}

func TestService_Pull(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	resp, err := svc.Pull(context.Background(), "user-123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if resp.Events == nil {
		t.Error("Pull() Events should not be nil")
	}

	if resp.ServerTime == "" {
		t.Error("Pull() ServerTime should not be empty")
	}
}

func TestService_Status(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	status, err := svc.Status(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if status.ServerTime == "" {
		t.Error("Status() ServerTime should not be empty")
	}
}

func TestService_Push_EmptyEvents(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events:   []Event{},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 0 {
		t.Errorf("Push() Processed = %d, want 0", resp.Processed)
	}

	if resp.Failed != 0 {
		t.Errorf("Push() Failed = %d, want 0", resp.Failed)
	}
}

func TestService_Push_SleepUpdate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	sleepSvc.sleeps["sleep-123"] = &sleep.Sleep{ID: "sleep-123"}
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeSleep,
				Action:    "update",
				EntityID:  "sleep-123",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "night",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}
}

func TestService_Push_SleepDelete(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	sleepSvc.sleeps["sleep-123"] = &sleep.Sleep{ID: "sleep-123"}
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeSleep,
				Action:    "delete",
				EntityID:  "sleep-123",
				Timestamp: time.Now(),
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if sleepSvc.sleeps["sleep-123"] != nil {
		t.Error("Push() should have deleted the sleep record")
	}
}

func TestService_Push_SleepUnknownAction(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeSleep,
				Action:    "unknown_action",
				Timestamp: time.Now(),
				Data:      map[string]any{},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1 for unknown action", resp.Failed)
	}
}

func TestService_Push_MedicationUpdate(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	medSvc.medications["med-123"] = &medication.Medication{ID: "med-123", Name: "Original"}
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedication,
				Action:    "update",
				EntityID:  "med-123",
				Timestamp: time.Now(),
				Data: map[string]any{
					"name":   "Updated Med",
					"dosage": "10",
					"unit":   "mg",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}
}

func TestService_Push_MedicationDelete(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	medSvc.medications["med-123"] = &medication.Medication{ID: "med-123"}
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedication,
				Action:    "delete",
				EntityID:  "med-123",
				Timestamp: time.Now(),
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Processed != 1 {
		t.Errorf("Push() Processed = %d, want 1", resp.Processed)
	}

	if medSvc.medications["med-123"] != nil {
		t.Error("Push() should have deleted the medication")
	}
}

func TestService_Push_MedicationUnknownAction(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedication,
				Action:    "unknown_action",
				Timestamp: time.Now(),
				Data:      map[string]any{},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1 for unknown action", resp.Failed)
	}
}

func TestService_Push_MedicationLogUnknownAction(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedicationLog,
				Action:    "unknown_action",
				Timestamp: time.Now(),
				Data:      map[string]any{},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1 for unknown action", resp.Failed)
	}
}

func TestService_Push_NoteUnknownAction(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeNote,
				Action:    "unknown_action",
				Timestamp: time.Now(),
				Data:      map[string]any{},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1 for unknown action", resp.Failed)
	}
}

func TestService_Push_SleepServiceError(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	sleepSvc.createErr = errors.New("sleep service error")
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeSleep,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "nap",
					"start_time": time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1", resp.Failed)
	}
}

func TestService_Push_MedicationServiceError(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	medSvc.createErr = errors.New("medication service error")
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedication,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id": "child-123",
					"name":     "Test Med",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1", resp.Failed)
	}
}

func TestService_Push_MedicationLogServiceError(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	medSvc.logErr = errors.New("log service error")
	notesSvc := newMockNotesService()

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeMedicationLog,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"medication_id": "med-123",
					"given_at":      time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1", resp.Failed)
	}
}

func TestService_Push_NoteServiceError(t *testing.T) {
	feedingSvc := newMockFeedingService()
	sleepSvc := newMockSleepService()
	medSvc := newMockMedicationService()
	notesSvc := newMockNotesService()
	notesSvc.createErr = errors.New("note service error")

	svc := NewService(feedingSvc, sleepSvc, medSvc, notesSvc)

	req := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeNote,
				Action:    "create",
				Timestamp: time.Now(),
				Data: map[string]any{
					"child_id": "child-123",
					"content":  "Test note",
				},
			},
		},
	}

	resp, err := svc.Push(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Failed != 1 {
		t.Errorf("Push() Failed = %d, want 1", resp.Failed)
	}
}
