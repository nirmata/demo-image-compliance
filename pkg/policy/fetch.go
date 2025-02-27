package policy

import (
	"os"

	"github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
)

type Fetcher func() ([]*v1alpha1.ImageVerificationPolicy, error)

func FSPolicyFetcher() ([]*v1alpha1.ImageVerificationPolicy, error) {
	policiesDir := os.Getenv("POLICY_DIR")
	if len(policiesDir) == 0 {
		policiesDir = "policies"
	}

	return Load(policiesDir)
}
