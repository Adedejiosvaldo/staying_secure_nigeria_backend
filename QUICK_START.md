# Quick Start Guide

## ✅ What Was Fixed

1. **Enhanced Error Reporting**: The backend now returns detailed error messages in the API response
2. **Server-Side Logging**: All errors are logged with full context for debugging
3. **Database Migrations**: Created a proper migration system using golang-migrate
4. **Fixed Missing Table**: Created the `blackbox_trails` table that was causing the 500 error

## 🚀 Quick Commands

### Database Migrations
```bash
cd backend

# Source environment variables (or use direnv)
source .envrc

# Run all pending migrations
make migrate-up

# Check current version
make migrate-version

# Rollback last migration
make migrate-down

# Create new migration
make migration name=my_new_migration
```

### Development
```bash
# Build the application
make build

# Run the application
make run

# Run with hot reload (if air is installed)
make dev
```

## 📊 Migration Status

Current migration version: **5**

All tables created:
- ✅ users
- ✅ heartbeats
- ✅ last_gasps
- ✅ alerts
- ✅ blackbox_trails

## 🔍 Debugging API Errors

When the mobile app uploads offline data, the backend now returns detailed errors:

**Before:**
```json
{
  "error": "failed to store trail"
}
```

**After:**
```json
{
  "error": "failed to store trail",
  "details": "ERROR: relation \"blackbox_trails\" does not exist (SQLSTATE 42P01)",
  "trail_id": "cc623cae-c2c5-4887-8f08-cc2ed4939524"
}
```

**Server Logs:**
```
2025/11/20 17:47:54 INFO: Blackbox upload request: UserID=550e8400-..., DataPoints=311
2025/11/20 17:47:54 ERROR: Failed to create blackbox trail: [detailed error]
2025/11/20 17:47:54 Trail details: ID=..., DataPoints=311, StartTs=..., EndTs=...
```

## 🧪 Test the Fix

Now try uploading offline data from the mobile app again. You should either:
1. ✅ Get a successful upload (status 200)
2. ❌ Get a clear error message showing what went wrong

Check the backend logs for detailed error information:
```bash
# If running with air, logs are in the terminal
# If running with docker, use:
docker logs safetrace-api -f
```

## 📁 File Structure

```
backend/
├── cmd/
│   └── migrate/
│       └── migrations/          # Migration files (.up.sql & .down.sql)
├── internal/
│   └── handlers/
│       └── blackbox.go          # Enhanced with error logging
├── Makefile                     # Migration commands
├── .envrc                       # Environment variables (DATABASE_URL, etc.)
├── MIGRATIONS.md                # Detailed migration guide
└── QUICK_START.md               # This file
```

## 🔐 Environment Variables

The `.envrc` file contains:
```bash
DATABASE_URL="postgres://safetrace:safetrace_dev_password@localhost:5432/safetrace?sslmode=disable"
REDIS_URL="localhost:6379"
PORT="8080"
# ... etc
```

**Tip**: Install [direnv](https://direnv.net/) to automatically load these variables:
```bash
# Install
brew install direnv  # macOS
sudo apt install direnv  # Ubuntu

# Setup (add to ~/.bashrc or ~/.zshrc)
eval "$(direnv hook bash)"  # or zsh, fish

# Enable for this directory
cd backend
direnv allow .
```

## 📚 Documentation

- `MIGRATIONS.md` - Complete migration system documentation
- `README.md` - Main project documentation

## 🆘 Troubleshooting

### Database connection issues
```bash
# Check if postgres is running
docker compose ps postgres

# Restart database
docker compose restart postgres
```

### Migration version mismatch
```bash
# Check current version
make migrate-version

# Force to specific version
make migrate-force v=5
```

### Clear database and start fresh
```bash
# ⚠️ This will delete all data!
make migrate-drop
make migrate-up
```

## ✨ Next Steps

1. Try uploading offline data from the mobile app
2. Check the backend logs for any errors
3. If you see errors, they will now be detailed and actionable
4. Use `make migration name=...` to create new migrations as needed

## 🎯 Summary of Changes

| File | Change |
|------|--------|
| `backend/internal/handlers/blackbox.go` | Added detailed error logging and response messages |
| `backend/cmd/migrate/migrations/*` | Created 10 migration files (5 up + 5 down) |
| `backend/Makefile` | Added migration management commands |
| `backend/.envrc` | Created environment configuration file |
| `backend/MIGRATIONS.md` | Comprehensive migration documentation |

**Database Status**: ✅ All tables created, migration version 5

**Error Reporting**: ✅ Enhanced with detailed error messages

**Ready to test!** 🚀

