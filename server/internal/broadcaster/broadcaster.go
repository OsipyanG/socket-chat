package broadcaster

import (
	"log/slog"
	"net"
	"sync"
)

const bufferSize = 15

type Broadcaster struct {
	mu          *sync.RWMutex
	subscribers map[net.Conn]chan string
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		mu:          &sync.RWMutex{},
		subscribers: make(map[net.Conn]chan string),
	}
}

func (b *Broadcaster) Subscribe(conn net.Conn) chan string {
	ch := make(chan string, bufferSize)

	b.mu.Lock()
	b.subscribers[conn] = ch
	b.mu.Unlock()

	return ch
}

func (b *Broadcaster) Unsubscribe(conn net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscribers[conn]; ok {
		close(ch)
		delete(b.subscribers, conn)
	}
}

func (b *Broadcaster) Broadcast(message string, senderConn net.Conn) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for conn, ch := range b.subscribers {
		if conn == senderConn {
			continue
		}
		select {
		case ch <- message:
		default:
			slog.Info("Client is not ready to receive messages", "addr", conn.RemoteAddr())
		}
	}
}

func (b *Broadcaster) CloseAll() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for conn, ch := range b.subscribers {
		close(ch)
		conn.Close()
	}
}
