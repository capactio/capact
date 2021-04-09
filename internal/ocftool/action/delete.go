package action

import (
	"context"
	"io"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"

	"github.com/fatih/color"
)

type DeleteOptions struct {
	ActionName string
	Namespace  string
}

func Delete(ctx context.Context, opts DeleteOptions, w io.Writer) error {
	server, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)
	err = actionCli.DeleteAction(ctxWithNs, opts.ActionName)
	if err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintlnFunc()
	okCheck(w, "Action deleted successfully\n")

	return nil
}
