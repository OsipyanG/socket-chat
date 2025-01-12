package clients

import (
	"net"
	"sync"
)

type ConnectionRegistry struct {
	mu      *sync.RWMutex
	clients map[net.Conn]string
}

func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		mu:      &sync.RWMutex{},
		clients: make(map[net.Conn]string),
	}
}

func (r *ConnectionRegistry) Register(conn net.Conn, nickname string) {
	r.mu.Lock()
	r.clients[conn] = nickname
	r.mu.Unlock()
}

func (r *ConnectionRegistry) Unregister(conn net.Conn) {
	r.mu.Lock()
	delete(r.clients, conn)
	conn.Close()
	r.mu.Unlock()
}

func (r *ConnectionRegistry) LookupNickname(conn net.Conn) string {
	r.mu.RLock()
	nickname := r.clients[conn]
	r.mu.RUnlock()
	return nickname
}

func (r *ConnectionRegistry) TerminateAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for conn := range r.clients {
		delete(r.clients, conn)
		conn.Close()
	}
}
