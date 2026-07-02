package completiongen

import (
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

type Spec struct {
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Helpers            map[string]string   `json:"helpers"`
	Values             map[string][]string `json:"values"`
	OptionGroups       map[string][]Option `json:"optionGroups"`
	GlobalOptionGroups []string            `json:"globalOptionGroups"`
	GlobalOptions      []Option            `json:"globalOptions"`
	Commands           []Command           `json:"commands"`
}

type Command struct {
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Aliases      []string   `json:"aliases"`
	OptionGroups []string   `json:"optionGroups"`
	Options      []Option   `json:"options"`
	Arguments    []Argument `json:"arguments"`
	Commands     []Command  `json:"commands"`
}

type Option struct {
	Flags       []string `json:"flags"`
	Description string   `json:"description"`
	Argument    string   `json:"argument"`
	Completion  string   `json:"completion"`
	Repeatable  bool     `json:"repeatable"`
}

type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Completion  string `json:"completion"`
	Repeatable  bool   `json:"repeatable"`
}

func LoadSpec(path string) (Spec, error) {
	schemaPath := filepath.Join(filepath.Dir(path), "CompletionSpec.cue")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return Spec{}, fmt.Errorf("read schema %s: %w", schemaPath, err)
	}

	ctx := cuecontext.New()

	schemaValue := ctx.CompileString(string(schemaBytes), cue.Filename(schemaPath))
	if err := schemaValue.Err(); err != nil {
		return Spec{}, fmt.Errorf("compile schema %s: %w", schemaPath, err)
	}

	specSchema := schemaValue.LookupPath(cue.ParsePath("#Spec"))
	if !specSchema.Exists() {
		return Spec{}, fmt.Errorf("schema %s does not define #Spec", schemaPath)
	}

	specBytes, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("read spec %s: %w", path, err)
	}

	specValue := ctx.CompileString(string(specBytes), cue.Filename(path))
	if err := specValue.Err(); err != nil {
		return Spec{}, fmt.Errorf("compile spec %s: %w", path, err)
	}

	unified := specSchema.Unify(specValue)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return Spec{}, fmt.Errorf("validate spec %s: %w", path, err)
	}

	var spec Spec
	if err := unified.Decode(&spec); err != nil {
		return Spec{}, fmt.Errorf("decode spec %s: %w", path, err)
	}
	if err := validateSpec(spec); err != nil {
		return Spec{}, fmt.Errorf("validate spec %s: %w", path, err)
	}
	return spec, nil
}

func validateSpec(spec Spec) error {
	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}
	for group := range spec.OptionGroups {
		if group == "" {
			return fmt.Errorf("option group name is required")
		}
	}
	if err := validateOptions(spec.GlobalOptions); err != nil {
		return err
	}
	for name, options := range spec.OptionGroups {
		if err := validateOptions(options); err != nil {
			return fmt.Errorf("option group %s: %w", name, err)
		}
	}
	return validateCommands(spec.Commands)
}

func validateCommands(commands []Command) error {
	for _, command := range commands {
		if command.Name == "" {
			return fmt.Errorf("command name is required")
		}
		if err := validateOptions(command.Options); err != nil {
			return fmt.Errorf("command %s: %w", command.Name, err)
		}
		if err := validateArguments(command.Arguments); err != nil {
			return fmt.Errorf("command %s: %w", command.Name, err)
		}
		if err := validateCommands(command.Commands); err != nil {
			return fmt.Errorf("command %s: %w", command.Name, err)
		}
	}
	return nil
}

func validateOptions(options []Option) error {
	for _, option := range options {
		if len(option.Flags) == 0 {
			return fmt.Errorf("option flags are required")
		}
		for _, flag := range option.Flags {
			if flag == "" {
				return fmt.Errorf("option flag is required")
			}
		}
	}
	return nil
}

func validateArguments(arguments []Argument) error {
	for _, argument := range arguments {
		if argument.Name == "" {
			return fmt.Errorf("argument name is required")
		}
	}
	return nil
}
