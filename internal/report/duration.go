package report

import "fmt"

// formatDuration renders a duration in seconds as a compact, human-readable
// string (ex: "97ms", "1.2s").
func formatDuration(seconds float64) string {
	if seconds < 1 {
		return fmt.Sprintf("%.0fms", seconds*1000)
	}

	return fmt.Sprintf("%.1fs", seconds)
}