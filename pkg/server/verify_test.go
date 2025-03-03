package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	eval "github.com/kyverno/kyverno/pkg/imageverification/evaluator"
	"github.com/nirmata/image-verification-service/pkg/policy"
	"github.com/stretchr/testify/assert"
)

var (
	obj = func(image string) map[string]any {
		return map[string]any{
			"foo": map[string]string{
				"bar": image,
			},
		}
	}

	signedImage   = "ghcr.io/kyverno/test-verify-image:signed"
	unsignedImage = "ghcr.io/kyverno/test-verify-image:unsigned"
)

func Test_Verify_Pass(t *testing.T) {
	w := verifyImage(t, signedImage)
	assert.Equal(t, w.Code, http.StatusOK)

	var result []*eval.EvaluationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.True(t, len(result) == 1)
	assert.True(t, result[0].Result)
}

func Test_Verify_Fail(t *testing.T) {
	w := verifyImage(t, unsignedImage)
	assert.Equal(t, w.Code, http.StatusOK)

	var result []*eval.EvaluationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.True(t, len(result) == 1)
	assert.False(t, result[0].Result)
	assert.Equal(t, result[0].Message, "failed to verify image with notary cert")
}

func verifyImage(t *testing.T, image string) *httptest.ResponseRecorder {
	o, err := policy.NewOCIPolicyFetcher(context.Background(), logr.Discard(), "ghcr.io/vishal-chdhry/ivpol:sample", 0, nil, nil)
	assert.NoError(t, err)

	handler := VerifyImagesHandler(logr.Discard(), o.Fetch, nil)
	data, err := json.Marshal(obj(image))
	assert.NoError(t, err)

	r, err := http.NewRequest(http.MethodPost, "/verifyimages", bytes.NewBuffer(data))
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	handler(w, r)

	return w
}
