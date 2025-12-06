package config

import "strings"

// parseBool converts common boolean string representations to bool.
// Accepts: "true", "yes", "1" (case-insensitive) as true.
// All other values return false.
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "yes" || s == "1"
}
