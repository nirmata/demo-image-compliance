package policy

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kyverno/image-verification-service/pkg/data"
	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	fileinfo "github.com/kyverno/pkg/ext/file-info"
	"github.com/kyverno/pkg/ext/resource/convert"
	"github.com/kyverno/pkg/ext/resource/loader"
	yamlutils "github.com/kyverno/pkg/ext/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kubectl-validate/pkg/openapiclient"
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

func Parse(content []byte) ([]*policiesv1alpha1.ImageVerificationPolicy, error) {
	documents, err := yamlutils.SplitDocuments(content)
	if err != nil {
		return nil, err
	}
	crds, err := data.Crds()
	if err != nil {
		return nil, err
	}
	loader, err := loader.New(openapiclient.NewLocalCRDFiles(crds))
	if err != nil {
		return nil, err
	}
	var policies []*policiesv1alpha1.ImageVerificationPolicy
	for _, document := range documents {
		gvk, untyped, err := loader.Load(document)
		if err != nil {
			return nil, err
		}
		switch gvk {
		case ivpol_v1alpha1:
			policy, err := convert.To[policiesv1alpha1.ImageVerificationPolicy](untyped)
			if err != nil {
				return nil, err
			}
			policies = append(policies, policy)
		default:
			return nil, fmt.Errorf("policy type not supported %s", gvk)
		}
	}
	return policies, nil
}
