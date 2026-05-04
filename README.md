# Home Weather & Flood Alert System

A Go backend that monitors weather and AQI for two locations, stores readings in SQLite, and sends Discord updates through both bot commands and scheduled webhooks.

## Current Behavior

- Mae Sai and CNX both run automatic weather checks every 10 minutes.
- Mae Sai and CNX both send automatic webhook reports every 3 hours.
- Mae Sai and CNX both run AQI checks and urgent AQI alerts.
- Mae Sai and CNX both run weather risk classification with independent alert history.
- `/maesai` and `/cnx` reply through the Discord bot in the channel where the user typed the command.
- `/weather` is kept as an alias for Mae Sai.

## Project Structure

```text
weather-GO/
├── cmd/server/              # App wiring, cron, handlers, Discord bot
├── internal/config/         # Env loading and location config
├── internal/store/          # SQLite schema and queries
├── internal/weather/        # Open-Meteo, WAQI, risk logic, models
├── internal/notify/         # Discord payload builders and senders
├── .env.example             # Config template
├── deploy.sh                # Pull, rebuild, restart helper
├── simulate.sh              # Local sensor simulator
├── CLAUDE.md                # Maintainer notes for this repo
├── planing.md               # High-level project notes
└── weather-server.service   # Example systemd service
```

## Prerequisites

- Go 1.21 or newer
- `gcc` / build tools for `github.com/mattn/go-sqlite3`
- `curl` for `simulate.sh`

Ubuntu / Debian:

```bash
sudo apt update
sudo apt install -y golang-go build-essential curl
```

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/YOUR_USERNAME/weather-GO.git
cd weather-GO
```

### 2. Download dependencies

```bash
go mod download
```

### 3. Create `.env`

```bash
cp .env.example .env
```

Example:

```env
PORT=3000
ENV=production
DB_PATH=./data.db
WEATHER_BOT_KEY=your_discord_bot_token

MAESAI_CHANNEL=your_maesai_channel_id
MAESAI_LAT=19.922944
MAESAI_LON=99.867034
MAESAI_CODE=@1832
DISCORD_WEBHOOK_MAESAI_URL=https://discord.com/api/webhooks/...

CNX_CHANNEL=your_cnx_channel_id
CNX_LAT=18.7883
CNX_LON=98.9853
CNX_CODE=@5775
DISCORD_WEBHOOK_CNX_URL=https://discord.com/api/webhooks/...

DISCORD_WEBHOOK_TEST_URL=https://discord.com/api/webhooks/...
AQI_TOKEN=your_waqi_token
```

Notes:

- `ENV=production` uses the real location webhooks.
- `ENV=debug` or `ENV=development` sends webhook traffic to `DISCORD_WEBHOOK_TEST_URL`.
- AQI station codes may be written as `5775` or `@5775`; the app normalizes numeric values automatically.

### 4. Run the server

```bash
go run ./cmd/server
```

Or build a binary:

```bash
go build -o server ./cmd/server
./server
```

## Discord Usage

- `/maesai` or `/weather`: send Mae Sai AQI + weather report back through the bot to the current channel
- `/cnx`: send CNX AQI + weather report back through the bot to the current channel

Scheduled behavior:

- automatic weather cycles every 10 minutes for both locations
- automatic webhook reports every 3 hours for both locations
- urgent risk and AQI alerts for both locations

## API

- `GET /health`
- `GET /api/alert/latest`
- `GET /api/alert/latest?location=cnx`
- `POST /api/weather/fetch`
- `GET /api/weather/report`
- `GET /api/weather/report?location=cnx`
- `POST /api/test/high-risk`
- `POST /api/test/high-risk?location=cnx`
- `POST /api/test/peroid`
- `POST /api/test/peroid?location=cnx`
- `POST /api/test/urgent-aqi`
- `POST /api/test/urgent-aqi?location=cnx`

Quick checks:

```bash
curl http://localhost:3000/health
curl http://localhost:3000/api/alert/latest
curl http://localhost:3000/api/alert/latest?location=cnx
curl http://localhost:3000/api/weather/report
curl http://localhost:3000/api/weather/report?location=cnx
curl -X POST http://localhost:3000/api/weather/fetch
```

## Testing Without ESP32

Start the server:

```bash
go run ./cmd/server
```

Run the simulator in another terminal:

```bash
./simulate.sh
./simulate.sh high
./simulate.sh medium
./simulate.sh low
```

Useful options:

```bash
INTERVAL=10 ./simulate.sh high
SERVER_URL=http://192.168.1.50:3000 ./simulate.sh
```

Manual sensor post:

```bash
curl -X POST http://localhost:3000/api/sensor \
  -H "Content-Type: application/json" \
  -d '{"location":"west","humidity":85,"temperature":28,"water_detected":0}'
```

## Deployment

One-time setup on the server:

```bash
git clone https://github.com/YOUR_USERNAME/weather-GO.git
cd weather-GO

cp .env.example .env
nano .env

go build -o server ./cmd/server
nano weather-server.service

sudo cp weather-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable weather-server
sudo systemctl start weather-server
```

Before enabling the service, update `weather-server.service` paths for your machine.

Deploy updates:

```bash
./deploy.sh
```

## Useful Commands

```bash
go run ./cmd/server
go build -o server ./cmd/server
go test ./...
./simulate.sh high
sudo systemctl status weather-server
journalctl -u weather-server -f
sqlite3 data.db
```

Example SQLite queries:

```sql
SELECT * FROM weather_data WHERE location = 'maesai' ORDER BY id DESC LIMIT 10;
SELECT * FROM weather_data WHERE location = 'cnx' ORDER BY id DESC LIMIT 10;
SELECT * FROM alerts WHERE location = 'maesai' ORDER BY id DESC LIMIT 20;
SELECT * FROM alerts WHERE location = 'cnx' ORDER BY id DESC LIMIT 20;
SELECT * FROM aqi_data ORDER BY id DESC LIMIT 20;
```
