package notifications

import (
	"encoding/json"
	"sync"
	"time"
)

// EventType represents the type of notification event
type EventType string

const (
	EventMedicationDue   EventType = "medication_due"
	EventVaccinationDue  EventType = "vaccination_due"
	EventAppointmentSoon EventType = "appointment_soon"
	EventSleepInsight    EventType = "sleep_insight"
)

// Event represents a notification event to be sent to clients
type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	ChildID   string    `json:"childId,omitempty"`
	ChildName string    `json:"childName,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Client represents a connected SSE client
type Client struct {
	UserID string
	Send   chan []byte
}

// Hub manages all SSE client connections and broadcasts events
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	mu         sync.RWMutex
}

// NewHub creates a new notification hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 100),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- data:
				default:
					// Client buffer full, skip
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a new client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(event Event) {
	h.broadcast <- event
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
