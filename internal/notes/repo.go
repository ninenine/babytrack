package notes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Note, error)
	List(ctx context.Context, filter *NoteFilter) ([]Note, error)
	Create(ctx context.Context, note *Note) error
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, childID, query string) ([]Note, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Note, error) {
	query := `
		SELECT id, child_id, author_id, title, content, tags, pinned,
		       created_at, updated_at, synced_at
		FROM notes
		WHERE id = $1
	`

	var n Note
	var title sql.NullString
	var tags pq.StringArray
	var syncedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.ChildID, &n.AuthorID, &title, &n.Content, &tags,
		&n.Pinned, &n.CreatedAt, &n.UpdatedAt, &syncedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if title.Valid {
		n.Title = title.String
	}
	n.Tags = tags
	if syncedAt.Valid {
		n.SyncedAt = &syncedAt.Time
	}

	return &n, nil
}

func (r *repository) List(ctx context.Context, filter *NoteFilter) ([]Note, error) {
	query := `
		SELECT id, child_id, author_id, title, content, tags, pinned,
		       created_at, updated_at, synced_at
		FROM notes
		WHERE 1=1
	`
	args := []any{}
	argIndex := 1

	if filter.ChildID != "" {
		query += fmt.Sprintf(` AND child_id = $%d`, argIndex)
		args = append(args, filter.ChildID)
		argIndex++
	}

	if filter.AuthorID != "" {
		query += fmt.Sprintf(` AND author_id = $%d`, argIndex)
		args = append(args, filter.AuthorID)
		argIndex++
	}

	if filter.PinnedOnly {
		query += fmt.Sprintf(` AND pinned = $%d`, argIndex)
		args = append(args, true)
		argIndex++
	}

	if len(filter.Tags) > 0 {
		query += fmt.Sprintf(` AND tags && $%d`, argIndex)
		args = append(args, pq.Array(filter.Tags))
		argIndex++
	}

	query += ` ORDER BY pinned DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		var title sql.NullString
		var tags pq.StringArray
		var syncedAt sql.NullTime

		if err := rows.Scan(
			&n.ID, &n.ChildID, &n.AuthorID, &title, &n.Content, &tags,
			&n.Pinned, &n.CreatedAt, &n.UpdatedAt, &syncedAt,
		); err != nil {
			return nil, err
		}

		if title.Valid {
			n.Title = title.String
		}
		n.Tags = tags
		if syncedAt.Valid {
			n.SyncedAt = &syncedAt.Time
		}

		notes = append(notes, n)
	}

	if notes == nil {
		return []Note{}, nil
	}

	return notes, rows.Err()
}

func (r *repository) Create(ctx context.Context, note *Note) error {
	query := `
		INSERT INTO notes (id, child_id, author_id, title, content, tags, pinned,
		                   created_at, updated_at, synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var title *string
	if note.Title != "" {
		title = &note.Title
	}

	_, err := r.db.ExecContext(ctx, query,
		note.ID, note.ChildID, note.AuthorID, title, note.Content,
		pq.Array(note.Tags), note.Pinned, note.CreatedAt, note.UpdatedAt, note.SyncedAt,
	)

	return err
}

func (r *repository) Update(ctx context.Context, note *Note) error {
	query := `
		UPDATE notes
		SET title = $2, content = $3, tags = $4, pinned = $5, updated_at = $6, synced_at = $7
		WHERE id = $1
	`

	var title *string
	if note.Title != "" {
		title = &note.Title
	}

	_, err := r.db.ExecContext(ctx, query,
		note.ID, title, note.Content, pq.Array(note.Tags),
		note.Pinned, note.UpdatedAt, note.SyncedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) Search(ctx context.Context, childID, query string) ([]Note, error) {
	sqlQuery := `
		SELECT id, child_id, author_id, title, content, tags, pinned,
		       created_at, updated_at, synced_at
		FROM notes
		WHERE child_id = $1
		  AND (title ILIKE $2 OR content ILIKE $2)
		ORDER BY pinned DESC, created_at DESC
		LIMIT 50
	`

	searchPattern := "%" + query + "%"

	rows, err := r.db.QueryContext(ctx, sqlQuery, childID, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		var title sql.NullString
		var tags pq.StringArray
		var syncedAt sql.NullTime

		if err := rows.Scan(
			&n.ID, &n.ChildID, &n.AuthorID, &title, &n.Content, &tags,
			&n.Pinned, &n.CreatedAt, &n.UpdatedAt, &syncedAt,
		); err != nil {
			return nil, err
		}

		if title.Valid {
			n.Title = title.String
		}
		n.Tags = tags
		if syncedAt.Valid {
			n.SyncedAt = &syncedAt.Time
		}

		notes = append(notes, n)
	}

	if notes == nil {
		return []Note{}, nil
	}

	return notes, rows.Err()
}
