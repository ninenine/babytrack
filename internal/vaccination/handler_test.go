package vaccination

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
	createFn                   func(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error)
	getFn                      func(ctx context.Context, id string) (*Vaccination, error)
	listFn                     func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error)
	updateFn                   func(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error)
	deleteFn                   func(ctx context.Context, id string) error
	recordAdministrationFn     func(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error)
	getUpcomingFn              func(ctx context.Context, childID string, days int) ([]Vaccination, error)
	getScheduleFn              func() []VaccinationSchedule
	generateScheduleForChildFn func(ctx context.Context, childID string, birthDate string) ([]Vaccination, error)
}

func (m *mockService) Create(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}

func (m *mockService) Get(ctx context.Context, id string) (*Vaccination, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) List(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockService) Update(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error) {
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

func (m *mockService) RecordAdministration(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
	if m.recordAdministrationFn != nil {
		return m.recordAdministrationFn(ctx, id, req)
	}
	return nil, nil
}

func (m *mockService) GetUpcoming(ctx context.Context, childID string, days int) ([]Vaccination, error) {
	if m.getUpcomingFn != nil {
		return m.getUpcomingFn(ctx, childID, days)
	}
	return nil, nil
}

func (m *mockService) GetSchedule() []VaccinationSchedule {
	if m.getScheduleFn != nil {
		return m.getScheduleFn()
	}
	return nil
}

func (m *mockService) GenerateScheduleForChild(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
	if m.generateScheduleForChildFn != nil {
		return m.generateScheduleForChildFn(ctx, childID, birthDate)
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

	group := router.Group("/vaccinations")
	handler.RegisterRoutes(group)
	return router
}

// Helper to create a sample vaccination
func sampleVaccination() *Vaccination {
	now := time.Now()
	return &Vaccination{
		ID:          "vax-123",
		ChildID:     "child-456",
		Name:        "DTaP",
		Dose:        1,
		ScheduledAt: time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC),
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Helper to create a completed vaccination
func completedVaccination() *Vaccination {
	now := time.Now()
	administeredAt := time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)
	return &Vaccination{
		ID:             "vax-123",
		ChildID:        "child-456",
		Name:           "DTaP",
		Dose:           1,
		ScheduledAt:    time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC),
		AdministeredAt: &administeredAt,
		Provider:       "Dr. Smith",
		Location:       "Pediatric Clinic",
		LotNumber:      "LOT123ABC",
		Notes:          "No reactions observed",
		Completed:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Helper to create a sample vaccination schedule
func sampleVaccinationSchedule() []VaccinationSchedule {
	return []VaccinationSchedule{
		{
			ID:          "sched-1",
			Name:        "DTaP",
			Description: "Diphtheria, Tetanus, Pertussis",
			AgeWeeks:    8,
			AgeMonths:   2,
			AgeLabel:    "2 months",
			Dose:        1,
		},
		{
			ID:          "sched-2",
			Name:        "DTaP",
			Description: "Diphtheria, Tetanus, Pertussis",
			AgeWeeks:    16,
			AgeMonths:   4,
			AgeLabel:    "4 months",
			Dose:        2,
		},
	}
}

// Helper to create a valid vaccination request body
func validVaccinationRequest() *CreateVaccinationRequest {
	return &CreateVaccinationRequest{
		ChildID:     "child-456",
		Name:        "DTaP",
		Dose:        1,
		ScheduledAt: time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC),
	}
}

// Helper to create a valid record vaccination request body
func validRecordVaccinationRequest() *RecordVaccinationRequest {
	return &RecordVaccinationRequest{
		AdministeredAt: time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC),
		Provider:       "Dr. Smith",
		Location:       "Pediatric Clinic",
		LotNumber:      "LOT123ABC",
		Notes:          "No reactions observed",
	}
}

// =====================
// List Handler Tests
// =====================

func TestList_Success(t *testing.T) {
	vaccinations := []Vaccination{*sampleVaccination()}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			return vaccinations, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 vaccination, got %d", len(result))
	}
}

