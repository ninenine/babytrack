package feeding

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
	createFn         func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error)
	getFn            func(ctx context.Context, id string) (*Feeding, error)
	listFn           func(ctx context.Context, filter *FeedingFilter) ([]Feeding, error)
	updateFn         func(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error)
	deleteFn         func(ctx context.Context, id string) error
	getLastFeedingFn func(ctx context.Context, childID string) (*Feeding, error)
}

func (m *mockService) Create(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}

func (m *mockService) Get(ctx context.Context, id string) (*Feeding, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) List(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockService) Update(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error) {
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

func (m *mockService) GetLastFeeding(ctx context.Context, childID string) (*Feeding, error) {
	if m.getLastFeedingFn != nil {
		return m.getLastFeedingFn(ctx, childID)
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

	group := router.Group("/feedings")
	handler.RegisterRoutes(group)
	return router
}

// Helper to create a sample feeding
func sampleFeeding() *Feeding {
	amount := 120.0
	return &Feeding{
		ID:        "feeding-123",
		ChildID:   "child-456",
		Type:      FeedingTypeBottle,
		StartTime: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		Amount:    &amount,
		Unit:      "ml",
		Notes:     "Fed well",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Helper to create a valid request body
func validRequestBody() *CreateFeedingRequest {
	amount := 120.0
	return &CreateFeedingRequest{
		ChildID:   "child-456",
		Type:      FeedingTypeBottle,
		StartTime: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		Amount:    &amount,
		Unit:      "ml",
		Notes:     "Fed well",
	}
}

// =====================
// List Handler Tests
// =====================

func TestList_Success(t *testing.T) {
	feedings := []Feeding{*sampleFeeding()}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
			return feedings, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Feeding
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 feeding, got %d", len(result))
	}
}

func TestList_WithFilter(t *testing.T) {
	var capturedFilter *FeedingFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
			capturedFilter = filter
			return []Feeding{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings?child_id=child-456", http.NoBody)
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
		listFn: func(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings", http.NoBody)
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
		listFn: func(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
			return []Feeding{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Feeding
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
	feeding := sampleFeeding()
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Feeding, error) {
			if id == "feeding-123" {
				return feeding, nil
			}
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings/feeding-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Feeding
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "feeding-123" {
		t.Errorf("Expected ID feeding-123, got %s", result.ID)
	}
}

func TestGet_ServiceError(t *testing.T) {
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Feeding, error) {
			return nil, errors.New("feeding not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "feeding not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGet_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Feeding, error) {
			capturedID = id
			return sampleFeeding(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings/my-specific-id", http.NoBody)
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
	feeding := sampleFeeding()
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
			return feeding, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result Feeding
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "feeding-123" {
		t.Errorf("Expected ID feeding-123, got %s", result.ID)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader([]byte("invalid json")))
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
	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreate_ServiceError(t *testing.T) {
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
			return nil, errors.New("failed to create feeding")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader(body))
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
	if result["error"] != "failed to create feeding" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestCreate_VerifiesRequestData(t *testing.T) {
	var capturedReq *CreateFeedingRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
			capturedReq = req
			return sampleFeeding(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validRequestBody()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader(body))
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
	if capturedReq.Unit != reqBody.Unit {
		t.Errorf("Expected Unit %s, got %s", reqBody.Unit, capturedReq.Unit)
	}
}

func TestCreate_AllFeedingTypes(t *testing.T) {
	feedingTypes := []FeedingType{
		FeedingTypeBreast,
		FeedingTypeBottle,
		FeedingTypeFormula,
		FeedingTypeSolid,
	}

	for _, feedingType := range feedingTypes {
		t.Run(string(feedingType), func(t *testing.T) {
			var capturedType FeedingType
			svc := &mockService{
				createFn: func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
					capturedType = req.Type
					return sampleFeeding(), nil
				},
			}
			router := setupRouter(svc)

			reqBody := validRequestBody()
			reqBody.Type = feedingType
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/feedings", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("Expected status 201, got %d", w.Code)
			}
			if capturedType != feedingType {
				t.Errorf("Expected type %s, got %s", feedingType, capturedType)
			}
		})
	}
}

func TestCreate_WithOptionalFields(t *testing.T) {
	var capturedReq *CreateFeedingRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
			capturedReq = req
			return sampleFeeding(), nil
		},
	}
	router := setupRouter(svc)

	endTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	amount := 150.0
	reqBody := &CreateFeedingRequest{
		ChildID:   "child-456",
		Type:      FeedingTypeBreast,
		StartTime: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		EndTime:   &endTime,
		Amount:    &amount,
		Unit:      "ml",
		Side:      "left",
		Notes:     "Good feeding session",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
	if capturedReq.Side != "left" {
		t.Errorf("Expected Side left, got %s", capturedReq.Side)
	}
	if capturedReq.EndTime == nil {
		t.Error("Expected EndTime to be set")
	}
	if capturedReq.Amount == nil || *capturedReq.Amount != 150.0 {
		t.Error("Expected Amount to be 150.0")
	}
}

// =====================
// Update Handler Tests
// =====================

func TestUpdate_Success(t *testing.T) {
	feeding := sampleFeeding()
	feeding.Notes = "Updated notes"
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error) {
			return feeding, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("PUT", "/feedings/feeding-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Feeding
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

	req := httptest.NewRequest("PUT", "/feedings/feeding-123", bytes.NewReader([]byte("invalid json")))
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
		"notes": "Some notes",
	})
	req := httptest.NewRequest("PUT", "/feedings/feeding-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error) {
			return nil, errors.New("feeding not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("PUT", "/feedings/nonexistent", bytes.NewReader(body))
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
	if result["error"] != "feeding not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestUpdate_VerifiesIDAndRequest(t *testing.T) {
	var capturedID string
	var capturedReq *CreateFeedingRequest
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error) {
			capturedID = id
			capturedReq = req
			return sampleFeeding(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validRequestBody()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/feedings/update-id-456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "update-id-456" {
		t.Errorf("Expected ID update-id-456, got %s", capturedID)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.ChildID != reqBody.ChildID {
		t.Errorf("Expected ChildID %s, got %s", reqBody.ChildID, capturedReq.ChildID)
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

	req := httptest.NewRequest("DELETE", "/feedings/feeding-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDelete_ServiceError(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("feeding not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/feedings/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "feeding not found" {
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

	req := httptest.NewRequest("DELETE", "/feedings/delete-me-789", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "delete-me-789" {
		t.Errorf("Expected ID delete-me-789, got %s", capturedID)
	}
}

// =====================
// GetLast Handler Tests
// =====================

func TestGetLast_Success(t *testing.T) {
	feeding := sampleFeeding()
	svc := &mockService{
		getLastFeedingFn: func(ctx context.Context, childID string) (*Feeding, error) {
			return feeding, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings/last/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Feeding
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "feeding-123" {
		t.Errorf("Expected ID feeding-123, got %s", result.ID)
	}
}

func TestGetLast_ServiceError(t *testing.T) {
	svc := &mockService{
		getLastFeedingFn: func(ctx context.Context, childID string) (*Feeding, error) {
			return nil, errors.New("no feedings found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings/last/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "no feedings found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGetLast_VerifiesChildIDParam(t *testing.T) {
	var capturedChildID string
	svc := &mockService{
		getLastFeedingFn: func(ctx context.Context, childID string) (*Feeding, error) {
			capturedChildID = childID
			return sampleFeeding(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings/last/specific-child-id", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedChildID != "specific-child-id" {
		t.Errorf("Expected childID specific-child-id, got %s", capturedChildID)
	}
}

func TestGetLast_NilResult(t *testing.T) {
	svc := &mockService{
		getLastFeedingFn: func(ctx context.Context, childID string) (*Feeding, error) {
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings/last/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// =====================
// Route Registration Tests
// =====================

func TestRegisterRoutes(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
			return []Feeding{}, nil
		},
		getFn: func(ctx context.Context, id string) (*Feeding, error) {
			return sampleFeeding(), nil
		},
		createFn: func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
			return sampleFeeding(), nil
		},
		updateFn: func(ctx context.Context, id string, req *CreateFeedingRequest) (*Feeding, error) {
			return sampleFeeding(), nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
		getLastFeedingFn: func(ctx context.Context, childID string) (*Feeding, error) {
			return sampleFeeding(), nil
		},
	}
	router := setupRouter(svc)

	testCases := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"GET", "/feedings", "", http.StatusOK},
		{"GET", "/feedings/feeding-123", "", http.StatusOK},
		{"POST", "/feedings", `{"child_id":"c1","type":"bottle","start_time":"2025-01-15T10:00:00Z"}`, http.StatusCreated},
		{"PUT", "/feedings/feeding-123", `{"child_id":"c1","type":"bottle","start_time":"2025-01-15T10:00:00Z"}`, http.StatusOK},
		{"DELETE", "/feedings/feeding-123", "", http.StatusNoContent},
		{"GET", "/feedings/last/child-456", "", http.StatusOK},
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
				t.Errorf("%s %s: expected status %d, got %d", tc.method, tc.path, tc.expectedCode, w.Code)
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
// Edge Case Tests
// =====================

func TestCreate_EmptyBody(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_EmptyBody(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("PUT", "/feedings/feeding-123", bytes.NewReader([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreate_MalformedJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	// JSON with trailing comma
	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader([]byte(`{"child_id":"c1",}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_MalformedJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	// JSON with missing closing brace
	req := httptest.NewRequest("PUT", "/feedings/feeding-123", bytes.NewReader([]byte(`{"child_id":"c1"`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestList_MultipleFeedings(t *testing.T) {
	feeding1 := sampleFeeding()
	feeding2 := sampleFeeding()
	feeding2.ID = "feeding-456"
	feeding2.Type = FeedingTypeBreast

	feedings := []Feeding{*feeding1, *feeding2}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *FeedingFilter) ([]Feeding, error) {
			return feedings, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/feedings", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Feeding
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 feedings, got %d", len(result))
	}
}

func TestCreate_WithNilOptionalFields(t *testing.T) {
	var capturedReq *CreateFeedingRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateFeedingRequest) (*Feeding, error) {
			capturedReq = req
			return sampleFeeding(), nil
		},
	}
	router := setupRouter(svc)

	// Request with only required fields, no optional fields
	reqBody := &CreateFeedingRequest{
		ChildID:   "child-456",
		Type:      FeedingTypeFormula,
		StartTime: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/feedings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
	if capturedReq.EndTime != nil {
		t.Error("Expected EndTime to be nil")
	}
	if capturedReq.Amount != nil {
		t.Error("Expected Amount to be nil")
	}
	if capturedReq.Side != "" {
		t.Errorf("Expected Side to be empty, got %s", capturedReq.Side)
	}
}
