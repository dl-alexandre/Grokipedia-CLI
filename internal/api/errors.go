package api

import (
	"errors"
	"fmt"
)

// Exit codes as defined in the spec
const (
	ExitSuccess      = 0
	ExitGenericError = 1
	ExitNotFound     = 2
	ExitRateLimited  = 3
	ExitInvalidArgs  = 4
)

// APIError is the base error type for API errors
type APIError struct {
	Message  string
	ExitCode int
}

func (e *APIError) Error() string {
	return e.Message
}

// NotFoundError represents a 404 or missing resource error
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Page not found: %s", e.Resource)
}

func (e *NotFoundError) ExitCode() int {
	return ExitNotFound
}

// RateLimitError represents a 429 rate limit error
type RateLimitError struct {
	RetryAfter int // seconds
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("Rate limited. Retry after %d seconds", e.RetryAfter)
	}
	return "Rate limited. Please try again later"
}

func (e *RateLimitError) ExitCode() int {
	return ExitRateLimited
}

// InvalidArgsError represents invalid arguments or flags
type InvalidArgsError struct {
	Message string
}

func (e *InvalidArgsError) Error() string {
	return e.Message
}

func (e *InvalidArgsError) ExitCode() int {
	return ExitInvalidArgs
}

// NetworkError represents network/timeout errors
type NetworkError struct {
	Message string
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("Network error: %s", e.Message)
}

func (e *NetworkError) ExitCode() int {
	return ExitGenericError
}

// UnknownConstantError represents an unknown constant key
type UnknownConstantError struct {
	Key string
}

func (e *UnknownConstantError) Error() string {
	return fmt.Sprintf("Unknown constant: %s", e.Key)
}

func (e *UnknownConstantError) ExitCode() int {
	return ExitNotFound
}

// GetExitCode returns the exit code for an error
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var exitCoder interface{ ExitCode() int }
	if errors.As(err, &exitCoder) {
		return exitCoder.ExitCode()
	}
	return ExitGenericError
}
