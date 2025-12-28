package sleep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockService implements the Service interface for testing
type mockService struct {
	createFn         func(ctx context.Context, req *CreateSleepRequest) (*Sleep, error)
	getFn            func(ctx context.Context, id string) (*Sleep, error)
	listFn           func(ctx context.Context, filter *SleepFilter) ([]Sleep, error)
	updateFn         func(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error)
	deleteFn         func(ctx context.Context, id string) error
	startSleepFn     func(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error)
	endSleepFn       func(ctx context.Context, id string) (*Sleep, error)
	getActiveSleepFn func(ctx context.Context, childID string) (*Sleep, error)
}

func (m *mockService) Create(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}

func (m *mockService) Get(ctx context.Context, id string) (*Sleep, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) List(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockService) Update(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return nil, nil
}

func (m *mockService) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockService) StartSleep(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
	if m.startSleepFn != nil {
		return m.startSleepFn(ctx, childID, sleepType)
	}
	return nil, nil
}

func (m *mockService) EndSleep(ctx context.Context, id string) (*Sleep, error) {
	if m.endSleepFn != nil {
		return m.endSleepFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) GetActiveSleep(ctx context.Context, childID string) (*Sleep, error) {
	if m.getActiveSleepFn != nil {
		return m.getActiveSleepFn(ctx, childID)
	}
	return nil, nil
}

// setupRouter creates a test router with the handler registered
func setupRouter(svc Service) *gin.Engine {
	router := gin.New()
	handler := NewHandler(svc)

	// Add middleware to set user_id in context (simulating auth)
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Next()
	})

	group := router.Group("/sleep")
	handler.RegisterRoutes(group)
	return router
}

