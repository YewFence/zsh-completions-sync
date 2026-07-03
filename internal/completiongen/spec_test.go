package completiongen

import (
	"strings"
	"testing"
)

func TestValidateSpecRejectsUnknownGlobalOptionGroup(t *testing.T) {
	err := validateSpec(Spec{
		Name:               "demo",
		GlobalOptionGroups: []string{"missing"},
	})
	assertValidationError(t, err, "option group missing is not defined")
}

func TestValidateSpecRejectsUnknownCommandOptionGroup(t *testing.T) {
	err := validateSpec(Spec{
		Name: "demo",
		Commands: []Command{
			{Name: "run", OptionGroups: []string{"missing"}},
		},
	})
	assertValidationError(t, err, "command run: option group missing is not defined")
}

func TestValidateSpecRejectsUnknownOptionValueRef(t *testing.T) {
	err := validateSpec(Spec{
		Name: "demo",
		GlobalOptions: []Option{
			{Flags: []string{"--format"}, Completion: "value:formats"},
		},
	})
	assertValidationError(t, err, "value formats is not defined")
}

func TestValidateSpecRejectsUnknownArgumentValuesRef(t *testing.T) {
	err := validateSpec(Spec{
		Name: "demo",
		Commands: []Command{
			{
				Name: "run",
				Arguments: []Argument{
					{Name: "format", Completion: "values:formats"},
				},
			},
		},
	})
	assertValidationError(t, err, "command run: value formats is not defined")
}

func TestValidateSpecAcceptsKnownRefs(t *testing.T) {
	err := validateSpec(Spec{
		Name: "demo",
		Values: map[string][]string{
			"formats": {"json", "yaml"},
		},
		OptionGroups: map[string][]Option{
			"output": {
				{Flags: []string{"--format"}, Completion: "value:formats"},
			},
		},
		GlobalOptionGroups: []string{"output"},
		Commands: []Command{
			{
				Name:         "run",
				OptionGroups: []string{"output"},
				Arguments: []Argument{
					{Name: "format", Completion: "values:formats"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("validateSpec() error = %v", err)
	}
}

func assertValidationError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("validateSpec() error = nil, want %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("validateSpec() error = %q, want to contain %q", err.Error(), want)
	}
}
