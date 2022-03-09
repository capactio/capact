package cmd

import (
	"capact.io/capact/internal/frontmatter"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	docsTargetDir = "./cmd/populator/docs"
)

// NewDocs returns a cobra.Command for generating Populator documentation.
func NewDocs(cliName string) *cobra.Command {
	return &cobra.Command{
		Use:    "gen-usage-docs",
		Hidden: true,
		Short:  "Generate usage documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := NewRoot(cliName)
			root.DisableAutoGenTag = true

			defaultLinkHandler := func(s string) string { return s }
			return doc.GenMarkdownTreeCustom(root, docsTargetDir, frontmatter.FilePrepender, defaultLinkHandler)
		},
	}
}
