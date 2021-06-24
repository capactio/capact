package typeinstance

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	heredocx "capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/printer"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/MakeNowJust/heredoc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

type applyOptions struct {
	TypeInstancesFiles []string
}

// NewApply returns a cobra.Command for the "typeinstance apply" command.
func NewApply() *cobra.Command {
	var opts applyOptions

	resourcePrinter := printer.NewForResource(
		os.Stdout,
		printer.WithJSON(),
		printer.WithYAML(),
		printer.WithTable(tableDataOnGet),
	)

	cmd := &cobra.Command{
		Use:   "apply -f file...",
		Short: "Apply a given TypeInstance(s)",
		Long: heredoc.Doc(`
			Updates a given TypeInstance(s).
			CAUTION: Race updates may occur as TypeInstance locking is not used by CLI.
		`),
		Example: heredocx.WithCLIName(`
			# Apply TypeInstances from the given file.
			<cli> typeinstance apply -f /tmp/typeinstances.yaml
		`, cli.Name),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return applyTI(cmd.Context(), opts, resourcePrinter)
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opts.TypeInstancesFiles, cli.FromFileFlagName, "f", []string{}, "The TypeInstances input in YAML format (can specify multiple)")
	panicOnError(cmd.MarkFlagRequired(cli.FromFileFlagName)) // this cannot happen

	resourcePrinter.RegisterFlags(flags)

	return cmd
}

func applyTI(ctx context.Context, opts applyOptions, resourcePrinter *printer.ResourcePrinter) error {
	server := config.GetDefaultContext()

	hubCli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	typeInstanceToUpdate, err := typeInstancesFromFile(opts.TypeInstancesFiles)
	if err != nil {
		return err
	}

	updatedTI, err := hubCli.UpdateTypeInstances(ctx, typeInstanceToUpdate)
	if err != nil {
		return err
	}

	return resourcePrinter.Print(updatedTI)
}

func typeInstancesFromFile(typeInstancesFiles []string) ([]gqllocalapi.UpdateTypeInstancesInput, error) {
	var typeInstanceToUpdate []gqllocalapi.UpdateTypeInstancesInput

	for _, path := range typeInstancesFiles {
		out, err := loadUpdateTypeInstanceFromFile(path)
		if err != nil {
			return nil, err
		}
		typeInstanceToUpdate = append(typeInstanceToUpdate, out...)
	}

	return typeInstanceToUpdate, nil
}

func loadUpdateTypeInstanceFromFile(path string) ([]gqllocalapi.UpdateTypeInstancesInput, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "cannot open file with TypeInstance input")
	}

	d := yamlutil.NewYAMLOrJSONDecoder(f, decodeBufferSize)
	var out []gqllocalapi.UpdateTypeInstancesInput
	for {
		item := gqllocalapi.UpdateTypeInstancesInput{}
		if err := d.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error parsing %s: %v", path, err)
		}
		out = append(out, item)
	}

	return out, nil
}
