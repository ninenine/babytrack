package app

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"family-tracker/internal/auth"
	"family-tracker/internal/db"
	"family-tracker/internal/family"
	"family-tracker/internal/feeding"
	"family-tracker/internal/medication"
	"family-tracker/internal/sleep"

	"github.com/gin-gonic/gin"
)

//go:embed all:web_dist
var uiFS embed.FS

type Server struct {
	cfg               *Config
	db                *db.DB
	router            *gin.Engine
	httpServer        *http.Server
	authService       auth.Service
	authHandler       *auth.Handler
	familyHandler     *family.Handler
	feedingHandler    *feeding.Handler
	sleepHandler      *sleep.Handler
	medicationHandler *medication.Handler
}

func NewServer(cfg *Config, database *db.DB) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)

	// Initialize auth components
	googleClient := auth.NewGoogleOAuthClient(&auth.GoogleOAuthConfig{
		ClientID:     cfg.Auth.GoogleClientID,
		ClientSecret: cfg.Auth.GoogleClientSecret,
		RedirectURL:  cfg.Server.BaseURL + "/api/auth/google/callback",
	})

	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, 24*time.Hour)

	// Use postgres repository
	authRepo := auth.NewPostgresRepository(database.DB)
	authService := auth.NewService(authRepo, googleClient, jwtManager)
	authHandler := auth.NewHandler(authService)

	// Initialize family components
	familyRepo := family.NewRepository(database.DB)
	familyService := family.NewService(familyRepo)
	familyHandler := family.NewHandler(familyService)

	// Initialize feeding components
	feedingRepo := feeding.NewRepository(database.DB)
	feedingService := feeding.NewService(feedingRepo)
	feedingHandler := feeding.NewHandler(feedingService)

	// Initialize sleep components
	sleepRepo := sleep.NewRepository(database.DB)
	sleepService := sleep.NewService(sleepRepo)
	sleepHandler := sleep.NewHandler(sleepService)

	// Initialize medication components
	medicationRepo := medication.NewRepository(database.DB)
	medicationService := medication.NewService(medicationRepo)
	medicationHandler := medication.NewHandler(medicationService)

	s := &Server{
		cfg:               cfg,
		db:                database,
		router:            gin.New(),
		authService:       authService,
		authHandler:       authHandler,
		familyHandler:     familyHandler,
		feedingHandler:    feedingHandler,
		sleepHandler:      sleepHandler,
		medicationHandler: medicationHandler,
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

	fileServer := http.FileServer(http.FS(subFS))

	s.router.NoRoute(func(c *gin.Context) {
		reqPath := c.Request.URL.Path

		// Don't serve UI for API routes
		if strings.HasPrefix(reqPath, "/api") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}

		// Try to serve the file directly
		filePath := strings.TrimPrefix(reqPath, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Check if file exists
		if f, err := subFS.Open(filePath); err == nil {
			f.Close()
			// Set correct content type based on extension
			ext := path.Ext(filePath)
			switch ext {
			case ".js":
				c.Header("Content-Type", "application/javascript")
			case ".css":
				c.Header("Content-Type", "text/css")
			case ".svg":
				c.Header("Content-Type", "image/svg+xml")
			case ".png":
				c.Header("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				c.Header("Content-Type", "image/jpeg")
			case ".woff":
				c.Header("Content-Type", "font/woff")
			case ".woff2":
				c.Header("Content-Type", "font/woff2")
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// File not found, serve index.html for SPA routing
		indexFile, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			c.String(500, "Internal Server Error")
			return
		}
		c.Data(200, "text/html; charset=utf-8", indexFile)
	})
}
