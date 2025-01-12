package sender

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	bufferSize  = 15
	sendTimeout = 5 * time.Second
	maxRetries  = 3
)

type Sender struct {
	mu          sync.RWMutex
	subscribers map[net.Conn]chan string
}

func NewSender() *Sender {
	return &Sender{
		subscribers: make(map[net.Conn]chan string),
	}
}

func (s *Sender) AddSub(conn net.Conn) {
	ch := make(chan string, bufferSize)

	s.mu.Lock()
	s.subscribers[conn] = ch
	s.mu.Unlock()

	go s.startSending(conn, ch)
}

func (s *Sender) RemoveSub(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[conn]; ok {
		close(ch)
		delete(s.subscribers, conn)
	}
}

func (s *Sender) Broadcast(message string, senderConn net.Conn) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn, ch := range s.subscribers {
		if conn == senderConn {
			continue
		}

		if err := s.sendWithRetries(conn, ch, message); err != nil {
			slog.Warn("Failed to broadcast message", "addr", conn.RemoteAddr(), "error", err.Error())
		}
	}
}

func (s *Sender) SendDirect(conn net.Conn, message string) error {
	s.mu.RLock()
	ch, exists := s.subscribers[conn]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found in subscribers")
	}

	return s.sendWithRetries(conn, ch, message)
}

func (s *Sender) startSending(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		if _, err := conn.Write([]byte(msg + "\n")); err != nil {
			slog.Warn("Failed to send message to client", "addr", conn.RemoteAddr(), "error", err.Error())
			break
		}
	}
}

func (s *Sender) sendWithRetries(conn net.Conn, ch chan string, message string) error {
	var lastErr error
	ticker := time.NewTicker(sendTimeout)

	for i := range maxRetries {
		select {
		case ch <- message:
			return nil
		case <-ticker.C:
			lastErr = fmt.Errorf("timeout while sending message")
			slog.Warn("Retrying to send message", "addr", conn.RemoteAddr(), "retry", i+1, "message", message)
		}
	}

	return fmt.Errorf("failed to send message after %d retries: %w", maxRetries, lastErr)
}

func (s *Sender) CloseAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for conn, ch := range s.subscribers {
		close(ch)
		delete(s.subscribers, conn)
	}
}
