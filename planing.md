# Planning Notes

## Current Goal State

The system now treats Mae Sai and CNX as parallel monitored locations.

Each location has:

- its own coordinates and AQI station code from config
- its own webhook target
- its own bot channel ID
- its own weather history in `weather_data.location`
- its own risk history in `alerts.location`

## Automatic Triggers

- Every 10 minutes:
  - fetch weather for Mae Sai
  - fetch weather for CNX
  - store both readings
  - run weather-risk checks for both locations
  - run AQI-risk checks for both locations

- Every 3 hours:
  - send Mae Sai AQI and weather webhook report
  - send CNX AQI and weather webhook report

## Manual Triggers

- `/weather` and `/maesai`
  - reply with Mae Sai AQI + weather through the bot in the requesting channel

- `/cnx`
  - reply with CNX AQI + weather through the bot in the requesting channel

## Data Model

### weather_data

- stores `location`
- keeps readings separated by `maesai` and `cnx`

### alerts

- stores `location`
- keeps risk transitions separated by `maesai` and `cnx`

### aqi_data

- no schema change required for this feature
- city name comes from WAQI response

## API Notes

- `GET /api/alert/latest` defaults to Mae Sai
- `GET /api/alert/latest?location=cnx` returns CNX latest alert
- `GET /api/weather/report?location=cnx` returns CNX report
- test endpoints accept `?location=cnx`

## Operational Rule

- user-requested bot commands answer through the bot only
- scheduled 3-hour reports go through webhooks only
- urgent risk and AQI alerts continue to use the notifier flow for that location
