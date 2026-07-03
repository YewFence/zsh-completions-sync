package completiongen

import (
	"strings"
	"testing"
)

func TestRenderZshOptionSpecs(t *testing.T) {
	output, err := RenderZsh(Spec{
		Name: "demo",
		OptionGroups: map[string][]Option{
			"global": {
				{Flags: []string{"-h", "--help"}, Description: "show help"},
				{Flags: []string{"-o", "--output"}, Description: "write output", Argument: "file", Completion: "_files"},
				{Flags: []string{"--config"}, Description: "read config", Argument: "file", Completion: "_files"},
				{Flags: []string{"--tag"}, Description: "add tag", Argument: "tag", Completion: "_values tags alpha beta", Repeatable: true},
				{Flags: []string{"-v", "--verbose"}, Description: "increase verbosity", Repeatable: true},
			},
		},
		GlobalOptionGroups: []string{"global"},
		Commands: []Command{
			{Name: "run", Description: "run command"},
		},
	})
	if err != nil {
		t.Fatalf("RenderZsh() error = %v", err)
	}

	content := string(output)
	assertContains(t, content, "'(-h --help)'{-h,--help}'[show help]'")
	assertContains(t, content, "'(-o --output)'{-o,--output}'[write output]:file:_files'")
	assertContains(t, content, "'--config=[read config]:file:_files'")
	assertContains(t, content, "'*--tag=[add tag]:tag:_values tags alpha beta'")
	assertContains(t, content, "'*(-v --verbose)'{-v,--verbose}'[increase verbosity]'")
	assertNotContains(t, content, "'(-h --help){-h,--help}[show help]'")
}

func assertContains(t *testing.T, content, want string) {
	t.Helper()
	if !strings.Contains(content, want) {
		t.Fatalf("rendered output does not contain %q\n\n%s", want, content)
	}
}

func assertNotContains(t *testing.T, content, want string) {
	t.Helper()
	if strings.Contains(content, want) {
		t.Fatalf("rendered output contains %q\n\n%s", want, content)
	}
}
