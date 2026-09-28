// Package realtime is the push-only WebSocket listener. Browsers redeem a
// single-use ticket for a socket; the hub then forwards to it every bus event
// published on the channels its ticket entitles it to. Sockets only receive:
// what changed is fetched again through the API, where authorization lives.
package realtime

import (
	"sync"

	"github.com/google/uuid"
)

// client is one open socket and the channels it hears.
type client struct {
	userID   uuid.UUID
	channels []string
	send     chan []byte
	// slow is closed, once, when the socket fell behind and must be dropped.
	slow     chan struct{}
	slowOnce sync.Once
}

func newClient(userID uuid.UUID, channels []string, buffer int) *client {
	return &client{
		userID:   userID,
		channels: channels,
		send:     make(chan []byte, buffer),
		slow:     make(chan struct{}),
	}
}

func (c *client) markSlow() {
	c.slowOnce.Do(func() { close(c.slow) })
}

// Hub fans events out to the sockets of this process.
type Hub struct {
	mu         sync.RWMutex
	byChannel  map[string]map[*client]struct{}
	perUser    map[uuid.UUID]int
	maxPerUser int
}

// NewHub returns a hub that admits at most maxPerUser sockets per user.
func NewHub(maxPerUser int) *Hub {
	return &Hub{
		byChannel:  map[string]map[*client]struct{}{},
		perUser:    map[uuid.UUID]int{},
		maxPerUser: maxPerUser,
	}
}

// add registers c, or reports false when its user already holds the maximum.
func (h *Hub) add(c *client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.perUser[c.userID] >= h.maxPerUser {
		return false
	}
	h.perUser[c.userID]++
	for _, channel := range c.channels {
		if h.byChannel[channel] == nil {
			h.byChannel[channel] = map[*client]struct{}{}
		}
		h.byChannel[channel][c] = struct{}{}
	}
	return true
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, channel := range c.channels {
		delete(h.byChannel[channel], c)
		if len(h.byChannel[channel]) == 0 {
			delete(h.byChannel, channel)
		}
	}
	if h.perUser[c.userID]--; h.perUser[c.userID] <= 0 {
		delete(h.perUser, c.userID)
	}
}

// Deliver queues payload for every socket on channel. It never blocks: a socket
// whose queue is full has fallen behind and is dropped; the browser reconnects
// and refetches what it missed.
func (h *Hub) Deliver(channel string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.byChannel[channel] {
		select {
		case c.send <- payload:
		default:
			c.markSlow()
		}
	}
}
