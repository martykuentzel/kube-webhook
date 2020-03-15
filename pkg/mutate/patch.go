package mutate

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	crypto "github.com/MartyKuentzel/kube-webhook/pkg/crypto"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// loop through secret values, check for "secman:" prefix and create map of patches
func secretPatch(ctx context.Context, secret *corev1.Secret) []map[string]string {

	p := []map[string]string{}
	patch := map[string]string{}
	for k, v := range secret.Data {
		log.Debugf("key: %s, value: %s found", k, v)

		if strings.HasPrefix(string(v), "secman:") {
			log.Infof("Mutating '%s/%s/%s'.", secret.Namespace, secret.Name, k)
			patch = replaceSecManKey(ctx, k, v)
			p = append(p, patch)
		}

	}
	log.Debugf("Created following patch: %v", p)
	return p
}

// the actual mutation is done by a string in JSONPatch style
func replaceSecManKey(ctx context.Context, secretKey string, secretValueRaw []byte) map[string]string {

	secManKey := trimKey(secretValueRaw)
	log.Infof("Retrieving Secret for secManKey '%s'", secManKey)
	retrievedSecret, err := crypto.GetSecret(ctx, secManKey)

	patch := map[string]string{}
	if err != nil {
		log.Errorf("Because secret cannot be retrieved from SecretManager the secret `%s` will not be muatated", secretKey)
		patch = map[string]string{
			"op":    "replace",
			"path":  fmt.Sprintf("/data/%s", secretKey),
			"value": base64.StdEncoding.EncodeToString(secretValueRaw),
		}
	} else {
		log.Debugf("Secret with secManKey %s could be successfully retrieved from Secret Manager", secManKey)
		patch = map[string]string{
			"op":    "replace",
			"path":  fmt.Sprintf("/data/%s", secretKey),
			"value": base64.StdEncoding.EncodeToString([]byte(retrievedSecret)),
		}
	}
	return patch
}

func trimKey(secretValueRaw []byte) string {

	secManKeyRaw := strings.TrimPrefix(string(secretValueRaw), "secman:")
	secManKey := strings.TrimRight(secManKeyRaw, "\n")
	return secManKey
}
