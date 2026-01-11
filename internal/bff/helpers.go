// Package bff provides Backend-For-Frontend functionality including
// template rendering and helper functions for the web UI.
package bff

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"arabica/internal/models"
)

// FormatTemp formats a temperature value with unit detection.
// Returns "N/A" if temp is 0, otherwise determines C/F based on >100 threshold.
func FormatTemp(temp float64) string {
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

// FormatTempValue formats a temperature for use in input fields (numeric value only).
func FormatTempValue(temp float64) string {
	return fmt.Sprintf("%.1f", temp)
}

// FormatTime formats seconds into a human-readable time string (e.g., "3m 30s").
// Returns "N/A" if seconds is 0.
func FormatTime(seconds int) string {
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

// FormatRating formats a rating as "X/10".
// Returns "N/A" if rating is 0.
func FormatRating(rating int) string {
	if rating == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d/10", rating)
}

// FormatID converts an int to string.
func FormatID(id int) string {
	return fmt.Sprintf("%d", id)
}

// FormatInt converts an int to string.
func FormatInt(val int) string {
	return fmt.Sprintf("%d", val)
}

// FormatRoasterID formats a nullable roaster ID.
// Returns "null" if id is nil, otherwise the ID as a string.
func FormatRoasterID(id *int) string {
	if id == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *id)
}

// PoursToJSON serializes a slice of pours to JSON for use in JavaScript.
func PoursToJSON(pours []*models.Pour) string {
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

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// PtrEquals checks if a pointer equals a value.
// Returns false if the pointer is nil.
func PtrEquals[T comparable](p *T, val T) bool {
	if p == nil {
		return false
	}
	return *p == val
}

// PtrValue returns the dereferenced value of a pointer, or zero value if nil.
func PtrValue[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// Iterate returns a slice of ints from 0 to n-1, useful for range loops in templates.
func Iterate(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}
	return result
}

// IterateRemaining returns a slice of ints for the remaining count, useful for star ratings.
// For example, IterateRemaining(3, 5) returns [0, 1] for the 2 remaining empty stars.
func IterateRemaining(current, total int) []int {
	remaining := total - current
	if remaining <= 0 {
		return nil
	}
	result := make([]int, remaining)
	for i := range result {
		result[i] = i
	}
	return result
}

// HasTemp returns true if temperature is greater than zero
func HasTemp(temp float64) bool {
	return temp > 0
}

// HasValue returns true if the int value is greater than zero
func HasValue(val int) bool {
	return val > 0
}

// SafeAvatarURL validates and sanitizes avatar URLs to prevent XSS and other attacks.
// Only allows HTTPS URLs from trusted domains (Bluesky CDN) or relative paths.
// Returns a safe URL or empty string if invalid.
func SafeAvatarURL(avatarURL string) string {
	if avatarURL == "" {
		return ""
	}

	// Allow relative paths (e.g., /static/icon-placeholder.svg)
	if strings.HasPrefix(avatarURL, "/") {
		// Basic validation - must start with /static/
		if strings.HasPrefix(avatarURL, "/static/") {
			return avatarURL
		}
		return ""
	}

	// Parse the URL
	parsedURL, err := url.Parse(avatarURL)
	if err != nil {
		return ""
	}

	// Only allow HTTPS scheme
	if parsedURL.Scheme != "https" {
		return ""
	}

	// Whitelist trusted domains for avatar images
	// Bluesky uses cdn.bsky.app for avatars
	trustedDomains := []string{
		"cdn.bsky.app",
		"av-cdn.bsky.app",
	}

	hostLower := strings.ToLower(parsedURL.Host)
	for _, domain := range trustedDomains {
		if hostLower == domain || strings.HasSuffix(hostLower, "."+domain) {
			return avatarURL
		}
	}

	// URL is not from a trusted domain
	return ""
}

// SafeWebsiteURL validates and sanitizes website URLs for display.
// Only allows HTTP/HTTPS URLs and performs basic validation.
// Returns a safe URL or empty string if invalid.
func SafeWebsiteURL(websiteURL string) string {
	if websiteURL == "" {
		return ""
	}

	// Parse the URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return ""
	}

	// Only allow HTTP and HTTPS schemes
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return ""
	}

	// Basic hostname validation - must have at least one dot
	if !strings.Contains(parsedURL.Host, ".") {
		return ""
	}

	return websiteURL
}
