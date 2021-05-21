package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docsTargetDir = "./cmd/cli/docs"

func NewDocs() *cobra.Command {
	return &cobra.Command{
		Use:    "gen-usage-docs",
		Hidden: true,
		Short:  "Generate usage documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := NewRoot()
			root.DisableAutoGenTag = true

			return doc.GenMarkdownTree(root, docsTargetDir)
		},
	}
}
