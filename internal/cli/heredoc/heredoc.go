package heredoc

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
)

const cliTag = "<cli>"

// WithCLIName returns unindented and formatted string as here-document.
// Replace all <cli> with a given name.
func WithCLIName(raw string, cliName string) string {
	return strings.ReplaceAll(heredoc.Doc(raw), cliTag, cliName)
}

// Doc returns un-indented string as here-document.
func Doc(raw string) string {
	return heredoc.Doc(raw)
}

// Docf returns unindented and formatted string as here-document.
// Formatting is done as for fmt.Printf().
func Docf(raw string, args ...interface{}) string {
	return heredoc.Docf(raw, args...)
}
