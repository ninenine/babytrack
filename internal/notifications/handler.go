package notifications

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles SSE notification endpoints
type Handler struct {
	hub *Hub
}

// NewHandler creates a new notification handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// RegisterRoutes registers the notification routes
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/stream", h.Stream)
}

// Stream handles the SSE connection
func (h *Handler) Stream(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Create client
	client := &Client{
		UserID: userID.(string),
		Send:   make(chan []byte, 256),
	}

	// Register client
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	log.Printf("[SSE] Client connected: %s", client.UserID)

	// Send initial connection event
	c.Writer.WriteString(fmt.Sprintf("id: %s\n", uuid.New().String()))
	c.Writer.WriteString("event: connected\n")
	c.Writer.WriteString("data: {\"status\":\"connected\"}\n\n")
	c.Writer.Flush()

	// Create channel for client disconnect
	clientGone := c.Request.Context().Done()

	for {
		select {
		case <-clientGone:
			log.Printf("[SSE] Client disconnected: %s", client.UserID)
			return
		case data, ok := <-client.Send:
			if !ok {
				return
			}
			c.Writer.WriteString(fmt.Sprintf("id: %s\n", uuid.New().String()))
			c.Writer.WriteString("event: notification\n")
			c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", data))
			c.Writer.Flush()
		}
	}
}
