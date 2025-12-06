package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test error constants for validation tests.
var (
	errHostnameEmpty   = errors.New("hostname is required")
	errEmailInvalid    = errors.New("email format is invalid")
	errPasswordTooWeak = errors.New("password must be at least 8 characters")
)

func TestValidationError_Error_Empty(t *testing.T) {
	ve := &ValidationError{}

	result := ve.Error()

	assert.Equal(t, "", result)
}

func TestValidationError_Error_SingleError(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	result := ve.Error()

	assert.Equal(t, "hostname is required", result)
}

func TestValidationError_Error_MultipleErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
			errPasswordTooWeak,
		},
	}

	result := ve.Error()

	assert.Equal(t, "hostname is required; email format is invalid; password must be at least 8 characters", result)
}

func TestValidationError_Error_TwoErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
		},
	}

	result := ve.Error()

	assert.Equal(t, "hostname is required; email format is invalid", result)
}

func TestValidationError_HasErrors_Empty(t *testing.T) {
	ve := &ValidationError{}

	assert.False(t, ve.HasErrors())
}

func TestValidationError_HasErrors_WithErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	assert.True(t, ve.HasErrors())
}

func TestValidationError_HasErrors_MultipleErrors(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{
			errHostnameEmpty,
			errEmailInvalid,
		},
	}

	assert.True(t, ve.HasErrors())
}

func TestValidationError_Unwrap_Empty(t *testing.T) {
	ve := &ValidationError{}

	result := ve.Unwrap()

	assert.Nil(t, result)
}

func TestValidationError_Unwrap_SingleError(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	result := ve.Unwrap()

	assert.Equal(t, errHostnameEmpty, result)
}

func TestValidationError_Unwrap_MultipleErrors_ReturnsFirst(t *testing.T) {
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

func TestValidationError_Unwrap_WorksWithErrorsIs(t *testing.T) {
	ve := &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	assert.True(t, errors.Is(ve, errHostnameEmpty))
}

func TestValidationError_Unwrap_ErrorsIsWithMultiple(t *testing.T) {
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

func TestValidationError_ImplementsErrorInterface(t *testing.T) {
	var err error = &ValidationError{
		Errors: []error{errHostnameEmpty},
	}

	assert.NotNil(t, err)
	assert.Equal(t, "hostname is required", err.Error())
}

func TestValidationError_Add_NilError(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(nil)

	assert.False(t, ve.HasErrors())
	assert.Empty(t, ve.Errors)
}

func TestValidationError_Add_SingleError(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(errHostnameEmpty)

	require.Len(t, ve.Errors, 1)
	assert.Equal(t, errHostnameEmpty, ve.Errors[0])
}

func TestValidationError_Add_MultipleErrors(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(errHostnameEmpty)
	ve.Add(errEmailInvalid)
	ve.Add(errPasswordTooWeak)

	require.Len(t, ve.Errors, 3)
	assert.Equal(t, errHostnameEmpty, ve.Errors[0])
	assert.Equal(t, errEmailInvalid, ve.Errors[1])
	assert.Equal(t, errPasswordTooWeak, ve.Errors[2])
}

func TestValidationError_Add_IgnoresNilBetweenErrors(t *testing.T) {
	ve := &ValidationError{}

	ve.Add(errHostnameEmpty)
	ve.Add(nil)
	ve.Add(errEmailInvalid)
	ve.Add(nil)

	require.Len(t, ve.Errors, 2)
	assert.Equal(t, errHostnameEmpty, ve.Errors[0])
	assert.Equal(t, errEmailInvalid, ve.Errors[1])
}

func TestValidationError_Add_PreservesOrder(t *testing.T) {
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
		assert.Equal(t, err.Error(), ve.Errors[i].Error())
	}
}

func TestValidationError_ZeroValue(t *testing.T) {
	var ve ValidationError

	assert.False(t, ve.HasErrors())
	assert.Equal(t, "", ve.Error())
	assert.Nil(t, ve.Unwrap())
}

func TestValidationError_TableDriven(t *testing.T) {
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
			expectedString: "hostname is required",
			hasErrors:      true,
			unwrapResult:   errHostnameEmpty,
		},
		{
			name:           "two errors",
			errors:         []error{errHostnameEmpty, errEmailInvalid},
			expectedString: "hostname is required; email format is invalid",
			hasErrors:      true,
			unwrapResult:   errHostnameEmpty,
		},
		{
			name:           "three errors",
			errors:         []error{errHostnameEmpty, errEmailInvalid, errPasswordTooWeak},
			expectedString: "hostname is required; email format is invalid; password must be at least 8 characters",
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
