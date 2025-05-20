package backend

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

func init() {
	registerBuilder("aws", func() Builder { return new(AWSStoreBuilder) })
}

// AWSStoreBuilder implements the Builder interface for AWS Secrets Manager.
type AWSStoreBuilder struct {
	Region  string `toml:"region"`
	Profile string `toml:"profile,omitempty"`
}

// Build returns a new AWS Secrets Manager store.
func (ab *AWSStoreBuilder) Build(ctx context.Context, name string) (Store, error) {
	if ab.Region == "" {
		return nil, fmt.Errorf("missing region")
	}
	return NewAWSStore(ctx, ab)
}

// AWSSecretsManagerClient defines the interface for AWS Secrets Manager operations.
type AWSSecretsManagerClient interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	CreateSecret(context.Context, *secretsmanager.CreateSecretInput, ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
	PutSecretValue(context.Context, *secretsmanager.PutSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error)
	DeleteSecret(context.Context, *secretsmanager.DeleteSecretInput, ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error)
}

// AWSStore represents an AWS Secrets Manager store.
type AWSStore struct {
	client AWSSecretsManagerClient
	region string
}

// NewAWSStore creates a new AWS Secrets Manager Store.
func NewAWSStore(ctx context.Context, builder *AWSStoreBuilder) (*AWSStore, error) {
	var configOpts []func(*config.LoadOptions) error
	
	// Add region
	configOpts = append(configOpts, config.WithRegion(builder.Region))
	
	// Add profile if specified
	if builder.Profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(builder.Profile))
	}
	
	// Note: AWS SSO is automatically handled by the AWS SDK when using a profile
	// that is configured for SSO. No additional configuration is needed here.
	
	// Load the AWS configuration with all options
	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	
	client := secretsmanager.NewFromConfig(cfg)
	return &AWSStore{
		client: client,
		region: builder.Region,
	}, nil
}

// Get implements the Store.Get method.
func (a *AWSStore) Get(ctx context.Context, key string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}
	
	result, err := a.client.GetSecretValue(ctx, input)
	if err != nil {
		// Check if the error is a ResourceNotFoundException
		var rnfe *types.ResourceNotFoundException
		if errors.As(err, &rnfe) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("get aws secret: %w", err)
	}
	
	// Return the secret string
	if result.SecretString != nil {
		return *result.SecretString, nil
	}
	
	// If the secret is binary, we don't support it
	return "", fmt.Errorf("aws secret is binary, not string")
}

// Set implements the Store.Set method.
func (a *AWSStore) Set(ctx context.Context, key, value string) error {
	// Try to get the secret first to check if it exists
	_, err := a.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	})
	
	if err != nil {
		// If the secret doesn't exist, create it
		var rnfe *types.ResourceNotFoundException
		if errors.As(err, &rnfe) {
			_, err = a.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
				Name:         aws.String(key),
				SecretString: aws.String(value),
			})
			if err != nil {
				return fmt.Errorf("create aws secret: %w", err)
			}
			return nil
		}
		return fmt.Errorf("check aws secret existence: %w", err)
	}
	
	// If the secret exists, update it
	_, err = a.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(key),
		SecretString: aws.String(value),
	})
	if err != nil {
		return fmt.Errorf("update aws secret: %w", err)
	}
	
	return nil
}

// Delete implements the Store.Delete method.
func (a *AWSStore) Delete(ctx context.Context, key string) error {
	_, err := a.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(key),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		// Check if the error is a ResourceNotFoundException
		var rnfe *types.ResourceNotFoundException
		if errors.As(err, &rnfe) {
			return nil // Already deleted, not an error
		}
		return fmt.Errorf("delete aws secret: %w", err)
	}
	
	return nil
}