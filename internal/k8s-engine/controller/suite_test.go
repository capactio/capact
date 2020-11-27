//+build controllertests

package controller

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	ochgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

const (
	crdDirectory            = "../../../deploy/kubernetes/crds"
	maxConcurrentReconciles = 1
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths:     []string{toOSPath(crdDirectory)},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = corev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&ActionReconciler{
		Client:           k8sManager.GetClient(),
		Log:              ctrl.Log.WithName("controllers").WithName("Action"),
		gatewayInterface: &gatewayClientMock{},
	}).SetupWithManager(k8sManager, maxConcurrentReconciles)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

type gatewayClientMock struct{}

func (c *gatewayClientMock) GetImplementationLatestRevision(ctx context.Context, path string) (*ochgraphql.ImplementationRevision, error) {
	return &ochgraphql.ImplementationRevision{
		Spec: &ochgraphql.ImplementationSpec{
			Action: &ochgraphql.ImplementationAction{
				Args: map[string]interface{}{
					"template": "main",
				},
			},
		},
	}, nil
}

// returns path with OS specific Separator
func toOSPath(path string) string {
	return filepath.Join(strings.Split(path, "/")...)
}
