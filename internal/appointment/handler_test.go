package appointment

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
	createFn      func(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error)
	getFn         func(ctx context.Context, id string) (*Appointment, error)
	listFn        func(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error)
	updateFn      func(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error)
	deleteFn      func(ctx context.Context, id string) error
	completeFn    func(ctx context.Context, id string) error
	cancelFn      func(ctx context.Context, id string) error
	getUpcomingFn func(ctx context.Context, childID string, days int) ([]Appointment, error)
}

func (m *mockService) Create(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}

func (m *mockService) Get(ctx context.Context, id string) (*Appointment, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) List(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockService) Update(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error) {
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

func (m *mockService) Complete(ctx context.Context, id string) error {
	if m.completeFn != nil {
		return m.completeFn(ctx, id)
	}
	return nil
}

func (m *mockService) Cancel(ctx context.Context, id string) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, id)
	}
	return nil
}

func (m *mockService) GetUpcoming(ctx context.Context, childID string, days int) ([]Appointment, error) {
	if m.getUpcomingFn != nil {
		return m.getUpcomingFn(ctx, childID, days)
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

	group := router.Group("/appointments")
	handler.RegisterRoutes(group)
	return router
}

// Helper to create a sample appointment
func sampleAppointment() *Appointment {
	return &Appointment{
		ID:          "apt-123",
		ChildID:     "child-456",
		Type:        AppointmentTypeWellVisit,
		Title:       "Annual Checkup",
		Provider:    "Dr. Smith",
		Location:    "123 Medical Center",
		ScheduledAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		Duration:    30,
		Notes:       "Bring immunization records",
		Completed:   false,
		Canceled:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Helper to create a valid request body
func validRequestBody() *CreateAppointmentRequest {
	return &CreateAppointmentRequest{
		ChildID:     "child-456",
		Type:        AppointmentTypeWellVisit,
		Title:       "Annual Checkup",
		Provider:    "Dr. Smith",
		Location:    "123 Medical Center",
		ScheduledAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		Duration:    30,
		Notes:       "Bring immunization records",
	}
}

// =====================
// List Handler Tests
// =====================

func TestList_Success(t *testing.T) {
	appointments := []Appointment{*sampleAppointment()}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
			return appointments, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Appointment
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 appointment, got %d", len(result))
	}
}

func TestList_WithFilter(t *testing.T) {
	var capturedFilter *AppointmentFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
			capturedFilter = filter
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments?child_id=child-456&upcoming_only=true", http.NoBody)
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
	if !capturedFilter.UpcomingOnly {
		t.Error("Expected UpcomingOnly to be true")
	}
}

func TestList_ServiceError(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments", http.NoBody)
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
		listFn: func(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Appointment
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
	apt := sampleAppointment()
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Appointment, error) {
			if id == "apt-123" {
				return apt, nil
			}
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/apt-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Appointment
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "apt-123" {
		t.Errorf("Expected ID apt-123, got %s", result.ID)
	}
}

func TestGet_ServiceError(t *testing.T) {
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Appointment, error) {
			return nil, errors.New("appointment not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "appointment not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGet_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Appointment, error) {
			capturedID = id
			return sampleAppointment(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/my-specific-id", http.NoBody)
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
	apt := sampleAppointment()
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error) {
			return apt, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result Appointment
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "apt-123" {
		t.Errorf("Expected ID apt-123, got %s", result.ID)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/appointments", bytes.NewReader([]byte("invalid json")))
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

	// Missing required fields (child_id, type, title, scheduled_at)
	body, _ := json.Marshal(map[string]any{
		"provider": "Dr. Smith",
	})
	req := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreate_ServiceError(t *testing.T) {
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error) {
			return nil, errors.New("failed to create appointment")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
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
	if result["error"] != "failed to create appointment" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestCreate_VerifiesRequestData(t *testing.T) {
	var capturedReq *CreateAppointmentRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error) {
			capturedReq = req
			return sampleAppointment(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validRequestBody()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/appointments", bytes.NewReader(body))
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
	if capturedReq.Title != reqBody.Title {
		t.Errorf("Expected Title %s, got %s", reqBody.Title, capturedReq.Title)
	}
}

// =====================
// Update Handler Tests
// =====================

func TestUpdate_Success(t *testing.T) {
	apt := sampleAppointment()
	apt.Title = "Updated Checkup"
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error) {
			return apt, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("PUT", "/appointments/apt-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Appointment
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Title != "Updated Checkup" {
		t.Errorf("Expected Title 'Updated Checkup', got %s", result.Title)
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("PUT", "/appointments/apt-123", bytes.NewReader([]byte("invalid json")))
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
		"provider": "Dr. Jones",
	})
	req := httptest.NewRequest("PUT", "/appointments/apt-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error) {
			return nil, errors.New("appointment not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRequestBody())
	req := httptest.NewRequest("PUT", "/appointments/nonexistent", bytes.NewReader(body))
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
	if result["error"] != "appointment not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestUpdate_VerifiesIDAndRequest(t *testing.T) {
	var capturedID string
	var capturedReq *CreateAppointmentRequest
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error) {
			capturedID = id
			capturedReq = req
			return sampleAppointment(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validRequestBody()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/appointments/update-id-456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "update-id-456" {
		t.Errorf("Expected ID update-id-456, got %s", capturedID)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.Title != reqBody.Title {
		t.Errorf("Expected Title %s, got %s", reqBody.Title, capturedReq.Title)
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

	req := httptest.NewRequest("DELETE", "/appointments/apt-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDelete_ServiceError(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("appointment not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/appointments/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "appointment not found" {
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

	req := httptest.NewRequest("DELETE", "/appointments/delete-me-789", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "delete-me-789" {
		t.Errorf("Expected ID delete-me-789, got %s", capturedID)
	}
}

// =====================
// Complete Handler Tests
// =====================

func TestComplete_Success(t *testing.T) {
	svc := &mockService{
		completeFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/appointments/apt-123/complete", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestComplete_ServiceError(t *testing.T) {
	svc := &mockService{
		completeFn: func(ctx context.Context, id string) error {
			return errors.New("appointment not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/appointments/nonexistent/complete", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "appointment not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestComplete_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		completeFn: func(ctx context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/appointments/complete-this/complete", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "complete-this" {
		t.Errorf("Expected ID complete-this, got %s", capturedID)
	}
}

// =====================
// Cancel Handler Tests
// =====================

func TestCancel_Success(t *testing.T) {
	svc := &mockService{
		cancelFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/appointments/apt-123/cancel", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCancel_ServiceError(t *testing.T) {
	svc := &mockService{
		cancelFn: func(ctx context.Context, id string) error {
			return errors.New("appointment not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/appointments/nonexistent/cancel", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "appointment not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestCancel_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		cancelFn: func(ctx context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/appointments/cancel-this/cancel", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "cancel-this" {
		t.Errorf("Expected ID cancel-this, got %s", capturedID)
	}
}

// =====================
// GetUpcoming Handler Tests
// =====================

func TestGetUpcoming_Success(t *testing.T) {
	appointments := []Appointment{*sampleAppointment()}
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			return appointments, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/upcoming/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Appointment
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 appointment, got %d", len(result))
	}
}

func TestGetUpcoming_DefaultDays(t *testing.T) {
	var capturedDays int
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			capturedDays = days
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/upcoming/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedDays != 30 {
		t.Errorf("Expected default days 30, got %d", capturedDays)
	}
}

func TestGetUpcoming_CustomDays(t *testing.T) {
	var capturedDays int
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			capturedDays = days
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/upcoming/child-456?days=7", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedDays != 7 {
		t.Errorf("Expected days 7, got %d", capturedDays)
	}
}

func TestGetUpcoming_InvalidDaysParam(t *testing.T) {
	var capturedDays int
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			capturedDays = days
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	// Invalid days parameter should default to 30
	req := httptest.NewRequest("GET", "/appointments/upcoming/child-456?days=invalid", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedDays != 30 {
		t.Errorf("Expected default days 30 for invalid param, got %d", capturedDays)
	}
}

func TestGetUpcoming_ServiceError(t *testing.T) {
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			return nil, errors.New("database error")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/upcoming/child-456", http.NoBody)
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

func TestGetUpcoming_VerifiesChildIDParam(t *testing.T) {
	var capturedChildID string
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			capturedChildID = childID
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/upcoming/specific-child-id", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedChildID != "specific-child-id" {
		t.Errorf("Expected childID specific-child-id, got %s", capturedChildID)
	}
}

func TestGetUpcoming_EmptyResult(t *testing.T) {
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/appointments/upcoming/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Appointment
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty list, got %d", len(result))
	}
}

// =====================
// Route Registration Tests
// =====================

func TestRegisterRoutes(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *AppointmentFilter) ([]Appointment, error) {
			return []Appointment{}, nil
		},
		getFn: func(ctx context.Context, id string) (*Appointment, error) {
			return sampleAppointment(), nil
		},
		createFn: func(ctx context.Context, req *CreateAppointmentRequest) (*Appointment, error) {
			return sampleAppointment(), nil
		},
		updateFn: func(ctx context.Context, id string, req *CreateAppointmentRequest) (*Appointment, error) {
			return sampleAppointment(), nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
		completeFn: func(ctx context.Context, id string) error {
			return nil
		},
		cancelFn: func(ctx context.Context, id string) error {
			return nil
		},
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Appointment, error) {
			return []Appointment{}, nil
		},
	}
	router := setupRouter(svc)

	testCases := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"GET", "/appointments", "", http.StatusOK},
		{"GET", "/appointments/apt-123", "", http.StatusOK},
		{"POST", "/appointments", `{"child_id":"c1","type":"well_visit","title":"Test","scheduled_at":"2025-01-15T10:00:00Z"}`, http.StatusCreated},
		{"PUT", "/appointments/apt-123", `{"child_id":"c1","type":"well_visit","title":"Test","scheduled_at":"2025-01-15T10:00:00Z"}`, http.StatusOK},
		{"DELETE", "/appointments/apt-123", "", http.StatusNoContent},
		{"POST", "/appointments/apt-123/complete", "", http.StatusOK},
		{"POST", "/appointments/apt-123/cancel", "", http.StatusOK},
		{"GET", "/appointments/upcoming/child-456", "", http.StatusOK},
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
