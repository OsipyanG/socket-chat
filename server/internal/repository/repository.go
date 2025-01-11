package repository

import (
	"net"
	"sync"
)

type ClientRepository struct {
	mu      *sync.RWMutex
	clients map[net.Conn]string
}

func NewClientRepository() *ClientRepository {
	return &ClientRepository{
		mu:      &sync.RWMutex{},
		clients: make(map[net.Conn]string),
	}
}

func (r *ClientRepository) Add(conn net.Conn, nickname string) {
	r.mu.Lock()
	r.clients[conn] = nickname
	r.mu.Unlock()
}

func (r *ClientRepository) Remove(conn net.Conn) {
	r.mu.Lock()
	delete(r.clients, conn)
	r.mu.Unlock()
}

func (r *ClientRepository) GetNickname(conn net.Conn) string {
	r.mu.RLock()
	nickname := r.clients[conn]
	r.mu.RUnlock()

	return nickname
}
