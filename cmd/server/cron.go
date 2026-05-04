package main

import (
	"log"
	"time"

	"wheather-go/internal/config"
	"wheather-go/internal/notify"
	"wheather-go/internal/weather"
)

func (a *App) startCron() {
	// Goroutine 1: fetch weather + check risk every 10 minutes
	go func() {
		for {
			a.cronWeatherCycle(a.cfg.Maesai, a.notifier)
			a.cronWeatherCycle(a.cfg.CNX, a.cnxNotifier)
			time.Sleep(10 * time.Minute)
		}
	}()

	// Goroutine 2: send periodic Discord report every 3 hours
	go func() {
		for {
			a.cronReportCycle(a.cfg.Maesai, a.notifier)
			a.cronReportCycle(a.cfg.CNX, a.cnxNotifier)
			time.Sleep(3 * time.Hour)
		}
	}()

	log.Println("Cron started: weather check every 10min, report every 3hr for Mae Sai and CNX")
}

func (a *App) cronWeatherCycle(location config.Location, notifier *notify.Notifier) {
	if location.Lat == "" || location.Lon == "" {
		log.Printf("Weather cycle skipped: %s location not configured", location.Name)
		return
	}

	report, err := weather.FetchReport(location.Lat, location.Lon)
	if err != nil {
		log.Printf("fetchWeather error for %s: %v", location.Name, err)
		return
	}

	rainProb := report.Next1Hour.PrecipitationProb
	if err := a.store.InsertWeather(report.Current, rainProb, location.Name); err != nil {
		log.Printf("Failed to insert weather data for %s: %v", location.Name, err)
	} else {
		c := report.Current
		log.Printf("Weather fetched for %s: %s  temp=%.1f°C  rain_prob=%d%%  wind=%.0f°",
			location.Name, c.WeatherCodeText, c.Temperature, rainProb, c.WindDirection)
	}

	a.checkWeatherRisk(location, notifier)
	a.checkAQIRisk(location, notifier)
}

func (a *App) cronReportCycle(location config.Location, notifier *notify.Notifier) {
	if location.Lat == "" || location.Lon == "" {
		log.Printf("Periodic report skipped: %s location not configured", location.Name)
		return
	}

	report, err := weather.FetchReport(location.Lat, location.Lon)
	if err != nil {
		log.Printf("Periodic report skipped for %s: could not fetch weather: %v", location.Name, err)
		return
	}

	risk, _ := a.store.LatestAlertLevel(location.Name)

	if location.AQICode != "" && a.cfg.AQIToken != "" {
		aqi, err := weather.FetchAQIReport(location.AQICode, a.cfg.AQIToken)
		if err != nil {
			log.Printf("Periodic AQI report skipped for %s: %v", location.Name, err)
		} else {
			c := aqi.CurrentAQI
			if err := a.store.InsertAQI(c.City, c.AQI, weather.AQICodeText(c.AQI), c.PM25, c.PM10); err != nil {
				log.Printf("Failed to insert AQI data for %s: %v", location.Name, err)
			} else {
				log.Printf("AQI fetched for %s: %s  aqi=%d (%s)  pm25=%.1f  pm10=%.1f",
					location.Name, c.City, c.AQI, weather.AQICodeText(c.AQI), c.PM25, c.PM10)
			}
			notifier.AQIReportWebhookOnly(aqi)
		}
	}

	notifier.PeriodicReportWebhookOnly(report, risk)
}
