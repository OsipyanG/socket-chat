package handler

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	errClientExit       = errors.New("client disconnected")
	errUnknownCommand   = errors.New("unknown command")
	countOfCommandParts = 2
)

// ProcessCommand обрабатывает команды
func (h *ChatHandler) ProcessCommand(command string, conn net.Conn) error {
	parts := strings.SplitN(command, " ", countOfCommandParts)
	cmd := parts[0]

	switch cmd {
	case "/nick":
		return h.handleNickCommand(parts, conn)
	case "/exit":
		return errClientExit
	default:
		return h.handleUnknownCommand(conn)
	}
}

func (h *ChatHandler) handleNickCommand(parts []string, conn net.Conn) error {
	if len(parts) < countOfCommandParts {
		_, _ = conn.Write([]byte("Usage: /nick <new_nickname>\n"))
		return nil
	}

	newNickname := strings.TrimSpace(parts[1])
	oldNickname := h.Repo.GetNickname(conn)

	h.Repo.Remove(conn)
	h.Repo.Add(conn, newNickname)

	message := fmt.Sprintf("User changed nickname from %s to %s", oldNickname, newNickname)
	h.Sender.Broadcast(message, conn)
	return nil
}

func (h *ChatHandler) handleUnknownCommand(conn net.Conn) error {
	_, _ = conn.Write([]byte("Unknown command\n"))
	return errUnknownCommand
}
