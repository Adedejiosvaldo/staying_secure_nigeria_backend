# Database Migrations Guide

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

## Prerequisites

Install golang-migrate CLI:

```bash
# macOS
brew install golang-migrate

# Linux (using binary)
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Arch Linux
yay -S golang-migrate

# Or using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Configuration

The database connection string is stored in `.envrc`:

```bash
export DATABASE_URL="postgres://safetrace:safetrace_pass@localhost:5432/safetrace?sslmode=disable"
```

You can source this file manually:

```bash
source .envrc
```

Or use [direnv](https://direnv.net/) for automatic loading:

```bash
# Install direnv
brew install direnv  # macOS
# or
sudo apt install direnv  # Ubuntu/Debian

# Add to your shell config (.bashrc, .zshrc, etc.)
eval "$(direnv hook bash)"  # or zsh, fish, etc.

# Allow the directory
direnv allow .
```

## Usage

### Run All Pending Migrations

```bash
make migrate-up
```

This will apply all migrations that haven't been run yet.

### Create a New Migration

```bash
make migration name=add_user_email
```

This creates two files:

- `cmd/migrate/migrations/000006_add_user_email.up.sql`
- `cmd/migrate/migrations/000006_add_user_email.down.sql`

Edit these files to add your schema changes.

### Rollback Migrations

Rollback the last migration:

```bash
make migrate-down
```

Rollback the last 3 migrations:

```bash
make migrate-down n=3
```

### Check Migration Version

```bash
make migrate-version
```

### Force Migration Version

If migrations get out of sync (e.g., after manual database changes):

```bash
make migrate-force v=5
```

This sets the migration version to 5 without running any migrations.

### Drop All Tables (⚠️ DANGEROUS!)

```bash
make migrate-drop
```

This will prompt for confirmation before dropping all tables.

## Migration Files

Migration files are located in `cmd/migrate/migrations/` and follow this naming convention:

```
000001_create_users.up.sql
000001_create_users.down.sql
000002_create_heartbeats.up.sql
000002_create_heartbeats.down.sql
...
```

- **`.up.sql`**: Applies the migration (e.g., CREATE TABLE)
- **`.down.sql`**: Reverts the migration (e.g., DROP TABLE)

## Current Migrations

1. **000001_create_users** - Creates users table with phone, name, contacts
2. **000002_create_heartbeats** - Creates heartbeats table for location tracking
3. **000003_create_last_gasps** - Creates last_gasps table for emergency signals
4. **000004_create_alerts** - Creates alerts table for user safety alerts
5. **000005_create_blackbox_trails** - Creates blackbox_trails for offline data

## Best Practices

1. **Always create both up and down migrations** - This allows rollbacks
2. **Test migrations locally first** - Run `migrate-up` and `migrate-down` to test
3. **Use transactions** - Wrap complex migrations in BEGIN/COMMIT
4. **Keep migrations small** - One logical change per migration
5. **Never edit existing migrations** - Create a new migration to fix issues
6. **Use IF NOT EXISTS** - Makes migrations idempotent

## Example Migration

**000006_add_user_email.up.sql:**

```sql
-- Add email column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

**000006_add_user_email.down.sql:**

```sql
-- Remove email column from users table
DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users DROP COLUMN IF EXISTS email;
```

## Troubleshooting

### "Dirty database version"

If you see this error, a migration failed partway through:

```bash
# Fix by forcing to the last known good version
make migrate-force v=4

# Then try running migrations again
make migrate-up
```

### "No change" when running migrate-up

All migrations are already applied. Check version:

```bash
make migrate-version
```

### Connection refused

Make sure PostgreSQL is running:

```bash
docker-compose up -d postgres
```

Check the DATABASE_URL in `.envrc` is correct.

## Quick Start (First Time Setup)

```bash
# 1. Source environment variables
source .envrc

# 2. Start the database
docker-compose up -d postgres

# 3. Run all migrations
make migrate-up

# 4. Verify
make migrate-version
```

You should see:

```
Current migration version:
5
```

## Additional Make Commands

```bash
make build       # Build the application
make run         # Run the application
make dev         # Run with hot reload (air)
make test        # Run tests
make clean       # Clean build artifacts
make help        # Show all available commands
```

