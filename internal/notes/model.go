package notes

import "time"

type Note struct {
	ID        string     `json:"id"`
	ChildID   string     `json:"child_id"`
	AuthorID  string     `json:"author_id"`
	Title     string     `json:"title,omitempty"`
	Content   string     `json:"content"`
	Tags      []string   `json:"tags,omitempty"`
	Pinned    bool       `json:"pinned"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	SyncedAt  *time.Time `json:"synced_at,omitempty"`
}

type CreateNoteRequest struct {
	ChildID string   `json:"child_id" binding:"required"`
	Title   string   `json:"title,omitempty"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags,omitempty"`
	Pinned  bool     `json:"pinned"`
}

type UpdateNoteRequest struct {
	Title   string   `json:"title,omitempty"`
	Content string   `json:"content"`
	Tags    []string `json:"tags,omitempty"`
	Pinned  bool     `json:"pinned"`
}

type NoteFilter struct {
	ChildID    string
	AuthorID   string
	Tags       []string
	PinnedOnly bool
	Search     string
}
