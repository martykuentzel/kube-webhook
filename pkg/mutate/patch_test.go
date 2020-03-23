package mutate

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

func TestPatchSecrets(t *testing.T) {

	assert := assert.New(t)
	ctx := context.Background()
	m := new(mockSecretManager)
	retrievedSecret := []byte("blabla")

	secret := corev1.Secret{
		Data: map[string][]byte{
			"dnt_mutate": []byte{100, 111, 32, 110, 111, 116, 32, 109, 117, 116, 97, 116, 101, 32, 116, 104, 105, 115, 32, 107, 101, 121},
			"mutate":     []byte{115, 101, 99, 109, 97, 110, 58, 112, 114, 111, 106, 101, 99, 116, 115, 47, 55, 55, 54, 50, 52, 49, 54, 56, 48, 51, 52, 48, 47, 115, 101, 99, 114, 101, 116, 115, 47, 116, 101, 115, 116, 101, 114, 47, 118, 101, 114, 115, 105, 111, 110, 115, 47, 108, 97, 116, 101, 115, 116, 10},
			"mutate1":    []byte{115, 101, 99, 109, 97, 110, 58, 116, 104, 105, 115, 32, 105, 115, 32, 116, 104, 101, 32, 102, 97, 107, 101, 32, 115, 101, 99, 114, 101, 116},
			"mutate3":    []byte{115, 101, 99, 109, 97, 110, 58, 112, 114, 111, 106, 101, 99, 116, 115, 47, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100, 45, 109, 97, 114, 116, 121, 45, 107, 117, 101, 110, 116, 122, 101, 108, 47, 115, 101, 99, 114, 101, 116, 115, 47, 98, 108, 97, 98, 108, 97, 47, 118, 101, 114, 115, 105, 111, 110, 115, 47, 108, 97, 116, 101, 115, 116, 10},
			"mutate4":    []byte(" secman: /project/34343/bla/latest"),
		},
	}
	var dumySecret *corev1.Secret = &secret
	secManEntries := findAllSecManEntries(dumySecret.Data)
	m.On("GetSecret", ctx, mock.AnythingOfType("string")).Return(retrievedSecret, nil).Times(len(secManEntries))

	actual := patchSecrets(ctx, dumySecret, m)
	expected := []map[string]string(
		[]map[string]string{
			map[string]string{"op": "replace", "path": "/data/mutate", "value": "YmxhYmxh"},
			map[string]string{"op": "replace", "path": "/data/mutate1", "value": "YmxhYmxh"},
			map[string]string{"op": "replace", "path": "/data/mutate3", "value": "YmxhYmxh"},
			map[string]string{"op": "replace", "path": "/data/mutate4", "value": "YmxhYmxh"},
		},
	)

	m.AssertExpectations(t)
	assert.ElementsMatch(expected, actual, "Patching was not successfull.")
}

func TestFindAllSecManEntries(t *testing.T) {

	assert := assert.New(t)
	for _, tt := range []struct {
		n        map[string][]byte // input
		expected map[string]string // expected result
	}{
		{map[string][]byte{"key1": []byte("secman: /first/key"), "key2": []byte("secretPassword")}, map[string]string{"key1": "secman: /first/key"}},
		{map[string][]byte{"key1": []byte("secman/first/key"), "key2": []byte("secretPassword"), "key3": []byte("Secman:/project/bla")}, map[string]string{}},
		{map[string][]byte{"key1": []byte(" secman: /first/key"), "key2": []byte(" secman:/second/key ")}, map[string]string{"key1": " secman: /first/key", "key2": " secman:/second/key "}},
	} {

		actual := findAllSecManEntries(tt.n)
		assert.Equal(tt.expected, actual, "Filtering of Map was not successfull", tt.n, tt.expected)
	}
}

func TestHasSecManPrefixt(t *testing.T) {

	assert := assert.New(t)
	for _, tt := range []struct {
		n        string // input
		expected bool   // expected result
	}{
		{"secman:/abc", true},
		{"Secman: /abc", false},
		{"SecMan:/abc", false},
		{" secman:/abc", true},
		{" secman: /abc ", true},
		{" SECMAN: /abc", false},
		{"secman/abc", false},
		{"secbla: /abc", false},
	} {
		actual := hasSecManPrefix(tt.n)
		assert.Equal(tt.expected, actual, "The Prefix '%s' two Booleans should be tested %t.", tt.n, tt.expected)
	}
}

func TestReplaceSecManVals_1(t *testing.T) {

	ctx := context.Background()
	m := new(mockSecretManager)

	for _, tt := range []struct {
		n map[string]string // input replaceSecManVals
		e string            // expected
	}{
		{map[string]string{"key1": " secman:/second/key1 "}, "/second/key1"},
		{map[string]string{"key2": "secman: /second/key2 "}, "/second/key2"},
		{map[string]string{"key3": "secman:/second/key3 "}, "/second/key3"},
		{map[string]string{"key4": "secman: /project/Bla"}, "/project/Bla"},
		{map[string]string{"key5": "Secman:/second/key5 "}, "Secman:/second/key5"},
	} {
		m.On("GetSecret", ctx, tt.e).Return([]byte("geheim"), nil).Once()
		replaceSecManVals(ctx, m, tt.n)

		// assert that the expectations were met -> removal of secman prefix
		m.AssertExpectations(t)
	}
}

