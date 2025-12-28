package notes

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	notes     map[string]*Note
	createErr error
	updateErr error
	deleteErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		notes: make(map[string]*Note),
	}
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Note, error) {
	note, ok := m.notes[id]
	if !ok {
		return nil, nil
	}
	return note, nil
}

func (m *mockRepository) List(ctx context.Context, filter *NoteFilter) ([]Note, error) {
	var result []Note
	for _, note := range m.notes {
		if filter.ChildID != "" && note.ChildID != filter.ChildID {
			continue
		}
		if filter.AuthorID != "" && note.AuthorID != filter.AuthorID {
			continue
		}
		if filter.PinnedOnly && !note.Pinned {
			continue
		}
		if len(filter.Tags) > 0 {
			hasTag := false
			for _, filterTag := range filter.Tags {
				if slices.Contains(note.Tags, filterTag) {
					hasTag = true
				}
			}
			if !hasTag {
				continue
			}
		}
		result = append(result, *note)
	}
	return result, nil
}

func (m *mockRepository) Create(ctx context.Context, note *Note) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.notes[note.ID] = note
	return nil
}

func (m *mockRepository) Update(ctx context.Context, note *Note) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.notes[note.ID] = note
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.notes, id)
	return nil
}

func (m *mockRepository) Search(ctx context.Context, childID, query string) ([]Note, error) {
	var result []Note
	queryLower := strings.ToLower(query)
	for _, note := range m.notes {
		if childID != "" && note.ChildID != childID {
			continue
		}
		// Simple search in title and content
		if strings.Contains(strings.ToLower(note.Title), queryLower) ||
			strings.Contains(strings.ToLower(note.Content), queryLower) {
			result = append(result, *note)
		}
	}
	return result, nil
}

func TestService_Create(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "First Steps",
		Content: "Baby took first steps today!",
		Tags:    []string{"milestone", "walking"},
		Pinned:  true,
	}

	note, err := svc.Create(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if note.ID == "" {
		t.Error("Create() should generate an ID")
	}

	if note.ChildID != req.ChildID {
		t.Errorf("Create() ChildID = %v, want %v", note.ChildID, req.ChildID)
	}

	if note.AuthorID != "user-123" {
		t.Errorf("Create() AuthorID = %v, want user-123", note.AuthorID)
	}

	if note.Title != "First Steps" {
		t.Errorf("Create() Title = %v, want First Steps", note.Title)
	}

	if note.Content != "Baby took first steps today!" {
		t.Errorf("Create() Content = %v, want 'Baby took first steps today!'", note.Content)
	}

	if len(note.Tags) != 2 {
		t.Errorf("Create() Tags count = %v, want 2", len(note.Tags))
	}

	if !note.Pinned {
		t.Error("Create() Pinned should be true")
	}
}

func TestService_Create_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Test note",
	}

	_, err := svc.Create(context.Background(), "user-123", req)
	if err == nil {
		t.Error("Create() should return error when repo fails")
	}
}

func TestService_Get(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Test note",
	}
	created, _ := svc.Create(context.Background(), "user-123", req)

	note, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if note == nil {
		t.Fatal("Get() returned nil for existing note")
	}

	if note.ID != created.ID {
		t.Errorf("Get() ID = %v, want %v", note.ID, created.ID)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	note, err := svc.Get(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if note != nil {
		t.Error("Get() should return nil for non-existent note")
	}
}

func TestService_List(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create multiple notes
	for i := range 3 {
		req := &CreateNoteRequest{
			ChildID: "child-123",
			Content: "Note content " + string(rune('A'+i)),
		}
		svc.Create(context.Background(), "user-123", req)
	}

	// Create one for different child
	req := &CreateNoteRequest{
		ChildID: "child-456",
		Content: "Other child note",
	}
	svc.Create(context.Background(), "user-123", req)

	filter := &NoteFilter{ChildID: "child-123"}
	notes, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 3 {
		t.Errorf("List() returned %d notes, want 3", len(notes))
	}
}

func TestService_List_PinnedOnly(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create pinned note
	pinnedReq := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Pinned note",
		Pinned:  true,
	}
	svc.Create(context.Background(), "user-123", pinnedReq)

	// Create unpinned note
	unpinnedReq := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Unpinned note",
		Pinned:  false,
	}
	svc.Create(context.Background(), "user-123", unpinnedReq)

	filter := &NoteFilter{ChildID: "child-123", PinnedOnly: true}
	notes, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("List() with PinnedOnly returned %d notes, want 1", len(notes))
	}
}

func TestService_List_ByAuthor(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create notes by different authors
	req1 := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Note by user-123",
	}
	svc.Create(context.Background(), "user-123", req1)

	req2 := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Note by user-456",
	}
	svc.Create(context.Background(), "user-456", req2)

	filter := &NoteFilter{ChildID: "child-123", AuthorID: "user-123"}
	notes, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("List() by author returned %d notes, want 1", len(notes))
	}
}

