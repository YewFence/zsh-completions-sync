package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(output, "Available") {
		t.Fatalf("availability column missing: %q", output)
	}
	for _, unexpected := range []string{"Scopes", "Pre-command", "Config loaded from"} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("compact table contains implementation column %q: %q", unexpected, output)
		}
	}
}

func TestListCommandRendersHomepage(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
homepage = "https://example.com/local-tool"
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "list", "--scope", "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute list command: %v", err)
	}

	if !strings.Contains(buffer.String(), "https://example.com/local-tool") {
		t.Fatalf("homepage missing from output: %q", buffer.String())
	}
}

func TestListCommandSupportsJSONFormat(t *testing.T) {
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
	command := newTestRootCommand(buffer, "list", "--scope", "project", "--format", "json")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute list command: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{`"name": "local-tool"`, `"available": true`, `"scopes": [`} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in json output: %q", expected, output)
		}
	}
}

func TestListCommandJSONIncludesHomepage(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
homepage = "https://example.com/local-tool"
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "list", "--scope", "project", "--format", "json")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute list command: %v", err)
	}
	if !strings.Contains(buffer.String(), `"homepage": "https://example.com/local-tool"`) {
		t.Fatalf("homepage missing from json output: %q", buffer.String())
	}
}

func TestListCommandShowsUnavailableTool(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = "zcs-definitely-missing-tool"
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "list", "--scope", "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute list command: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "local-tool") || !strings.Contains(output, "no") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestInfoCommandShowsMergedToolDetails(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
homepage = "https://example.com/local-tool"
pre-command = ["prepare", "local-tool"]
env = { ZCS_TEST = "value" }
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "info", "local-tool")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute info command: %v", err)
	}
	output := buffer.String()
	for _, expected := range []string{"local-tool", "enabled", "https://example.com/local-tool", "project", "prepare local-tool", "ZCS_TEST=value", sourcePath, "project config"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output: %q", expected, output)
		}
	}
}

func TestInfoCommandShowsDisabledTool(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, `[tools.local-tool]
disabled = true
homepage = "https://example.com/local-tool"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "info", "local-tool")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute info command: %v", err)
	}
	if output := buffer.String(); !strings.Contains(output, "disabled") || !strings.Contains(output, "https://example.com/local-tool") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestInfoCommandRejectsUnknownTool(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "info", "missing-tool")
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), `unknown tool "missing-tool"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInfoCommandSupportsJSONFormat(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
homepage = "https://example.com/local-tool"
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "info", "local-tool", "--format", "json")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute info command: %v", err)
	}
	for _, expected := range []string{`"name": "local-tool"`, `"status": "enabled"`, `"homepage": "https://example.com/local-tool"`} {
		if !strings.Contains(buffer.String(), expected) {
			t.Fatalf("expected %q in json output: %q", expected, buffer.String())
		}
	}
}

func TestGenerateCommandCompletesToolArgs(t *testing.T) {
	tempDir := t.TempDir()
	localSourcePath := writeNamedTestCompletionSource(t, tempDir, "local-tool")
	otherSourcePath := writeNamedTestCompletionSource(t, tempDir, "other-tool")
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
file = "`+localSourcePath+`"

[tools.other-tool]
scopes = ["project"]
check = false
file = "`+otherSourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "__complete", "generate", "--scope", "project", "local-tool", "")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute completion command: %v", err)
	}

	output := buffer.String()
	if strings.Contains(output, "local-tool") || !strings.Contains(output, "other-tool") || !strings.Contains(output, ":4") {
		t.Fatalf("unexpected completion output: %q", output)
	}
}

