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
	cliprinter "capact.io/capact/internal/cli/printer"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/spf13/cobra"
)

const yamlFileSeparator = "---"

// GetOptions is used to store the configuration flags for the Get command.
type GetOptions struct {
	RequestedTypeInstancesIDs []string
	ExportToUpdateFormat      bool
}

// ErrTableFormatWithExportFlag is used to inform that --export flag was used with table output,
// which is not supported.
var ErrTableFormatWithExportFlag = fmt.Errorf("cannot use --export with table output")

// NewGet returns a cobra.Command for the "typeinstance get" command.
func NewGet() *cobra.Command {
	var opts GetOptions
	out := os.Stdout

	resourcePrinter := cliprinter.NewForResource(
		out,
		cliprinter.WithJSON(),
		cliprinter.WithYAML(),
		cliprinter.WithTable(tableDataOnGet),
	)

	cmd := &cobra.Command{
		Use:   "get [TYPE_INSTANCE_ID...]",
		Short: "Displays one or multiple TypeInstances",
		Example: heredoc.WithCLIName(`
			# Display TypeInstances with IDs 'c49b' and '4793'
			<cli> typeinstance get c49b 4793
			
			# Save TypeInstances with IDs 'c49b' and '4793' to file in the update format which later can be submitted for update by: 
			# <cli> typeinstance apply --from-file /tmp/typeinstances.yaml
			<cli> typeinstance get c49b 4793 -oyaml --export > /tmp/typeinstances.yaml
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RequestedTypeInstancesIDs = args

			if opts.ExportToUpdateFormat && resourcePrinter.PrintFormat() == cliprinter.TableFormat {
				return ErrTableFormatWithExportFlag
			}

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
		return hubCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
	}

	var (
		out  []gqllocalapi.TypeInstance
		errs []error
	)

	// TODO: make it client-side
	for _, id := range opts.RequestedTypeInstancesIDs {
		ti, err := hubCli.FindTypeInstance(ctx, id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if ti == nil {
			errs = append(errs, fmt.Errorf("TypeInstance %s not found", id))
			continue
		}

		out = append(out, *ti)
	}

	cliprinter.PrintErrors(errs)
	return out, nil
}

func tableDataOnGet(inRaw interface{}) (cliprinter.TableData, error) {
	out := cliprinter.TableData{}

	switch in := inRaw.(type) {
	case []gqllocalapi.TypeInstance:
		out.Headers = []string{"TYPE INSTANCE ID", "TYPE", "USES", "USED BY", "REVISION", "LOCKED"}
		for _, ti := range in {
			out.MultipleRows = append(out.MultipleRows, []string{
				ti.ID,
				fmt.Sprintf("%s:%s", ti.TypeRef.Path, ti.TypeRef.Revision),
				toTypeInstanceIDs(ti.Uses),
				toTypeInstanceIDs(ti.UsedBy),
				strconv.FormatInt(int64(ti.LatestResourceVersion.ResourceVersion), 10),
				strconv.FormatBool(ti.LockedBy != nil),
			})
		}
	case gqllocalapi.UpdateTypeInstancesInput: // this shouldn't happen because of previous options validation
		return cliprinter.TableData{}, ErrTableFormatWithExportFlag
	default:
		return cliprinter.TableData{}, fmt.Errorf("got unexpected input type, expected []gqllocalapi.TypeInstance, got %T", inRaw)
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
