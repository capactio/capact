package helmstoragebackend

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"capact.io/capact/internal/ptr"
)

// NewHelmGet create a new Helm release get client.
func NewHelmGet(flags *genericclioptions.ConfigFlags, driver, ns string) (*action.Get, error) {
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

	return action.NewGet(actionConfig), nil
}
