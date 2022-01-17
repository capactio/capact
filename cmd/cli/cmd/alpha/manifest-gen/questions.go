package manifestgen

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
)

func askForManifestType() ([]string, error) {
	var manifestTypes []string
	availableManifestsType := []string{
		string(types.AttributeManifestKind),
		string(types.TypeManifestKind),
		string(types.InterfaceGroupManifestKind),
		string(types.InterfaceManifestKind),
		string(types.ImplementationManifestKind)}
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

func askForCommonMetadataInformation() (*types.ImplementationMetadata, error) {
	type Answers struct {
		DocumentationURL string
		SupportURL       string
		IconURL          string
	}
	var answers Answers
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
				Message: "Icon URL",
				Default: "",
			},
			Validate: common.ValidateURL,
		},
	}
	err := survey.Ask(qs, &answers)
	if err != nil {
		return nil, errors.Wrap(err, "while asking for metadata")
	}
	maintainers, err := askForMaintainers()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for maintainers")
	}
	metadata := types.ImplementationMetadata{
		Maintainers: maintainers,
	}
	if answers.DocumentationURL != "" {
		metadata.DocumentationURL = &answers.DocumentationURL
	}
	if answers.SupportURL != "" {
		metadata.SupportURL = &answers.SupportURL
	}
	if answers.IconURL != "" {
		metadata.IconURL = &answers.IconURL
	}
	return &metadata, nil
}

func askForMaintainers() ([]types.Maintainer, error) {
	var maintainers []types.Maintainer
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

func askForMaintainer() (types.Maintainer, error) {
	type Answers struct {
		Email string
		Name  string
		URL   string
	}
	var answer Answers
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
	err := survey.Ask(qs, &answer)
	if err != nil {
		return types.Maintainer{}, err
	}

	maintainer := types.Maintainer{
		Email: answer.Email,
		Name:  &answer.Name,
		URL:   &answer.URL,
	}

	return maintainer, nil
}

func askIfOverwrite() (bool, error) {
	overwrite := false
	prompt := &survey.Confirm{
		Message: "Do you want to overwrite existing manifest files?",
	}
	err := survey.AskOne(prompt, &overwrite)
	return overwrite, err
}
