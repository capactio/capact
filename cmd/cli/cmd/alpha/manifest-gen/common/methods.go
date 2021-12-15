package common

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"os"
	"path/filepath"
)

// AskForOutputDirectory ask for a directory output
func AskForOutputDirectory(msg string, defaultValue string) string{
	directory := ""
	directoryPrompt := &survey.Input{
		Message: msg,
		Suggest: func (toComplete string) []string {
			files, _ := filepath.Glob(toComplete + "*")
			var dirs []string
			for _, match := range files {
				file, err := os.Stat(match)
				if err != nil {
					fmt.Println("Cannot stat the file")
				}
				if file.IsDir() {
					dirs = append(dirs, match)
				}
			}
			return dirs
		},
	}
	if defaultValue != "" {
		directoryPrompt.Default = defaultValue
	}

	survey.AskOne(directoryPrompt, &directory)
	return directory
}
