package utils

import (
	"fmt"
	"strconv"
	"time"
)

// Helper functions to get max and min values
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func ToFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ToInt converts a string to int.
func ToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func ToTimestamp(ts string) int64 {
    t, _ := time.Parse("2006-01-02 15:04:05", ts)
    return t.Unix()
}

// IsUSMarketOpen checks if the current time is within US stock market regular trading hours, excluding weekends.
func IsUSMarketOpen(currentTime time.Time) bool {
	// Load EST time zone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return false
	}

	// Convert current time to EST
	currentEST := currentTime.In(loc)

	// Check if it's a weekend (Saturday or Sunday)
	if currentEST.Weekday() == time.Saturday || currentEST.Weekday() == time.Sunday {
		return false
	}

	// Define regular market hours: 9:30 AM to 4:00 PM EST
	marketOpen := time.Date(currentEST.Year(), currentEST.Month(), currentEST.Day(), 9, 30, 0, 0, loc)
	marketClose := time.Date(currentEST.Year(), currentEST.Month(), currentEST.Day(), 16, 0, 0, 0, loc)

	// Check if current time is within market hours
	return currentEST.After(marketOpen) && currentEST.Before(marketClose)
}

