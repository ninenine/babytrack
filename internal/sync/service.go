package sync

import (
	"context"
	"time"
)

type EventType string

const (
	EventTypeFeeding     EventType = "feeding"
	EventTypeSleep       EventType = "sleep"
	EventTypeMedication  EventType = "medication"
	EventTypeNote        EventType = "note"
	EventTypeVaccination EventType = "vaccination"
	EventTypeAppointment EventType = "appointment"
)

type Event struct {
	ID        string          `json:"id"`
	Type      EventType       `json:"type"`
	Action    string          `json:"action"` // create, update, delete
	EntityID  string          `json:"entity_id"`
	Data      interface{}     `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	ClientID  string          `json:"client_id"`
}

type PushRequest struct {
	Events   []Event `json:"events"`
	ClientID string  `json:"client_id"`
}

type PushResponse struct {
	Processed   int      `json:"processed"`
	Failed      int      `json:"failed"`
	Conflicts   []string `json:"conflicts,omitempty"`
	ServerTime  string   `json:"server_time"`
}

type PullResponse struct {
	Events     []Event `json:"events"`
	ServerTime string  `json:"server_time"`
	HasMore    bool    `json:"has_more"`
}

type SyncStatus struct {
	LastSync   string `json:"last_sync"`
	Pending    int    `json:"pending"`
	ServerTime string `json:"server_time"`
}

type Service interface {
	Push(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error)
	Pull(ctx context.Context, userID string, lastSync string) (*PullResponse, error)
	Status(ctx context.Context, userID string) (*SyncStatus, error)
}

type service struct {
	// Add dependencies as needed
}

func NewService() Service {
	return &service{}
}

func (s *service) Push(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
	// TODO: implement
	// 1. Validate events
	// 2. Check for conflicts
	// 3. Apply changes to database
	// 4. Return results
	return &PushResponse{
		Processed:  len(req.Events),
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *service) Pull(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
	// TODO: implement
	// 1. Parse lastSync timestamp
	// 2. Get all events since lastSync for user's families
	// 3. Return events
	return &PullResponse{
		Events:     []Event{},
		ServerTime: time.Now().UTC().Format(time.RFC3339),
		HasMore:    false,
	}, nil
}

func (s *service) Status(ctx context.Context, userID string) (*SyncStatus, error) {
	// TODO: implement
	return &SyncStatus{
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
