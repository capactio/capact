package action

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"

	"github.com/fatih/color"
)

type DeleteOptions struct {
	ActionNames []string
	Namespace   string
	NameRegex   string
	Phase       string
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

func Delete(ctx context.Context, opts DeleteOptions, w io.Writer) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintfFunc()

	server, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

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
		err = actionCli.DeleteAction(ctxWithNs, name)
		if err != nil {
			return err
		}
		okCheck(w, "Action '%s/%s' deleted successfully\n", opts.Namespace, name)
	}

	return nil
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
	var names []string
	for _, i := range out {
		names = append(names, i.Name)
	}

	return names, nil
}
