package action

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"

	"k8s.io/apimachinery/pkg/util/wait"
)

const deletionCheckPollInterval = time.Second

type DeleteOptions struct {
	ActionNames []string
	Namespace   string
	NameRegex   string
	Phase       string
	Timeout     time.Duration
	Wait        bool
}

func (d *DeleteOptions) Validate() error {
	if len(d.ActionNames) == 0 && d.NameRegex == "" {
		return ErrMissingActionToDeleteOpt
	}

	if len(d.ActionNames) > 0 && d.NameRegex != "" {
		return ErrMutuallyExclusiveOpts
	}

	if d.Phase != "" && d.NameRegex == "" {
		return ErrNotSupportedPhaseOpt
	}

	return nil
}

var (
	ErrMissingActionToDeleteOpt = errors.New("exact name, or regex option need to be specified")
	ErrMutuallyExclusiveOpts    = errors.New("exact name cannot be provided when regex option is specified")
	ErrNoActionToDelete         = errors.New("no Action to delete")
	ErrNotSupportedPhaseOpt     = errors.New("phase filter is supported only when regex option is used")
)

func Delete(ctx context.Context, opts DeleteOptions, w io.Writer) (err error) {
	status := printer.NewStatus(w, "")
	defer func() {
		status.End(err == nil)
	}()

	if err := opts.Validate(); err != nil {
		return err
	}

	server := config.GetDefaultContext()

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)

	actionsToDelete := opts.ActionNames
	if opts.NameRegex != "" {
		actionsToDelete, err = listActionsForDeletion(ctxWithNs, actionCli, opts)
		if err != nil {
			return err
		}
	}

	if len(actionsToDelete) == 0 {
		return ErrNoActionToDelete
	}

	for _, name := range actionsToDelete {
		status.Step("Scheduling Action '%s/%s' deletion", opts.Namespace, name)
		if err = actionCli.DeleteAction(ctxWithNs, name); err != nil {
			return err
		}
	}

	if !opts.Wait {
		return nil
	}

	status.Step("Waiting ≤ %s for deletion process to complete", opts.Timeout)
	return waitUntilDeleted(ctxWithNs, actionCli, actionsToDelete, opts.Timeout)
}

func waitUntilDeleted(ctxWithNs context.Context, actCli client.ClusterClient, names []string, timeout time.Duration) error {
	toBeDeletedNameRegex := mapToStrictOrRegex(names)

	var lastErr error
	err := wait.Poll(deletionCheckPollInterval, timeout, func() (done bool, err error) {
		out, err := actCli.ListActions(ctxWithNs, &gqlengine.ActionFilter{
			NameRegex: &toBeDeletedNameRegex,
		})
		if err != nil { // may be network issue, ignoring
			lastErr = err
			return false, nil
		}
		if len(out) != 0 {
			lastErr = fmt.Errorf("the following Actions are still not deleted: [ %s ]", strings.Join(toNamesList(out), ", "))
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return lastErr
		}
		return err
	}

	return nil
}

func toNamesList(in []*gqlengine.Action) []string {
	var names []string
	for _, i := range in {
		names = append(names, i.Name)
	}
	return names
}

func mapToStrictOrRegex(in []string) string {
	out := strings.Join(in, "$|^")
	return fmt.Sprintf("(^%s$)", out)
}

func AllowedPhases() string {
	var out []string
	for _, p := range gqlengine.AllActionStatusPhase {
		out = append(out, string(p))
	}
	return strings.Join(out, ", ")
}

func listActionsForDeletion(ctxWithNs context.Context, actionCli client.ClusterClient, opts DeleteOptions) ([]string, error) {
	var phase *gqlengine.ActionStatusPhase
	if opts.Phase != "" {
		p := gqlengine.ActionStatusPhase(opts.Phase)
		if !p.IsValid() {
			return nil, fmt.Errorf("not valid phase option, allowed values: %s", AllowedPhases())
		}
		phase = &p
	}

	out, err := actionCli.ListActions(ctxWithNs, &gqlengine.ActionFilter{
		Phase:     phase,
		NameRegex: &opts.NameRegex,
	})
	if err != nil {
		return nil, err
	}

	return toNamesList(out), nil
}
