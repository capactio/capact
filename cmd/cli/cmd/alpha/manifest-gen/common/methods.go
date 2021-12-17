package common

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// ValidateFun is a function that validates the user's answers. It is used in the survey library.
type ValidateFun func(ans interface{}) error

// AskForDirectory asks for a directory. It suggests to a user a list of dirs that can be used.
func AskForDirectory(msg string, defaultDir string) (string, error) {
	chosenDir := ""
	directoryPrompt := &survey.Input{
		Message: msg,
		Suggest: func(toComplete string) []string {
			files, err := filepath.Glob(toComplete + "*")
			if err != nil {
				fmt.Println("Cannot getting the names of files")
				return nil
			}
			var dirs []string
			for _, match := range files {
				file, err := os.Stat(match)
				if err != nil {
					fmt.Println("Cannot getting the information about the file")
					return nil
				}
				if file.IsDir() {
					dirs = append(dirs, match)
				}
			}
			return dirs
		},
	}
	if defaultDir != "" {
		directoryPrompt.Default = defaultDir
	}

	err := survey.AskOne(directoryPrompt, &chosenDir)
	return chosenDir, err
}

// AskForManifestPathSuffix asks for a manifest path suffix.
func AskForManifestPathSuffix(msg string) (string, error) {
	var manifestPath string
	prompt := []*survey.Question{
		{
			Prompt: &survey.Input{
				Message: msg,
			},
			Validate: func(ans interface{}) error {
				if str, ok := ans.(string); !ok || len(strings.Split(str, ".")) < 2 {
					return errors.New(`manifest path suffix must be in format "[PREFIX].[NAME]"`)
				}
				return nil
			},
		},
	}
	err := survey.Ask(prompt, &manifestPath)
	return manifestPath, err
}

// AskForManifestRevision asks for a manifest path revision.
func AskForManifestRevision() (string, error) {
	var manifestRevision string
	prompt := []*survey.Question{
		{
			Prompt: &survey.Input{
				Message: "Revision of the manifests",
				Default: "0.1.0",
			},
			Validate: func(ans interface{}) error {
				if str, ok := ans.(string); !ok || len(strings.Split(str, ".")) < 3 {
					return errors.New(`revision must be in format "[major].[minor].[patch]"`)
				}
				return nil
			},
		},
	}
	err := survey.Ask(prompt, &manifestRevision)
	return manifestRevision, err
}

// CreateManifestPath create a manifest path based on a manifest type and suffix.
func CreateManifestPath(manifestType string, suffix string) string {
	suffixes := map[string]string{
		AttributeManifest:      "attribute",
		TypeManifest:           "type",
		InterfaceManifest:      "interface",
		InterfaceGroupManifest: "interfaceGroup",
		ImplementationManifest: "implementation",
	}
	return "cap." + suffixes[manifestType] + "." + suffix
}

// AddRevisionToPath adds revision to manifest path
func AddRevisionToPath(path string, revision string) string {
	return path + ":" + revision
}

// GetDefaultMetadata creates a new Metadata object and sets default values.
func GetDefaultMetadata() Metadata {
	var metadata Metadata
	metadata.DocumentationURL = "https://example.com"
	metadata.SupportURL = "https://example.com"
	metadata.IconURL = "https://example.com/icon.png"
	metadata.Maintainers = []Maintainers{
		{
			Email: "dev@example.com",
			Name:  "Example Dev",
			URL:   "https://example.com",
		},
	}
	metadata.License.Name = &ApacheLicense
	return metadata
}

//ManyValidators allow using many validators function in the Survey validator.
func ManyValidators(validateFuns []ValidateFun) func(ans interface{}) error {
	return func(ans interface{}) error {
		for _, fun := range validateFuns {
			if err := fun(ans); err != nil {
				return err
			}
		}
		return nil
	}
}

// ValidateURL validates a URL.
func ValidateURL(ans interface{}) error {
	if str, ok := ans.(string); !ok || (ans != "" && !isURL(str)) {
		return errors.New("URL is not valid")
	}
	return nil
}

// ValidateEmail validates an email.
func ValidateEmail(ans interface{}) error {
	if str, ok := ans.(string); !ok || (ans != "" && !isEmail(str)) {
		return errors.New("email is not valid")
	}
	return nil
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
