package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestRootCommand(buffer *bytes.Buffer, args ...string) *cobra.Command {
	command := NewRootCommand("test")
	command.SetOut(buffer)
	command.SetErr(buffer)
	command.SetArgs(args)
	return command
}

func TestRootCommandHelp(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "--help")

	if err := command.Execute(); err != nil {
		t.Fatalf("execute root command: %v", err)
	}

	if got := buffer.String(); !strings.Contains(got, "Synchronize zsh completion scripts") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestCompletionCommand(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := NewRootCommand("test")

	if err := command.GenBashCompletionV2(buffer, true); err != nil {
		t.Fatalf("generate bash completion: %v", err)
	}

	if got := buffer.String(); !strings.Contains(got, "# bash completion V2 for zcs") {
		t.Fatalf("unexpected completion output: %q", got)
	}
}

func TestVersionCommand(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "version")

	if err := command.Execute(); err != nil {
		t.Fatalf("execute version command: %v", err)
	}

	if got := buffer.String(); got != "zcs test\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestListCommand(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "list", "--scope", "project")

	if err := command.Execute(); err != nil {
		t.Fatalf("execute list command: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Tool") || !strings.Contains(output, "kubectl") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestProjectCommandWritesConfiguredCompletion(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	sourcePath := filepath.Join(sourceDir, "_local-tool")
	if err := os.WriteFile(sourcePath, []byte("#compdef local-tool\n"), 0o644); err != nil {
		t.Fatalf("write source completion: %v", err)
	}

	configDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	config := `[tools.local-tool]
scopes = ["project"]
check = false
file = "` + sourcePath + `"
`
	if err := os.WriteFile(filepath.Join(configDir, "zsh-completions-sync.toml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working dir: %v", err)
	}

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute project command: %v", err)
	}

	destination := filepath.Join(tempDir, ".completions", "zsh", "_local-tool")
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read generated completion: %v", err)
	}
	if got := string(content); got != "#compdef local-tool\n" {
		t.Fatalf("unexpected completion content: %q", got)
	}
}
