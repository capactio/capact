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
				Message: "What is documentation URL?",
				Default: "",
			},
			Validate: common.ValidateURL,
		},
		{
			Name: "SupportURL",
			Prompt: &survey.Input{
				Message: "What is support URL?",
				Default: "",
			},
			Validate: common.ValidateURL,
		},
		{
			Name: "IconURL",
			Prompt: &survey.Input{
				Message: "What is icon URL?",
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
		message := ""
		if len(maintainers) < 1 {
			message = "Do you want to add maintainer?"
		} else {
			message = "Do you want to add another maintainer?"
		}
		prompt := &survey.Confirm{
			Message: message,
		}
		err := survey.AskOne(prompt, &addMore)
		if err != nil {
			return nil, errors.Wrap(err, "while asking if add maintainers")
		}
		if !addMore {
			return maintainers, nil
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
				Message: "What is email",
				Default: "",
			},
			Validate: common.ManyValidators([]common.ValidateFun{survey.Required, common.ValidateEmail}),
		},
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "What is a name?",
				Default: "",
			},
			Validate: survey.Required,
		},
		{
			Name: "URL",
			Prompt: &survey.Input{
				Message: "What is a Url?",
				Default: "",
			},
			Validate: common.ManyValidators([]common.ValidateFun{survey.Required, common.ValidateURL}),
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
