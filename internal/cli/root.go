package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func NewRootCommand(version string) *cobra.Command {
	var silent bool

	rootCmd := &cobra.Command{
		Use:   "zcs",
		Short: "Synchronize zsh completion scripts",
		Long:  "Synchronize zsh completion scripts into global and project-local completion directories.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if silent {
				silenceCommandOutput(cmd.Root())
			}
		},
	}
	rootCmd.PersistentFlags().Var(&silentFlag{enabled: &silent, command: rootCmd}, "silent", "Suppress all command output.")
	rootCmd.PersistentFlags().Lookup("silent").NoOptDefVal = "true"
	rootCmd.AddCommand(newGenerateCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newCheckUpdateCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newVersionCommand(version))
	return rootCmd
}

func newGenerateCommand() *cobra.Command {
	var scope string
	var outputDir string
	var jobs int

	command := &cobra.Command{
		Use:   "generate [tool...]",
		Short: "Generate completion scripts.",
		Args:  cobra.ArbitraryArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			projectDir, err := os.Getwd()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			loadedRegistry, err := loadRegistry(projectDir, cmd.ErrOrStderr())
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			return completeToolNames(parseScopeTools(loadedRegistry.Registry, scope, cmd.ErrOrStderr()), args), cobra.ShellCompDirectiveNoFileComp
		},
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
			result, err := syncTools(tools, resolvedOutputDir, jobs, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			return printSyncSummary(result, cmd.OutOrStdout())
		},
	}
	command.Flags().StringVarP(&scope, "scope", "s", "global", "Generate completions for the selected scope.")
	command.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated completion scripts.")
	command.Flags().IntVarP(&jobs, "jobs", "j", 8, "Maximum number of tools to synchronize concurrently.")
	_ = command.RegisterFlagCompletionFunc("scope", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"global", "project"}, cobra.ShellCompDirectiveNoFileComp
	})
	command.PreRunE = func(cmd *cobra.Command, args []string) error {
		if scope == "global" || scope == "project" {
			return nil
		}
		return fmt.Errorf("invalid scope %q, expected global or project", scope)
	}
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
	command.Flags().BoolVar(&noSync, "no-sync", false, "Do not run zcs generate --scope project in the generated snippet.")
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
	var format string

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

			return listTools(loadedRegistry, scope, format, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	command.Flags().StringVar(&scope, "scope", "", "Only show tools enabled for the selected scope.")
	command.Flags().StringVar(&format, "format", "table", "Output format: table or json.")
	_ = command.RegisterFlagCompletionFunc("scope", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"global", "project"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = command.RegisterFlagCompletionFunc("format", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	command.PreRunE = func(cmd *cobra.Command, args []string) error {
		if scope != "" && scope != "global" && scope != "project" {
			return fmt.Errorf("invalid scope %q, expected global or project", scope)
		}
		if format != "table" && format != "json" {
			return fmt.Errorf("invalid format %q, expected table or json", format)
		}
		return nil
	}
	return command
}

func Execute(version string) {
	rootCmd := NewRootCommand(version)
	silent := argsContainSilent(os.Args[1:])
	if silent {
		silenceCommandOutput(rootCmd)
	}
	if err := rootCmd.Execute(); err != nil {
		if !silent {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

type silentFlag struct {
	enabled *bool
	command *cobra.Command
}

func (flag *silentFlag) Set(value string) error {
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	*flag.enabled = enabled
	if enabled {
		silenceCommandOutput(flag.command)
	}
	return nil
}

func (flag *silentFlag) String() string {
	if flag == nil || flag.enabled == nil {
		return "false"
	}
	return strconv.FormatBool(*flag.enabled)
}

func (flag *silentFlag) Type() string {
	return "bool"
}

func (flag *silentFlag) IsBoolFlag() bool {
	return true
}

func silenceCommandOutput(command *cobra.Command) {
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SilenceErrors = true
	command.SilenceUsage = true
	for _, child := range command.Commands() {
		silenceCommandOutput(child)
	}
}

func argsContainSilent(args []string) bool {
	silent := false
	for _, arg := range args {
		if arg == "--silent" || arg == "--silent=true" {
			silent = true
		}
		if strings.HasPrefix(arg, "--silent=") {
			silent = arg == "--silent=true"
		}
	}
	return silent
}
