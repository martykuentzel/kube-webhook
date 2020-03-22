package vault

import (
	"context"
	"errors"

	secretmanagerSDK "cloud.google.com/go/secretmanager/apiv1beta1"
	log "github.com/sirupsen/logrus"
	secretspb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

// VaultClient bladsds
type VaultClient interface {
	GetSecret(context.Context, string) ([]byte, error)
}

type secretManager struct {
	client *secretmanagerSDK.Client
}

// GetSecret retrieves the content of an SecretManager Key
func (sM secretManager) GetSecret(ctx context.Context, secHookAddr string) ([]byte, error) {

	// Build the request.
	accessRequest := &secretspb.AccessSecretVersionRequest{
		Name: secHookAddr,
	}

	// Call the API.
	result, err := sM.client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		log.Errorf("failed to access secret version: %v", err)
		return nil, err
	}

	return result.Payload.Data, err
}

// New initializes VaultClient
func New(ctx context.Context) (VaultClient, error) {

	// Create the client.
	client, err := secretmanagerSDK.NewClient(ctx)
	if err != nil {
		log.Errorf("Failed to create secretmanager client: %v", err)
		return nil, errors.New("Failed to initialize vault client.")
	}

	return &secretManager{
		client: client,
	}, err
}
