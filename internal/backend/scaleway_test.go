package backend

import (
	"context"
	"errors"
	"testing"

	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockScalewaySecretClient is a mock implementation of ScalewaySecretClient
type MockScalewaySecretClient struct {
	mock.Mock
}

func (m *MockScalewaySecretClient) AccessSecretVersionByPath(req *secret.AccessSecretVersionByPathRequest, opts ...scw.RequestOption) (*secret.AccessSecretVersionResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secret.AccessSecretVersionResponse), args.Error(1)
}

func (m *MockScalewaySecretClient) CreateSecret(req *secret.CreateSecretRequest, opts ...scw.RequestOption) (*secret.Secret, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secret.Secret), args.Error(1)
}

func (m *MockScalewaySecretClient) CreateSecretVersion(req *secret.CreateSecretVersionRequest, opts ...scw.RequestOption) (*secret.SecretVersion, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secret.SecretVersion), args.Error(1)
}

func (m *MockScalewaySecretClient) ListSecrets(req *secret.ListSecretsRequest, opts ...scw.RequestOption) (*secret.ListSecretsResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secret.ListSecretsResponse), args.Error(1)
}

func (m *MockScalewaySecretClient) DeleteSecret(req *secret.DeleteSecretRequest, opts ...scw.RequestOption) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestScalewayStoreBuilder_Build(t *testing.T) {
	t.Run("missing region", func(t *testing.T) {
		builder := ScalewayStoreBuilder{ProjectID: "proj-id"}
		store, err := builder.Build(context.Background(), "test")
		assert.Nil(t, store)
		assert.EqualError(t, err, "missing region")
	})

	t.Run("missing project-id", func(t *testing.T) {
		builder := ScalewayStoreBuilder{Region: "fr-par"}
		store, err := builder.Build(context.Background(), "test")
		assert.Nil(t, store)
		assert.EqualError(t, err, "missing project-id")
	})
}

func TestScalewayStore_Get(t *testing.T) {
	ctx := context.Background()
	region := scw.Region("fr-par")
	projectID := "test-project"

	t.Run("successful get", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("AccessSecretVersionByPath", &secret.AccessSecretVersionByPathRequest{
			Region:     region,
			SecretPath: "/",
			SecretName: "test-key",
			Revision:   "latest_enabled",
			ProjectID:  projectID,
		}).Return(&secret.AccessSecretVersionResponse{
			Data: []byte("test-value"),
		}, nil)

		value, err := store.Get(ctx, "test-key")
		assert.NoError(t, err)
		assert.Equal(t, "test-value", value)
		mockClient.AssertExpectations(t)
	})

	t.Run("key not found", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("AccessSecretVersionByPath", &secret.AccessSecretVersionByPathRequest{
			Region:     region,
			SecretPath: "/",
			SecretName: "test-key",
			Revision:   "latest_enabled",
			ProjectID:  projectID,
		}).Return(nil, &scw.ResourceNotFoundError{Resource: "secret"})

		value, err := store.Get(ctx, "test-key")
		assert.Equal(t, "", value)
		assert.Equal(t, ErrKeyNotFound, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("AccessSecretVersionByPath", &secret.AccessSecretVersionByPathRequest{
			Region:     region,
			SecretPath: "/",
			SecretName: "test-key",
			Revision:   "latest_enabled",
			ProjectID:  projectID,
		}).Return(nil, errors.New("api error"))

		value, err := store.Get(ctx, "test-key")
		assert.Equal(t, "", value)
		assert.EqualError(t, err, "access scaleway secret: api error")
		mockClient.AssertExpectations(t)
	})
}

