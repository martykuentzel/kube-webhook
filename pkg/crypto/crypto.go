package crypto

import (
	"context"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	log "github.com/sirupsen/logrus"
	secretspb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

// GetSecret retrieves the content of an object
func GetSecret(ctx context.Context, key string) (encSecret []byte, err error) {

	// Create the client.
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Errorf("failed to create secretmanager client: %v", err)
		return nil, err
	}
	// Build the request.
	accessRequest := &secretspb.AccessSecretVersionRequest{
		Name: "projects/776241680340/secrets/tester/versions/1",
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		log.Errorf("failed to access secret version: %v", err)
		return nil, err
	}

	return result.Payload.Data, err
}
