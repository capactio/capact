package policy

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/pkg/engine/api/graphql"
	"github.com/fatih/color"
)

// ApplyOptions holds configuration for updating Capact Policy.
type ApplyOptions struct {
	PolicyFilePath string
}

// Validate validates if provided options are valid.
func (opts *ApplyOptions) Validate() error {
	if opts.PolicyFilePath == "" {
		return errors.New("Policy YAML file path cannot be empty")
	}

	return nil
}

// Apply updates Capact policy with a given input.
func Apply(ctx context.Context, opts ApplyOptions, w io.Writer) error {
	err := opts.Validate()
	if err != nil {
		return err
	}

	server := config.GetDefaultContext()

	engineCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	policyInput, err := loadPolicyInputFromFile(opts.PolicyFilePath)
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

func loadPolicyInputFromFile(path string) (*graphql.PolicyInput, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return toPolicyInput(bytes)
}
