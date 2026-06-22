package cmd

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
	rootCmd.AddCommand(newSyncCommand("project", "Generate project-local completions."))
	rootCmd.AddCommand(newSyncCommand("global", "Generate global completions."))
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newVersionCommand(version))
	return rootCmd
}

func newSyncCommand(scope string, short string) *cobra.Command {
	var outputDir string
	var jobs int

	command := &cobra.Command{
		Use:   scope + " [tool...]",
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
	var project bool
	var globalSync bool
	var noSync bool
	var noCompinit bool

	command := &cobra.Command{
		Use:   "init",
		Short: "Print a zsh initialization snippet.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := InitOptions{
				Project:    project,
				GlobalSync: globalSync,
				Sync:       !noSync,
				Compinit:   !noCompinit,
			}
			return writeInitScript(options, cmd.OutOrStdout())
		},
	}
	command.Flags().BoolVar(&project, "project", false, "Include project-local completions and run zcs project before updating fpath.")
	command.Flags().BoolVar(&globalSync, "global-sync", false, "Refresh stale global completions before updating fpath.")
	command.Flags().BoolVar(&noSync, "no-sync", false, "Do not run zcs project in the generated snippet.")
	command.Flags().BoolVar(&noCompinit, "no-compinit", false, "Do not include autoload -Uz compinit and compinit in the generated snippet.")
	return command
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
