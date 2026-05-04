package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type waqiRawResponse struct {
	Status string `json:"status"`
	Data   struct {
		AQI  int `json:"aqi"`
		City struct {
			Name string `json:"name"`
		} `json:"city"`
		IAQI struct {
			PM25 struct {
				V float64 `json:"v"`
			} `json:"pm25"`
			PM10 struct {
				V float64 `json:"v"`
			} `json:"pm10"`
		} `json:"iaqi"`
		Time struct {
			S string `json:"s"`
		} `json:"time"`
		Forecast struct {
			Daily struct {
				PM10 []PM10Detail `json:"pm10"`
				PM25 []PM25Detail `json:"pm25"`
			} `json:"daily"`
		} `json:"forecast"`
	} `json:"data"`
}

func BuildAQIURL(city, token string) string {
	return fmt.Sprintf("https://api.waqi.info/feed/%s/?token=%s", city, token)
}

// FetchAQIReport queries the WAQI API and returns a parsed AQIResponse.
// Pure HTTP + parse — no DB, no env reads.
func FetchAQIReport(city, token string) (*AQIResponse, error) {
	resp, err := http.Get(BuildAQIURL(city, token))
	if err != nil {
		return nil, fmt.Errorf("AQI API fetch error: %w", err)
	}
	defer resp.Body.Close()

	var raw waqiRawResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("AQI API parse error: %w", err)
	}
	if raw.Status != "ok" {
		return nil, fmt.Errorf("AQI API returned status: %s", raw.Status)
	}

	result := &AQIResponse{}
	result.CurrentAQI.CodeText = AQICodeText(raw.Data.AQI)
	result.CurrentAQI.AQI = raw.Data.AQI
	result.CurrentAQI.PM25 = raw.Data.IAQI.PM25.V
	result.CurrentAQI.PM10 = raw.Data.IAQI.PM10.V
	result.CurrentAQI.Time = raw.Data.Time.S
	result.CurrentAQI.City = raw.Data.City.Name
	result.DailyAQI.PM10 = raw.Data.Forecast.Daily.PM10
	result.DailyAQI.PM25 = raw.Data.Forecast.Daily.PM25
	return result, nil
}

// AQICodeText maps a numeric AQI value to the US EPA category label.
func AQICodeText(code int) string {
	switch {
	case code <= 50:
		return "Good"
	case code <= 100:
		return "Moderate"
	case code <= 150:
		return "Unhealthy for Sensitive Groups"
	case code <= 200:
		return "Unhealthy"
	case code <= 300:
		return "Very Unhealthy"
	default:
		return "Hazardous"
	}
}
