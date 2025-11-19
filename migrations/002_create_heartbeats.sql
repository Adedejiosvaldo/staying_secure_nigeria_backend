-- Create heartbeats table
CREATE TABLE IF NOT EXISTS heartbeats (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source VARCHAR(10) NOT NULL CHECK (source IN ('http', 'sms')),
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    accuracy_m INT NOT NULL,
    cell_info JSONB NOT NULL,
    battery_pct INT,
    speed DOUBLE PRECISION,
    last_gasp BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP NOT NULL,
    signature TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_heartbeats_user_id ON heartbeats(user_id);
CREATE INDEX idx_heartbeats_timestamp ON heartbeats(timestamp DESC);
CREATE INDEX idx_heartbeats_user_timestamp ON heartbeats(user_id, timestamp DESC);
CREATE INDEX idx_heartbeats_last_gasp ON heartbeats(user_id, last_gasp) WHERE last_gasp = TRUE;
