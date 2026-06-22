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
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newVersionCommand(version))
	return rootCmd
}

func newSyncCommand(scope string, short string) *cobra.Command {
	return &cobra.Command{
		Use:   scope,
		Short: short,
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

			outputDir, err := defaultOutputDir(scope)
			if err != nil {
				return err
			}

			tools := parseScopeTools(loadedRegistry.Registry, scope, cmd.ErrOrStderr())
			return syncTools(tools, outputDir, cmd.ErrOrStderr())
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
