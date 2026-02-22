package ws

import (
	"sync"

	"gorm.io/gorm"
)

// Manager owns all Hubs, keyed by game ID.
type Manager struct {
	mu   sync.RWMutex
	hubs map[string]*Hub
}

func NewManager() *Manager {
	return &Manager{hubs: make(map[string]*Hub)}
}

// GetOrCreate returns the Hub for gameID, creating it if necessary.
func (m *Manager) GetOrCreate(gameID string, database *gorm.DB) *Hub {
	m.mu.Lock()
	defer m.mu.Unlock()
	if h, ok := m.hubs[gameID]; ok {
		return h
	}
	h := &Hub{
		gameID:  gameID,
		clients: make(map[*Client]bool),
		db:      database,
	}
	m.hubs[gameID] = h
	return h
}

// Get returns the Hub for gameID, or nil if it does not exist.
func (m *Manager) Get(gameID string) *Hub {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hubs[gameID]
}

// Hub manages all clients connected to a single game.
type Hub struct {
	gameID  string
	mu      sync.Mutex
	clients map[*Client]bool
	db      *gorm.DB
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

// Broadcast sends msg to all registered clients (non-blocking per client).
func (h *Hub) Broadcast(msg []byte) {
	h.mu.Lock()
	channels := make([]chan []byte, 0, len(h.clients))
	for c := range h.clients {
		channels = append(channels, c.send)
	}
	h.mu.Unlock()

	for _, ch := range channels {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (h *Hub) ClientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}
