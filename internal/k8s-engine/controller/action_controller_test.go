//+build controllertests

package controller

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	graphqldomain "capact.io/capact/internal/k8s-engine/graphql/domain/action"
	corev1alpha1 "capact.io/capact/pkg/engine/k8s/api/v1alpha1"
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
		It("should render the action workflow without user input", func() {
			key := types.NamespacedName{
				Name:      "action-without-input",
				Namespace: "default",
			}

			created := fixActionCR(key, nil)

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
				if action.Status.Rendering.Input != nil {
					return errors.New(".Status.Rendering.Input field should be empty")
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})

		It("should render the action workflow with user input parameters", func() {
			key := types.NamespacedName{
				Name:      "action-with-input",
				Namespace: "default",
			}

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "user-input",
					Namespace: key.Namespace,
				},
				StringData: map[string]string{
					graphqldomain.ParametersSecretDataKey: `{ "message": { "pico": "bello"}}`,
				},
			}

			action := fixActionCR(key, secret)

			// Simulate that Secret with user input is created
			Expect(k8sClient.Create(context.Background(), secret)).Should(Succeed())
			// Simulate that Action CR is created
			Expect(k8sClient.Create(context.Background(), action)).Should(Succeed())

			Eventually(func() error {
				action := &corev1alpha1.Action{}

				if err := k8sClient.Get(context.Background(), key, action); err != nil {
					return err
				}

				if action.Status.Rendering == nil || action.Status.Rendering.Action == nil {
					return errors.New(".Status.Rendering.Action field is empty")
				}
				if action.Status.Rendering.Input == nil || action.Status.Rendering.Input.Parameters == nil {
					return errors.New(".Status.Rendering.Input.Parameters field is empty")
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})
	})
})

func fixActionCR(key types.NamespacedName, secret *corev1.Secret) *corev1alpha1.Action {
	action := &corev1alpha1.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: corev1alpha1.ActionSpec{
			ActionRef: corev1alpha1.ManifestReference{
				Path: "cap.interface.anything",
			},
		},
	}

	if secret != nil {
		action.Spec.Input = &corev1alpha1.ActionInput{
			Parameters: &corev1alpha1.InputParameters{
				SecretRef: corev1.LocalObjectReference{Name: secret.Name},
			},
		}
	}

	return action
}
