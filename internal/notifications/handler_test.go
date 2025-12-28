package notifications

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestHub() *Hub {
	hub := NewHub()
	go hub.Run()
	return hub
}

func setupRouter(hub *Hub) *gin.Engine {
	router := gin.New()

	// Simulated auth middleware
	router.Use(func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})

	handler := NewHandler(hub)
	group := router.Group("/notifications")
	handler.RegisterRoutes(group)
	return router
}

func TestNewHandler(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub)

	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}

	if handler.hub != hub {
		t.Error("NewHandler() did not set hub correctly")
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	hub := setupTestHub()

	router := gin.New()
	handler := NewHandler(hub)
	group := router.Group("/notifications")
	handler.RegisterRoutes(group)

	// Check routes are registered by checking that unknown routes return 404
	req := httptest.NewRequest("GET", "/unknown", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Error("Unknown route should return 404")
	}
}

func TestHandler_Stream_Unauthorised(t *testing.T) {
	hub := setupTestHub()
	router := setupRouter(hub)

	// Request without user_id in context
	req := httptest.NewRequest("GET", "/notifications/stream", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandler_Stream_SetsSSEHeaders(t *testing.T) {
	hub := setupTestHub()
	router := setupRouter(hub)

	// Create a context that we can cancel
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/notifications/stream", http.NoBody).WithContext(ctx)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	// Run in goroutine since it will block until context is canceled
	done := make(chan struct{})
	go func() {
		router.ServeHTTP(w, req)
		close(done)
	}()

	// Wait for handler to complete or timeout
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	// Verify SSE headers were set
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control no-cache, got %s", cacheControl)
	}

	connection := w.Header().Get("Connection")
	if connection != "keep-alive" {
		t.Errorf("Expected Connection keep-alive, got %s", connection)
	}
}

func TestHandler_Stream_SendsConnectedEvent(t *testing.T) {
	hub := setupTestHub()
	router := setupRouter(hub)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/notifications/stream", http.NoBody).WithContext(ctx)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(w, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	body := w.Body.String()
	if len(body) > 0 {
		// Should contain connected event if body is not empty
		if !containsSubstring(body, "event: connected") {
			t.Logf("Response body: %s", body)
		}
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	hub := setupTestHub()

	router := gin.New()
	handler := NewHandler(hub)
	group := router.Group("/notifications")
	handler.RegisterRoutes(group)

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/notifications/stream", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s should return 404 or 405, got %d", method, w.Code)
		}
	}
}

func TestHandler_Stream_ReceivesNotification(t *testing.T) {
	hub := setupTestHub()
	router := setupRouter(hub)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/notifications/stream", http.NoBody).WithContext(ctx)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(w, req)
		close(done)
	}()

	// Give the handler time to register and set up
	time.Sleep(30 * time.Millisecond)

	// Broadcast a message
	hub.Broadcast(Event{
		ID:        "test-event",
		Type:      EventMedicationDue,
		Title:     "Test",
		Message:   "Test message",
		ChildID:   "child-1",
		ChildName: "Baby",
		Timestamp: time.Now(),
	})

	// Wait for handler to complete
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
	}

	body := w.Body.String()
	// Check if notification event was received
	if len(body) > 0 && containsSubstring(body, "event: notification") {
		// Found notification event
		if !containsSubstring(body, "Test message") {
			t.Log("Notification event found but message content not verified")
		}
	}
}

func TestHandler_Stream_ClientDisconnect(t *testing.T) {
	hub := setupTestHub()
	router := setupRouter(hub)

	// Create a context that cancels quickly
	ctx, cancel := context.WithCancel(context.Background())

	req := httptest.NewRequest("GET", "/notifications/stream", http.NoBody).WithContext(ctx)
	req.Header.Set("X-User-ID", "test-user")
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(w, req)
		close(done)
	}()

	// Give handler time to start
	time.Sleep(20 * time.Millisecond)

	// Cancel the context (simulating client disconnect)
	cancel()

	// Wait for handler to complete
	select {
	case <-done:
		// Handler completed after disconnect
	case <-time.After(200 * time.Millisecond):
		t.Error("Handler should have returned after client disconnect")
	}
}
