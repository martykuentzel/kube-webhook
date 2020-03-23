package mutate

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestMutateJSON(t *testing.T) {

	rawJSON := `{
	"kind":"AdmissionReview",
	"apiVersion":"admission.k8s.io/v1beta1",
	"request":{
	   "uid":"4f3ccf94-6c5c-11ea-86d0-4201ac100016",
	   "kind":{
		  "group":"",
		  "version":"v1",
		  "kind":"Secret"
	   },
	   "resource":{
		  "group":"",
		  "version":"v1",
		  "resource":"secrets"
	   },
	   "namespace":"default",
	   "operation":"CREATE",
	   "userInfo":{
		  "username":"bob.baumeister@gmail.de",
		  "groups":[
			 "system:authenticated"
		  ],
		  "extra":{
			 "user-assertion.cloud.google.com":[
				"Ajefiwfkwrnfkjrnjf6/sjdnsdnjskdjkflrelkgnlwtngnwtkleglreg53r34o="
			 ]
		  }
	   },
	   "object":{
		  "kind":"Secret",
		  "apiVersion":"v1",
		  "metadata":{
			 "name":"mutate-test",
			 "namespace":"default",
			 "creationTimestamp":null,
			 "annotations":{}
		  },
		  "data":{
			"mutate1": "c2VjSG9vazp0aGlzIGlzIHRoZSBmYWtlIHNlY3JldA==",
			"mutate": "c2VjSG9vazpwcm9qZWN0cy83NzYyNDE2ODAzNDAvc2VjcmV0cy90ZXN0ZXIvdmVyc2lvbnMvbGF0ZXN0",
			"mutate3": "c2VjSG9vazpwcm9qZWN0cy9wbGF5Z3JvdW5kLW1hcnR5LWt1ZW50emVsL3NlY3JldHMvYmxhYmxhL3ZlcnNpb25zL2xhdGVzdA==",
			"dnt_mutate": "ZG8gbm90IG11dGF0ZSB0aGlzIGtleQ=="
		  },
		  "type":"Opaque"
	   },
	   "oldObject":null,
	   "dryRun":false
	}
 }`
	ctx := context.Background()
	m := new(mockSecretManager)
	retrievedSecret := []byte("blabla")

	m.On("GetSecret", ctx, mock.AnythingOfType("string")).Return(retrievedSecret, nil).Times(3)

	response, err := mutate(ctx, []byte(rawJSON), m)
	if err != nil {
		t.Errorf("failed to mutate AdmissionRequest %s with error %s", string(response), err)
	}

	r := v1beta1.AdmissionReview{}
	err = json.Unmarshal(response, &r)
	assert.NoError(t, err, "failed to unmarshal with error %s", err)

	rr := r.Response
	actual := string(rr.Patch)
	//order of the three can be different, which does not matter
	expected1 := `[{"op":"replace","path":"/data/mutate","value":"YmxhYmxh"},{"op":"replace","path":"/data/mutate3","value":"YmxhYmxh"},{"op":"replace","path":"/data/mutate1","value":"YmxhYmxh"}]`
	expected2 := `[{"op":"replace","path":"/data/mutate1","value":"YmxhYmxh"},{"op":"replace","path":"/data/mutate","value":"YmxhYmxh"},{"op":"replace","path":"/data/mutate3","value":"YmxhYmxh"}]`
	expected3 := `[{"op":"replace","path":"/data/mutate3","value":"YmxhYmxh"},{"op":"replace","path":"/data/mutate1","value":"YmxhYmxh"},{"op":"replace","path":"/data/mutate","value":"YmxhYmxh"}]`

	if actual != expected1 && actual != expected2 && actual != expected3 {
		t.Errorf("Mutating JSON test failed")
	}
	assert.Contains(t, rr.AuditAnnotations, "kube-secHook")
	m.AssertExpectations(t)

}

func TestAdmResponse(t *testing.T) {

	assert := assert.New(t)

	var mockUID types.UID
	mockUID = "ed195e41-6c58-11ea-94e0-4201ac100014"

	var mockPatch []map[string]string
	mockPatch = []map[string]string(
		[]map[string]string{
			map[string]string{"op": "replace", "path": "/data/mutate", "value": "YmxhYmxh"},
			map[string]string{"op": "replace", "path": "/data/mutate1", "value": "YmxhYmxh"},
			map[string]string{"op": "replace", "path": "/data/mutate3", "value": "YmxhYmxh"},
			map[string]string{"op": "replace", "path": "/data/mutate4", "value": "YmxhYmxh"},
		},
	)
	mockJSONPatch, err := json.Marshal(mockPatch)
	if err != nil {
		t.Errorf("Cannot parse secret patch []map into Json: %v", err)
	}

	actual := admResponse(mockJSONPatch, mockUID)

	pT := v1beta1.PatchTypeJSONPatch
	expected := v1beta1.AdmissionResponse{
		Allowed:          true,
		UID:              mockUID,
		PatchType:        &pT,
		Patch:            mockJSONPatch,
		AuditAnnotations: map[string]string{"kube-secHook": "mutated"},
		Result:           &metav1.Status{Status: "Success"},
	}

	assert.Equal(expected, actual, "The two responses should be the same.")
}
