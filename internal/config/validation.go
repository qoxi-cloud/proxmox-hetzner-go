// Package config provides configuration structures and utilities for the
// Proxmox VE installer on Hetzner dedicated servers.
package config

import (
	"errors"
	"regexp"
	"strings"
)

// Hostname validation constants.
const (
	// MaxHostnameLength is the maximum allowed length for a hostname per RFC 1123.
	maxHostnameLength = 63
)

// Hostname validation errors.
var (
	// ErrHostnameEmpty is returned when hostname is empty.
	ErrHostnameEmpty = errors.New("hostname cannot be empty")
	// ErrHostnameTooLong is returned when hostname exceeds 63 characters.
	ErrHostnameTooLong = errors.New("hostname cannot exceed 63 characters")
	// ErrHostnameStartsWithHyphen is returned when hostname starts with a hyphen.
	ErrHostnameStartsWithHyphen = errors.New("hostname cannot start with a hyphen")
	// ErrHostnameEndsWithHyphen is returned when hostname ends with a hyphen.
	ErrHostnameEndsWithHyphen = errors.New("hostname cannot end with a hyphen")
	// ErrHostnameInvalidChars is returned when hostname contains invalid characters.
	ErrHostnameInvalidChars = errors.New("hostname can only contain alphanumeric characters and hyphens")
)

// hostnameRegex matches valid RFC 1123 hostname characters (alphanumeric and hyphens).
var hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// ValidateHostname validates a hostname according to RFC 1123.
// A valid hostname:
//   - Must not be empty
//   - Must not exceed 63 characters
//   - Can only contain alphanumeric characters (a-z, A-Z, 0-9) and hyphens (-)
//   - Cannot start or end with a hyphen
func ValidateHostname(hostname string) error {
	if hostname == "" {
		return ErrHostnameEmpty
	}

	if len(hostname) > maxHostnameLength {
		return ErrHostnameTooLong
	}

	if hostname[0] == '-' {
		return ErrHostnameStartsWithHyphen
	}

	if hostname[len(hostname)-1] == '-' {
		return ErrHostnameEndsWithHyphen
	}

	if !hostnameRegex.MatchString(hostname) {
		return ErrHostnameInvalidChars
	}

	return nil
}

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
