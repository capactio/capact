package implementation

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
)

type helmChart struct {
	// URL is address to helm repository
	URL string
	// Version defines a helm chart version
	Version string
}

func askForImplementationTool() (string, error) {
	var selectedTool string
	availableTool := []string{helmTool, terraformTool, emptyManifest}
	prompt := &survey.Select{
		Message: "Based on which tool do you want to generate implementation:",
		Options: availableTool,
	}
	err := survey.AskOne(prompt, &selectedTool)
	return selectedTool, err
}

func askForInterface() (string, error) {
	path, err := common.AskForManifestPathSuffix("Interface manifest path suffix")
	if err != nil {
		return "", errors.Wrap(err, "while asking for interface manifest path suffix")
	}

	revision, err := common.AskForManifestRevision()
	if err != nil {
		return "", errors.Wrap(err, "while asking for interface revision")
	}
	return common.AddRevisionToPath(path, revision), nil
}

func askForLicense() (types.License, error) {
	var licenseName, licenseRef string
	name := &survey.Input{
		Message: "Name of the license",
		Default: common.ApacheLicense,
	}
	err := survey.AskOne(name, &licenseName)
	if err != nil {
		return types.License{}, err
	}

	// TODO: can be extended to a list of licenses that do not need a ref
	if licenseName == common.ApacheLicense {
		return types.License{
			Name: &licenseName,
		}, nil
	}

	ref := &survey.Input{
		Message: "Reference for the license",
		Default: "",
	}
	err = survey.AskOne(ref, &licenseRef)
	if err != nil {
		return types.License{}, err
	}
	return types.License{
		Name: &licenseName,
		Ref:  &licenseRef,
	}, nil
}

func askForProvider() (manifestgen.Provider, error) {
	var selectedProvider string
	availableProviders := []string{string(manifestgen.ProviderAWS), string(manifestgen.ProviderGCP)}
	prompt := &survey.Select{
		Message: "Create a provider-specific workflow:",
		Options: availableProviders,
	}
	err := survey.AskOne(prompt, &selectedProvider)
	return manifestgen.Provider(selectedProvider), err
}

func askForSource() (string, error) {
	var source string
	prompt := []*survey.Question{
		{
			Prompt: &survey.Input{
				Message: "Path to the Terraform module, such as URL to Tarball or Git repository",
				Default: "",
			},
		},
	}
	err := survey.Ask(prompt, &source)
	return source, err
}

func askForHelmChartDetails() (helmChart, error) {
	var helmChartInfo helmChart
	var qs = []*survey.Question{
		{
			Name: "URL",
			Prompt: &survey.Input{
				Message: "URL of the Helm repository",
				Default: "",
			},
		},
		{
			Name: "Version",
			Prompt: &survey.Input{
				Message: "Version of the Helm chart",
				Default: "",
			},
		},
	}
	err := survey.Ask(qs, &helmChartInfo)
	return helmChartInfo, err
}
