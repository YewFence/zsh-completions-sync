package main

import "github.com/YewFence/zsh-completions-sync/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
