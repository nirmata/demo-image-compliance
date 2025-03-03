package policy

import (
	"path/filepath"
	"testing"

	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	basePath := "../../policies"
	tests := []struct {
		name    string
		path    string
		want    []*policiesv1alpha1.ImageVerificationPolicy
		wantErr bool
	}{{
		name:    "sample",
		path:    filepath.Join(basePath, "sample.yaml"),
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, len(got), 1)
			// assert.True(t, cmp.Equal(tt.want, got))
		})
	}
}
