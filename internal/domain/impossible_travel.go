package domain

import (
	"math"
	"strconv"
	"time"
)

type config struct {
	// Maximum realistic travel speed a human can have (km/h).
	// Commercial planes max ~900 km/h; we use 1000 for tolerance.
	MaxSpeedKMH float64

	// Minimum distance where calculation matters (noise reduction).
	MinDistanceKM float64

	// If true, logs or caller uses verbose info (optional).
	Debug bool
}

// DefaultConfig is the default configuration.
var DefaultConfig = config{
	MaxSpeedKMH:   1000,
	MinDistanceKM: 10.0,
	Debug:         false,
}

// ImpossibleTravel determines whether the user movement is physically impossible
func ImpossibleTravel(
	prev GeoInfo,
	prevTime time.Time,
	cur GeoInfo,
	curTime time.Time,
	cfg config,
) (bool, string) {
	// Require valid lat/lon in both samples.
	if prev.Latitude == 0 && prev.Longitude == 0 {
		return false, "prev location missing"
	}
	if cur.Latitude == 0 && cur.Longitude == 0 {
		return false, "current location missing"
	}

	// Calculate distance between two points.
	distanceKM := haversine(
		prev.Latitude,
		prev.Longitude,
		cur.Latitude,
		cur.Longitude,
	)

	if distanceKM < cfg.MinDistanceKM {
		return false, "distance too small"
	}

	// Avoid divide-by-zero.
	duration := curTime.Sub(prevTime)
	if duration <= 0 {
		return false, "invalid timestamps"
	}

	// Convert duration to hours for km/h.
	hours := duration.Hours()
	speed := distanceKM / hours

	if speed > cfg.MaxSpeedKMH {
		return true, impossibleTravelMessage(distanceKM, hours, speed)
	}

	return false, "speed acceptable"
}

func impossibleTravelMessage(distance, hours, speed float64) string {
	return "distance " + formatFloat(distance) +
		" km in " + formatFloat(hours) +
		" h → " + formatFloat(speed) + " km/h (impossible)"
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// haversine calculates great-circle distance between two lat/lon points.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const EarthRadiusKM = 6371.0

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1r := degreesToRadians(lat1)
	lat2r := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1r)*math.Cos(lat2r)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKM * c
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180.0
}
