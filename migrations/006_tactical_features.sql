-- Trips table for live trip sharing
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    destination VARCHAR(255) NOT NULL,
    estimated_arrival TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Trip locations to stream coordinates
CREATE TABLE IF NOT EXISTS trip_locations (
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_trip_locations_trip_id ON trip_locations(trip_id);

-- Timers for Sentinel Mode
CREATE TABLE IF NOT EXISTS timers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    duration_seconds INT NOT NULL,
    label VARCHAR(255) NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, SAFE, EXPIRED
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_timers_user_status ON timers(user_id, status);
CREATE INDEX IF NOT EXISTS idx_timers_expires_at ON timers(expires_at) WHERE status = 'ACTIVE';

-- Incidents reported by users
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hazard_type VARCHAR(100) NOT NULL,
    description TEXT,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_incidents_location ON incidents(lat, lng);

-- Note: user_settings and trusted_contacts are already JSONB columns in the users table.
-- We do not need schema changes for those, just struct updates in Go.
