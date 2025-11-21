-- Drop heartbeats table and related objects
DROP INDEX IF EXISTS idx_heartbeats_last_gasp;
DROP INDEX IF EXISTS idx_heartbeats_user_timestamp;
DROP INDEX IF EXISTS idx_heartbeats_timestamp;
DROP INDEX IF EXISTS idx_heartbeats_user_id;
DROP TABLE IF EXISTS heartbeats CASCADE;

