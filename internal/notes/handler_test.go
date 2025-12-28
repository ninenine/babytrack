package notes

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
	createFn func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error)
	getFn    func(ctx context.Context, id string) (*Note, error)
	listFn   func(ctx context.Context, filter *NoteFilter) ([]Note, error)
	updateFn func(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error)
	deleteFn func(ctx context.Context, id string) error
	pinFn    func(ctx context.Context, id string, pinned bool) error
	searchFn func(ctx context.Context, childID, query string) ([]Note, error)
}

func (m *mockService) Create(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockService) Get(ctx context.Context, id string) (*Note, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) List(ctx context.Context, filter *NoteFilter) ([]Note, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockService) Update(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
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

func (m *mockService) Pin(ctx context.Context, id string, pinned bool) error {
	if m.pinFn != nil {
		return m.pinFn(ctx, id, pinned)
	}
	return nil
}

func (m *mockService) Search(ctx context.Context, childID, query string) ([]Note, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, childID, query)
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

	group := router.Group("/notes")
	handler.RegisterRoutes(group)
	return router
}

// Helper to create a sample note
func sampleNote() *Note {
	now := time.Now()
	return &Note{
		ID:        "note-123",
		ChildID:   "child-456",
		AuthorID:  "test-user-123",
		Title:     "Sample Note",
		Content:   "This is a sample note content",
		Tags:      []string{"health", "appointment"},
		Pinned:    false,
		CreatedAt: now,
		UpdatedAt: now,
		SyncedAt:  &now,
	}
}

// Helper to create a valid note request body
func validCreateNoteRequest() *CreateNoteRequest {
	return &CreateNoteRequest{
		ChildID: "child-456",
		Title:   "New Note",
		Content: "Note content here",
		Tags:    []string{"tag1", "tag2"},
		Pinned:  false,
	}
}

// Helper to create a valid update note request body
func validUpdateNoteRequest() *UpdateNoteRequest {
	return &UpdateNoteRequest{
		Title:   "Updated Note",
		Content: "Updated content",
		Tags:    []string{"updated-tag"},
		Pinned:  true,
	}
}

// =====================
// List Handler Tests
// =====================

func TestList_Success(t *testing.T) {
	notes := []Note{*sampleNote()}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			return notes, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 note, got %d", len(result))
	}
}

func TestList_WithChildIDFilter(t *testing.T) {
	var capturedFilter *NoteFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			capturedFilter = filter
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes?child_id=child-456", http.NoBody)
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

func TestList_WithPinnedOnlyFilter(t *testing.T) {
	var capturedFilter *NoteFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			capturedFilter = filter
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes?pinned_only=true", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedFilter == nil {
		t.Fatal("Filter was not passed to service")
	}
	if !capturedFilter.PinnedOnly {
		t.Error("Expected PinnedOnly to be true")
	}
}

func TestList_WithAllFilters(t *testing.T) {
	var capturedFilter *NoteFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			capturedFilter = filter
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes?child_id=child-456&pinned_only=true", http.NoBody)
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
	if !capturedFilter.PinnedOnly {
		t.Error("Expected PinnedOnly to be true")
	}
}

func TestList_PinnedOnlyFalseWhenNotTrue(t *testing.T) {
	var capturedFilter *NoteFilter
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			capturedFilter = filter
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes?pinned_only=false", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedFilter.PinnedOnly {
		t.Error("Expected PinnedOnly to be false when not 'true'")
	}
}

func TestList_ServiceError(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes", http.NoBody)
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
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty list, got %d", len(result))
	}
}

func TestList_MultipleResults(t *testing.T) {
	note1 := sampleNote()
	note2 := sampleNote()
	note2.ID = "note-456"
	note2.Title = "Second Note"

	notes := []Note{*note1, *note2}
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			return notes, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(result))
	}
	if result[0].ID != "note-123" {
		t.Errorf("Expected first note ID note-123, got %s", result[0].ID)
	}
	if result[1].ID != "note-456" {
		t.Errorf("Expected second note ID note-456, got %s", result[1].ID)
	}
}

// =====================
// Get Handler Tests
// =====================

