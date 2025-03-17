package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	policiesv1alpha1 "github.com/kyverno/kyverno/api/policies.kyverno.io/v1alpha1"
	"github.com/kyverno/kyverno/pkg/imageverification/imagedataloader"
	tlsMgr "github.com/kyverno/pkg/tls"
	"github.com/nirmata/demo-image-compliance/pkg/certmanager"
	"github.com/nirmata/demo-image-compliance/pkg/policy"
	"github.com/nirmata/demo-image-compliance/pkg/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	resyncPeriod = 15 * time.Minute
)

func main() {
	var (
		flagTLS                   bool
		flagImagePullSecrets      string
		flagAllowInsecureRegistry bool
		enableLeaderElection      bool
		reconcileDuration         time.Duration
	)

	flag.BoolVar(&flagTLS, "notls", false, "Disable HTTPS")
	flag.StringVar(
		&flagImagePullSecrets, "imagePullSecrets", "",
		"Secret resource names for image registry access credentials.")
	flag.BoolVar(
		&flagAllowInsecureRegistry, "allowInsecureRegistry", false,
		"Whether to allow insecure connections to registries. Not recommended.")
	flag.DurationVar(
		&reconcileDuration, "reconcileDuration", time.Hour,
		"Frequency of reconciling policy artifact")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.Parse()
	zc := zap.NewDevelopmentConfig()
	zc.Level = zap.NewAtomicLevelAt(zapcore.Level(-2))
	zlogger, err := zc.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	logger := zapr.NewLogger(zlogger)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to get kubernetes cluster config: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to initialize kube client: %v", err)
	}

	signalCtx, sdown := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer sdown()

	tlsMgrConfig := &tlsMgr.Config{
		ServiceName: certmanager.ServiceName,
		Namespace:   certmanager.Namespace,
	}

	caStopCh := make(chan struct{}, 1)
	caInformer := certmanager.NewSecretInformer(
		kubeClient,
		certmanager.Namespace,
		tlsMgr.GenerateRootCASecretName(tlsMgrConfig),
		resyncPeriod)
	go caInformer.Informer().Run(caStopCh)

	tlsStopCh := make(chan struct{}, 1)
	tlsInformer := certmanager.NewSecretInformer(kubeClient,
		certmanager.Namespace,
		tlsMgr.GenerateTLSPairSecretName(tlsMgrConfig),
		resyncPeriod)
	go tlsInformer.Informer().Run(tlsStopCh)

	if err != nil {
		log.Fatalf("failed to initialize leader election: %v", err)
		os.Exit(1)
	}

	err = certmanager.StartCertManager(logger, signalCtx, kubeClient, caInformer, tlsInformer, tlsMgrConfig)
	if err != nil {
		log.Fatalf("failed to initialize leader election: %v", err)
		os.Exit(1)
	}

	secrets := make([]string, 0)
	if len(flagImagePullSecrets) != 0 {
		secrets = append(secrets, strings.Split(flagImagePullSecrets, ",")...)
	}
	rOpts, nOpts, err := policy.RegistryOpts(
		kubeClient.CoreV1().Secrets(certmanager.Namespace),
		flagAllowInsecureRegistry,
		secrets...,
	)
	if err != nil {
		log.Fatalf("failed to initialize remote options: %v", err)
		os.Exit(1)
	}

	fetcher, err := policy.PolicyFetcher(signalCtx, logger, reconcileDuration, rOpts, nOpts)
	if err != nil {
		log.Fatalf("failed to initialize policy fetcher: %v", err)
		os.Exit(1)
	}

	errChan := server.NewServer(
		logger,
		flagTLS,
		tlsInformer,
		fetcher,
		kubeClient.CoreV1().Secrets(certmanager.Namespace),
		BuildRemoteOpts(secrets, []string{
			string(policiesv1alpha1.ACR),
			string(policiesv1alpha1.GCP),
			string(policiesv1alpha1.AWS),
			string(policiesv1alpha1.GHCR),
			string(policiesv1alpha1.DEFAULT),
		},
			flagAllowInsecureRegistry)...,
	)

	logger.Info("Listening for requests...")
	select {
	case err := <-errChan:
		logger.Info("TLS server error", "error", err)
		Shutdown(zlogger, &caStopCh, &tlsStopCh)
		os.Exit(-1)
	case <-signalCtx.Done():
		logger.Info("Shutting down service")
		Shutdown(zlogger, &caStopCh, &tlsStopCh)
		os.Exit(-1)
	}
}

func Shutdown(zlogger *zap.Logger, caStopCh *chan struct{}, tlsStopCh *chan struct{}) {
	_ = zlogger.Sync()
	*caStopCh <- struct{}{}
	*tlsStopCh <- struct{}{}
}

func BuildRemoteOpts(secrets []string, providers []string, insecure bool) []imagedataloader.Option {
	opts := make([]imagedataloader.Option, 0)

	if insecure {
		opts = append(opts, imagedataloader.WithInsecure(insecure))
	}
	if len(providers) != 0 {
		opts = append(opts, imagedataloader.WithCredentialProviders(providers...))
	}
	if len(secrets) != 0 {
		opts = append(opts, imagedataloader.WithPullSecret(secrets))
	}

	return opts
}
