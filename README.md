# SafeTrace Backend API

Zero-interaction safety monitoring system backend with hybrid connectivity support (HTTP → SMS → Blackbox).

## Features

- ✅ **Heartbeat Management**: HTTP and SMS fallback
- ✅ **Safety Evaluation Engine**: Real-time scoring and state machine
- ✅ **LastGasp Protocol**: Emergency location recording
- ✅ **Alert System**: SMS/WhatsApp via Twilio + FCM push notifications
- ✅ **Blackbox Trail Storage**: Offline data recovery
- ✅ **Rate Limiting**: Protects against spam
- ✅ **HMAC Authentication**: Secure payload verification

## Architecture

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Postgres & Redis clients
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   │   ├── evaluator.go      # Safety scoring engine
│   │   ├── alert_engine.go   # Twilio SMS/WhatsApp + FCM
│   │   └── sms_parser.go     # SMS payload parsing
│   ├── handlers/        # HTTP handlers
│   └── utils/           # Crypto & helpers
├── migrations/          # Database migrations
└── docker-compose.yml   # Docker setup
```

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **SMS**: Twilio
- **Push**: Firebase Cloud Messaging
- **Maps**: Mapbox (optional)

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Twilio account with phone number
- (Optional) Firebase project for FCM

### 1. Clone and Setup

```bash
cd backend
cp .env.example .env
# Edit .env with your credentials
```

### 2. Start with Docker

```bash
docker-compose up -d
```

This will start:
- PostgreSQL on port 5432
- Redis on port 6379
- API server on port 8080

### 3. Run Migrations

Migrations run automatically when Postgres starts via Docker Compose. For manual migration:

```bash
# Connect to Postgres
psql postgresql://safetrace:safetrace_dev_password@localhost:5432/safetrace

# Run each migration file
\i migrations/001_create_users.sql
\i migrations/002_create_heartbeats.sql
\i migrations/003_create_last_gasps.sql
\i migrations/004_create_alerts.sql
\i migrations/005_create_blackbox_trails.sql
```

### 4. Verify

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "safetrace-api",
  "time": "2025-11-19T12:00:00Z"
}
```

## API Endpoints

### Heartbeat

**POST /v1/heartbeat**

Submit location and sensor data.

```bash
curl -X POST http://localhost:8080/v1/heartbeat \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "uuid-here",
    "timestamp": "2025-11-19T12:00:00Z",
    "lat": 6.5244,
    "lng": 3.3792,
    "accuracy_m": 20,
    "cell_info": {
      "mcc": 621,
      "mnc": 20,
      "cid": 12345,
      "lac": 678,
      "rssi": -75,
      "network_type": "4G"
    },
    "battery_pct": 48,
    "speed": 0,
    "last_gasp": false,
    "signature": "hmac-sha256-signature"
  }'
```

### SMS Webhook

**POST /v1/sms/webhook**

Twilio webhook for incoming SMS heartbeats.

Configure in Twilio Console:
```
Webhook URL: https://your-domain.com/v1/sms/webhook
Method: POST
```

### User Status

**GET /v1/user/:id/status**

Get current safety state for a user.

```bash
curl http://localhost:8080/v1/user/{user-id}/status
```

### Blackbox Upload

**POST /v1/blackbox/upload**

Upload offline trail data.

```bash
curl -X POST http://localhost:8080/v1/blackbox/upload \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "uuid",
    "start_ts": "2025-11-19T10:00:00Z",
    "end_ts": "2025-11-19T12:00:00Z",
    "data_points": [...]
  }'
```

### Resolve Alert

**POST /v1/alert/:id/resolve**

Mark an alert as resolved.

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | No | Server port (default: 8080) |
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `REDIS_URL` | Yes | Redis connection string |
| `HMAC_SECRET` | Yes | Secret for HMAC signing (min 32 chars) |
| `JWT_SECRET` | Yes | Secret for JWT tokens (min 32 chars) |
| `TWILIO_ACCOUNT_SID` | Yes | Twilio Account SID |
| `TWILIO_AUTH_TOKEN` | Yes | Twilio Auth Token |
| `TWILIO_PHONE_NUMBER` | Yes | Twilio phone number (E.164 format) |
| `FCM_CREDENTIALS_PATH` | No | Path to Firebase credentials JSON |
| `MAPBOX_TOKEN` | No | Mapbox API token for map links |

### Safety Thresholds

| Variable | Default | Description |
|----------|---------|-------------|
| `HEARTBEAT_INTERVAL_SECONDS` | 180 | Expected heartbeat frequency |
| `HEARTBEAT_WINDOW_SECONDS` | 600 | Grace period before concern |
| `LASTGASP_TIMEOUT_SECONDS` | 3600 | LastGasp validity window |
| `SILENT_PROMPT_SECONDS` | 10 | User response timeout |
| `BLACKBOX_RETENTION_HOURS` | 12 | Local trail retention |

## Safety Evaluation Logic

