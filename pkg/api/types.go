package api

import "github.com/kyverno/kyverno/pkg/cel/policy"

type RequestData any

type ResponseData struct {
	Results []*policy.EvaluationResult `json:"results"`
}
