package helmstoragebackend

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"capact.io/capact/internal/ptr"
)

const defaultHelmDriver = "secrets"

type (
	// HelmRelease holds details about Helm release.
	HelmRelease struct {
		// Name specifies Helm release name for a given request.
		Name string `json:"name"`
		// Namespace specifies in which Kubernetes Namespace Helm release is located.
		Namespace string `json:"namespace"`
		// Driver specifies drivers used for storing the Helm release.
		Driver *string `json:"driver,omitempty"`
	}

	actionConfigurationProducerFn func(flags *genericclioptions.ConfigFlags, driver string, ns string) (*action.Configuration, error)
)

// HelmReleaseFetcher provides functionality to fetch Helm release.
type HelmReleaseFetcher struct {
	helmCfgFlags                *genericclioptions.ConfigFlags
	actionConfigurationProducer actionConfigurationProducerFn
}

// NewHelmReleaseFetcher returns a new HelmReleaseFetcher instance.
func NewHelmReleaseFetcher(flags *genericclioptions.ConfigFlags) *HelmReleaseFetcher {
	return &HelmReleaseFetcher{helmCfgFlags: flags, actionConfigurationProducer: actionConfigurationProducer}
}

// FetchHelmRelease returns a given Helm release. It already handles the gRPC errors properly.
func (f *HelmReleaseFetcher) FetchHelmRelease(helmRelease HelmRelease, additionalErrMsg *string) (*release.Release, error) {
	cfg, err := f.actionConfigurationProducer(f.helmCfgFlags, *helmRelease.Driver, helmRelease.Namespace)
	if err != nil {
		return nil, gRPCInternalError(errors.Wrap(err, "while creating Helm get release client"))
	}

	helmGet := action.NewGet(cfg)

	// NOTE: req.resourceVersion is ignored on purpose.
	// Based on our contract we always return the latest Helm release revision.
	helmGet.Version = latestRevisionIndicator

	rel, err := helmGet.Run(helmRelease.Name)
	switch {
	case err == nil:
	case errors.Is(err, driver.ErrReleaseNotFound):
		var additionalErrCtx string
		if additionalErrMsg != nil {
			additionalErrCtx = fmt.Sprintf(" (%s)", *additionalErrMsg)
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Helm release '%s/%s'%s was not found", helmRelease.Namespace, helmRelease.Name, additionalErrCtx))
	default:
		return nil, gRPCInternalError(errors.Wrap(err, "while fetching Helm release"))
	}

	return rel, nil
}

// actionConfigurationProducer returns Configuration with a given input settings.
func actionConfigurationProducer(flags *genericclioptions.ConfigFlags, driver, ns string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	helmCfg := &genericclioptions.ConfigFlags{
		APIServer:   flags.APIServer,
		Insecure:    flags.Insecure,
		CAFile:      flags.CAFile,
		BearerToken: flags.BearerToken,
		Namespace:   ptr.String(ns),
	}

	debugLog := func(format string, v ...interface{}) {
		// noop
	}

	err := actionConfig.Init(helmCfg, ns, driver, debugLog)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Helm configuration")
	}

	return actionConfig, nil
}
