package policy

import (
	"fmt"

	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	"github.com/kyverno/pkg/ext/resource/convert"
	"github.com/kyverno/pkg/ext/resource/loader"
	yamlutils "github.com/kyverno/pkg/ext/yaml"
	"github.com/nirmata/image-verification-service/pkg/data"
	"sigs.k8s.io/kubectl-validate/pkg/openapiclient"
)

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
