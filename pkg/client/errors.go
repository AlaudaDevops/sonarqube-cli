package client

import (
	"errors"
	"fmt"
)

// ErrAlreadyExists indicates the target resource already exists.
var ErrAlreadyExists = errors.New("resource already exists")

func alreadyExistsError(resourceType, resourceName string, err error) error {
	return fmt.Errorf("%w: %s %q: %v", ErrAlreadyExists, resourceType, resourceName, err)
}
