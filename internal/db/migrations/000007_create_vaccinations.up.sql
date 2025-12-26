CREATE TABLE vaccinations (
    id VARCHAR(64) PRIMARY KEY,
    child_id VARCHAR(64) NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    dose INTEGER NOT NULL,
    scheduled_at DATE NOT NULL,
    administered_at TIMESTAMPTZ,
    provider VARCHAR(255),
    location VARCHAR(255),
    lot_number VARCHAR(100),
    notes TEXT,
    completed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vaccinations_child_id ON vaccinations(child_id);
CREATE INDEX idx_vaccinations_scheduled ON vaccinations(child_id, scheduled_at);
CREATE INDEX idx_vaccinations_completed ON vaccinations(child_id, completed);
