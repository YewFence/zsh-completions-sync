package registry

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func ValidateBuiltin(data map[string]any) error {
	tools, ok := data["tools"].(map[string]any)
	if !ok {
		return fmt.Errorf("tools table is missing")
	}

	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		config, ok := tools[name].(map[string]any)
		if !ok {
			return fmt.Errorf("tool %q has invalid configuration", name)
		}
		homepage, ok := config["homepage"].(string)
		if !ok || strings.TrimSpace(homepage) == "" {
			return fmt.Errorf("tool %q has no homepage", name)
		}
		parsed, err := url.Parse(homepage)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || !parsed.IsAbs() {
			return fmt.Errorf("tool %q has invalid homepage %q", name, homepage)
		}
	}
	return nil
}
