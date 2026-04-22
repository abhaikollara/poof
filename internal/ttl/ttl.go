package ttl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var dwPattern = regexp.MustCompile(`(\d+)\s*([dw])`)

// Parse parses a duration string, extending Go's time.ParseDuration with
// support for 'd' (days, 24h) and 'w' (weeks, 168h).
func Parse(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Preprocess: replace Nd and Nw with their hour equivalents.
	converted := dwPattern.ReplaceAllStringFunc(s, func(match string) string {
		parts := dwPattern.FindStringSubmatch(match)
		n, _ := strconv.Atoi(parts[1])
		switch parts[2] {
		case "d":
			return fmt.Sprintf("%dh", n*24)
		case "w":
			return fmt.Sprintf("%dh", n*7*24)
		default:
			return match
		}
	})

	d, err := time.ParseDuration(converted)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be positive, got %q", s)
	}
	return d, nil
}
