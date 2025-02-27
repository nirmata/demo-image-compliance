package server

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyverno/image-verification-service/pkg/certmanager"
	"github.com/kyverno/image-verification-service/pkg/policy"
	"github.com/kyverno/kyverno/pkg/imageverification/imagedataloader"
	tlsMgr "github.com/kyverno/pkg/tls"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/informers/core/v1"
)

func NewServer(logger logr.Logger, tlsDisabled bool, tlsInformer v1.SecretInformer, opts ...imagedataloader.Option) chan error {
	mux := http.NewServeMux()
	mux.HandleFunc("/verifyimages", VerifyImagesHandler(logger, policy.FSPolicyFetcher, opts...))

	errsTLS := make(chan error, 1)
	if !tlsDisabled {
		tlsMgrConfig := &tlsMgr.Config{
			ServiceName: certmanager.ServiceName,
			Namespace:   certmanager.Namespace,
		}

		tlsConf := &tls.Config{
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				secret, err := tlsInformer.Lister().Secrets(certmanager.Namespace).Get(tlsMgr.GenerateTLSPairSecretName(tlsMgrConfig))
				if err != nil {
					return nil, err
				} else if secret == nil {
					return nil, errors.New("tls secret not found")
				} else if secret.Type != corev1.SecretTypeTLS {
					return nil, errors.New("secret is not a TLS secret")
				}

				cert, err := tls.X509KeyPair(secret.Data[corev1.TLSCertKey], secret.Data[corev1.TLSPrivateKeyKey])
				if err != nil {
					return nil, err
				}

				return &cert, nil
			},
		}
		srv := &http.Server{
			Addr:              ":9443",
			Handler:           mux,
			TLSConfig:         tlsConf,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			ReadHeaderTimeout: 30 * time.Second,
			IdleTimeout:       1 * time.Minute,
		}

		go func() {
			errsTLS <- srv.ListenAndServeTLS("", "")
		}()

		return errsTLS
	} else {
		errsHTTP := make(chan error, 1)
		go func() {
			errsHTTP <- http.ListenAndServe(":9080", mux)
		}()

		return errsHTTP
	}
}
