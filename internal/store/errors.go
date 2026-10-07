package store

import "errors"

// ErrNotFound is returned when a looked-up row does not exist.
var ErrNotFound = errors.New("not found")
