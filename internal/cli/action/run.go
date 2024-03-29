package action

import (
	"context"
	"io"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// RunOptions holds configuration for running Action.
type RunOptions struct {
	ActionName string `survey:"name"`
	Namespace  string `survey:"namespace"`
}

// Run executes a given Action. Possible only if Action is in the `READY_TO_RUN` phase.
func Run(ctx context.Context, opts RunOptions, w io.Writer) error {
	var qs []*survey.Question
	if opts.Namespace == "" {
		qs = append(qs, namespaceQuestion())
	}

	if opts.ActionName == "" {
		qs = append(qs, actionNameQuestion(""))
	}

	if err := survey.Ask(qs, &opts); err != nil {
		return err
	}

	server := config.GetDefaultContext()

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)
	err = actionCli.RunAction(ctxWithNs, opts.ActionName)
	if err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintlnFunc()
	okCheck(w, "Action run successfully\n")

	return nil
}
