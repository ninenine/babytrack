CREATE TABLE medications (
    id VARCHAR(64) PRIMARY KEY,
    child_id VARCHAR(64) NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    dosage VARCHAR(100) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    frequency VARCHAR(100) NOT NULL,
    instructions TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE medication_logs (
    id VARCHAR(64) PRIMARY KEY,
    medication_id VARCHAR(64) NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
    child_id VARCHAR(64) NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    given_at TIMESTAMPTZ NOT NULL,
    given_by VARCHAR(64) NOT NULL REFERENCES users(id),
    dosage VARCHAR(100) NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    synced_at TIMESTAMPTZ
);

CREATE INDEX idx_medications_child_id ON medications(child_id);
CREATE INDEX idx_medications_active ON medications(child_id, active);
CREATE INDEX idx_medication_logs_medication_id ON medication_logs(medication_id);
CREATE INDEX idx_medication_logs_child_id ON medication_logs(child_id);
CREATE INDEX idx_medication_logs_given_at ON medication_logs(given_at DESC);
