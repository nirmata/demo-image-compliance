package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	eval "github.com/kyverno/kyverno/pkg/imageverification/evaluator"
	"github.com/nirmata/demo-image-compliance/pkg/policy"
	"github.com/stretchr/testify/assert"
)

var (
	obj = func(image string) map[string]string {
		return map[string]string{
			"imageReference": image,
		}
	}

	signedImage   = "ghcr.io/kyverno/test-verify-image:signed"
	unsignedImage = "ghcr.io/kyverno/test-verify-image:unsigned"
)

func Test_Verify_Pass(t *testing.T) {
	w := verifyImage(t, signedImage, "ghcr.io/nirmata/image-compliance-policies:block-critical-vulnerabilities")
	assert.Equal(t, w.Code, http.StatusOK)

	var result map[string]*eval.EvaluationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.True(t, len(result) == 1)
	assert.True(t, result["sample"].Result)
}

func Test_Verify_Fail(t *testing.T) {
	w := verifyImage(t, unsignedImage, "ghcr.io/nirmata/image-compliance-policies:block-critical-vulnerabilities")
	assert.Equal(t, w.Code, http.StatusOK)

	var result map[string]*eval.EvaluationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.True(t, len(result) == 1)
	assert.False(t, result["sample"].Result)
	assert.Equal(t, result["sample"].Message, "failed to verify image with notary cert")
}

func Test_Verify_Attestation_Fail(t *testing.T) {
	w := verifyImage(t, signedImage, "ghcr.io/nirmata/image-compliance-policies:block-high-and-critical-vulnerabilities")
	assert.Equal(t, w.Code, http.StatusOK)

	var result map[string]*eval.EvaluationResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.True(t, len(result) == 1)
	assert.False(t, result["sample"].Result)
	assert.Equal(t, result["sample"].Message, "the image has vulnerabilities of HIGH or CRITICAL severity")
}

func verifyImage(t *testing.T, image, policyImage string) *httptest.ResponseRecorder {
	o, err := policy.NewOCIPolicyFetcher(context.Background(), logr.Discard(), policyImage, 0, nil, nil)
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

func VerifyImageLocal(t *testing.T, image string) *httptest.ResponseRecorder {
	handler := VerifyImagesHandler(logr.Discard(), func() ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
		return policy.Load("../../policies/critical.yaml")
	}, nil)
	data, err := json.Marshal(obj(image))
	assert.NoError(t, err)

	r, err := http.NewRequest(http.MethodPost, "/verifyimages", bytes.NewBuffer(data))
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	handler(w, r)

	return w
}
