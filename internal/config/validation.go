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
// Returns an empty string if there are no errors or if the receiver is nil.
func (v *ValidationError) Error() string {
	if v == nil || len(v.Errors) == 0 {
		return ""
	}

	messages := make([]string, 0, len(v.Errors))
	for _, err := range v.Errors {
		if err == nil {
			continue
		}
		messages = append(messages, err.Error())
	}

	if len(messages) == 0 {
		return ""
	}

	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors.
// Returns false if the receiver is nil.
func (v *ValidationError) HasErrors() bool {
	return v != nil && len(v.Errors) > 0
}

// Unwrap returns the first non-nil error for compatibility with errors.Is() and errors.As().
// Returns nil if there are no errors or if the receiver is nil.
func (v *ValidationError) Unwrap() error {
	if v == nil {
		return nil
	}

	for _, err := range v.Errors {
		if err != nil {
			return err
		}
	}

	return nil
}

// Add appends an error to the validation errors list.
// Nil errors are ignored.
func (v *ValidationError) Add(err error) {
	if err != nil {
		v.Errors = append(v.Errors, err)
	}
}
