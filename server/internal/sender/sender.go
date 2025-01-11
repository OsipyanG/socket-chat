package sender

import (
	"log/slog"
	"net"
	"sync"
)

const bufferSize = 15

type Sender struct {
	mu          *sync.RWMutex
	subscribers map[net.Conn]chan string
}

func NewSender() *Sender {
	return &Sender{
		mu:          &sync.RWMutex{},
		subscribers: make(map[net.Conn]chan string),
	}
}

// AddSubscriber добавляет клиента в список подписчиков
func (s *Sender) AddSubscriber(conn net.Conn) chan string {
	ch := make(chan string, bufferSize)

	s.mu.Lock()
	s.subscribers[conn] = ch
	s.mu.Unlock()

	go s.startSending(conn, ch)

	return ch
}

// RemoveSubscriber удаляет клиента из списка подписчиков (не закрывает conn)
func (s *Sender) RemoveSubscriber(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[conn]; ok {
		close(ch)
		delete(s.subscribers, conn)
	}
}

// Broadcast отправляет сообщение всем подписчикам, кроме отправителя
func (s *Sender) Broadcast(message string, senderConn net.Conn) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn, ch := range s.subscribers {
		if conn == senderConn {
			continue
		}
		select {
		case ch <- message:
		default:
			slog.Warn("Client is not ready to receive messages", "addr", conn.RemoteAddr())
		}
	}
}

// SendDirect отправляет сообщение напрямую конкретному клиенту
func (s *Sender) SendDirect(conn net.Conn, message string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.subscribers[conn]; !exists {
		return nil // Неактивный клиент
	}

	_, err := conn.Write([]byte(message + "\n"))
	if err != nil {
		slog.Warn("Failed to send direct message", "addr", conn.RemoteAddr(), "error", err.Error())
	}
	return err
}

// startSending запускает горутину для отправки сообщений клиенту
func (s *Sender) startSending(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		if _, err := conn.Write([]byte(msg + "\n")); err != nil {
			slog.Warn("Failed to send message to client", "addr", conn.RemoteAddr(), "error", err.Error())
			break
		}
	}
}