// Helper to create a sample sleep record
func sampleSleep() *Sleep {
	now := time.Now()
	quality := 4
	return &Sleep{
		ID:        "sleep-123",
		ChildID:   "child-456",
		Type:      SleepTypeNap,
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   &now,
		Quality:   &quality,
		Notes:     "Good nap",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper to create an active sleep record (no end time)
func sampleActiveSleep() *Sleep {
	now := time.Now()
	return &Sleep{
		ID:        "sleep-active",
		ChildID:   "child-456",
		Type:      SleepTypeNight,
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper to create a valid request body
func validRequestBody() *CreateSleepRequest {
	now := time.Now()
	endTime := now.Add(2 * time.Hour)
	quality := 4
	return &CreateSleepRequest{
		ChildID:   "child-456",
		Type:      SleepTypeNap,
		StartTime: now,
		EndTime:   &endTime,
		Quality:   &quality,
		Notes:     "Good nap",
	}
}

// =====================
// List Handler Tests
// =====================

func TestList_Success(t *testing.T) {
	sleeps := []Sleep{*sampleSleep()}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
			return sleeps, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 sleep record, got %d", len(result))
	}
}

func TestList_WithChildIDFilter(t *testing.T) {
	var capturedFilter *SleepFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
			capturedFilter = filter
			return []Sleep{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep?child_id=child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedFilter == nil {
		t.Fatal("Filter was not passed to service")
	}
	if capturedFilter.ChildID != "child-456" {
		t.Errorf("Expected ChildID child-456, got %s", capturedFilter.ChildID)
	}
}

func TestList_ServiceError(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "database connection failed" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestList_EmptyResult(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
			return []Sleep{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty list, got %d", len(result))
	}
}

// =====================
// Get Handler Tests
// =====================

func TestGet_Success(t *testing.T) {
	slp := sampleSleep()
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Sleep, error) {
			if id == "sleep-123" {
				return slp, nil
			}
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep/sleep-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "sleep-123" {
		t.Errorf("Expected ID sleep-123, got %s", result.ID)
	}
}

func TestGet_ServiceError(t *testing.T) {
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Sleep, error) {
			return nil, errors.New("sleep record not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "sleep record not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGet_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Sleep, error) {
			capturedID = id
			return sampleSleep(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep/my-specific-id", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "my-specific-id" {
		t.Errorf("Expected captured ID my-specific-id, got %s", capturedID)
	}
}

// =====================
// Create Handler Tests
// =====================

func TestCreate_Success(t *testing.T) {
	slp := sampleSleep()
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
			return slp, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("POST", "/sleep", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "sleep-123" {
		t.Errorf("Expected ID sleep-123, got %s", result.ID)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/sleep", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] == "" {
		t.Error("Expected error message for invalid JSON")
	}
}

func TestCreate_MissingRequiredFields(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	// Missing required fields (child_id, type, start_time)
	body, _ := json.Marshal(map[string]any{
		"notes": "Some notes",
	})
	req := httptest.NewRequest("POST", "/sleep", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreate_ServiceError(t *testing.T) {
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
			return nil, errors.New("failed to create sleep record")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("POST", "/sleep", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "failed to create sleep record" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestCreate_VerifiesRequestData(t *testing.T) {
	var capturedReq *CreateSleepRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
			capturedReq = req
			return sampleSleep(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validRequestBody()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/sleep", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.ChildID != reqBody.ChildID {
		t.Errorf("Expected ChildID %s, got %s", reqBody.ChildID, capturedReq.ChildID)
	}
	if capturedReq.Type != reqBody.Type {
		t.Errorf("Expected Type %s, got %s", reqBody.Type, capturedReq.Type)
	}
}

// =====================
// Update Handler Tests
// =====================

func TestUpdate_Success(t *testing.T) {
	slp := sampleSleep()
	slp.Notes = "Updated notes"
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error) {
			return slp, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("PUT", "/sleep/sleep-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Notes != "Updated notes" {
		t.Errorf("Expected Notes 'Updated notes', got %s", result.Notes)
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("PUT", "/sleep/sleep-123", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_MissingRequiredFields(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"notes": "Just notes",
	})
	req := httptest.NewRequest("PUT", "/sleep/sleep-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error) {
			return nil, errors.New("sleep record not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("PUT", "/sleep/nonexistent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "sleep record not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestUpdate_VerifiesIDAndRequest(t *testing.T) {
	var capturedID string
	var capturedReq *CreateSleepRequest
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error) {
			capturedID = id
			capturedReq = req
			return sampleSleep(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validRequestBody()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/sleep/update-id-456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "update-id-456" {
		t.Errorf("Expected ID update-id-456, got %s", capturedID)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.Notes != reqBody.Notes {
		t.Errorf("Expected Notes %s, got %s", reqBody.Notes, capturedReq.Notes)
	}
}

// =====================
// Delete Handler Tests
// =====================

func TestDelete_Success(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/sleep/sleep-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDelete_ServiceError(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("sleep record not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/sleep/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "sleep record not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestDelete_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/sleep/delete-me-789", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "delete-me-789" {
		t.Errorf("Expected ID delete-me-789, got %s", capturedID)
	}
}

// =====================
// StartSleep Handler Tests
// =====================

func TestStartSleep_Success(t *testing.T) {
	slp := sampleActiveSleep()
	svc := &mockService{
		startSleepFn: func(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
			return slp, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "child-456",
		"type":     "nap",
	})
	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.EndTime != nil {
		t.Error("Expected EndTime to be nil for active sleep")
	}
}

func TestStartSleep_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestStartSleep_MissingChildID(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"type": "nap",
	})
	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestStartSleep_MissingType(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "child-456",
	})
	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestStartSleep_ServiceError(t *testing.T) {
	svc := &mockService{
		startSleepFn: func(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
			return nil, errors.New("failed to start sleep")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "child-456",
		"type":     "nap",
	})
	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "failed to start sleep" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestStartSleep_VerifiesRequestData(t *testing.T) {
	var capturedChildID string
	var capturedType SleepType
	svc := &mockService{
		startSleepFn: func(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
			capturedChildID = childID
			capturedType = sleepType
			return sampleActiveSleep(), nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "specific-child",
		"type":     "night",
	})
	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedChildID != "specific-child" {
		t.Errorf("Expected ChildID specific-child, got %s", capturedChildID)
	}
	if capturedType != SleepTypeNight {
		t.Errorf("Expected Type night, got %s", capturedType)
	}
}

// =====================
// EndSleep Handler Tests
// =====================

func TestEndSleep_Success(t *testing.T) {
	slp := sampleSleep() // Has EndTime set
	svc := &mockService{
		endSleepFn: func(ctx context.Context, id string) (*Sleep, error) {
			return slp, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/sleep/sleep-123/end", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.EndTime == nil {
		t.Error("Expected EndTime to be set after ending sleep")
	}
}

func TestEndSleep_ServiceError(t *testing.T) {
	svc := &mockService{
		endSleepFn: func(ctx context.Context, id string) (*Sleep, error) {
			return nil, errors.New("sleep record not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/sleep/nonexistent/end", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "sleep record not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestEndSleep_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		endSleepFn: func(ctx context.Context, id string) (*Sleep, error) {
			capturedID = id
			return sampleSleep(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/sleep/end-this-sleep/end", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "end-this-sleep" {
		t.Errorf("Expected ID end-this-sleep, got %s", capturedID)
	}
}

// =====================
// GetActive Handler Tests
// =====================

func TestGetActive_Success(t *testing.T) {
	slp := sampleActiveSleep()
	svc := &mockService{
		getActiveSleepFn: func(ctx context.Context, childID string) (*Sleep, error) {
			return slp, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep/active/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Sleep
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "sleep-active" {
		t.Errorf("Expected ID sleep-active, got %s", result.ID)
	}
}

func TestGetActive_NoActiveSleep(t *testing.T) {
	svc := &mockService{
		getActiveSleepFn: func(ctx context.Context, childID string) (*Sleep, error) {
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep/active/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should return null when no active sleep
	if w.Body.String() != "null" {
		t.Errorf("Expected null response, got %s", w.Body.String())
	}
}

func TestGetActive_ServiceError(t *testing.T) {
	svc := &mockService{
		getActiveSleepFn: func(ctx context.Context, childID string) (*Sleep, error) {
			return nil, errors.New("database error")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep/active/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "database error" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGetActive_VerifiesChildIDParam(t *testing.T) {
	var capturedChildID string
	svc := &mockService{
		getActiveSleepFn: func(ctx context.Context, childID string) (*Sleep, error) {
			capturedChildID = childID
			return sampleActiveSleep(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep/active/specific-child-id", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedChildID != "specific-child-id" {
		t.Errorf("Expected childID specific-child-id, got %s", capturedChildID)
	}
}

// =====================
// Route Registration Tests
// =====================

func TestRegisterRoutes(t *testing.T) {
	now := time.Now()
	svc := &mockService{
		listFn: func(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
			return []Sleep{}, nil
		},
		getFn: func(ctx context.Context, id string) (*Sleep, error) {
			return sampleSleep(), nil
		},
		createFn: func(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
			return sampleSleep(), nil
		},
		updateFn: func(ctx context.Context, id string, req *CreateSleepRequest) (*Sleep, error) {
			return sampleSleep(), nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
		startSleepFn: func(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
			return sampleActiveSleep(), nil
		},
		endSleepFn: func(ctx context.Context, id string) (*Sleep, error) {
			return sampleSleep(), nil
		},
		getActiveSleepFn: func(ctx context.Context, childID string) (*Sleep, error) {
			return sampleActiveSleep(), nil
		},
	}
	router := setupRouter(svc)

	testCases := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"GET", "/sleep", "", http.StatusOK},
		{"GET", "/sleep/sleep-123", "", http.StatusOK},
		{"POST", "/sleep", `{"child_id":"c1","type":"nap","start_time":"` + now.Format(time.RFC3339) + `"}`, http.StatusCreated},
		{"PUT", "/sleep/sleep-123", `{"child_id":"c1","type":"nap","start_time":"` + now.Format(time.RFC3339) + `"}`, http.StatusOK},
		{"DELETE", "/sleep/sleep-123", "", http.StatusNoContent},
		{"POST", "/sleep/start", `{"child_id":"c1","type":"nap"}`, http.StatusCreated},
		{"POST", "/sleep/sleep-123/end", "", http.StatusOK},
		{"GET", "/sleep/active/child-456", "", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, http.NoBody)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("%s %s: expected status %d, got %d (body: %s)", tc.method, tc.path, tc.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}

// =====================
// NewHandler Tests
// =====================

func TestNewHandler(t *testing.T) {
	svc := &mockService{}
	handler := NewHandler(svc)

	if handler == nil {
		t.Fatal("Expected handler to be created")
	}
	if handler.service == nil {
		t.Error("Expected service to be set")
	}
}

// =====================
// Sleep Type Tests
// =====================

func TestStartSleep_NapType(t *testing.T) {
	var capturedType SleepType
	svc := &mockService{
		startSleepFn: func(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
			capturedType = sleepType
			return sampleActiveSleep(), nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "child-456",
		"type":     "nap",
	})
	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedType != SleepTypeNap {
		t.Errorf("Expected type nap, got %s", capturedType)
	}
}

func TestStartSleep_NightType(t *testing.T) {
	var capturedType SleepType
	svc := &mockService{
		startSleepFn: func(ctx context.Context, childID string, sleepType SleepType) (*Sleep, error) {
			capturedType = sleepType
			return sampleActiveSleep(), nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "child-456",
		"type":     "night",
	})
	req := httptest.NewRequest("POST", "/sleep/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedType != SleepTypeNight {
		t.Errorf("Expected type night, got %s", capturedType)
	}
}

// =====================
// Edge Case Tests
// =====================

func TestList_NilFilter(t *testing.T) {
	var capturedFilter *SleepFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *SleepFilter) ([]Sleep, error) {
			capturedFilter = filter
			return []Sleep{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sleep", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedFilter == nil {
		t.Fatal("Expected filter to be passed, even if empty")
	}
	if capturedFilter.ChildID != "" {
		t.Errorf("Expected empty ChildID, got %s", capturedFilter.ChildID)
	}
}

func TestCreate_WithOptionalFields(t *testing.T) {
	var capturedReq *CreateSleepRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
			capturedReq = req
			return sampleSleep(), nil
		},
	}
	router := setupRouter(svc)

	now := time.Now()
	endTime := now.Add(2 * time.Hour)
	quality := 5
	reqBody := &CreateSleepRequest{
		ChildID:   "child-456",
		Type:      SleepTypeNight,
		StartTime: now,
		EndTime:   &endTime,
		Quality:   &quality,
		Notes:     "Excellent sleep",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/sleep", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.Quality == nil || *capturedReq.Quality != 5 {
		t.Error("Expected Quality to be 5")
	}
	if capturedReq.Notes != "Excellent sleep" {
		t.Errorf("Expected Notes 'Excellent sleep', got %s", capturedReq.Notes)
	}
}

func TestCreate_WithoutOptionalFields(t *testing.T) {
	var capturedReq *CreateSleepRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateSleepRequest) (*Sleep, error) {
			capturedReq = req
			return sampleSleep(), nil
		},
	}
	router := setupRouter(svc)

	now := time.Now()
	reqBody := map[string]any{
		"child_id":   "child-456",
		"type":       "nap",
		"start_time": now.Format(time.RFC3339),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/sleep", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.EndTime != nil {
		t.Error("Expected EndTime to be nil")
	}
	if capturedReq.Quality != nil {
		t.Error("Expected Quality to be nil")
	}
}