func TestGenerateCommandDefaultsToGlobalScope(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := writeTestCompletionSource(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["global"]
check = false
file = "`+sourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	outputDir := filepath.Join(tempDir, "global-output")
	t.Setenv("ZCS_OUTPUT_DIR", outputDir)
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(outputDir, "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandSupportsProjectScope(t *testing.T) {
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
	command := newTestRootCommand(buffer, "generate", "--scope", "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
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

func TestGenerateCommandSupportsProjectScopeOutputFlag(t *testing.T) {
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
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "--output", outputDir)
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(outputDir, "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandSupportsProjectScopeOutputEnv(t *testing.T) {
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
	command := newTestRootCommand(buffer, "generate", "--scope", "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(outputDir, "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandSupportsProjectScopeSettingsOutputDir(t *testing.T) {
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
	command := newTestRootCommand(buffer, "generate", "--scope", "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(tempDir, "settings-output", "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandProjectScopeFlagOutputOverridesEnvAndSettings(t *testing.T) {
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
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "--output", flagOutputDir)
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(flagOutputDir, "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandSupportsProjectScopeJobsFlag(t *testing.T) {
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
	command := newTestRootCommand(buffer, "generate", "-s", "project", "--jobs", "1")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(tempDir, ".completions", "zsh", "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandSupportsProjectScopeToolArgs(t *testing.T) {
	tempDir := t.TempDir()
	localSourcePath := writeNamedTestCompletionSource(t, tempDir, "local-tool")
	otherSourcePath := writeNamedTestCompletionSource(t, tempDir, "other-tool")
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
file = "`+localSourcePath+`"

[tools.other-tool]
scopes = ["project"]
check = false
file = "`+otherSourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "local-tool")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	outputDir := filepath.Join(tempDir, ".completions", "zsh")
	assertFileContent(t, filepath.Join(outputDir, "_local-tool"), "#compdef local-tool\n")
	if _, err := os.Stat(filepath.Join(outputDir, "_other-tool")); !os.IsNotExist(err) {
		t.Fatalf("expected unrequested tool to be skipped, stat error: %v", err)
	}
}

func TestGenerateCommandPassesEnvToCommandSource(t *testing.T) {
	tempDir := t.TempDir()
	toolPath := writeEnvCompletionCommand(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
env = { ZCS_TEST_COMPLETION = "local-tool" }
command = ["`+toolPath+`"]
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "local-tool")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(tempDir, ".completions", "zsh", "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandPassesEnvToCommandCheck(t *testing.T) {
	tempDir := t.TempDir()
	toolPath := writeEnvCompletionCommand(t, tempDir)
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = ["`+toolPath+`", "check"]
env = { ZCS_TEST_COMPLETION = "local-tool" }
command = ["`+toolPath+`"]
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "local-tool")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	assertFileContent(t, filepath.Join(tempDir, ".completions", "zsh", "_local-tool"), "#compdef local-tool\n")
}

func TestGenerateCommandPrintsSyncSummary(t *testing.T) {
	tempDir := t.TempDir()
	localSourcePath := writeNamedTestCompletionSource(t, tempDir, "local-tool")
	missingSourcePath := filepath.Join(tempDir, "missing", "_missing-tool")
	writeProjectConfig(t, tempDir, `[tools.local-tool]
scopes = ["project"]
check = false
file = "`+localSourcePath+`"

[tools.missing-tool]
scopes = ["project"]
check = false
file = "`+missingSourcePath+`"

[tools.unavailable-tool]
scopes = ["project"]
check = "zcs-definitely-missing-tool"
file = "`+localSourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "local-tool", "missing-tool", "unavailable-tool")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{
		"Generated 1 completion in .completions/zsh: local-tool.",
		"Skipped 2 tools: missing-tool, unavailable-tool.",
		"warn: skip missing-tool:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output: %q", expected, output)
		}
	}
}

func TestSilentFlagSuppressesCommandOutput(t *testing.T) {
	tempDir := t.TempDir()
	missingSourcePath := filepath.Join(tempDir, "missing", "_missing-tool")
	writeProjectConfig(t, tempDir, `[tools.missing-tool]
scopes = ["project"]
check = false
file = "`+missingSourcePath+`"
`)
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "--silent", "generate", "--scope", "project", "missing-tool")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	if got := buffer.String(); got != "" {
		t.Fatalf("expected silent output, got: %q", got)
	}
}

func TestSilentFlagSuppressesExistingStdoutCommands(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "version", "--silent")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute version command: %v", err)
	}

	if got := buffer.String(); got != "" {
		t.Fatalf("expected silent output, got: %q", got)
	}
}

func TestSilentFlagSuppressesValidationOutput(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, "")
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "--silent", "generate", "--scope", "workspace")
	if err := command.Execute(); err == nil {
		t.Fatal("expected invalid scope error")
	}

	if got := buffer.String(); got != "" {
		t.Fatalf("expected silent output, got: %q", got)
	}
}

func TestGenerateCommandProjectScopeRejectsUnknownToolArg(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, "")
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "missing-tool")
	if err := command.Execute(); err == nil {
		t.Fatal("expected unknown tool error")
	}
}

