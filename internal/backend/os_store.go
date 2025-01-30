package backend

import (
	"context"

	"github.com/zalando/go-keyring"
)

func init() {
	registerBuilder("osstore", func() Builder { return new(OSStoreBuilder) })
}

type OSStoreBuilder struct {
	Namespace string `toml:"namespace"`
}

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
