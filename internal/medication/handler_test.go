package medication

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
	createFn        func(ctx context.Context, req *CreateMedicationRequest) (*Medication, error)
	getFn           func(ctx context.Context, id string) (*Medication, error)
	listFn          func(ctx context.Context, filter *MedicationFilter) ([]Medication, error)
	updateFn        func(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error)
	deleteFn        func(ctx context.Context, id string) error
	deactivateFn    func(ctx context.Context, id string) error
	logMedicationFn func(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error)
	getLogsFn       func(ctx context.Context, medicationID string) ([]MedicationLog, error)
	getLastLogFn    func(ctx context.Context, medicationID string) (*MedicationLog, error)
}

func (m *mockService) Create(ctx context.Context, req *CreateMedicationRequest) (*Medication, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}

func (m *mockService) Get(ctx context.Context, id string) (*Medication, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) List(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockService) Update(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error) {
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

func (m *mockService) Deactivate(ctx context.Context, id string) error {
	if m.deactivateFn != nil {
		return m.deactivateFn(ctx, id)
	}
	return nil
}

func (m *mockService) LogMedication(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
	if m.logMedicationFn != nil {
		return m.logMedicationFn(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockService) GetLogs(ctx context.Context, medicationID string) ([]MedicationLog, error) {
	if m.getLogsFn != nil {
		return m.getLogsFn(ctx, medicationID)
	}
	return nil, nil
}

func (m *mockService) GetLastLog(ctx context.Context, medicationID string) (*MedicationLog, error) {
	if m.getLastLogFn != nil {
		return m.getLastLogFn(ctx, medicationID)
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

	group := router.Group("/medications")
	handler.RegisterRoutes(group)
	return router
}

// Helper to create a sample medication
func sampleMedication() *Medication {
	return &Medication{
		ID:           "med-123",
		ChildID:      "child-456",
		Name:         "Amoxicillin",
		Dosage:       "250",
		Unit:         "mg",
		Frequency:    "twice_daily",
		Instructions: "Take with food",
		StartDate:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      nil,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Helper to create a sample medication log
func sampleMedicationLog() *MedicationLog {
	now := time.Now()
	return &MedicationLog{
		ID:           "log-123",
		MedicationID: "med-123",
		ChildID:      "child-456",
		GivenAt:      now,
		GivenBy:      "test-user-123",
		Dosage:       "250mg",
		Notes:        "Given with breakfast",
		CreatedAt:    now,
		SyncedAt:     &now,
	}
}

// Helper to create a valid medication request body
func validMedicationRequest() *CreateMedicationRequest {
	return &CreateMedicationRequest{
		ChildID:      "child-456",
		Name:         "Amoxicillin",
		Dosage:       "250",
		Unit:         "mg",
		Frequency:    "twice_daily",
		Instructions: "Take with food",
		StartDate:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Helper to create a valid log medication request body
func validLogMedicationRequest() *LogMedicationRequest {
	return &LogMedicationRequest{
		MedicationID: "med-123",
		GivenAt:      time.Now(),
		Dosage:       "250mg",
		Notes:        "Given with breakfast",
	}
}

// =====================
// List Handler Tests
// =====================

func TestList_Success(t *testing.T) {
	medications := []Medication{*sampleMedication()}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			return medications, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Medication
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 medication, got %d", len(result))
	}
}

func TestList_WithChildIDFilter(t *testing.T) {
	var capturedFilter *MedicationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			capturedFilter = filter
			return []Medication{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications?child_id=child-456", http.NoBody)
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

func TestList_WithActiveOnlyFilter(t *testing.T) {
	var capturedFilter *MedicationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			capturedFilter = filter
			return []Medication{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications?active_only=true", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedFilter == nil {
		t.Fatal("Filter was not passed to service")
	}
	if !capturedFilter.ActiveOnly {
		t.Error("Expected ActiveOnly to be true")
	}
}

func TestList_WithAllFilters(t *testing.T) {
	var capturedFilter *MedicationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			capturedFilter = filter
			return []Medication{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications?child_id=child-456&active_only=true", http.NoBody)
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
	if !capturedFilter.ActiveOnly {
		t.Error("Expected ActiveOnly to be true")
	}
}

func TestList_ActiveOnlyFalseWhenNotTrue(t *testing.T) {
	var capturedFilter *MedicationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			capturedFilter = filter
			return []Medication{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications?active_only=false", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedFilter.ActiveOnly {
		t.Error("Expected ActiveOnly to be false when not 'true'")
	}
}

func TestList_ServiceError(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications", http.NoBody)
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
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			return []Medication{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Medication
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
	med := sampleMedication()
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Medication, error) {
			if id == "med-123" {
				return med, nil
			}
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/med-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Medication
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "med-123" {
		t.Errorf("Expected ID med-123, got %s", result.ID)
	}
}

func TestGet_ServiceError(t *testing.T) {
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Medication, error) {
			return nil, errors.New("medication not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "medication not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGet_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Medication, error) {
			capturedID = id
			return sampleMedication(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/my-specific-id", http.NoBody)
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
	med := sampleMedication()
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateMedicationRequest) (*Medication, error) {
			return med, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validMedicationRequest())
	req := httptest.NewRequest("POST", "/medications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result Medication
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "med-123" {
		t.Errorf("Expected ID med-123, got %s", result.ID)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/medications", bytes.NewReader([]byte("invalid json")))
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

	// Missing required fields (child_id, name, dosage, unit, frequency, start_date)
	body, _ := json.Marshal(map[string]any{
		"instructions": "Take with food",
	})
	req := httptest.NewRequest("POST", "/medications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreate_MissingChildID(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"name":       "Amoxicillin",
		"dosage":     "250",
		"unit":       "mg",
		"frequency":  "twice_daily",
		"start_date": "2025-01-01T00:00:00Z",
	})
	req := httptest.NewRequest("POST", "/medications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing child_id, got %d", w.Code)
	}
}

func TestCreate_MissingName(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id":   "child-456",
		"dosage":     "250",
		"unit":       "mg",
		"frequency":  "twice_daily",
		"start_date": "2025-01-01T00:00:00Z",
	})
	req := httptest.NewRequest("POST", "/medications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing name, got %d", w.Code)
	}
}

func TestCreate_ServiceError(t *testing.T) {
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateMedicationRequest) (*Medication, error) {
			return nil, errors.New("failed to create medication")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validMedicationRequest())
	req := httptest.NewRequest("POST", "/medications", bytes.NewReader(body))
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
	if result["error"] != "failed to create medication" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestCreate_VerifiesRequestData(t *testing.T) {
	var capturedReq *CreateMedicationRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateMedicationRequest) (*Medication, error) {
			capturedReq = req
			return sampleMedication(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validMedicationRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/medications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.ChildID != reqBody.ChildID {
		t.Errorf("Expected ChildID %s, got %s", reqBody.ChildID, capturedReq.ChildID)
	}
	if capturedReq.Name != reqBody.Name {
		t.Errorf("Expected Name %s, got %s", reqBody.Name, capturedReq.Name)
	}
	if capturedReq.Dosage != reqBody.Dosage {
		t.Errorf("Expected Dosage %s, got %s", reqBody.Dosage, capturedReq.Dosage)
	}
	if capturedReq.Unit != reqBody.Unit {
		t.Errorf("Expected Unit %s, got %s", reqBody.Unit, capturedReq.Unit)
	}
	if capturedReq.Frequency != reqBody.Frequency {
		t.Errorf("Expected Frequency %s, got %s", reqBody.Frequency, capturedReq.Frequency)
	}
}

// =====================
// Update Handler Tests
// =====================

func TestUpdate_Success(t *testing.T) {
	med := sampleMedication()
	med.Name = "Updated Medication"
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error) {
			return med, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validMedicationRequest())
	req := httptest.NewRequest("PUT", "/medications/med-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Medication
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Name != "Updated Medication" {
		t.Errorf("Expected Name 'Updated Medication', got %s", result.Name)
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("PUT", "/medications/med-123", bytes.NewReader([]byte("invalid json")))
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
		"instructions": "New instructions",
	})
	req := httptest.NewRequest("PUT", "/medications/med-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error) {
			return nil, errors.New("medication not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validMedicationRequest())
	req := httptest.NewRequest("PUT", "/medications/nonexistent", bytes.NewReader(body))
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
	if result["error"] != "medication not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestUpdate_VerifiesIDAndRequest(t *testing.T) {
	var capturedID string
	var capturedReq *CreateMedicationRequest
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error) {
			capturedID = id
			capturedReq = req
			return sampleMedication(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validMedicationRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/medications/update-id-456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "update-id-456" {
		t.Errorf("Expected ID update-id-456, got %s", capturedID)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.Name != reqBody.Name {
		t.Errorf("Expected Name %s, got %s", reqBody.Name, capturedReq.Name)
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

	req := httptest.NewRequest("DELETE", "/medications/med-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDelete_ServiceError(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("medication not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/medications/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "medication not found" {
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

	req := httptest.NewRequest("DELETE", "/medications/delete-me-789", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "delete-me-789" {
		t.Errorf("Expected ID delete-me-789, got %s", capturedID)
	}
}

// =====================
// Deactivate Handler Tests
// =====================

func TestDeactivate_Success(t *testing.T) {
	svc := &mockService{
		deactivateFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/medications/med-123/deactivate", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeactivate_ServiceError(t *testing.T) {
	svc := &mockService{
		deactivateFn: func(ctx context.Context, id string) error {
			return errors.New("medication not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/medications/nonexistent/deactivate", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "medication not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestDeactivate_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		deactivateFn: func(ctx context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/medications/deactivate-this/deactivate", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "deactivate-this" {
		t.Errorf("Expected ID deactivate-this, got %s", capturedID)
	}
}

// =====================
// LogMedication Handler Tests
// =====================

func TestLogMedication_Success(t *testing.T) {
	log := sampleMedicationLog()
	svc := &mockService{
		logMedicationFn: func(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
			return log, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validLogMedicationRequest())
	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result MedicationLog
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "log-123" {
		t.Errorf("Expected ID log-123, got %s", result.ID)
	}
}

func TestLogMedication_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestLogMedication_MissingRequiredFields(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	// Missing required fields (medication_id, given_at, dosage)
	body, _ := json.Marshal(map[string]any{
		"notes": "Some notes",
	})
	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestLogMedication_MissingMedicationID(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"given_at": time.Now().Format(time.RFC3339),
		"dosage":   "250mg",
	})
	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing medication_id, got %d", w.Code)
	}
}

func TestLogMedication_MissingDosage(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"medication_id": "med-123",
		"given_at":      time.Now().Format(time.RFC3339),
	})
	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing dosage, got %d", w.Code)
	}
}

func TestLogMedication_ServiceError(t *testing.T) {
	svc := &mockService{
		logMedicationFn: func(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
			return nil, errors.New("medication not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validLogMedicationRequest())
	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader(body))
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
	if result["error"] != "medication not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestLogMedication_VerifiesUserIDAndRequest(t *testing.T) {
	var capturedUserID string
	var capturedReq *LogMedicationRequest
	svc := &mockService{
		logMedicationFn: func(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
			capturedUserID = userID
			capturedReq = req
			return sampleMedicationLog(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validLogMedicationRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "test-user-123" {
		t.Errorf("Expected userID test-user-123, got %s", capturedUserID)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.MedicationID != reqBody.MedicationID {
		t.Errorf("Expected MedicationID %s, got %s", reqBody.MedicationID, capturedReq.MedicationID)
	}
	if capturedReq.Dosage != reqBody.Dosage {
		t.Errorf("Expected Dosage %s, got %s", reqBody.Dosage, capturedReq.Dosage)
	}
}

// =====================
// GetLogs Handler Tests
// =====================

func TestGetLogs_Success(t *testing.T) {
	logs := []MedicationLog{*sampleMedicationLog()}
	svc := &mockService{
		getLogsFn: func(ctx context.Context, medicationID string) ([]MedicationLog, error) {
			return logs, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/med-123/logs", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []MedicationLog
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 log, got %d", len(result))
	}
}

func TestGetLogs_ServiceError(t *testing.T) {
	svc := &mockService{
		getLogsFn: func(ctx context.Context, medicationID string) ([]MedicationLog, error) {
			return nil, errors.New("database error")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/med-123/logs", http.NoBody)
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

func TestGetLogs_VerifiesMedicationIDParam(t *testing.T) {
	var capturedMedicationID string
	svc := &mockService{
		getLogsFn: func(ctx context.Context, medicationID string) ([]MedicationLog, error) {
			capturedMedicationID = medicationID
			return []MedicationLog{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/specific-med-id/logs", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedMedicationID != "specific-med-id" {
		t.Errorf("Expected medicationID specific-med-id, got %s", capturedMedicationID)
	}
}

func TestGetLogs_EmptyResult(t *testing.T) {
	svc := &mockService{
		getLogsFn: func(ctx context.Context, medicationID string) ([]MedicationLog, error) {
			return []MedicationLog{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/med-123/logs", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []MedicationLog
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty list, got %d", len(result))
	}
}

// =====================
// GetLastLog Handler Tests
// =====================

func TestGetLastLog_Success(t *testing.T) {
	log := sampleMedicationLog()
	svc := &mockService{
		getLastLogFn: func(ctx context.Context, medicationID string) (*MedicationLog, error) {
			return log, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/med-123/logs/last", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result MedicationLog
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "log-123" {
		t.Errorf("Expected ID log-123, got %s", result.ID)
	}
}

func TestGetLastLog_ServiceError(t *testing.T) {
	svc := &mockService{
		getLastLogFn: func(ctx context.Context, medicationID string) (*MedicationLog, error) {
			return nil, errors.New("no logs found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/med-123/logs/last", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "no logs found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGetLastLog_VerifiesMedicationIDParam(t *testing.T) {
	var capturedMedicationID string
	svc := &mockService{
		getLastLogFn: func(ctx context.Context, medicationID string) (*MedicationLog, error) {
			capturedMedicationID = medicationID
			return sampleMedicationLog(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/specific-med-id/logs/last", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedMedicationID != "specific-med-id" {
		t.Errorf("Expected medicationID specific-med-id, got %s", capturedMedicationID)
	}
}

// =====================
// Route Registration Tests
// =====================

func TestRegisterRoutes(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			return []Medication{}, nil
		},
		getFn: func(ctx context.Context, id string) (*Medication, error) {
			return sampleMedication(), nil
		},
		createFn: func(ctx context.Context, req *CreateMedicationRequest) (*Medication, error) {
			return sampleMedication(), nil
		},
		updateFn: func(ctx context.Context, id string, req *CreateMedicationRequest) (*Medication, error) {
			return sampleMedication(), nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
		deactivateFn: func(ctx context.Context, id string) error {
			return nil
		},
		logMedicationFn: func(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
			return sampleMedicationLog(), nil
		},
		getLogsFn: func(ctx context.Context, medicationID string) ([]MedicationLog, error) {
			return []MedicationLog{}, nil
		},
		getLastLogFn: func(ctx context.Context, medicationID string) (*MedicationLog, error) {
			return sampleMedicationLog(), nil
		},
	}
	router := setupRouter(svc)

	testCases := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"GET", "/medications", "", http.StatusOK},
		{"GET", "/medications/med-123", "", http.StatusOK},
		{"POST", "/medications", `{"child_id":"c1","name":"Med","dosage":"250","unit":"mg","frequency":"daily","start_date":"2025-01-01T00:00:00Z"}`, http.StatusCreated},
		{"PUT", "/medications/med-123", `{"child_id":"c1","name":"Med","dosage":"250","unit":"mg","frequency":"daily","start_date":"2025-01-01T00:00:00Z"}`, http.StatusOK},
		{"DELETE", "/medications/med-123", "", http.StatusNoContent},
		{"POST", "/medications/med-123/deactivate", "", http.StatusOK},
		{"POST", "/medications/log", `{"medication_id":"med-123","given_at":"2025-01-01T10:00:00Z","dosage":"250mg"}`, http.StatusCreated},
		{"GET", "/medications/med-123/logs", "", http.StatusOK},
		{"GET", "/medications/med-123/logs/last", "", http.StatusOK},
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
// Edge Case Tests
// =====================

func TestCreate_WithOptionalFields(t *testing.T) {
	var capturedReq *CreateMedicationRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateMedicationRequest) (*Medication, error) {
			capturedReq = req
			return sampleMedication(), nil
		},
	}
	router := setupRouter(svc)

	endDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	reqBody := &CreateMedicationRequest{
		ChildID:      "child-456",
		Name:         "Amoxicillin",
		Dosage:       "250",
		Unit:         "mg",
		Frequency:    "twice_daily",
		Instructions: "Take with food and plenty of water",
		StartDate:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      &endDate,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/medications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if capturedReq.Instructions != "Take with food and plenty of water" {
		t.Errorf("Expected Instructions to be set, got %s", capturedReq.Instructions)
	}
	if capturedReq.EndDate == nil {
		t.Error("Expected EndDate to be set")
	}
}

func TestLogMedication_WithOptionalNotes(t *testing.T) {
	var capturedReq *LogMedicationRequest
	svc := &mockService{
		logMedicationFn: func(ctx context.Context, userID string, req *LogMedicationRequest) (*MedicationLog, error) {
			capturedReq = req
			return sampleMedicationLog(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := &LogMedicationRequest{
		MedicationID: "med-123",
		GivenAt:      time.Now(),
		Dosage:       "250mg",
		Notes:        "Child was feeling better after dose",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/medications/log", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedReq.Notes != "Child was feeling better after dose" {
		t.Errorf("Expected Notes to be set, got %s", capturedReq.Notes)
	}
}

func TestList_MultipleResults(t *testing.T) {
	med1 := sampleMedication()
	med2 := sampleMedication()
	med2.ID = "med-456"
	med2.Name = "Ibuprofen"

	medications := []Medication{*med1, *med2}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *MedicationFilter) ([]Medication, error) {
			return medications, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Medication
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 medications, got %d", len(result))
	}
	if result[0].ID != "med-123" {
		t.Errorf("Expected first medication ID med-123, got %s", result[0].ID)
	}
	if result[1].ID != "med-456" {
		t.Errorf("Expected second medication ID med-456, got %s", result[1].ID)
	}
}

func TestGetLogs_MultipleResults(t *testing.T) {
	log1 := sampleMedicationLog()
	log2 := sampleMedicationLog()
	log2.ID = "log-456"
	log2.GivenAt = time.Now().Add(-24 * time.Hour)

	logs := []MedicationLog{*log1, *log2}
	svc := &mockService{
		getLogsFn: func(ctx context.Context, medicationID string) ([]MedicationLog, error) {
			return logs, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/medications/med-123/logs", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []MedicationLog
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(result))
	}
}
