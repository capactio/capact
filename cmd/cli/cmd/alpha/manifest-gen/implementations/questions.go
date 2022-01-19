package implementations

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/pkg/runner/helm"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
)

type helmChartLocation string

const (
	localHelmChartLocation  helmChartLocation = "local"
	remoteHelmChartLocation helmChartLocation = "remote"
)

func askForImplementationTool() (string, error) {
	var selectedTool string
	var options []string

	availableTool := []implGeneratorType{helmTool, terraformTool, emptyManifest}
	for _, tool := range availableTool {
		options = append(options, string(tool))
	}

	prompt := &survey.Select{
		Message: "Based on which tool do you want to generate implementation:",
		Options: options,
	}
	err := survey.AskOne(prompt, &selectedTool)
	return selectedTool, err
}

func askForInterface() (string, error) {
	path, err := common.AskForManifestPathSuffix("Interface manifest path suffix")
	if err != nil {
		return "", errors.Wrap(err, "while asking for interface manifest path suffix")
	}

	revision, err := common.AskForManifestRevision("Interface manifest revision")
	if err != nil {
		return "", errors.Wrap(err, "while asking for interface revision")
	}
	return common.AddRevisionToPath(path, revision), nil
}

func askForLicense() (string, error) {
	var licenseName string
	name := &survey.Input{
		Message: "License name",
		Default: *common.ApacheLicense,
	}
	err := survey.AskOne(name, &licenseName)
	return licenseName, err
}

func askForProvider() (manifestgen.Provider, error) {
	var selectedProvider string
	availableProviders := []string{string(manifestgen.ProviderAWS), string(manifestgen.ProviderGCP)}
	prompt := &survey.Select{
		Message: "Terraform provider",
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
				Message: "Location of the hosted Terraform module, such as URL to Tarball or Git repository",
				Default: "",
			},
		},
	}
	err := survey.Ask(prompt, &source)
	return source, err
}

func askForHelmLocation() (string, error) {
	var selectedLocation string
	availableLocations := []string{string(localHelmChartLocation), string(remoteHelmChartLocation)}
	prompt := &survey.Select{
		Message: "Select Helm chart location",
		Options: availableLocations,
	}
	err := survey.AskOne(prompt, &selectedLocation)
	return selectedLocation, err
}

// TODO: in case of error ask for helm chart details in a loop
func askForHelmChartDetails() (helm.Chart, error) {
	var helmChartInfo helm.Chart

	location, err := askForHelmLocation()
	if err != nil {
		return helm.Chart{}, errors.Wrap(err, "while asking for selecting Helm location")
	}

	if location == string(localHelmChartLocation) {
		helmTemplate, err := common.AskForDirectory("Path to Helm chart", "")
		if err != nil {
			return helm.Chart{}, errors.Wrap(err, "while asking for path to Helm chart")
		}
		helmChartInfo.Name = helmTemplate
		return helmChartInfo, nil
	}

	var qs = []*survey.Question{
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "Helm chart name",
				Default: "",
			},
			Validate: survey.Required,
		},
		{
			Name: "Version",
			Prompt: &survey.Input{
				Message: "Helm chart version",
				Default: "",
			},
			Validate: survey.Required,
		},
		{
			Name: "Repo",
			Prompt: &survey.Input{
				Message: "Helm repository URL",
				Default: "",
			},
			Validate: common.ManyValidators([]survey.Validator{survey.Required, common.ValidateURL}),
		},
	}
	err = survey.Ask(qs, &helmChartInfo)
	return helmChartInfo, err
}
