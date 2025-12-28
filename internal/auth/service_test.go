package auth

import (
	"context"
	"testing"
	"time"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	users        map[string]*User
	usersByEmail map[string]*User
	createErr    error
	updateErr    error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:        make(map[string]*User),
		usersByEmail: make(map[string]*User),
	}
}

func (m *mockRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, ok := m.usersByEmail[email]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockRepository) CreateUser(ctx context.Context, user *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockRepository) UpdateUser(ctx context.Context, user *User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func TestService_ValidateToken(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	// Create a user in the mock repo
	user := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.users[user.ID] = user

	// Generate a token
	token, _ := jwtManager.Generate(user.ID, user.Email)

	// Validate it
	validatedUser, err := svc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if validatedUser.ID != user.ID {
		t.Errorf("ValidateToken() user ID = %v, want %v", validatedUser.ID, user.ID)
	}
}

func TestService_ValidateToken_InvalidToken(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	_, err := svc.ValidateToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("ValidateToken() should return error for invalid token")
	}
}

func TestService_ValidateToken_UserNotFound(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	// Generate a valid token for a user that doesn't exist
	token, _ := jwtManager.Generate("non-existent-user", "test@example.com")

	_, err := svc.ValidateToken(context.Background(), token)
	if err == nil {
		t.Error("ValidateToken() should return error when user not found")
	}
}

func TestService_RefreshToken(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	// Create a user
	user := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.users[user.ID] = user

	// Generate an original token
	originalToken, _ := jwtManager.Generate(user.ID, user.Email)

	// Refresh it
	response, err := svc.RefreshToken(context.Background(), originalToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if response.Token == "" {
		t.Error("RefreshToken() returned empty token")
	}

	if response.User.ID != user.ID {
		t.Errorf("RefreshToken() user ID = %v, want %v", response.User.ID, user.ID)
	}
}

func TestService_RefreshToken_InvalidToken(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	_, err := svc.RefreshToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("RefreshToken() should return error for invalid token")
	}
}

func TestService_GetUserByID(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	// Create a user
	user := &User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.users[user.ID] = user

	// Get user
	retrieved, err := svc.GetUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetUserByID() returned nil for existing user")
	}

	if retrieved.ID != user.ID {
		t.Errorf("GetUserByID() ID = %v, want %v", retrieved.ID, user.ID)
	}
}

func TestService_GetUserByID_NotFound(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	user, err := svc.GetUserByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if user != nil {
		t.Error("GetUserByID() should return nil for non-existent user")
	}
}

func TestService_HandleGoogleCallback_InvalidState(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	_, err := svc.HandleGoogleCallback(context.Background(), "code", "invalid-state")
	if err == nil {
		t.Error("HandleGoogleCallback() should return error for invalid state")
	}
}

func TestService_GetGoogleAuthURL(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	googleClient := NewGoogleOAuthClient(&GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	})
	svc := NewService(repo, googleClient, jwtManager)

	url, state := svc.GetGoogleAuthURL()

	if url == "" {
		t.Error("GetGoogleAuthURL() returned empty URL")
	}

	if state == "" {
		t.Error("GetGoogleAuthURL() returned empty state")
	}

	// URL should contain the state
	if !containsSubstring(url, state) {
		t.Error("GetGoogleAuthURL() URL should contain state")
	}
}

func TestService_GetGoogleAuthURL_GeneratesUniqueStates(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	googleClient := NewGoogleOAuthClient(&GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	})
	svc := NewService(repo, googleClient, jwtManager)

	_, state1 := svc.GetGoogleAuthURL()
	_, state2 := svc.GetGoogleAuthURL()

	if state1 == state2 {
		t.Error("GetGoogleAuthURL() should generate unique states")
	}
}

func TestService_HandleGoogleCallback_ExpiredState(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	svc := NewService(repo, nil, jwtManager)

	// Access the internal service to manually add an expired state
	internalSvc := svc.(*service)
	expiredState := "expired-state-123"
	internalSvc.states[expiredState] = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago

	_, err := svc.HandleGoogleCallback(context.Background(), "code", expiredState)
	if err == nil {
		t.Error("HandleGoogleCallback() should return error for expired state")
	}

	expectedErr := "invalid or expired state"
	if err.Error() != expectedErr {
		t.Errorf("HandleGoogleCallback() error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestService_HandleGoogleCallback_ValidState_CodeExchangeFails(t *testing.T) {
	repo := newMockRepository()
	jwtManager := NewJWTManager("test-secret", time.Hour)
	googleClient := NewGoogleOAuthClient(&GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	})
	svc := NewService(repo, googleClient, jwtManager)

	// Get a valid state
	_, state := svc.GetGoogleAuthURL()

	// Try callback with invalid code (will fail because Google returns error)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := svc.HandleGoogleCallback(ctx, "invalid-code", state)
	if err == nil {
		t.Error("HandleGoogleCallback() should return error when code exchange fails")
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
