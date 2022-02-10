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
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/printer"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type createOptions struct {
	FilePath           string
	TypeInstancesFiles []string
}

// NewCreate returns a cobra.Command for creating a TypeInstance on a Local Hub.
func NewCreate() *cobra.Command {
	var opts createOptions

	resourcePrinter := printer.NewForResource(
		os.Stdout,
		printer.WithJSON(),
		printer.WithYAML(),
		printer.WithTable(tableDataOnCreate),
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a new TypeInstance(s)",
		Long: heredoc.Doc(`
			Create one or multiple TypeInstances from a given file.

			Syntax:
				
				typeInstances:
				  - alias: parent # required when submitting more than one TypeInstance
				    attributes: # optional
				      - path: cap.attribute.cloud.provider.aws
				        revision: 0.1.0
				    typeRef: # required
				      path: cap.type.aws.auth.credentials
				      revision: 0.1.0
				    value: # required
				      accessKeyID: fake-123
				      secretAccessKey: fake-456
				
				usesRelations: # optional
				  - from: parent
				    to: 123-4313 # ID of already existing TypeInstance, or TypeInstance alias from a given request


			NOTE: Supported syntax are YAML and JSON.
		`),
		Example: heredoc.WithCLIName(`
			# Create TypeInstances defined in a given file
			<cli> typeinstance create -f ./tmp/typeinstances.yaml
		`, cli.Name),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createTI(cmd.Context(), opts, resourcePrinter)
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&opts.TypeInstancesFiles, cli.FromFileFlagName, "f", []string{}, "The TypeInstances input in YAML format (can specify multiple)")
	panicOnError(cmd.MarkFlagRequired(cli.FromFileFlagName)) // this cannot happen

	resourcePrinter.RegisterFlags(flags)
	client.RegisterFlags(flags)

	return cmd
}

func createTI(ctx context.Context, opts createOptions, resourcePrinter *printer.ResourcePrinter) error {
	typeInstanceToCreate := &gqllocalapi.CreateTypeInstancesInput{}

	server := config.GetDefaultContext()
	hubCli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	for _, path := range opts.TypeInstancesFiles {
		out, err := loadCreateTypeInstanceFromFile(path)
		if err != nil {
			return err
		}

		typeInstanceToCreate = mergeCreateTypeInstances(typeInstanceToCreate, out)
	}

	validationResult, err := validation.ValidateTypeInstancesToCreate(ctx, hubCli, typeInstanceToCreate)
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}
	if validationResult.Len() > 0 {
		return validationResult.ErrorOrNil()
	}

	// HACK: UsesRelations are required on GQL side so at least empty array needs to be send
	if typeInstanceToCreate.UsesRelations == nil {
		typeInstanceToCreate.UsesRelations = []*gqllocalapi.TypeInstanceUsesRelationInput{}
	}

	createdTI, err := hubCli.CreateTypeInstances(ctx, typeInstanceToCreate)
	if err != nil {
		return err
	}

	return resourcePrinter.Print(createdTI)
}

func loadCreateTypeInstanceFromFile(path string) (*gqllocalapi.CreateTypeInstancesInput, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "cannot open file with TypeInstance input")
	}

	d := yaml.NewYAMLOrJSONDecoder(f, decodeBufferSize)
	out := &gqllocalapi.CreateTypeInstancesInput{}
	for {
		item := &gqllocalapi.CreateTypeInstancesInput{}
		if err := d.Decode(item); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error parsing %s: %v", path, err)
		}

		if err := validateInput(item); err != nil {
			return nil, err
		}

		out = mergeCreateTypeInstances(out, item)
	}

	return out, nil
}

func mergeCreateTypeInstances(a, b *gqllocalapi.CreateTypeInstancesInput) *gqllocalapi.CreateTypeInstancesInput {
	a.UsesRelations = append(a.UsesRelations, b.UsesRelations...)
	a.TypeInstances = append(a.TypeInstances, b.TypeInstances...)

	return a
}

func tableDataOnCreate(in interface{}) (printer.TableData, error) {
	out := printer.TableData{}

	typeInstances, ok := in.([]gqllocalapi.CreateTypeInstanceOutput)
	if !ok {
		return printer.TableData{}, fmt.Errorf("got unexpected input type, expected []gqllocalapi.CreateTypeInstanceOutput, got %T", in)
	}

	out.Headers = []string{"ALIAS", "ASSIGNED ID"}
	for _, ti := range typeInstances {
		out.MultipleRows = append(out.MultipleRows, []string{ti.Alias, ti.ID})
	}

	return out, nil
}

func validateInput(in *gqllocalapi.CreateTypeInstancesInput) error {
	// Single TypeInstance can be without alias. Submitting multiple TypeInstances without alias (even if relations are not defined)
	// are hard to represent relations between input and returned IDs.
	if len(in.TypeInstances) > 1 {
		for _, ti := range in.TypeInstances {
			if ti.Alias == nil || *ti.Alias == "" {
				return fmt.Errorf("when submitting more than one TypeInstance, all must have alias property set to easily relate it with returned ID")
			}
		}
	}

	return nil
}
