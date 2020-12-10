package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docsTargetDir = "./cmd/ocftool/docs"

func NewDocs() *cobra.Command {
	return &cobra.Command{
		Use:    "gen-usage-docs",
		Hidden: true,
		Short:  "Generate usage docs",
		Run: func(cmd *cobra.Command, args []string) {
			root := NewRoot()
			root.DisableAutoGenTag = true

			err := doc.GenMarkdownTree(root, docsTargetDir)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}
