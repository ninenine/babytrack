package auth

import (
	"testing"
	"time"
)

func TestJWTManager_Generate(t *testing.T) {
	manager := NewJWTManager("test-secret-key", time.Hour)

	token, err := manager.Generate("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if token == "" {
		t.Error("Generate() returned empty token")
	}
}

func TestJWTManager_Validate(t *testing.T) {
	manager := NewJWTManager("test-secret-key", time.Hour)

	// Generate a valid token
	token, err := manager.Generate("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Validate it
	claims, err := manager.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("Validate() UserID = %v, want %v", claims.UserID, "user-123")
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Validate() Email = %v, want %v", claims.Email, "test@example.com")
	}
}

func TestJWTManager_Validate_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", time.Hour)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "not-a-valid-token"},
		{"wrong format", "header.payload.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.Validate(tt.token)
			if err != ErrInvalidToken {
				t.Errorf("Validate() error = %v, want %v", err, ErrInvalidToken)
			}
		})
	}
}

func TestJWTManager_Validate_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-1", time.Hour)
	manager2 := NewJWTManager("secret-2", time.Hour)

	// Generate with one secret
	token, err := manager1.Generate("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Validate with different secret
	_, err = manager2.Validate(token)
	if err != ErrInvalidToken {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestJWTManager_Validate_ExpiredToken(t *testing.T) {
	// Create manager with very short duration
	manager := NewJWTManager("test-secret-key", -time.Hour)

	// Generate an already-expired token
	token, err := manager.Generate("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Validate should return expired error
	_, err = manager.Validate(token)
	if err != ErrExpiredToken {
		t.Errorf("Validate() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestJWTManager_RefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", time.Hour)

	// Generate a valid token
	originalToken, err := manager.Generate("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Refresh it
	newToken, err := manager.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if newToken == "" {
		t.Error("RefreshToken() returned empty token")
	}

	// Validate the new token
	claims, err := manager.Validate(newToken)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("RefreshToken() preserved UserID = %v, want %v", claims.UserID, "user-123")
	}
}

func TestJWTManager_RefreshToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", time.Hour)

	_, err := manager.RefreshToken("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("RefreshToken() error = %v, want %v", err, ErrInvalidToken)
	}
}
