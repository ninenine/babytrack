package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ninenine/babytrack/internal/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockAuthService is a test double for auth.Service
type mockAuthService struct {
	validateTokenFn func(ctx context.Context, token string) (*auth.User, error)
}

func (m *mockAuthService) GetGoogleAuthURL() (string, string) {
	return "https://google.com/auth", "test-state"
}

func (m *mockAuthService) HandleGoogleCallback(ctx context.Context, code, state string) (*auth.AuthResponse, error) {
	return nil, nil
}

func (m *mockAuthService) ValidateToken(ctx context.Context, token string) (*auth.User, error) {
	if m.validateTokenFn != nil {
		return m.validateTokenFn(ctx, token)
	}
	return &auth.User{ID: "user-123", Email: "test@example.com"}, nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, token string) (*auth.AuthResponse, error) {
	return nil, nil
}

func (m *mockAuthService) GetUserByID(ctx context.Context, id string) (*auth.User, error) {
	return nil, nil
}

// createTestServer creates a minimal server for testing middleware
func createTestServer(authService auth.Service) *Server {
	return &Server{
		router:      gin.New(),
		authService: authService,
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockService := &mockAuthService{
		validateTokenFn: func(ctx context.Context, token string) (*auth.User, error) {
			return &auth.User{ID: "user-123", Email: "test@example.com"}, nil
		},
	}
	server := createTestServer(mockService)

	router := gin.New()
	router.Use(server.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID := c.GetString("user_id")
		email := c.GetString("user_email")
		c.JSON(200, gin.H{"user_id": userID, "email": email})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	mockService := &mockAuthService{}
	server := createTestServer(mockService)

	router := gin.New()
	router.Use(server.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockService := &mockAuthService{
		validateTokenFn: func(ctx context.Context, token string) (*auth.User, error) {
			return nil, errors.New("invalid token")
		},
	}
	server := createTestServer(mockService)

	router := gin.New()
	router.Use(server.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	mockService := &mockAuthService{
		validateTokenFn: func(ctx context.Context, token string) (*auth.User, error) {
			return nil, auth.ErrExpiredToken
		},
	}
	server := createTestServer(mockService)

	router := gin.New()
	router.Use(server.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_TokenFromQuery(t *testing.T) {
	mockService := &mockAuthService{
		validateTokenFn: func(ctx context.Context, token string) (*auth.User, error) {
			if token == "query-token" {
				return &auth.User{ID: "user-456", Email: "query@example.com"}, nil
			}
			return nil, errors.New("invalid token")
		},
	}
	server := createTestServer(mockService)

	router := gin.New()
	router.Use(server.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(200, gin.H{"user_id": userID})
	})

	req := httptest.NewRequest("GET", "/test?token=query-token", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_SetsUserInContext(t *testing.T) {
	testUser := &auth.User{ID: "user-789", Email: "context@example.com", Name: "Test User"}
	mockService := &mockAuthService{
		validateTokenFn: func(ctx context.Context, token string) (*auth.User, error) {
			return testUser, nil
		},
	}
	server := createTestServer(mockService)

	var capturedUserID, capturedEmail string
	var capturedUser *auth.User

	router := gin.New()
	router.Use(server.authMiddleware())
	router.GET("/test", func(c *gin.Context) {
		capturedUserID = c.GetString("user_id")
		capturedEmail = c.GetString("user_email")
		if u, exists := c.Get("user"); exists {
			capturedUser = u.(*auth.User)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "user-789" {
		t.Errorf("Expected user_id user-789, got %s", capturedUserID)
	}
	if capturedEmail != "context@example.com" {
		t.Errorf("Expected email context@example.com, got %s", capturedEmail)
	}
	if capturedUser == nil || capturedUser.ID != "user-789" {
		t.Error("Expected user object in context")
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	server := createTestServer(&mockAuthService{})
	server.router.Use(server.corsMiddleware())
	server.router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestCORSMiddleware_AllowOrigin(t *testing.T) {
	server := createTestServer(&mockAuthService{})
	server.router.Use(server.corsMiddleware())
	server.router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin *, got %s", allowOrigin)
	}

	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Expected Access-Control-Allow-Methods to be set")
	}

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders == "" {
		t.Error("Expected Access-Control-Allow-Headers to be set")
	}
}

func TestExtractToken_BearerHeader(t *testing.T) {
	router := gin.New()
	var extractedToken string

	router.GET("/test", func(c *gin.Context) {
		extractedToken = extractToken(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer my-token-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if extractedToken != "my-token-123" {
		t.Errorf("Expected token my-token-123, got %s", extractedToken)
	}
}

func TestExtractToken_BearerHeaderCaseInsensitive(t *testing.T) {
	router := gin.New()
	var extractedToken string

	router.GET("/test", func(c *gin.Context) {
		extractedToken = extractToken(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "BEARER my-token-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if extractedToken != "my-token-123" {
		t.Errorf("Expected token my-token-123, got %s", extractedToken)
	}
}

func TestExtractToken_QueryParam(t *testing.T) {
	router := gin.New()
	var extractedToken string

	router.GET("/test", func(c *gin.Context) {
		extractedToken = extractToken(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test?token=query-token-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if extractedToken != "query-token-456" {
		t.Errorf("Expected token query-token-456, got %s", extractedToken)
	}
}

func TestExtractToken_HeaderTakesPrecedence(t *testing.T) {
	router := gin.New()
	var extractedToken string

	router.GET("/test", func(c *gin.Context) {
		extractedToken = extractToken(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test?token=query-token", http.NoBody)
	req.Header.Set("Authorization", "Bearer header-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if extractedToken != "header-token" {
		t.Errorf("Expected header token to take precedence, got %s", extractedToken)
	}
}

func TestExtractToken_NoToken(t *testing.T) {
	router := gin.New()
	var extractedToken string

	router.GET("/test", func(c *gin.Context) {
		extractedToken = extractToken(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if extractedToken != "" {
		t.Errorf("Expected empty token, got %s", extractedToken)
	}
}

func TestExtractToken_InvalidAuthHeaderFormat(t *testing.T) {
	router := gin.New()
	var extractedToken string

	router.GET("/test", func(c *gin.Context) {
		extractedToken = extractToken(c)
		c.JSON(200, gin.H{"ok": true})
	})

	// Only "Basic" scheme, not Bearer
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if extractedToken != "" {
		t.Errorf("Expected empty token for Basic auth, got %s", extractedToken)
	}
}
