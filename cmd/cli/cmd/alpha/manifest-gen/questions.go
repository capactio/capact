package manifestgen

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
)

func askForManifestType() ([]string, error) {
	var manifestTypes []string
	availableManifestsType := []string{common.AttributeManifest, common.TypeManifest, common.InterfaceGroupManifest, common.InterfaceManifest, common.ImplementationManifest}
	prompt := []*survey.Question{
		{
			Prompt: &survey.MultiSelect{
				Message: "Which manifests do you want to generate:",
				Options: availableManifestsType,
			},
			Validate: survey.MinItems(1),
		},
	}
	err := survey.Ask(prompt, &manifestTypes)
	return manifestTypes, err
}

func askForCommonMetadataInformation() (*common.Metadata, error) {
	var metadata common.Metadata
	var qs = []*survey.Question{
		{
			Name: "DocumentationURL",
			Prompt: &survey.Input{
				Message: "Documentation URL",
				Default: "",
			},
			Validate: common.ValidateURL,
		},
		{
			Name: "SupportURL",
			Prompt: &survey.Input{
				Message: "Support URL",
				Default: "",
			},
			Validate: common.ValidateURL,
		},
		{
			Name: "IconURL",
			Prompt: &survey.Input{
				Message: "Icon URL?",
				Default: "",
			},
			Validate: common.ValidateURL,
		},
	}
	err := survey.Ask(qs, &metadata)
	if err != nil {
		return nil, errors.Wrap(err, "while asking for metadata")
	}

	maintainers, err := askForMaintainers()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for maintainers")
	}
	metadata.Maintainers = maintainers
	return &metadata, nil
}

func askForMaintainers() ([]common.Maintainers, error) {
	var maintainers []common.Maintainers
	for {
		addMore := false
		if len(maintainers) >= 1 {
			prompt := &survey.Confirm{
				Message: "Do you want to add more maintainers",
			}
			err := survey.AskOne(prompt, &addMore)
			if err != nil {
				return nil, errors.Wrap(err, "while asking if add maintainers")
			}
			if !addMore {
				return maintainers, nil
			}
		}

		maintainer, err := askForMaintainer()
		if err != nil {
			return nil, errors.Wrap(err, "while asking for maintainer details")
		}
		maintainers = append(maintainers, maintainer)
	}
}

func askForMaintainer() (common.Maintainers, error) {
	var maintainer common.Maintainers
	var qs = []*survey.Question{
		{
			Name: "Email",
			Prompt: &survey.Input{
				Message: "Maintainer's e-mail address",
				Default: "",
			},
			Validate: common.ManyValidators([]survey.Validator{survey.Required, common.ValidateEmail}),
		},
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "Maintainer's name",
				Default: "",
			},
			Validate: survey.Required,
		},
		{
			Name: "URL",
			Prompt: &survey.Input{
				Message: "Maintainer's URL",
				Default: "",
			},
			Validate: common.ManyValidators([]survey.Validator{survey.Required, common.ValidateURL}),
		},
	}
	err := survey.Ask(qs, &maintainer)
	return maintainer, err
}

func askIfOverwrite() (bool, error) {
	overwrite := false
	prompt := &survey.Confirm{
		Message: "Do you want to overwrite existing manifest files?",
	}
	err := survey.AskOne(prompt, &overwrite)
	return overwrite, err
}
