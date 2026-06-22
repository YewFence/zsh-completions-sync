package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCommand(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "zcs",
		Short: "Synchronize zsh completion scripts",
		Long:  "Synchronize zsh completion scripts into global and project-local completion directories.",
	}
	rootCmd.AddCommand(newGenerateCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newCheckUpdateCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newVersionCommand(version))
	return rootCmd
}

func newGenerateCommand() *cobra.Command {
	command := newGenerateScopeCommand("global", "generate [tool...]", "Generate global completions.")
	command.AddCommand(newGenerateScopeCommand("global", "global [tool...]", "Generate global completions."))
	command.AddCommand(newGenerateScopeCommand("project", "project [tool...]", "Generate project-local completions."))
	return command
}

func newGenerateScopeCommand(scope string, use string, short string) *cobra.Command {
	var outputDir string
	var jobs int

	command := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}

			loadedRegistry, err := loadRegistry(projectDir, cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			resolvedOutputDir, err := resolveOutputDir(loadedRegistry.Registry, scope, outputDir)
			if err != nil {
				return err
			}

			tools := parseScopeTools(loadedRegistry.Registry, scope, cmd.ErrOrStderr())
			tools, err = filterTools(tools, args)
			if err != nil {
				return err
			}
			return syncTools(tools, resolvedOutputDir, jobs, cmd.ErrOrStderr())
		},
	}
	command.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated completion scripts.")
	command.Flags().IntVarP(&jobs, "jobs", "j", 8, "Maximum number of tools to synchronize concurrently.")
	return command
}

func newInitCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "init",
		Short: "Print a zsh initialization snippet.",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(newInitGlobalCommand())
	command.AddCommand(newInitProjectCommand())
	return command
}

func newInitGlobalCommand() *cobra.Command {
	var noCompinit bool

	command := &cobra.Command{
		Use:   "global",
		Short: "Print a global zsh initialization snippet.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := InitOptions{
				Sync:     true,
				Compinit: !noCompinit,
			}
			return writeInitScript(options, cmd.OutOrStdout())
		},
	}
	command.Flags().BoolVar(&noCompinit, "no-compinit", false, "Do not include autoload -Uz compinit and compinit in the generated snippet.")
	return command
}

func newInitProjectCommand() *cobra.Command {
	var noSync bool
	var noCompinit bool

	command := &cobra.Command{
		Use:   "project",
		Short: "Print a project-local zsh initialization snippet.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := InitOptions{
				Project:  true,
				Sync:     !noSync,
				Compinit: !noCompinit,
			}
			return writeInitScript(options, cmd.OutOrStdout())
		},
	}
	command.Flags().BoolVar(&noSync, "no-sync", false, "Do not run zcs generate project in the generated snippet.")
	command.Flags().BoolVar(&noCompinit, "no-compinit", false, "Do not include autoload -Uz compinit and compinit in the generated snippet.")
	return command
}

func newCheckUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check-update",
		Short: "Print a zsh snippet that refreshes stale global completions.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeCheckUpdateScript(cmd.OutOrStdout())
		},
	}
}

func newListCommand() *cobra.Command {
	var scope string

	command := &cobra.Command{
		Use:   "list",
		Short: "List configured completion tools.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}

			loadedRegistry, err := loadRegistry(projectDir, cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			return listTools(loadedRegistry, scope, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	command.Flags().StringVar(&scope, "scope", "", "Only show tools enabled for the selected scope.")
	_ = command.RegisterFlagCompletionFunc("scope", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"global", "project"}, cobra.ShellCompDirectiveNoFileComp
	})
	command.PreRunE = func(cmd *cobra.Command, args []string) error {
		if scope == "" || scope == "global" || scope == "project" {
			return nil
		}
		return fmt.Errorf("invalid scope %q, expected global or project", scope)
	}
	return command
}

func Execute(version string) {
	rootCmd := NewRootCommand(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
