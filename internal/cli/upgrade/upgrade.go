package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	"capact.io/capact/internal/ptr"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/avast/retry-go"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	capactUpgradeInterfacePath = "cap.interface.capactio.capact.upgrade"
	capactTypeRefPath          = "cap.type.capactio.capact.config"
	helmReleaseTypeRefPath     = "cap.type.helm.chart.release"

	randomSuffixLength = 5
	pollInterval       = time.Second
)

// Options holds configuration for Capact upgrade operation.
type Options struct {
	Timeout          time.Duration
	Wait             bool
	Parameters       capact.InputParameters
	ActionNamePrefix string

	MaxQueueTime time.Duration
}

// Upgrade provides functionality to upgrade Capact cluster to a given version.
type Upgrade struct {
	hubCli client.Hub
	actCli client.ClusterClient
	writer io.Writer
}

// New returns a new Upgrade instance.
func New(w io.Writer) (*Upgrade, error) {
	server := config.GetDefaultContext()

	hubCli, err := client.NewHub(server)
	if err != nil {
		return nil, errors.Wrap(err, "while creating hub client")
	}

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return nil, errors.Wrap(err, "while creating cluster client")
	}

	return &Upgrade{
		hubCli: hubCli,
		actCli: actionCli,
		writer: w,
	}, nil
}

// Run executes Capact cluster upgrade action with a input configuration.
func (u *Upgrade) Run(ctx context.Context, opts Options) (err error) {
	status := printer.NewStatus(u.writer, "Upgrading Capact on cluster...")
	defer func() {
		status.End(err == nil)
	}()

	ctxWithNs := namespace.NewContext(ctx, capact.Namespace)

	status.Step("Waiting for other upgrade actions to complete...")
	if err := u.waitForOtherUpgradesToComplete(ctxWithNs, opts); err != nil {
		return errors.Wrap(err, "while waiting for other upgrade actions to complete")
	}

	status.Step("Getting Capact config 📜")
	capactCfg, err := u.getCapactConfigTypeInstance(ctx)
	if err != nil {
		return err
	}

	if err := u.resolveInputParameters(&opts); err != nil {
		return err
	}

	status.Step("Creating upgrade Action for %s 💾", opts.Parameters.Version)

	inputParams, err := mapToInputParameters(opts.Parameters)
	if err != nil {
		return err
	}

	inputTI, err := mapToInputTypeInstances(capactCfg)
	if err != nil {
		return err
	}

	act, err := u.createAction(ctxWithNs, generateActionName(opts.ActionNamePrefix), inputParams, inputTI)
	if err != nil {
		return err
	}

	status.Step("Rendering upgrade Action 📽️")
	err = u.waitUntilReadyToRun(ctxWithNs, act.Name, opts.Timeout)
	if err != nil {
		return err
	}

	status.Step("Running upgrade Action 🏃")
	if err = u.actCli.RunAction(ctxWithNs, act.Name); err != nil {
		return err
	}

	if !opts.Wait {
		status.Step("Action '%s/%s' successfully scheduled.", capact.Namespace, act.Name)
		return nil
	}

	status.Step("Waiting ≤ %s for the '%s/%s' Action to finish 🏁", opts.Timeout, capact.Namespace, act.Name)
	err = u.waitUntilFinished(ctxWithNs, act.Name, opts.Timeout)
	if err != nil {
		return err
	}

	return nil
}

func (u *Upgrade) waitForOtherUpgradesToComplete(ctxWithNs context.Context, opts Options) error {
	var (
		attempts   uint
		ctx        context.Context
		retryDelay = 5 * time.Second
	)

	if opts.MaxQueueTime == 0 {
		attempts = 1
		ctx = ctxWithNs
	} else {
		attempts = 1e6
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctxWithNs, opts.MaxQueueTime)
		defer cancel()
	}

	err := retry.Do(func() error {
		actions, err := u.getRunningUpgradeActions(ctxWithNs)
		if err != nil {
			return errors.Wrap(err, "while getting running upgrade actions")
		}

		if len(actions) > 0 {
			return NewErrAnotherUpgradeIsRunning(actions[0].Name)
		}

		return nil
	}, retry.Delay(retryDelay),
		retry.Attempts(attempts),
		retry.Context(ctx),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
	)

	return errors.Wrap(err, "Timeout waiting for another upgrade action to finish.")
}

func (u *Upgrade) getRunningUpgradeActions(nsCtx context.Context) ([]*gqlengine.Action, error) {
	actions, err := u.actCli.ListActions(nsCtx, &gqlengine.ActionFilter{
		InterfaceRef: &gqlengine.ManifestReferenceInput{
			Path: capactUpgradeInterfacePath,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while listing existing upgrade actions")
	}

	var runningActions []*gqlengine.Action
	for i := range actions {
		action := actions[i]
		if action.Status.Phase != gqlengine.ActionStatusPhaseRunning {
			continue
		}
		runningActions = append(runningActions, action)
	}

	return runningActions, nil
}

func (u *Upgrade) resolveInputParameters(opts *Options) error {
	err := opts.Parameters.ResolveVersion()
	if err != nil {
		return err
	}

	if opts.Parameters.IncreaseResourceLimits {
		opts.Parameters.Override.CapactValues.Gateway.Resources = capact.IncreasedGatewayResources()
		opts.Parameters.Override.CapactValues.HubPublic.Resources = capact.IncreasedHubPublicResources()
		opts.Parameters.Override.CapactValues.HubLocal.Resources = capact.IncreasedHubLocalResources()
		opts.Parameters.Override.Neo4jValues.Neo4j.Core.Resources = capact.IncreasedNeo4jResources()
	}

	return nil
}

func (u *Upgrade) getCapactConfigTypeInstance(ctx context.Context) (gqllocalapi.TypeInstance, error) {
	capactCfg, err := u.hubCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		TypeRef: &gqllocalapi.TypeRefFilterInput{
			Path:     capactTypeRefPath,
			Revision: ptr.String("0.1.0"),
		},
	})
	if err != nil {
		return gqllocalapi.TypeInstance{}, err
	}

	if len(capactCfg) != 1 {
		return gqllocalapi.TypeInstance{}, errors.Errorf("Unexpected number of Capact config TypeInstance, expected 1, got %d", len(capactCfg))
	}

	return capactCfg[0], nil
}

