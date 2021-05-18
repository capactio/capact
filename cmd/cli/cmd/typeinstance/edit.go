package typeinstance

import (
	"context"
	"fmt"
	"os"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/printer"
	gqllocalapi "capact.io/capact/pkg/och/api/graphql/local"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type editOptions struct {
	EditTypeInstanceID string
}

func NewEdit() *cobra.Command {
	var opts editOptions

	resourcePrinter := printer.NewForResource(
		os.Stdout,
		printer.WithJSON(),
		printer.WithYAML(),
		printer.WithTable(tableDataOnGet),
	)

	cmd := &cobra.Command{
		Use:   "edit TYPE_INSTANCE_ID",
		Short: "Edit a given TypeInstance via editor",
		Long: heredoc.Doc(`
			Update a given TypeInstance.
			CAUTION: Race update may occur as TypeInstance locking is not used by CLI.
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.EditTypeInstanceID = args[0]
			return editTI(cmd.Context(), opts, resourcePrinter)
		},
	}

	flags := cmd.Flags()
	resourcePrinter.RegisterFlags(flags)

	return cmd
}

func editTI(ctx context.Context, opts editOptions, resourcePrinter *printer.ResourcePrinter) error {
	server := config.GetDefaultContext()

	hubCli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	typeInstanceToUpdate, err := typeInstanceViaEditor(ctx, hubCli, opts.EditTypeInstanceID)
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
