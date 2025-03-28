package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"maragu.dev/env"

	"app/ai"
	"app/http"
	"app/sql"
)

func main() {
	// Set up a logger that is used throughout the app
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Start the app, exit with a non-zero exit code on errors
	if err := start(log); err != nil {
		log.Error("Error", "error", err)
		os.Exit(1)
	}
}

func start(log *slog.Logger) error {
	log.Info("Starting app")

	// We load environment variables from .env if it exists
	_ = env.Load()

	// Catch signals to gracefully shut down the app
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Set up the database, which is injected as a dependency into the HTTP server
	db := sql.NewDatabase(sql.NewDatabaseOptions{
		Log:  log,
		Path: env.GetStringOrDefault("DATABASE_PATH", "app.db"),
	})
	if err := db.Connect(); err != nil {
		return err
	}
	if err := db.MigrateUp(ctx); err != nil {
		return err
	}

	// Set up the AI client for chat completion and embeddings
	ai := ai.NewClient(ai.NewClientOptions{
		Log:                  log,
		ChatCompleterBaseURL: env.GetStringOrDefault("AI_CHAT_COMPLETER_BASE_URL", "http://localhost:8081/v1"),
		EmbedderBaseURL:      env.GetStringOrDefault("AI_EMBEDDER_BASE_URL", "http://localhost:8082/v1"),
	})

	// Set up the HTTP server, injecting the database, AI client, and logger
	s := http.NewServer(http.NewServerOptions{
		AI:  ai,
		DB:  db,
		Log: log,
	})

	// Use an errgroup to wait for separate goroutines which can error
	eg, ctx := errgroup.WithContext(ctx)

	// Start the server within the errgroup.
	// You can do this for other dependencies as well.
	eg.Go(func() error {
		return s.Start()
	})

	// Wait for the context to be done, which happens when a signal is caught
	<-ctx.Done()
	log.Info("Stopping app")

	// Stop the server gracefully
	eg.Go(func() error {
		return s.Stop()
	})

	// Wait for the server to stop
	if err := eg.Wait(); err != nil {
		return err
	}

	log.Info("Stopped app")

	return nil
}
