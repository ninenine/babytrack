package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/google", h.googleAuth)
	rg.GET("/google/callback", h.googleCallback)
	rg.POST("/refresh", h.refreshToken)
	rg.GET("/me", h.getCurrentUser)
}

// GET /api/auth/google - Redirect to Google OAuth
func (h *Handler) googleAuth(c *gin.Context) {
	url, state := h.service.GetGoogleAuthURL()

	// Set state in cookie for validation
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GET /api/auth/google/callback - Handle OAuth callback
func (h *Handler) googleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error="+errorParam)
		return
	}

	if code == "" || state == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=missing_params")
		return
	}

	resp, err := h.service.HandleGoogleCallback(c.Request.Context(), code, state)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=auth_failed")
		return
	}

	// Redirect to frontend with token
	c.Redirect(http.StatusTemporaryRedirect, "/login?token="+resp.Token)
}

// POST /api/auth/refresh - Refresh JWT token
func (h *Handler) refreshToken(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/auth/me - Get current user
func (h *Handler) getCurrentUser(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	user, err := h.service.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
	}

	// Try query parameter
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}
