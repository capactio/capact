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
	Version   string
	Revision  string
	Branch    string
	BuildDate string
	BuildUser string
	GoVersion = runtime.Version()
	Platform  = runtime.GOOS + "/" + runtime.GOARCH

	versionInfoTmpl = `
{{.program}}
  version:          {{.version}}
  branch:           {{.branch}}
  revision:         {{.revision}}
  build user:       {{.buildUser}}
  build date:       {{.buildDate}}
  go version:       {{.goVersion}}
  platform:         {{.platform}}
`
)

func NewVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information about this binary",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
}

func printVersion() {
	m := map[string]string{
		"program":   cli.Name,
		"version":   Version,
		"revision":  Revision,
		"branch":    Branch,
		"buildUser": BuildUser,
		"buildDate": BuildDate,
		"goVersion": GoVersion,
		"platform":  Platform,
	}

	t := template.Must(template.New("version").Parse(versionInfoTmpl))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "version", m); err != nil {
		panic(err)
	}

	fmt.Println(strings.TrimSpace(buf.String()))
}
