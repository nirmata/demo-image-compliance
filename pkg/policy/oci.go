package policy

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
)

func NewOCIPolicyFetcher(ctx context.Context, logger logr.Logger, artifact string, reconcileDuration time.Duration, rOpts []remote.Option, nOpts []name.Option) (*ociPolicyFetcher, error) {
	var ticker *time.Ticker
	if reconcileDuration != 0 {
		ticker = time.NewTicker(reconcileDuration)
	}

	ivpols, err := fetchPoliciesFromArtifact(artifact, rOpts, nOpts)
	if err != nil {
		return nil, err
	}
	logger.Info("fetched policies", "artifact", artifact, "policies", len(ivpols))

	o := &ociPolicyFetcher{
		logger:   logger,
		artifact: artifact,
		ticker:   ticker,
		rOpts:    rOpts,
		nOpts:    nOpts,
		ivpols:   ivpols,
	}

	o.Reconcile(ctx)
	return o, nil
}

type ociPolicyFetcher struct {
	mu       sync.RWMutex
	logger   logr.Logger
	artifact string
	ticker   *time.Ticker
	rOpts    []remote.Option
	nOpts    []name.Option
	ivpols   []*policiesv1alpha1.ImageVerificationPolicy
}

func (o *ociPolicyFetcher) Fetch() ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.ivpols, nil
}

func (o *ociPolicyFetcher) Reconcile(ctx context.Context) {
	if o.ticker == nil {
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-o.ticker.C:
				policies, err := fetchPoliciesFromArtifact(o.artifact, o.rOpts, o.nOpts)
				if err != nil {
					o.logger.Error(err, "failed to reconcile policies", "artifact", o.artifact)
				}
				o.logger.Info("reconciled policies", "artifact", o.artifact, "policies", len(policies))
				o.mu.Lock()
				o.ivpols = policies
				o.mu.Unlock()
			}
		}
	}()
}

func fetchPoliciesFromArtifact(image string, rOpts []remote.Option, nOpts []name.Option) ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
	policies := make([]*policiesv1alpha1.ImageVerificationPolicy, 0)
	ref, err := name.ParseReference(image, nOpts...)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref, rOpts...)
	if err != nil {
		return nil, err
	}

	tr := tar.NewReader(mutate.Extract(img))
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		name := header.Name
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			var buf bytes.Buffer
			buf.ReadFrom(tr)
			p, err := Parse(buf.Bytes())
			if err != nil {
				return nil, err
			}
			policies = append(policies, p...)
		}
	}

	return policies, nil
}
