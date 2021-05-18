package policy

import (
	"context"
	"fmt"
	"io"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/pkg/engine/api/graphql"
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	"sigs.k8s.io/yaml"
)

func Edit(ctx context.Context, w io.Writer) error {
	server := config.GetDefaultContext()

	engineCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	existingPolicy, err := engineCli.GetPolicy(ctx)
	if err != nil {
		return err
	}

	policyInput, err := askForPolicyInput(existingPolicy)
	if err != nil {
		return err
	}

	_, err = engineCli.UpdatePolicy(ctx, policyInput)
	if err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintlnFunc()
	okCheck(w, "Policy updated successfully")

	return nil
}

// TODO: Do not require additional confirmation to enter the editor
// Once this feature is implemented: https://github.com/AlecAivazis/survey/issues/313
// Currently we would need to use `survey` internals, which doesn't seem to be the right way to do it.
func askForPolicyInput(existingPolicy *graphql.Policy) (*graphql.PolicyInput, error) {
	policyStr, err := toYAMLString(existingPolicy)
	if err != nil {
		return nil, err
	}

	editor := ""
	prompt := &survey.Editor{
		Message:       "Edit current Policy using YAML syntax",
		Default:       heredoc.Doc(policyStr),
		AppendDefault: true,
		HideDefault:   true,
	}

	err = survey.AskOne(prompt, &editor, survey.WithValidator(validatePolicy))
	if err != nil {
		return nil, err
	}

	return toPolicyInput([]byte(editor))
}

func toPolicyInput(rawInput []byte) (*graphql.PolicyInput, error) {
	var policyInput *graphql.PolicyInput
	err := yaml.Unmarshal(rawInput, &policyInput)
	if err != nil {
		return nil, err
	}

	return policyInput, nil
}

func toYAMLString(policy *graphql.Policy) (string, error) {
	bytes, err := yaml.Marshal(policy)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func validatePolicy(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("Cannot enforce YAML syntax validation on response of type %T", val)
	}

	_, err := toPolicyInput([]byte(str))
	return err
}