func TestScalewayStore_Set(t *testing.T) {
	ctx := context.Background()
	region := scw.Region("fr-par")
	projectID := "test-project"

	t.Run("create new secret", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		// Secret does not exist
		mockClient.On("ListSecrets", &secret.ListSecretsRequest{
			Region:    region,
			ProjectID: &projectID,
			Name:      strPtr("test-key"),
		}).Return(&secret.ListSecretsResponse{
			Secrets:    []*secret.Secret{},
			TotalCount: 0,
		}, nil)

		mockClient.On("CreateSecret", &secret.CreateSecretRequest{
			Region:    region,
			ProjectID: projectID,
			Name:      "test-key",
			Type:      secret.SecretTypeOpaque,
		}).Return(&secret.Secret{ID: "secret-id-123"}, nil)

		disablePrevious := true
		mockClient.On("CreateSecretVersion", &secret.CreateSecretVersionRequest{
			Region:          region,
			SecretID:        "secret-id-123",
			Data:            []byte("test-value"),
			DisablePrevious: &disablePrevious,
		}).Return(&secret.SecretVersion{}, nil)

		err := store.Set(ctx, "test-key", "test-value")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("update existing secret", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		// Secret exists
		mockClient.On("ListSecrets", &secret.ListSecretsRequest{
			Region:    region,
			ProjectID: &projectID,
			Name:      strPtr("test-key"),
		}).Return(&secret.ListSecretsResponse{
			Secrets:    []*secret.Secret{{ID: "existing-id", Name: "test-key"}},
			TotalCount: 1,
		}, nil)

		disablePrevious := true
		mockClient.On("CreateSecretVersion", &secret.CreateSecretVersionRequest{
			Region:          region,
			SecretID:        "existing-id",
			Data:            []byte("new-value"),
			DisablePrevious: &disablePrevious,
		}).Return(&secret.SecretVersion{}, nil)

		err := store.Set(ctx, "test-key", "new-value")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("list error", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("ListSecrets", mock.Anything).Return(nil, errors.New("list error"))

		err := store.Set(ctx, "test-key", "test-value")
		assert.EqualError(t, err, "list scaleway secrets: list error")
		mockClient.AssertExpectations(t)
	})

	t.Run("create secret error", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("ListSecrets", mock.Anything).Return(&secret.ListSecretsResponse{
			Secrets:    []*secret.Secret{},
			TotalCount: 0,
		}, nil)

		mockClient.On("CreateSecret", mock.Anything).Return(nil, errors.New("create error"))

		err := store.Set(ctx, "test-key", "test-value")
		assert.EqualError(t, err, "create scaleway secret: create error")
		mockClient.AssertExpectations(t)
	})

	t.Run("create version error", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("ListSecrets", mock.Anything).Return(&secret.ListSecretsResponse{
			Secrets:    []*secret.Secret{{ID: "existing-id", Name: "test-key"}},
			TotalCount: 1,
		}, nil)

		mockClient.On("CreateSecretVersion", mock.Anything).Return(nil, errors.New("version error"))

		err := store.Set(ctx, "test-key", "test-value")
		assert.EqualError(t, err, "create scaleway secret version: version error")
		mockClient.AssertExpectations(t)
	})
}

func TestScalewayStore_Delete(t *testing.T) {
	ctx := context.Background()
	region := scw.Region("fr-par")
	projectID := "test-project"

	t.Run("successful delete", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("ListSecrets", &secret.ListSecretsRequest{
			Region:    region,
			ProjectID: &projectID,
			Name:      strPtr("test-key"),
		}).Return(&secret.ListSecretsResponse{
			Secrets:    []*secret.Secret{{ID: "secret-id", Name: "test-key"}},
			TotalCount: 1,
		}, nil)

		mockClient.On("DeleteSecret", &secret.DeleteSecretRequest{
			Region:   region,
			SecretID: "secret-id",
		}).Return(nil)

		err := store.Delete(ctx, "test-key")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("secret not found (not an error)", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("ListSecrets", mock.Anything).Return(&secret.ListSecretsResponse{
			Secrets:    []*secret.Secret{},
			TotalCount: 0,
		}, nil)

		err := store.Delete(ctx, "test-key")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("list error", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("ListSecrets", mock.Anything).Return(nil, errors.New("list error"))

		err := store.Delete(ctx, "test-key")
		assert.EqualError(t, err, "list scaleway secrets: list error")
		mockClient.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		mockClient := new(MockScalewaySecretClient)
		store := ScalewayStore{client: mockClient, region: region, projectID: projectID}

		mockClient.On("ListSecrets", mock.Anything).Return(&secret.ListSecretsResponse{
			Secrets:    []*secret.Secret{{ID: "secret-id", Name: "test-key"}},
			TotalCount: 1,
		}, nil)

		mockClient.On("DeleteSecret", &secret.DeleteSecretRequest{
			Region:   region,
			SecretID: "secret-id",
		}).Return(errors.New("delete error"))

		err := store.Delete(ctx, "test-key")
		assert.EqualError(t, err, "delete scaleway secret: delete error")
		mockClient.AssertExpectations(t)
	})
}

func strPtr(s string) *string {
	return &s
}
