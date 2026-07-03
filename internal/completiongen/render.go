package completiongen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

func RenderZsh(spec Spec) ([]byte, error) {
	model := newRenderModel(spec)
	var output bytes.Buffer
	tmpl := template.Must(template.New("zsh").Parse(zshTemplate))
	if err := tmpl.Execute(&output, model); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

type renderModel struct {
	Name                string
	Helpers             []string
	Values              []valueModel
	OptionArrays        []optionArrayModel
	Functions           []functionModel
	RootCommands        []commandEntryModel
	RootArguments       []lineModel
	RootCommandArgument string
	RootCases           []caseModel
}

type valueModel struct {
	Function    string
	Tag         string
	Description string
	Items       []string
}

type optionArrayModel struct {
	Name    string
	Options []string
}

type functionModel struct {
	Name            string
	CommandLabel    string
	Subcommands     []commandEntryModel
	Arguments       []lineModel
	Cases           []caseModel
	CommandArgument string
	HasSubcommands  bool
}

type commandEntryModel struct {
	Name        string
	Description string
}

type caseModel struct {
	Pattern   string
	Call      string
	Arguments []lineModel
}

type lineModel struct {
	Text    string
	HasNext bool
}

type modelBuilder struct {
	spec Spec
}

func newRenderModel(spec Spec) renderModel {
	builder := modelBuilder{spec: spec}
	model := renderModel{
		Name:                spec.Name,
		Helpers:             builder.helpers(),
		Values:              builder.values(),
		OptionArrays:        builder.optionArrays(),
		Functions:           builder.functions(),
		RootCommands:        builder.commandEntries(spec.Commands),
		RootArguments:       lineModels(builder.rootArguments()),
		RootCommandArgument: argumentReference(1),
		RootCases:           builder.cases(spec.Commands, nil),
	}
	return model
}

func (builder modelBuilder) helpers() []string {
	names := sortedKeys(builder.spec.Helpers)
	helpers := make([]string, 0, len(names))
	for _, name := range names {
		helper := strings.TrimSpace(builder.spec.Helpers[name])
		if helper != "" {
			helpers = append(helpers, helper)
		}
	}
	return helpers
}

func (builder modelBuilder) values() []valueModel {
	names := sortedKeys(builder.spec.Values)
	values := make([]valueModel, 0, len(names))
	for _, name := range names {
		items := make([]string, 0, len(builder.spec.Values[name]))
		for _, item := range builder.spec.Values[name] {
			items = append(items, shellSingleQuote(item))
		}
		values = append(values, valueModel{
			Function:    builder.valueFunction(name),
			Tag:         shellWord(name),
			Description: shellSingleQuote(singularDescription(name)),
			Items:       items,
		})
	}
	return values
}

func (builder modelBuilder) optionArrays() []optionArrayModel {
	arrays := []optionArrayModel{}
	if len(builder.spec.GlobalOptions) > 0 || len(builder.spec.GlobalOptionGroups) > 0 {
		arrays = append(arrays, optionArrayModel{
			Name:    builder.optionArrayName("global_options"),
			Options: builder.optionSpecs(builder.collectOptions(builder.spec.GlobalOptionGroups, builder.spec.GlobalOptions)),
		})
	}
	for _, name := range sortedKeys(builder.spec.OptionGroups) {
		if contains(builder.spec.GlobalOptionGroups, name) {
			continue
		}
		arrays = append(arrays, optionArrayModel{
			Name:    builder.optionArrayName(name + "_options"),
			Options: builder.optionSpecs(builder.spec.OptionGroups[name]),
		})
	}
	return arrays
}

func (builder modelBuilder) functions() []functionModel {
	functions := []functionModel{}
	var walk func([]Command, []string)
	walk = func(commands []Command, path []string) {
		for _, command := range commands {
			if len(command.Commands) == 0 {
				continue
			}
			commandPath := append(path, command.Name)
			functions = append(functions, functionModel{
				Name:            builder.commandFunctionName(commandPath),
				CommandLabel:    shellSingleQuote(command.Name + " command"),
				Subcommands:     builder.commandEntries(command.Commands),
				Arguments:       lineModels(builder.commandStateArguments(command, commandPath)),
				Cases:           builder.cases(command.Commands, commandPath),
				CommandArgument: argumentReference(commandArgumentPosition(commandPath)),
				HasSubcommands:  len(command.Commands) > 0,
			})
			walk(command.Commands, commandPath)
		}
	}
	walk(builder.spec.Commands, nil)
	return functions
}

func (builder modelBuilder) commandEntries(commands []Command) []commandEntryModel {
	entries := []commandEntryModel{}
	for _, command := range commands {
		entries = append(entries, commandEntryModel{
			Name:        shellSingleQuote(command.Name + ":" + command.Description),
			Description: command.Description,
		})
		for _, alias := range command.Aliases {
			entries = append(entries, commandEntryModel{
				Name:        shellSingleQuote(alias + ":" + command.Description),
				Description: command.Description,
			})
		}
	}
	return entries
}

func (builder modelBuilder) rootArguments() []string {
	lines := []string{}
	if len(builder.spec.GlobalOptions) > 0 || len(builder.spec.GlobalOptionGroups) > 0 {
		lines = append(lines, "$"+builder.optionArrayName("global_options"))
	}
	return append(lines, "'1:command:->command'", "'*::arg:->arg'")
}

func (builder modelBuilder) commandStateArguments(command Command, commandPath []string) []string {
	lines := builder.optionLines(command, true)
	commandPosition := commandArgumentPosition(commandPath)
	return append(lines, shellSingleQuote(fmt.Sprintf("%d:command:->command", commandPosition)), "'*::arg:->arg'")
}

func (builder modelBuilder) leafArguments(command Command, commandPath []string) []string {
	lines := builder.optionLines(command, false)
	for index, argument := range command.Arguments {
		lines = append(lines, builder.argumentSpec(argumentPosition(commandPath, index), argument))
	}
	return lines
}

func (builder modelBuilder) optionLines(command Command, includeGlobal bool) []string {
	lines := []string{}
	if includeGlobal && (len(builder.spec.GlobalOptions) > 0 || len(builder.spec.GlobalOptionGroups) > 0) {
		lines = append(lines, "$"+builder.optionArrayName("global_options"))
	}
	for _, group := range command.OptionGroups {
		lines = append(lines, "$"+builder.optionArrayName(group+"_options"))
	}
	return append(lines, builder.optionSpecs(command.Options)...)
}

func (builder modelBuilder) cases(commands []Command, path []string) []caseModel {
	cases := make([]caseModel, 0, len(commands))
	for _, command := range commands {
		model := caseModel{Pattern: patternForCommand(command)}
		commandPath := append(path, command.Name)
		if len(command.Commands) > 0 {
			model.Call = builder.commandFunctionName(commandPath)
		} else {
			model.Arguments = lineModels(builder.leafArguments(command, commandPath))
		}
		cases = append(cases, model)
	}
	return cases
}

func commandArgumentPosition(commandPath []string) int {
	return len(commandPath) + 1
}

func argumentPosition(commandPath []string, argumentIndex int) int {
	return len(commandPath) + argumentIndex + 1
}

func argumentReference(position int) string {
	return fmt.Sprintf("$line[%d]", position)
}

func lineModels(lines []string) []lineModel {
	models := make([]lineModel, 0, len(lines))
	for index, line := range lines {
		models = append(models, lineModel{
			Text:    line,
			HasNext: index < len(lines)-1,
		})
	}
	return models
}

func (builder modelBuilder) optionSpecs(options []Option) []string {
	specs := make([]string, 0, len(options))
	for _, option := range options {
		specs = append(specs, builder.optionSpec(option))
	}
	return specs
}

func (builder modelBuilder) collectOptions(groups []string, options []Option) []Option {
	collected := []Option{}
	for _, group := range groups {
		collected = append(collected, builder.spec.OptionGroups[group]...)
	}
	return append(collected, options...)
}

func (builder modelBuilder) optionSpec(option Option) string {
	flags := option.Flags
	prefix := ""
	if option.Repeatable {
		prefix = "*"
	}
	description := "[" + escapeZshDescription(option.Description) + "]"
	argument := optionArgumentSpec(option, builder.completionExpression(option.Completion))

	if len(flags) == 1 {
		flagSpec := flags[0]
		if strings.HasPrefix(flags[0], "--") && option.Argument != "" {
			flagSpec = flags[0] + "="
		}
		return shellSingleQuote(prefix + flagSpec + description + argument)
	}

	flagPrefix := "(" + strings.Join(flags, " ") + ")"
	return shellSingleQuote(prefix+flagPrefix) + zshBraceExpansion(flags) + shellSingleQuote(description+argument)
}

func optionArgumentSpec(option Option, completion string) string {
	if option.Argument == "" {
		return ""
	}
	return ":" + escapeZshDescription(option.Argument) + ":" + completion
}

func (builder modelBuilder) argumentSpec(position int, argument Argument) string {
	prefix := fmt.Sprintf("%d", position)
	if argument.Repeatable {
		prefix = "*"
	}
	name := argument.Name
	if argument.Description != "" {
		name = argument.Description
	}
	spec := prefix + ":" + escapeZshDescription(name) + ":" + builder.completionExpression(argument.Completion)
	return shellSingleQuote(spec)
}

func (builder modelBuilder) completionExpression(completion string) string {
	completion = strings.TrimSpace(completion)
	if strings.HasPrefix(completion, "value:") || strings.HasPrefix(completion, "values:") {
		_, name, _ := strings.Cut(completion, ":")
		return builder.valueFunction(name)
	}
	if strings.HasPrefix(completion, "alternative:") {
		return renderAlternative(strings.TrimPrefix(completion, "alternative:"))
	}
	return completion
}

func renderAlternative(value string) string {
	parts := strings.Split(strings.TrimSpace(value), "|")
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, shellSingleQuote(strings.TrimSpace(part)))
	}
	return "_alternative " + strings.Join(quoted, " ")
}

