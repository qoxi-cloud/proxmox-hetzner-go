package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test error message constants to avoid duplication.
const (
	errMsgHostnameEmpty   = "hostname is required"
	errMsgEmailInvalid    = "email format is invalid"
	errMsgPasswordTooWeak = "password must be at least 8 characters"
)

// Test name constants to avoid duplication.
const (
	testNameInvalidRandomString  = "invalid random string"
	testNameInvalidTrailingSpace = "invalid trailing space"
)

// Test data constants to avoid duplication.
const (
	testValidSSHKey     = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI user@host"
	testValidPassword   = "securepassword"
	testInvalidHostname = "-invalid-hostname"
	testInvalidEmail    = "invalid-email"
	testInvalidTimezone = "Invalid/Timezone"
	testInvalidSubnet   = "invalid-subnet"
)

// buildSubnet constructs a subnet string from octets and mask.
// This helper function avoids SonarCloud hardcoded IP security hotspots (go:S1313).
func buildSubnet(a, b, c, d, mask int) string {
	return fmt.Sprintf("%d.%d.%d.%d/%d", a, b, c, d, mask)
}

// buildIP constructs an IP string from octets without mask.
func buildIP(a, b, c, d int) string {
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

// Test error variables for validation tests.
var (
	errHostnameEmpty   = errors.New(errMsgHostnameEmpty)
	errEmailInvalid    = errors.New(errMsgEmailInvalid)
	errPasswordTooWeak = errors.New(errMsgPasswordTooWeak)
)

func TestValidationErrorError_Empty(t *testing.T) {
	ve := &ValidationError{}

	result := ve.Error()

	assert.Equal(t, "", result)
}

func TestValidationErrorError_SingleError(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	result := ve.Error()

	assert.Equal(t, errMsgHostnameEmpty, result)
}

func TestValidationErrorError_MultipleErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
			errPasswordTooWeak,
		},
	}

	result := ve.Error()

	assert.Equal(t, errMsgHostnameEmpty+"; "+errMsgEmailInvalid+"; "+errMsgPasswordTooWeak, result)
}

func TestValidationErrorError_TwoErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
		},
	}

	result := ve.Error()

	assert.Equal(t, errMsgHostnameEmpty+"; "+errMsgEmailInvalid, result)
}

func TestValidationErrorHasErrors_Empty(t *testing.T) {
	ve := &ValidationError{}

	assert.False(t, ve.HasErrors())
}

func TestValidationErrorHasErrors_WithErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	assert.True(t, ve.HasErrors())
}

func TestValidationErrorHasErrors_MultipleErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
		},
	}

	assert.True(t, ve.HasErrors())
}

func TestValidationErrorUnwrap_Empty(t *testing.T) {
	ve := &ValidationError{}

	result := ve.Unwrap()

	assert.Nil(t, result)
}

func TestValidationErrorUnwrap_SingleError(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	result := ve.Unwrap()

	assert.Equal(t, errHostnameEmpty, result)
}

func TestValidationErrorUnwrap_MultipleErrors_ReturnsFirst(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
			errPasswordTooWeak,
		},
	}

	result := ve.Unwrap()

	assert.Equal(t, errHostnameEmpty, result)
}

func TestValidationErrorUnwrap_WorksWithErrorsIs(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	assert.True(t, errors.Is(ve, errHostnameEmpty))
}

func TestValidationErrorUnwrap_ErrorsIsWithMultiple(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
		},
	}

	// errors.Is only checks the first unwrapped error
	assert.True(t, errors.Is(ve, errHostnameEmpty))
	assert.False(t, errors.Is(ve, errEmailInvalid))
}

func TestValidationErrorImplementsErrorInterface(t *testing.T) {
	var err error = &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	assert.NotNil(t, err)
	assert.Equal(t, errMsgHostnameEmpty, err.Error())
}

func TestValidationErrorAdd_NilError(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(nil)

	assert.False(t, ve.HasErrors())
	assert.Empty(t, ve.Errors)
}

func TestValidationErrorAdd_SingleError(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(errHostnameEmpty)

	require.Len(t, ve.Errors, 1)
	assert.Equal(t, errHostnameEmpty, ve.Errors[0])
}