func TestList_WithChildIDFilter(t *testing.T) {
	var capturedFilter *VaccinationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			capturedFilter = filter
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations?child_id=child-456", http.NoBody)
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

func TestList_WithCompletedTrueFilter(t *testing.T) {
	var capturedFilter *VaccinationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			capturedFilter = filter
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations?completed=true", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedFilter == nil {
		t.Fatal("Filter was not passed to service")
	}
	if capturedFilter.Completed == nil {
		t.Fatal("Expected Completed to be set")
	}
	if *capturedFilter.Completed != true {
		t.Error("Expected Completed to be true")
	}
}

func TestList_WithCompletedFalseFilter(t *testing.T) {
	var capturedFilter *VaccinationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			capturedFilter = filter
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations?completed=false", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedFilter == nil {
		t.Fatal("Filter was not passed to service")
	}
	if capturedFilter.Completed == nil {
		t.Fatal("Expected Completed to be set")
	}
	if *capturedFilter.Completed != false {
		t.Error("Expected Completed to be false")
	}
}

func TestList_WithUpcomingOnlyFilter(t *testing.T) {
	var capturedFilter *VaccinationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			capturedFilter = filter
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations?upcoming_only=true", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedFilter == nil {
		t.Fatal("Filter was not passed to service")
	}
	if !capturedFilter.UpcomingOnly {
		t.Error("Expected UpcomingOnly to be true")
	}
}

func TestList_WithAllFilters(t *testing.T) {
	var capturedFilter *VaccinationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			capturedFilter = filter
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations?child_id=child-456&completed=false&upcoming_only=true", http.NoBody)
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
	if capturedFilter.Completed == nil || *capturedFilter.Completed != false {
		t.Error("Expected Completed to be false")
	}
	if !capturedFilter.UpcomingOnly {
		t.Error("Expected UpcomingOnly to be true")
	}
}

func TestList_CompletedNilWhenNotProvided(t *testing.T) {
	var capturedFilter *VaccinationFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			capturedFilter = filter
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedFilter.Completed != nil {
		t.Error("Expected Completed to be nil when not provided")
	}
}

func TestList_ServiceError(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations", http.NoBody)
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
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Vaccination
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
	vax := sampleVaccination()
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Vaccination, error) {
			if id == "vax-123" {
				return vax, nil
			}
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/vax-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "vax-123" {
		t.Errorf("Expected ID vax-123, got %s", result.ID)
	}
}

func TestGet_ServiceError(t *testing.T) {
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Vaccination, error) {
			return nil, errors.New("vaccination not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "vaccination not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGet_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Vaccination, error) {
			capturedID = id
			return sampleVaccination(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/my-specific-id", http.NoBody)
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
	vax := sampleVaccination()
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error) {
			return vax, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validVaccinationRequest())
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "vax-123" {
		t.Errorf("Expected ID vax-123, got %s", result.ID)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader([]byte("invalid json")))
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

	// Missing required fields (child_id, name, dose, scheduled_at)
	body, _ := json.Marshal(map[string]any{
		"notes": "Some notes",
	})
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
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
		"name":         "DTaP",
		"dose":         1,
		"scheduled_at": "2025-03-15T10:00:00Z",
	})
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
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
		"child_id":     "child-456",
		"dose":         1,
		"scheduled_at": "2025-03-15T10:00:00Z",
	})
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing name, got %d", w.Code)
	}
}

func TestCreate_MissingDose(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id":     "child-456",
		"name":         "DTaP",
		"scheduled_at": "2025-03-15T10:00:00Z",
	})
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing dose, got %d", w.Code)
	}
}

func TestCreate_MissingScheduledAt(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "child-456",
		"name":     "DTaP",
		"dose":     1,
	})
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing scheduled_at, got %d", w.Code)
	}
}

