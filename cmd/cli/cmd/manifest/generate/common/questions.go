package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

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
func AskForManifestRevision(msg string) (string, error) {
	var manifestRevision string
	prompt := []*survey.Question{
		{
			Prompt: &survey.Input{
				Message: msg,
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
