package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wsppppp/manage-pull-request/internal/app"
	"github.com/wsppppp/manage-pull-request/internal/config"
	"github.com/wsppppp/manage-pull-request/internal/repository/postgres"
	transport "github.com/wsppppp/manage-pull-request/internal/transport/http"
	"github.com/wsppppp/manage-pull-request/pkg/database"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewFromEnv()

	log.Println("Running database migrations...")
	migrationURL := cfg.MigrationURL() + "&x-migrations-table=schema_migrations"
	m, err := migrate.New("file://migrations", migrationURL)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to apply migrations: %v", err)
	}
	log.Println("Migrations applied successfully")

	dbPool, err := database.NewClient(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	repo := postgres.New(dbPool)
	service := app.New(repo)
	handler := transport.NewHandler(service)
	router := handler.NewRouter()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Server is starting on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
