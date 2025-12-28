package medication

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	medications  map[string]*Medication
	logs         map[string][]*MedicationLog
	createErr    error
	updateErr    error
	deleteErr    error
	createLogErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		medications: make(map[string]*Medication),
		logs:        make(map[string][]*MedicationLog),
	}
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Medication, error) {
	med, ok := m.medications[id]
	if !ok {
		return nil, nil
	}
	return med, nil
}

func (m *mockRepository) List(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
	var result []Medication
	for _, med := range m.medications {
		if filter.ChildID != "" && med.ChildID != filter.ChildID {
			continue
		}
		if filter.ActiveOnly && !med.Active {
			continue
		}
		result = append(result, *med)
	}
	return result, nil
}

func (m *mockRepository) Create(ctx context.Context, med *Medication) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.medications[med.ID] = med
	return nil
}

func (m *mockRepository) Update(ctx context.Context, med *Medication) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.medications[med.ID] = med
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.medications, id)
	return nil
}

func (m *mockRepository) CreateLog(ctx context.Context, log *MedicationLog) error {
	if m.createLogErr != nil {
		return m.createLogErr
	}
	m.logs[log.MedicationID] = append(m.logs[log.MedicationID], log)
	return nil
}

func (m *mockRepository) ListLogs(ctx context.Context, medicationID string) ([]MedicationLog, error) {
	logs := m.logs[medicationID]
	var result []MedicationLog
	for _, log := range logs {
		result = append(result, *log)
	}
	return result, nil
}

func (m *mockRepository) GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error) {
	logs := m.logs[medicationID]
	if len(logs) == 0 {
		return nil, nil
	}
	var latest *MedicationLog
	for _, log := range logs {
		if latest == nil || log.GivenAt.After(latest.GivenAt) {
			latest = log
		}
	}
	return latest, nil
}

func (m *mockRepository) GetLogByID(ctx context.Context, id string) (*MedicationLog, error) {
	for _, logs := range m.logs {
		for _, log := range logs {
			if log.ID == id {
				return log, nil
			}
		}
	}
	return nil, nil
}

func TestService_Create(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	startDate := time.Now()
	endDate := startDate.Add(30 * 24 * time.Hour)

	req := &CreateMedicationRequest{
		ChildID:      "child-123",
		Name:         "Acetaminophen",
		Dosage:       "5",
		Unit:         "ml",
		Frequency:    "every_6_hours",
		Instructions: "Give with food",
		StartDate:    startDate,
		EndDate:      &endDate,
	}

	med, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if med.ID == "" {
		t.Error("Create() should generate an ID")
	}

	if med.ChildID != req.ChildID {
		t.Errorf("Create() ChildID = %v, want %v", med.ChildID, req.ChildID)
	}

	if med.Name != "Acetaminophen" {
		t.Errorf("Create() Name = %v, want Acetaminophen", med.Name)
	}

	if !med.Active {
		t.Error("Create() should set Active to true")
	}

	if med.Instructions != "Give with food" {
		t.Errorf("Create() Instructions = %v, want 'Give with food'", med.Instructions)
	}
}

func TestService_Create_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Error("Create() should return error when repo fails")
	}
}

