package backend

import (
	"context"
	"fmt"

	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	registerBuilder("scaleway", func() Builder { return new(ScalewayStoreBuilder) })
}

// ScalewayStoreBuilder implements the Builder interface for Scaleway Secret Manager.
type ScalewayStoreBuilder struct {
	Region    string `toml:"region"`
	ProjectID string `toml:"project-id"`
	// AccessKey is the Scaleway access key (SCW_ACCESS_KEY). Optional if env/config is set.
	AccessKey string `toml:"access-key,omitempty"`
	// SecretKey is the Scaleway secret key (SCW_SECRET_KEY). Optional if env/config is set.
	SecretKey string `toml:"secret-key,omitempty"`
}

// Build returns a new Scaleway Secret Manager store.
func (sb *ScalewayStoreBuilder) Build(ctx context.Context, name string) (Store, error) {
	if sb.Region == "" {
		return nil, fmt.Errorf("missing region")
	}
	if sb.ProjectID == "" {
		return nil, fmt.Errorf("missing project-id")
	}
	return NewScalewayStore(sb)
}

// ScalewaySecretClient defines the interface for Scaleway Secret Manager operations.
type ScalewaySecretClient interface {
	AccessSecretVersionByPath(req *secret.AccessSecretVersionByPathRequest, opts ...scw.RequestOption) (*secret.AccessSecretVersionResponse, error)
	CreateSecret(req *secret.CreateSecretRequest, opts ...scw.RequestOption) (*secret.Secret, error)
	CreateSecretVersion(req *secret.CreateSecretVersionRequest, opts ...scw.RequestOption) (*secret.SecretVersion, error)
	ListSecrets(req *secret.ListSecretsRequest, opts ...scw.RequestOption) (*secret.ListSecretsResponse, error)
	DeleteSecret(req *secret.DeleteSecretRequest, opts ...scw.RequestOption) error
}

// ScalewayStore represents a Scaleway Secret Manager store.
type ScalewayStore struct {
	client    ScalewaySecretClient
	region    scw.Region
	projectID string
}

// NewScalewayStore creates a new Scaleway Secret Manager Store.
func NewScalewayStore(builder *ScalewayStoreBuilder) (*ScalewayStore, error) {
	var clientOpts []scw.ClientOption

	// Load from environment/config file first, then override with explicit values
	clientOpts = append(clientOpts, scw.WithEnv())
	clientOpts = append(clientOpts, scw.WithDefaultRegion(scw.Region(builder.Region)))
	clientOpts = append(clientOpts, scw.WithDefaultProjectID(builder.ProjectID))

	if builder.AccessKey != "" && builder.SecretKey != "" {
		clientOpts = append(clientOpts, scw.WithAuth(builder.AccessKey, builder.SecretKey))
	}

	client, err := scw.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("create scaleway client: %w", err)
	}

	api := secret.NewAPI(client)
	return &ScalewayStore{
		client:    api,
		region:    scw.Region(builder.Region),
		projectID: builder.ProjectID,
	}, nil
}

// Get implements the Store.Get method.
func (s *ScalewayStore) Get(ctx context.Context, key string) (string, error) {
	resp, err := s.client.AccessSecretVersionByPath(&secret.AccessSecretVersionByPathRequest{
		Region:     s.region,
		SecretPath: "/",
		SecretName: key,
		Revision:   "latest_enabled",
		ProjectID:  s.projectID,
	})
	if err != nil {
		if isScalewayNotFound(err) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("access scaleway secret: %w", err)
	}

	return string(resp.Data), nil
}

// Set implements the Store.Set method.
func (s *ScalewayStore) Set(ctx context.Context, key, value string) error {
	// Try to find the secret by name
	secretID, err := s.findSecretIDByName(key)
	if err != nil {
		return err
	}

	if secretID == "" {
		// Secret does not exist, create it with initial version
		created, err := s.client.CreateSecret(&secret.CreateSecretRequest{
			Region:    s.region,
			ProjectID: s.projectID,
			Name:      key,
			Type:      secret.SecretTypeOpaque,
		})
		if err != nil {
			return fmt.Errorf("create scaleway secret: %w", err)
		}
		secretID = created.ID
	}

	// Create a new version with the value
	disablePrevious := true
	_, err = s.client.CreateSecretVersion(&secret.CreateSecretVersionRequest{
		Region:          s.region,
		SecretID:        secretID,
		Data:            []byte(value),
		DisablePrevious: &disablePrevious,
	})
	if err != nil {
		return fmt.Errorf("create scaleway secret version: %w", err)
	}

	return nil
}

// Delete implements the Store.Delete method.
func (s *ScalewayStore) Delete(ctx context.Context, key string) error {
	secretID, err := s.findSecretIDByName(key)
	if err != nil {
		return err
	}
	if secretID == "" {
		return nil // Already gone, not an error
	}

	err = s.client.DeleteSecret(&secret.DeleteSecretRequest{
		Region:   s.region,
		SecretID: secretID,
	})
	if err != nil {
		if isScalewayNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete scaleway secret: %w", err)
	}

	return nil
}

// findSecretIDByName looks up a secret by name and returns its ID, or "" if not found.
func (s *ScalewayStore) findSecretIDByName(name string) (string, error) {
	resp, err := s.client.ListSecrets(&secret.ListSecretsRequest{
		Region:    s.region,
		ProjectID: &s.projectID,
		Name:      &name,
	})
	if err != nil {
		return "", fmt.Errorf("list scaleway secrets: %w", err)
	}

	for _, sec := range resp.Secrets {
		if sec.Name == name {
			return sec.ID, nil
		}
	}

	return "", nil
}

func isScalewayNotFound(err error) bool {
	notFoundErr, ok := err.(*scw.ResourceNotFoundError)
	_ = notFoundErr
	return ok
}
