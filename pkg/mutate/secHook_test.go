package mutate

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1beta1 "k8s.io/api/admission/v1beta1"
)

type mockSecretManager struct {
	mock.Mock
}

func (m *mockSecretManager) GetSecret(ctx context.Context, secHookAddr string) ([]byte, error) {
	args := m.Called(ctx, secHookAddr)
	return args.Get(0).([]byte), args.Error(1)
}

func TestHandleMutateErrors(t *testing.T) {

	m := new(mockSecretManager)
	secHook := SecHook{Vault: m}
	ts := httptest.NewServer(http.HandlerFunc(secHook.HandleMutate))
	defer ts.Close()

	// TODO: Catch actual errors (not noerror)
	// default GET on the handle should throw an error trying to convert from empty JSON
	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.NoError(t, err)

	admReview := v1beta1.AdmissionReview{}
	assert.Errorf(t, json.Unmarshal(body, &admReview), "body: %s", string(body))

	// Post should throw an error trying to convert from JSON with wrong format - AdmissionReview.Request will be empty
	bodyReader := strings.NewReader(`{"test": "test"}`)
	_, err = http.Post(fmt.Sprintf("%s/mutate", ts.URL), "application/json", bodyReader)
	assert.NoError(t, err)
}
