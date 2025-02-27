package certmanager

import (
	"os"
	"time"
)

var (
	Namespace      = os.Getenv("POD_NAMESPACE")
	PodName        = os.Getenv("POD_NAME")
	ServiceName    = getEnvWithFallback("SERVICE_NAME", "svc")
	DeploymentName = getEnvWithFallback("DEPLOYMENT_NAME", "kyverno-notation-aws")

	CertRenewalInterval = 12 * time.Hour
	CAValidityDuration  = 365 * 24 * time.Hour
	TLSValidityDuration = 150 * 24 * time.Hour
)

func getEnvWithFallback(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
