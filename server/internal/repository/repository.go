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
	defer r.mu.Unlock()
	r.clients[conn] = nickname
}

func (r *ClientRepository) Remove(conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, conn)
}

func (r *ClientRepository) GetNickname(conn net.Conn) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.clients[conn]
}
