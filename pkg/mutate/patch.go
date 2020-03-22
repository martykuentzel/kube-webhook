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

func patchSecrets(ctx context.Context, secret *corev1.Secret, v vault.VaultClient) []map[string]string {

	secManEntries := findAllSecManEntries(secret.Data)
	patch := replaceSecManVals(ctx, v, secManEntries)
	log.Debugf("Created following patch: %v", patch)
	return patch
}

func findAllSecManEntries(secretContent map[string][]byte) map[string]string {

	ss := map[string]string{}
	for k, v := range secretContent {
		if hasSecManPrefix(string(v)) {
			log.Debugf("key: %s, value: %s found", k, v)
			ss[k] = string(v)
		}
	}
	return ss
}

func hasSecManPrefix(s string) bool {
	s1 := strings.TrimSpace(s)
	if strings.HasPrefix(s1, "secman:") {
		return true
	}
	return false
}

// the actual mutation is done by a string in JSONPatch style
func replaceSecManVals(ctx context.Context, vault vault.VaultClient, secManEntries map[string]string) []map[string]string {

	pp := []map[string]string{}

	for k, v := range secManEntries {
		secManAddr := removeSecManPrefix(v)
		log.Infof("Retrieving Secret for secManAddr '%s'", secManAddr)
		retrievedSecret, err := vault.GetSecret(ctx, secManAddr)

		patch := map[string]string{}
		if err != nil {
			log.Errorf("Because secret cannot be retrieved from SecretManager the secret `%s` will not be muatated", v)
			patch = createPatch(k, []byte(v))
			pp = append(pp, patch)
		} else {
			log.Debugf("Secret with secManAddr %s could be successfully retrieved from Secret Manager", secManAddr)
			patch = createPatch(k, retrievedSecret)
			pp = append(pp, patch)
		}
	}
	return pp
}

func removeSecManPrefix(secretValueRaw string) string {
	s1 := strings.TrimSpace(secretValueRaw)
	s2 := strings.TrimPrefix(s1, "secman:")
	secManAddr := strings.TrimSpace(s2)
	return secManAddr
}

func createPatch(secretKey string, secretValue []byte) map[string]string {

	patch := map[string]string{
		"op":    "replace",
		"path":  fmt.Sprintf("/data/%s", secretKey),
		"value": base64.StdEncoding.EncodeToString(secretValue),
	}
	return patch
}
