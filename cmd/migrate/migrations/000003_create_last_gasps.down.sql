-- Drop last_gasps table and related objects
DROP INDEX IF EXISTS idx_last_gasps_created;
DROP INDEX IF EXISTS idx_last_gasps_expiry;
DROP INDEX IF EXISTS idx_last_gasps_user_id;
DROP TABLE IF EXISTS last_gasps CASCADE;

