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
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		runSeed(os.Args[2:])
		return
	}

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

	server := &api.Server{DB: database}
	handler := api.NewRouter(server)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Printf("cassiopeia-backend is listening on :%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

// runSeed is the entrypoint for the `cassiopeia-backend seed <initial|personal>`
// CLI invocation used by the initial-seed Helm hook Job and the local
// personal-seed Job (see cassiopeia-cluster/backend).
func runSeed(args []string) {
	if len(args) != 1 || (args[0] != "initial" && args[0] != "personal") {
		log.Fatalf("usage: cassiopeia-backend seed [initial|personal]")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connString := getenv("DATABASE_URL", "postgres://cassiopeia:cassiopeia@localhost:5432/cassiopeia?sslmode=disable")

	database, err := connectWithRetry(ctx, connString)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	switch args[0] {
	case "initial":
		if err := database.SeedInitial(ctx); err != nil {
			log.Fatalf("failed to run initial seed: %v", err)
		}
	case "personal":
		if err := database.SeedPersonal(ctx); err != nil {
			log.Fatalf("failed to run personal seed: %v", err)
		}
	}
	log.Printf("seed %s complete", args[0])
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
