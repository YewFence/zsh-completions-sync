package main

import (
	"fmt"
	"os"

	"github.com/YewFence/zsh-completions-sync/internal/registry"
)

func main() {
	data, err := registry.Builtin()
	if err == nil {
		err = registry.ValidateBuiltin(data)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "builtin registry: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("builtin registry is valid")
}
