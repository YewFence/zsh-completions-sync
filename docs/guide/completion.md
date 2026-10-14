# Shell Completion

zcs uses Cobra's native completion support and does not require an additional generator.

## Generate Completion Scripts

After building the binary, use the `completion` subcommand to generate completion scripts for Bash, Zsh, Fish, and PowerShell.

```bash
mise run build
./bin/zcs completion zsh > _zcs
./bin/zcs completion bash > zcs.bash
./bin/zcs completion fish > zcs.fish
./bin/zcs completion powershell > zcs.ps1
```

## Installation Examples

For Zsh, place the generated `_zcs` file in an existing directory from `$fpath`, or place it in a custom directory and add that directory in `~/.zshrc`.

```bash
mkdir -p ~/.zsh/completions
./bin/zcs completion zsh > ~/.zsh/completions/_zcs
```

```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit
compinit
```

For Bash, place the completion script in a local directory and source it manually, or let the system bash-completion directory manage it.

```bash
mkdir -p ~/.bash_completion.d
./bin/zcs completion bash > ~/.bash_completion.d/zcs.bash
source ~/.bash_completion.d/zcs.bash
```

For Fish, write the script directly to the user completion directory.

```bash
mkdir -p ~/.config/fish/completions
./bin/zcs completion fish > ~/.config/fish/completions/zcs.fish
```
