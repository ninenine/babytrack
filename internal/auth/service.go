package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Service interface {
	GetGoogleAuthURL() (url string, state string)
	HandleGoogleCallback(ctx context.Context, code, state string) (*AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
	RefreshToken(ctx context.Context, token string) (*AuthResponse, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type service struct {
	repo         Repository
	googleClient *GoogleOAuthClient
	jwtManager   *JWTManager
	states       map[string]time.Time // In production, use Redis
}

func NewService(repo Repository, googleClient *GoogleOAuthClient, jwtManager *JWTManager) Service {
	return &service{
		repo:         repo,
		googleClient: googleClient,
		jwtManager:   jwtManager,
		states:       make(map[string]time.Time),
	}
}

func (s *service) GetGoogleAuthURL() (string, string) {
	state := generateState()
	s.states[state] = time.Now().Add(10 * time.Minute)
	url := s.googleClient.GetAuthURL(state)
	return url, state
}

func (s *service) HandleGoogleCallback(ctx context.Context, code, state string) (*AuthResponse, error) {
	// Validate state
	expiry, exists := s.states[state]
	if !exists || time.Now().After(expiry) {
		return nil, fmt.Errorf("invalid or expired state")
	}
	delete(s.states, state)

	// Exchange code for tokens
	tokenResp, err := s.googleClient.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	userInfo, err := s.googleClient.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.repo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// Create new user
		user = &User{
			ID:        generateID(),
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			AvatarURL: userInfo.Picture,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if createErr := s.repo.CreateUser(ctx, user); createErr != nil {
			return nil, fmt.Errorf("failed to create user: %w", createErr)
		}
	} else {
		// Update existing user info
		user.Name = userInfo.Name
		user.AvatarURL = userInfo.Picture
		user.UpdatedAt = time.Now()
		if updateErr := s.repo.UpdateUser(ctx, user); updateErr != nil {
			return nil, fmt.Errorf("failed to update user: %w", updateErr)
		}
	}

	// Generate JWT
	token, err := s.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *service) ValidateToken(ctx context.Context, token string) (*User, error) {
	claims, err := s.jwtManager.Validate(token)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *service) RefreshToken(ctx context.Context, token string) (*AuthResponse, error) {
	newToken, err := s.jwtManager.RefreshToken(token)
	if err != nil {
		return nil, err
	}

	claims, err := s.jwtManager.Validate(newToken)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &AuthResponse{
		User:  user,
		Token: newToken,
	}, nil
}

func (s *service) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}
