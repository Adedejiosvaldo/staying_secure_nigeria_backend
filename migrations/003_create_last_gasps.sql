-- Create last_gasps table
CREATE TABLE IF NOT EXISTS last_gasps (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    accuracy_m INT NOT NULL,
    cell_info JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expiry_ts TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX idx_last_gasps_user_id ON last_gasps(user_id);
CREATE INDEX idx_last_gasps_expiry ON last_gasps(user_id, expiry_ts) WHERE expiry_ts > NOW();
CREATE INDEX idx_last_gasps_created ON last_gasps(created_at DESC);
