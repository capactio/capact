package action

import (
	"context"
	"io"

	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	"capact.io/capact/internal/ocftool/client"
	"capact.io/capact/internal/ocftool/config"
	"capact.io/capact/internal/ptr"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	"github.com/fatih/color"
)

type CreateOutput struct {
	Action    *gqlengine.Action
	Namespace string
}

func Create(ctx context.Context, opts CreateOptions, w io.Writer) (*CreateOutput, error) {
	if err := opts.Resolve(); err != nil {
		return nil, err
	}

	server, err := config.GetDefaultContext()
	if err != nil {
		return nil, err
	}

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return nil, err
	}

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)
	act, err := actionCli.CreateAction(ctxWithNs, &gqlengine.ActionDetailsInput{
		Name:  opts.ActionName,
		Input: opts.ActionInput(),
		ActionRef: &gqlengine.ManifestReferenceInput{
			Path: opts.InterfacePath,
		},
		DryRun: ptr.Bool(opts.DryRun),
	})
	if err != nil {
		return nil, err
	}

	okCheck := color.New(color.FgGreen).FprintfFunc()
	okCheck(w, "Action %s/%s created successfully\n", opts.Namespace, act.Name)

	return &CreateOutput{
		Action:    act,
		Namespace: opts.Namespace,
	}, nil
}
