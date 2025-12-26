package app

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"family-tracker/internal/auth"

	"github.com/gin-gonic/gin"
)

//go:embed all:web_dist
var uiFS embed.FS

type Server struct {
	cfg         *Config
	router      *gin.Engine
	httpServer  *http.Server
	authService auth.Service
	authHandler *auth.Handler
}

func NewServer(cfg *Config) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)

	// Initialize auth components
	googleClient := auth.NewGoogleOAuthClient(&auth.GoogleOAuthConfig{
		ClientID:     cfg.Auth.GoogleClientID,
		ClientSecret: cfg.Auth.GoogleClientSecret,
		RedirectURL:  cfg.Server.BaseURL + "/api/auth/google/callback",
	})

	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, 24*time.Hour)

	// For now, use a nil repo (in-memory) - will be replaced with real DB
	authRepo := auth.NewInMemoryRepository()
	authService := auth.NewService(authRepo, googleClient, jwtManager)
	authHandler := auth.NewHandler(authService)

	s := &Server{
		cfg:         cfg,
		router:      gin.New(),
		authService: authService,
		authHandler: authHandler,
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	fmt.Printf("Server starting on port %d\n", s.cfg.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) serveUI() {
	subFS, err := fs.Sub(uiFS, "web_dist")
	if err != nil {
		panic(err)
	}

	// Serve static files
	s.router.StaticFS("/assets", http.FS(subFS))

	// Serve index.html for all non-API routes (SPA)
	s.router.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}

		indexFile, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			c.String(500, "Internal Server Error")
			return
		}
		c.Data(200, "text/html; charset=utf-8", indexFile)
	})
}