func (u *Upgrade) createAction(ctx context.Context, name string, inputParams gqlengine.JSON, inputTI []*gqlengine.InputTypeInstanceData) (*gqlengine.Action, error) {
	act, err := u.actCli.CreateAction(ctx, &gqlengine.ActionDetailsInput{
		Name: name,
		Input: &gqlengine.ActionInputData{
			Parameters:    &inputParams,
			TypeInstances: inputTI,
		},
		ActionRef: &gqlengine.ManifestReferenceInput{
			Path: capactUpgradeInterfacePath,
		},
	})
	return act, err
}

func generateActionName(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, utilrand.String(randomSuffixLength))
}

func (u *Upgrade) waitUntilReadyToRun(ctx context.Context, name string, timeout time.Duration) error {
	var lastErr error
	err := wait.Poll(pollInterval, timeout, func() (done bool, err error) {
		act, err := u.actCli.GetAction(ctx, name)
		if err != nil { // may be network issue, ignoring
			lastErr = err
			return false, nil
		}

		if act == nil {
			return true, ErrActionDeleted
		}

		if act.Status == nil {
			lastErr = ErrActionWithoutStatus
			return false, nil
		}
		switch act.Status.Phase {
		case gqlengine.ActionStatusPhaseReadyToRun:
			return true, nil
		case gqlengine.ActionStatusPhaseCanceled, gqlengine.ActionStatusPhaseFailed:
			lastErr = fmt.Errorf("unexpected Action state, expected %s, got %s%s", gqlengine.ActionStatusPhaseReadyToRun, act.Status.Phase, printMessage(act.Status.Message))
			return true, lastErr
		default:
			lastErr = ErrActionNotFinished
			return false, nil
		}
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return lastErr
		}
		return err
	}

	return nil
}

// printMessage prints message if provided. Message is printed in new line.
func printMessage(message *string) string {
	if message == nil {
		return ""
	}
	return fmt.Sprintf(".\n Message: %q", *message)
}

func (u *Upgrade) waitUntilFinished(ctx context.Context, name string, timeout time.Duration) error {
	var lastErr error
	err := wait.Poll(pollInterval, timeout, func() (done bool, err error) {
		act, err := u.actCli.GetAction(ctx, name)
		if err != nil {
			lastErr = err
			return false, nil
		}

		if act == nil {
			// Action has been deleted, no reason to wait further
			// as an Action can be deleted only once it's completed.
			return true, ErrActionDeleted
		}
		if act.Status == nil {
			// Status not available, this shouldn't happen for running action
			return false, nil
		}
		switch act.Status.Phase {
		case gqlengine.ActionStatusPhaseSucceeded:
			return true, nil
		case gqlengine.ActionStatusPhaseCanceled, gqlengine.ActionStatusPhaseFailed:
			lastErr = fmt.Errorf("Unexpected Action state, expected %s, got %s", gqlengine.ActionStatusPhaseSucceeded, act.Status.Phase)
			return true, lastErr
		default:
			lastErr = ErrActionNotFinished
			return false, nil
		}
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return lastErr
		}
		return err
	}

	return nil
}

// mapToInputTypeInstances converts capactCfg.Uses into input TypeInstance required for upgrade Action.
// Returned TypeInstance:
// - capact-config
// - capact-helm-release
// - argo-helm-release
// - ingress-nginx-helm-release
// - kubed-helm-release
// - monitoring-helm-release
// - neo4j-helm-release
// - cert-manager-helm-release
func mapToInputTypeInstances(capactCfg gqllocalapi.TypeInstance) ([]*gqlengine.InputTypeInstanceData, error) {
	inputTI := []*gqlengine.InputTypeInstanceData{
		{
			Name: fmt.Sprintf("%s-config", capact.Name),
			ID:   capactCfg.ID,
		},
	}

	for _, ti := range capactCfg.Uses {
		unpacked := struct {
			Name string
		}{}

		if err := validateTypeInstance(ti); err != nil {
			return nil, err
		}

		err := mapstructure.Decode(ti.LatestResourceVersion.Spec.Value, &unpacked)
		if err != nil {
			return nil, err
		}

		inputTI = append(inputTI, &gqlengine.InputTypeInstanceData{
			Name: fmt.Sprintf("%s-helm-release", unpacked.Name),
			ID:   ti.ID,
		})
	}
	return inputTI, nil
}

func mapToInputParameters(params capact.InputParameters) (gqlengine.JSON, error) {
	marshalled, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	return gqlengine.JSON(marshalled), nil
}

func validateTypeInstance(ti *gqllocalapi.TypeInstance) error {
	if ti == nil || ti.LatestResourceVersion == nil || ti.LatestResourceVersion.Spec == nil || ti.LatestResourceVersion.Spec.Value == nil {
		return errors.New("TypeInstance.LatestResourceVersion.Spec.Value cannot be nil")
	}

	if ti.TypeRef == nil {
		return errors.New("TypeInstance.TypeRef cannot be nil")
	}

	if ti.TypeRef.Path != helmReleaseTypeRefPath {
		return fmt.Errorf("unexpected TypeRef, expected %q, got %q", helmReleaseTypeRefPath, ti.TypeRef.Path)
	}
	return nil
}
