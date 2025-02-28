package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	eval "github.com/kyverno/kyverno/pkg/imageverification/evaluator"
	"github.com/stretchr/testify/assert"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
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

	ivpol = &policiesv1alpha1.ImageVerificationPolicy{
		Spec: policiesv1alpha1.ImageVerificationPolicySpec{
			ImageRules: []policiesv1alpha1.ImageRule{
				{
					Glob: "ghcr.io/*",
				},
			},
			Images: []policiesv1alpha1.Image{
				{
					Name:       "bar",
					Expression: "[request.foo.bar]",
				},
			},
			Attestors: []policiesv1alpha1.Attestor{
				{
					Name: "notary",
					Notary: &policiesv1alpha1.Notary{
						Certs: `-----BEGIN CERTIFICATE-----
MIIDTTCCAjWgAwIBAgIJAPI+zAzn4s0xMA0GCSqGSIb3DQEBCwUAMEwxCzAJBgNV
BAYTAlVTMQswCQYDVQQIDAJXQTEQMA4GA1UEBwwHU2VhdHRsZTEPMA0GA1UECgwG
Tm90YXJ5MQ0wCwYDVQQDDAR0ZXN0MB4XDTIzMDUyMjIxMTUxOFoXDTMzMDUxOTIx
MTUxOFowTDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAldBMRAwDgYDVQQHDAdTZWF0
dGxlMQ8wDQYDVQQKDAZOb3RhcnkxDTALBgNVBAMMBHRlc3QwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQDNhTwv+QMk7jEHufFfIFlBjn2NiJaYPgL4eBS+
b+o37ve5Zn9nzRppV6kGsa161r9s2KkLXmJrojNy6vo9a6g6RtZ3F6xKiWLUmbAL
hVTCfYw/2n7xNlVMjyyUpE+7e193PF8HfQrfDFxe2JnX5LHtGe+X9vdvo2l41R6m
Iia04DvpMdG4+da2tKPzXIuLUz/FDb6IODO3+qsqQLwEKmmUee+KX+3yw8I6G1y0
Vp0mnHfsfutlHeG8gazCDlzEsuD4QJ9BKeRf2Vrb0ywqNLkGCbcCWF2H5Q80Iq/f
ETVO9z88R7WheVdEjUB8UrY7ZMLdADM14IPhY2Y+tLaSzEVZAgMBAAGjMjAwMAkG
A1UdEwQCMAAwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMDMA0G
CSqGSIb3DQEBCwUAA4IBAQBX7x4Ucre8AIUmXZ5PUK/zUBVOrZZzR1YE8w86J4X9
kYeTtlijf9i2LTZMfGuG0dEVFN4ae3CCpBst+ilhIndnoxTyzP+sNy4RCRQ2Y/k8
Zq235KIh7uucq96PL0qsF9s2RpTKXxyOGdtp9+HO0Ty5txJE2txtLDUIVPK5WNDF
ByCEQNhtHgN6V20b8KU2oLBZ9vyB8V010dQz0NRTDLhkcvJig00535/LUylECYAJ
5/jn6XKt6UYCQJbVNzBg/YPGc1RF4xdsGVDBben/JXpeGEmkdmXPILTKd9tZ5TC0
uOKpF5rWAruB5PCIrquamOejpXV9aQA/K2JQDuc0mcKz
-----END CERTIFICATE-----`,
					},
				},
			},
			Attestations: []policiesv1alpha1.Attestation{
				{
					Name: "sbom",
					Referrer: &policiesv1alpha1.Referrer{
						Type: "sbom/cyclone-dx",
					},
				},
			},
			Verifications: []admissionregistrationv1.Validation{
				{
					Expression: "images.bar.map(image, verifyImageSignatures(image, [attestors.notary])).all(e, e > 0)",
					Message:    "failed to verify image with notary cert",
				},
				{
					Expression: "images.bar.map(image, verifyAttestationSignatures(image, attestations.sbom ,[attestors.notary])).all(e, e > 0)",
					Message:    "failed to verify attestation with notary cert",
				},
			},
		},
	}

	policyFetcher = func() ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
		return []*policiesv1alpha1.ImageVerificationPolicy{ivpol}, nil
	}
)

func Test_Verify_Pass(t *testing.T) {
	w := verifyImage(t, unsignedImage)
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
	handler := VerifyImagesHandler(logr.Discard(), policyFetcher)
	data, err := json.Marshal(obj(image))
	assert.NoError(t, err)

	r, err := http.NewRequest(http.MethodPost, "/verifyimages", bytes.NewBuffer(data))
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	handler(w, r)

	return w
}
