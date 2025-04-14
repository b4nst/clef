package backend

import (
	"context"
	"errors"
)

var (
	// ErrKeyNotFound means the key hasn't been found on the store
	ErrKeyNotFound = errors.New("key not found")
	// ErrReservedStoreName means the store name cannot be used
	ErrReservedStoreName = errors.New("reserved store name")
)

// Store represents a store abstraction.
type Store interface {
	// Get returns the value at key from the store, or an error.
	Get(ctx context.Context, key string) (string, error)
	// Set store the value in the store at key.
	Set(ctx context.Context, key, value string) error
	// Delete delete the secret from the store
	Delete(ctx context.Context, key string) error
}
