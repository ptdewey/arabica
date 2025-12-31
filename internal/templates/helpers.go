package templates

import "fmt"

func formatTemp(temp float64) string {
	if temp == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.1f°C", temp)
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