func (builder modelBuilder) commandFunctionName(path []string) string {
	parts := []string{"", builder.spec.Name}
	for _, part := range path {
		parts = append(parts, zshIdent(part))
	}
	return strings.Join(parts, "_")
}

func (builder modelBuilder) optionArrayName(name string) string {
	return "_" + builder.spec.Name + "_" + zshIdent(name)
}

func (builder modelBuilder) valueFunction(name string) string {
	return "_" + builder.spec.Name + "_" + zshIdent(name)
}

func sortedKeys[V any](mapping map[string]V) []string {
	keys := make([]string, 0, len(mapping))
	for key := range mapping {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func patternForCommand(command Command) string {
	values := append([]string{command.Name}, command.Aliases...)
	if len(values) == 1 {
		return values[0]
	}
	return strings.Join(values, "|")
}

func zshIdent(value string) string {
	replacer := strings.NewReplacer("-", "_", ":", "_", ".", "_")
	return replacer.Replace(value)
}

func shellWord(value string) string {
	return strings.NewReplacer("'", "", " ", "-").Replace(value)
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func zshBraceExpansion(values []string) string {
	return "{" + strings.Join(values, ",") + "}"
}

func escapeZshDescription(value string) string {
	replacer := strings.NewReplacer("[", "\\[", "]", "\\]", ":", "\\:")
	return replacer.Replace(value)
}

func singularDescription(value string) string {
	if strings.HasSuffix(value, "ies") {
		return strings.TrimSuffix(value, "ies") + "y"
	}
	if strings.HasSuffix(value, "s") {
		return strings.TrimSuffix(value, "s")
	}
	return value
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

const zshTemplate = `#compdef {{ .Name }}

# Copyright (c) 2026 YewFence
# Generated from completion-spec/{{ .Name }}.cue; do not edit by hand.
{{ range .Helpers }}
{{ . }}
{{ end }}
{{ range .Values }}
{{ .Function }}() {
  local -a values
  values=(
{{- range .Items }}
    {{ . }}
{{- end }}
  )
  _describe -t {{ .Tag }} {{ .Description }} values
}
{{ end }}
{{ range .OptionArrays }}
{{ .Name }}=(
{{- range .Options }}
  {{ . }}
{{- end }}
)
{{ end }}
{{ range .Functions }}
{{ .Name }}() {
  local context state state_descr line curcontext="$curcontext"
  typeset -A opt_args

  local -a subcommands
  subcommands=(
{{- range .Subcommands }}
    {{ .Name }}
{{- end }}
  )

  _arguments -C \
{{- range .Arguments }}
    {{ .Text }}{{ if .HasNext }} \{{ end }}
{{- end }}

  case $state in
    command)
      _describe -t commands {{ .CommandLabel }} subcommands
      ;;
    arg)
      case {{ .CommandArgument }} in
{{- range .Cases }}
        {{ .Pattern }})
{{- if .Call }}
          {{ .Call }}
{{- else if .Arguments }}
          _arguments \
{{- range .Arguments }}
            {{ .Text }}{{ if .HasNext }} \{{ end }}
{{- end }}
{{- end }}
          ;;
{{- end }}
      esac
      ;;
  esac
}
{{ end }}
_{{ .Name }}() {
  local context state state_descr line curcontext="$curcontext"
  typeset -A opt_args

  local -a commands
  commands=(
{{- range .RootCommands }}
    {{ .Name }}
{{- end }}
  )

  _arguments -C \
{{- range .RootArguments }}
    {{ .Text }}{{ if .HasNext }} \{{ end }}
{{- end }}

  case $state in
    command)
      _describe -t commands '{{ .Name }} command' commands
      ;;
    arg)
      curcontext="${curcontext%:*:*}:{{ .Name }}-{{ .RootCommandArgument }}:"
      case {{ .RootCommandArgument }} in
{{- range .RootCases }}
        {{ .Pattern }})
{{- if .Call }}
          {{ .Call }}
{{- else if .Arguments }}
          _arguments \
{{- range .Arguments }}
            {{ .Text }}{{ if .HasNext }} \{{ end }}
{{- end }}
{{- end }}
          ;;
{{- end }}
      esac
      ;;
  esac
}

_{{ .Name }} "$@"
`
