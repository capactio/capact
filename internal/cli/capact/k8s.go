package capact

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

func CreateNamespace(config *rest.Config, namespace string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nsName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(nsName)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func AnnotateSecret(config *rest.Config, secretName, namespace, key, val string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	secretsClient := clientset.CoreV1().Secrets(namespace)

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Secret before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		secret, getErr := secretsClient.Get(secretName, metav1.GetOptions{})
		if getErr != nil {
			return errors.Wrapf(getErr, "while getting the secret %s", secretName)
		}

		if secret.ObjectMeta.Annotations == nil {
			secret.ObjectMeta.Annotations = map[string]string{}
		}
		secret.ObjectMeta.Annotations[key] = val

		_, updateErr := secretsClient.Update(secret)
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "while updating the secret %s", secretName)
	}
	return nil
}

func CreateUpdateSecret(config *rest.Config, newSecret *corev1.Secret, namespace string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	secretsClient := clientset.CoreV1().Secrets(namespace)
	_, err = secretsClient.Create(newSecret)
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Secret before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		secret, getErr := secretsClient.Get(newSecret.Name, metav1.GetOptions{})
		if getErr != nil {
			return errors.Wrapf(getErr, "while getting the secret %s", secret.Name)
		}

		secret.Data = newSecret.Data
		_, updateErr := secretsClient.Update(secret)
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "while updating the secret %s", newSecret.Name)
	}

	return nil
}
