package registry

import (
	"embed"

	"github.com/pelletier/go-toml/v2"
)

//go:embed builtin.toml
var files embed.FS

func Builtin() (map[string]any, error) {
	content, err := files.ReadFile("builtin.toml")
	if err != nil {
		return nil, err
	}

	data := map[string]any{}
	if err := toml.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	return data, nil
}
