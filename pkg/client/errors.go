package client

import (
	"errors"
	"fmt"
	"strings"
)

// ErrAlreadyExists indicates the target resource already exists.
var ErrAlreadyExists = errors.New("resource already exists")

// alreadyExistsError wraps vendor-specific API errors into a stable sentinel error for callers.
func alreadyExistsError(resourceType, resourceName string, err error) error {
	return fmt.Errorf("%w: %s %q: %v", ErrAlreadyExists, resourceType, resourceName, err)
}

// APIError captures a non-successful SonarQube API response.
type APIError struct {
	StatusCode int
	Method     string
	URL        string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d on %s %s: %s", e.StatusCode, e.Method, e.URL, e.Body)
}

// hasStatus reports whether err is an APIError with one of the provided HTTP status codes.
func hasStatus(err error, statusCodes ...int) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, code := range statusCodes {
		if apiErr.StatusCode == code {
			return true
		}
	}
	return false
}

// bodyContains reports whether err is an APIError whose response body contains any given marker.
func bodyContains(err error, substrings ...string) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, substring := range substrings {
		if strings.Contains(apiErr.Body, substring) {
			return true
		}
	}
	return false
}
