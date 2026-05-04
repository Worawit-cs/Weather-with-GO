package notify

import (
	"fmt"
	"strings"
	"time"

	"wheather-go/internal/weather"
)

const (
	colorRed    = 0xFF0000
	colorYellow = 0xFFA500
	colorGreen  = 0x00CC44
)

type Embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields,omitempty"`
	Footer      *Footer `json:"footer,omitempty"`
	Timestamp   string  `json:"timestamp"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Footer struct {
	Text string `json:"text"`
}

type Payload struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds"`
}

func CodeToColour(code string) int {
	switch strings.ToLower(code) {
	case "good":
		return 5857280
	case "moderate":
		return 16776960
	case "unhealthy for sensitive groups":
		return 16744448
	case "unhealthy":
		return 16711680
	case "very unhealthy":
		return 9381719
	case "hazardous":
		return 8257539
	default:
		return 8421504
	}
}

func AQIReportLabel(aqiVal int, codeText string) (string, int) {
	var title string
	switch {
	case aqiVal > 300:
		title = "🌫️ Air Quality Report — HAZARDOUS"
	case aqiVal > 200:
		title = "🌫️ Air Quality Report — VERY UNHEALTHY"
	case aqiVal > 150:
		title = "🌫️ Air Quality Report — UNHEALTHY"
	case aqiVal > 100:
		title = "🌫️ Air Quality Report — UNHEALTHY FOR SENSITIVE GROUPS"
	case aqiVal > 50:
		title = "🌫️ Air Quality Report — MODERATE"
	default:
		title = "🌫️ Air Quality Report — GOOD"
	}
	return title, CodeToColour(codeText)
}

func HourlyField(label, t string, snap weather.HourlySnapshot) Field {
	return Field{
		Name: label + " (" + ShortTime(t) + ")",
		Value: fmt.Sprintf(
			"🌡️ %.1f°C  💧 %d%%  🌬️ %.1f km/h\n🌧️ Rain: %.1fmm  📊 Prob: %d%%  ☁️ %s",
			snap.Temperature, snap.RelativeHumidity, snap.WindSpeed,
			snap.Rain, snap.PrecipitationProb, snap.WeatherCodeText,
		),
		Inline: false,
	}
}

func ForecastField(label, t string, snap weather.HourlySnapshot) Field {
	return Field{
		Name: label + " (" + ShortTime(t) + ")",
		Value: fmt.Sprintf(
			"🌡️ %.1f°C  📊 Rain prob: %d%%  ☁️ %s",
			snap.Temperature, snap.PrecipitationProb, snap.WeatherCodeText,
		),
		Inline: false,
	}
}

func ShortTime(t string) string {
	if len(t) >= 16 {
		return t[11:16]
	}
	return t
}

