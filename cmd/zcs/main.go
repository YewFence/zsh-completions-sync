package main

import "github.com/YewFence/zsh-completions-sync/internal/cli"

var version = "dev"

func main() {
	cli.Execute(version)
}
