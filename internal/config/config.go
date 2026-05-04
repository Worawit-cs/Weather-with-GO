package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Location groups coordinates and identity for a monitored site.
// The Name field is used as the DB location key.
type Location struct {
	Name    string // e.g. "maesai" or "cnx" — stored in weather_data.location
	Lat     string
	Lon     string
	AQICode string // WAQI station code
}

type Config struct {
	Port             string
	DBPath           string
	WeatherBotKey    string
	MaesaiChannel    string
	CNXChannel       string
	WebhookMaesaiURL string
	WebhookTestURL   string
	WebhookCNXURL    string
	AQIToken         string
	Maesai           Location
	CNX              Location
	Debug            bool // true unless ENV=production
}

func Load() (Config, error) {
	_ = godotenv.Load() // silently ignore missing .env; systemd sets vars directly

	envMode := strings.ToLower(envTrim("ENV"))
	cfg := Config{
		Port:             envOr("PORT", "3000"),
		DBPath:           envOr("DB_PATH", "./data.db"),
		WeatherBotKey:    envTrim("WEATHER_BOT_KEY"),
		MaesaiChannel:    envTrim("MAESAI_CHANNEL"),
		CNXChannel:       envTrim("CNX_CHANNEL"),
		WebhookMaesaiURL: envTrim("DISCORD_WEBHOOK_MAESAI_URL"),
		WebhookTestURL:   envTrim("DISCORD_WEBHOOK_TEST_URL"),
		WebhookCNXURL:    envTrim("DISCORD_WEBHOOK_CNX_URL"),
		AQIToken:         envTrim("AQI_TOKEN"),
		Maesai: Location{
			Name:    "maesai",
			Lat:     envTrim("MAESAI_LAT"),
			Lon:     envTrim("MAESAI_LON"),
			AQICode: normalizeAQICode(envTrim("MAESAI_CODE")),
		},
		CNX: Location{
			Name:    "cnx",
			Lat:     envTrim("CNX_LAT"),
			Lon:     envTrim("CNX_LON"),
			AQICode: normalizeAQICode(envTrim("CNX_CODE")),
		},
		Debug: envMode == "debug" || envMode == "development",
	}
	return cfg, nil
}

func envOr(key, def string) string {
	if v := envTrim(key); v != "" {
		return v
	}
	return def
}

func envTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func normalizeAQICode(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "@") {
		return v
	}
	allDigits := true
	for _, r := range v {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return "@" + v
	}
	return v
}
