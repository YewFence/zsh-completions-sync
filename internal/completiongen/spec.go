package completiongen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	command := exec.Command("pkl", "eval", "--color=never", "-f", "json", path)
	command.Stderr = os.Stderr
	output, err := command.Output()
	if err != nil {
		return Spec{}, fmt.Errorf("evaluate %s: %w", path, err)
	}

	var spec Spec
	if err := json.Unmarshal(output, &spec); err != nil {
		return Spec{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if err := validateSpec(spec); err != nil {
		return Spec{}, fmt.Errorf("validate %s: %w", path, err)
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
