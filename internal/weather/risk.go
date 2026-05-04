package weather

// IsWestWind returns true for wind coming from the west (225°–315°).
func IsWestWind(degrees float64) bool {
	return degrees >= 225 && degrees <= 315
}

// RiskRank converts a risk level string to an integer for directional comparison.
func RiskRank(level string) int {
	switch level {
	case "HIGH":
		return 2
	case "MEDIUM":
		return 1
	default:
		return 0
	}
}

// Classify returns HIGH/MEDIUM/LOW based on rain probability and wind direction.
func Classify(rainProb float64, windDir float64) string {
	if rainProb > 70 && IsWestWind(windDir) {
		return "HIGH"
	}
	if rainProb > 50 {
		return "MEDIUM"
	}
	return "LOW"
}
