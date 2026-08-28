package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fogshaper/cassiopeia-backend/internal/api"
	"github.com/fogshaper/cassiopeia-backend/internal/db"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connString := getenv("DATABASE_URL", "postgres://cassiopeia:cassiopeia@localhost:5432/cassiopeia?sslmode=disable")
	port := getenv("PORT", "8080")

	database, err := connectWithRetry(ctx, connString)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	if err := database.SeedIfEmpty(ctx); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	server := &api.Server{DB: database}
	handler := api.NewRouter(server)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Printf("cassiopeia-backend listening on :%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func connectWithRetry(ctx context.Context, connString string) (*db.DB, error) {
	var lastErr error
	for attempt := 0; attempt < 10; attempt++ {
		database, err := db.Connect(ctx, connString)
		if err == nil {
			return database, nil
		}
		lastErr = err
		log.Printf("database not ready (attempt %d/10): %v", attempt+1, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, lastErr
}
