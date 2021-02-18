//+build controllertests

package controller

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"path/filepath"
	ochclient "projectvoltron.dev/voltron/pkg/och/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"testing"
	"time"

	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"projectvoltron.dev/voltron/pkg/sdk/renderer/argo"
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

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	cfg := Config{
		BuiltinRunner: BuiltinRunnerConfig{
			Timeout: time.Second,
			Image:   "not-needed",
		},
		ClusterPolicy: ClusterPolicyConfig{},
	}
	err = (&ActionReconciler{
		log: ctrl.Log.WithName("controllers").WithName("Action"),
		svc: NewActionService(zap.NewRaw(zap.WriteTo(ioutil.Discard)), mgr.GetClient(), &argoRendererFake{}, cfg),
	}).SetupWithManager(mgr, maxConcurrentReconciles)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = mgr.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = mgr.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

type argoRendererFake struct{}

func (c *argoRendererFake) Render(ctx context.Context, runnerCtx argo.RunnerContextSecretRef, ref types.InterfaceRef, opts ...argo.RendererOption) (*types.Action, error) {
	return &types.Action{
		Args: map[string]interface{}{
			"workflow": "{}",
		},
		RunnerInterface: "argo.run",
	}, nil
}

func (c *argoRendererFake) PolicyEnforcer() argo.PolicyEnforcedOCHClient {
	return &ochclient.PolicyEnforcedClient{}
}

// returns path with OS specific Separator
func toOSPath(path string) string {
	return filepath.Join(strings.Split(path, "/")...)
}
