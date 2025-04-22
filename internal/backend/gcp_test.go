package backend

import (
	"context"
	"errors"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestGCPStore_Get(t *testing.T) {
	t.Parallel()

	t.Run("access error", func(t *testing.T) {
		t.Parallel()

		therr := errors.New("error")
		smc := NewMockSecretManagerClient(t)
		smc.EXPECT().
			AccessSecretVersion(mock.Anything, &secretmanagerpb.AccessSecretVersionRequest{Name: "projects/pid/secrets/foo/versions/latest"}).
			Return(nil, therr)
		store := &GCPStore{client: smc, projectID: "pid"}

		_, err := store.Get(context.TODO(), "foo")
		assert.ErrorIs(t, err, therr)
	})

	t.Run("nominal", func(t *testing.T) {
		t.Parallel()

		smc := NewMockSecretManagerClient(t)
		smc.EXPECT().
			AccessSecretVersion(mock.Anything, &secretmanagerpb.AccessSecretVersionRequest{Name: "projects/pid/secrets/foo/versions/latest"}).
			Return(&secretmanagerpb.AccessSecretVersionResponse{
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("bar"),
				},
			}, nil)
		store := &GCPStore{client: smc, projectID: "pid"}

		value, err := store.Get(context.TODO(), "foo")
		assert.NoError(t, err)
		assert.Equal(t, "bar", value)
	})
}

func TestGCPStore_Set(t *testing.T) {
	t.Parallel()

	t.Run("existing secret", func(t *testing.T) {
		t.Parallel()
		t.Skip("Need a way to mock secret version iterator")

		smc := NewMockSecretManagerClient(t)
		smc.EXPECT().
			GetSecret(mock.Anything, mock.Anything).
			Return(&secretmanagerpb.Secret{Name: "the/secret"}, nil)
		smc.EXPECT().
			AddSecretVersion(mock.Anything, mock.Anything).
			Return(&secretmanagerpb.SecretVersion{Name: "the/secret/versions/latest"}, nil)
		smc.EXPECT().
			ListSecretVersions(mock.Anything, mock.Anything).
			Return(&secretmanager.SecretVersionIterator{})

		store := &GCPStore{client: smc, projectID: "pid"}
		assert.NoError(t, store.Set(context.TODO(), "foo", "bar"))

	})
}

func TestGCPStore_Delete(t *testing.T) {
	t.Parallel()

	t.Run("gcp error", func(t *testing.T) {
		t.Parallel()

		therr := errors.New("error")
		smc := NewMockSecretManagerClient(t)
		smc.EXPECT().
			DeleteSecret(mock.Anything, &secretmanagerpb.DeleteSecretRequest{Name: "projects/pid/secrets/foo"}).
			Return(therr)
		store := &GCPStore{client: smc, projectID: "pid"}

		assert.ErrorIs(t, store.Delete(context.TODO(), "foo"), therr)
	})

	t.Run("nominal", func(t *testing.T) {
		t.Parallel()
		smc := NewMockSecretManagerClient(t)
		smc.EXPECT().
			DeleteSecret(mock.Anything, &secretmanagerpb.DeleteSecretRequest{Name: "projects/pid/secrets/foo"}).
			Return(nil)
		store := &GCPStore{client: smc, projectID: "pid"}

		assert.NoError(t, store.Delete(context.TODO(), "foo"))
	})
}
