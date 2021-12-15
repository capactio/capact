package capact

import (
	"context"
	"fmt"

	"github.com/avast/retry-go"
	"github.com/jetstack/cert-manager/pkg/api/util"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certmanager "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	k8sretry "k8s.io/client-go/util/retry"
)

// ApplyClusterIssuer creates or, if it already exists, updates a ClusterIssuer for cert-manager.
func ApplyClusterIssuer(ctx context.Context, config *rest.Config, new *certv1.ClusterIssuer) error {
	clientset, err := certmanager.NewForConfig(config)
	if err != nil {
		return err
	}

	cli := clientset.ClusterIssuers()
	err = retryCreatingClusterIssuer(ctx, cli, new)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}

		err = retryUpdatingClusterIssuer(ctx, cli, new)
		if err != nil {
			return errors.Wrapf(err, "while updating the ClusterIssuer %s", new.Name)
		}
	}

	return waitForClusterIssuer(ctx, cli, new.Name)
}

func retryCreatingClusterIssuer(ctx context.Context, cli certmanager.ClusterIssuerInterface, new *certv1.ClusterIssuer) error {
	return retryForFn(
		func() error {
			_, err := cli.Create(ctx, new, metav1.CreateOptions{})
			return err
		},
		retry.RetryIf(func(err error) bool {
			return !apierrors.IsAlreadyExists(err)
		}),
		retry.LastErrorOnly(true),
	)
}

func retryUpdatingClusterIssuer(ctx context.Context, cli certmanager.ClusterIssuerInterface, new *certv1.ClusterIssuer) error {
	return k8sretry.RetryOnConflict(k8sretry.DefaultRetry, func() error {
		old, err := cli.Get(ctx, new.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "while getting the ClusterIssuer %s", old.Name)
		}

		old.Spec = new.Spec
		_, updateErr := cli.Update(ctx, old, metav1.UpdateOptions{})
		return updateErr
	})
}

func waitForClusterIssuer(ctx context.Context, cli certmanager.ClusterIssuerInterface, name string) error {
	return retryForFn(func() error {
		got, err := cli.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		readyCond := certv1.IssuerCondition{
			Type:   certv1.IssuerConditionReady,
			Status: certmeta.ConditionTrue,
		}

		if !util.IssuerHasCondition(got, readyCond) {
			return fmt.Errorf("ClusterIssuer %q is not ready", name)
		}

		return nil
	})
}
