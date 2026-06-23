package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/YewFence/zsh-completions-sync/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const outputDir = "docs/reference"

func main() {
	rootCmd := cli.NewRootCommand("dev")
	disableAutoGenTag(rootCmd)

	if err := os.RemoveAll(outputDir); err != nil {
		fail(err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fail(err)
	}

	if err := doc.GenMarkdownTreeCustom(rootCmd, outputDir, filePrepender, linkHandler); err != nil {
		fail(err)
	}
}

func disableAutoGenTag(command *cobra.Command) {
	command.DisableAutoGenTag = true
	for _, child := range command.Commands() {
		disableAutoGenTag(child)
	}
}

func filePrepender(filename string) string {
	title := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	return fmt.Sprintf("---\ntitle: %s\n---\n\n", title)
}

func linkHandler(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
