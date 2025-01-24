package backend

import "context"

// Store represents a store abstraction.
type Store interface {
	// Get returns the value at key from the store, or an error.
	Get(ctx context.Context, key string) (error, string)
	// Set store the value in the store at key.
	Set(ctx context.Context, key, value string) error
}
