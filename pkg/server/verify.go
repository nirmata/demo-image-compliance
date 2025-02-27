package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/kyverno/image-verification-service/pkg/api"
	"github.com/kyverno/image-verification-service/pkg/policy"
	"github.com/kyverno/kyverno/pkg/imageverification/imagedataloader"
)

func VerifyImagesHandler(logger logr.Logger, policyFetcher policy.Fetcher, opts ...imagedataloader.Option) func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}
		logger.Info("Request recieved", "data", requestData)

		var data []byte

		policies, err := policyFetcher()
		if err != nil {
			logger.Info("failed to fetch policies", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		policies = policies
		// image verify

		logger.Info("Sending response", "data", string(data))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