### State Machine

```
SAFE (80-100) ─────> Normal operation
    │
    ├──> CAUTION (50-79) ─────> Silent ping sent
    │         │
    │         └──> No response ───> AT_RISK
    │
    └──> AT_RISK (<50) ─────> Alert trusted contacts
              │
              └──> ALERT ─────> Escalated emergency
```

### Scoring Components

1. **Heartbeat Recency** (30 pts): Time since last update
2. **GPS Accuracy** (20 pts): Location precision
3. **Movement Pattern** (20 pts): Speed consistency
4. **Signal Quality** (10 pts): Cell signal strength
5. **Source Reliability** (5 pts): HTTP vs SMS
6. **Battery Level** (15 pts): Device power status

### Deterministic Rules

Override scoring for immediate action:

- **Sudden Stop**: Speed drop >40 km/h in <60s
- **Tower Jump**: Location change >5km in <2min
- **No Heartbeat**: Missed window by >10min

## Twilio Setup

### 1. Get Twilio Credentials

1. Sign up at [twilio.com](https://www.twilio.com)
2. Get Account SID and Auth Token from Console
3. Buy a phone number with SMS capability

### 2. Configure SMS Webhook

In Twilio Console → Phone Numbers → Your Number:

```
Messaging Configuration:
  A MESSAGE COMES IN: Webhook
  URL: https://your-domain.com/v1/sms/webhook
  HTTP POST
```

### 3. WhatsApp (Optional)

Enable WhatsApp in Twilio and use `whatsapp:` prefix:

```go
params.SetTo("whatsapp:+2348012345678")
params.SetFrom("whatsapp:" + cfg.TwilioPhoneNumber)
```

## Firebase FCM Setup (Optional)

### 1. Create Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com)
2. Create new project
3. Add Android app

### 2. Download Credentials

1. Project Settings → Service Accounts
2. Generate new private key
3. Save as `firebase-credentials.json`
4. Set path in `.env`:

```
FCM_CREDENTIALS_PATH=/path/to/firebase-credentials.json
```

## Development

### Local Run (without Docker)

```bash
# Start Postgres
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15

# Start Redis
docker run -d -p 6379:6379 redis:7

# Run migrations
psql postgresql://postgres:password@localhost:5432/postgres < migrations/*.sql

# Run server
go run cmd/api/main.go
```

### Build

```bash
go build -o safetrace-api cmd/api/main.go
./safetrace-api
```

### Testing

```bash
go test ./...
```

## Deployment

### Production Checklist

- [ ] Change `HMAC_SECRET` and `JWT_SECRET` to strong random strings
- [ ] Use managed PostgreSQL (AWS RDS, DigitalOcean, etc.)
- [ ] Use managed Redis (AWS ElastiCache, Redis Cloud, etc.)
- [ ] Enable HTTPS with valid SSL certificate
- [ ] Configure Twilio production credentials
- [ ] Set up monitoring and logging
- [ ] Configure backup strategy for database
- [ ] Set up rate limiting and DDoS protection
- [ ] Configure proper CORS for mobile app

### Docker Production

```bash
# Build optimized image
docker build -t safetrace-api:latest .

# Run with production .env
docker run -d \
  --name safetrace-api \
  -p 8080:8080 \
  --env-file .env.production \
  safetrace-api:latest
```

### Cloud Platforms

#### DigitalOcean App Platform

```yaml
name: safetrace-api
services:
  - name: api
    github:
      repo: your-org/safetrace
      branch: main
      deploy_on_push: true
    dockerfile_path: backend/Dockerfile
    http_port: 8080
    envs:
      - key: DATABASE_URL
        value: ${db.DATABASE_URL}
```

#### AWS ECS

1. Push image to ECR
2. Create task definition
3. Configure ECS service with load balancer
4. Set environment variables

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### Logs

```bash
# Docker logs
docker-compose logs -f api

# Or specific service
docker logs -f safetrace-api
```

### Metrics to Monitor

- Heartbeat ingestion rate
- Alert trigger frequency
- SMS delivery success rate
- Database query latency
- Redis hit rate
- API response times

## Troubleshooting

### Database Connection Failed

```bash
# Check Postgres is running
docker ps | grep postgres

# Test connection
psql postgresql://safetrace:password@localhost:5432/safetrace
```

### Twilio SMS Not Sending

1. Verify credentials in `.env`
2. Check Twilio Console → Logs
3. Verify phone number format (E.164: +234...)
4. Check account balance

### Redis Connection Failed

```bash
# Test Redis
redis-cli ping
# Should return: PONG
```

## Security

- All heartbeats must be HMAC-signed
- Secrets must be at least 32 characters
- Use HTTPS in production
- Rate limiting prevents abuse
- User data isolated by UUID

## Support

For issues or questions:
- Check logs: `docker-compose logs api`
- Review migrations
- Verify environment variables
- Check Twilio webhook configuration

## License

Proprietary - All Rights Reserved
