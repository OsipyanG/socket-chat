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

func (h *ChatHandler) ProcessCommand(command string, conn net.Conn) error {
	parts := strings.SplitN(command, " ", countOfCommandParts)
	cmd := parts[0]

	switch cmd {
	case "/nick":
		return h.handleNickCommand(parts, conn)
	default:
		return h.handleUnknownCommand(conn)
	}
}

func (h *ChatHandler) handleNickCommand(parts []string, conn net.Conn) error {
	if len(parts) < countOfCommandParts {
		h.Sender.SendDirect(conn, "Usage: /nick <new_nickname>")
		return nil
	}

	newNickname := strings.TrimSpace(parts[1])
	oldNickname := h.Registry.LookupNickname(conn)
	h.Registry.Register(conn, newNickname)

	message := fmt.Sprintf("User changed nickname from %s to %s", oldNickname, newNickname)
	h.Sender.Broadcast(message, conn)
	return nil
}

func (h *ChatHandler) handleUnknownCommand(conn net.Conn) error {
	h.Sender.SendDirect(conn, "Unknown command")
	return errUnknownCommand
}
