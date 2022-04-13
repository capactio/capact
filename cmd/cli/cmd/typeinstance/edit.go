package typeinstance

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	storagebackend "capact.io/capact/pkg/hub/storage-backend"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	"github.com/pkg/errors"
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

	flags := cmd.Flags()
	client.RegisterFlags(flags)

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

	r := validation.ResultAggregator{}
	err = r.Report(validation.ValidateTypeInstanceToUpdate(ctx, hubCli, typeInstanceToUpdate))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstance")
	}
	if r.ErrorOrNil() != nil {
		return r.ErrorOrNil()
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
		return nil, errors.Wrap(err, "while finding TypeInstance")
	}
	if out == nil {
		return nil, fmt.Errorf("TypeInstance %s not found", tiID)
	}

	backendData, err := storagebackend.NewTypeInstanceValue(ctx, cli, out)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching storage backend data")
	}

	updateTI := mapTypeInstanceToUpdateType(out)

	var valueData string
	if backendData != nil && !backendData.AcceptValue {
		valueData, err = getCommentedOutTypeInstanceValue(&updateTI)
		if err != nil {
			return nil, errors.Wrap(err, "while getting commented out TypeInstance value")
		}
	}

	setTypeInstanceValueForMarshaling(backendData, &updateTI)
	rawInput, err := yaml.Marshal(updateTI)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling updated TypeInstance")
	}

	prompt := &survey.Editor{
		Message:       "Edit TypeInstance in YAML format",
		Default:       fmt.Sprintf("%s%s", string(rawInput), valueData),
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

func getCommentedOutTypeInstanceValue(in *gqllocalapi.UpdateTypeInstancesInput) (string, error) {
	type valueOnly struct {
		Value interface{} `json:"value,omitempty"`
	}
	encodedValue, err := yaml.Marshal(valueOnly{in.TypeInstance.Value})
	if err != nil {
		return "", errors.Wrap(err, "while marshaling storage backend value")
	}
	lines := strings.Split(string(encodedValue), "\n")
	for i := range lines {
		if lines[i] == "" {
			continue
		}
		lines[i] = "# " + lines[i]
	}
	return strings.Join(lines, "\n"), nil
}

func isValidUpdateTypeInstancesInput(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("cannot enforce UpdateTypeInstancesInput syntax validation on response of type %T", val)
	}

	out := gqllocalapi.UpdateTypeInstancesInput{}
	return yaml.UnmarshalStrict([]byte(str), &out)
}
