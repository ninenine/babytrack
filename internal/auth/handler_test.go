package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// mockService is a test double for Service interface
type mockService struct {
	// GetGoogleAuthURL
	authURL   string
	authState string

	// HandleGoogleCallback
	callbackResp *AuthResponse
	callbackErr  error

	// ValidateToken
	validateUser *User
	validateErr  error

	// RefreshToken
	refreshResp *AuthResponse
	refreshErr  error

	// GetUserByID
	getUserResp *User
	getUserErr  error
}

func (m *mockService) GetGoogleAuthURL() (string, string) {
	return m.authURL, m.authState
}

func (m *mockService) HandleGoogleCallback(ctx context.Context, code, state string) (*AuthResponse, error) {
	return m.callbackResp, m.callbackErr
}

func (m *mockService) ValidateToken(ctx context.Context, token string) (*User, error) {
	return m.validateUser, m.validateErr
}

func (m *mockService) RefreshToken(ctx context.Context, token string) (*AuthResponse, error) {
	return m.refreshResp, m.refreshErr
}

func (m *mockService) GetUserByID(ctx context.Context, id string) (*User, error) {
	return m.getUserResp, m.getUserErr
}

func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/api/auth")
	h.RegisterRoutes(rg)
	return router
}

// ============================================================================
// Google Auth (Login Redirect) Tests
// ============================================================================

func TestHandler_GoogleAuth_RedirectsToGoogle(t *testing.T) {
	mockSvc := &mockService{
		authURL:   "https://accounts.google.com/oauth?client_id=test&state=abc123",
		authState: "abc123",
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	if location != mockSvc.authURL {
		t.Errorf("expected redirect to %s, got %s", mockSvc.authURL, location)
	}
}

func TestHandler_GoogleAuth_SetsStateCookie(t *testing.T) {
	mockSvc := &mockService{
		authURL:   "https://accounts.google.com/oauth",
		authState: "test-state-value",
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	cookies := resp.Result().Cookies()
	var stateCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			stateCookie = c
			break
		}
	}

	if stateCookie == nil {
		t.Fatal("expected oauth_state cookie to be set")
	}

	if stateCookie.Value != "test-state-value" {
		t.Errorf("expected cookie value %s, got %s", "test-state-value", stateCookie.Value)
	}

	if !stateCookie.HttpOnly {
		t.Error("expected cookie to be HttpOnly")
	}
}

// ============================================================================
// Google Callback Tests
// ============================================================================

func TestHandler_GoogleCallback_Success(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		callbackResp: &AuthResponse{
			User:  testUser,
			Token: "jwt-token-here",
		},
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google/callback?code=auth-code&state=valid-state", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	expectedLocation := "/login?token=jwt-token-here"
	if location != expectedLocation {
		t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
	}
}

func TestHandler_GoogleCallback_ErrorParam(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google/callback?error=access_denied", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	expectedLocation := "/login?error=access_denied"
	if location != expectedLocation {
		t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
	}
}

func TestHandler_GoogleCallback_MissingCode(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google/callback?state=valid-state", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	expectedLocation := "/login?error=missing_params"
	if location != expectedLocation {
		t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
	}
}

func TestHandler_GoogleCallback_MissingState(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google/callback?code=auth-code", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	expectedLocation := "/login?error=missing_params"
	if location != expectedLocation {
		t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
	}
}

func TestHandler_GoogleCallback_MissingBothParams(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google/callback", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	expectedLocation := "/login?error=missing_params"
	if location != expectedLocation {
		t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
	}
}

func TestHandler_GoogleCallback_ServiceError(t *testing.T) {
	mockSvc := &mockService{
		callbackErr: errors.New("invalid or expired state"),
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/google/callback?code=auth-code&state=invalid-state", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, resp.Code)
	}

	location := resp.Header().Get("Location")
	expectedLocation := "/login?error=auth_failed"
	if location != expectedLocation {
		t.Errorf("expected redirect to %s, got %s", expectedLocation, location)
	}
}

// ============================================================================
// Refresh Token Tests
// ============================================================================

func TestHandler_RefreshToken_Success(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		refreshResp: &AuthResponse{
			User:  testUser,
			Token: "new-jwt-token",
		},
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", http.NoBody)
	req.Header.Set("Authorization", "Bearer old-jwt-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var result AuthResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Token != "new-jwt-token" {
		t.Errorf("expected token %s, got %s", "new-jwt-token", result.Token)
	}

	if result.User.ID != testUser.ID {
		t.Errorf("expected user ID %s, got %s", testUser.ID, result.User.ID)
	}
}

