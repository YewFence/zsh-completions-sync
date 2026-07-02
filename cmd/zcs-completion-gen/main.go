package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/YewFence/zsh-completions-sync/internal/completiongen"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	specPaths := args
	if len(specPaths) == 0 {
		matches, err := filepath.Glob(filepath.Join("completion-spec", "*.cue"))
		if err != nil {
			return err
		}
		for _, match := range matches {
			if strings.HasSuffix(match, "CompletionSpec.cue") {
				continue
			}
			specPaths = append(specPaths, match)
		}
	}

	for _, specPath := range specPaths {
		spec, err := completiongen.LoadSpec(specPath)
		if err != nil {
			return err
		}
		content, err := completiongen.RenderZsh(spec)
		if err != nil {
			return err
		}
		outputPath := filepath.Join("completions", "_"+spec.Name)
		if err := os.WriteFile(outputPath, content, 0o644); err != nil {
			return err
		}
		fmt.Printf("generated %s from %s\n", outputPath, specPath)
	}
	return nil
}