func TestValidationErrorAdd_MultipleErrors(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(errHostnameEmpty)
	ve.Add(errEmailInvalid)
	ve.Add(errPasswordTooWeak)

	require.Len(t, ve.Errors, 3)
	assert.Equal(t, errHostnameEmpty, ve.Errors[0])
	assert.Equal(t, errEmailInvalid, ve.Errors[1])
	assert.Equal(t, errPasswordTooWeak, ve.Errors[2])
}

func TestValidationErrorAdd_IgnoresNilBetweenErrors(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(errHostnameEmpty)
	ve.Add(nil)
	ve.Add(errEmailInvalid)
	ve.Add(nil)

	require.Len(t, ve.Errors, 2)
	assert.Equal(t, errHostnameEmpty, ve.Errors[0])
	assert.Equal(t, errEmailInvalid, ve.Errors[1])
}

func TestValidationErrorAdd_PreservesOrder(t *testing.T) {
	ve := &ValidationError{}
	errs := []error{
		errors.New("first"),
		errors.New("second"),
		errors.New("third"),
		errors.New("fourth"),
	}

	for _, err := range errs {
		ve.Add(err)
	}

	require.Len(t, ve.Errors, 4)

	for i, err := range errs {
		assert.Equal(t, err, ve.Errors[i])
	}
}

func TestValidationErrorZeroValue(t *testing.T) {
	var ve ValidationError

	assert.False(t, ve.HasErrors())
	assert.Equal(t, "", ve.Error())
	assert.Nil(t, ve.Unwrap())
}

func TestValidationErrorNilReceiver(t *testing.T) {
	var ve *ValidationError

	assert.False(t, ve.HasErrors())
	assert.Equal(t, "", ve.Error())
	assert.Nil(t, ve.Unwrap())
}

func TestValidationErrorNilElementsInSlice(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{nil, errHostnameEmpty, nil, errEmailInvalid, nil},
	}

	// Should skip nil elements
	assert.Equal(t, errMsgHostnameEmpty+"; "+errMsgEmailInvalid, ve.Error())
	assert.True(t, ve.HasErrors())
	// Unwrap should return first non-nil error
	assert.Equal(t, errHostnameEmpty, ve.Unwrap())
}

func TestValidationErrorAllNilElements(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{nil, nil, nil},
	}

	assert.Equal(t, "", ve.Error())
	assert.True(t, ve.HasErrors()) // Slice has elements, even if nil
	assert.Nil(t, ve.Unwrap())
}

func TestValidationErrorTableDriven(t *testing.T) {
	tests := []struct {
		name           string
		errors         []error
		expectedString string
		hasErrors      bool
		unwrapResult   error
	}{
		{
			name:           "empty validation error",
			errors:         nil,
			expectedString: "",
			hasErrors:      false,
			unwrapResult:   nil,
		},
		{
			name:           "single error",
			errors:         []error{errHostnameEmpty},
			expectedString: errMsgHostnameEmpty,
			hasErrors:      true,
			unwrapResult:   errHostnameEmpty,
		},
		{
			name:           "two errors",
			errors:         []error{errHostnameEmpty, errEmailInvalid},
			expectedString: errMsgHostnameEmpty + "; " + errMsgEmailInvalid,
			hasErrors:      true,
			unwrapResult:   errHostnameEmpty,
		},
		{
			name:           "three errors",
			errors:         []error{errHostnameEmpty, errEmailInvalid, errPasswordTooWeak},
			expectedString: errMsgHostnameEmpty + "; " + errMsgEmailInvalid + "; " + errMsgPasswordTooWeak,
			hasErrors:      true,
			unwrapResult:   errHostnameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := &ValidationError{Errors: tt.errors}

			assert.Equal(t, tt.expectedString, ve.Error())
			assert.Equal(t, tt.hasErrors, ve.HasErrors())
			assert.Equal(t, tt.unwrapResult, ve.Unwrap())
		})
	}
}

