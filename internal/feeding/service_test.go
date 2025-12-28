package feeding

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	feedings  map[string]*Feeding
	createErr error
	updateErr error
	deleteErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		feedings: make(map[string]*Feeding),
	}
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Feeding, error) {
	f, ok := m.feedings[id]
	if !ok {
		return nil, nil
	}
	return f, nil
}

func (m *mockRepository) List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
	var result []Feeding
	for _, f := range m.feedings {
		if filter.ChildID != "" && f.ChildID != filter.ChildID {
			continue
		}
		if filter.Type != nil && f.Type != *filter.Type {
			continue
		}
		if filter.StartDate != nil && f.StartTime.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && f.StartTime.After(*filter.EndDate) {
			continue
		}
		result = append(result, *f)
	}
	return result, nil
}

func (m *mockRepository) Create(ctx context.Context, feeding *Feeding) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.feedings[feeding.ID] = feeding
	return nil
}

func (m *mockRepository) Update(ctx context.Context, feeding *Feeding) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.feedings[feeding.ID] = feeding
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.feedings, id)
	return nil
}

func (m *mockRepository) GetLastFeeding(ctx context.Context, childID string) (*Feeding, error) {
	var latest *Feeding
	for _, f := range m.feedings {
		if f.ChildID == childID {
			if latest == nil || f.StartTime.After(latest.StartTime) {
				latest = f
			}
		}
	}
	return latest, nil
}

func TestService_Create(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	startTime := time.Now()
	endTime := startTime.Add(30 * time.Minute)
	amount := 120.0

	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
		StartTime: startTime,
		EndTime:   &endTime,
		Amount:    &amount,
		Unit:      "ml",
		Notes:     "Test feeding",
	}

	feeding, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if feeding.ID == "" {
		t.Error("Create() should generate an ID")
	}

	if feeding.ChildID != req.ChildID {
		t.Errorf("Create() ChildID = %v, want %v", feeding.ChildID, req.ChildID)
	}

	if feeding.Type != FeedingTypeBottle {
		t.Errorf("Create() Type = %v, want %v", feeding.Type, FeedingTypeBottle)
	}

	if feeding.Amount == nil || *feeding.Amount != amount {
		t.Errorf("Create() Amount = %v, want %v", feeding.Amount, amount)
	}

	if feeding.Unit != "ml" {
		t.Errorf("Create() Unit = %v, want ml", feeding.Unit)
	}
}

func TestService_Create_BreastFeeding(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBreast,
		StartTime: time.Now(),
		Side:      "left",
	}

	feeding, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if feeding.Type != FeedingTypeBreast {
		t.Errorf("Create() Type = %v, want %v", feeding.Type, FeedingTypeBreast)
	}

	if feeding.Side != "left" {
		t.Errorf("Create() Side = %v, want left", feeding.Side)
	}
}

func TestService_Create_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
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

	// Create a feeding first
	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
		StartTime: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	// Get it back
	feeding, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if feeding == nil {
		t.Fatal("Get() returned nil for existing feeding")
	}

	if feeding.ID != created.ID {
		t.Errorf("Get() ID = %v, want %v", feeding.ID, created.ID)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	feeding, err := svc.Get(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if feeding != nil {
		t.Error("Get() should return nil for non-existent feeding")
	}
}

func TestService_List(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create multiple feedings
	for i := range 3 {
		req := &CreateFeedingRequest{
			ChildID:   "child-123",
			Type:      FeedingTypeBottle,
			StartTime: time.Now().Add(time.Duration(i) * time.Hour),
		}
		svc.Create(context.Background(), req)
	}

	// Also create one for different child
	req := &CreateFeedingRequest{
		ChildID:   "child-456",
		Type:      FeedingTypeBreast,
		StartTime: time.Now(),
	}
	svc.Create(context.Background(), req)

	// List for child-123
	filter := &FeedingFilter{ChildID: "child-123"}
	feedings, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(feedings) != 3 {
		t.Errorf("List() returned %d feedings, want 3", len(feedings))
	}
}

func TestService_List_WithTypeFilter(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create feedings of different types
	bottleReq := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
		StartTime: time.Now(),
	}
	svc.Create(context.Background(), bottleReq)

	breastReq := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBreast,
		StartTime: time.Now(),
	}
	svc.Create(context.Background(), breastReq)

	// Filter by type
	breastType := FeedingTypeBreast
	filter := &FeedingFilter{
		ChildID: "child-123",
		Type:    &breastType,
	}
	feedings, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(feedings) != 1 {
		t.Errorf("List() returned %d feedings, want 1", len(feedings))
	}

	if feedings[0].Type != FeedingTypeBreast {
		t.Errorf("List() returned type %v, want %v", feedings[0].Type, FeedingTypeBreast)
	}
}