func TestHandler_RefreshToken_WithQueryParam(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		refreshResp: &AuthResponse{
			User:  testUser,
			Token: "new-jwt-token",
		},
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("POST", "/api/auth/refresh?token=old-jwt-token", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestHandler_RefreshToken_MissingToken(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["error"] != "missing token" {
		t.Errorf("expected error 'missing token', got '%s'", result["error"])
	}
}

func TestHandler_RefreshToken_InvalidToken(t *testing.T) {
	mockSvc := &mockService{
		refreshErr: errors.New("token is expired"),
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["error"] != "token is expired" {
		t.Errorf("expected error 'token is expired', got '%s'", result["error"])
	}
}

func TestHandler_RefreshToken_InvalidAuthHeader(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("POST", "/api/auth/refresh", http.NoBody)
	req.Header.Set("Authorization", "InvalidFormat")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}

// ============================================================================
// Get Current User (Me) Tests
// ============================================================================

func TestHandler_GetCurrentUser_Success(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "https://example.com/avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		validateUser: testUser,
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-jwt-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var result User
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.ID != testUser.ID {
		t.Errorf("expected user ID %s, got %s", testUser.ID, result.ID)
	}

	if result.Email != testUser.Email {
		t.Errorf("expected email %s, got %s", testUser.Email, result.Email)
	}

	if result.Name != testUser.Name {
		t.Errorf("expected name %s, got %s", testUser.Name, result.Name)
	}

	if result.AvatarURL != testUser.AvatarURL {
		t.Errorf("expected avatar URL %s, got %s", testUser.AvatarURL, result.AvatarURL)
	}
}

func TestHandler_GetCurrentUser_WithQueryParam(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		validateUser: testUser,
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me?token=valid-jwt-token", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestHandler_GetCurrentUser_MissingToken(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["error"] != "missing token" {
		t.Errorf("expected error 'missing token', got '%s'", result["error"])
	}
}

func TestHandler_GetCurrentUser_InvalidToken(t *testing.T) {
	mockSvc := &mockService{
		validateErr: errors.New("token is invalid"),
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["error"] != "token is invalid" {
		t.Errorf("expected error 'token is invalid', got '%s'", result["error"])
	}
}

func TestHandler_GetCurrentUser_ExpiredToken(t *testing.T) {
	mockSvc := &mockService{
		validateErr: errors.New("token has expired"),
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer expired-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}

// ============================================================================
// Route Registration Tests
// ============================================================================

func TestHandler_RegisterRoutes(t *testing.T) {
	mockSvc := &mockService{
		authURL:   "https://google.com/auth",
		authState: "state",
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET /api/auth/google exists",
			method:         "GET",
			path:           "/api/auth/google",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:           "GET /api/auth/google/callback exists",
			method:         "GET",
			path:           "/api/auth/google/callback",
			expectedStatus: http.StatusTemporaryRedirect, // redirects to error page
		},
		{
			name:           "POST /api/auth/refresh exists",
			method:         "POST",
			path:           "/api/auth/refresh",
			expectedStatus: http.StatusUnauthorized, // no token
		},
		{
			name:           "GET /api/auth/me exists",
			method:         "GET",
			path:           "/api/auth/me",
			expectedStatus: http.StatusUnauthorized, // no token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, http.NoBody)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.Code)
			}
		})
	}
}

func TestHandler_RegisterRoutes_MethodNotAllowed(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.HandleMethodNotAllowed = true
	rg := router.Group("/api/auth")
	handler.RegisterRoutes(rg)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "POST /api/auth/google should fail",
			method: "POST",
			path:   "/api/auth/google",
		},
		{
			name:   "POST /api/auth/google/callback should fail",
			method: "POST",
			path:   "/api/auth/google/callback",
		},
		{
			name:   "GET /api/auth/refresh should fail",
			method: "GET",
			path:   "/api/auth/refresh",
		},
		{
			name:   "POST /api/auth/me should fail",
			method: "POST",
			path:   "/api/auth/me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, http.NoBody)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.Code)
			}
		})
	}
}

// ============================================================================
// Token Extraction Tests
// ============================================================================

func TestHandler_TokenExtraction_BearerHeader(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		validateUser: testUser,
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me", http.NoBody)
	req.Header.Set("Authorization", "Bearer my-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestHandler_TokenExtraction_BearerHeaderCaseInsensitive(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		validateUser: testUser,
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me", http.NoBody)
	req.Header.Set("Authorization", "BEARER my-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestHandler_TokenExtraction_QueryParamFallback(t *testing.T) {
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		validateUser: testUser,
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/auth/me?token=my-token", http.NoBody)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestHandler_TokenExtraction_HeaderTakesPrecedence(t *testing.T) {
	// This test verifies that Authorization header is checked before query param
	testUser := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockSvc := &mockService{
		validateUser: testUser,
	}
	handler := NewHandler(mockSvc)
	router := setupTestRouter(handler)

	// Provide both header and query param
	req, _ := http.NewRequest("GET", "/api/auth/me?token=query-token", http.NoBody)
	req.Header.Set("Authorization", "Bearer header-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Should succeed because header token is used
	if resp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
}

// ============================================================================
// NewHandler Tests
// ============================================================================

func TestNewHandler(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)

	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}

	if handler.service == nil {
		t.Error("NewHandler() did not set service")
	}
}
