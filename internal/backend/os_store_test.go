package backend

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func TestOSStoreBuilderBuild(t *testing.T) {
	t.Parallel()

	t.Run("empty namespace", func(t *testing.T) {
		t.Parallel()

		builder := &OSStoreBuilder{}
		s, err := builder.Build("default")
		if assert.NoError(t, err) {
			assert.Equal(t, "default", builder.Namespace)
			assert.Implements(t, (*Store)(nil), s)
		}
	})

	t.Run("reserved namespace", func(t *testing.T) {
		t.Parallel()

		builder := &OSStoreBuilder{Namespace: SystemStoreNameSpace}
		s, err := builder.Build("default")
		if assert.ErrorIs(t, err, ErrReservedStoreName) {
			assert.Nil(t, s)
		}
	})

	t.Run("nominal", func(t *testing.T) {
		t.Parallel()

		builder := &OSStoreBuilder{Namespace: "foo"}
		s, err := builder.Build("bar")
		if assert.NoError(t, err) {
			assert.Equal(t, "foo", builder.Namespace)
			assert.Implements(t, (*Store)(nil), s)
		}
	})
}

func TestNewOSStore(t *testing.T) {
	t.Parallel()

	t.Run("nominal", func(t *testing.T) {
		t.Parallel()

		s, err := NewOSStore("foo")
		if assert.NoError(t, err) {
			assert.NotNil(t, s)
			assert.Contains(t, s.service, "foo")
		}
	})
}

func TestOSStore_Get(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		kerr := errors.New("test error")
		keyring.MockInitWithError(kerr)

		s := newOSStore("test_foo")
		v, err := s.Get(context.TODO(), "test_no_var")
		if assert.ErrorIs(t, err, kerr) {
			assert.Empty(t, v)
		}
	})

	t.Run("nominal", func(t *testing.T) {
		keyring.MockInit()

		s := newOSStore("test_foo")
		keyring.Set(s.service, "test_key", "test_value")
		t.Cleanup(func() { keyring.Delete(s.service, "test_key") })

		v, err := s.Get(context.TODO(), "test_key")
		if assert.NoError(t, err) {
			assert.Equal(t, "test_value", v)
		}
	})
}

func TestOSStore_Set(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		kerr := errors.New("test set error")
		keyring.MockInitWithError(kerr)

		s := newOSStore("test_foo")
		assert.ErrorIs(t, s.Set(context.TODO(), "test_key", "test_value"), kerr)
	})

	t.Run("nominal", func(t *testing.T) {
		keyring.MockInit()

		s := newOSStore("test_foo")
		if err := s.Set(context.TODO(), "test_key", "test_value"); assert.NoError(t, err) {
			v, err := keyring.Get(s.service, "test_key")
			if assert.NoError(t, err) {
				assert.Equal(t, "test_value", v)
			}
		}
	})
}
