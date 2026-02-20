package api

import (
	"errors"
	"testing"
)

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{Resource: "test-resource"}

	expectedMsg := "Page not found: test-resource"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}

	if err.ExitCode() != ExitNotFound {
		t.Errorf("Expected exit code %d, got %d", ExitNotFound, err.ExitCode())
	}
}

func TestRateLimitError(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter int
		expected   string
	}{
		{
			name:       "with retry after",
			retryAfter: 60,
			expected:   "Rate limited. Retry after 60 seconds",
		},
		{
			name:       "without retry after",
			retryAfter: 0,
			expected:   "Rate limited. Please try again later",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &RateLimitError{RetryAfter: tt.retryAfter}

			if err.Error() != tt.expected {
				t.Errorf("Expected error message %q, got %q", tt.expected, err.Error())
			}

			if err.ExitCode() != ExitRateLimited {
				t.Errorf("Expected exit code %d, got %d", ExitRateLimited, err.ExitCode())
			}
		})
	}
}

func TestInvalidArgsError(t *testing.T) {
	err := &InvalidArgsError{Message: "invalid argument test"}

	expectedMsg := "invalid argument test"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}

	if err.ExitCode() != ExitInvalidArgs {
		t.Errorf("Expected exit code %d, got %d", ExitInvalidArgs, err.ExitCode())
	}
}

func TestNetworkError(t *testing.T) {
	err := &NetworkError{Message: "connection refused"}

	expectedMsg := "Network error: connection refused"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}

	if err.ExitCode() != ExitGenericError {
		t.Errorf("Expected exit code %d, got %d", ExitGenericError, err.ExitCode())
	}
}

func TestUnknownConstantError(t *testing.T) {
	err := &UnknownConstantError{Key: "unknown_key"}

	expectedMsg := "Unknown constant: unknown_key"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}

	if err.ExitCode() != ExitNotFound {
		t.Errorf("Expected exit code %d, got %d", ExitNotFound, err.ExitCode())
	}
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: ExitSuccess,
		},
		{
			name:     "not found error",
			err:      &NotFoundError{Resource: "test"},
			expected: ExitNotFound,
		},
		{
			name:     "rate limit error",
			err:      &RateLimitError{},
			expected: ExitRateLimited,
		},
		{
			name:     "invalid args error",
			err:      &InvalidArgsError{},
			expected: ExitInvalidArgs,
		},
		{
			name:     "network error",
			err:      &NetworkError{},
			expected: ExitGenericError,
		},
		{
			name:     "unknown constant error",
			err:      &UnknownConstantError{},
			expected: ExitNotFound,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: ExitGenericError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			if got != tt.expected {
				t.Errorf("GetExitCode() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestExitCodes(t *testing.T) {
	// Verify all exit codes have expected values
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess should be 0, got %d", ExitSuccess)
	}
	if ExitGenericError != 1 {
		t.Errorf("ExitGenericError should be 1, got %d", ExitGenericError)
	}
	if ExitNotFound != 2 {
		t.Errorf("ExitNotFound should be 2, got %d", ExitNotFound)
	}
	if ExitRateLimited != 3 {
		t.Errorf("ExitRateLimited should be 3, got %d", ExitRateLimited)
	}
	if ExitInvalidArgs != 4 {
		t.Errorf("ExitInvalidArgs should be 4, got %d", ExitInvalidArgs)
	}
}
