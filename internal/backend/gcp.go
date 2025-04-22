package backend

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	registerBuilder("gcp", func() Builder { return new(GCPStoreBuilder) })
}

// GCPStoreBuilder implements the Builder interface for GCP Secret Manager.
type GCPStoreBuilder struct {
	ProjectID string `toml:"project-id"`
}

// Build returns a new OSStore store.
func (ob *GCPStoreBuilder) Build(ctx context.Context, name string) (Store, error) {
	if ob.ProjectID == "" {
		return nil, fmt.Errorf("missing project-id")
	}
	return NewGCPStore(ctx, ob.ProjectID)
}

type SecretManagerClient interface {
	AccessSecretVersion(context.Context, *secretmanagerpb.AccessSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	GetSecret(context.Context, *secretmanagerpb.GetSecretRequest, ...gax.CallOption) (*secretmanagerpb.Secret, error)
	CreateSecret(context.Context, *secretmanagerpb.CreateSecretRequest, ...gax.CallOption) (*secretmanagerpb.Secret, error)
	AddSecretVersion(context.Context, *secretmanagerpb.AddSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	ListSecretVersions(context.Context, *secretmanagerpb.ListSecretVersionsRequest, ...gax.CallOption) *secretmanager.SecretVersionIterator
	DisableSecretVersion(context.Context, *secretmanagerpb.DisableSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	DestroySecretVersion(context.Context, *secretmanagerpb.DestroySecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	DeleteSecret(context.Context, *secretmanagerpb.DeleteSecretRequest, ...gax.CallOption) error
}

type GCPStore struct {
	client SecretManagerClient

	projectID string
}

// NewGCPStore creates a new GCP Secret Manager Store.
func NewGCPStore(ctx context.Context, projectID string) (*GCPStore, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("setup client: %w", err)
	}
	return &GCPStore{
		client,
		projectID,
	}, nil
}

// Get implements the Store.Get method.
func (o *GCPStore) Get(ctx context.Context, k string) (string, error) {
	res, err := o.client.AccessSecretVersion(ctx,
		&secretmanagerpb.AccessSecretVersionRequest{
			Name: secretLatestVersion(o, k),
		})
	if err != nil {
		return "", fmt.Errorf("access gcp secret version: %w", err)
	}

	return string(res.Payload.GetData()), nil
}

// Set implements the Store.Set method
func (o *GCPStore) Set(ctx context.Context, k, v string) error {
	// Try getting the secret, if it already exists
	secret, err := o.client.GetSecret(ctx,
		&secretmanagerpb.GetSecretRequest{
			Name: secretName(o, k),
		})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			secret, err = o.client.CreateSecret(ctx,
				&secretmanagerpb.CreateSecretRequest{
					Parent:   fmt.Sprintf("projects/%s", o.projectID),
					SecretId: k,
					Secret: &secretmanagerpb.Secret{
						Replication: &secretmanagerpb.Replication{
							Replication: &secretmanagerpb.Replication_Automatic_{
								Automatic: &secretmanagerpb.Replication_Automatic{},
							},
						},
					},
				})
			if err != nil {
				return fmt.Errorf("create secret: %w", err)
			}
		} else {
			return fmt.Errorf("retrieve secret: %w", err)
		}
	}

	version, err := o.client.AddSecretVersion(ctx,
		&secretmanagerpb.AddSecretVersionRequest{
			Parent:  secret.GetName(),
			Payload: &secretmanagerpb.SecretPayload{Data: []byte(v)},
		})
	if err != nil {
		return fmt.Errorf("add secret version: %w", err)
	}

	// Cleanup old versions
	for v, err := range o.client.ListSecretVersions(ctx,
		&secretmanagerpb.ListSecretVersionsRequest{
			Parent: secret.GetName(),
		}).All() {
		if err != nil {
			continue
		}
		if v.GetName() != version.GetName() {
			switch v.State {
			case secretmanagerpb.SecretVersion_ENABLED:
				_, _ = o.client.DisableSecretVersion(ctx,
					&secretmanagerpb.DisableSecretVersionRequest{
						Name: v.GetName(),
					})
			case secretmanagerpb.SecretVersion_DISABLED:
				_, _ = o.client.DestroySecretVersion(ctx,
					&secretmanagerpb.DestroySecretVersionRequest{
						Name: v.GetName(),
					})
			}
		}
	}

	return nil
}

// Delete implements the Store.Delete method.
func (o *GCPStore) Delete(ctx context.Context, k string) error {
	return o.client.DeleteSecret(ctx,
		&secretmanagerpb.DeleteSecretRequest{
			Name: secretName(o, k),
		})
}

func secretName(store *GCPStore, k string) string {
	return fmt.Sprintf("projects/%s/secrets/%s", store.projectID, k)
}

func secretLatestVersion(store *GCPStore, k string) string {
	return fmt.Sprintf("%s/versions/latest", secretName(store, k))
}
