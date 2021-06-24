package implementations

import (
	"context"
	"fmt"
	"os"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/printer"
	cliprinter "capact.io/capact/internal/cli/printer"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"

	"github.com/spf13/cobra"
)

type getOptions struct {
	implementationPaths []string
}

// NewGet returns a cobra.Command for getting Implementations from a public Hub.
func NewGet() *cobra.Command {
	var opts getOptions

	resourcePrinter := printer.NewForResource(os.Stdout, printer.WithJSON(), printer.WithYAML(), printer.WithTable(tableDataOnGet))

	get := &cobra.Command{
		Use:   "get",
		Short: "Displays one or multiple Implementations available on the Hub server",
		Example: heredoc.WithCLIName(`
			# Show all Implementation Revisions in table format
			<cli> hub implementations get
			
			# Show "cap.implementation.gcp.cloudsql.postgresql.install" Implementation Revisions in YAML format			
			<cli> hub implementations get cap.interface.database.postgresql.install -oyaml
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.implementationPaths = args
			return getImpl(cmd.Context(), opts, resourcePrinter)
		},
	}

	flags := get.Flags()
	resourcePrinter.RegisterFlags(flags)

	return get
}

func getImpl(ctx context.Context, opts getOptions, printer *cliprinter.ResourcePrinter) error {
	server := config.GetDefaultContext()

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	var (
		implementationRevisions []*gqlpublicapi.ImplementationRevision
		errors                  []error
	)

	impls, err := cli.ListImplementationRevisions(ctx, nil)
	if err != nil {
		return err
	}

	if len(opts.implementationPaths) == 0 {
		implementationRevisions = impls
	} else {
		implMap := implementationSliceToMap(impls)

		for _, path := range opts.implementationPaths {
			foundImpls, found := implMap[path]
			if !found {
				errors = append(errors, errNotFound(path))
				continue
			}

			implementationRevisions = append(implementationRevisions, foundImpls...)
		}
	}

	cliprinter.PrintErrors(errors)
	return printer.Print(implementationRevisions)
}

func implementationSliceToMap(impls []*gqlpublicapi.ImplementationRevision) map[string][]*gqlpublicapi.ImplementationRevision {
	res := make(map[string][]*gqlpublicapi.ImplementationRevision)

	for i := range impls {
		impl := impls[i]
		res[impl.Metadata.Path] = append(res[impl.Metadata.Path], impl)
	}

	return res
}

func errNotFound(name string) error {
	return fmt.Errorf(`NotFound: Implementation "%s" not found`, name)
}

func tableDataOnGet(in interface{}) (printer.TableData, error) {
	out := printer.TableData{}

	implementations, ok := in.([]*gqlpublicapi.ImplementationRevision)
	if !ok {
		return printer.TableData{}, fmt.Errorf("got unexpected input type, expected []gqlpublicapi.ImplementationRevision, got %T", in)
	}

	out.Headers = []string{"PATH", "REVISION", "ATTRIBUTES"}
	for _, impl := range implementations {
		out.MultipleRows = append(out.MultipleRows, []string{
			impl.Metadata.Path,
			impl.Revision,
			attrNames(impl.Metadata.Attributes),
		})
	}

	return out, nil
}

func attrNames(attrs []*gqlpublicapi.AttributeRevision) string {
	var out []string
	for _, a := range attrs {
		if a == nil || a.Metadata == nil {
			continue
		}
		out = append(out, a.Metadata.Path)
	}

	return strings.Join(out, "\n")
}
