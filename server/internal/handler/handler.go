package handler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"server/internal/chatlog"
	"server/internal/repository"
	"server/internal/sender"
	"strings"
)

type ChatHandler struct {
	Repo       *repository.ClientRepository
	Sender     *sender.Sender
	ChatLogger *chatlog.ChatLogger
}

func NewChatHandler(repo *repository.ClientRepository, sender *sender.Sender, chatLogger *chatlog.ChatLogger) *ChatHandler {
	return &ChatHandler{
		Repo:       repo,
		Sender:     sender,
		ChatLogger: chatLogger,
	}
}

func (h *ChatHandler) HandleConnection(conn net.Conn) {
	defer h.cleanUpConnection(conn)

	nickname, err := h.registerNickname(conn)
	if err != nil {
		slog.Warn("Failed to register nickname", slog.String("error", err.Error()))
		return
	}

	h.sendUserJoinedMessage(nickname, conn)
	h.sendChatHistory(conn)

	clientChannel := h.Sender.AddSubscriber(conn)

	h.listenForMessages(conn, nickname, clientChannel)
}

func (h *ChatHandler) cleanUpConnection(conn net.Conn) {
	nickname := h.Repo.GetNickname(conn)
	h.Repo.Remove(conn)
	h.Sender.RemoveSubscriber(conn)
	conn.Close() // Закрытие соединения полностью в ответственности обработчика

	leaveMessage := fmt.Sprintf("User %s has left the chat", nickname)
	h.Sender.Broadcast(leaveMessage, nil)
}

func (h *ChatHandler) registerNickname(conn net.Conn) (string, error) {
	if err := h.Sender.SendDirect(conn, "Enter your nickname:"); err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)
	nickname, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading nickname: %w", err)
	}

	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return "", errors.New("invalid nickname")
	}

	h.Repo.Add(conn, nickname)
	return nickname, nil
}

func (h *ChatHandler) sendUserJoinedMessage(nickname string, conn net.Conn) {
	message := fmt.Sprintf("User %s joined the chat", nickname)
	h.Sender.Broadcast(message, conn)
}

func (h *ChatHandler) sendChatHistory(conn net.Conn) {
	lastMessages, err := h.ChatLogger.GetLastMessages(10)
	if err != nil {
		slog.Warn("Failed to retrieve chat history", slog.String("error", err.Error()))
		return
	}

	for _, msg := range lastMessages {
		h.Sender.SendDirect(conn, msg)
	}
}

func (h *ChatHandler) listenForMessages(conn net.Conn, nickname string, clientChannel <-chan string) {
	go func() {
		for msg := range clientChannel {
			if err := h.Sender.SendDirect(conn, msg); err != nil {
				break
			}
		}
	}()

	reader := bufio.NewReader(conn)
	for {
		inputMessage, err := reader.ReadString('\n')
		if err != nil {
			h.handleReadError(err, nickname)
			return
		}

		message := strings.TrimSpace(inputMessage)
		if strings.HasPrefix(message, "/") {
			h.handleCommand(message, conn, nickname)
		} else {
			h.broadcastMessage(nickname, message, conn)
		}
	}
}

func (h *ChatHandler) handleReadError(err error, nickname string) {
	if errors.Is(err, io.EOF) {
		slog.Info("Client disconnected", slog.String("nickname", nickname))
	} else {
		slog.Warn("Error reading message", slog.String("nickname", nickname), slog.String("error", err.Error()))
	}
}

func (h *ChatHandler) handleCommand(command string, conn net.Conn, nickname string) {
	err := h.ProcessCommand(command, conn)
	if err != nil {
		if errors.Is(err, errClientExit) {
			slog.Info("Client exited the chat", slog.String("nickname", nickname))
		} else {
			slog.Warn("Failed to process command", slog.String("command", command), slog.String("nickname", nickname), slog.String("error", err.Error()))
		}
	}
}

func (h *ChatHandler) broadcastMessage(nickname, message string, senderConn net.Conn) {
	formattedMessage := fmt.Sprintf("(%s): %s", nickname, message)
	h.Sender.Broadcast(formattedMessage, senderConn)

	if err := h.ChatLogger.SaveMessage(formattedMessage); err != nil {
		slog.Warn("Failed to save message", slog.String("error", err.Error()))
	}
}
