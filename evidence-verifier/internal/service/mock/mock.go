package mock

import (
	"os"
	"strings"
)

func IsMockMode() bool {
	val := strings.ToLower(os.Getenv("MOCK_MODE"))
	return val == "true" || val == "1"
}
