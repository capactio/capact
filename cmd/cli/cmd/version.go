package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime"
	"strings"

	"capact.io/capact/internal/cli"
	"github.com/spf13/cobra"
)

// Build information. Populated at build-time.
var (
	showShort bool

	Version   string
	Revision  string
	Branch    string
	BuildDate string
	GoVersion = runtime.Version()
	Platform  = runtime.GOOS + "/" + runtime.GOARCH

	buildInfoTmpl = `
{{.program}}
  version:          {{.version}}
  branch:           {{.branch}}
  revision:         {{.revision}}
  build date:       {{.buildDate}}
  go version:       {{.goVersion}}
  platform:         {{.platform}}
`
)

// NewVersion returns a cobra.Command for showing Capact CLI binary information.
func NewVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information about this binary",
		Run: func(cmd *cobra.Command, args []string) {
			if showShort {
				printShortVersion()
			} else {
				printVersion()
			}
		},
	}

	cmd.Flags().BoolVar(&showShort, "short", false, "Show short version information")

	return cmd
}

func printShortVersion() {
	fmt.Printf("%s %s on %s\n", cli.Name, Version, Platform)
}

func printVersion() {
	m := map[string]string{
		"program":   cli.Name,
		"version":   Version,
		"revision":  Revision,
		"branch":    Branch,
		"buildDate": BuildDate,
		"goVersion": GoVersion,
		"platform":  Platform,
	}

	t := template.Must(template.New("version").Parse(buildInfoTmpl))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "version", m); err != nil {
		panic(err)
	}

	fmt.Println(strings.TrimSpace(buf.String()))
}
