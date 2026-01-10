package templates

import (
	"encoding/json"
	"fmt"

	"arabica/internal/models"
)

func formatTemp(temp float64) string {
	if temp == 0 {
		return "N/A"
	}

	// REFACTOR: This probably isn't the best way to deal with units
	unit := 'C'
	if temp > 100 {
		unit = 'F'
	}

	return fmt.Sprintf("%.1f°%c", temp, unit)
}

func formatTempValue(temp float64) string {
	// For use in input fields - returns just the numeric value
	return fmt.Sprintf("%.1f", temp)
}

func formatTime(seconds int) string {
	if seconds == 0 {
		return "N/A"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	remaining := seconds % 60
	if remaining == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, remaining)
}

func formatRating(rating int) string {
	if rating == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d/10", rating)
}

func formatID(id int) string {
	return fmt.Sprintf("%d", id)
}

func formatInt(val int) string {
	return fmt.Sprintf("%d", val)
}

func formatRoasterID(id *int) string {
	if id == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *id)
}

func poursToJSON(pours []*models.Pour) string {
	if len(pours) == 0 {
		return "[]"
	}

	type pourData struct {
		Water int `json:"water"`
		Time  int `json:"time"`
	}

	data := make([]pourData, len(pours))
	for i, p := range pours {
		data[i] = pourData{
			Water: p.WaterAmount,
			Time:  p.TimeSeconds,
		}
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "[]"
	}

	return string(jsonBytes)
}

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

// ptrEquals checks if a pointer equals a value
func ptrEquals[T comparable](p *T, val T) bool {
	if p == nil {
		return false
	}
	return *p == val
}

// ptrValue returns the dereferenced value of a pointer, or zero value if nil
func ptrValue[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
