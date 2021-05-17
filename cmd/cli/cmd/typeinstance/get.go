package typeinstance

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/printer"
	gqllocalapi "capact.io/capact/pkg/och/api/graphql/local"

	"github.com/spf13/cobra"
)

const yamlFileSeparator = "---"

type GetOptions struct {
	RequestedTypeInstancesIDs []string
	ExportToUpdateFormat      bool
}

func NewGet() *cobra.Command {
	var opts GetOptions
	out := os.Stdout

	resourcePrinter := printer.NewForResource(
		out,
		printer.WithJSON(),
		printer.WithYAML(),
		printer.WithTable(tableDataOnGet),
	)

	cmd := &cobra.Command{
		Use:   "get [TYPE_INSTANCE_ID...]",
		Short: "Displays one or multiple TypeInstances",
		Example: heredoc.WithCLIName(`
			# Display TypeInstances with IDs c49b and 4793
			<cli> typeinstance get c49b 4793
			
			# Save TypeInstances with IDs c49b and 4793 to file in the update format which later can be submitted for update by: 
			# <cli> typeinstance update --from-file /tmp/typeinstances.yaml
			<cli> typeinstance get c49b 4793 -oyaml --export > /tmp/typeinstances.yaml
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RequestedTypeInstancesIDs = args

			tis, err := getTI(cmd.Context(), opts)
			if err != nil {
				return err
			}

			if opts.ExportToUpdateFormat {
				for idx := range tis {
					conv := mapTypeInstanceToUpdateType(&tis[idx])
					fmt.Fprintln(out, yamlFileSeparator)
					if err := resourcePrinter.Print(conv); err != nil {
						return err
					}
				}
				return nil
			}

			return resourcePrinter.Print(tis)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.ExportToUpdateFormat, "export", false, "Converts TypeInstance to update format.")
	resourcePrinter.RegisterFlags(flags)

	return cmd
}

func getTI(ctx context.Context, opts GetOptions) ([]gqllocalapi.TypeInstance, error) {
	server := config.GetDefaultContext()

	hubCli, err := client.NewHub(server)
	if err != nil {
		return nil, err
	}

	if len(opts.RequestedTypeInstancesIDs) == 0 {
		return hubCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
			Attributes: nil,
			TypeRef:    nil,
		})
	}

	// TODO: make it client-side
	var out []gqllocalapi.TypeInstance
	for _, id := range opts.RequestedTypeInstancesIDs {
		ti, err := hubCli.FindTypeInstance(ctx, id)
		if err != nil {
			return nil, err
		}

		out = append(out, *ti)
	}

	return out, nil
}

func tableDataOnGet(inRaw interface{}) (printer.TableData, error) {
	out := printer.TableData{}

	switch in := inRaw.(type) {
	case []gqllocalapi.TypeInstance:
		out.Headers = []string{"TYPE INSTANCE ID", "TYPE", "USES", "USED BY", "REVISION", "LOCKED"}
		for _, ti := range in {
			out.MultipleRows = append(out.MultipleRows, []string{
				ti.ID,
				ti.TypeRef.Path,
				toTypeInstanceIDs(ti.Uses),
				toTypeInstanceIDs(ti.UsedBy),
				strconv.FormatInt(int64(ti.LatestResourceVersion.ResourceVersion), 10),
				strconv.FormatBool(ti.LockedBy != nil),
			})
		}
	case gqllocalapi.UpdateTypeInstancesInput: // this is a rare case when someone specify only `--export` or `--export -o=table`
		return printer.TableData{}, fmt.Errorf("cannot use --export with table output")
	default:
		return printer.TableData{}, fmt.Errorf("got unexpected input type, expected []gqllocalapi.TypeInstance, got %T", inRaw)
	}

	return out, nil
}

func toTypeInstanceIDs(in []*gqllocalapi.TypeInstance) string {
	var out []string
	for _, ti := range in {
		out = append(out, ti.ID)
	}

	if len(out) == 0 {
		return " —— "
	}
	return strings.Join(out, ", ")
}
