-- Create blackbox_trails table
CREATE TABLE IF NOT EXISTS blackbox_trails (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_ts TIMESTAMP NOT NULL,
    end_ts TIMESTAMP NOT NULL,
    data_points INT NOT NULL,
    file_url TEXT NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_blackbox_trails_user_id ON blackbox_trails(user_id);
CREATE INDEX idx_blackbox_trails_uploaded ON blackbox_trails(uploaded_at DESC);
CREATE INDEX idx_blackbox_trails_time_range ON blackbox_trails(user_id, start_ts, end_ts);
