package typeinstance

import (
	"context"
	"fmt"
	"io"
	"os"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type editOptions struct {
	EditTypeInstanceID string
}

// NewEdit returns a cobra.Command for editing a TypeInstance on a Local Hub.
func NewEdit() *cobra.Command {
	var opts editOptions

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
			return editTI(cmd.Context(), opts, os.Stdout)
		},
	}

	return cmd
}

func editTI(ctx context.Context, opts editOptions, w io.Writer) error {
	server := config.GetDefaultContext()

	hubCli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	typeInstanceToUpdate, err := typeInstanceViaEditor(ctx, hubCli, opts.EditTypeInstanceID)
	if err != nil {
		return err
	}

	_, err = hubCli.UpdateTypeInstances(ctx, typeInstanceToUpdate)
	if err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintfFunc()
	okCheck(w, "TypeInstance %s updated successfully\n", opts.EditTypeInstanceID)

	return nil
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
	if err := survey.AskOne(prompt, &rawEdited, survey.WithValidator(isValidUpdateTypeInstancesInput)); err != nil {
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

func isValidUpdateTypeInstancesInput(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("cannot enforce UpdateTypeInstancesInput syntax validation on response of type %T", val)
	}

	out := gqllocalapi.UpdateTypeInstancesInput{}
	return yaml.UnmarshalStrict([]byte(str), &out)
}
