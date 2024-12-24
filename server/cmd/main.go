package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"server/internal/broadcaster"
	"server/internal/chatlog"
	"server/internal/config"
	"server/internal/handler"
	"server/internal/repository"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	listener, err := net.Listen("tcp", net.JoinHostPort(cfg.Host, cfg.Port))
	if err != nil {
		slog.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer listener.Close()

	repo := repository.NewClientRepository()
	broadcaster := broadcaster.NewBroadcaster()
	chatLogger := chatlog.NewChatLogger("messages.log")

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down server...")
		broadcaster.CloseAll()
		os.Exit(0)
	}()

	slog.Info("Server started", slog.String("host", cfg.Host), slog.String("port", cfg.Port))

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Warn("Failed to accept connection", slog.String("error", err.Error()))

			continue
		}

		slog.Info("New client connected", slog.String("address", conn.RemoteAddr().String()))
		go handler.HandleConnection(conn, repo, broadcaster, chatLogger)
	}
}