func UrgentWeatherPayload(report *weather.WeatherReport) Payload {
	c := report.Current
	embed := Embed{
		Title: "🚨 FLOOD RISK ALERT — HIGH",
		Color: colorRed,
		Fields: []Field{
			{Name: "🌧️ Rain Probability (next 1h)", Value: fmt.Sprintf("%d%%", report.Next1Hour.PrecipitationProb), Inline: true},
			{Name: "🌬️ Wind Direction", Value: fmt.Sprintf("%.0f°", c.WindDirection), Inline: true},
			{Name: "💨 Wind Speed", Value: fmt.Sprintf("%.1f km/h", c.WindSpeed), Inline: true},
			{Name: "🌡️ Temperature", Value: fmt.Sprintf("%.1f°C", c.Temperature), Inline: true},
			{Name: "💧 Humidity", Value: fmt.Sprintf("%d%%", c.RelativeHumidity), Inline: true},
			{Name: "☁️ Condition", Value: c.WeatherCodeText, Inline: true},
			{Name: "🔬 Sensor", Value: "Location: -  |  Humidity: -  |  Temp: -  |  Water: -", Inline: false},
		},
		Footer:    &Footer{Text: "West side of house at risk"},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	return Payload{Content: "@everyone", Embeds: []Embed{embed}}
}

func UrgentAQIPayload(aqi *weather.AQIResponse) Payload {
	c := aqi.CurrentAQI
	embed := Embed{
		Title: fmt.Sprintf("🚨 AQI RISK ALERT — %s 💨\n ⚠️status: %s", c.City, c.CodeText),
		Color: CodeToColour(c.CodeText),
		Fields: []Field{
			{Name: "🕒 TIME", Value: ShortTime(c.Time), Inline: true},
			{Name: "📊 AQI", Value: fmt.Sprintf("%d", c.AQI), Inline: true},
			{Name: "💨 PM2.5", Value: fmt.Sprintf("%.1f μg/m³", c.PM25), Inline: true},
			{Name: "🌫️ PM10", Value: fmt.Sprintf("%.1f μg/m³", c.PM10), Inline: true},
		},
		Footer:    &Footer{Text: "Avoid outdoor activities ⚠️"},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	return Payload{Content: "@everyone", Embeds: []Embed{embed}}
}

func AllClearPayload() Payload {
	embed := Embed{
		Title:       "✅ All Clear — Risk Resolved",
		Description: "Rain risk has dropped back to LOW. West side is safe.",
		Color:       colorGreen,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	return Payload{Embeds: []Embed{embed}}
}

func PeriodicReportPayload(report *weather.WeatherReport, risk string) Payload {
	color := colorGreen
	title := "📊 Weather Report — LOW"
	switch risk {
	case "MEDIUM":
		color = colorYellow
		title = "📊 Weather Report — MEDIUM ⚠️"
	case "HIGH":
		color = colorRed
		title = "📊 Weather Report — HIGH 🚨"
	}

	c := report.Current
	fields := []Field{
		HourlyField("⏪ 1 Hour Ago", report.PastHour.Time, report.PastHour),
		{
			Name: "📍 Now (" + ShortTime(c.Time) + ")",
			Value: fmt.Sprintf(
				"🌡️ %.1f°C  💧 %d%%  🌬️ %.1f km/h (%.0f°)\n🌧️ Rain: %.1fmm  ☁️ %s",
				c.Temperature, c.RelativeHumidity, c.WindSpeed, c.WindDirection,
				c.Rain, c.WeatherCodeText,
			),
			Inline: false,
		},
		ForecastField("🔮 +1h", report.Next1Hour.Time, report.Next1Hour),
		ForecastField("🔮 +2h", report.Next2Hours.Time, report.Next2Hours),
		ForecastField("🔮 +3h", report.Next3Hours.Time, report.Next3Hours),
		{Name: "🔬 Sensor", Value: "Location: -  |  Humidity: -  |  Temp: -  |  Water: -", Inline: false},
	}

	embed := Embed{
		Title:     title,
		Color:     color,
		Fields:    fields,
		Footer:    &Footer{Text: "Next report in 3 hours"},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	return Payload{Embeds: []Embed{embed}}
}

func AQIReportPayload(aqi *weather.AQIResponse) Payload {
	c := aqi.CurrentAQI
	title, color := AQIReportLabel(c.AQI, c.CodeText)

	// Derive "today" from the WAQI response timestamp (station's local timezone),
	// not time.Now(), to avoid a server UTC vs. station UTC+7 mismatch.
	aqiDateStr := c.Time
	if len(aqiDateStr) >= 10 {
		aqiDateStr = aqiDateStr[:10]
	}
	todayT, err := time.Parse("2006-01-02", aqiDateStr)
	if err != nil {
		todayT = time.Now()
	}
	today := todayT.Format("2006-01-02")
	yesterday := todayT.AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := todayT.AddDate(0, 0, 1).Format("2006-01-02")

	findPM10 := func(day string) *weather.PM10Detail {
		for i := range aqi.DailyAQI.PM10 {
			if aqi.DailyAQI.PM10[i].Day == day {
				return &aqi.DailyAQI.PM10[i]
			}
		}
		return nil
	}
	findPM25 := func(day string) *weather.PM25Detail {
		for i := range aqi.DailyAQI.PM25 {
			if aqi.DailyAQI.PM25[i].Day == day {
				return &aqi.DailyAQI.PM25[i]
			}
		}
		return nil
	}
	fmtPM10 := func(d *weather.PM10Detail) string {
		if d == nil {
			return "—"
		}
		return fmt.Sprintf("⬇️ %d   ↔️ %d   ⬆️ %d", d.Min, d.Avg, d.Max)
	}
	fmtPM25 := func(d *weather.PM25Detail) string {
		if d == nil {
			return "—"
		}
		return fmt.Sprintf("⬇️ %d   ↔️ %d   ⬆️ %d", d.Min, d.Avg, d.Max)
	}

	embed := Embed{
		Title: title,
		Color: color,
		Fields: []Field{
			{
				Name:   "📍 City / Time",
				Value:  fmt.Sprintf("%s\n🕐 %s", c.City, ShortTime(c.Time)),
				Inline: false,
			},
			{
				Name: "📊 Now",
				Value: fmt.Sprintf(
					"💨 PM10:   %.1f μg/m³\n🌫️ PM2.5:   %.1f μg/m³\nAQI: %d — %s",
					c.PM10, c.PM25, c.AQI, weather.AQICodeText(c.AQI),
				),
				Inline: false,
			},
			{
				Name: "📅 Daily Forecast",
				Value: fmt.Sprintf(
					"💨 PM10\n  ⏪ Yesterday   %s\n  📍 Today       %s\n  🔮 Tomorrow    %s\n\n🌫️ PM2.5\n  ⏪ Yesterday   %s\n  📍 Today       %s\n  🔮 Tomorrow    %s",
					fmtPM10(findPM10(yesterday)), fmtPM10(findPM10(today)), fmtPM10(findPM10(tomorrow)),
					fmtPM25(findPM25(yesterday)), fmtPM25(findPM25(today)), fmtPM25(findPM25(tomorrow)),
				),
				Inline: false,
			},
		},
		Footer:    &Footer{Text: "Mae Sai AQI — updated every 3 hours"},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	return Payload{Embeds: []Embed{embed}}
}
