package main

import (
	"context"
	"log/slog"
	"os/signal"
	"server/internal/app"
	"server/internal/config"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.String("error", err.Error()))
		return
	}

	app, err := app.NewApp(cfg)
	if err != nil {
		slog.Error("Failed to initialize application", slog.String("error", err.Error()))
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go app.Start(ctx)

	<-ctx.Done()

	slog.Info("Shutting down server...")
	app.Stop()
	slog.Info("Server gracefully stopped")
}
