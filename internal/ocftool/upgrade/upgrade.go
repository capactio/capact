package upgrade

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
	"io"
	"io/ioutil"
	"net/http"
	"sigs.k8s.io/yaml"
	"time"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ptr"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"

	"github.com/mitchellh/mapstructure"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	LatestVersionTag = "@latest"

	capactSystemNS = "voltron-system"
	capactOldName  = "voltron"

	capactUpgradeInterfacePath = "cap.interface.capactio.capact.upgrade"
	capactTypeRefPath          = "cap.type.capactio.capact.config"
	helmReleaseTypeRefPath     = "cap.type.helm.chart.release"

	randomSuffixLength = 5
	pollInterval       = time.Second

	capactioHelmRepoIndexURL = "https://capactio-awesome-charts.storage.googleapis.com/index.yaml"
)

var actionNotFinishedErr = errors.New("Action still not finished")

type Options struct {
	Version                string
	Timeout                time.Duration
	Wait                   bool
	IncreaseResourceLimits bool
}

type Upgrade struct {
	hubCli client.Hub
	actCli client.ClusterClient
	writer io.Writer
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

	return &Upgrade{
		hubCli: hubCli,
		actCli: actionCli,
		writer: w,
	}, nil
}

func (u *Upgrade) Run(ctx context.Context, opts Options) (err error) {
	status := ocftool.NewStatusPrinter(u.writer, "Upgrading Capact on cluster...")
	defer func() {
		status.End(err == nil)
	}()

	status.Step("Getting Capact config 📜")
	capactCfg, err := u.getCapactConfigTypeInstance(ctx)
	if err != nil {
		return err
	}

	if opts.Version == LatestVersionTag {
		ver, err := getLatestVersion()
		if err != nil {
			return err
		}
		opts.Version = ver
	}

	status.Step("Creating upgrade Action for %s 💾", opts.Version)
	var (
		inputParams = mapToInputParameters(opts)
		ctxWithNs   = namespace.NewContext(ctx, capactSystemNS)
	)

	inputTI, err := mapToInputTypeInstances(capactCfg)
	if err != nil {
		return err
	}

	act, err := u.createAction(ctxWithNs, inputParams, inputTI)
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
		status.Step("Action %q successfully scheduled.", act.Name)
		return nil
	}

	status.Step("Waiting ≤ %s for upgrade Action finish 🏁", opts.Timeout)
	err = u.waitUntilFinish(ctxWithNs, act.Name, opts.Timeout)
	if err != nil {
		return err
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
		return gqllocalapi.TypeInstance{}, err
	}

	return capactCfg[0], nil
}

func (u *Upgrade) createAction(ctx context.Context, inputParams gqlengine.JSON, inputTI []*gqlengine.InputTypeInstanceData) (*gqlengine.Action, error) {
	act, err := u.actCli.CreateAction(ctx, &gqlengine.ActionDetailsInput{
		// TODO: should we support server-side GenerateName parameter?
		Name: generateActionName(),
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

func generateActionName() string {
	return fmt.Sprintf("capact-upgrade-%s", utilrand.String(randomSuffixLength))
}

func (u *Upgrade) waitUntilReadyToRun(ctx context.Context, name string, timeout time.Duration) error {
	var lastErr error
	err := wait.Poll(pollInterval, timeout, func() (done bool, err error) {
		act, err := u.actCli.GetAction(ctx, name)
		if err != nil { // may be network issue, ignoring
			lastErr = err
			return false, nil
		}
		switch act.Status.Phase {
		case gqlengine.ActionStatusPhaseReadyToRun:
			return true, nil
		default:
			lastErr = actionNotFinishedErr
			return false, nil
		}
	})
	if err != nil {
		return lastErr
	}

	return nil
}

func (u *Upgrade) waitUntilFinish(ctx context.Context, name string, timeout time.Duration) error {
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
			lastErr = actionNotFinishedErr
			return false, nil
		}
	})
	if err != nil {
		return lastErr
	}
	return nil
}

// mapToInputTypeInstances converts capactCfg.Uses into input TypeInstance required for upgrade Action.
// Returned TypeInstance:
// - capact-helm-release
// - argo-helm-release
// - ingress-nginx-helm-release
// - kubed-helm-release
// - monitoring-helm-release
// - neo4j-helm-release
func mapToInputTypeInstances(capactCfg gqllocalapi.TypeInstance) ([]*gqlengine.InputTypeInstanceData, error) {
	inputTI := []*gqlengine.InputTypeInstanceData{
		{Name: "capact-config", ID: capactCfg.ID},
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

		if unpacked.Name == capactOldName {
			unpacked.Name = "capact"
		}

		inputTI = append(inputTI, &gqlengine.InputTypeInstanceData{
			Name: fmt.Sprintf("%s-helm-release", unpacked.Name),
			ID:   ti.ID,
		})

	}
	return inputTI, nil
}

func mapToInputParameters(opts Options) gqlengine.JSON {
	return gqlengine.JSON(fmt.Sprintf(`{"version": "%s", "increaseResourceLimits": %t}`, opts.Version, opts.IncreaseResourceLimits))
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

// loadIndex loads an index file and does minimal validity checking.
// Assumption that all charts are versioned in the same way.
func getLatestVersion() (string, error) {
	resp, err := http.Get(capactioHelmRepoIndexURL)
	if err != nil {
		return "", errors.Wrap(err, "while getting capactio Helm Chart repository index.yaml")
	}
	defer resp.Body.Close()

	// TODO(mszostok): read with fixed size, so we will not blow up app if request is malformed
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	i := &repo.IndexFile{}
	if err := yaml.UnmarshalStrict(data, i); err != nil {
		return "", err
	}
	i.SortEntries()

	return i.Entries[capactOldName][0].Version, nil
}
