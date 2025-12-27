package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ninenine/babytrack/internal/feeding"
	"github.com/ninenine/babytrack/internal/medication"
	"github.com/ninenine/babytrack/internal/notes"
	"github.com/ninenine/babytrack/internal/sleep"
)

type EventType string

const (
	EventTypeFeeding       EventType = "feeding"
	EventTypeSleep         EventType = "sleep"
	EventTypeMedication    EventType = "medication"
	EventTypeMedicationLog EventType = "medication_log"
	EventTypeNote          EventType = "note"
	EventTypeVaccination   EventType = "vaccination"
	EventTypeAppointment   EventType = "appointment"
)

type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Action    string      `json:"action"` // create, update, delete
	EntityID  string      `json:"entity_id"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	ClientID  string      `json:"client_id"`
}

type PushRequest struct {
	Events   []Event `json:"events"`
	ClientID string  `json:"client_id"`
}

type PushResponse struct {
	Processed  int               `json:"processed"`
	Failed     int               `json:"failed"`
	FailedIDs  []string          `json:"failed_ids,omitempty"`
	Results    map[string]string `json:"results,omitempty"` // eventID -> new server ID
	ServerTime string            `json:"server_time"`
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
	feedingService    feeding.Service
	sleepService      sleep.Service
	medicationService medication.Service
	notesService      notes.Service
}

func NewService(
	feedingService feeding.Service,
	sleepService sleep.Service,
	medicationService medication.Service,
	notesService notes.Service,
) Service {
	return &service{
		feedingService:    feedingService,
		sleepService:      sleepService,
		medicationService: medicationService,
		notesService:      notesService,
	}
}

func (s *service) Push(ctx context.Context, userID string, req *PushRequest) (*PushResponse, error) {
	resp := &PushResponse{
		Results:    make(map[string]string),
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}

	for _, event := range req.Events {
		err := s.processEvent(ctx, userID, &event, resp)
		if err != nil {
			resp.Failed++
			resp.FailedIDs = append(resp.FailedIDs, event.ID)
		} else {
			resp.Processed++
		}
	}

	return resp, nil
}

func (s *service) processEvent(ctx context.Context, userID string, event *Event, resp *PushResponse) error {
	switch event.Type {
	case EventTypeFeeding:
		return s.processFeedingEvent(ctx, event, resp)
	case EventTypeSleep:
		return s.processSleepEvent(ctx, event, resp)
	case EventTypeMedication:
		return s.processMedicationEvent(ctx, event, resp)
	case EventTypeMedicationLog:
		return s.processMedicationLogEvent(ctx, userID, event, resp)
	case EventTypeNote:
		return s.processNoteEvent(ctx, userID, event, resp)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (s *service) processFeedingEvent(ctx context.Context, event *Event, resp *PushResponse) error {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	switch event.Action {
	case "create":
		var req feeding.CreateFeedingRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		result, err := s.feedingService.Create(ctx, &req)
		if err != nil {
			return err
		}
		resp.Results[event.ID] = result.ID
		return nil

	case "update":
		var req feeding.CreateFeedingRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		_, err := s.feedingService.Update(ctx, event.EntityID, &req)
		return err

	case "delete":
		return s.feedingService.Delete(ctx, event.EntityID)

	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}

func (s *service) processSleepEvent(ctx context.Context, event *Event, resp *PushResponse) error {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	switch event.Action {
	case "create":
		var req sleep.CreateSleepRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		result, err := s.sleepService.Create(ctx, &req)
		if err != nil {
			return err
		}
		resp.Results[event.ID] = result.ID
		return nil

	case "update":
		var req sleep.CreateSleepRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		_, err := s.sleepService.Update(ctx, event.EntityID, &req)
		return err

	case "delete":
		return s.sleepService.Delete(ctx, event.EntityID)

	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}

func (s *service) processMedicationEvent(ctx context.Context, event *Event, resp *PushResponse) error {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	switch event.Action {
	case "create":
		var req medication.CreateMedicationRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		result, err := s.medicationService.Create(ctx, &req)
		if err != nil {
			return err
		}
		resp.Results[event.ID] = result.ID
		return nil

	case "update":
		var req medication.CreateMedicationRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		_, err := s.medicationService.Update(ctx, event.EntityID, &req)
		return err

	case "delete":
		return s.medicationService.Delete(ctx, event.EntityID)

	case "deactivate":
		return s.medicationService.Deactivate(ctx, event.EntityID)

	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}

func (s *service) processMedicationLogEvent(ctx context.Context, userID string, event *Event, resp *PushResponse) error {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	switch event.Action {
	case "create":
		var req medication.LogMedicationRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		result, err := s.medicationService.LogMedication(ctx, userID, &req)
		if err != nil {
			return err
		}
		resp.Results[event.ID] = result.ID
		return nil

	default:
		return fmt.Errorf("unknown action for medication_log: %s", event.Action)
	}
}

func (s *service) processNoteEvent(ctx context.Context, userID string, event *Event, resp *PushResponse) error {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	switch event.Action {
	case "create":
		var req notes.CreateNoteRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		result, err := s.notesService.Create(ctx, userID, &req)
		if err != nil {
			return err
		}
		resp.Results[event.ID] = result.ID
		return nil

	case "update":
		var req notes.UpdateNoteRequest
		if err := json.Unmarshal(dataBytes, &req); err != nil {
			return err
		}
		_, err := s.notesService.Update(ctx, event.EntityID, &req)
		return err

	case "delete":
		return s.notesService.Delete(ctx, event.EntityID)

	default:
		return fmt.Errorf("unknown action for note: %s", event.Action)
	}
}

func (s *service) Pull(ctx context.Context, userID string, lastSync string) (*PullResponse, error) {
	// For now, return empty - pull sync is more complex and requires
	// tracking server-side changes. The push-first approach handles
	// the primary offline use case.
	return &PullResponse{
		Events:     []Event{},
		ServerTime: time.Now().UTC().Format(time.RFC3339),
		HasMore:    false,
	}, nil
}

func (s *service) Status(ctx context.Context, userID string) (*SyncStatus, error) {
	return &SyncStatus{
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
