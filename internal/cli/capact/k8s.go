package capact

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

// CreateNamespace creates a k8s namespaces. If it already exists it does nothing.
func CreateNamespace(ctx context.Context, config *rest.Config, namespace string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	nsName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err = clientset.CoreV1().Namespaces().Create(ctx, nsName, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// AnnotateSecret adds an annotation to the Secret.
func AnnotateSecret(ctx context.Context, config *rest.Config, secretName, namespace, key, val string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	secretsClient := clientset.CoreV1().Secrets(namespace)

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Secret before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		secret, getErr := secretsClient.Get(ctx, secretName, metav1.GetOptions{})
		if getErr != nil {
			return errors.Wrapf(getErr, "while getting the secret %s", secretName)
		}

		if secret.ObjectMeta.Annotations == nil {
			secret.ObjectMeta.Annotations = map[string]string{}
		}
		secret.ObjectMeta.Annotations[key] = val

		_, updateErr := secretsClient.Update(ctx, secret, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "while updating the secret %s", secretName)
	}
	return nil
}

// ApplySecret creates or, if it already exists, updates a secret.
func ApplySecret(ctx context.Context, config *rest.Config, newSecret *corev1.Secret, namespace string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	secretsClient := clientset.CoreV1().Secrets(namespace)
	_, err = secretsClient.Create(ctx, newSecret, metav1.CreateOptions{})
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Secret before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		secret, getErr := secretsClient.Get(ctx, newSecret.Name, metav1.GetOptions{})
		if getErr != nil {
			return errors.Wrapf(getErr, "while getting the secret %s", secret.Name)
		}

		secret.Data = newSecret.Data
		_, updateErr := secretsClient.Update(ctx, secret, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "while updating the secret %s", newSecret.Name)
	}

	return nil
}