func TestCreate_ServiceError(t *testing.T) {
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error) {
			return nil, errors.New("failed to create vaccination")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validVaccinationRequest())
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
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
	if result["error"] != "failed to create vaccination" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestCreate_VerifiesRequestData(t *testing.T) {
	var capturedReq *CreateVaccinationRequest
	svc := &mockService{
		createFn: func(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error) {
			capturedReq = req
			return sampleVaccination(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validVaccinationRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
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
	if capturedReq.Dose != reqBody.Dose {
		t.Errorf("Expected Dose %d, got %d", reqBody.Dose, capturedReq.Dose)
	}
}

// =====================
// Update Handler Tests
// =====================

func TestUpdate_Success(t *testing.T) {
	vax := sampleVaccination()
	vax.Name = "Updated Vaccination"
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error) {
			return vax, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validVaccinationRequest())
	req := httptest.NewRequest("PUT", "/vaccinations/vax-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Name != "Updated Vaccination" {
		t.Errorf("Expected Name 'Updated Vaccination', got %s", result.Name)
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("PUT", "/vaccinations/vax-123", bytes.NewReader([]byte("invalid json")))
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
		"notes": "New notes",
	})
	req := httptest.NewRequest("PUT", "/vaccinations/vax-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error) {
			return nil, errors.New("vaccination not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validVaccinationRequest())
	req := httptest.NewRequest("PUT", "/vaccinations/nonexistent", bytes.NewReader(body))
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
	if result["error"] != "vaccination not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestUpdate_VerifiesIDAndRequest(t *testing.T) {
	var capturedID string
	var capturedReq *CreateVaccinationRequest
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error) {
			capturedID = id
			capturedReq = req
			return sampleVaccination(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validVaccinationRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/vaccinations/update-id-456", bytes.NewReader(body))
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

	req := httptest.NewRequest("DELETE", "/vaccinations/vax-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDelete_ServiceError(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("vaccination not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/vaccinations/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "vaccination not found" {
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

	req := httptest.NewRequest("DELETE", "/vaccinations/delete-me-789", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "delete-me-789" {
		t.Errorf("Expected ID delete-me-789, got %s", capturedID)
	}
}

// =====================
// RecordAdministration Handler Tests
// =====================

func TestRecordAdministration_Success(t *testing.T) {
	vax := completedVaccination()
	svc := &mockService{
		recordAdministrationFn: func(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
			return vax, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRecordVaccinationRequest())
	req := httptest.NewRequest("POST", "/vaccinations/vax-123/record", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if !result.Completed {
		t.Error("Expected vaccination to be marked as completed")
	}
	if result.Provider != "Dr. Smith" {
		t.Errorf("Expected Provider 'Dr. Smith', got %s", result.Provider)
	}
}

func TestRecordAdministration_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/vaccinations/vax-123/record", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRecordAdministration_MissingRequiredFields(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	// Missing required administered_at field
	body, _ := json.Marshal(map[string]any{
		"provider": "Dr. Smith",
		"notes":    "Some notes",
	})
	req := httptest.NewRequest("POST", "/vaccinations/vax-123/record", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRecordAdministration_MissingAdministeredAt(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"provider":   "Dr. Smith",
		"location":   "Clinic",
		"lot_number": "LOT123",
	})
	req := httptest.NewRequest("POST", "/vaccinations/vax-123/record", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing administered_at, got %d", w.Code)
	}
}

func TestRecordAdministration_ServiceError(t *testing.T) {
	svc := &mockService{
		recordAdministrationFn: func(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
			return nil, errors.New("vaccination not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validRecordVaccinationRequest())
	req := httptest.NewRequest("POST", "/vaccinations/nonexistent/record", bytes.NewReader(body))
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
	if result["error"] != "vaccination not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestRecordAdministration_VerifiesIDAndRequest(t *testing.T) {
	var capturedID string
	var capturedReq *RecordVaccinationRequest
	svc := &mockService{
		recordAdministrationFn: func(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
			capturedID = id
			capturedReq = req
			return completedVaccination(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validRecordVaccinationRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/vaccinations/record-id-123/record", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "record-id-123" {
		t.Errorf("Expected ID record-id-123, got %s", capturedID)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.Provider != reqBody.Provider {
		t.Errorf("Expected Provider %s, got %s", reqBody.Provider, capturedReq.Provider)
	}
	if capturedReq.Location != reqBody.Location {
		t.Errorf("Expected Location %s, got %s", reqBody.Location, capturedReq.Location)
	}
	if capturedReq.LotNumber != reqBody.LotNumber {
		t.Errorf("Expected LotNumber %s, got %s", reqBody.LotNumber, capturedReq.LotNumber)
	}
}

func TestRecordAdministration_WithOptionalFieldsOnly(t *testing.T) {
	var capturedReq *RecordVaccinationRequest
	svc := &mockService{
		recordAdministrationFn: func(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
			capturedReq = req
			return completedVaccination(), nil
		},
	}
	router := setupRouter(svc)

	// Only required field (administered_at) provided
	reqBody := &RecordVaccinationRequest{
		AdministeredAt: time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/vaccinations/vax-123/record", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedReq.Provider != "" {
		t.Errorf("Expected empty Provider, got %s", capturedReq.Provider)
	}
	if capturedReq.Location != "" {
		t.Errorf("Expected empty Location, got %s", capturedReq.Location)
	}
}

// =====================
// GetUpcoming Handler Tests
// =====================

func TestGetUpcoming_Success(t *testing.T) {
	vaccinations := []Vaccination{*sampleVaccination()}
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			return vaccinations, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 vaccination, got %d", len(result))
	}
}

func TestGetUpcoming_DefaultDays(t *testing.T) {
	var capturedDays int
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			capturedDays = days
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedDays != 30 {
		t.Errorf("Expected default days to be 30, got %d", capturedDays)
	}
}

func TestGetUpcoming_CustomDays(t *testing.T) {
	var capturedDays int
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			capturedDays = days
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/child-456?days=60", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedDays != 60 {
		t.Errorf("Expected days to be 60, got %d", capturedDays)
	}
}

func TestGetUpcoming_InvalidDays(t *testing.T) {
	var capturedDays int
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			capturedDays = days
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/child-456?days=invalid", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fall back to default when invalid
	if capturedDays != 30 {
		t.Errorf("Expected default days 30 for invalid input, got %d", capturedDays)
	}
}

func TestGetUpcoming_VerifiesChildIDParam(t *testing.T) {
	var capturedChildID string
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			capturedChildID = childID
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/specific-child-id", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedChildID != "specific-child-id" {
		t.Errorf("Expected childID specific-child-id, got %s", capturedChildID)
	}
}

func TestGetUpcoming_ServiceError(t *testing.T) {
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			return nil, errors.New("database error")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/child-456", http.NoBody)
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

func TestGetUpcoming_EmptyResult(t *testing.T) {
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty list, got %d", len(result))
	}
}

// =====================
// GetSchedule Handler Tests
// =====================

func TestGetSchedule_Success(t *testing.T) {
	schedule := sampleVaccinationSchedule()
	svc := &mockService{
		getScheduleFn: func() []VaccinationSchedule {
			return schedule
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/schedule", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []VaccinationSchedule
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 schedule items, got %d", len(result))
	}
	if result[0].Name != "DTaP" {
		t.Errorf("Expected Name 'DTaP', got %s", result[0].Name)
	}
}

func TestGetSchedule_EmptySchedule(t *testing.T) {
	svc := &mockService{
		getScheduleFn: func() []VaccinationSchedule {
			return []VaccinationSchedule{}
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/schedule", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []VaccinationSchedule
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty schedule, got %d", len(result))
	}
}

func TestGetSchedule_NilSchedule(t *testing.T) {
	svc := &mockService{
		getScheduleFn: func() []VaccinationSchedule {
			return nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/schedule", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// =====================
// GenerateSchedule Handler Tests
// =====================

func TestGenerateSchedule_Success(t *testing.T) {
	vaccinations := []Vaccination{*sampleVaccination()}
	svc := &mockService{
		generateScheduleForChildFn: func(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
			return vaccinations, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]string{
		"birth_date": "2025-01-01",
	})
	req := httptest.NewRequest("POST", "/vaccinations/generate/child-456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result []Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 vaccination, got %d", len(result))
	}
}

func TestGenerateSchedule_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/vaccinations/generate/child-456", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGenerateSchedule_MissingBirthDate(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest("POST", "/vaccinations/generate/child-456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing birth_date, got %d", w.Code)
	}
}

func TestGenerateSchedule_ServiceError(t *testing.T) {
	svc := &mockService{
		generateScheduleForChildFn: func(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
			return nil, errors.New("invalid birth date")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]string{
		"birth_date": "invalid-date",
	})
	req := httptest.NewRequest("POST", "/vaccinations/generate/child-456", bytes.NewReader(body))
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
	if result["error"] != "invalid birth date" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGenerateSchedule_VerifiesChildIDAndBirthDate(t *testing.T) {
	var capturedChildID string
	var capturedBirthDate string
	svc := &mockService{
		generateScheduleForChildFn: func(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
			capturedChildID = childID
			capturedBirthDate = birthDate
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]string{
		"birth_date": "2025-01-15",
	})
	req := httptest.NewRequest("POST", "/vaccinations/generate/specific-child-789", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedChildID != "specific-child-789" {
		t.Errorf("Expected childID specific-child-789, got %s", capturedChildID)
	}
	if capturedBirthDate != "2025-01-15" {
		t.Errorf("Expected birthDate 2025-01-15, got %s", capturedBirthDate)
	}
}

func TestGenerateSchedule_MultipleVaccinations(t *testing.T) {
	vax1 := sampleVaccination()
	vax2 := sampleVaccination()
	vax2.ID = "vax-456"
	vax2.Name = "Polio"
	vax2.Dose = 1
	vax3 := sampleVaccination()
	vax3.ID = "vax-789"
	vax3.Name = "HepB"
	vax3.Dose = 1

	vaccinations := []Vaccination{*vax1, *vax2, *vax3}
	svc := &mockService{
		generateScheduleForChildFn: func(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
			return vaccinations, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]string{
		"birth_date": "2025-01-01",
	})
	req := httptest.NewRequest("POST", "/vaccinations/generate/child-456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result []Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 vaccinations, got %d", len(result))
	}
}

// =====================
// Route Registration Tests
// =====================

func TestRegisterRoutes(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			return []Vaccination{}, nil
		},
		getFn: func(ctx context.Context, id string) (*Vaccination, error) {
			return sampleVaccination(), nil
		},
		createFn: func(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error) {
			return sampleVaccination(), nil
		},
		updateFn: func(ctx context.Context, id string, req *CreateVaccinationRequest) (*Vaccination, error) {
			return sampleVaccination(), nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
		recordAdministrationFn: func(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
			return completedVaccination(), nil
		},
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			return []Vaccination{}, nil
		},
		getScheduleFn: func() []VaccinationSchedule {
			return []VaccinationSchedule{}
		},
		generateScheduleForChildFn: func(ctx context.Context, childID string, birthDate string) ([]Vaccination, error) {
			return []Vaccination{}, nil
		},
	}
	router := setupRouter(svc)

	testCases := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"GET", "/vaccinations", "", http.StatusOK},
		{"GET", "/vaccinations/vax-123", "", http.StatusOK},
		{"POST", "/vaccinations", `{"child_id":"c1","name":"DTaP","dose":1,"scheduled_at":"2025-03-15T10:00:00Z"}`, http.StatusCreated},
		{"PUT", "/vaccinations/vax-123", `{"child_id":"c1","name":"DTaP","dose":1,"scheduled_at":"2025-03-15T10:00:00Z"}`, http.StatusOK},
		{"DELETE", "/vaccinations/vax-123", "", http.StatusNoContent},
		{"POST", "/vaccinations/vax-123/record", `{"administered_at":"2025-03-15T10:30:00Z"}`, http.StatusOK},
		{"GET", "/vaccinations/upcoming/child-456", "", http.StatusOK},
		{"GET", "/vaccinations/schedule", "", http.StatusOK},
		{"POST", "/vaccinations/generate/child-456", `{"birth_date":"2025-01-01"}`, http.StatusCreated},
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

func TestList_MultipleResults(t *testing.T) {
	vax1 := sampleVaccination()
	vax2 := sampleVaccination()
	vax2.ID = "vax-456"
	vax2.Name = "Polio"

	vaccinations := []Vaccination{*vax1, *vax2}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *VaccinationFilter) ([]Vaccination, error) {
			return vaccinations, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 vaccinations, got %d", len(result))
	}
	if result[0].ID != "vax-123" {
		t.Errorf("Expected first vaccination ID vax-123, got %s", result[0].ID)
	}
	if result[1].ID != "vax-456" {
		t.Errorf("Expected second vaccination ID vax-456, got %s", result[1].ID)
	}
}

func TestRecordAdministration_WithAllOptionalFields(t *testing.T) {
	var capturedReq *RecordVaccinationRequest
	svc := &mockService{
		recordAdministrationFn: func(ctx context.Context, id string, req *RecordVaccinationRequest) (*Vaccination, error) {
			capturedReq = req
			return completedVaccination(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := &RecordVaccinationRequest{
		AdministeredAt: time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC),
		Provider:       "Dr. Smith",
		Location:       "Pediatric Clinic",
		LotNumber:      "LOT123ABC",
		Notes:          "No adverse reactions observed. Child was calm.",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/vaccinations/vax-123/record", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if capturedReq.Provider != "Dr. Smith" {
		t.Errorf("Expected Provider 'Dr. Smith', got %s", capturedReq.Provider)
	}
	if capturedReq.Location != "Pediatric Clinic" {
		t.Errorf("Expected Location 'Pediatric Clinic', got %s", capturedReq.Location)
	}
	if capturedReq.LotNumber != "LOT123ABC" {
		t.Errorf("Expected LotNumber 'LOT123ABC', got %s", capturedReq.LotNumber)
	}
	if capturedReq.Notes != "No adverse reactions observed. Child was calm." {
		t.Errorf("Expected Notes to be set, got %s", capturedReq.Notes)
	}
}

func TestGetUpcoming_MultipleResults(t *testing.T) {
	vax1 := sampleVaccination()
	vax2 := sampleVaccination()
	vax2.ID = "vax-456"
	vax2.Name = "Polio"
	vax2.ScheduledAt = time.Date(2025, 3, 20, 10, 0, 0, 0, time.UTC)

	vaccinations := []Vaccination{*vax1, *vax2}
	svc := &mockService{
		getUpcomingFn: func(ctx context.Context, childID string, days int) ([]Vaccination, error) {
			return vaccinations, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/upcoming/child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Vaccination
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 vaccinations, got %d", len(result))
	}
}

func TestGetSchedule_FullScheduleDetails(t *testing.T) {
	schedule := []VaccinationSchedule{
		{
			ID:          "sched-1",
			Name:        "DTaP",
			Description: "Diphtheria, Tetanus, Pertussis",
			AgeWeeks:    8,
			AgeMonths:   2,
			AgeLabel:    "2 months",
			Dose:        1,
		},
	}
	svc := &mockService{
		getScheduleFn: func() []VaccinationSchedule {
			return schedule
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/vaccinations/schedule", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []VaccinationSchedule
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 schedule item, got %d", len(result))
	}

	item := result[0]
	if item.ID != "sched-1" {
		t.Errorf("Expected ID 'sched-1', got %s", item.ID)
	}
	if item.Name != "DTaP" {
		t.Errorf("Expected Name 'DTaP', got %s", item.Name)
	}
	if item.Description != "Diphtheria, Tetanus, Pertussis" {
		t.Errorf("Expected Description, got %s", item.Description)
	}
	if item.AgeWeeks != 8 {
		t.Errorf("Expected AgeWeeks 8, got %d", item.AgeWeeks)
	}
	if item.AgeMonths != 2 {
		t.Errorf("Expected AgeMonths 2, got %d", item.AgeMonths)
	}
	if item.AgeLabel != "2 months" {
		t.Errorf("Expected AgeLabel '2 months', got %s", item.AgeLabel)
	}
	if item.Dose != 1 {
		t.Errorf("Expected Dose 1, got %d", item.Dose)
	}
}

func TestCreate_WithDifferentDoses(t *testing.T) {
	testCases := []struct {
		name string
		dose int
	}{
		{"first dose", 1},
		{"second dose", 2},
		{"third dose", 3},
		{"booster", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedReq *CreateVaccinationRequest
			svc := &mockService{
				createFn: func(ctx context.Context, req *CreateVaccinationRequest) (*Vaccination, error) {
					capturedReq = req
					return sampleVaccination(), nil
				},
			}
			router := setupRouter(svc)

			reqBody := &CreateVaccinationRequest{
				ChildID:     "child-456",
				Name:        "DTaP",
				Dose:        tc.dose,
				ScheduledAt: time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC),
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/vaccinations", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("Expected status 201, got %d", w.Code)
			}
			if capturedReq.Dose != tc.dose {
				t.Errorf("Expected Dose %d, got %d", tc.dose, capturedReq.Dose)
			}
		})
	}
}
