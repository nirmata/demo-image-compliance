package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	eval "github.com/kyverno/kyverno/pkg/imageverification/evaluator"
	"github.com/kyverno/kyverno/pkg/imageverification/imagedataloader"
	"github.com/nirmata/demo-image-compliance/pkg/api"
	"github.com/nirmata/demo-image-compliance/pkg/policy"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func VerifyImagesHandler(logger logr.Logger,
	policyFetcher policy.Fetcher,
	lister k8scorev1.SecretInterface,
	opts ...imagedataloader.Option) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		var requestData api.RequestData
		raw, _ := io.ReadAll(r.Body)

		err := json.Unmarshal(raw, &requestData)
		if err != nil {
			logger.Info("failed to decode", "data", string(raw), "error", err)
			http.Error(w, errors.Wrapf(err, "failed to decode").Error(), http.StatusNotAcceptable)
			return
		}
		logger.Info("Request received", "data", requestData)

		policies, err := policyFetcher()
		if err != nil {
			logger.Info("failed to fetch policies", "error", err)
			http.Error(w, errors.Wrapf(err, "failed to fetch policies").Error(), http.StatusInternalServerError)
			return
		}

		result, err := eval.Evaluate(context.Background(), logger, policies, requestData, nil, nil, nil, opts...)
		if err != nil {
			logger.Info("failed to evaluate request", "error", err)
			http.Error(w, errors.Wrapf(err, "failed to evaluate request").Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(result)
		if err != nil {
			logger.Info("failed to decode result", "error", err)
			http.Error(w, errors.Wrapf(err, "failed to decode result").Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Sending response", "data", string(data))
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			logger.Info("failed to write result", "error", err)
			http.Error(w, errors.Wrapf(err, "failed to write result").Error(), http.StatusInternalServerError)
			return
		}
	}
}
