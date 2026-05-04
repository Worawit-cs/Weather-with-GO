# CLAUDE.md

Repository notes for agents and maintainers.

## Commands

```bash
# Run locally
go run ./cmd/server

# Build binary
go build -o server ./cmd/server

# Validate code
go test ./...

# Simulate sensor traffic
./simulate.sh [high|medium|low]

# Deploy on the server machine
./deploy.sh
```

## Runtime Summary

- Two monitored locations: `maesai` and `cnx`
- Both locations run:
  - weather fetch every 10 minutes
  - AQI check during the weather cycle
  - periodic webhook report every 3 hours
  - independent weather-risk transitions and alert history
- Bot commands:
  - `/weather` and `/maesai` for Mae Sai
  - `/cnx` for Chiang Mai
- Bot commands reply only in the channel where the user asked
- Scheduled reports go to the configured webhooks

## Important Code Paths

| Path | Purpose |
|------|---------|
| `cmd/server/main.go` | App wiring, notifier setup, Discord session |
| `cmd/server/cron.go` | Automatic weather and periodic report loops for both locations |
| `cmd/server/risk.go` | Per-location risk and AQI alert checks |
| `cmd/server/bot.go` | Discord bot commands |
| `cmd/server/handlers.go` | HTTP endpoints and test triggers |
| `internal/config/config.go` | `.env` loading and location normalization |
| `internal/store/store.go` | SQLite schema, migrations, location-aware queries |
| `internal/notify/discord.go` | Webhook-only, bot-only, and broadcast send paths |
| `internal/weather/` | API clients, models, and classification logic |

## Config Notes

Expected env vars:

- `PORT`
- `ENV`
- `DB_PATH`
- `WEATHER_BOT_KEY`
- `AQI_TOKEN`
- `MAESAI_CHANNEL`
- `MAESAI_LAT`
- `MAESAI_LON`
- `MAESAI_CODE`
- `DISCORD_WEBHOOK_MAESAI_URL`
- `CNX_CHANNEL`
- `CNX_LAT`
- `CNX_LON`
- `CNX_CODE`
- `DISCORD_WEBHOOK_CNX_URL`
- `DISCORD_WEBHOOK_TEST_URL`

Behavior:

- `ENV=production` uses real location webhooks
- `ENV=debug` or `ENV=development` routes webhook sends to the test webhook
- AQI codes like `5775` are normalized to `@5775`

## API Notes

- `GET /api/alert/latest` defaults to Mae Sai
- `GET /api/alert/latest?location=cnx` returns CNX alert state
- `GET /api/weather/report?location=cnx` fetches CNX weather report
- test endpoints accept `?location=cnx`

## Deployment

`deploy.sh` performs:

1. `git pull --ff-only`
2. `go build -o server ./cmd/server`
3. `sudo systemctl restart weather-server`