func TestGenerateCommandRejectsInvalidScope(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, "")
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "--scope", "workspace")
	if err := command.Execute(); err == nil {
		t.Fatal("expected invalid scope error")
	}
}

func TestGenerateCommandRejectsInvalidJobs(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, "")
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "--scope", "project", "--jobs", "0")
	if err := command.Execute(); err == nil {
		t.Fatal("expected invalid jobs error")
	}
}

func TestGenerateCommandProjectScopeSkipsDisabledTool(t *testing.T) {
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
	command := newTestRootCommand(buffer, "generate", "--scope", "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute generate command: %v", err)
	}

	destination := filepath.Join(tempDir, ".completions", "zsh", "_local-tool")
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("expected disabled tool to be skipped, stat error: %v", err)
	}
}

func TestGenerateCommandDoesNotSupportLegacyProjectSubcommand(t *testing.T) {
	tempDir := t.TempDir()
	writeProjectConfig(t, tempDir, "")
	restoreWorkingDir := chdir(t, tempDir)
	defer restoreWorkingDir()

	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "generate", "project")
	if err := command.Execute(); err == nil {
		t.Fatal("expected legacy project subcommand to be rejected")
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

func TestInitGlobalCommandDefaultsToGlobalOnly(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "init", "global")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute init global command: %v", err)
	}

	output := buffer.String()
	if strings.Contains(output, "zcs generate") {
		t.Fatalf("default init global should not run zcs generate: %q", output)
	}
	if strings.Contains(output, "$PWD/.completions/zsh") {
		t.Fatalf("default init global should not include project directory: %q", output)
	}
	if !strings.Contains(output, "ZCS_GLOBAL_OUTPUT_DIR") || !strings.Contains(output, "$HOME/.zsh/completions") || !strings.Contains(output, "compinit") {
		t.Fatalf("unexpected output: %q", output)
	}
	if strings.Contains(output, ".zcompdump") {
		t.Fatalf("global init should use default compinit cache: %q", output)
	}
}

func TestInitProjectCommandSupportsNoFlags(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "init", "project", "--no-sync", "--no-compinit")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute init project command: %v", err)
	}

	output := buffer.String()
	if strings.Contains(output, "zcs generate --scope project") || strings.Contains(output, "compinit") {
		t.Fatalf("unexpected disabled sections: %q", output)
	}
	if !strings.Contains(output, "$PWD/.completions/zsh") || !strings.Contains(output, "$HOME/.zsh/completions") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestInitProjectCommandUsesProjectCompinitDump(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "init", "project", "--no-sync")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute init project command: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, `compinit -d "$_zcs_project_compdump"`) || !strings.Contains(output, `$PWD/.completions/zsh/.zcompdump`) {
		t.Fatalf("project compinit dump missing: %q", output)
	}
	if strings.Contains(output, "zcs generate --scope project") {
		t.Fatalf("project sync command should be disabled: %q", output)
	}
}

func TestInitProjectCommandSyncsProjectScope(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "init", "project")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute init project command: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "zcs generate --scope project") {
		t.Fatalf("project sync command missing: %q", output)
	}
}

