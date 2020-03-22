package mutate

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// func TestMutateJSON(t *testing.T) {
// 	rawJSON := `{
// 		"kind": "AdmissionReview",
// 		"apiVersion": "admission.k8s.io/v1beta1",
// 		"request": {
// 			"uid": "7f0b2891-916f-4ed6-b7cd-27bff1815a8c",
// 			"kind": {
// 				"group": "",
// 				"version": "v1",
// 				"kind": "Pod"
// 			},
// 			"resource": {
// 				"group": "",
// 				"version": "v1",
// 				"resource": "pods"
// 			},
// 			"requestKind": {
// 				"group": "",
// 				"version": "v1",
// 				"kind": "Pod"
// 			},
// 			"requestResource": {
// 				"group": "",
// 				"version": "v1",
// 				"resource": "pods"
// 			},
// 			"namespace": "yolo",
// 			"operation": "CREATE",
// 			"userInfo": {
// 				"username": "kubernetes-admin",
// 				"groups": [
// 					"system:masters",
// 					"system:authenticated"
// 				]
// 			},
// 			"object": {
// 				"kind": "Pod",
// 				"apiVersion": "v1",
// 				"metadata": {
// 					"name": "c7m",
// 					"namespace": "yolo",
// 					"creationTimestamp": null,
// 					"labels": {
// 						"name": "c7m"
// 					},
// 					"annotations": {
// 						"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"annotations\":{},\"labels\":{\"name\":\"c7m\"},\"name\":\"c7m\",\"namespace\":\"yolo\"},\"spec\":{\"containers\":[{\"args\":[\"-c\",\"trap \\\"killall sleep\\\" TERM; trap \\\"kill -9 sleep\\\" KILL; sleep infinity\"],\"command\":[\"/bin/bash\"],\"image\":\"centos:7\",\"name\":\"c7m\"}]}}\n"
// 					}
// 				},
// 				"spec": {
// 					"volumes": [
// 						{
// 							"name": "default-token-5z7xl",
// 							"secret": {
// 								"secretName": "default-token-5z7xl"
// 							}
// 						}
// 					],
// 					"containers": [
// 						{
// 							"name": "c7m",
// 							"image": "centos:7",
// 							"command": [
// 								"/bin/bash"
// 							],
// 							"args": [
// 								"-c",
// 								"trap \"killall sleep\" TERM; trap \"kill -9 sleep\" KILL; sleep infinity"
// 							],
// 							"resources": {},
// 							"volumeMounts": [
// 								{
// 									"name": "default-token-5z7xl",
// 									"readOnly": true,
// 									"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
// 								}
// 							],
// 							"terminationMessagePath": "/dev/termination-log",
// 							"terminationMessagePolicy": "File",
// 							"imagePullPolicy": "IfNotPresent"
// 						}
// 					],
// 					"restartPolicy": "Always",
// 					"terminationGracePeriodSeconds": 30,
// 					"dnsPolicy": "ClusterFirst",
// 					"serviceAccountName": "default",
// 					"serviceAccount": "default",
// 					"securityContext": {},
// 					"schedulerName": "default-scheduler",
// 					"tolerations": [
// 						{
// 							"key": "node.kubernetes.io/not-ready",
// 							"operator": "Exists",
// 							"effect": "NoExecute",
// 							"tolerationSeconds": 300
// 						},
// 						{
// 							"key": "node.kubernetes.io/unreachable",
// 							"operator": "Exists",
// 							"effect": "NoExecute",
// 							"tolerationSeconds": 300
// 						}
// 					],
// 					"priority": 0,
// 					"enableServiceLinks": true
// 				},
// 				"status": {}
// 			},
// 			"oldObject": null,
// 			"dryRun": false,
// 			"options": {
// 				"kind": "CreateOptions",
// 				"apiVersion": "meta.k8s.io/v1"
// 			}
// 		}
// 	}`
// 	response, err := Mutate([]byte(rawJSON))
// 	if err != nil {
// 		t.Errorf("failed to mutate AdmissionRequest %s with error %s", string(response), err)
// 	}

// 	r := v1beta1.AdmissionReview{}
// 	err = json.Unmarshal(response, &r)
// 	assert.NoError(t, err, "failed to unmarshal with error %s", err)

// 	rr := r.Response
// 	assert.Equal(t, `[{"op":"replace","path":"/spec/containers/0/image","value":"debian"}]`, string(rr.Patch))
// 	assert.Contains(t, rr.AuditAnnotations, "mutateme")

// }

//JSON
//[{\"op\":\"replace\",\"path\":\"/data/mutate1\",\"value\":\"c2VjbWFuOnRoaXMgaXMgdGhlIGZha2Ugc2VjcmV0\"},{\"op\":\"replace\",\"path\":\"/data/mutate3\",\"value\":\"c2Zrc2RsZmRza2YK\"},{\"op\":\"replace\",\"path\":\"/data/mutate\",\"value\":\"dmVyc2lvbiAyISBhd2Vzb21l\"}]

// build Response for Admission Review Response
// func admResponse(JSONPatch []byte, UID types.UID) v1beta1.AdmissionResponse {

// 	log.Debug("Creating Admission Response")
// 	pT := v1beta1.PatchTypeJSONPatch
// 	resp := v1beta1.AdmissionResponse{
// 		Allowed:          true,
// 		UID:              UID,
// 		PatchType:        &pT,
// 		Patch:            JSONPatch,
// 		AuditAnnotations: map[string]string{"kube-secman": "mutated"},
// 		Result:           &metav1.Status{Status: "Success"},
// 	}
// 	log.Debug("Admission Response successfully created")
// 	return resp
// }

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
		AuditAnnotations: map[string]string{"kube-secman": "mutated"},
		Result:           &metav1.Status{Status: "Success"},
	}

	assert.Equal(expected, actual, "The two responses should be the same.")
}
