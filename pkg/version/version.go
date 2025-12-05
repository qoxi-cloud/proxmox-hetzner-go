// Package version provides build version information.
package version

// Build information set via ldflags.
var (
	// Version is the semantic version (set by goreleaser).
	Version = "dev"

	// Commit is the git commit SHA (set by goreleaser).
	Commit = "none"

	// Date is the build date (set by goreleaser).
	Date = "unknown"
)

// Full returns the full version string.
func Full() string {
	return Version + " (" + Commit + ") built on " + Date
}