func TestReplaceSecManVals_2(t *testing.T) {

	ctx := context.Background()
	m := new(mockSecretManager)
	assert := assert.New(t)
	retrievedSecret := []byte("blabla")

	for _, tt := range []struct {
		n        map[string]string   // input replaceSecManVals
		err      error               // error of GetSecret
		expected []map[string]string // expected result
	}{
		{map[string]string{"key1": "secman:/first/key", "key2": " secman:/second/key "}, // input replaceSecManVals
			nil, // error of GetSecret
			[]map[string]string([]map[string]string{map[string]string{"op": "replace", "path": "/data/key1", "value": "YmxhYmxh"}, // expected result
				map[string]string{"op": "replace", "path": "/data/key2", "value": "YmxhYmxh"}})},

		{map[string]string{"key1": "secman:/first/key", "key2": " secman:/second/key "}, // input replaceSecManVals
			errors.New("Some Error"), // error of GetSecret
			[]map[string]string([]map[string]string{map[string]string{"op": "replace", "path": "/data/key1", "value": "c2VjbWFuOi9maXJzdC9rZXk="}, // expected result
				map[string]string{"op": "replace", "path": "/data/key2", "value": "IHNlY21hbjovc2Vjb25kL2tleSA="}})},
	} {
		m.On("GetSecret", ctx, mock.AnythingOfType("string")).Return(retrievedSecret, tt.err).Times(len(tt.n))
		actual := replaceSecManVals(ctx, m, tt.n)

		//assert that result-patch is created correctly
		assert.ElementsMatch(tt.expected, actual, "The two Maps should be the same.")
	}
}

func TestRemoveSecManPrefix(t *testing.T) {

	assert := assert.New(t)
	for _, tt := range []struct {
		n        string // input
		expected string // expected result
	}{
		{"secman:/project/3423443/secrets/test/latest\n", "/project/3423443/secrets/test/latest"},
		{"secman: /project/3423443/secrets/test/2\n", "/project/3423443/secrets/test/2"},
		{" secman:/project/3423443/secrets/test/3", "/project/3423443/secrets/test/3"},
		{"secman:  /project/3423443/secrets/test/4\n\n", "/project/3423443/secrets/test/4"},
		{"secman:/project/3423443/secrets/test/5 ", "/project/3423443/secrets/test/5"},
		{"secman: /project/3423443/secrets/test/6 \n", "/project/3423443/secrets/test/6"},
		{"secman:/project/3423443/secrets/test/7\n ", "/project/3423443/secrets/test/7"},
		{"secman:\t /project/3423443/secrets/test/7\n ", "/project/3423443/secrets/test/7"},
		{"secman:/project/3423443/secrets/test/7\n \t", "/project/3423443/secrets/test/7"},
		{"secman: ", ""},
		{"secman:", ""},
	} {
		actual := removeSecManPrefix(tt.n)
		assert.Equal(tt.expected, actual, "The two strings should be the same.")
	}
}

var fakeSaCredentials string = `{
	"type": "service_account",
	"project_id": "project-bob-baumeister",
	"private_key_id": "32395849584hjrjgkgkr4854b",
	"private_key": "-----BEGIN PRIVATE KEY-----\nMII4r4rwgNehuBKiEs\npewjFA5445ttreggWDKWDjdedDDDDDDKJm+6\nzKWDWDDWdewforeofrefrefrefrWe4f6j\nG/wKWDjdwd838CpWV5xN4VmXI5NQ2Iz+tJFfKQHg\nFTFYr8Yscjecnejd/sxx/xwxwedFu7h//vCm0eHityq15ZSEude\nPpbOEycEca1YiRKbI3j3wECpKnrgid0QFZTuxwpuSciZwKpuFVsHSIx2jBdQQ8mK\nZLV2G2WZVBXsCWTewSDhSW4PjXSc/ZLREMsryoN8hoIWp1GFiQMCuPLCCzAobm8/\nGVuoPuNR7HUdypKMk2ctsy7U85zllqHBqPXoEubb4MoHb5JXGdJF+9pyERV3NQGd\nGKpI7Eo0eVXWloKnzzu4kEDkvrej5Q5fwop+IPwEBHHHLRkTOLxD/qhleoGrPZc2\nNTmfpwam4BYi/JV23qMOstsDRNBjonKeTI1gILwVvwKBgQDhTsdX0f4rWnROP/Vg\nwJ/moZa0VCN1fQRdRo3nJoE5OvVxAf/1d1JfKnUlJn64r48r94SxQfl+14IhjuOA==\n-----END PRIVATE KEY-----\n",
	"client_email": "boby@project-bob-baumeiste.iam.gserviceaccount.com",
	"client_id": "99999999999",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/boboycewcblablaproject.nz"
  }`

