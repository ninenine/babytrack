package sync

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

// mockSyncService implements the Service interface for testing
type mockSyncService struct {
	pushFn   func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error)
	pullFn   func(ctx context.Context, userID string, lastSync string) (*PullResponse, error)
	statusFn func(ctx context.Context, userID string) (*SyncStatus, error)
}

func (m *mockSyncService) Push(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
	if m.pushFn != nil {
		return m.pushFn(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockSyncService) Pull(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
	if m.pullFn != nil {
		return m.pullFn(ctx, userID, lastSync)
	}
	return nil, nil
}

func (m *mockSyncService) Status(ctx context.Context, userID string) (*SyncStatus, error) {
	if m.statusFn != nil {
		return m.statusFn(ctx, userID)
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

	group := router.Group("/sync")
	handler.RegisterRoutes(group)
	return router
}

// setupRouterWithUserID creates a test router with a custom user_id
func setupRouterWithUserID(svc Service, userID string) *gin.Engine {
	router := gin.New()
	handler := NewHandler(svc)

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	group := router.Group("/sync")
	handler.RegisterRoutes(group)
	return router
}

// Helper to create a sample push request
func samplePushRequest() *PushRequest {
	return &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "create",
				EntityID:  "",
				Timestamp: time.Now().UTC(),
				ClientID:  "client-123",
				Data: map[string]any{
					"child_id":   "child-123",
					"type":       "bottle",
					"start_time": time.Now().UTC().Format(time.RFC3339),
				},
			},
		},
	}
}

// Helper to create a sample push response
func samplePushResponse() *PushResponse {
	return &PushResponse{
		Processed:  1,
		Failed:     0,
		FailedIDs:  nil,
		Results:    map[string]string{"event-1": "feeding-new-id"},
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}
}

// Helper to create a sample pull response
func samplePullResponse() *PullResponse {
	return &PullResponse{
		Events: []Event{
			{
				ID:        "event-1",
				Type:      EventTypeFeeding,
				Action:    "create",
				EntityID:  "feeding-123",
				Timestamp: time.Now().UTC(),
				ClientID:  "server",
			},
		},
		ServerTime: time.Now().UTC().Format(time.RFC3339),
		HasMore:    false,
	}
}

// Helper to create a sample sync status
func sampleSyncStatus() *SyncStatus {
	return &SyncStatus{
		LastSync:   time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339),
		Pending:    5,
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}
}

// =====================
// Push Handler Tests
// =====================

func TestPush_Success(t *testing.T) {
	pushResp := samplePushResponse()
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			return pushResp, nil
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(samplePushRequest())
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PushResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Processed != 1 {
		t.Errorf("Expected Processed 1, got %d", result.Processed)
	}
	if result.Failed != 0 {
		t.Errorf("Expected Failed 0, got %d", result.Failed)
	}
	if result.Results["event-1"] != "feeding-new-id" {
		t.Errorf("Expected result event-1 to be feeding-new-id, got %s", result.Results["event-1"])
	}
}

func TestPush_ServiceError(t *testing.T) {
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	body, _ := json.Marshal(samplePushRequest())
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
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
	if result["error"] != "database connection failed" {
		t.Errorf("Expected error message 'database connection failed', got %s", result["error"])
	}
}

func TestPush_InvalidJSON(t *testing.T) {
	svc := &mockSyncService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader([]byte("invalid json")))
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

