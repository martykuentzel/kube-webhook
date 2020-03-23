// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package mutate

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/MartyKuentzel/kube-webhook/pkg/vault"
	log "github.com/sirupsen/logrus"
	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Mutate receives AdmissionReview and responds with mutated AdmissionReview
func mutate(ctx context.Context, body []byte, vault vault.VaultClient) ([]byte, error) {

	log.Debugf("Received Request: %s\n", string(body))

	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		log.Errorf("Unmarshaling request failed with %v", err)
		return nil, err
	}

	responseBody := []byte{}
	ar := admReview.Request
	if ar == nil {
		return responseBody, errors.New("AdmissionReview.Request is empty")
	}

	var secret *corev1.Secret
	if err := json.Unmarshal(ar.Object.Raw, &secret); err != nil {
		log.Errorf("Unable unmarshal secret json object %v", err)
		return nil, err
	}

	patch := patchSecrets(ctx, secret, vault)
	JSONPatch, err := json.Marshal(patch)
	if err != nil {
		log.Errorf("Cannot parse secret patch []map into Json: %v", err)
		return nil, err
	}

	resp := admResponse(JSONPatch, ar.UID)
	admReview.Response = &resp
	responseBody, err = json.Marshal(admReview)
	if err != nil {
		log.Errorf("Cannot parse admReview []map into Json: %v", err)
		return nil, err
	}

	log.Debugf("resp: %s\n", string(responseBody))
	return responseBody, nil
}

// build Response for Admission Review Response
func admResponse(JSONPatch []byte, UID types.UID) v1beta1.AdmissionResponse {

	log.Debug("Creating Admission Response")
	pT := v1beta1.PatchTypeJSONPatch
	resp := v1beta1.AdmissionResponse{
		Allowed:          true,
		UID:              UID,
		PatchType:        &pT,
		Patch:            JSONPatch,
		AuditAnnotations: map[string]string{"kube-secHook": "mutated"},
		Result:           &metav1.Status{Status: "Success"},
	}
	log.Debug("Admission Response successfully created")
	return resp
}
