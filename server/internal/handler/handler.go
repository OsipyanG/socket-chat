package handler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"server/internal/broadcaster"
	"server/internal/chatlog"
	"server/internal/repository"
	"strings"
)

func HandleConnection(conn net.Conn, repo *repository.ClientRepository, broadcaster *broadcaster.Broadcaster, chatLogger *chatlog.ChatLogger) {
	defer func() {
		sendMessage := fmt.Sprintf("User %s join to the chat", repo.GetNickname(conn))
		broadcaster.Broadcast(sendMessage, conn)

		repo.Remove(conn)
		broadcaster.Unsubscribe(conn)
		conn.Close()
	}()

	nickname, err := registerNickname(conn, repo)
	if err != nil {
		slog.Warn("Failed to register nickname", slog.String("error", err.Error()))

		return
	}
	sendMessage := fmt.Sprintf("User %s join to the chat", nickname)
	broadcaster.Broadcast(sendMessage, conn)

	clientChannel := broadcaster.Subscribe(conn)

	sendChatHistory(conn, chatLogger)

	go func() {
		for msg := range clientChannel {
			_, err := conn.Write([]byte(msg + "\n"))
			if err != nil {
				slog.Warn("Failed to send message to client", slog.String("nickname", nickname), slog.String("error", err.Error()))

				break
			}
		}
	}()

	reader := bufio.NewReader(conn)
	for {
		inputMessage, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("Client disconnected", slog.String("nickname", nickname))

				return
			}
			slog.Warn("Error reading message", slog.String("nickname", nickname), slog.String("error", err.Error()))

			continue
		}

		message := strings.TrimSpace(inputMessage)
		if strings.HasPrefix(message, "/") {
			if err := processCommand(message, conn, repo, broadcaster); err != nil {
				if errors.Is(err, errClientExit) {
					slog.Warn("Client sended /exit command")

					return
				}
				slog.Warn("Failed to process command", slog.String("command", message), slog.String("nickname", nickname), slog.String("error", err.Error()))
			}
		} else {
			sendMessage := fmt.Sprintf("(%s): %s", nickname, message)
			broadcaster.Broadcast(sendMessage, conn)

			if err := chatLogger.SaveMessage(sendMessage); err != nil {
				slog.Warn("Failed to save message", slog.String("error", err.Error()))
			}
		}
	}
}

func registerNickname(conn net.Conn, repo *repository.ClientRepository) (string, error) {
	conn.Write([]byte("Enter your nickname: \n"))
	reader := bufio.NewReader(conn)
	nickname, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return "", errors.New("invalid nickname")
	}

	repo.Add(conn, nickname)

	return nickname, nil
}

func sendChatHistory(conn net.Conn, chatLogger *chatlog.ChatLogger) {
	lastMessages, err := chatLogger.GetLastMessages(10)
	if err != nil {
		slog.Warn("Failed to send chat history", slog.String("error", err.Error()))

		return
	}

	for _, msg := range lastMessages {
		conn.Write([]byte(msg + "\n"))
	}
}
