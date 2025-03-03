package policy

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
)

type Fetcher func() ([]*policiesv1alpha1.ImageVerificationPolicy, error)

func PolicyFetcher(ctx context.Context, logger logr.Logger, reconcileDuration time.Duration, rOpts []remote.Option, nOpts []name.Option) (Fetcher, error) {
	policiesPath := os.Getenv("POLICY_PATH")
	if len(policiesPath) == 0 {
		policiesPath = "/policies"
	}

	if strings.HasPrefix(policiesPath, "oci://") {
		artifact := strings.TrimPrefix(policiesPath, "oci://")
		o, err := NewOCIPolicyFetcher(ctx, logger, artifact, reconcileDuration, rOpts, nOpts)
		if err != nil {
			return nil, err
		}
		return o.Fetch, nil
	} else {
		return func() ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
			return Load(policiesPath)
		}, nil
	}
}
