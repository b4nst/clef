package backend

import (
	"context"

	"github.com/zalando/go-keyring"
)

func init() {
	registerBuilder("osstore", func() Builder { return new(OSStoreBuilder) })
}

// OSStoreBuilder implements the Builder interface for OSStore.
type OSStoreBuilder struct {
	Namespace string `toml:"namespace"`
}

// Build returns a new OSStore store.
func (ob *OSStoreBuilder) Build(name string) (Store, error) {
	if ob.Namespace == "" {
		ob.Namespace = name
	}
	if ob.Namespace == SystemStoreNameSpace {
		return nil, ErrReservedStoreName
	}
	return NewOSStore(ob.Namespace)
}

// OSStore uses the operating system keyring to store secrets
//
// # OS Specific Details
//
// OS X
// The OS X implementation depends on the /usr/bin/security binary for interfacing with the OS X keychain. It should be available by default.
//
// Linux and *BSD
// The Linux and *BSD implementation depends on the [Secret Service] dbus interface,
// which is provided by [GNOME Keyring].
// It's expected that the default collection login exists in the keyring, because it's the default in most distros.
// If it doesn't exist, you can create it through the keyring frontend program [Seahorse]:
// Open seahorse
// Go to File > New > Password Keyring
// Click Continue
// When asked for a name, use: login
//
// [Secret Service]: https://specifications.freedesktop.org/secret-service/latest/
// [GNOME Keyring]: https://wiki.gnome.org/Projects/GnomeKeyring
// [Seahorse]: https://wiki.gnome.org/Apps/Seahorse
type OSStore struct {
	service string
}

// NewOSStore creates a new namespaced OS Store.
func NewOSStore(namespace string) (*OSStore, error) {
	return newOSStore(namespace), nil
}

func newOSStore(namespace string) *OSStore {
	service := "clef:" + namespace

	return &OSStore{service}
}

// Get implements the Store.Get method.
func (o *OSStore) Get(ctx context.Context, k string) (string, error) {
	secret, err := keyring.Get(o.service, k)
	if err != nil {
		return "", err
	}
	return secret, nil
}

// Set implements the Store.Set method
func (o *OSStore) Set(ctx context.Context, k, v string) error {
	return keyring.Set(o.service, k, v)
}

// Delete implements the Store.Delete method.
func (o *OSStore) Delete(ctx context.Context, k string) error {
	return keyring.Delete(o.service, k)
}
