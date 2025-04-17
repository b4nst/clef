package profile

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/b4nst/clef/internal/backend"
)

var ErrEmptyKey = fmt.Errorf("key cannot be empty")

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

// Decode implements a custom mapper for kong.
// It is not possible to use [encoding.TextUnmarshaler] instead, as it would conflict with TOML.
func (s *Secret) Decode(ktx *kong.DecodeContext) error {
	var text string
	err := ktx.Scan.PopValueInto("value", &text)
	if err != nil {
		return err
	}

	return s.DecodeText(text)
}

// DecodeText decode a [Secret] from a string.
func (s *Secret) DecodeText(text string) error {
	keyStart := 0
	keyEnd := len(text)
	targetStart := -1
	storeEnd := -1

	// Scan from right to left
	for i := len(text) - 1; i >= 0; i-- {
		if targetStart == -1 && text[i] == '=' {
			// Found the first '=' from the right
			keyEnd = i
			targetStart = i + 1
		} else if text[i] == '.' && storeEnd == -1 {
			// Found the first '.' from the right before any '='
			storeEnd = i
			keyStart = i + 1
			// We can break early as we've found all we need
			break
		}
	}

	s.Key = text[keyStart:keyEnd]
	// Ensure Key is not empty
	if s.Key == "" {
		return ErrEmptyKey
	}

	// Extract the components
	if storeEnd != -1 {
		s.Store = text[:storeEnd]
	}

	if targetStart != -1 {
		s.Target = text[targetStart:]
	}

	return nil
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
