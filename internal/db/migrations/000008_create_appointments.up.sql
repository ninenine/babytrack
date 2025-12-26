CREATE TABLE appointments (
    id VARCHAR(64) PRIMARY KEY,
    child_id VARCHAR(64) NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    provider VARCHAR(255),
    location VARCHAR(255),
    scheduled_at TIMESTAMPTZ NOT NULL,
    duration INTEGER NOT NULL DEFAULT 30,
    notes TEXT,
    completed BOOLEAN NOT NULL DEFAULT false,
    cancelled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_appointments_child_id ON appointments(child_id);
CREATE INDEX idx_appointments_scheduled ON appointments(child_id, scheduled_at);
