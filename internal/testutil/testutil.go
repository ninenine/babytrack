package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ninenine/babytrack/internal/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestJWTManager returns a JWT manager for testing
var TestJWTManager = auth.NewJWTManager("test-secret-key-for-testing", 24*time.Hour)

// SetupTestRouter creates a new Gin router in test mode
func SetupTestRouter() *gin.Engine {
	return gin.New()
}

// SetupAuthContext adds authenticated user info to the Gin context
func SetupAuthContext(c *gin.Context, userID, email string) {
	c.Set("user_id", userID)
	c.Set("user_email", email)
	c.Set("user", &auth.User{
		ID:    userID,
		Email: email,
	})
}

// NewJSONRequest creates an HTTP request with JSON body
func NewJSONRequest(method, path string, body any) *http.Request {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// NewAuthenticatedRequest creates an HTTP request with Authorization header
func NewAuthenticatedRequest(method, path string, body any, token string) *http.Request {
	req := NewJSONRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// GenerateTestToken creates a valid JWT token for testing
func GenerateTestToken(userID, email string) string {
	token, _ := TestJWTManager.Generate(userID, email) //nolint:errcheck // Test helper - known good inputs
	return token
}

// AssertStatus checks the HTTP response status code
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("Expected status %d, got %d. Body: %s", expected, w.Code, w.Body.String())
	}
}

// AssertJSONResponse checks the status and unmarshals the response body
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, status int, target any) {
	t.Helper()
	AssertStatus(t, w, status)
	if target != nil {
		if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
			t.Errorf("Failed to unmarshal response: %v. Body: %s", err, w.Body.String())
		}
	}
}

// AssertContains checks if the response body contains a substring
func AssertContains(t *testing.T, w *httptest.ResponseRecorder, substring string) {
	t.Helper()
	if !bytes.Contains(w.Body.Bytes(), []byte(substring)) {
		t.Errorf("Expected response to contain %q, got: %s", substring, w.Body.String())
	}
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// AuthMiddlewareStub returns a middleware that sets user context without validation
func AuthMiddlewareStub(userID, email string) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetupAuthContext(c, userID, email)
		c.Next()
	}
}
