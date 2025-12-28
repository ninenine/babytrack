package notes

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	return db, mock
}

var noteColumns = []string{
	"id", "child_id", "author_id", "title", "content", "tags", "pinned",
	"created_at", "updated_at", "synced_at",
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestRepository_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-123", "child-456", "author-789", "Test Title", "Test content", pq.Array([]string{"tag1", "tag2"}), true, now, now, syncedAt)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("note-123").
		WillReturnRows(rows)

	note, err := repo.GetByID(context.Background(), "note-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if note == nil {
		t.Fatal("GetByID() returned nil")
	}

	if note.ID != "note-123" {
		t.Errorf("GetByID() ID = %v, want note-123", note.ID)
	}

	if note.ChildID != "child-456" {
		t.Errorf("GetByID() ChildID = %v, want child-456", note.ChildID)
	}

	if note.AuthorID != "author-789" {
		t.Errorf("GetByID() AuthorID = %v, want author-789", note.AuthorID)
	}

	if note.Title != "Test Title" {
		t.Errorf("GetByID() Title = %v, want Test Title", note.Title)
	}

	if note.Content != "Test content" {
		t.Errorf("GetByID() Content = %v, want Test content", note.Content)
	}

	if len(note.Tags) != 2 || note.Tags[0] != "tag1" || note.Tags[1] != "tag2" {
		t.Errorf("GetByID() Tags = %v, want [tag1, tag2]", note.Tags)
	}

	if !note.Pinned {
		t.Error("GetByID() Pinned = false, want true")
	}

	if note.SyncedAt == nil {
		t.Error("GetByID() SyncedAt should not be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	note, err := repo.GetByID(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if note != nil {
		t.Error("GetByID() should return nil for non-existent note")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("note-123").
		WillReturnError(errors.New("database error"))

	_, err := repo.GetByID(context.Background(), "note-123")
	if err == nil {
		t.Error("GetByID() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetByID_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-123", "child-456", "author-789", nil, "Test content", pq.Array([]string{}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("note-123").
		WillReturnRows(rows)

	note, err := repo.GetByID(context.Background(), "note-123")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if note.Title != "" {
		t.Errorf("GetByID() Title should be empty for NULL, got %v", note.Title)
	}

	if note.SyncedAt != nil {
		t.Errorf("GetByID() SyncedAt should be nil for NULL, got %v", note.SyncedAt)
	}

	if len(note.Tags) != 0 {
		t.Errorf("GetByID() Tags should be empty, got %v", note.Tags)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// List Tests
// =============================================================================

func TestRepository_List(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "Title 1", "Content 1", pq.Array([]string{"tag1"}), true, now, now, syncedAt).
		AddRow("note-2", "child-456", "author-2", nil, "Content 2", pq.Array([]string{}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456").
		WillReturnRows(rows)

	filter := &NoteFilter{ChildID: "child-456"}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("List() returned %d notes, want 2", len(notes))
	}

	if notes[0].Title != "Title 1" {
		t.Errorf("List() first note Title = %v, want Title 1", notes[0].Title)
	}

	if notes[1].Title != "" {
		t.Errorf("List() second note Title should be empty, got %v", notes[1].Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_WithAuthorFilter(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-123", "Title 1", "Content 1", pq.Array([]string{}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "author-123").
		WillReturnRows(rows)

	filter := &NoteFilter{ChildID: "child-456", AuthorID: "author-123"}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("List() returned %d notes, want 1", len(notes))
	}

	if notes[0].AuthorID != "author-123" {
		t.Errorf("List() AuthorID = %v, want author-123", notes[0].AuthorID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_PinnedOnly(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "Pinned Note", "Content", pq.Array([]string{}), true, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", true).
		WillReturnRows(rows)

	filter := &NoteFilter{ChildID: "child-456", PinnedOnly: true}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("List() returned %d notes, want 1", len(notes))
	}

	if !notes[0].Pinned {
		t.Error("List() with PinnedOnly should return only pinned notes")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_WithTagsFilter(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "Tagged Note", "Content", pq.Array([]string{"important", "health"}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", pq.Array([]string{"important"})).
		WillReturnRows(rows)

	filter := &NoteFilter{ChildID: "child-456", Tags: []string{"important"}}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("List() returned %d notes, want 1", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_AllFilters(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-123", "Full Filter Note", "Content", pq.Array([]string{"urgent"}), true, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "author-123", true, pq.Array([]string{"urgent"})).
		WillReturnRows(rows)

	filter := &NoteFilter{
		ChildID:    "child-456",
		AuthorID:   "author-123",
		PinnedOnly: true,
		Tags:       []string{"urgent"},
	}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("List() returned %d notes, want 1", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(noteColumns)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WillReturnRows(rows)

	filter := &NoteFilter{}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if notes == nil {
		t.Error("List() should return empty slice, not nil")
	}

	if len(notes) != 0 {
		t.Errorf("List() returned %d notes, want 0", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WillReturnError(errors.New("database error"))

	filter := &NoteFilter{}
	_, err := repo.List(context.Background(), filter)
	if err == nil {
		t.Error("List() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_List_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create rows with wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "child_id"}).
		AddRow("note-1", "child-456")

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WillReturnRows(rows)

	filter := &NoteFilter{}
	_, err := repo.List(context.Background(), filter)
	if err == nil {
		t.Error("List() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// GetPinned Tests (using List with PinnedOnly filter)
// =============================================================================

func TestRepository_GetPinned(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "Pinned 1", "Content 1", pq.Array([]string{"important"}), true, now, now, nil).
		AddRow("note-2", "child-456", "author-2", "Pinned 2", "Content 2", pq.Array([]string{}), true, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", true).
		WillReturnRows(rows)

	filter := &NoteFilter{ChildID: "child-456", PinnedOnly: true}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("GetPinned via List() error = %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("GetPinned returned %d notes, want 2", len(notes))
	}

	for _, note := range notes {
		if !note.Pinned {
			t.Errorf("GetPinned returned non-pinned note: %v", note.ID)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_GetPinned_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(noteColumns)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", true).
		WillReturnRows(rows)

	filter := &NoteFilter{ChildID: "child-456", PinnedOnly: true}
	notes, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("GetPinned via List() error = %v", err)
	}

	if notes == nil {
		t.Error("GetPinned should return empty slice, not nil")
	}

	if len(notes) != 0 {
		t.Errorf("GetPinned returned %d notes, want 0", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Create Tests
// =============================================================================

func TestRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	note := &Note{
		ID:        "new-note",
		ChildID:   "child-123",
		AuthorID:  "author-456",
		Title:     "New Note Title",
		Content:   "New note content",
		Tags:      []string{"tag1", "tag2"},
		Pinned:    true,
		CreatedAt: now,
		UpdatedAt: now,
		SyncedAt:  &syncedAt,
	}

	mock.ExpectExec("INSERT INTO notes").
		WithArgs(note.ID, note.ChildID, note.AuthorID, &note.Title, note.Content,
			pq.Array(note.Tags), note.Pinned, note.CreatedAt, note.UpdatedAt, note.SyncedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), note)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Create_NoOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	note := &Note{
		ID:        "new-note",
		ChildID:   "child-123",
		AuthorID:  "author-456",
		Content:   "Note without title",
		Tags:      []string{},
		Pinned:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO notes").
		WithArgs(note.ID, note.ChildID, note.AuthorID, nil, note.Content,
			pq.Array(note.Tags), note.Pinned, note.CreatedAt, note.UpdatedAt, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), note)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Create_WithTags(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	note := &Note{
		ID:        "tagged-note",
		ChildID:   "child-123",
		AuthorID:  "author-456",
		Title:     "Tagged Note",
		Content:   "Content with tags",
		Tags:      []string{"health", "important", "follow-up"},
		Pinned:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO notes").
		WithArgs(note.ID, note.ChildID, note.AuthorID, &note.Title, note.Content,
			pq.Array(note.Tags), note.Pinned, note.CreatedAt, note.UpdatedAt, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), note)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Create_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	note := &Note{
		ID:        "error-note",
		ChildID:   "child-123",
		AuthorID:  "author-456",
		Content:   "Error note",
		Tags:      []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO notes").
		WithArgs(note.ID, note.ChildID, note.AuthorID, nil, note.Content,
			pq.Array(note.Tags), note.Pinned, note.CreatedAt, note.UpdatedAt, nil).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(context.Background(), note)
	if err == nil {
		t.Error("Create() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Update Tests
// =============================================================================

func TestRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	note := &Note{
		ID:        "update-note",
		Title:     "Updated Title",
		Content:   "Updated content",
		Tags:      []string{"updated", "tag"},
		Pinned:    true,
		UpdatedAt: now,
		SyncedAt:  &syncedAt,
	}

	mock.ExpectExec("UPDATE notes SET title").
		WithArgs(note.ID, &note.Title, note.Content, pq.Array(note.Tags),
			note.Pinned, note.UpdatedAt, note.SyncedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), note)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update_NoTitle(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	note := &Note{
		ID:        "update-note",
		Title:     "", // Empty title should be stored as NULL
		Content:   "Updated content without title",
		Tags:      []string{},
		Pinned:    false,
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE notes SET title").
		WithArgs(note.ID, nil, note.Content, pq.Array(note.Tags),
			note.Pinned, note.UpdatedAt, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), note)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Update_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	note := &Note{
		ID:        "error-update",
		Content:   "Error Update",
		Tags:      []string{},
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE notes SET title").
		WithArgs(note.ID, nil, note.Content, pq.Array(note.Tags),
			note.Pinned, note.UpdatedAt, nil).
		WillReturnError(errors.New("database error"))

	err := repo.Update(context.Background(), note)
	if err == nil {
		t.Error("Update() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Pin/Unpin Tests (using Update)
// =============================================================================

func TestRepository_Pin(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	note := &Note{
		ID:        "pin-note",
		Title:     "Note to Pin",
		Content:   "Content",
		Tags:      []string{},
		Pinned:    true, // Setting pinned to true
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE notes SET title").
		WithArgs(note.ID, &note.Title, note.Content, pq.Array(note.Tags),
			true, note.UpdatedAt, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), note)
	if err != nil {
		t.Fatalf("Pin via Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Unpin(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	note := &Note{
		ID:        "unpin-note",
		Title:     "Note to Unpin",
		Content:   "Content",
		Tags:      []string{},
		Pinned:    false, // Setting pinned to false
		UpdatedAt: now,
	}

	mock.ExpectExec("UPDATE notes SET title").
		WithArgs(note.ID, &note.Title, note.Content, pq.Array(note.Tags),
			false, note.UpdatedAt, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), note)
	if err != nil {
		t.Fatalf("Unpin via Update() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM notes WHERE id").
		WithArgs("delete-note").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "delete-note")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Delete returns no error even if row doesn't exist (0 rows affected)
	mock.ExpectExec("DELETE FROM notes WHERE id").
		WithArgs("non-existent").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Delete_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec("DELETE FROM notes WHERE id").
		WithArgs("error-delete").
		WillReturnError(errors.New("database error"))

	err := repo.Delete(context.Background(), "error-delete")
	if err == nil {
		t.Error("Delete() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// =============================================================================
// Search Tests
// =============================================================================

func TestRepository_Search(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	syncedAt := now.Add(time.Hour)
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "Doctor Visit", "Visited the doctor today", pq.Array([]string{"health"}), true, now, now, syncedAt).
		AddRow("note-2", "child-456", "author-2", nil, "Doctor recommended vitamins", pq.Array([]string{}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%doctor%").
		WillReturnRows(rows)

	notes, err := repo.Search(context.Background(), "child-456", "doctor")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Search() returned %d notes, want 2", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Search_ByTitle(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "Vaccination Record", "Got flu shot", pq.Array([]string{"health"}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%vaccination%").
		WillReturnRows(rows)

	notes, err := repo.Search(context.Background(), "child-456", "vaccination")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Search() returned %d notes, want 1", len(notes))
	}

	if notes[0].Title != "Vaccination Record" {
		t.Errorf("Search() Title = %v, want Vaccination Record", notes[0].Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Search_ByContent(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "General Note", "Remember to buy milk for baby", pq.Array([]string{}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%milk%").
		WillReturnRows(rows)

	notes, err := repo.Search(context.Background(), "child-456", "milk")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Search() returned %d notes, want 1", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Search_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	rows := sqlmock.NewRows(noteColumns)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%nonexistent%").
		WillReturnRows(rows)

	notes, err := repo.Search(context.Background(), "child-456", "nonexistent")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if notes == nil {
		t.Error("Search() should return empty slice, not nil")
	}

	if len(notes) != 0 {
		t.Errorf("Search() returned %d notes, want 0", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Search_Error(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%test%").
		WillReturnError(errors.New("database error"))

	_, err := repo.Search(context.Background(), "child-456", "test")
	if err == nil {
		t.Error("Search() should return error on database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Search_ScanError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	// Create rows with wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "child_id"}).
		AddRow("note-1", "child-456")

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%test%").
		WillReturnRows(rows)

	_, err := repo.Search(context.Background(), "child-456", "test")
	if err == nil {
		t.Error("Search() should return error on scan failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Search_NullOptionalFields(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", nil, "Content with null title", pq.Array([]string{}), false, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%content%").
		WillReturnRows(rows)

	notes, err := repo.Search(context.Background(), "child-456", "content")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Search() returned %d notes, want 1", len(notes))
	}

	if notes[0].Title != "" {
		t.Errorf("Search() Title should be empty for NULL, got %v", notes[0].Title)
	}

	if notes[0].SyncedAt != nil {
		t.Errorf("Search() SyncedAt should be nil for NULL, got %v", notes[0].SyncedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestRepository_Search_WithTags(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()
	repo := NewRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(noteColumns).
		AddRow("note-1", "child-456", "author-1", "Health Note", "Regular checkup notes", pq.Array([]string{"health", "checkup", "routine"}), true, now, now, nil)

	mock.ExpectQuery("SELECT id, child_id, author_id, title, content, tags, pinned").
		WithArgs("child-456", "%checkup%").
		WillReturnRows(rows)

	notes, err := repo.Search(context.Background(), "child-456", "checkup")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Search() returned %d notes, want 1", len(notes))
	}

	if len(notes[0].Tags) != 3 {
		t.Errorf("Search() Tags length = %d, want 3", len(notes[0].Tags))
	}

	expectedTags := []string{"health", "checkup", "routine"}
	for i, tag := range notes[0].Tags {
		if tag != expectedTags[i] {
			t.Errorf("Search() Tag[%d] = %v, want %v", i, tag, expectedTags[i])
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
