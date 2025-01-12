package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"server/internal/chatlog"
	"server/internal/config"
	"server/internal/handler"
	"server/internal/repository"
	"server/internal/sender"
)

type App struct {
	listener net.Listener
	handler  *handler.ChatHandler
}

func NewApp(cfg *config.Config) (*App, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}

	repo := repository.NewClientRepository()
	sender := sender.NewSender()
	chatLogger, err := chatlog.NewChatLogger("messages.log")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat logger: %w", err)
	}

	chatHandler := handler.NewChatHandler(repo, sender, chatLogger)

	return &App{
		listener: listener,
		handler:  chatHandler,
	}, nil
}

func (app *App) Start(ctx context.Context) {
	slog.Info("Server started", slog.String("address", app.listener.Addr().String()))

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := app.listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.Warn("Failed to accept connection", slog.String("error", err.Error()))
				continue
			}

			slog.Info("New client connected", slog.String("address", conn.RemoteAddr().String()))
			go app.handler.HandleConnection(conn)
		}
	}
}

func (app *App) Stop() {
	if err := app.listener.Close(); err != nil {
		slog.Warn("Failed to close listener", slog.String("error", err.Error()))
	}

	if err := app.handler.ChatLogger.Close(); err != nil {
		slog.Warn("Failed to close ChatLogger", slog.String("error", err.Error()))
	}

	app.handler.Repo.CloseAllConnections()
	app.handler.Sender.CloseAll()
}
