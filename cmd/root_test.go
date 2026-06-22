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

func TestProjectCommandSupportsOutputFlag(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	outputDir := filepath.Join(tempDir, "custom-output")
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project", "--output", outputDir)
	if err := command.Execute(); err != nil {
		t.Fatalf("execute project command: %v", err)
	}

	assertFileContent(t, filepath.Join(outputDir, "_local-tool"), "#compdef local-tool\n")
}

func TestProjectCommandSupportsOutputEnv(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	outputDir := filepath.Join(tempDir, "env-output")
	t.Setenv("ZCS_OUTPUT_DIR", outputDir)
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute project command: %v", err)
	}

	assertFileContent(t, filepath.Join(outputDir, "_local-tool"), "#compdef local-tool\n")
}

func TestProjectCommandSupportsSettingsOutputDir(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[settings]
output_dir = "`+filepath.Join(tempDir, "settings-output")+`"

[tools.local-tool]
scopes = ["project"]
check = false
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute project command: %v", err)
	}

	assertFileContent(t, filepath.Join(tempDir, "settings-output", "_local-tool"), "#compdef local-tool\n")
}

func TestProjectCommandFlagOutputOverridesEnvAndSettings(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[settings]
output_dir = "`+filepath.Join(tempDir, "settings-output")+`"

[tools.local-tool]
scopes = ["project"]
check = false
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	flagOutputDir := filepath.Join(tempDir, "flag-output")
	t.Setenv("ZCS_OUTPUT_DIR", filepath.Join(tempDir, "env-output"))
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project", "--output", flagOutputDir)
	if err := command.Execute(); err != nil {
		t.Fatalf("execute project command: %v", err)
	}

	assertFileContent(t, filepath.Join(flagOutputDir, "_local-tool"), "#compdef local-tool\n")
}

func TestProjectCommandSupportsJobsFlag(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project", "--jobs", "1")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute project command: %v", err)
	}

	assertFileContent(t, filepath.Join(tempDir, ".completions", "zsh", "_local-tool"), "#compdef local-tool\n")
}

func TestProjectCommandRejectsInvalidJobs(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, "")
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project", "--jobs", "0")
	if err := command.Execute(); err == nil {
		t.Fatal("expected invalid jobs error")
	}
}

func TestProjectCommandSkipsDisabledTool(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
disabled = true
scopes = ["project"]
check = false
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute project command: %v", err)
	}

	destination := filepath.Join(tempDir, ".completions", "zsh", "_local-tool")
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("expected disabled tool to be skipped, stat error: %v", err)
	}
}

func TestListCommandShowsDisabledTool(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, `[tools.local-tool]
disabled = true
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "list")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute list command: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "local-tool") || !strings.Contains(output, "disabled") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestInitCommandDefaultsToGlobalOnly(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "init")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute init command: %v", err)
	}

	output := buffer.String()
	if strings.Contains(output, "zcs project") {
		t.Fatalf("default init should not run zcs project: %q", output)
	}
	if strings.Contains(output, "$PWD/.completions/zsh") {
		t.Fatalf("default init should not include project directory: %q", output)
	}
	if !strings.Contains(output, "$HOME/.zsh/completions") || !strings.Contains(output, "compinit") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestInitCommandSupportsProjectAndNoFlags(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "init", "--project", "--no-sync", "--no-compinit")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute init command: %v", err)
	}

	output := buffer.String()
	if strings.Contains(output, "zcs project") || strings.Contains(output, "compinit") {
		t.Fatalf("unexpected disabled sections: %q", output)
	}
	if !strings.Contains(output, "$PWD/.completions/zsh") || !strings.Contains(output, "$HOME/.zsh/completions") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func writeTestCompletionSource(t *testing.T, tempDir string) string {
	t.Helper()

	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	sourcePath := filepath.Join(sourceDir, "_local-tool")
	if err := os.WriteFile(sourcePath, []byte("#compdef local-tool\n"), 0o644); err != nil {
		t.Fatalf("write source completion: %v", err)
	}
	return sourcePath
}

func writeProjectConfig(t *testing.T, tempDir string, content string) {
	t.Helper()

	configDir := filepath.Join(tempDir, ".config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "zsh-completions-sync.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working dir: %v", err)
	}
	return func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	}
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	if got := string(content); got != want {
		t.Fatalf("unexpected content for %s: %q", path, got)
	}
}
