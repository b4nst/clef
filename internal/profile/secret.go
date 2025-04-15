package profile

import (
	"context"
	"fmt"

	"github.com/b4nst/clef/internal/backend"
)

// Injector is a function that takes a key (secret name) and value (secret plain) and inject it into the system.
// It returns an error on failed injection.
//
// example: [os.Setenv]
type Injector func(k, v string) error

// Secret represents a secret configuration with its source and target information.
// It defines where a secret should be loaded from and where it should be injected.
type Secret struct {
	// Key is the identifier of the secret in the storage backend
	Key string `toml:"key"`
	// Store specifies which backend store to use (optional, uses default if empty)
	Store string `toml:"store,omitempty"`
	// Target is the name to use when injecting the secret (defaults to Key if empty)
	Target string `toml:"target,omitempty"`
}

// Inject loads a secret from the specified store and injects it using the provided function.
// It will use the Key to fetch the secret and inject it with the Target name (or Key if Target is empty).
func (s *Secret) Inject(ctx context.Context, injectf Injector, loader backend.StoreLoader) error {
	store, err := loader.Backend(s.Store)
	if err != nil {
		return fmt.Errorf("load store '%s': %w", s.Store, err)
	}

	plain, err := store.Get(ctx, s.Key)
	if err != nil {
		return fmt.Errorf("get %s: %w", s.Key, err)
	}

	// default target to key
	if s.Target == "" {
		s.Target = s.Key
	}

	if err := injectf(s.Target, plain); err != nil {
		return fmt.Errorf("inject %s: %w", s.Target, err)
	}

	return nil
}
