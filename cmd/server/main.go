package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ninenine/babytrack/internal/app"
	"github.com/ninenine/babytrack/internal/db"
)

func main() {
	configPath := flag.String("config", "./configs/config.yaml", "path to config file")
	migrateOnly := flag.Bool("migrate", false, "run migrations and exit")
	flag.Parse()

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.New(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			log.Printf("error closing database: %v", closeErr)
		}
	}()

	// Run migrations
	log.Println("running database migrations...")
	if migrateErr := database.Migrate(); migrateErr != nil {
		log.Fatalf("failed to run migrations: %v", migrateErr) //nolint:gocritic // Acceptable in CLI - OS closes db on exit
	}
	log.Println("migrations completed")

	if *migrateOnly {
		log.Println("migrate-only mode, exiting")
		return
	}

	srv, err := app.NewServer(cfg, database)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	if err := srv.Shutdown(); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
}
