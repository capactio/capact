package interfaces

import (
	"context"
	"fmt"
	"os"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/heredoc"
	cliprinter "capact.io/capact/internal/cli/printer"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"

	"github.com/spf13/cobra"
)

type getOptions struct {
	interfacePaths []string
}

var (
	allPathPrefix = "cap.interface.*"
)

// NewGet returns a cobra.Command for getting available Implementations in a Public Hub.
func NewGet() *cobra.Command {
	var opts getOptions

	resourcePrinter := cliprinter.NewForResource(os.Stdout, cliprinter.WithJSON(), cliprinter.WithYAML(), cliprinter.WithTable(tableDataOnGet))

	get := &cobra.Command{
		Use:   "get",
		Short: "Displays one or multiple Interfaces available on the Hub server",
		Example: heredoc.WithCLIName(`
			# Show all Interfaces in table format:
			<cli> hub interfaces get
			
			# Show "cap.interface.database.postgresql.install" Interface in JSON format:
			<cli> hub interfaces get cap.interface.database.postgresql.install -ojson
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.interfacePaths = args
			return listInterfaces(cmd.Context(), opts, resourcePrinter)
		},
	}

	flags := get.Flags()
	resourcePrinter.RegisterFlags(flags)

	return get
}

func listInterfaces(ctx context.Context, opts getOptions, printer *cliprinter.ResourcePrinter) error {
	server := config.GetDefaultContext()

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	var (
		interfaces []*gqlpublicapi.Interface
		errors     []error
	)

	ifaces, err := cli.ListInterfacesWithLatestRevision(ctx, gqlpublicapi.InterfaceFilter{
		PathPattern: &allPathPrefix,
	})
	if err != nil {
		return err
	}

	if len(opts.interfacePaths) == 0 {
		interfaces = ifaces
	} else {
		ifaceMap := interfaceSliceToMap(ifaces)

		for _, path := range opts.interfacePaths {
			iface, found := ifaceMap[path]
			if !found {
				errors = append(errors, errNotFound(path))
				continue
			}

			interfaces = append(interfaces, iface)
		}
	}

	cliprinter.PrintErrors(errors)
	return printer.Print(interfaces)
}

func interfaceSliceToMap(ifaces []*gqlpublicapi.Interface) map[string]*gqlpublicapi.Interface {
	res := make(map[string]*gqlpublicapi.Interface)

	for i := range ifaces {
		iface := ifaces[i]
		res[iface.Path] = iface
	}

	return res
}

func errNotFound(name string) error {
	return fmt.Errorf(`NotFound: Interface "%s" not found`, name)
}

func tableDataOnGet(in interface{}) (cliprinter.TableData, error) {
	out := cliprinter.TableData{}

	interfaces, ok := in.([]*gqlpublicapi.Interface)
	if !ok {
		return cliprinter.TableData{}, fmt.Errorf("got unexpected input type, expected []*gqlpublicapi.Interface, got %T", in)
	}

	out.Headers = []string{"PATH", "LATEST REVISION", "IMPLEMENTATIONS"}
	for _, i := range interfaces {
		out.MultipleRows = append(out.MultipleRows, []string{
			i.Path,
			i.LatestRevision.Revision,
			implList(i.LatestRevision.ImplementationRevisions)},
		)
	}

	return out, nil
}

func implList(revisions []*gqlpublicapi.ImplementationRevision) string {
	var out []string
	for _, r := range revisions {
		if r == nil || r.Metadata == nil {
			continue
		}
		out = append(out, r.Metadata.Path)
	}
	return strings.Join(out, "\n")
}
