package main

import (
	"log"
	"strings"

	"wheather-go/internal/weather"

	"github.com/bwmarrin/discordgo"
)

func (a *App) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	switch strings.ToLower(m.Content) {
	case "/weather", "/maesai":
		report, err := weather.FetchReport(a.cfg.Maesai.Lat, a.cfg.Maesai.Lon)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Cannot fetch Mae Sai weather report.")
			return
		}

		risk, _ := a.store.LatestAlertLevel(a.cfg.Maesai.Name)

		if a.cfg.Maesai.AQICode != "" && a.cfg.AQIToken != "" {
			aqi, err := weather.FetchAQIReport(a.cfg.Maesai.AQICode, a.cfg.AQIToken)
			if err != nil {
				log.Println("/weather: AQI fetch failed:", err)
			} else {
				a.notifier.AQIReportToChannel(m.ChannelID, aqi)
			}
		}

		a.notifier.PeriodicReportToChannel(m.ChannelID, report, risk)

	case "/cnx":
		if a.cfg.CNX.Lat == "" || a.cfg.CNX.Lon == "" {
			s.ChannelMessageSend(m.ChannelID, "❌ CNX location not configured.")
			return
		}

		report, err := weather.FetchReport(a.cfg.CNX.Lat, a.cfg.CNX.Lon)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Cannot fetch Chiang Mai weather.")
			return
		}

		if err := a.store.InsertWeather(report.Current, report.Next1Hour.PrecipitationProb, a.cfg.CNX.Name); err != nil {
			log.Println("/cnx: insert weather failed:", err)
		}

		if a.cfg.CNX.AQICode != "" && a.cfg.AQIToken != "" {
			aqi, err := weather.FetchAQIReport(a.cfg.CNX.AQICode, a.cfg.AQIToken)
			if err != nil {
				log.Println("/cnx: AQI fetch failed:", err)
			} else {
				c := aqi.CurrentAQI
				if err := a.store.InsertAQI(c.City, c.AQI, weather.AQICodeText(c.AQI), c.PM25, c.PM10); err != nil {
					log.Println("/cnx: insert AQI failed:", err)
				}
				a.cnxNotifier.AQIReportToChannel(m.ChannelID, aqi)
			}
		}

		risk, _ := a.store.LatestAlertLevel(a.cfg.CNX.Name)
		a.cnxNotifier.PeriodicReportToChannel(m.ChannelID, report, risk)
	}
}
