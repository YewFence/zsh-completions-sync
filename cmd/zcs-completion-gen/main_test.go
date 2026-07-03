package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckSucceedsWhenCompletionsMatchSpecs(t *testing.T) {
	setupCompletionGenTest(t)

	if err := run(nil); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if err := run([]string{"--check"}); err != nil {
		t.Fatalf("run(--check) error = %v", err)
	}
}

func TestRunCheckFailsWithoutMutatingOutdatedCompletion(t *testing.T) {
	setupCompletionGenTest(t)

	outputPath := filepath.Join("completions", "_demo")
	if err := os.WriteFile(outputPath, []byte("stale content\n"), 0o644); err != nil {
		t.Fatalf("write stale completion: %v", err)
	}

	err := run([]string{"--check"})
	if !errors.Is(err, errCompletionsOutOfDate) {
		t.Fatalf("run(--check) error = %v, want %v", err, errCompletionsOutOfDate)
	}
	assertFileContent(t, outputPath, "stale content\n")
}

func TestRunCheckIgnoresManualCompletions(t *testing.T) {
	setupCompletionGenTest(t)

	if err := run(nil); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join("completions", "_manual"), []byte("# manual completion\n"), 0o644); err != nil {
		t.Fatalf("write manual completion: %v", err)
	}

	if err := run([]string{"--check"}); err != nil {
		t.Fatalf("run(--check) error = %v", err)
	}
}

func setupCompletionGenTest(t *testing.T) {
	t.Helper()

	t.Chdir(t.TempDir())
	if err := os.MkdirAll("completion-spec", 0o755); err != nil {
		t.Fatalf("create completion-spec: %v", err)
	}
	if err := os.MkdirAll("completions", 0o755); err != nil {
		t.Fatalf("create completions: %v", err)
	}
	writeFile(t, filepath.Join("completion-spec", "CompletionSpec.cue"), `package completionspec

#Spec: {
	name: string
	description?: string
	commands?: [...#Command]
}

#Command: {
	name: string
	description?: string
}
`)
	writeFile(t, filepath.Join("completion-spec", "demo.cue"), `package completionspec

name: "demo"
commands: [{
	name: "run"
	description: "run command"
}]
`)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(content) != want {
		t.Fatalf("%s content = %q, want %q", path, string(content), want)
	}
}
