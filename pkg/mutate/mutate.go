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

	"github.com/MartyKuentzel/kube-webhook/pkg/crypto"
	log "github.com/sirupsen/logrus"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Mutate mutates
func Mutate(ctx context.Context, body []byte) ([]byte, error) {

	log.Debugf("recv: %s\n", string(body))

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var secret *corev1.Secret

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1beta1.AdmissionResponse{}

	if ar == nil {
		return responseBody, errors.New("admissionReview.Request is empty")
	}

	// get the Secret object and unmarshal it into its struct
	if err := json.Unmarshal(ar.Object.Raw, &secret); err != nil {
		return nil, fmt.Errorf("unable unmarshal secret json object %v", err)
	}
	// set response options
	resp.Allowed = true
	resp.UID = ar.UID
	pT := v1beta1.PatchTypeJSONPatch
	resp.PatchType = &pT

	// add some audit annotations
	resp.AuditAnnotations = map[string]string{
		"kube-secman": "mutated",
	}

	// the actual mutation is done by a string in JSONPatch style, i.e. we don't _actually_ modify the object, but
	// tell K8S how it should modifiy it
	p := []map[string]string{}
	patch := map[string]string{}
	for k, v := range secret.Data {
		log.Debugf("key: %s, value: *** found.", k)
		if strings.HasPrefix(string(v), "secman") {

			log.Infof("Mutating ns/sercet/key `%s/%s/%s`. ", secret.Namespace, secret.Name, k)

			retrievedSecret, err := crypto.GetSecret(ctx, secret.Name)

			if err != nil {
				log.Errorf("Cannot retrieve secret from SecretManager: %v", err)
			} else {
				patch = map[string]string{
					"op":    "replace",
					"path":  fmt.Sprintf("/data/%s", k),
					"value": base64.StdEncoding.EncodeToString([]byte(retrievedSecret)),
				}
			}

			p = append(p, patch)

		}
		//for deeper debugging :), don't use it in prod
		log.Debugf("patch: %v", p)
	}
	// parse the []map into JSON
	resp.Patch, err = json.Marshal(p)

	// Success, of course ;)
	resp.Result = &metav1.Status{
		Status: "Success",
	}

	admReview.Response = &resp
	// back into JSON so we can return the finished AdmissionReview w/ Response directly
	// w/o needing to convert things in the http handler
	responseBody, err = json.Marshal(admReview)
	if err != nil {
		return nil, err
	}

	log.Printf("resp: %s\n", string(responseBody))
	return responseBody, nil
}
