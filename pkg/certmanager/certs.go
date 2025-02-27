package certmanager

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyverno/kyverno/pkg/leaderelection"
	"github.com/kyverno/pkg/certmanager"
	tlsMgr "github.com/kyverno/pkg/tls"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

func StartCertManager(logger logr.Logger, ctx context.Context, kubeClient kubernetes.Interface, caInformer v1.SecretInformer, tlsInformer v1.SecretInformer, tlsConfig *tlsMgr.Config) error {
	tlsMgrConfig := &tlsMgr.Config{
		ServiceName: ServiceName,
		Namespace:   Namespace,
	}

	le, err := leaderelection.New(
		logger.WithName("leader-election"),
		DeploymentName,
		Namespace,
		kubeClient,
		PodName,
		2*time.Second,
		func(ctx context.Context) {

			certRenewer := tlsMgr.NewCertRenewer(
				logger.WithName("tls").WithValues("pod", PodName),
				kubeClient.CoreV1().Secrets(Namespace),
				CertRenewalInterval,
				CAValidityDuration,
				TLSValidityDuration,
				"",
				tlsMgrConfig,
			)

			certManager := certmanager.NewController(
				logger.WithName("certmanager").WithValues("pod", PodName),
				caInformer,
				tlsInformer,
				certRenewer,
				tlsMgrConfig,
			)

			leaderControllers := []Controller{NewController("cert-manager", certManager, 1)}

			// start leader controllers
			var wg sync.WaitGroup
			for _, controller := range leaderControllers {
				controller.Run(ctx, logger.WithName("controllers"), &wg)
			}
			// wait all controllers shut down
			wg.Wait()
		},
		nil,
	)

	if err != nil {
		return err
	}
	// start leader election
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			le.Run(ctx)
		}
	}()

	return nil
}
