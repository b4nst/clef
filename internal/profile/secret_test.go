package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/b4nst/clef/internal/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSecret_Inject(t *testing.T) {
	t.Parallel()

	t.Run("no error", func(t *testing.T) {
		system := map[string]string{}
		injector := func(k, v string) error {
			system[k] = v
			return nil
		}

		store := backend.NewMockStore(t)
		store.EXPECT().Get(mock.Anything, "foo").Return("bar", nil).Once()
		loader := backend.NewMockStoreLoader(t)
		loader.EXPECT().Backend(context.TODO(), "default").Return(store, nil).Once()

		secret := Secret{
			Key:   "foo",
			Store: "default",
		}
		require.NoError(t, secret.Inject(context.TODO(), injector, loader))
		assert.Contains(t, system, "foo")
		assert.Equal(t, "bar", system["foo"])
	})

	t.Run("load store failure", func(t *testing.T) {
		therr := errors.New("oops")
		loader := backend.NewMockStoreLoader(t)
		loader.EXPECT().Backend(context.TODO(), "default").Return(nil, therr)

		secret := Secret{Key: "foo", Store: "default", Target: "bar"}
		assert.ErrorIs(t, secret.Inject(context.TODO(), nil, loader), therr)
	})

	t.Run("get secret failure", func(t *testing.T) {
		therr := errors.New("oops")
		store := backend.NewMockStore(t)
		store.EXPECT().Get(mock.Anything, "foo").Return("", therr).Once()
		loader := backend.NewMockStoreLoader(t)
		loader.EXPECT().Backend(context.TODO(), mock.AnythingOfType("string")).Return(store, nil)

		secret := Secret{Key: "foo", Store: "default", Target: "bar"}
		assert.ErrorIs(t, secret.Inject(context.TODO(), nil, loader), therr)
	})

	t.Run("inject secret failure", func(t *testing.T) {
		therr := errors.New("oops")
		injector := func(k, v string) error {
			return therr
		}

		store := backend.NewMockStore(t)
		store.EXPECT().Get(mock.Anything, "foo").Return("bar", nil).Once()
		loader := backend.NewMockStoreLoader(t)
		loader.EXPECT().Backend(context.TODO(), "default").Return(store, nil).Once()

		secret := Secret{Key: "foo", Store: "default", Target: "bar"}
		assert.ErrorIs(t, secret.Inject(context.TODO(), injector, loader), therr)
	})
}

func TestSecret_UnmarshalText(t *testing.T) {
	tcs := map[string]struct {
		in  string
		out Secret
		err error
	}{
		"nostore":  {"key=target", Secret{"key", "", "target"}, nil},
		"notarget": {"store.key", Secret{"key", "store", ""}, nil},
		"keyonly":  {"key", Secret{"key", "", ""}, nil},
		"all":      {"store.key=target", Secret{"key", "store", "target"}, nil},
		"empty":    {"", Secret{}, ErrEmptyKey},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := &Secret{}
			err := s.DecodeText(tc.in)

			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, &tc.out, s)
		})
	}
}
