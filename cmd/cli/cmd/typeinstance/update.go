package typeinstance

import (
	"context"
	"fmt"
	"io"
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	heredocx "capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/stringsx"
	gqllocalapi "capact.io/capact/pkg/och/api/graphql/local"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type updateOptions struct {
	TypeInstancesFiles []string
	UpdateTIID         string
}

func NewUpdate() *cobra.Command {
	var opts updateOptions

	resourcePrinter := printer.NewForResource(
		os.Stdout,
		printer.WithJSON(),
		printer.WithYAML(),
		printer.WithTable(tableDataOnGet),
	)

	cmd := &cobra.Command{
		Use:   "update [-f file]... | TYPE_INSTANCE_ID",
		Short: "Updates a given TypeInstance(s)",
		Long: heredoc.Doc(`
			Updates a given TypeInstance(s).
			CAUTION: Race updates may occur as TypeInstance locking is not used by CLI.
		`),
		Example: heredocx.WithCLIName(`
			# Apply TypeInstances from the given file
			<cli> typeinstance update -f /tmp/typeinstances.yaml 

			# Update TypeInstance in editor mode 
			<cli> typeinstance update TYPE_INSTANCE_ID
		`, cli.Name),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case stringsx.AreAllSlicesEmpty(opts.TypeInstancesFiles, args):
				return fmt.Errorf("must specify one of %s or TypeInstance ID to update in interactive mode", fromFileFlagName)
			case stringsx.AreAllSlicesNotEmpty(opts.TypeInstancesFiles, args):
				return fmt.Errorf("cannot specify both %s and TypeInstance ID", fromFileFlagName)
			}

			if len(args) == 1 {
				opts.UpdateTIID = args[0]
			}

			return updateTI(cmd.Context(), opts, resourcePrinter)
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opts.TypeInstancesFiles, fromFileFlagName, "f", []string{}, "The TypeInstances input in YAML format (can specify multiple)")
	resourcePrinter.RegisterFlags(flags)

	return cmd
}

func updateTI(ctx context.Context, opts updateOptions, resourcePrinter *printer.ResourcePrinter) error {
	server := config.GetDefaultContext()

	hubCli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	load := func() ([]gqllocalapi.UpdateTypeInstancesInput, error) {
		if opts.UpdateTIID != "" {
			return typeInstanceViaEditor(ctx, hubCli, opts.UpdateTIID)
		}
		return typeInstancesFromFile(opts.TypeInstancesFiles)
	}

	typeInstanceToUpdate, err := load()
	if err != nil {
		return err
	}

	updatedTI, err := hubCli.UpdateTypeInstances(ctx, typeInstanceToUpdate)
	if err != nil {
		return err
	}

	return resourcePrinter.Print(updatedTI)
}

func typeInstanceViaEditor(ctx context.Context, cli client.Hub, tiID string) ([]gqllocalapi.UpdateTypeInstancesInput, error) {
	out, err := cli.FindTypeInstance(ctx, tiID)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("TypeInstance %s not found", tiID)
	}

	rawInput, err := yaml.Marshal(mapTypeInstanceToUpdateType(out))
	if err != nil {
		return nil, err
	}

	prompt := &survey.Editor{
		Message:       "Edit TypeInstance in YAML format",
		Default:       string(rawInput),
		AppendDefault: true,
		HideDefault:   true,
	}

	rawEdited := ""
	if err := survey.AskOne(prompt, &rawEdited); err != nil {
		return nil, err
	}

	edited := gqllocalapi.UpdateTypeInstancesInput{}
	if err := yaml.Unmarshal([]byte(rawEdited), &edited); err != nil {
		return nil, err
	}

	return []gqllocalapi.UpdateTypeInstancesInput{
		edited,
	}, nil
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
	f, err := os.Open(path)
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

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
