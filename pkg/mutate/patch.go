package mutate

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/MartyKuentzel/kube-webhook/pkg/vault"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// the actual mutation is done by a string in JSONPatch style
func patchSecrets(ctx context.Context, secret *corev1.Secret, v vault.VaultClient) []map[string]string {

	secHookEntries := findAllSecHookEntries(secret.Data)
	patch := replaceSecHookVals(ctx, v, secHookEntries)
	log.Debugf("Created following patch: %v", patch)
	return patch
}

func findAllSecHookEntries(secretContent map[string][]byte) map[string]string {

	ss := map[string]string{}
	for k, v := range secretContent {
		if hasSecHookPrefix(string(v)) {
			log.Debugf("key: %s, value: %s found", k, v)
			ss[k] = string(v)
		}
	}
	return ss
}

func hasSecHookPrefix(s string) bool {
	s1 := strings.TrimSpace(s)
	if strings.HasPrefix(s1, "secHook:") {
		return true
	}
	return false
}

// Placeholder called 'secHookAddr' will be replaced with secret, if the value points to an actual key
func replaceSecHookVals(ctx context.Context, vault vault.VaultClient, secHookEntries map[string]string) []map[string]string {

	pp := []map[string]string{}

	for k, v := range secHookEntries {
		secHookAddr := removeSecHookPrefix(v)
		log.Infof("Retrieving Secret for Placeholder: '%s'", secHookAddr)
		retrievedSecret, err := vault.GetSecret(ctx, secHookAddr)

		patch := map[string]string{}
		if err != nil {
			log.Errorf("Because Secret cannot be retrieved from SecretManager the Placeholder `%s` will not be muatated", v)
			patch = createPatch(k, []byte(v))
			pp = append(pp, patch)
		} else {
			log.Debugf("Secret with Placeholder %s could be successfully retrieved from Secret Manager", secHookAddr)
			patch = createPatch(k, retrievedSecret)
			pp = append(pp, patch)
		}
	}
	return pp
}

func removeSecHookPrefix(secretValueRaw string) string {
	s1 := strings.TrimSpace(secretValueRaw)
	s2 := strings.TrimPrefix(s1, "secHook:")
	secHookAddr := strings.TrimSpace(s2)
	return secHookAddr
}

func createPatch(secretKey string, secretValue []byte) map[string]string {
	patch := map[string]string{
		"op":    "replace",
		"path":  fmt.Sprintf("/data/%s", secretKey),
		"value": base64.StdEncoding.EncodeToString(secretValue),
	}
	return patch
}
