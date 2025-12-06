// Package config provides configuration structures and utilities for the
// Proxmox VE installer on Hetzner dedicated servers.
package config

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"time"
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

// Email validation errors.
var (
	// ErrEmailEmpty is returned when email is empty.
	ErrEmailEmpty = errors.New("email is required")
	// ErrEmailInvalid is returned when email format is invalid.
	ErrEmailInvalid = errors.New("email format is invalid")
)

// Password validation constants.
const (
	// MinPasswordLength is the minimum allowed length for a password.
	minPasswordLength = 8
)

// Password validation errors.
var (
	// ErrPasswordEmpty is returned when password is empty.
	ErrPasswordEmpty = errors.New("password cannot be empty")
	// ErrPasswordTooShort is returned when password is less than 8 characters.
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)

// SSH key validation errors.
var (
	// ErrSSHKeyEmpty is returned when SSH public key is empty.
	ErrSSHKeyEmpty = errors.New("SSH public key is required")
	// ErrSSHKeyInvalidPrefix is returned when SSH key doesn't start with a valid prefix.
	ErrSSHKeyInvalidPrefix = errors.New("SSH key must start with ssh-rsa, ssh-ed25519, ssh-ecdsa, or ecdsa-sha2-")
)

// Timezone validation errors.
var (
	// ErrTimezoneEmpty is returned when timezone is empty.
	ErrTimezoneEmpty = errors.New("timezone is required")
	// ErrTimezoneInvalid is returned when timezone is not in the IANA timezone database.
	ErrTimezoneInvalid = errors.New("timezone not found in IANA timezone database")
)

// Bridge mode validation errors.
var (
	// ErrBridgeModeEmpty is returned when bridge mode is empty.
	ErrBridgeModeEmpty = errors.New("bridge mode is required")
	// ErrBridgeModeInvalid is returned when bridge mode is not a valid value.
	ErrBridgeModeInvalid = errors.New("bridge mode must be one of: internal, external, both")
)

// ZFS RAID validation errors.
var (
	// ErrZFSRaidEmpty is returned when ZFS RAID level is empty.
	ErrZFSRaidEmpty = errors.New("ZFS RAID level is required")
	// ErrZFSRaidInvalid is returned when ZFS RAID level is not a valid value.
	ErrZFSRaidInvalid = errors.New("ZFS RAID level must be one of: single, raid0, raid1")
)

// Subnet validation errors.
var (
	// ErrSubnetEmpty is returned when subnet is empty.
	ErrSubnetEmpty = errors.New("subnet is required")
	// ErrSubnetInvalid is returned when subnet is not in valid CIDR notation.
	ErrSubnetInvalid = errors.New("subnet must be in valid CIDR notation (e.g., 10.0.0.0/24)")
)

// hostnameRegex matches valid RFC 1123 hostname characters (alphanumeric and hyphens).
var hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// emailRegex provides basic email format validation.
// This is intentionally simple - full RFC 5322 compliance is complex.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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

// ValidateEmail validates an email address format.
// A valid email:
//   - Must not be empty
//   - Must contain @ symbol
//   - Must have valid local and domain parts
//   - Domain must have a TLD of at least 2 characters
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailEmpty
	}

	if !emailRegex.MatchString(email) {
		return ErrEmailInvalid
	}

	return nil
}

// ValidatePassword validates a password for Proxmox root user.
// A valid password:
//   - Must not be empty
//   - Must be at least 8 characters long
//
// Note: No complexity rules (special characters, uppercase, numbers) are required.
// The length is measured in runes to properly handle unicode characters.
func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordEmpty
	}

	if len([]rune(password)) < minPasswordLength {
		return ErrPasswordTooShort
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

// sshKeyPrefixes contains valid SSH public key prefixes.
var sshKeyPrefixes = []string{
	"ssh-rsa ",
	"ssh-ed25519 ",
	"ssh-ecdsa ",
	"ecdsa-sha2-",
}

// ValidateSSHKey validates an SSH public key format.
// A valid SSH public key:
//   - Must not be empty
//   - Must start with one of: ssh-rsa, ssh-ed25519, ssh-ecdsa, or ecdsa-sha2-
func ValidateSSHKey(key string) error {
	if key == "" {
		return ErrSSHKeyEmpty
	}

	for _, prefix := range sshKeyPrefixes {
		if strings.HasPrefix(key, prefix) {
			return nil
		}
	}

	return ErrSSHKeyInvalidPrefix
}

// ValidateTimezone validates a timezone against the IANA timezone database.
// A valid timezone:
//   - Must not be empty
//   - Must exist in the IANA timezone database (e.g., "UTC", "Europe/Kyiv", "America/New_York")
//   - "Local" is a valid special case representing the system's local timezone
func ValidateTimezone(timezone string) error {
	if timezone == "" {
		return ErrTimezoneEmpty
	}

	_, err := time.LoadLocation(timezone)
	if err != nil {
		return ErrTimezoneInvalid
	}

	return nil
}

// ValidateBridgeMode validates a network bridge mode.
// A valid bridge mode:
//   - Must not be empty
//   - Must be one of: internal, external, both
//
// This function uses the BridgeMode.IsValid() method for validation.
func ValidateBridgeMode(mode BridgeMode) error {
	if mode == "" {
		return ErrBridgeModeEmpty
	}

	if !mode.IsValid() {
		return ErrBridgeModeInvalid
	}

	return nil
}

// ValidateZFSRaid validates a ZFS RAID level.
// A valid ZFS RAID level:
//   - Must not be empty
//   - Must be one of: single, raid0, raid1
//
// This function uses the ZFSRaid.IsValid() method for validation.
func ValidateZFSRaid(raid ZFSRaid) error {
	if raid == "" {
		return ErrZFSRaidEmpty
	}

	if !raid.IsValid() {
		return ErrZFSRaidInvalid
	}

	return nil
}

// ValidateSubnet validates a subnet in CIDR notation.
// A valid subnet:
//   - Must not be empty
//   - Must be in valid CIDR notation (e.g., "10.0.0.0/24", "192.168.1.0/24")
//
// This function uses Go's net.ParseCIDR for validation.
func ValidateSubnet(subnet string) error {
	if subnet == "" {
		return ErrSubnetEmpty
	}

	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return ErrSubnetInvalid
	}

	return nil
}
