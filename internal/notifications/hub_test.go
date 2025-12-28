package notifications

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}

	if hub.clients == nil {
		t.Error("NewHub() clients map should be initialised")
	}

	if hub.register == nil {
		t.Error("NewHub() register channel should be initialised")
	}

	if hub.unregister == nil {
		t.Error("NewHub() unregister channel should be initialised")
	}

	if hub.broadcast == nil {
		t.Error("NewHub() broadcast channel should be initialised")
	}
}

func TestHub_ClientCount_Empty(t *testing.T) {
	hub := NewHub()

	count := hub.ClientCount()
	if count != 0 {
		t.Errorf("ClientCount() = %d, want 0 for empty hub", count)
	}
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Give the hub time to start
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		UserID: "user-123",
		Send:   make(chan []byte, 256),
	}

	// Register client
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	count := hub.ClientCount()
	if count != 1 {
		t.Errorf("ClientCount() after register = %d, want 1", count)
	}

	// Unregister client
	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	count = hub.ClientCount()
	if count != 0 {
		t.Errorf("ClientCount() after unregister = %d, want 0", count)
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	client1 := &Client{
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}
	client2 := &Client{
		UserID: "user-2",
		Send:   make(chan []byte, 256),
	}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	event := Event{
		ID:        "event-123",
		Type:      EventMedicationDue,
		Title:     "Medication Reminder",
		Message:   "Time to give medicine",
		ChildID:   "child-123",
		ChildName: "Baby",
		Timestamp: time.Now(),
	}

	hub.Broadcast(event)
	time.Sleep(10 * time.Millisecond)

	// Check both clients received the event
	select {
	case data := <-client1.Send:
		var received Event
		if err := json.Unmarshal(data, &received); err != nil {
			t.Fatalf("Failed to unmarshal event for client1: %v", err)
		}
		if received.ID != event.ID {
			t.Errorf("Client1 received event ID = %v, want %v", received.ID, event.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 did not receive the broadcast")
	}

	select {
	case data := <-client2.Send:
		var received Event
		if err := json.Unmarshal(data, &received); err != nil {
			t.Fatalf("Failed to unmarshal event for client2: %v", err)
		}
		if received.ID != event.ID {
			t.Errorf("Client2 received event ID = %v, want %v", received.ID, event.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client2 did not receive the broadcast")
	}
}

func TestHub_BroadcastToNoClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	event := Event{
		ID:      "event-123",
		Type:    EventMedicationDue,
		Title:   "Test",
		Message: "Test message",
	}

	// Should not panic when broadcasting to no clients
	hub.Broadcast(event)
	time.Sleep(10 * time.Millisecond)
}

func TestHub_MultipleRegistrations(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	clients := make([]*Client, 5)
	for i := range 5 {
		clients[i] = &Client{
			UserID: "user-" + string(rune('A'+i)),
			Send:   make(chan []byte, 256),
		}
		hub.Register(clients[i])
	}

	time.Sleep(20 * time.Millisecond)

	count := hub.ClientCount()
	if count != 5 {
		t.Errorf("ClientCount() = %d, want 5", count)
	}

	// Unregister some
	hub.Unregister(clients[0])
	hub.Unregister(clients[2])
	time.Sleep(20 * time.Millisecond)

	count = hub.ClientCount()
	if count != 3 {
		t.Errorf("ClientCount() after unregistering 2 = %d, want 3", count)
	}
}

func TestHub_UnregisterNonExistent(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	client := &Client{
		UserID: "user-123",
		Send:   make(chan []byte, 256),
	}

	// Unregistering a client that was never registered should not panic
	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	count := hub.ClientCount()
	if count != 0 {
		t.Errorf("ClientCount() = %d, want 0", count)
	}
}

func TestEvent_Types(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventMedicationDue, "medication_due"},
		{EventVaccinationDue, "vaccination_due"},
		{EventAppointmentSoon, "appointment_soon"},
		{EventSleepInsight, "sleep_insight"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("EventType = %v, want %v", string(tt.eventType), tt.expected)
			}
		})
	}
}

func TestEvent_JSONSerialization(t *testing.T) {
	event := Event{
		ID:        "event-123",
		Type:      EventMedicationDue,
		Title:     "Medication Reminder",
		Message:   "Time to give medicine",
		ChildID:   "child-123",
		ChildName: "Baby",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.ID != event.ID {
		t.Errorf("Decoded ID = %v, want %v", decoded.ID, event.ID)
	}

	if decoded.Type != event.Type {
		t.Errorf("Decoded Type = %v, want %v", decoded.Type, event.Type)
	}

	if decoded.Title != event.Title {
		t.Errorf("Decoded Title = %v, want %v", decoded.Title, event.Title)
	}

	if decoded.Message != event.Message {
		t.Errorf("Decoded Message = %v, want %v", decoded.Message, event.Message)
	}

	if decoded.ChildID != event.ChildID {
		t.Errorf("Decoded ChildID = %v, want %v", decoded.ChildID, event.ChildID)
	}

	if decoded.ChildName != event.ChildName {
		t.Errorf("Decoded ChildName = %v, want %v", decoded.ChildName, event.ChildName)
	}
}

func TestClient_BufferFull(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	// Create a client with a very small buffer
	client := &Client{
		UserID: "user-123",
		Send:   make(chan []byte, 1), // Very small buffer
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Fill the buffer
	event1 := Event{ID: "event-1", Type: EventMedicationDue}
	hub.Broadcast(event1)
	time.Sleep(10 * time.Millisecond)

	// This should not block or panic even with full buffer
	event2 := Event{ID: "event-2", Type: EventMedicationDue}
	event3 := Event{ID: "event-3", Type: EventMedicationDue}
	hub.Broadcast(event2)
	hub.Broadcast(event3)
	time.Sleep(20 * time.Millisecond)

	// Drain the first event
	select {
	case <-client.Send:
		// Got the first event
	default:
		t.Error("Expected at least one event in buffer")
	}
}

func TestHub_ConcurrentAccess(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	done := make(chan bool)

	// Spawn multiple goroutines that register/unregister clients
	for i := range 10 {
		go func(id int) {
			client := &Client{
				UserID: "user-" + string(rune('0'+id)),
				Send:   make(chan []byte, 256),
			}
			hub.Register(client)
			time.Sleep(5 * time.Millisecond)
			hub.Unregister(client)
			done <- true
		}(i)
	}

	// Also broadcast events concurrently
	go func() {
		for i := range 20 {
			event := Event{
				ID:   "event-" + string(rune('0'+i)),
				Type: EventMedicationDue,
			}
			hub.Broadcast(event)
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for range 11 {
		<-done
	}

	// Should complete without deadlock or panic
}
