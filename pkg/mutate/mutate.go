// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package mutate

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	crypto "github.com/MartyKuentzel/kube-webhook/pkg/crypto"
	log "github.com/sirupsen/logrus"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Mutate mutating Admission Request
func Mutate(ctx context.Context, body []byte) ([]byte, error) {

	log.Debugf("Received Request: %s\n", string(body))

	var err error
	var secret *corev1.Secret

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err = json.Unmarshal(body, &admReview); err != nil {
		log.Errorf("Unmarshaling request failed with %v", err)
		return nil, err
	}

	responseBody := []byte{}
	ar := admReview.Request
	if ar == nil {
		return responseBody, errors.New("AdmissionReview.Request is empty")
	}

	// get the Secret object and unmarshal it into its struct
	if err := json.Unmarshal(ar.Object.Raw, &secret); err != nil {
		log.Errorf("Unable unmarshal secret json object %v", err)
		return nil, err
	}

	p := patchSecrets(ctx, secret)
	resp, err := responseCreator(p, ar.UID)
	if err != nil {
		log.Errorf("Creation of response failed")
		return nil, err
	}

	admReview.Response = &resp
	// back into JSON so we can return the finished AdmissionReview w/ Response directly
	// w/o needing to convert things in the http handler
	responseBody, err = json.Marshal(admReview)
	if err != nil {
		log.Errorf("Cannot parse admReview []map into Json: %v", err)
		return nil, err
	}

	log.Debugf("resp: %s\n", string(responseBody))
	return responseBody, nil
}

// loop through secret values, check for "secman:" prefix and create map of patches
func patchSecrets(ctx context.Context, secret *corev1.Secret) []map[string]string {

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

// build Response for Admission Review Response
func responseCreator(secretPatch []map[string]string, UID types.UID) (v1beta1.AdmissionResponse, error) {

	log.Debug("Creating Response")
	var err error

	resp := v1beta1.AdmissionResponse{}
	resp.Allowed = true
	resp.UID = UID
	pT := v1beta1.PatchTypeJSONPatch
	resp.PatchType = &pT
	resp.Patch, err = json.Marshal(secretPatch)
	if err != nil {
		log.Errorf("Cannot parse secret patch []map into Json: %v", err)
		return resp, err
	}
	resp.AuditAnnotations = map[string]string{
		"kube-secman": "mutated",
	}
	resp.Result = &metav1.Status{
		Status: "Success",
	}
	log.Debug("Response successfully created")
	return resp, nil
}

func trimKey(secretValueRaw []byte) string {

	secManKeyRaw := strings.TrimPrefix(string(secretValueRaw), "secman:")
	secManKey := strings.TrimRight(secManKeyRaw, "\n")

	return secManKey
}
