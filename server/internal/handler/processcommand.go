package handler

import (
	"errors"
	"fmt"
	"net"
	"server/internal/broadcaster"
	"server/internal/repository"
	"strings"
)

var errClientExit = errors.New("client disconnected")

const countOfCommandParts = 2

func processCommand(command string, conn net.Conn, repo *repository.ClientRepository, broadcaster *broadcaster.Broadcaster) error {
	parts := strings.SplitN(command, " ", countOfCommandParts)
	cmd := parts[0]

	switch cmd {
	case "/nick":
		if len(parts) < countOfCommandParts {
			_, _ = conn.Write([]byte("Usage: /nick <new_nickname>\n"))

			return nil
		}
		newNickname := strings.TrimSpace(parts[1])
		repo.Remove(conn)
		repo.Add(conn, newNickname)
		broadcaster.Broadcast(fmt.Sprintf("User changed nickname fron %s to %s", repo.GetNickname(conn), newNickname), conn)
	case "/exit":
		return errClientExit
	default:
		_, _ = conn.Write([]byte("Unknown command\n"))
	}

	return nil
}
