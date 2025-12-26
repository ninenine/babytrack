CREATE TABLE sleep_records (
    id VARCHAR(64) PRIMARY KEY,
    child_id VARCHAR(64) NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    quality INTEGER CHECK (quality >= 1 AND quality <= 5),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    synced_at TIMESTAMPTZ
);

CREATE INDEX idx_sleep_child_id ON sleep_records(child_id);
CREATE INDEX idx_sleep_start_time ON sleep_records(start_time DESC);
CREATE INDEX idx_sleep_child_start ON sleep_records(child_id, start_time DESC);
