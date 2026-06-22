package cli

import (
	"embed"
	"io"
)

//go:embed init-snippets/*.zsh
var initSnippetFS embed.FS

type InitOptions struct {
	Project    bool
	GlobalSync bool
	Sync       bool
	Compinit   bool
}

func writeInitScript(options InitOptions, stdout io.Writer) error {
	snippets := []string{}
	if options.GlobalSync {
		snippets = append(snippets, "init-snippets/global-sync.zsh")
	}
	if options.Project && options.Sync {
		snippets = append(snippets, "init-snippets/project-sync.zsh")
	}
	if options.Project {
		snippets = append(snippets, "init-snippets/project-fpath.zsh")
	} else {
		snippets = append(snippets, "init-snippets/global-fpath.zsh")
	}
	if options.Compinit {
		snippets = append(snippets, "init-snippets/compinit.zsh")
	}

	return writeInitSnippets(snippets, stdout)
}

func writeCheckUpdateScript(stdout io.Writer) error {
	return writeInitSnippets([]string{"init-snippets/global-sync.zsh"}, stdout)
}

func writeInitSnippets(snippets []string, stdout io.Writer) error {
	for index, snippet := range snippets {
		if index > 0 {
			if _, err := io.WriteString(stdout, "\n"); err != nil {
				return err
			}
		}
		content, err := initSnippetFS.ReadFile(snippet)
		if err != nil {
			return err
		}
		if _, err := stdout.Write(content); err != nil {
			return err
		}
	}
	return nil
}