var base64FakeCreds string = "ewoJInR5cGUiOiAic2VydmljZV9hY2NvdW50IiwKCSJwcm9qZWN0X2lkIjogInByb2plY3QtYm9iLWJhdW1laXN0ZXIiLAoJInByaXZhdGVfa2V5X2lkIjogIjMyMzk1ODQ5NTg0aGpyamdrZ2tyNDg1NGIiLAoJInByaXZhdGVfa2V5IjogIi0tLS0tQkVHSU4gUFJJVkFURSBLRVktLS0tLVxuTUlJNHI0cndnTmVodUJLaUVzXG5wZXdqRkE1NDQ1dHRyZWdnV0RLV0RqZGVkRERERERES0ptKzZcbnpLV0RXRERXZGV3Zm9yZW9mcmVmcmVmcmVmcldlNGY2alxuRy93S1dEamR3ZDgzOENwV1Y1eE40Vm1YSTVOUTJJeit0SkZmS1FIZ1xuRlRGWXI4WXNjamVjbmVqZC9zeHgveHd4d2VkRnU3aC8vdkNtMGVIaXR5cTE1WlNFdWRlXG5QcGJPRXljRWNhMVlpUktiSTNqM3dFQ3BLbnJnaWQwUUZaVHV4d3B1U2NpWndLcHVGVnNIU0l4MmpCZFFROG1LXG5aTFYyRzJXWlZCWHNDV1Rld1NEaFNXNFBqWFNjL1pMUkVNc3J5b044aG9JV3AxR0ZpUU1DdVBMQ0N6QW9ibTgvXG5HVnVvUHVOUjdIVWR5cEtNazJjdHN5N1U4NXpsbHFIQnFQWG9FdWJiNE1vSGI1SlhHZEpGKzlweUVSVjNOUUdkXG5HS3BJN0VvMGVWWFdsb0tuenp1NGtFRGt2cmVqNVE1ZndvcCtJUHdFQkhISExSa1RPTHhEL3FobGVvR3JQWmMyXG5OVG1mcHdhbTRCWWkvSlYyM3FNT3N0c0RSTkJqb25LZVRJMWdJTHdWdndLQmdRRGhUc2RYMGY0clduUk9QL1ZnXG53Si9tb1phMFZDTjFmUVJkUm8zbkpvRTVPdlZ4QWYvMWQxSmZLblVsSm42NHI0OHI5NFN4UWZsKzE0SWhqdU9BPT1cbi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS1cbiIsCgkiY2xpZW50X2VtYWlsIjogImJvYnlAcHJvamVjdC1ib2ItYmF1bWVpc3RlLmlhbS5nc2VydmljZWFjY291bnQuY29tIiwKCSJjbGllbnRfaWQiOiAiOTk5OTk5OTk5OTkiLAoJImF1dGhfdXJpIjogImh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbS9vL29hdXRoMi9hdXRoIiwKCSJ0b2tlbl91cmkiOiAiaHR0cHM6Ly9vYXV0aDIuZ29vZ2xlYXBpcy5jb20vdG9rZW4iLAoJImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLAoJImNsaWVudF94NTA5X2NlcnRfdXJsIjogImh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL3JvYm90L3YxL21ldGFkYXRhL3g1MDkvYm9ib3ljZXdjYmxhYmxhcHJvamVjdC5ueiIKICB9"

func TestCreatePatch(t *testing.T) {

	assert := assert.New(t)
	for _, tt := range []struct {
		secretKey   string            // input
		secretValue []byte            // input
		expected    map[string]string // expected result
	}{
		{"mutate", []byte("s€cr€T_Pa$$Word"), map[string]string{"op": "replace", "path": "/data/mutate", "value": "c+KCrGNy4oKsVF9QYSQkV29yZA=="}},
		{"mutate1", []byte(fakeSaCredentials), map[string]string{"op": "replace", "path": "/data/mutate1", "value": base64FakeCreds}},
		{"mutate2", []byte("3j i4j 934j 34 j3 K9 L0 j$ k3 w4 3r"), map[string]string{"op": "replace", "path": "/data/mutate2", "value": "M2ogaTRqIDkzNGogMzQgajMgSzkgTDAgaiQgazMgdzQgM3I="}},
		{"failedMuation", []byte("secman: /project/3423443/secrets/test/7"), map[string]string{"op": "replace", "path": "/data/failedMuation", "value": "c2VjbWFuOiAvcHJvamVjdC8zNDIzNDQzL3NlY3JldHMvdGVzdC83"}},
	} {
		actual := createPatch(tt.secretKey, tt.secretValue)
		actualDecoded, _ := base64.StdEncoding.DecodeString(actual["value"])

		// assert that the passwords are identical
		assert.Equal(string(tt.secretValue), string(actualDecoded), "The two values should be the same.")
		// assert that the base64 strings are identical
		assert.Equal(tt.expected["value"], actual["value"], "The two encodings should be the same.")
		// assert that the paths are identical
		assert.Equal(tt.expected["path"], actual["path"], "The two encodings should be the same.")
	}
}
