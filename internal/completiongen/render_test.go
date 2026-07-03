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

func TestRenderZshUsesCommandDepthForDispatch(t *testing.T) {
	output, err := RenderZsh(Spec{
		Name: "demo",
		Commands: []Command{
			{
				Name:        "get",
				Description: "get item",
				Arguments: []Argument{
					{Name: "target", Completion: "_files"},
				},
			},
			{
				Name:        "group",
				Description: "manage group",
				Commands: []Command{
					{
						Name:        "add",
						Description: "add item",
						Arguments: []Argument{
							{Name: "item", Completion: "_files"},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderZsh() error = %v", err)
	}

	content := string(output)
	assertContains(t, content, `curcontext="${curcontext%:*:*}:demo-$line[1]:"`)
	assertContains(t, content, "case $line[1] in")
	assertContains(t, content, "case $line[2] in")
	assertContains(t, content, "'2:command:->command'")
	assertContains(t, content, "'2:target:_files'")
	assertContains(t, content, "'3:item:_files'")
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
