package cmd

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	zcsOutputDirEnv = "ZCS_OUTPUT_DIR"

	projectConfig     = ".zsh-completions-sync.toml"
	projectConfigDir  = ".config"
	projectConfigFile = "zsh-completions-sync.toml"

	userConfigDir        = "zsh-completions-sync"
	userConfigFile       = "registry.toml"
	userLegacyConfigFile = "zsh-completions-sync-registry.toml"
)

//go:embed registry.toml
var registryFS embed.FS

type RegistryLayer struct {
	Label    string
	Registry map[string]any
}

type LoadedRegistry struct {
	Registry map[string]any
	Layers   []RegistryLayer
}

func resolveOutputDir(registry map[string]any, scope string, flagOutputDir string) (string, error) {
	if flagOutputDir != "" {
		return expandUser(os.ExpandEnv(flagOutputDir)), nil
	}
	if outputDir := os.Getenv(zcsOutputDirEnv); outputDir != "" {
		return expandUser(os.ExpandEnv(outputDir)), nil
	}
	if settingsOutputDir, ok := settingsOutputDir(registry); ok {
		return expandUser(os.ExpandEnv(settingsOutputDir)), nil
	}
	return defaultOutputDir(scope)
}

func settingsOutputDir(registry map[string]any) (string, bool) {
	settings, ok := registry["settings"].(map[string]any)
	if !ok {
		return "", false
	}
	outputDir, ok := settings["output_dir"].(string)
	return outputDir, ok && outputDir != ""
}

func defaultOutputDir(scope string) (string, error) {
	switch scope {
	case "project":
		return filepath.Join(".", ".completions", "zsh"), nil
	case "global":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".zsh", "completions"), nil
	default:
		return "", fmt.Errorf("unsupported scope: %s", scope)
	}
}

func loadRegistry(projectDir string, stderr io.Writer) (LoadedRegistry, error) {
	builtIn, err := readResourceTOML("registry.toml")
	if err != nil {
		return LoadedRegistry{}, err
	}

	userPreferredPath, userFallbackPath := userConfigPaths()
	projectPreferredPath, projectFallbackPath := projectConfigPaths(projectDir)
	layers := []RegistryLayer{
		{Label: "built-in registry", Registry: builtIn},
	}

	userLayer, err := readPreferredConfigLayer("user config", userPreferredPath, userFallbackPath, stderr)
	if err != nil {
		return LoadedRegistry{}, err
	}
	projectLayer, err := readPreferredConfigLayer("project config", projectPreferredPath, projectFallbackPath, stderr)
	if err != nil {
		return LoadedRegistry{}, err
	}
	layers = append(layers, userLayer, projectLayer)

	registry := map[string]any{}
	for _, layer := range layers {
		mergeMapping(registry, layer.Registry)
	}

	return LoadedRegistry{Registry: registry, Layers: layers}, nil
}

func userConfigPaths() (string, string) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			configHome = filepath.Join(home, ".config")
		}
	}
	return filepath.Join(expandUser(configHome), userConfigDir, userConfigFile),
		filepath.Join(expandUser(configHome), userLegacyConfigFile)
}

func projectConfigPaths(projectDir string) (string, string) {
	return filepath.Join(projectDir, projectConfigDir, projectConfigFile),
		filepath.Join(projectDir, projectConfig)
}

func readPreferredConfigLayer(label string, preferredPath string, fallbackPath string, stderr io.Writer) (RegistryLayer, error) {
	preferredExists := fileExists(preferredPath)
	fallbackExists := fileExists(fallbackPath)

	if preferredExists && fallbackExists {
		warnDuplicateConfig(preferredPath, fallbackPath, stderr)
	}

	if preferredExists {
		registry, err := readTOML(preferredPath)
		return RegistryLayer{Label: formatConfigLabel(label, preferredPath), Registry: registry}, err
	}
	registry, err := readTOML(fallbackPath)
	return RegistryLayer{Label: formatConfigLabel(label, fallbackPath), Registry: registry}, err
}

func formatConfigLabel(label string, path string) string {
	if label == "" {
		return path
	}
	return fmt.Sprintf("%s: %s", label, path)
}

func warnDuplicateConfig(preferredPath string, ignoredPath string, stderr io.Writer) {
	_, _ = fmt.Fprintf(
		stderr,
		"warn: duplicate zsh-completions-sync registry config; using %s and ignoring %s\n",
		preferredPath,
		ignoredPath,
	)
}

func readTOML(path string) (map[string]any, error) {
	data := map[string]any{}
	if _, err := toml.DecodeFile(path, &data); err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	return data, nil
}

func readResourceTOML(name string) (map[string]any, error) {
	content, err := registryFS.ReadFile(name)
	if err != nil {
		return nil, err
	}

	data := map[string]any{}
	if _, err := toml.Decode(string(content), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func mergeMapping(base map[string]any, override map[string]any) {
	for key, value := range override {
		valueMap, valueOK := value.(map[string]any)
		baseMap, baseOK := base[key].(map[string]any)
		if valueOK && baseOK {
			mergeMapping(baseMap, valueMap)
			continue
		}
		if valueOK {
			base[key] = cloneMapping(valueMap)
			continue
		}
		base[key] = value
	}
}

func cloneMapping(value map[string]any) map[string]any {
	cloned := map[string]any{}
	for key, item := range value {
		if itemMap, ok := item.(map[string]any); ok {
			cloned[key] = cloneMapping(itemMap)
			continue
		}
		cloned[key] = item
	}
	return cloned
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
