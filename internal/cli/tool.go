package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var supportedScopes = map[string]struct{}{
	"global":  {},
	"project": {},
}

type CompletionTool struct {
	Name       string
	Source     any
	Check      any
	PreCommand []string
	Env        map[string]string
}

type WhichCheck struct {
	Executable string
}

type CommandCheck struct {
	Command []string
}

type ListedTool struct {
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	Available     *bool             `json:"available"`
	Scopes        []string          `json:"scopes"`
	PreCommand    []string          `json:"pre_command,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Source        string            `json:"source"`
	ConfigSources []string          `json:"config_sources"`
}

type SyncResult struct {
	OutputDir           string
	GeneratedTools      []string
	SkippedTools        []string
	WarningSkippedTools []string
}

type syncToolResult struct {
	Generated          bool
	SkippedUnavailable bool
	Warning            string
}

func parseScopeTools(registry map[string]any, scope string, stderr io.Writer) []CompletionTool {
	toolTable, ok := registry["tools"].(map[string]any)
	if !ok {
		return nil
	}

	tools := []CompletionTool{}
	for name, config := range toolTable {
		toolConfig, ok := config.(map[string]any)
		if !ok || name == "" {
			continue
		}
		if toolDisabled(toolConfig) {
			continue
		}

		scopes, ok := parseScopes(toolConfig["scopes"])
		if !ok {
			warnTool(name, "invalid scopes config", stderr)
			continue
		}
		if _, ok := scopes[scope]; !ok {
			continue
		}

		tool, ok := parseTool(name, toolConfig, stderr)
		if ok {
			tools = append(tools, tool)
		}
	}

	return tools
}

func filterTools(tools []CompletionTool, names []string) ([]CompletionTool, error) {
	if len(names) == 0 {
		return tools, nil
	}

	requested := map[string]struct{}{}
	for _, name := range names {
		requested[name] = struct{}{}
	}

	filtered := []CompletionTool{}
	for _, tool := range tools {
		if _, ok := requested[tool.Name]; !ok {
			continue
		}
		filtered = append(filtered, tool)
		delete(requested, tool.Name)
	}
	if len(requested) == 0 {
		return filtered, nil
	}

	missing := make([]string, 0, len(requested))
	for name := range requested {
		missing = append(missing, name)
	}
	sort.Strings(missing)
	return nil, fmt.Errorf("unknown tool for selected scope: %s", strings.Join(missing, ", "))
}

func completeToolNames(tools []CompletionTool, args []string) []string {
	seen := map[string]struct{}{}
	for _, arg := range args {
		seen[arg] = struct{}{}
	}

	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		if _, ok := seen[tool.Name]; ok {
			continue
		}
		names = append(names, tool.Name)
	}
	sort.Strings(names)
	return names
}

func parseScopes(value any) (map[string]struct{}, bool) {
	values, ok := value.([]any)
	if !ok || len(values) == 0 {
		return nil, false
	}

	scopes := map[string]struct{}{}
	for _, item := range values {
		scope, ok := item.(string)
		if !ok || scope == "" {
			return nil, false
		}
		if _, ok := supportedScopes[scope]; !ok {
			return nil, false
		}
		scopes[scope] = struct{}{}
	}
	return scopes, true
}

func toolDisabled(config map[string]any) bool {
	disabled, ok := config["disabled"].(bool)
	return ok && disabled
}

func parseTool(name string, config map[string]any, stderr io.Writer) (CompletionTool, bool) {
	source, ok := parseSource(config)
	if !ok {
		warnTool(name, "invalid source config", stderr)
		return CompletionTool{}, false
	}

	check, ok := parseCheck(config["check"], name)
	if !ok {
		warnTool(name, "invalid check config", stderr)
		return CompletionTool{}, false
	}

	preCommand, ok := parsePreCommand(config["pre-command"])
	if !ok {
		warnTool(name, "invalid pre-command config", stderr)
		return CompletionTool{}, false
	}

	env, ok := parseEnv(config["env"])
	if !ok {
		warnTool(name, "invalid env config", stderr)
		return CompletionTool{}, false
	}

	return CompletionTool{Name: name, Source: source, Check: check, PreCommand: preCommand, Env: env}, true
}

func parseCheck(value any, defaultExecutable string) (any, bool) {
	if value == nil {
		return WhichCheck{Executable: defaultExecutable}, true
	}
	if disabled, ok := value.(bool); ok && !disabled {
		return nil, true
	}
	if executable, ok := value.(string); ok && executable != "" {
		return WhichCheck{Executable: executable}, true
	}
	if command, ok := parseCommand(value); ok {
		return CommandCheck{Command: command}, true
	}
	return nil, false
}

func parsePreCommand(value any) ([]string, bool) {
	if value == nil {
		return nil, true
	}
	return parseCommand(value)
}

func parseEnv(value any) (map[string]string, bool) {
	if value == nil {
		return nil, true
	}

	typed, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	env := map[string]string{}
	for key, rawValue := range typed {
		value, ok := rawValue.(string)
		if key == "" || strings.Contains(key, "=") || !ok {
			return nil, false
		}
		env[key] = value
	}
	return env, true
}

func listTools(loadedRegistry LoadedRegistry, scope string, format string, stdout io.Writer, stderr io.Writer) error {
	rows := listedTools(loadedRegistry, scope, stderr)
	if format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(rows)
	}

	if len(rows) == 0 {
		_, err := fmt.Fprintln(stdout, "No configured tools.")
		return err
	}

	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{
			row.Name,
			row.Status,
			formatAvailability(row.Available),
			formatScopes(row.Scopes),
			formatOptionalCommand(row.PreCommand),
			formatEnv(row.Env),
			row.Source,
			strings.Join(row.ConfigSources, " -> "),
		})
	}
	return printTable([]string{"Tool", "Status", "Available", "Scopes", "Pre-command", "Env", "Source", "Config loaded from"}, tableRows, stdout)
}

func listedTools(loadedRegistry LoadedRegistry, scope string, stderr io.Writer) []ListedTool {
	toolTable, ok := loadedRegistry.Registry["tools"].(map[string]any)
	if !ok {
		return nil
	}

	names := make([]string, 0, len(toolTable))
	for name := range toolTable {
		names = append(names, name)
	}
	sort.Strings(names)

	rows := []ListedTool{}
	for _, name := range names {
		config, ok := toolTable[name].(map[string]any)
		if !ok {
			continue
		}
		if toolDisabled(config) {
			rows = append(rows, ListedTool{
				Name:          name,
				Status:        "disabled",
				Available:     nil,
				Scopes:        nil,
				PreCommand:    nil,
				Env:           nil,
				Source:        "-",
				ConfigSources: toolConfigSources(loadedRegistry.Layers, name),
			})
			continue
		}

		scopes, ok := parseScopes(config["scopes"])
		if !ok {
			warnTool(name, "invalid scopes config", stderr)
			continue
		}
		if scope != "" {
			if _, ok := scopes[scope]; !ok {
				continue
			}
		}

		source, ok := parseSource(config)
		if !ok {
			warnTool(name, "invalid source config", stderr)
			continue
		}
		preCommand, ok := parsePreCommand(config["pre-command"])
		if !ok {
			warnTool(name, "invalid pre-command config", stderr)
			continue
		}
		env, ok := parseEnv(config["env"])
		if !ok {
			warnTool(name, "invalid env config", stderr)
			continue
		}
		check, ok := parseCheck(config["check"], name)
		if !ok {
			warnTool(name, "invalid check config", stderr)
			continue
		}
		available := toolEnabled(check, env)

		rows = append(rows, ListedTool{
			Name:          name,
			Status:        "enabled",
			Available:     &available,
			Scopes:        sortedScopes(scopes),
			PreCommand:    preCommand,
			Env:           env,
			Source:        formatSource(source),
			ConfigSources: toolConfigSources(loadedRegistry.Layers, name),
		})
	}

	return rows
}

func toolConfigSources(layers []RegistryLayer, toolName string) []string {
	sources := []string{}
	for _, layer := range layers {
		toolTable, ok := layer.Registry["tools"].(map[string]any)
		if ok {
			if _, exists := toolTable[toolName]; exists {
				sources = append(sources, layer.Label)
			}
		}
	}
	return sources
}

func sortedScopes(scopes map[string]struct{}) []string {
	values := make([]string, 0, len(scopes))
	for scope := range scopes {
		values = append(values, scope)
	}
	sort.Strings(values)
	return values
}

func formatScopes(scopes []string) string {
	if len(scopes) == 0 {
		return "-"
	}
	return strings.Join(scopes, ", ")
}

func formatAvailability(available *bool) string {
	if available == nil {
		return "-"
	}
	if *available {
		return "yes"
	}
	return "no"
}

func formatSource(source any) string {
	switch typed := source.(type) {
	case CommandSource:
		return fmt.Sprintf("command: %s", formatCommand(typed.Command))
	case FileSource:
		switch fileSource := typed.File.(type) {
		case LocalFileSource:
			return fmt.Sprintf("file: %s", fileSource.Path)
		case HTTPFileSource:
			return fmt.Sprintf("http: %s", fileSource.URL)
		case GitFileSource:
			ref := ""
			if fileSource.Ref != "" {
				ref = fmt.Sprintf(" @ %s", fileSource.Ref)
			}
			return fmt.Sprintf("git: %s//%s%s", fileSource.Repository, fileSource.Path, ref)
		default:
			return "unknown"
		}
	default:
		return "unknown"
	}
}

func formatOptionalCommand(command []string) string {
	if len(command) == 0 {
		return "-"
	}
	return formatCommand(command)
}

func formatEnv(env map[string]string) string {
	if len(env) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, env[key]))
	}
	return strings.Join(pairs, " ")
}

func printTable(headers []string, rows [][]string, stdout io.Writer) error {
	widths := make([]int, len(headers))
	for index, header := range headers {
		widths[index] = len(header)
	}
	for _, row := range rows {
		for index, value := range row {
			if len(value) > widths[index] {
				widths[index] = len(value)
			}
		}
	}

	if _, err := fmt.Fprintln(stdout, formatTableRow(headers, widths)); err != nil {
		return err
	}
	separators := make([]string, len(headers))
	for index, width := range widths {
		separators[index] = strings.Repeat("-", width)
	}
	if _, err := fmt.Fprintln(stdout, formatTableRow(separators, widths)); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(stdout, formatTableRow(row, widths)); err != nil {
			return err
		}
	}
	return nil
}

func formatTableRow(row []string, widths []int) string {
	paddedCells := make([]string, len(row))
	for index, value := range row {
		paddedCells[index] = value + strings.Repeat(" ", widths[index]-len(value))
	}
	return strings.TrimRight(strings.Join(paddedCells, "  "), " ")
}

func syncTools(tools []CompletionTool, outputDir string, jobs int, stderr io.Writer) (SyncResult, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return SyncResult{}, err
	}
	if jobs <= 0 {
		return SyncResult{}, fmt.Errorf("jobs must be greater than 0")
	}
	if jobs > len(tools) && len(tools) > 0 {
		jobs = len(tools)
	}

	results := make([]syncToolResult, len(tools))
	work := make(chan int)
	var waitGroup sync.WaitGroup
	for range jobs {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for index := range work {
				results[index] = syncTool(tools[index], outputDir)
			}
		}()
	}
	for index := range tools {
		work <- index
	}
	close(work)
	waitGroup.Wait()

	result := SyncResult{OutputDir: outputDir}
	for index, toolResult := range results {
		toolName := tools[index].Name
		switch {
		case toolResult.Generated:
			result.GeneratedTools = append(result.GeneratedTools, toolName)
		case toolResult.SkippedUnavailable:
			result.SkippedTools = append(result.SkippedTools, toolName)
		case toolResult.Warning != "":
			result.WarningSkippedTools = append(result.WarningSkippedTools, toolName)
			warnToolMessage(toolResult.Warning, stderr)
		}
	}
	sort.Strings(result.GeneratedTools)
	sort.Strings(result.SkippedTools)
	sort.Strings(result.WarningSkippedTools)
	return result, nil
}

func syncTool(tool CompletionTool, outputDir string) syncToolResult {
	if !toolEnabled(tool.Check, tool.Env) {
		return syncToolResult{SkippedUnavailable: true}
	}

	if preCommandError := runPreCommand(tool.PreCommand, tool.Env); preCommandError != "" {
		return syncToolResult{Warning: formatToolWarning(tool.Name, preCommandError)}
	}

	result := readSourceWithEnv(tool.Source, tool.Env)
	if result.Error != "" {
		return syncToolResult{Warning: formatToolWarning(tool.Name, result.Error)}
	}

	destination := filepath.Join(outputDir, fmt.Sprintf("_%s", tool.Name))
	if err := writeAtomic(destination, result.Content); err != nil {
		return syncToolResult{Warning: formatToolWarning(tool.Name, fmt.Sprintf("failed to write %s: %v", destination, err))}
	}
	return syncToolResult{Generated: true}
}

func printSyncSummary(result SyncResult, stdout io.Writer) error {
	generated := len(result.GeneratedTools)
	skipped := len(result.SkippedTools) + len(result.WarningSkippedTools)

	generatedTools := formatToolList(result.GeneratedTools)
	if generatedTools != "" {
		generatedTools = ": " + generatedTools
	}
	if _, err := fmt.Fprintf(stdout, "Generated %d %s in %s%s.\n", generated, pluralize("completion", generated), result.OutputDir, generatedTools); err != nil {
		return err
	}
	if skipped > 0 {
		skippedTools := append([]string(nil), result.SkippedTools...)
		skippedTools = append(skippedTools, result.WarningSkippedTools...)
		sort.Strings(skippedTools)
		if _, err := fmt.Fprintf(stdout, "Skipped %d %s: %s.\n", skipped, pluralize("tool", skipped), formatToolList(skippedTools)); err != nil {
			return err
		}
	}
	return nil
}

func formatToolList(tools []string) string {
	return strings.Join(tools, ", ")
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func runPreCommand(command []string, env map[string]string) string {
	if len(command) == 0 {
		return ""
	}
	if !commandExists(command) {
		return fmt.Sprintf("pre-command not found: %s", command[0])
	}

	result, err := runCommandWithEnv(command, io.Discard, env)
	if err != nil {
		return fmt.Sprintf("failed to run pre-command %s: %v", formatCommand(command), err)
	}
	if result.exitCode == 0 {
		return ""
	}

	message := fmt.Sprintf("pre-command failed with exit code %d: %s", result.exitCode, formatCommand(command))
	stderr := strings.TrimSpace(string(result.stderr))
	if stderr != "" {
		message = fmt.Sprintf("%s; %s", message, stderr)
	}
	return message
}

func toolEnabled(check any, env map[string]string) bool {
	if check == nil {
		return true
	}

	switch typed := check.(type) {
	case WhichCheck:
		_, err := exec.LookPath(typed.Executable)
		return err == nil
	case CommandCheck:
		if _, err := exec.LookPath(typed.Command[0]); err != nil {
			return false
		}
		result, err := runCommandWithEnv(typed.Command, io.Discard, env)
		return err == nil && result.exitCode == 0
	default:
		return false
	}
}

func warnTool(name string, message string, stderr io.Writer) {
	warnToolMessage(formatToolWarning(name, message), stderr)
}

func formatToolWarning(name string, message string) string {
	return fmt.Sprintf("warn: skip %s: %s", name, message)
}

func warnToolMessage(message string, stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, message)
}

func writeAtomic(destination string, content []byte) error {
	tempFile, err := os.CreateTemp(filepath.Dir(destination), "."+filepath.Base(destination)+".*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, destination)
}
