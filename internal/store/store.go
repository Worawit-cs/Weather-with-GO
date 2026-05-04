package store

import (
	"database/sql"
	"log"

	"wheather-go/internal/weather"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

// New opens the SQLite DB, enables WAL mode, creates tables, and runs migrations.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS weather_data (
			id                INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp         DATETIME DEFAULT CURRENT_TIMESTAMP,
			temperature       REAL,
			humidity          REAL,
			rain_probability  REAL,
			rainfall          REAL,
			wind_speed        REAL,
			wind_direction    REAL,
			weather_code      INTEGER DEFAULT 0,
			weather_code_text TEXT DEFAULT '',
			location          TEXT DEFAULT 'maesai'
		)`,
		`CREATE TABLE IF NOT EXISTS sensor_data (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp        DATETIME DEFAULT CURRENT_TIMESTAMP,
			sensor_location  TEXT,
			humidity         REAL,
			temperature      REAL,
			water_detected   INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp  DATETIME DEFAULT CURRENT_TIMESTAMP,
			risk_level TEXT,
			message    TEXT,
			location   TEXT DEFAULT 'maesai'
		)`,
		`CREATE TABLE IF NOT EXISTS aqi_data (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			city      TEXT,
			aqi       INTEGER,
			aqi_text  TEXT,
			pm25      REAL,
			pm10      REAL
		)`,
	}

	for _, stmt := range statements {
		if _, err = db.Exec(stmt); err != nil {
			return nil, err
		}
	}

	// Migrations for existing DB — errors ignored when column already exists
	db.Exec(`ALTER TABLE weather_data ADD COLUMN weather_code INTEGER DEFAULT 0`)
	db.Exec(`ALTER TABLE weather_data ADD COLUMN weather_code_text TEXT DEFAULT ''`)
	db.Exec(`ALTER TABLE weather_data ADD COLUMN location TEXT DEFAULT 'maesai'`)
	// alerts.location was added later so old databases can keep working without a destructive migration.
	db.Exec(`ALTER TABLE alerts ADD COLUMN location TEXT DEFAULT 'maesai'`)

	log.Println("Database initialized")
	return &Store{db: db}, nil
}

func (s *Store) InsertWeather(c weather.CurrentWeather, rainProb int, location string) error {
	_, err := s.db.Exec(
		`INSERT INTO weather_data
			(temperature, humidity, rain_probability, rainfall, wind_speed, wind_direction, weather_code, weather_code_text, location)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Temperature, c.RelativeHumidity, rainProb, c.Rain,
		c.WindSpeed, c.WindDirection, c.WeatherCode, c.WeatherCodeText, location,
	)
	return err
}

// LatestWeather returns the most recent weather row for the given location.
func (s *Store) LatestWeather(location string) (weather.WeatherData, error) {
	var w weather.WeatherData
	err := s.db.QueryRow(
		`SELECT temperature, humidity, rain_probability, rainfall, wind_speed, wind_direction
		 FROM weather_data WHERE location = ? ORDER BY id DESC LIMIT 1`,
		location,
	).Scan(&w.Temperature, &w.Humidity, &w.RainProbability, &w.Rainfall, &w.WindSpeed, &w.WindDirection)
	return w, err
}

func (s *Store) InsertAQI(city string, aqi int, text string, pm25, pm10 float64) error {
	_, err := s.db.Exec(
		`INSERT INTO aqi_data (city, aqi, aqi_text, pm25, pm10) VALUES (?, ?, ?, ?, ?)`,
		city, aqi, text, pm25, pm10,
	)
	return err
}

func (s *Store) InsertAlert(level, message, location string) error {
	_, err := s.db.Exec(
		`INSERT INTO alerts (risk_level, message, location) VALUES (?, ?, ?)`,
		level, message, location,
	)
	return err
}

// LatestAlertLevel returns the most recent risk level for a location, defaulting to "LOW" if no alerts exist.
func (s *Store) LatestAlertLevel(location string) (string, error) {
	var level string
	err := s.db.QueryRow(`SELECT risk_level FROM alerts WHERE location = ? ORDER BY id DESC LIMIT 1`, location).Scan(&level)
	if err == sql.ErrNoRows {
		return "LOW", nil
	}
	return level, err
}

// LatestAlert powers both the API and debugging tools, so it returns the location field too.
func (s *Store) LatestAlert(location string) (weather.Alert, error) {
	var a weather.Alert
	err := s.db.QueryRow(
		`SELECT risk_level, message, timestamp, location FROM alerts WHERE location = ? ORDER BY id DESC LIMIT 1`,
		location,
	).Scan(&a.RiskLevel, &a.Message, &a.Timestamp, &a.Location)
	return a, err
}

func (s *Store) InsertSensor(d weather.SensorData) error {
	_, err := s.db.Exec(
		`INSERT INTO sensor_data (sensor_location, humidity, temperature, water_detected) VALUES (?, ?, ?, ?)`,
		d.Location, d.Humidity, d.Temperature, d.WaterDetected,
	)
	return err
}