func TestService_Update(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a feeding
	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
		StartTime: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	// Update it
	newAmount := 200.0
	newEndTime := time.Now().Add(45 * time.Minute)
	updateReq := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeFormula,
		StartTime: created.StartTime,
		EndTime:   &newEndTime,
		Amount:    &newAmount,
		Unit:      "ml",
		Notes:     "Updated notes",
	}

	updated, err := svc.Update(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Type != FeedingTypeFormula {
		t.Errorf("Update() Type = %v, want %v", updated.Type, FeedingTypeFormula)
	}

	if updated.Amount == nil || *updated.Amount != newAmount {
		t.Errorf("Update() Amount = %v, want %v", updated.Amount, newAmount)
	}

	if updated.Notes != "Updated notes" {
		t.Errorf("Update() Notes = %v, want Updated notes", updated.Notes)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
		StartTime: time.Now(),
	}

	_, err := svc.Update(context.Background(), "non-existent", req)
	if err == nil {
		t.Error("Update() should return error for non-existent feeding")
	}
}

func TestService_Update_RepoError(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a feeding
	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
		StartTime: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	// Set error and try update
	repo.updateErr = errors.New("database error")

	_, err := svc.Update(context.Background(), created.ID, req)
	if err == nil {
		t.Error("Update() should return error when repo fails")
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a feeding
	req := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBottle,
		StartTime: time.Now(),
	}
	created, _ := svc.Create(context.Background(), req)

	// Delete it
	err := svc.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	feeding, _ := svc.Get(context.Background(), created.ID)
	if feeding != nil {
		t.Error("Delete() should remove the feeding")
	}
}

func TestService_Delete_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.deleteErr = errors.New("database error")
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "some-id")
	if err == nil {
		t.Error("Delete() should return error when repo fails")
	}
}

func TestService_GetLastFeeding(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create multiple feedings at different times
	now := time.Now()
	for i := range 3 {
		req := &CreateFeedingRequest{
			ChildID:   "child-123",
			Type:      FeedingTypeBottle,
			StartTime: now.Add(time.Duration(-i) * time.Hour), // Earlier times
		}
		svc.Create(context.Background(), req)
	}

	// Create the most recent one
	latestReq := &CreateFeedingRequest{
		ChildID:   "child-123",
		Type:      FeedingTypeBreast,
		StartTime: now.Add(1 * time.Hour), // Most recent
	}
	latest, _ := svc.Create(context.Background(), latestReq)

	// Get last feeding
	lastFeeding, err := svc.GetLastFeeding(context.Background(), "child-123")
	if err != nil {
		t.Fatalf("GetLastFeeding() error = %v", err)
	}

	if lastFeeding == nil {
		t.Fatal("GetLastFeeding() returned nil")
	}

	if lastFeeding.ID != latest.ID {
		t.Errorf("GetLastFeeding() returned ID = %v, want %v", lastFeeding.ID, latest.ID)
	}
}

func TestService_GetLastFeeding_NoFeedings(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	lastFeeding, err := svc.GetLastFeeding(context.Background(), "child-no-feedings")
	if err != nil {
		t.Fatalf("GetLastFeeding() error = %v", err)
	}

	if lastFeeding != nil {
		t.Error("GetLastFeeding() should return nil when no feedings exist")
	}
}

func TestFeedingTypes(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	types := []FeedingType{
		FeedingTypeBreast,
		FeedingTypeBottle,
		FeedingTypeFormula,
		FeedingTypeSolid,
	}

	for _, feedingType := range types {
		t.Run(string(feedingType), func(t *testing.T) {
			req := &CreateFeedingRequest{
				ChildID:   "child-123",
				Type:      feedingType,
				StartTime: time.Now(),
			}

			feeding, err := svc.Create(context.Background(), req)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			if feeding.Type != feedingType {
				t.Errorf("Create() Type = %v, want %v", feeding.Type, feedingType)
			}
		})
	}
}