func TestService_List_ByTags(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create note with tags
	taggedReq := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Tagged note",
		Tags:    []string{"milestone", "important"},
	}
	svc.Create(context.Background(), "user-123", taggedReq)

	// Create note without matching tags
	otherReq := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Other note",
		Tags:    []string{"general"},
	}
	svc.Create(context.Background(), "user-123", otherReq)

	filter := &NoteFilter{ChildID: "child-123", Tags: []string{"milestone"}}
	notes, err := svc.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("List() by tags returned %d notes, want 1", len(notes))
	}
}

func TestService_Update(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "Original Title",
		Content: "Original content",
	}
	created, _ := svc.Create(context.Background(), "user-123", req)

	updateReq := &UpdateNoteRequest{
		Title:   "Updated Title",
		Content: "Updated content",
		Tags:    []string{"updated"},
		Pinned:  true,
	}

	updated, err := svc.Update(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Update() Title = %v, want Updated Title", updated.Title)
	}

	if updated.Content != "Updated content" {
		t.Errorf("Update() Content = %v, want Updated content", updated.Content)
	}

	if !updated.Pinned {
		t.Error("Update() Pinned should be true")
	}

	if len(updated.Tags) != 1 || updated.Tags[0] != "updated" {
		t.Errorf("Update() Tags = %v, want [updated]", updated.Tags)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	updateReq := &UpdateNoteRequest{
		Title:   "Updated Title",
		Content: "Updated content",
	}

	_, err := svc.Update(context.Background(), "non-existent", updateReq)
	if err == nil {
		t.Error("Update() should return error for non-existent note")
	}
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Test note",
	}
	created, _ := svc.Create(context.Background(), "user-123", req)

	err := svc.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	note, _ := svc.Get(context.Background(), created.ID)
	if note != nil {
		t.Error("Delete() should remove the note")
	}
}

func TestService_Pin(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Test note",
		Pinned:  false,
	}
	created, _ := svc.Create(context.Background(), "user-123", req)

	// Pin the note
	err := svc.Pin(context.Background(), created.ID, true)
	if err != nil {
		t.Fatalf("Pin(true) error = %v", err)
	}

	note, _ := svc.Get(context.Background(), created.ID)
	if !note.Pinned {
		t.Error("Pin(true) should set Pinned to true")
	}

	// Unpin the note
	err = svc.Pin(context.Background(), created.ID, false)
	if err != nil {
		t.Fatalf("Pin(false) error = %v", err)
	}

	note, _ = svc.Get(context.Background(), created.ID)
	if note.Pinned {
		t.Error("Pin(false) should set Pinned to false")
	}
}

func TestService_Pin_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	err := svc.Pin(context.Background(), "non-existent", true)
	if err == nil {
		t.Error("Pin() should return error for non-existent note")
	}
}

func TestService_Search(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create notes with different content
	note1 := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "First Steps",
		Content: "Baby took first steps today!",
	}
	svc.Create(context.Background(), "user-123", note1)

	note2 := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "Feeding",
		Content: "Started solid foods today",
	}
	svc.Create(context.Background(), "user-123", note2)

	note3 := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "Sleep",
		Content: "Slept through the night",
	}
	svc.Create(context.Background(), "user-123", note3)

	// Search for "steps"
	results, err := svc.Search(context.Background(), "child-123", "steps")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search('steps') returned %d notes, want 1", len(results))
	}
}

func TestService_Search_TitleMatch(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "Important Milestone",
		Content: "Details here",
	}
	svc.Create(context.Background(), "user-123", req)

	// Search should match title
	results, err := svc.Search(context.Background(), "child-123", "milestone")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search('milestone') returned %d notes, want 1", len(results))
	}
}

func TestService_Search_CaseInsensitive(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "UPPERCASE TITLE",
		Content: "lowercase content",
	}
	svc.Create(context.Background(), "user-123", req)

	// Search should be case insensitive
	results, err := svc.Search(context.Background(), "child-123", "uppercase")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search('uppercase') returned %d notes, want 1 (case insensitive)", len(results))
	}
}

func TestService_Search_NoResults(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateNoteRequest{
		ChildID: "child-123",
		Title:   "Test Note",
		Content: "Some content here",
	}
	svc.Create(context.Background(), "user-123", req)

	// Search for non-existent term
	results, err := svc.Search(context.Background(), "child-123", "nonexistent")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Search('nonexistent') returned %d notes, want 0", len(results))
	}
}

func TestService_Search_ChildFilter(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create note for child-123
	req1 := &CreateNoteRequest{
		ChildID: "child-123",
		Content: "Matching content",
	}
	svc.Create(context.Background(), "user-123", req1)

	// Create note for child-456
	req2 := &CreateNoteRequest{
		ChildID: "child-456",
		Content: "Matching content",
	}
	svc.Create(context.Background(), "user-123", req2)

	// Search should filter by child
	results, err := svc.Search(context.Background(), "child-123", "matching")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Search() with child filter returned %d notes, want 1", len(results))
	}
}
