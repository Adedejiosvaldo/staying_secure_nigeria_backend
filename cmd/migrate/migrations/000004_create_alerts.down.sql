-- Drop alerts table and related objects
DROP INDEX IF EXISTS idx_alerts_unresolved;
DROP INDEX IF EXISTS idx_alerts_created;
DROP INDEX IF EXISTS idx_alerts_user_id;
DROP TABLE IF EXISTS alerts CASCADE;

