// Package config provides configuration structures and utilities for the
// Proxmox VE installer on Hetzner dedicated servers.
package config

import (
	"strings"
)

// ValidationError aggregates multiple validation errors, enabling users
// to see all configuration issues at once instead of one at a time.
type ValidationError struct {
	// Errors holds all collected validation errors.
	Errors []error
}

// Error returns all validation errors joined by semicolons.
// Returns an empty string if there are no errors.
func (v *ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return ""
	}

	messages := make([]string, len(v.Errors))
	for i, err := range v.Errors {
		messages[i] = err.Error()
	}

	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors.
func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

// Unwrap returns the first error for compatibility with errors.Is() and errors.As().
// Returns nil if there are no errors.
func (v *ValidationError) Unwrap() error {
	if len(v.Errors) == 0 {
		return nil
	}

	return v.Errors[0]
}

// Add appends an error to the validation errors list.
// Nil errors are ignored.
func (v *ValidationError) Add(err error) {
	if err != nil {
		v.Errors = append(v.Errors, err)
	}
}
