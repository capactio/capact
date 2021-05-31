package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	docsTargetDir = "./docs/cli/commands"

	frontmatterFormat = `---
title: %s
---

`
)

func NewDocs() *cobra.Command {
	return &cobra.Command{
		Use:    "gen-usage-docs",
		Hidden: true,
		Short:  "Generate usage documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := NewRoot()
			root.DisableAutoGenTag = true

			defaultLinkHandler := func(s string) string { return s }
			return doc.GenMarkdownTreeCustom(root, docsTargetDir, frontmatterFilePrepender, defaultLinkHandler)
		},
	}
}

func frontmatterFilePrepender(filePath string) string {
	fileName := filepath.Base(filePath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	title := strings.Replace(fileNameWithoutExt, "_", " ", -1)
	return fmt.Sprintf(frontmatterFormat, title)
}
