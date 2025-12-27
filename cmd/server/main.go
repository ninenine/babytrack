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
	defer database.Close()

	// Run migrations
	log.Println("running database migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
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
