-- Drop blackbox_trails table and related objects
DROP INDEX IF EXISTS idx_blackbox_trails_time_range;
DROP INDEX IF EXISTS idx_blackbox_trails_uploaded;
DROP INDEX IF EXISTS idx_blackbox_trails_user_id;
DROP TABLE IF EXISTS blackbox_trails CASCADE;

