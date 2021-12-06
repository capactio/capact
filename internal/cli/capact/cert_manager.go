package capact

import (
	"context"

	"github.com/pkg/errors"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

// ApplyClusterIssuer creates or, if it already exists, updates a ClusterIssuer for cert-manager.
// TODO: ensure issuer is ready.
func ApplyClusterIssuer(ctx context.Context, config *rest.Config, new *certv1.ClusterIssuer) error {
	clientset, err := certmanager.NewForConfig(config)
	if err != nil {
		return err
	}

	cli := clientset.ClusterIssuers()
	_, err = cli.Create(ctx, new, metav1.CreateOptions{})
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		old, err := cli.Get(ctx, new.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "while getting the ClusterIssuer %s", old.Name)
		}

		old.Spec = new.Spec
		_, updateErr := cli.Update(ctx, old, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "while updating the ClusterIssuer %s", new.Name)
	}

	return nil
}
