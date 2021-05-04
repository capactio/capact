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

type RunOptions struct {
	ActionName string `survey:"name"`
	Namespace  string `survey:"namespace"`
}

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

	server, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

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