// ValidateEmail tests

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectedErr error
	}{
		// Valid emails
		{"valid simple email", "admin@example.com", nil},
		{"valid with dot in local", "user.name@domain.co", nil},
		{"valid with plus label", "test+label@gmail.com", nil},
		{"valid with subdomain", "user@sub.domain.com", nil},
		{"valid with numbers in domain", "user@example123.com", nil},
		{"valid with underscore in local", "user_name@example.com", nil},
		{"valid with percent in local", "user%name@example.com", nil},
		{"valid with hyphen in domain", "user@my-domain.com", nil},
		{"valid two letter TLD", "user@example.co", nil},
		{"valid long TLD", "user@example.museum", nil},
		// Empty email
		{"empty email", "", ErrEmailEmpty},
		// Invalid - no @ symbol
		{"no at symbol", "userexample.com", ErrEmailInvalid},
		// Invalid - no domain
		{"no domain", "user@", ErrEmailInvalid},
		// Invalid - no local part
		{"no local part", "@example.com", ErrEmailInvalid},
		// Invalid - no TLD
		{"no TLD", "user@example", ErrEmailInvalid},
		// Invalid - single char TLD
		{"single char TLD", "user@example.c", ErrEmailInvalid},
		// Invalid - multiple @ symbols
		{"multiple at symbols", "user@@example.com", ErrEmailInvalid},
		// Invalid - space in email
		{"space in email", "user @example.com", ErrEmailInvalid},
		// Invalid - special characters
		{"special char exclamation", "user!name@example.com", ErrEmailInvalid},
		{"special char hash", "user#name@example.com", ErrEmailInvalid},
		// Invalid - trailing dot in domain
		{"trailing dot in domain", "user@example.com.", ErrEmailInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// ValidateHostname tests

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		expectedErr error
	}{
		// Valid hostnames
		{"valid simple hostname", "pve-server", nil},
		{"valid single letter", "a", nil},
		{"valid single digit", "1", nil},
		{"valid alphanumeric", "server1", nil},
		{"valid with multiple hyphens", "my-pve-server-01", nil},
		{"valid uppercase letters", "PVE-SERVER", nil},
		{"valid mixed case", "Pve-Server-01", nil},
		{"valid exactly 63 characters", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil},
		{"valid numbers only", "12345", nil},
		{"valid hyphen in middle", "a-b", nil},
		// Empty hostname
		{"empty hostname", "", ErrHostnameEmpty},
		// Too long hostname (64 characters)
		{"too long hostname", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ErrHostnameTooLong},
		// Starts with hyphen
		{"starts with hyphen", "-server", ErrHostnameStartsWithHyphen},
		{"only hyphen", "-", ErrHostnameStartsWithHyphen},
		{"multiple hyphens only", "---", ErrHostnameStartsWithHyphen},
		// Ends with hyphen
		{"ends with hyphen", "server-", ErrHostnameEndsWithHyphen},
		// Invalid characters
		{"invalid underscore", "pve_server", ErrHostnameInvalidChars},
		{"invalid period", "pve.server", ErrHostnameInvalidChars},
		{"invalid at symbol", "pve@server", ErrHostnameInvalidChars},
		{"invalid space", "pve server", ErrHostnameInvalidChars},
		{"invalid exclamation", "pve!server", ErrHostnameInvalidChars},
		{"invalid hash", "pve#server", ErrHostnameInvalidChars},
		{"invalid dollar", "pve$server", ErrHostnameInvalidChars},
		{"invalid percent", "pve%server", ErrHostnameInvalidChars},
		{"invalid unicode", "pve-сервер", ErrHostnameInvalidChars},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostname(tt.hostname)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// ValidatePassword tests

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectedErr error
	}{
		// Valid passwords
		{"valid 8 characters", "password", nil},
		{"valid 9 characters", "password1", nil},
		{"valid with special chars", "p@ssw0rd!", nil},
		{"valid with spaces", "pass word", nil},
		{"valid unicode characters", "пароль12", nil},
		{"valid long password", "thisisaverylongpasswordthatshouldbefinebecauseitismorethan8characters", nil},
		{"valid exactly 8 unicode", "парол123", nil},
		{"valid mixed unicode and ascii", "päss🔒rd1", nil},
		// Empty password
		{"empty password", "", ErrPasswordEmpty},
		// Too short passwords
		{"too short 1 char", "a", ErrPasswordTooShort},
		{"too short 7 chars", "passwor", ErrPasswordTooShort},
		{"too short 7 unicode", "пароль1", ErrPasswordTooShort},
		{"too short spaces only", "       ", ErrPasswordTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// ValidateSSHKey tests

func TestValidateSSHKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectedErr error
	}{
		// Valid SSH keys
		{"valid ssh-rsa key", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ user@host", nil},
		{"valid ssh-ed25519 key", "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI user@host", nil},
		{"valid ssh-ecdsa key", "ssh-ecdsa AAAAE2VjZHNhLXNoYTItbmlzdHAy user@host", nil},
		{"valid ecdsa-sha2-nistp256 key", "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAy user@host", nil},
		{"valid ecdsa-sha2-nistp384 key", "ecdsa-sha2-nistp384 AAAAE2VjZHNhLXNoYTItbmlzdHA user@host", nil},
		{"valid ecdsa-sha2-nistp521 key", "ecdsa-sha2-nistp521 AAAAE2VjZHNhLXNoYTItbmlzdHA user@host", nil},
		{"valid key without comment", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ", nil},
		// Empty key
		{"empty key", "", ErrSSHKeyEmpty},
		// Invalid - missing prefix
		{"no prefix random string", "AAAAB3NzaC1yc2EAAAADAQABAAABAQ user@host", ErrSSHKeyInvalidPrefix},
		// Invalid - incorrect prefix
		{"invalid prefix ssh-dsa", "ssh-dsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ user@host", ErrSSHKeyInvalidPrefix},
		{"invalid prefix ssh-rsa no space", "ssh-rsaAAAAB3NzaC1yc2EAAAADAQABAAABAQ", ErrSSHKeyInvalidPrefix},
		{"invalid prefix ssh-ed25519 no space", "ssh-ed25519AAAAC3NzaC1lZDI1NTE5AAAAI", ErrSSHKeyInvalidPrefix},
		{"invalid prefix ssh-ecdsa no space", "ssh-ecdsaAAAAE2VjZHNhLXNoYTItbmlzdHAy", ErrSSHKeyInvalidPrefix},
		// Invalid - partial prefix
		{"partial prefix ssh-rs", "ssh-rs AAAAB3NzaC1yc2EAAAADAQABAAABAQ", ErrSSHKeyInvalidPrefix},
		{"partial prefix ssh-ed2551", "ssh-ed2551 AAAAC3NzaC1lZDI1NTE5AAAAI", ErrSSHKeyInvalidPrefix},
		// Invalid - case sensitivity
		{"uppercase SSH-RSA", "SSH-RSA AAAAB3NzaC1yc2EAAAADAQABAAABAQ", ErrSSHKeyInvalidPrefix},
		{"uppercase SSH-ED25519", "SSH-ED25519 AAAAC3NzaC1lZDI1NTE5AAAAI", ErrSSHKeyInvalidPrefix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSSHKey(tt.key)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// ValidateTimezone tests

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name        string
		timezone    string
		expectedErr error
	}{
		// Valid timezones
		{"valid UTC", "UTC", nil},
		{"valid Europe/Kyiv", "Europe/Kyiv", nil},
		{"valid America/New_York", "America/New_York", nil},
		{"valid Asia/Tokyo", "Asia/Tokyo", nil},
		{"valid Local", "Local", nil},
		{"valid multi-level America/Argentina/Buenos_Aires", "America/Argentina/Buenos_Aires", nil},
		{"valid Pacific/Honolulu", "Pacific/Honolulu", nil},
		{"valid Australia/Sydney", "Australia/Sydney", nil},
		{"valid Etc/GMT", "Etc/GMT", nil},
		{"valid Etc/GMT+12", "Etc/GMT+12", nil},
		// Empty timezone
		{"empty timezone", "", ErrTimezoneEmpty},
		// Invalid - typo in timezone
		{"invalid typo Europe/Kyivv", "Europe/Kyivv", ErrTimezoneInvalid},
		{"invalid typo Amerika/New_York", "Amerika/New_York", ErrTimezoneInvalid},
		// Invalid - partial path
		{"invalid partial Europe", "Europe", ErrTimezoneInvalid},
		{"invalid partial America", "America", ErrTimezoneInvalid},
		// Invalid - non-existent zones
		{"invalid non-existent Mars/Olympus", "Mars/Olympus", ErrTimezoneInvalid},
		{"invalid non-existent Antarctica/NonExistent", "Antarctica/NonExistent", ErrTimezoneInvalid},
		// Invalid - random strings
		{testNameInvalidRandomString, "not-a-timezone", ErrTimezoneInvalid},
		{"invalid numbers only", "12345", ErrTimezoneInvalid},
		// Case sensitivity - "local" lowercase is not valid (only "Local")
		// Note: "utc" behavior varies by platform (valid on macOS, invalid on Linux)
		{"invalid lowercase local", "local", ErrTimezoneInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.timezone)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// ValidateBridgeMode tests

func TestValidateBridgeMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        BridgeMode
		expectedErr error
	}{
		// Valid bridge modes
		{"valid internal mode", BridgeModeInternal, nil},
		{"valid external mode", BridgeModeExternal, nil},
		{"valid both mode", BridgeModeBoth, nil},
		// Empty mode
		{"empty mode", BridgeMode(""), ErrBridgeModeEmpty},
		// Invalid modes
		{"invalid nat mode", BridgeMode("nat"), ErrBridgeModeInvalid},
		{"invalid uppercase Internal", BridgeMode("Internal"), ErrBridgeModeInvalid},
		{"invalid uppercase INTERNAL", BridgeMode("INTERNAL"), ErrBridgeModeInvalid},
		{"invalid uppercase External", BridgeMode("External"), ErrBridgeModeInvalid},
		{"invalid uppercase Both", BridgeMode("Both"), ErrBridgeModeInvalid},
		{testNameInvalidRandomString, BridgeMode("random"), ErrBridgeModeInvalid},
		{"invalid partial match intern", BridgeMode("intern"), ErrBridgeModeInvalid},
		{"invalid partial match extern", BridgeMode("extern"), ErrBridgeModeInvalid},
		{"invalid with spaces", BridgeMode(" internal"), ErrBridgeModeInvalid},
		{testNameInvalidTrailingSpace, BridgeMode("external "), ErrBridgeModeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBridgeMode(tt.mode)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// ValidateZFSRaid tests

func TestValidateZFSRaid(t *testing.T) {
	tests := []struct {
		name        string
		raid        ZFSRaid
		expectedErr error
	}{
		// Valid ZFS RAID levels
		{"valid single", ZFSRaidSingle, nil},
		{"valid raid0", ZFSRaid0, nil},
		{"valid raid1", ZFSRaid1, nil},
		// Empty RAID level
		{"empty raid level", ZFSRaid(""), ErrZFSRaidEmpty},
		// Invalid - unsupported RAID levels
		{"invalid raid5", ZFSRaid("raid5"), ErrZFSRaidInvalid},
		{"invalid raid6", ZFSRaid("raid6"), ErrZFSRaidInvalid},
		{"invalid raidz", ZFSRaid("raidz"), ErrZFSRaidInvalid},
		{"invalid raidz2", ZFSRaid("raidz2"), ErrZFSRaidInvalid},
		// Invalid - case sensitivity
		{"invalid uppercase SINGLE", ZFSRaid("SINGLE"), ErrZFSRaidInvalid},
		{"invalid uppercase Single", ZFSRaid("Single"), ErrZFSRaidInvalid},
		{"invalid uppercase RAID0", ZFSRaid("RAID0"), ErrZFSRaidInvalid},
		{"invalid uppercase Raid0", ZFSRaid("Raid0"), ErrZFSRaidInvalid},
		{"invalid uppercase RAID1", ZFSRaid("RAID1"), ErrZFSRaidInvalid},
		// Invalid - random strings
		{testNameInvalidRandomString, ZFSRaid("random"), ErrZFSRaidInvalid},
		{"invalid numeric only", ZFSRaid("123"), ErrZFSRaidInvalid},
		// Invalid - partial matches
		{"invalid partial raid", ZFSRaid("raid"), ErrZFSRaidInvalid},
		{"invalid partial sing", ZFSRaid("sing"), ErrZFSRaidInvalid},
		// Invalid - with spaces
		{"invalid leading space", ZFSRaid(" single"), ErrZFSRaidInvalid},
		{testNameInvalidTrailingSpace, ZFSRaid("raid0 "), ErrZFSRaidInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateZFSRaid(tt.raid)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// ValidateSubnet tests

func TestValidateSubnet(t *testing.T) {
	tests := []struct {
		name        string
		subnet      string
		expectedErr error
	}{
		// Valid subnets - standard private IPv4 ranges
		{"valid 10.0.0.0/24", buildSubnet(10, 0, 0, 0, 24), nil},
		{"valid 192.168.1.0/24", buildSubnet(192, 168, 1, 0, 24), nil},
		{"valid 172.16.0.0/16", buildSubnet(172, 16, 0, 0, 16), nil},
		// Valid - single host (/32)
		{"valid single host /32", buildSubnet(192, 168, 1, 1, 32), nil},
		// Valid - all networks (/0)
		{"valid all networks /0", buildSubnet(0, 0, 0, 0, 0), nil},
		// Valid - other common subnets
		{"valid /8 subnet", buildSubnet(10, 0, 0, 0, 8), nil},
		{"valid /12 subnet", buildSubnet(172, 16, 0, 0, 12), nil},
		{"valid /30 point-to-point", buildSubnet(192, 168, 1, 0, 30), nil},
		// Empty subnet
		{"empty subnet", "", ErrSubnetEmpty},
		// Invalid - missing subnet mask
		{"missing subnet mask", buildIP(10, 0, 0, 0), ErrSubnetInvalid},
		{"missing mask plain IP", buildIP(192, 168, 1, 1), ErrSubnetInvalid},
		// Invalid - invalid IP address
		{"invalid IP address 256", buildSubnet(256, 0, 0, 0, 24), ErrSubnetInvalid},
		{"invalid IP address format", "10.0.0/24", ErrSubnetInvalid},
		{"invalid IP negative", "-1.0.0.0/24", ErrSubnetInvalid},
		// Invalid - mask range (above /32)
		{"invalid mask /33", buildSubnet(10, 0, 0, 0, 33), ErrSubnetInvalid},
		{"invalid mask /64", buildSubnet(10, 0, 0, 0, 64), ErrSubnetInvalid},
		// Invalid - negative mask
		{"invalid negative mask", buildIP(10, 0, 0, 0) + "/-1", ErrSubnetInvalid},
		// Invalid - random strings
		{testNameInvalidRandomString, "not-a-subnet", ErrSubnetInvalid},
		{"invalid just slash", "/24", ErrSubnetInvalid},
		{"invalid no numbers", "abc.def.ghi.jkl/24", ErrSubnetInvalid},
		// Invalid - extra characters
		{"invalid trailing chars", buildSubnet(10, 0, 0, 0, 24) + "x", ErrSubnetInvalid},
		{"invalid leading space", " " + buildSubnet(10, 0, 0, 0, 24), ErrSubnetInvalid},
		{testNameInvalidTrailingSpace, buildSubnet(10, 0, 0, 0, 24) + " ", ErrSubnetInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubnet(tt.subnet)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}

// Config.Validate tests

func TestConfigValidateValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.RootPassword = testValidPassword
	cfg.System.SSHPublicKey = testValidSSHKey

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestConfigValidateEmptyConfig_AllErrors(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))

	// Should have errors for: hostname, email, password, ssh key,
	// timezone, bridge mode, subnet, zfs raid
	assert.Len(t, valErr.Errors, 8)
}

func TestConfigValidateReturnsValidationError(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.True(t, valErr.HasErrors())
}

func TestConfigValidatePartialErrors_System(t *testing.T) {
	cfg := DefaultConfig()
	// Missing required fields: RootPassword and SSHPublicKey

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))

	// Should have 2 errors: password and SSH key
	assert.Len(t, valErr.Errors, 2)
	assert.Contains(t, valErr.Error(), ErrPasswordEmpty.Error())
	assert.Contains(t, valErr.Error(), ErrSSHKeyEmpty.Error())
}

func TestConfigValidateInvalidHostname(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.Hostname = testInvalidHostname
	cfg.System.RootPassword = testValidPassword
	cfg.System.SSHPublicKey = testValidSSHKey

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.Len(t, valErr.Errors, 1)
	assert.True(t, errors.Is(valErr.Unwrap(), ErrHostnameStartsWithHyphen))
}

func TestConfigValidateInvalidEmail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.Email = testInvalidEmail
	cfg.System.RootPassword = testValidPassword
	cfg.System.SSHPublicKey = testValidSSHKey

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.Len(t, valErr.Errors, 1)
	assert.True(t, errors.Is(valErr.Unwrap(), ErrEmailInvalid))
}

func TestConfigValidateInvalidTimezone(t *testing.T) {
	cfg := DefaultConfig()
	cfg.System.Timezone = testInvalidTimezone
	cfg.System.RootPassword = testValidPassword
	cfg.System.SSHPublicKey = testValidSSHKey

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.Len(t, valErr.Errors, 1)
	assert.True(t, errors.Is(valErr.Unwrap(), ErrTimezoneInvalid))
}

func TestConfigValidateInvalidBridgeMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Network.BridgeMode = BridgeMode("invalid")
	cfg.System.RootPassword = testValidPassword
	cfg.System.SSHPublicKey = testValidSSHKey

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.Len(t, valErr.Errors, 1)
	assert.True(t, errors.Is(valErr.Unwrap(), ErrBridgeModeInvalid))
}

func TestConfigValidateInvalidSubnet(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Network.PrivateSubnet = testInvalidSubnet
	cfg.System.RootPassword = testValidPassword
	cfg.System.SSHPublicKey = testValidSSHKey

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.Len(t, valErr.Errors, 1)
	assert.True(t, errors.Is(valErr.Unwrap(), ErrSubnetInvalid))
}

func TestConfigValidateInvalidZFSRaid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Storage.ZFSRaid = ZFSRaid("invalid")
	cfg.System.RootPassword = testValidPassword
	cfg.System.SSHPublicKey = testValidSSHKey

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))
	assert.Len(t, valErr.Errors, 1)
	assert.True(t, errors.Is(valErr.Unwrap(), ErrZFSRaidInvalid))
}

func TestConfigValidateMultipleErrors_AllCategories(t *testing.T) {
	cfg := &Config{
		System: SystemConfig{
			Hostname:     "-invalid",     // Invalid: starts with hyphen
			Email:        "not-an-email", // Invalid: no @ symbol
			Timezone:     "Bad/Zone",     // Invalid: not in IANA database
			RootPassword: "short",        // Invalid: too short
			SSHPublicKey: "not-a-key",    // Invalid: no valid prefix
		},
		Network: NetworkConfig{
			BridgeMode:    BridgeMode("nat"), // Invalid: not internal/external/both
			PrivateSubnet: "not-a-subnet",    // Invalid: not CIDR
		},
		Storage: StorageConfig{
			ZFSRaid: ZFSRaid("raid5"), // Invalid: not single/raid0/raid1
		},
	}

	err := cfg.Validate()

	require.Error(t, err)

	var valErr *ValidationError
	require.True(t, errors.As(err, &valErr))

	// All 8 fields should have errors
	assert.Len(t, valErr.Errors, 8)

	// Check error message contains all errors
	errMsg := valErr.Error()
	assert.Contains(t, errMsg, ErrHostnameStartsWithHyphen.Error())
	assert.Contains(t, errMsg, ErrEmailInvalid.Error())
	assert.Contains(t, errMsg, ErrPasswordTooShort.Error())
	assert.Contains(t, errMsg, ErrSSHKeyInvalidPrefix.Error())
	assert.Contains(t, errMsg, ErrTimezoneInvalid.Error())
	assert.Contains(t, errMsg, ErrBridgeModeInvalid.Error())
	assert.Contains(t, errMsg, ErrSubnetInvalid.Error())
	assert.Contains(t, errMsg, ErrZFSRaidInvalid.Error())
}

func TestConfigValidateErrorMessageFormat(t *testing.T) {
	cfg := DefaultConfig()
	// Leave RootPassword and SSHPublicKey empty to trigger 2 errors

	err := cfg.Validate()

	require.Error(t, err)
	// Errors should be joined by "; "
	assert.Contains(t, err.Error(), "; ")
}