func TestService_Get(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	med, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if med == nil {
		t.Fatal("Get() returned nil for existing medication")
	}

	if med.ID != created.ID {
		t.Errorf("Get() ID = %v, want %v", med.ID, created.ID)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	med, err := svc.Get(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if med != nil {
		t.Error("Get() should return nil for non-existent medication")
	}
}

func TestService_List(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create multiple medications
	for i := range 3 {
		req := &CreateMedicationRequest{
			ChildID:   "child-123",
			Name:      "Med " + string(rune('A'+i)),
			Dosage:    "10",
			Unit:      "mg",
			Frequency: "daily",
			StartDate: time.Now(),
		}
		svc.Create(context.Background(), req)
	}

	// Create one for different child
	req := &CreateMedicationRequest{
		ChildID:   "child-456",
		Name:      "Other Med",
		Dosage:    "5",
		Unit:      "ml",
		Frequency: "twice_daily",
		StartDate: time.Now(),
	}
	svc.Create(context.Background(), req)

	filter := &MedicationFilter{ChildID: "child-123"}
	meds, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(meds) != 3 {
		t.Errorf("List() returned %d medications, want 3", len(meds))
	}
}

func TestService_List_ActiveOnly(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create an active medication
	activeReq := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Active Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	svc.Create(context.Background(), activeReq)

	// Create and deactivate another
	inactiveReq := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Inactive Med",
		Dosage:    "5",
		Unit:      "ml",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	inactive, _ := svc.Create(context.Background(), inactiveReq)
	svc.Deactivate(context.Background(), inactive.ID)

	filter := &MedicationFilter{ChildID: "child-123", ActiveOnly: true}
	meds, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(meds) != 1 {
		t.Errorf("List() with ActiveOnly returned %d medications, want 1", len(meds))
	}
}

func TestService_Update(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Original Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	updateReq := &CreateMedicationRequest{
		ChildID:      "child-123",
		Name:         "Updated Med",
		Dosage:       "20",
		Unit:         "mg",
		Frequency:    "twice_daily",
		Instructions: "New instructions",
		StartDate:    created.StartDate,
	}

	updated, err := svc.Update(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Name != "Updated Med" {
		t.Errorf("Update() Name = %v, want Updated Med", updated.Name)
	}

	if updated.Dosage != "20" {
		t.Errorf("Update() Dosage = %v, want 20", updated.Dosage)
	}

	if updated.Frequency != "twice_daily" {
		t.Errorf("Update() Frequency = %v, want twice_daily", updated.Frequency)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}

	_, err := svc.Update(context.Background(), "non-existent", req)
	if err == nil {
		t.Error("Update() should return error for non-existent medication")
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	err := svc.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	med, _ := svc.Get(context.Background(), created.ID)
	if med != nil {
		t.Error("Delete() should remove the medication")
	}
}

func TestService_Deactivate(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	err := svc.Deactivate(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}

	med, _ := svc.Get(context.Background(), created.ID)
	if med.Active {
		t.Error("Deactivate() should set Active to false")
	}

	if med.EndDate == nil {
		t.Error("Deactivate() should set EndDate")
	}
}

func TestService_Deactivate_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	err := svc.Deactivate(context.Background(), "non-existent")
	if err == nil {
		t.Error("Deactivate() should return error for non-existent medication")
	}
}

func TestService_LogMedication(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a medication first
	medReq := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	med, _ := svc.Create(context.Background(), medReq)

	logReq := &LogMedicationRequest{
		MedicationID: med.ID,
		GivenAt:      time.Now(),
		Dosage:       "10mg",
		Notes:        "Given without issues",
	}

	log, err := svc.LogMedication(context.Background(), "user-123", logReq)
	if err != nil {
		t.Fatalf("LogMedication() error = %v", err)
	}

	if log.ID == "" {
		t.Error("LogMedication() should generate an ID")
	}

	if log.MedicationID != med.ID {
		t.Errorf("LogMedication() MedicationID = %v, want %v", log.MedicationID, med.ID)
	}

	if log.ChildID != med.ChildID {
		t.Errorf("LogMedication() ChildID = %v, want %v", log.ChildID, med.ChildID)
	}

	if log.GivenBy != "user-123" {
		t.Errorf("LogMedication() GivenBy = %v, want user-123", log.GivenBy)
	}
}

func TestService_LogMedication_MedicationNotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	logReq := &LogMedicationRequest{
		MedicationID: "non-existent",
		GivenAt:      time.Now(),
		Dosage:       "10mg",
	}

	_, err := svc.LogMedication(context.Background(), "user-123", logReq)
	if err == nil {
		t.Error("LogMedication() should return error for non-existent medication")
	}
}

func TestService_GetLogs(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a medication
	medReq := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	med, _ := svc.Create(context.Background(), medReq)

	// Log it multiple times
	for i := range 3 {
		logReq := &LogMedicationRequest{
			MedicationID: med.ID,
			GivenAt:      time.Now().Add(time.Duration(i) * time.Hour),
			Dosage:       "10mg",
		}
		svc.LogMedication(context.Background(), "user-123", logReq)
	}

	logs, err := svc.GetLogs(context.Background(), med.ID)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("GetLogs() returned %d logs, want 3", len(logs))
	}
}

func TestService_GetLastLog(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a medication
	medReq := &CreateMedicationRequest{
		ChildID:   "child-123",
		Name:      "Test Med",
		Dosage:    "10",
		Unit:      "mg",
		Frequency: "daily",
		StartDate: time.Now(),
	}
	med, _ := svc.Create(context.Background(), medReq)

	// Log it at different times
	now := time.Now()
	for i := range 3 {
		logReq := &LogMedicationRequest{
			MedicationID: med.ID,
			GivenAt:      now.Add(time.Duration(-i) * time.Hour), // Earlier times
			Dosage:       "10mg",
		}
		svc.LogMedication(context.Background(), "user-123", logReq)
	}

	// Log the most recent one
	latestLogReq := &LogMedicationRequest{
		MedicationID: med.ID,
		GivenAt:      now.Add(1 * time.Hour), // Most recent
		Dosage:       "latest",
	}
	svc.LogMedication(context.Background(), "user-123", latestLogReq)

	lastLog, err := svc.GetLastLog(context.Background(), med.ID)
	if err != nil {
		t.Fatalf("GetLastLog() error = %v", err)
	}

	if lastLog == nil {
		t.Fatal("GetLastLog() returned nil")
	}

	if lastLog.Dosage != "latest" {
		t.Errorf("GetLastLog() returned wrong log, Dosage = %v, want 'latest'", lastLog.Dosage)
	}
}

func TestService_GetLastLog_NoLogs(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	lastLog, err := svc.GetLastLog(context.Background(), "med-no-logs")
	if err != nil {
		t.Fatalf("GetLastLog() error = %v", err)
	}

	if lastLog != nil {
		t.Error("GetLastLog() should return nil when no logs exist")
	}
}
