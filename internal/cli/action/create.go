package action

import (
	"context"
	"fmt"
	"io"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	"capact.io/capact/internal/ptr"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/validate/action"

	"github.com/fatih/color"
)

// CreateOutput defines output for Create function.
type CreateOutput struct {
	Action    *gqlengine.Action
	Namespace string
}

// Create creates a given Action.
func Create(ctx context.Context, opts CreateOptions, w io.Writer) (*CreateOutput, error) {
	server := config.GetDefaultContext()
	hubCli, err := client.NewHub(server)
	if err != nil {
		return nil, err
	}

	validator := action.NewValidator(hubCli)
	opts.validator = validator

	// TODO: In the future, we can use client.PolicyEnforcedClient
	// to get the Implementation and validate Implementation specific TypeInstances and additional input.
	// That would require some unification and re-using exactly the same logic for the Impl resolution.
	// For now, fetch latest - the same strategy is used by renderer.
	iface, err := hubCli.FindInterfaceRevision(ctx, gqlpublicapi.InterfaceReference{
		Path: opts.InterfacePath,
	})
	if err != nil {
		return nil, err
	}
	if iface == nil {
		return nil, fmt.Errorf("Interface %s was not found in Hub", opts.InterfacePath)
	}

	opts.ifaceSchemas, err = validator.LoadIfaceInputParametersSchemas(ctx, iface)
	if err != nil {
		return nil, err
	}
	opts.isInputParamsRequired, err = validator.HasRequiredProp(opts.ifaceSchemas)
	if err != nil {
		return nil, err
	}

	opts.ifaceTypes, err = validator.LoadIfaceInputTypeInstanceRefs(ctx, iface)
	if err != nil {
		return nil, err
	}
	opts.isInputTypesRequired = len(opts.ifaceTypes) > 0

	if err := opts.resolve(ctx); err != nil {
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
