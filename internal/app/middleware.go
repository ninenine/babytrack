package app

import (
	"strings"

	"github.com/ninenine/babytrack/internal/auth"

	"github.com/gin-gonic/gin"
)

func (s *Server) setupMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(s.corsMiddleware())
	s.router.Use(s.requestLogger())
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (s *Server) requestLogger() gin.HandlerFunc {
	return gin.Logger()
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization token"})
			return
		}

		user, err := s.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			if err == auth.ErrExpiredToken {
				c.AbortWithStatusJSON(401, gin.H{"error": "token expired"})
				return
			}
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user", user)

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Try query parameter (for WebSocket connections)
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}