func TestCheckUpdateCommandWarnsInsteadOfGenerating(t *testing.T) {
	buffer := new(bytes.Buffer)
	command := newTestRootCommand(buffer, "check-update")
	if err := command.Execute(); err != nil {
		t.Fatalf("execute check-update command: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "${commands[$_zcs_global_tool]}") {
		t.Fatalf("check update snippet missing: %q", output)
	}
	if !strings.Contains(output, "print -u2 -r --") {
		t.Fatalf("check update warning should use stderr: %q", output)
	}
	if !strings.Contains(output, "zcs: Some zsh completion scripts are out of date. Run 'zcs generate' to update them.") {
		t.Fatalf("check update warning missing: %q", output)
	}
	if strings.Count(output, "zcs generate") != 1 {
		t.Fatalf("check update snippet should not generate completions: %q", output)
	}

	testCases := []struct {
		name            string
		completionTime  time.Time
		executableTime  time.Time
		wantStaleOutput string
	}{
		{
			name:           "fresh completion is silent",
			completionTime: time.Unix(200, 0),
			executableTime: time.Unix(100, 0),
		},
		{
			name:            "stale completion warns on stderr",
			completionTime:  time.Unix(100, 0),
			executableTime:  time.Unix(200, 0),
			wantStaleOutput: "zcs: Some zsh completion scripts are out of date. Run 'zcs generate' to update them.\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir := t.TempDir()
			completionDir := filepath.Join(tempDir, "completions")
			binDir := filepath.Join(tempDir, "bin")
			if err := os.MkdirAll(completionDir, 0o755); err != nil {
				t.Fatalf("create completion dir: %v", err)
			}
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				t.Fatalf("create bin dir: %v", err)
			}

			completionPath := filepath.Join(completionDir, "_demo")
			if err := os.WriteFile(completionPath, []byte("#compdef demo\n"), 0o644); err != nil {
				t.Fatalf("write completion file: %v", err)
			}
			executablePath := filepath.Join(binDir, "demo")
			if err := os.WriteFile(executablePath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
				t.Fatalf("write executable: %v", err)
			}
			if err := os.Chtimes(completionPath, testCase.completionTime, testCase.completionTime); err != nil {
				t.Fatalf("set completion timestamp: %v", err)
			}
			if err := os.Chtimes(executablePath, testCase.executableTime, testCase.executableTime); err != nil {
				t.Fatalf("set executable timestamp: %v", err)
			}

			zsh := exec.Command("zsh", "-f")
			zsh.Env = append(os.Environ(),
				"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"ZCS_GLOBAL_OUTPUT_DIR="+completionDir,
			)
			zsh.Stdin = strings.NewReader(output)
			var stdout, stderr bytes.Buffer
			zsh.Stdout = &stdout
			zsh.Stderr = &stderr
			if err := zsh.Run(); err != nil {
				t.Fatalf("run check-update snippet: %v", err)
			}
			if got := stdout.String(); got != "" {
				t.Fatalf("unexpected stdout: %q", got)
			}
			if got := stderr.String(); got != testCase.wantStaleOutput {
				t.Fatalf("unexpected stderr: %q", got)
			}
		})
	}
}

func writeTestCompletionSource(t *testing.T, tempDir string) string {
	t.Helper()

	return writeNamedTestCompletionSource(t, tempDir, "local-tool")
}

func writeNamedTestCompletionSource(t *testing.T, tempDir string, name string) string {
	t.Helper()

	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	sourcePath := filepath.Join(sourceDir, "_"+name)
	if err := os.WriteFile(sourcePath, []byte("#compdef "+name+"\n"), 0o644); err != nil {
		t.Fatalf("write source completion: %v", err)
	}
	return sourcePath
}

func writeEnvCompletionCommand(t *testing.T, tempDir string) string {
	t.Helper()

	toolPath := filepath.Join(tempDir, "env-completion")
	content := `#!/bin/sh
if [ "$ZCS_TEST_COMPLETION" != "local-tool" ]; then
  exit 1
fi
if [ "$1" = "check" ]; then
  exit 0
fi
printf '#compdef %s\n' "$ZCS_TEST_COMPLETION"
`
	if err := os.WriteFile(toolPath, []byte(content), 0o755); err != nil {
		t.Fatalf("write test command: %v", err)
	}
	return toolPath
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
