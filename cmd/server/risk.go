package main

import (
	"database/sql"
	"log"

	"wheather-go/internal/config"
	"wheather-go/internal/notify"
	"wheather-go/internal/weather"
)

func (a *App) checkWeatherRisk(location config.Location, notifier *notify.Notifier) {
	// Risk state is tracked per location so Mae Sai and CNX do not overwrite each other's alerts.
	w, err := a.store.LatestWeather(location.Name)
	if err == sql.ErrNoRows {
		log.Printf("No weather data yet for %s, skipping risk check", location.Name)
		return
	}
	if err != nil {
		log.Printf("Risk check weather query error for %s: %v", location.Name, err)
		return
	}

	newRisk := weather.Classify(w.RainProbability, w.WindDirection)

	lastRisk, err := a.store.LatestAlertLevel(location.Name)
	if err != nil {
		log.Printf("Risk check alert query error for %s: %v", location.Name, err)
		return
	}

	if newRisk == lastRisk {
		return
	}

	if err := a.store.InsertAlert(newRisk, "Risk level changed to "+newRisk, location.Name); err != nil {
		log.Printf("Failed to insert alert for %s: %v", location.Name, err)
		return
	}

	report, fetchErr := weather.FetchReport(location.Lat, location.Lon)
	switch {
	case newRisk == "HIGH":
		if fetchErr != nil {
			log.Printf("Could not fetch report for urgent alert (%s): %v", location.Name, fetchErr)
		} else {
			notifier.UrgentWeather(report)
		}
	case newRisk == "MEDIUM":
		if fetchErr != nil {
			log.Printf("Could not fetch report for medium alert (%s): %v", location.Name, fetchErr)
		} else {
			notifier.PeriodicReport(report, "MEDIUM")
		}
	case newRisk == "LOW" && weather.RiskRank(lastRisk) > 0:
		notifier.AllClear()
	}

	log.Printf("Risk changed for %s: %s → %s", location.Name, lastRisk, newRisk)
}

func (a *App) checkAQIRisk(location config.Location, notifier *notify.Notifier) {
	// AQI alerts reuse the notifier for the same location so the message lands in the correct destination.
	if location.AQICode == "" || a.cfg.AQIToken == "" {
		return
	}
	aqi, err := weather.FetchAQIReport(location.AQICode, a.cfg.AQIToken)
	if err != nil {
		log.Printf("AQI risk check error for %s: %v", location.Name, err)
		return
	}
	c := aqi.CurrentAQI
	if err := a.store.InsertAQI(c.City, c.AQI, weather.AQICodeText(c.AQI), c.PM25, c.PM10); err != nil {
		log.Printf("Failed to insert AQI data for %s: %v", location.Name, err)
	}
	if aqi.CurrentAQI.AQI >= 151 {
		notifier.UrgentAQI(aqi)
	}
}
