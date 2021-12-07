package capact

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apihelpers"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

// LoadCRDDefinition loads the CRD from a given location. Supports local file or URL.
func LoadCRDDefinition(location string) (*apiextensionv1.CustomResourceDefinition, error) {
	var reader io.Reader
	if isLocalFile(location) {
		f, err := os.Open(filepath.Clean(location))
		if err != nil {
			return nil, errors.Wrapf(err, "while opening local CRD file%s", location)
		}
		defer f.Close()
		reader = f
	} else {
		// #nosec G107
		resp, err := http.Get(location)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting CRD %s", location)
		}
		defer resp.Body.Close()
		reader = resp.Body
	}

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading CRD %s", location)
	}

	scheme := runtime.NewScheme()
	if err := apiextensionv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	actionCRD := &apiextensionv1.CustomResourceDefinition{}
	if err := runtime.DecodeInto(decoder, content, actionCRD); err != nil {
		return nil, err
	}

	return actionCRD, nil
}

// ApplyCRD creates or, if it already exists, updates a CRD.
func ApplyCRD(ctx context.Context, config *rest.Config, new *apiextensionv1.CustomResourceDefinition) error {
	clientset, err := apiextension.NewForConfig(config)
	if err != nil {
		return err
	}

	crdClient := clientset.ApiextensionsV1().CustomResourceDefinitions()
	_, err = crdClient.Create(ctx, new, metav1.CreateOptions{})
	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		old, err := crdClient.Get(ctx, new.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "while getting the CRD %s", old.Name)
		}

		old.Spec = new.Spec
		_, updateErr := crdClient.Update(ctx, old, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return errors.Wrapf(retryErr, "while updating the CRD %s", new.Name)
	}

	return waitForCRD(ctx, crdClient, new.Name)
}

func waitForCRD(ctx context.Context, crdClient v1.CustomResourceDefinitionInterface, name string) error {
	return retryForFn(func() error {
		crd, err := crdClient.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if !apihelpers.IsCRDConditionTrue(crd, apiextensionv1.Established) {
			return errors.New("CRD is not active")
		}

		if !apihelpers.IsCRDConditionTrue(crd, apiextensionv1.NamesAccepted) {
			return errors.New("the CRD names were not accepted")
		}

		return nil
	})
}

func isLocalFile(in string) bool {
	f, err := os.Stat(in)
	return err == nil && !f.IsDir()
}
