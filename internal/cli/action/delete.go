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

// DeleteOptions holds configuration for Action deletion.
type DeleteOptions struct {
	ActionNames []string
	Namespace   string
	NameRegex   string
	Phase       string
	Timeout     time.Duration
	Wait        bool
}

// Validate validates if provided delete options are valid.
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
	// ErrMissingActionToDeleteOpt defines error indicating that at least Action name or Action regex name needs to be provided.
	ErrMissingActionToDeleteOpt = errors.New("exact name, or regex option need to be specified")
	// ErrMutuallyExclusiveOpts defines error indicating that Action name and Action regex name cannot be provided at the same time.
	ErrMutuallyExclusiveOpts = errors.New("exact name cannot be provided when regex option is specified")
	// ErrNoActionToDelete defines error indicating there are no Action to be deleted.
	ErrNoActionToDelete = errors.New("no Action to delete")
	// ErrNotSupportedPhaseOpt defines error indicating that you can filter Actions by the `phase` field only when Action regex name is used.
	ErrNotSupportedPhaseOpt = errors.New("phase filter is supported only when regex option is used")
)

// Delete schedules Action deletion. If requested, wait for deletion process to complete.
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

// AllowedPhases returns string with all possible Action phases separated by comma.
func AllowedPhases() string {
	var out []string
	for _, p := range gqlengine.AllActionStatusPhase {
		out = append(out, string(p))
	}
	return strings.Join(out, ", ")
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
