// Package mock provides mock mode detection for local development.
package mock

import (
	"os"
	"strings"
)

// IsMockMode returns true if MOCK_MODE environment variable is set to "true" or "1".
func IsMockMode() bool {
	val := strings.ToLower(os.Getenv("MOCK_MODE"))
	return val == "true" || val == "1"
}
