package api

import (
	"time"
)

// isTimezoneValid checks if the provided timezone is valid.
func isTimezoneValid(tz string) bool {
	if tz == "" {
		return false
	}

	if _, err := time.LoadLocation(tz); err != nil {
		return false
	}

	return true
}
