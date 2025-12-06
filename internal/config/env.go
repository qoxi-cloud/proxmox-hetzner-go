package config

import (
	"os"
	"strings"
)

// parseBool converts common boolean string representations to bool.
// Accepts: "true", "yes", "1" (case-insensitive) as true.
// All other values return false.
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "yes" || s == "1"
}

// EnvVarSet returns true if the environment variable with the given name
// was explicitly set, even if its value is empty.
// This distinguishes between unset variables and variables set to "".
func EnvVarSet(name string) bool {
	_, exists := os.LookupEnv(name)
	return exists
}
