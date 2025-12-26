CREATE TABLE notes (
    id VARCHAR(64) PRIMARY KEY,
    child_id VARCHAR(64) NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    author_id VARCHAR(64) NOT NULL REFERENCES users(id),
    title VARCHAR(255),
    content TEXT NOT NULL,
    tags TEXT[],
    pinned BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    synced_at TIMESTAMPTZ
);

CREATE INDEX idx_notes_child_id ON notes(child_id);
CREATE INDEX idx_notes_author_id ON notes(author_id);
CREATE INDEX idx_notes_pinned ON notes(child_id, pinned);
CREATE INDEX idx_notes_created ON notes(child_id, created_at DESC);
