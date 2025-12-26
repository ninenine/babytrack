CREATE TABLE feedings (
    id VARCHAR(64) PRIMARY KEY,
    child_id VARCHAR(64) NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    amount DECIMAL(10, 2),
    unit VARCHAR(20),
    side VARCHAR(20),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    synced_at TIMESTAMPTZ
);

CREATE INDEX idx_feedings_child_id ON feedings(child_id);
CREATE INDEX idx_feedings_start_time ON feedings(start_time DESC);
CREATE INDEX idx_feedings_child_start ON feedings(child_id, start_time DESC);
