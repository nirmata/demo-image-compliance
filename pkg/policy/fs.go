package policy

import (
	"io/fs"
	"os"
	"path/filepath"

	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	fileinfo "github.com/kyverno/pkg/ext/file-info"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	gv_v1alpha1    = schema.GroupVersion{Group: "policies.kyverno.io", Version: "v1alpha1"}
	ivpol_v1alpha1 = gv_v1alpha1.WithKind("ImageVerificationPolicy")
)

func Load(path ...string) ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
	var policies []*policiesv1alpha1.ImageVerificationPolicy
	for _, path := range path {
		p, err := load(path)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p...)
	}
	return policies, nil
}

func load(path string) ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
	var files []string
	err := filepath.Walk(path, func(file string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileinfo.IsYaml(info) {
			files = append(files, file)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var policies []*policiesv1alpha1.ImageVerificationPolicy
	for _, path := range files {
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, err
		}
		p, err := Parse(content)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p...)
	}
	return policies, nil
}
