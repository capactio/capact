package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/printer"
	"projectvoltron.dev/voltron/internal/ptr"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/httputil"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"
)

const (
	LatestVersionTag = "@latest"

	capactSystemNS = "voltron-system"
	capactOldName  = "voltron"
	capactName     = "capact"

	capactioHelmRepoMaster    = "https://storage.googleapis.com/capactio-master-charts"
	CapactioHelmRepoOfficial  = "https://storage.googleapis.com/capactio-awesome-charts"
	CapactioHelmRepoMasterTag = "@master"

	capactUpgradeInterfacePath = "cap.interface.capactio.capact.upgrade"
	capactTypeRefPath          = "cap.type.capactio.capact.config"
	helmReleaseTypeRefPath     = "cap.type.helm.chart.release"

	randomSuffixLength = 5
	pollInterval       = time.Second
)

var ErrActionNotFinished = errors.New("Action still not finished")
var ErrActionWithoutStatus = errors.New("Action doesn't have status")

type (
	Options struct {
		Timeout          time.Duration
		Wait             bool
		Parameters       InputParameters
		ActionNamePrefix string
	}
)

type Upgrade struct {
	hubCli     client.Hub
	actCli     client.ClusterClient
	writer     io.Writer
	httpClient *http.Client
}

func New(w io.Writer) (*Upgrade, error) {
	server, err := config.GetDefaultContext()
	if err != nil {
		return nil, errors.Wrap(err, "while getting default context")
	}

	hubCli, err := client.NewHub(server)
	if err != nil {
		return nil, errors.Wrap(err, "while creating hub client")
	}

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return nil, errors.Wrap(err, "while creating cluster client")
	}

	httpClient := httputil.NewClient(30 * time.Second)

	return &Upgrade{
		hubCli:     hubCli,
		actCli:     actionCli,
		httpClient: httpClient,
		writer:     w,
	}, nil
}

func (u *Upgrade) Run(ctx context.Context, opts Options) (err error) {
	status := printer.NewStatus(u.writer, "Upgrading Capact on cluster...")
	defer func() {
		status.End(err == nil)
	}()

	status.Step("Getting Capact config 📜")
	capactCfg, err := u.getCapactConfigTypeInstance(ctx)
	if err != nil {
		return err
	}

	if err := u.resolveInputParameters(&opts); err != nil {
		return err
	}

	status.Step("Creating upgrade Action for %s 💾", opts.Parameters.Version)
	ctxWithNs := namespace.NewContext(ctx, capactSystemNS)

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
		status.Step("Action '%s/%s' successfully scheduled.", capactSystemNS, act.Name)
		return nil
	}

	status.Step("Waiting ≤ %s for the '%s/%s' Action to finish 🏁", opts.Timeout, capactSystemNS, act.Name)
	err = u.waitUntilFinished(ctxWithNs, act.Name, opts.Timeout)
	if err != nil {
		return err
	}

	return nil
}

func (u *Upgrade) resolveInputParameters(opts *Options) error {
	if opts.Parameters.Override.HelmRepoURL == CapactioHelmRepoMasterTag {
		opts.Parameters.Override.HelmRepoURL = capactioHelmRepoMaster
	}

	if opts.Parameters.Version == LatestVersionTag {
		ver, err := u.getLatestVersion(opts.Parameters.Override.HelmRepoURL)
		if err != nil {
			return err
		}
		opts.Parameters.Version = ver
	}

	if opts.Parameters.IncreaseResourceLimits {
		opts.Parameters.Override.CapactValues.Gateway.Resources = increasedGatewayResources()
		opts.Parameters.Override.CapactValues.OCHPublic.Resources = increasedOCHPublicResources()
		opts.Parameters.Override.Neo4jValues.Neo4j.Core.Resources = increasedNeo4jResources()
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

// TODO: should we support server-side GenerateName parameter?
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
func mapToInputTypeInstances(capactCfg gqllocalapi.TypeInstance) ([]*gqlengine.InputTypeInstanceData, error) {
	inputTI := []*gqlengine.InputTypeInstanceData{
		{
			Name: fmt.Sprintf("%s-config", capactName),
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

		// TODO: remove after rebranding
		if unpacked.Name == capactOldName {
			unpacked.Name = capactName
		}

		inputTI = append(inputTI, &gqlengine.InputTypeInstanceData{
			Name: fmt.Sprintf("%s-helm-release", unpacked.Name),
			ID:   ti.ID,
		})
	}
	return inputTI, nil
}

func mapToInputParameters(params InputParameters) (gqlengine.JSON, error) {
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

// getLatestVersion loads an index file and returns version of the latest chart:
//	- for master Helm charts sort by Created field
//  - for all others repos sort by SemVer
//
// Assumption that all charts are versioned in the same way.
func (u *Upgrade) getLatestVersion(repoURL string) (string, error) {
	// by default sort by SemVer, so even if someone pushed bugfix of older
	// release we will not take it.
	sortFn := func(in *repo.IndexFile) {
		in.SortEntries()
	}

	// `master` charts are versioned via SHA commit, so we need to sort
	// them via Created time.
	if repoURL == capactioHelmRepoMaster {
		sortFn = func(in *repo.IndexFile) {
			sort.Sort(ByCreatedTime(in.Entries[capactOldName]))
		}
	}

	resp, err := u.httpClient.Get(fmt.Sprintf("%s/index.yaml", repoURL))
	if err != nil {
		return "", errors.Wrap(err, "while getting capactio Helm Chart repository index.yaml")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	i := &repo.IndexFile{}
	if err := yaml.UnmarshalStrict(data, i); err != nil {
		return "", err
	}
	sortFn(i)

	return i.Entries[capactOldName][0].Version, nil
}

type ByCreatedTime repo.ChartVersions

func (b ByCreatedTime) Len() int           { return len(b) }
func (b ByCreatedTime) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByCreatedTime) Less(i, j int) bool { return b[i].Created.Before(b[j].Created) }
