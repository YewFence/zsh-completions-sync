package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/YewFence/zsh-completions-sync/internal/completiongen"
)

const generatedMarker = "# Generated from completion-spec/"

var errCompletionsOutOfDate = errors.New("generated completions are out of date")

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("zcs-completion-gen", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	check := flags.Bool("check", false, "check generated completions without writing files")
	if err := flags.Parse(args); err != nil {
		return err
	}

	specPaths := flags.Args()
	allSpecs := len(specPaths) == 0
	if allSpecs {
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

	expectedOutputs := map[string]struct{}{}
	outOfDate := false
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
		expectedOutputs[outputPath] = struct{}{}
		if *check {
			current, err := os.ReadFile(outputPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fmt.Fprintf(os.Stderr, "%s is missing, run `mise run completions:gen`\n", outputPath)
					outOfDate = true
					continue
				}
				return err
			}
			if !bytes.Equal(current, content) {
				fmt.Fprintf(os.Stderr, "%s is out of date, run `mise run completions:gen`\n", outputPath)
				outOfDate = true
			}
			continue
		}
		if err := os.WriteFile(outputPath, content, 0o644); err != nil {
			return err
		}
		fmt.Printf("generated %s from %s\n", outputPath, specPath)
	}
	if *check && allSpecs {
		extraOutOfDate, err := checkGeneratedCompletions(expectedOutputs)
		if err != nil {
			return err
		}
		outOfDate = outOfDate || extraOutOfDate
	}
	if outOfDate {
		return errCompletionsOutOfDate
	}
	return nil
}

func checkGeneratedCompletions(expectedOutputs map[string]struct{}) (bool, error) {
	entries, err := os.ReadDir("completions")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	outOfDate := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		path := filepath.Join("completions", entry.Name())
		if _, ok := expectedOutputs[path]; ok {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}
		if bytes.Contains(content, []byte(generatedMarker)) {
			fmt.Fprintf(os.Stderr, "%s is generated but has no matching spec\n", path)
			outOfDate = true
		}
	}
	return outOfDate, nil
}
