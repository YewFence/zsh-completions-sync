package registry

import "testing"

func TestValidateBuiltin(t *testing.T) {
	data, err := Builtin()
	if err != nil {
		t.Fatalf("load builtin registry: %v", err)
	}
	if err := ValidateBuiltin(data); err != nil {
		t.Fatalf("validate builtin registry: %v", err)
	}
}

func TestValidateBuiltinRejectsMissingHomepage(t *testing.T) {
	data := map[string]any{"tools": map[string]any{"demo": map[string]any{}}}
	if err := ValidateBuiltin(data); err == nil {
		t.Fatal("expected missing homepage error")
	}
}

func TestValidateBuiltinRejectsInvalidHomepage(t *testing.T) {
	for _, homepage := range []string{"", "example.com", "ftp://example.com", "https:///path"} {
		t.Run(homepage, func(t *testing.T) {
			data := map[string]any{"tools": map[string]any{"demo": map[string]any{"homepage": homepage}}}
			if err := ValidateBuiltin(data); err == nil {
				t.Fatalf("expected invalid homepage error for %q", homepage)
			}
		})
	}
}