func TestGet_Success(t *testing.T) {
	note := sampleNote()
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Note, error) {
			if id == "note-123" {
				return note, nil
			}
			return nil, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/note-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "note-123" {
		t.Errorf("Expected ID note-123, got %s", result.ID)
	}
}

func TestGet_ServiceError(t *testing.T) {
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Note, error) {
			return nil, errors.New("note not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "note not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestGet_VerifiesIDParam(t *testing.T) {
	var capturedID string
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Note, error) {
			capturedID = id
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/my-specific-id", http.NoBody)
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
	note := sampleNote()
	svc := &mockService{
		createFn: func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
			return note, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validCreateNoteRequest())
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ID != "note-123" {
		t.Errorf("Expected ID note-123, got %s", result.ID)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/notes", bytes.NewReader([]byte("invalid json")))
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

	// Missing required fields (child_id, content)
	body, _ := json.Marshal(map[string]any{
		"title": "Just a title",
	})
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
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
		"content": "Note content without child_id",
	})
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing child_id, got %d", w.Code)
	}
}

func TestCreate_MissingContent(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]any{
		"child_id": "child-456",
		"title":    "Title without content",
	})
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing content, got %d", w.Code)
	}
}

func TestCreate_ServiceError(t *testing.T) {
	svc := &mockService{
		createFn: func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
			return nil, errors.New("failed to create note")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validCreateNoteRequest())
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
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
	if result["error"] != "failed to create note" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestCreate_VerifiesUserIDAndRequestData(t *testing.T) {
	var capturedUserID string
	var capturedReq *CreateNoteRequest
	svc := &mockService{
		createFn: func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
			capturedUserID = userID
			capturedReq = req
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validCreateNoteRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "test-user-123" {
		t.Errorf("Expected userID test-user-123, got %s", capturedUserID)
	}
	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.ChildID != reqBody.ChildID {
		t.Errorf("Expected ChildID %s, got %s", reqBody.ChildID, capturedReq.ChildID)
	}
	if capturedReq.Title != reqBody.Title {
		t.Errorf("Expected Title %s, got %s", reqBody.Title, capturedReq.Title)
	}
	if capturedReq.Content != reqBody.Content {
		t.Errorf("Expected Content %s, got %s", reqBody.Content, capturedReq.Content)
	}
}

func TestCreate_WithOptionalFields(t *testing.T) {
	var capturedReq *CreateNoteRequest
	svc := &mockService{
		createFn: func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
			capturedReq = req
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := &CreateNoteRequest{
		ChildID: "child-456",
		Title:   "Note with all fields",
		Content: "Full content here",
		Tags:    []string{"important", "health", "appointment"},
		Pinned:  true,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if capturedReq.Title != "Note with all fields" {
		t.Errorf("Expected Title to be set, got %s", capturedReq.Title)
	}
	if len(capturedReq.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(capturedReq.Tags))
	}
	if !capturedReq.Pinned {
		t.Error("Expected Pinned to be true")
	}
}

func TestCreate_WithoutOptionalTitle(t *testing.T) {
	var capturedReq *CreateNoteRequest
	svc := &mockService{
		createFn: func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
			capturedReq = req
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := map[string]any{
		"child_id": "child-456",
		"content":  "Content without title",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if capturedReq.Title != "" {
		t.Errorf("Expected Title to be empty, got %s", capturedReq.Title)
	}
}

// =====================
// Update Handler Tests
// =====================

func TestUpdate_Success(t *testing.T) {
	note := sampleNote()
	note.Title = "Updated Title"
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
			return note, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validUpdateNoteRequest())
	req := httptest.NewRequest("PUT", "/notes/note-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Title != "Updated Title" {
		t.Errorf("Expected Title 'Updated Title', got %s", result.Title)
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("PUT", "/notes/note-123", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdate_ServiceError(t *testing.T) {
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
			return nil, errors.New("note not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(validUpdateNoteRequest())
	req := httptest.NewRequest("PUT", "/notes/nonexistent", bytes.NewReader(body))
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
	if result["error"] != "note not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestUpdate_VerifiesIDAndRequest(t *testing.T) {
	var capturedID string
	var capturedReq *UpdateNoteRequest
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
			capturedID = id
			capturedReq = req
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := validUpdateNoteRequest()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/notes/update-id-456", bytes.NewReader(body))
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
	if capturedReq.Content != reqBody.Content {
		t.Errorf("Expected Content %s, got %s", reqBody.Content, capturedReq.Content)
	}
}

func TestUpdate_WithPartialFields(t *testing.T) {
	var capturedReq *UpdateNoteRequest
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
			capturedReq = req
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	// Only updating content, not title
	reqBody := map[string]any{
		"content": "Only updating content",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/notes/note-123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if capturedReq.Content != "Only updating content" {
		t.Errorf("Expected Content to be set, got %s", capturedReq.Content)
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

	req := httptest.NewRequest("DELETE", "/notes/note-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDelete_ServiceError(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("note not found")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("DELETE", "/notes/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "note not found" {
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

	req := httptest.NewRequest("DELETE", "/notes/delete-me-789", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "delete-me-789" {
		t.Errorf("Expected ID delete-me-789, got %s", capturedID)
	}
}

// =====================
// Search Handler Tests
// =====================

func TestSearch_Success(t *testing.T) {
	notes := []Note{*sampleNote()}
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			return notes, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?child_id=child-456&q=sample", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 note, got %d", len(result))
	}
}

func TestSearch_ServiceError(t *testing.T) {
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			return nil, errors.New("search failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?child_id=child-456&q=test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "search failed" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestSearch_VerifiesQueryParams(t *testing.T) {
	var capturedChildID, capturedQuery string
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			capturedChildID = childID
			capturedQuery = query
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?child_id=child-789&q=test+query", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedChildID != "child-789" {
		t.Errorf("Expected childID child-789, got %s", capturedChildID)
	}
	if capturedQuery != "test query" {
		t.Errorf("Expected query 'test query', got %s", capturedQuery)
	}
}

func TestSearch_EmptyResult(t *testing.T) {
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?child_id=child-456&q=nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty list, got %d", len(result))
	}
}

func TestSearch_MultipleResults(t *testing.T) {
	note1 := sampleNote()
	note2 := sampleNote()
	note2.ID = "note-456"
	note2.Title = "Another matching note"

	notes := []Note{*note1, *note2}
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			return notes, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?child_id=child-456&q=note", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result []Note
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(result))
	}
}

func TestSearch_WithoutChildID(t *testing.T) {
	var capturedChildID string
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			capturedChildID = childID
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?q=test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedChildID != "" {
		t.Errorf("Expected empty childID, got %s", capturedChildID)
	}
}

func TestSearch_WithoutQuery(t *testing.T) {
	var capturedQuery string
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			capturedQuery = query
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?child_id=child-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedQuery != "" {
		t.Errorf("Expected empty query, got %s", capturedQuery)
	}
}

// =====================
// Pin Handler Tests
// =====================

func TestPin_Success(t *testing.T) {
	svc := &mockService{
		pinFn: func(ctx context.Context, id string, pinned bool) error {
			return nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]bool{"pinned": true})
	req := httptest.NewRequest("POST", "/notes/note-123/pin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPin_Unpin(t *testing.T) {
	var capturedPinned bool
	svc := &mockService{
		pinFn: func(ctx context.Context, id string, pinned bool) error {
			capturedPinned = pinned
			return nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]bool{"pinned": false})
	req := httptest.NewRequest("POST", "/notes/note-123/pin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedPinned {
		t.Error("Expected pinned to be false for unpin operation")
	}
}

func TestPin_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/notes/note-123/pin", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestPin_ServiceError(t *testing.T) {
	svc := &mockService{
		pinFn: func(ctx context.Context, id string, pinned bool) error {
			return errors.New("note not found")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]bool{"pinned": true})
	req := httptest.NewRequest("POST", "/notes/nonexistent/pin", bytes.NewReader(body))
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
	if result["error"] != "note not found" {
		t.Errorf("Expected error message, got %s", result["error"])
	}
}

func TestPin_VerifiesIDAndPinned(t *testing.T) {
	var capturedID string
	var capturedPinned bool
	svc := &mockService{
		pinFn: func(ctx context.Context, id string, pinned bool) error {
			capturedID = id
			capturedPinned = pinned
			return nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(map[string]bool{"pinned": true})
	req := httptest.NewRequest("POST", "/notes/pin-this-note/pin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "pin-this-note" {
		t.Errorf("Expected ID pin-this-note, got %s", capturedID)
	}
	if !capturedPinned {
		t.Error("Expected pinned to be true")
	}
}

// =====================
// Route Registration Tests
// =====================

func TestRegisterRoutes(t *testing.T) {
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			return []Note{}, nil
		},
		getFn: func(ctx context.Context, id string) (*Note, error) {
			return sampleNote(), nil
		},
		createFn: func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
			return sampleNote(), nil
		},
		updateFn: func(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
			return sampleNote(), nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
		pinFn: func(ctx context.Context, id string, pinned bool) error {
			return nil
		},
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	testCases := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"GET", "/notes", "", http.StatusOK},
		{"GET", "/notes/note-123", "", http.StatusOK},
		{"POST", "/notes", `{"child_id":"c1","content":"Test content"}`, http.StatusCreated},
		{"PUT", "/notes/note-123", `{"content":"Updated content"}`, http.StatusOK},
		{"DELETE", "/notes/note-123", "", http.StatusNoContent},
		{"GET", "/notes/search?q=test", "", http.StatusOK},
		{"POST", "/notes/note-123/pin", `{"pinned":true}`, http.StatusOK},
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

func TestCreate_EmptyTags(t *testing.T) {
	var capturedReq *CreateNoteRequest
	svc := &mockService{
		createFn: func(ctx context.Context, userID string, req *CreateNoteRequest) (*Note, error) {
			capturedReq = req
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	reqBody := map[string]any{
		"child_id": "child-456",
		"content":  "Note without tags",
		"tags":     []string{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if len(capturedReq.Tags) != 0 {
		t.Errorf("Expected empty tags, got %v", capturedReq.Tags)
	}
}

func TestUpdate_EmptyBody(t *testing.T) {
	svc := &mockService{
		updateFn: func(ctx context.Context, id string, req *UpdateNoteRequest) (*Note, error) {
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	// Empty JSON object should be valid
	req := httptest.NewRequest("PUT", "/notes/note-123", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSearch_SpecialCharactersInQuery(t *testing.T) {
	var capturedQuery string
	svc := &mockService{
		searchFn: func(ctx context.Context, childID, query string) ([]Note, error) {
			capturedQuery = query
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes/search?q=hello%20world%21", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedQuery != "hello world!" {
		t.Errorf("Expected query 'hello world!', got %s", capturedQuery)
	}
}

func TestList_NilFilterHandling(t *testing.T) {
	var filterReceived bool
	svc := &mockService{
		listFn: func(ctx context.Context, filter *NoteFilter) ([]Note, error) {
			filterReceived = filter != nil
			return []Note{}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/notes", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !filterReceived {
		t.Error("Expected filter to be passed (not nil)")
	}
}

func TestGet_EmptyID(t *testing.T) {
	svc := &mockService{
		getFn: func(ctx context.Context, id string) (*Note, error) {
			return sampleNote(), nil
		},
	}
	router := setupRouter(svc)

	// This should match the list route, not get with empty ID
	req := httptest.NewRequest("GET", "/notes/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin may redirect or handle this differently
	// The test verifies the route behaviour
	if w.Code == http.StatusNotFound {
		t.Log("Empty ID path handled as 404, which is acceptable")
	}
}

func TestPin_MissingBody(t *testing.T) {
	svc := &mockService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/notes/note-123/pin", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing body, got %d", w.Code)
	}
}

func TestPin_DefaultPinnedValue(t *testing.T) {
	var capturedPinned bool
	svc := &mockService{
		pinFn: func(ctx context.Context, id string, pinned bool) error {
			capturedPinned = pinned
			return nil
		},
	}
	router := setupRouter(svc)

	// Empty JSON object - pinned should default to false
	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest("POST", "/notes/note-123/pin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedPinned {
		t.Error("Expected pinned to default to false")
	}
}
