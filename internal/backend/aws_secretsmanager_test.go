package backend

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAWSSecretsManagerClient is a mock implementation of AWSSecretsManagerClient
type MockAWSSecretsManagerClient struct {
	mock.Mock
}

func (m *MockAWSSecretsManagerClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

func (m *MockAWSSecretsManagerClient) CreateSecret(ctx context.Context, input *secretsmanager.CreateSecretInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secretsmanager.CreateSecretOutput), args.Error(1)
}

func (m *MockAWSSecretsManagerClient) PutSecretValue(ctx context.Context, input *secretsmanager.PutSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secretsmanager.PutSecretValueOutput), args.Error(1)
}

func (m *MockAWSSecretsManagerClient) DeleteSecret(ctx context.Context, input *secretsmanager.DeleteSecretInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secretsmanager.DeleteSecretOutput), args.Error(1)
}

func TestAWSStoreBuilder_Build(t *testing.T) {
	t.Run("missing region", func(t *testing.T) {
		builder := AWSStoreBuilder{}
		store, err := builder.Build(context.Background(), "test")
		assert.Nil(t, store)
		assert.EqualError(t, err, "missing region")
	})
}

func TestAWSStore_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		mockOutput := &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String("test-value"),
		}
		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(mockOutput, nil)

		value, err := store.Get(ctx, "test-key")
		assert.NoError(t, err)
		assert.Equal(t, "test-value", value)
		mockClient.AssertExpectations(t)
	})

	t.Run("key not found", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		// Create a ResourceNotFoundException
		rnfe := &types.ResourceNotFoundException{
			Message: aws.String("Secret not found"),
		}

		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(nil, rnfe)

		value, err := store.Get(ctx, "test-key")
		assert.Equal(t, "", value)
		assert.Equal(t, ErrKeyNotFound, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("binary secret not supported", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		// Simulate a binary secret by not setting SecretString
		mockOutput := &secretsmanager.GetSecretValueOutput{
			SecretBinary: []byte("binary-data"),
		}
		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(mockOutput, nil)

		value, err := store.Get(ctx, "test-key")
		assert.Equal(t, "", value)
		assert.EqualError(t, err, "aws secret is binary, not string")
		mockClient.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(nil, errors.New("other error"))

		value, err := store.Get(ctx, "test-key")
		assert.Equal(t, "", value)
		assert.EqualError(t, err, "get aws secret: other error")
		mockClient.AssertExpectations(t)
	})
}

func TestAWSStore_Set(t *testing.T) {
	ctx := context.Background()

	t.Run("create new secret", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		// Simulate secret not found error
		rnfe := &types.ResourceNotFoundException{
			Message: aws.String("Secret not found"),
		}

		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(nil, rnfe)

		mockClient.On("CreateSecret", ctx, &secretsmanager.CreateSecretInput{
			Name:         aws.String("test-key"),
			SecretString: aws.String("test-value"),
		}).Return(&secretsmanager.CreateSecretOutput{}, nil)

		err := store.Set(ctx, "test-key", "test-value")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("update existing secret", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		// Simulate existing secret
		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String("old-value"),
		}, nil)

		mockClient.On("PutSecretValue", ctx, &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String("test-key"),
			SecretString: aws.String("test-value"),
		}).Return(&secretsmanager.PutSecretValueOutput{}, nil)

		err := store.Set(ctx, "test-key", "test-value")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("create error", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		// Simulate secret not found error
		rnfe := &types.ResourceNotFoundException{
			Message: aws.String("Secret not found"),
		}

		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(nil, rnfe)

		mockClient.On("CreateSecret", ctx, &secretsmanager.CreateSecretInput{
			Name:         aws.String("test-key"),
			SecretString: aws.String("test-value"),
		}).Return(nil, errors.New("create error"))

		err := store.Set(ctx, "test-key", "test-value")
		assert.EqualError(t, err, "create aws secret: create error")
		mockClient.AssertExpectations(t)
	})

	t.Run("get error", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(nil, errors.New("get error"))

		err := store.Set(ctx, "test-key", "test-value")
		assert.EqualError(t, err, "check aws secret existence: get error")
		mockClient.AssertExpectations(t)
	})

	t.Run("update error", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		// Simulate existing secret
		mockClient.On("GetSecretValue", ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String("test-key"),
		}).Return(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String("old-value"),
		}, nil)

		mockClient.On("PutSecretValue", ctx, &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String("test-key"),
			SecretString: aws.String("test-value"),
		}).Return(nil, errors.New("update error"))

		err := store.Set(ctx, "test-key", "test-value")
		assert.EqualError(t, err, "update aws secret: update error")
		mockClient.AssertExpectations(t)
	})
}

func TestAWSStore_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		mockClient.On("DeleteSecret", ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String("test-key"),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		}).Return(&secretsmanager.DeleteSecretOutput{}, nil)

		err := store.Delete(ctx, "test-key")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("key not found (not an error)", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		// Simulate key not found error
		rnfe := &types.ResourceNotFoundException{
			Message: aws.String("Secret not found"),
		}

		mockClient.On("DeleteSecret", ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String("test-key"),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		}).Return(nil, rnfe)

		err := store.Delete(ctx, "test-key")
		assert.NoError(t, err) // Not found is not an error for delete
		mockClient.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		mockClient := new(MockAWSSecretsManagerClient)
		store := AWSStore{client: mockClient, region: "us-east-1"}

		mockClient.On("DeleteSecret", ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String("test-key"),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		}).Return(nil, errors.New("delete error"))

		err := store.Delete(ctx, "test-key")
		assert.EqualError(t, err, "delete aws secret: delete error")
		mockClient.AssertExpectations(t)
	})
}