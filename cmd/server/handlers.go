package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"wheather-go/internal/weather"
)

func (a *App) registerRoutes() {
	http.HandleFunc("/api/sensor", a.sensorHandler)
	http.HandleFunc("/api/alert/latest", a.latestAlertHandler)
	http.HandleFunc("/api/weather/fetch", a.weatherFetchHandler)
	http.HandleFunc("/api/weather/report", a.weatherReportHandler)
	http.HandleFunc("/api/test/high-risk", a.testHighRiskHandler)
	http.HandleFunc("/api/test/peroid", a.testPeroidWeatherHandler)
	http.HandleFunc("/api/test/urgent-aqi", a.testUrgentAQIAlertHandler)
	http.HandleFunc("/health", a.healthHandler)
}

// sensorHandler receives sensor readings from the ESP32 board.
// DB write is disabled until the board is reconnected.
func (a *App) sensorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var s weather.SensorData
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// ESP32 DB write disabled — uncomment when board is connected
	// if err := a.store.InsertSensor(s); err != nil { ... }
	_ = s
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (a *App) latestAlertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Default to Mae Sai to preserve the original endpoint behavior,
	// but allow callers to inspect CNX with ?location=cnx.
	location := r.URL.Query().Get("location")
	if location == "" {
		location = a.cfg.Maesai.Name
	}
	alert, err := a.store.LatestAlert(location)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"risk_level":"LOW","message":"","timestamp":"","location":"` + location + `"}`))
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

func (a *App) weatherFetchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Manual fetch mirrors the cron path so both locations are refreshed the same way.
	a.cronWeatherCycle(a.cfg.Maesai, a.notifier)
	a.cronWeatherCycle(a.cfg.CNX, a.cnxNotifier)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (a *App) weatherReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	location := a.cfg.Maesai
	switch r.URL.Query().Get("location") {
	case a.cfg.CNX.Name:
		// The read-only report endpoint can switch data source without changing payload shape.
		location = a.cfg.CNX
	}

	report, err := weather.FetchReport(location.Lat, location.Lon)
	if err != nil {
		log.Println("weatherReportHandler error:", err)
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (a *App) testHighRiskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mock := weather.CurrentWeather{
		Temperature:      28.0,
		RelativeHumidity: 85,
		Rain:             5.0,
		WindSpeed:        20.0,
		WindDirection:    270.0,
		WeatherCode:      63,
		WeatherCodeText:  "Moderate rain",
	}
	location := a.cfg.Maesai
	notifier := a.notifier
	if r.URL.Query().Get("location") == a.cfg.CNX.Name {
		// Test endpoints follow the same per-location routing rules as the real cron/bot flow.
		location = a.cfg.CNX
		notifier = a.cnxNotifier
	}
	a.store.InsertWeather(mock, 80, location.Name)
	a.checkWeatherRisk(location, notifier)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","note":"Injected HIGH risk data — check Discord"}`))
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (a *App) testPeroidWeatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	location := a.cfg.Maesai
	notifier := a.notifier
	if r.URL.Query().Get("location") == a.cfg.CNX.Name {
		location = a.cfg.CNX
		notifier = a.cnxNotifier
	}
	report, err := weather.FetchReport(location.Lat, location.Lon)
	if err != nil {
		log.Println("testPeroidWeatherHandler error:", err)
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}
	risk, _ := a.store.LatestAlertLevel(location.Name)
	notifier.PeriodicReport(report, risk)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","note":"Triggered periodic report — check Discord"}`))
}

func (a *App) testUrgentAQIAlertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mock := &weather.AQIResponse{}
	mock.CurrentAQI.AQI = 175
	mock.CurrentAQI.PM25 = 72.4
	mock.CurrentAQI.PM10 = 110.0
	mock.CurrentAQI.CodeText = "Unhealthy"
	mock.CurrentAQI.City = "Mae Sai (test)"
	mock.CurrentAQI.Time = time.Now().Format("15:04")
	notifier := a.notifier
	if r.URL.Query().Get("location") == a.cfg.CNX.Name {
		mock.CurrentAQI.City = "Chiang Mai (test)"
		notifier = a.cnxNotifier
	}
	notifier.UrgentAQI(mock)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","note":"Fired urgent AQI alert — check Discord"}`))
}
