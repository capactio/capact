//+build controllertests

package controller

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

var _ = Describe("Action Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("When Action CR is created", func() {
		It("Should render the action workflow", func() {
			key := types.NamespacedName{
				Name:      "action-test-1",
				Namespace: "default",
			}

			created := &corev1alpha1.Action{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: corev1alpha1.ActionSpec{
					Path: "bar",
				},
			}

			// Simulate that Action CR is created
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() error {
				action := &corev1alpha1.Action{}

				if err := k8sClient.Get(context.Background(), key, action); err != nil {
					return err
				}

				if action.Status.Rendering == nil || action.Status.Rendering.Action == nil {
					return errors.New(".Status.Rendering.Action field is empty")
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})
	})
})
