package backend

import (
	"context"
	"errors"
)

// ErrNotFound means the key hasn't been found on the store
var ErrNotFound = errors.New("resource not found")

// Store represents a store abstraction.
type Store interface {
	// Get returns the value at key from the store, or an error.
	Get(ctx context.Context, key string) (string, error)
	// Set store the value in the store at key.
	Set(ctx context.Context, key, value string) error
}