func TestPush_EmptyBody(t *testing.T) {
	svc := &mockSyncService{}
	router := setupRouter(svc)

	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestPush_MalformedJSON(t *testing.T) {
	svc := &mockSyncService{}
	router := setupRouter(svc)

	// JSON with trailing comma
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader([]byte(`{"client_id":"c1",}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestPush_VerifiesUserIDFromContext(t *testing.T) {
	var capturedUserID string
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			capturedUserID = userID
			return samplePushResponse(), nil
		},
	}
	router := setupRouterWithUserID(svc, "specific-user-456")

	body, _ := json.Marshal(samplePushRequest())
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "specific-user-456" {
		t.Errorf("Expected userID specific-user-456, got %s", capturedUserID)
	}
}

func TestPush_VerifiesRequestData(t *testing.T) {
	var capturedReq *PushRequest
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			capturedReq = req
			return samplePushResponse(), nil
		},
	}
	router := setupRouter(svc)

	pushReq := samplePushRequest()
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedReq == nil {
		t.Fatal("Request was not passed to service")
	}
	if capturedReq.ClientID != pushReq.ClientID {
		t.Errorf("Expected ClientID %s, got %s", pushReq.ClientID, capturedReq.ClientID)
	}
	if len(capturedReq.Events) != len(pushReq.Events) {
		t.Errorf("Expected %d events, got %d", len(pushReq.Events), len(capturedReq.Events))
	}
}

func TestPush_MultipleEvents(t *testing.T) {
	var capturedReq *PushRequest
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			capturedReq = req
			return &PushResponse{
				Processed:  3,
				Failed:     0,
				Results:    map[string]string{"event-1": "id-1", "event-2": "id-2", "event-3": "id-3"},
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	pushReq := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{ID: "event-1", Type: EventTypeFeeding, Action: "create", Timestamp: time.Now()},
			{ID: "event-2", Type: EventTypeSleep, Action: "create", Timestamp: time.Now()},
			{ID: "event-3", Type: EventTypeNote, Action: "create", Timestamp: time.Now()},
		},
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if len(capturedReq.Events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(capturedReq.Events))
	}

	var result PushResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Processed != 3 {
		t.Errorf("Expected Processed 3, got %d", result.Processed)
	}
}

func TestPush_EmptyEvents(t *testing.T) {
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			return &PushResponse{
				Processed:  0,
				Failed:     0,
				Results:    map[string]string{},
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	pushReq := &PushRequest{
		ClientID: "client-123",
		Events:   []Event{},
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PushResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Processed != 0 {
		t.Errorf("Expected Processed 0, got %d", result.Processed)
	}
}

func TestPush_WithPartialFailures(t *testing.T) {
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			return &PushResponse{
				Processed:  2,
				Failed:     1,
				FailedIDs:  []string{"event-2"},
				Results:    map[string]string{"event-1": "id-1", "event-3": "id-3"},
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	pushReq := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{ID: "event-1", Type: EventTypeFeeding, Action: "create", Timestamp: time.Now()},
			{ID: "event-2", Type: EventTypeSleep, Action: "create", Timestamp: time.Now()},
			{ID: "event-3", Type: EventTypeNote, Action: "create", Timestamp: time.Now()},
		},
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PushResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Processed != 2 {
		t.Errorf("Expected Processed 2, got %d", result.Processed)
	}
	if result.Failed != 1 {
		t.Errorf("Expected Failed 1, got %d", result.Failed)
	}
	if len(result.FailedIDs) != 1 || result.FailedIDs[0] != "event-2" {
		t.Errorf("Expected FailedIDs [event-2], got %v", result.FailedIDs)
	}
}

// =====================
// Pull Handler Tests
// =====================

func TestPull_Success(t *testing.T) {
	pullResp := samplePullResponse()
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			return pullResp, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull?last_sync=2025-01-01T00:00:00Z", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PullResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(result.Events))
	}
	if result.HasMore {
		t.Error("Expected HasMore false")
	}
}

func TestPull_ServiceError(t *testing.T) {
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			return nil, errors.New("database connection failed")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull?last_sync=2025-01-01T00:00:00Z", http.NoBody)
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
		t.Errorf("Expected error message 'database connection failed', got %s", result["error"])
	}
}

func TestPull_VerifiesUserIDFromContext(t *testing.T) {
	var capturedUserID string
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			capturedUserID = userID
			return samplePullResponse(), nil
		},
	}
	router := setupRouterWithUserID(svc, "specific-user-789")

	req := httptest.NewRequest("GET", "/sync/pull", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "specific-user-789" {
		t.Errorf("Expected userID specific-user-789, got %s", capturedUserID)
	}
}

func TestPull_VerifiesLastSyncQueryParam(t *testing.T) {
	var capturedLastSync string
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			capturedLastSync = lastSync
			return samplePullResponse(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull?last_sync=2025-06-15T10:30:00Z", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedLastSync != "2025-06-15T10:30:00Z" {
		t.Errorf("Expected lastSync 2025-06-15T10:30:00Z, got %s", capturedLastSync)
	}
}

func TestPull_EmptyLastSync(t *testing.T) {
	var capturedLastSync string
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			capturedLastSync = lastSync
			return samplePullResponse(), nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if capturedLastSync != "" {
		t.Errorf("Expected empty lastSync, got %s", capturedLastSync)
	}
}

func TestPull_EmptyEvents(t *testing.T) {
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			return &PullResponse{
				Events:     []Event{},
				ServerTime: time.Now().UTC().Format(time.RFC3339),
				HasMore:    false,
			}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull?last_sync=2025-01-01T00:00:00Z", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PullResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(result.Events))
	}
}

func TestPull_HasMoreTrue(t *testing.T) {
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			return &PullResponse{
				Events: []Event{
					{ID: "event-1", Type: EventTypeFeeding, Action: "create"},
				},
				ServerTime: time.Now().UTC().Format(time.RFC3339),
				HasMore:    true,
			}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull?last_sync=2025-01-01T00:00:00Z", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PullResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if !result.HasMore {
		t.Error("Expected HasMore true")
	}
}

func TestPull_MultipleEvents(t *testing.T) {
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			return &PullResponse{
				Events: []Event{
					{ID: "event-1", Type: EventTypeFeeding, Action: "create"},
					{ID: "event-2", Type: EventTypeSleep, Action: "update"},
					{ID: "event-3", Type: EventTypeNote, Action: "delete"},
				},
				ServerTime: time.Now().UTC().Format(time.RFC3339),
				HasMore:    false,
			}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PullResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(result.Events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(result.Events))
	}
}

// =====================
// Status Handler Tests
// =====================

func TestStatus_Success(t *testing.T) {
	statusResp := sampleSyncStatus()
	svc := &mockSyncService{
		statusFn: func(ctx context.Context, userID string) (*SyncStatus, error) {
			return statusResp, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/status", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result SyncStatus
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Pending != 5 {
		t.Errorf("Expected Pending 5, got %d", result.Pending)
	}
	if result.ServerTime == "" {
		t.Error("Expected ServerTime to be set")
	}
}

func TestStatus_ServiceError(t *testing.T) {
	svc := &mockSyncService{
		statusFn: func(ctx context.Context, userID string) (*SyncStatus, error) {
			return nil, errors.New("failed to get sync status")
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/status", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["error"] != "failed to get sync status" {
		t.Errorf("Expected error message 'failed to get sync status', got %s", result["error"])
	}
}

func TestStatus_VerifiesUserIDFromContext(t *testing.T) {
	var capturedUserID string
	svc := &mockSyncService{
		statusFn: func(ctx context.Context, userID string) (*SyncStatus, error) {
			capturedUserID = userID
			return sampleSyncStatus(), nil
		},
	}
	router := setupRouterWithUserID(svc, "status-user-111")

	req := httptest.NewRequest("GET", "/sync/status", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "status-user-111" {
		t.Errorf("Expected userID status-user-111, got %s", capturedUserID)
	}
}

func TestStatus_EmptyLastSync(t *testing.T) {
	svc := &mockSyncService{
		statusFn: func(ctx context.Context, userID string) (*SyncStatus, error) {
			return &SyncStatus{
				LastSync:   "",
				Pending:    0,
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/status", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result SyncStatus
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.LastSync != "" {
		t.Errorf("Expected empty LastSync, got %s", result.LastSync)
	}
}

func TestStatus_ZeroPending(t *testing.T) {
	svc := &mockSyncService{
		statusFn: func(ctx context.Context, userID string) (*SyncStatus, error) {
			return &SyncStatus{
				LastSync:   time.Now().UTC().Format(time.RFC3339),
				Pending:    0,
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/status", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result SyncStatus
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Pending != 0 {
		t.Errorf("Expected Pending 0, got %d", result.Pending)
	}
}

// =====================
// Route Registration Tests
// =====================

func TestRegisterRoutes(t *testing.T) {
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			return samplePushResponse(), nil
		},
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			return samplePullResponse(), nil
		},
		statusFn: func(ctx context.Context, userID string) (*SyncStatus, error) {
			return sampleSyncStatus(), nil
		},
	}
	router := setupRouter(svc)

	testCases := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{"POST", "/sync/push", `{"client_id":"c1","events":[]}`, http.StatusOK},
		{"GET", "/sync/pull", "", http.StatusOK},
		{"GET", "/sync/pull?last_sync=2025-01-01T00:00:00Z", "", http.StatusOK},
		{"GET", "/sync/status", "", http.StatusOK},
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

func TestRegisterRoutes_MethodNotAllowed(t *testing.T) {
	svc := &mockSyncService{}
	router := setupRouter(svc)

	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/sync/push"},
		{"POST", "/sync/pull"},
		{"POST", "/sync/status"},
		{"PUT", "/sync/push"},
		{"DELETE", "/sync/pull"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Gin returns 404 for method not allowed by default (unless HandleMethodNotAllowed is set)
			if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s: expected status 404 or 405, got %d", tc.method, tc.path, w.Code)
			}
		})
	}
}

// =====================
// NewHandler Tests
// =====================

func TestNewHandler(t *testing.T) {
	svc := &mockSyncService{}
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

func TestPush_WithAllEventTypes(t *testing.T) {
	var capturedReq *PushRequest
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			capturedReq = req
			return &PushResponse{
				Processed:  7,
				Failed:     0,
				Results:    make(map[string]string),
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	pushReq := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{ID: "e1", Type: EventTypeFeeding, Action: "create", Timestamp: time.Now()},
			{ID: "e2", Type: EventTypeSleep, Action: "create", Timestamp: time.Now()},
			{ID: "e3", Type: EventTypeMedication, Action: "create", Timestamp: time.Now()},
			{ID: "e4", Type: EventTypeMedicationLog, Action: "create", Timestamp: time.Now()},
			{ID: "e5", Type: EventTypeNote, Action: "create", Timestamp: time.Now()},
			{ID: "e6", Type: EventTypeVaccination, Action: "create", Timestamp: time.Now()},
			{ID: "e7", Type: EventTypeAppointment, Action: "create", Timestamp: time.Now()},
		},
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if len(capturedReq.Events) != 7 {
		t.Errorf("Expected 7 events, got %d", len(capturedReq.Events))
	}

	// Verify all event types were captured
	eventTypes := make(map[EventType]bool)
	for _, e := range capturedReq.Events {
		eventTypes[e.Type] = true
	}
	expectedTypes := []EventType{
		EventTypeFeeding, EventTypeSleep, EventTypeMedication,
		EventTypeMedicationLog, EventTypeNote, EventTypeVaccination, EventTypeAppointment,
	}
	for _, et := range expectedTypes {
		if !eventTypes[et] {
			t.Errorf("Expected event type %s to be present", et)
		}
	}
}

func TestPush_WithAllActions(t *testing.T) {
	var capturedReq *PushRequest
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			capturedReq = req
			return &PushResponse{
				Processed:  4,
				Failed:     0,
				Results:    make(map[string]string),
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	pushReq := &PushRequest{
		ClientID: "client-123",
		Events: []Event{
			{ID: "e1", Type: EventTypeFeeding, Action: "create", Timestamp: time.Now()},
			{ID: "e2", Type: EventTypeFeeding, Action: "update", EntityID: "feeding-1", Timestamp: time.Now()},
			{ID: "e3", Type: EventTypeFeeding, Action: "delete", EntityID: "feeding-2", Timestamp: time.Now()},
			{ID: "e4", Type: EventTypeMedication, Action: "deactivate", EntityID: "med-1", Timestamp: time.Now()},
		},
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify all actions were captured
	actions := make(map[string]bool)
	for _, e := range capturedReq.Events {
		actions[e.Action] = true
	}
	if !actions["create"] || !actions["update"] || !actions["delete"] || !actions["deactivate"] {
		t.Error("Expected all action types to be present")
	}
}

func TestPush_LargePayload(t *testing.T) {
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			return &PushResponse{
				Processed:  len(req.Events),
				Failed:     0,
				Results:    make(map[string]string),
				ServerTime: time.Now().UTC().Format(time.RFC3339),
			}, nil
		},
	}
	router := setupRouter(svc)

	// Create a large number of events
	events := make([]Event, 100)
	for i := range 100 {
		events[i] = Event{
			ID:        "event-" + string(rune('0'+i%10)) + string(rune('0'+i/10)),
			Type:      EventTypeFeeding,
			Action:    "create",
			Timestamp: time.Now(),
			Data: map[string]any{
				"child_id":   "child-123",
				"type":       "bottle",
				"start_time": time.Now().Format(time.RFC3339),
			},
		}
	}

	pushReq := &PushRequest{
		ClientID: "client-123",
		Events:   events,
	}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result PushResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Processed != 100 {
		t.Errorf("Expected Processed 100, got %d", result.Processed)
	}
}

func TestPull_ResponseContainsServerTime(t *testing.T) {
	serverTime := "2025-12-28T15:30:00Z"
	svc := &mockSyncService{
		pullFn: func(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
			return &PullResponse{
				Events:     []Event{},
				ServerTime: serverTime,
				HasMore:    false,
			}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/pull", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var result PullResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ServerTime != serverTime {
		t.Errorf("Expected ServerTime %s, got %s", serverTime, result.ServerTime)
	}
}

func TestStatus_ResponseContainsServerTime(t *testing.T) {
	serverTime := "2025-12-28T15:30:00Z"
	svc := &mockSyncService{
		statusFn: func(ctx context.Context, userID string) (*SyncStatus, error) {
			return &SyncStatus{
				LastSync:   "",
				Pending:    0,
				ServerTime: serverTime,
			}, nil
		},
	}
	router := setupRouter(svc)

	req := httptest.NewRequest("GET", "/sync/status", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var result SyncStatus
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ServerTime != serverTime {
		t.Errorf("Expected ServerTime %s, got %s", serverTime, result.ServerTime)
	}
}

func TestPush_ResponseContainsServerTime(t *testing.T) {
	serverTime := "2025-12-28T15:30:00Z"
	svc := &mockSyncService{
		pushFn: func(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
			return &PushResponse{
				Processed:  0,
				Failed:     0,
				Results:    make(map[string]string),
				ServerTime: serverTime,
			}, nil
		},
	}
	router := setupRouter(svc)

	pushReq := &PushRequest{ClientID: "c1", Events: []Event{}}
	body, _ := json.Marshal(pushReq)
	req := httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var result PushResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ServerTime != serverTime {
		t.Errorf("Expected ServerTime %s, got %s", serverTime, result.ServerTime)
	}
}
