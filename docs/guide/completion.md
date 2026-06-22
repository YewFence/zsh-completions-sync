# Shell 补全同步

`zcs` 用来同步 zsh 补全脚本，而不是生成自己的多 shell 补全文件。它会读取内置注册表、用户注册表和项目注册表，然后把不同工具的 zsh 补全写到对应目录。

## 基本命令

全局补全写入 `~/.zsh/completions`。

```sh
zcs global
```

项目补全写入当前目录的 `.completions/zsh`。

```sh
zcs project
```

查看当前合并后的工具注册表。

```sh
zcs list
zcs list --scope global
zcs list --scope project
```

同步时默认最多同时处理 8 个工具，可以用 `--jobs` 调整。

```sh
zcs global --jobs 4
zcs project --jobs 2
```

输出目录可以通过命令行参数、环境变量或配置文件覆盖。优先级从高到低是 `--output`、`ZCS_OUTPUT_DIR`、`[settings] output_dir`、默认目录。

```sh
zcs global --output ~/.local/share/zsh/completions
ZCS_OUTPUT_DIR=.completions/custom zcs project
```

## Zsh 加载

`zcs init` 会输出可以按需加入 `.zshrc` 的脚本片段。默认只加入全局补全目录，并包含 `compinit`。

```zsh
zcs init
```

如果希望同时加载项目级补全，可以使用 `--project`。这个模式会在脚本片段前面调用 `zcs project`，因为项目补全需要先生成再加入 `fpath`。

```zsh
zcs init --project
```

如果你已经用别的机制刷新项目补全，或者想自己管理 `compinit`，可以关掉对应输出。

```zsh
zcs init --project --no-sync
zcs init --no-compinit
```

把项目补全目录放在全局补全目录前面，当前项目的新 shell 就会优先使用项目内补全。

## 注册表位置

用户配置优先读取 `~/.config/zsh-completions-sync/registry.toml`，也兼容读取 `~/.config/zsh-completions-sync-registry.toml`。设置了 `XDG_CONFIG_HOME` 时，用户配置路径会跟随它。

项目配置优先读取当前项目的 `.config/zsh-completions-sync.toml`，也兼容读取 `.zsh-completions-sync.toml`。

## 注册表示例

```toml
[settings]
output_dir = ".completions/zsh"

[tools.mise]
scopes = ["global"]
command = ["mise", "completion", "zsh"]

[tools.pnpm]
disabled = true

[tools.local-tool]
scopes = ["project"]
file = "$PWD/completions/_local-tool"

[tools.installing-tool]
scopes = ["project"]
check = "installing-tool"
pre-command = ["installing-tool", "completion", "install", "--shell", "zsh", "--output", ".completions/vendor/_installing-tool"]
file = ".completions/vendor/_installing-tool"

[tools.remote-tool]
scopes = ["global"]
file = "https://example.com/completions/_remote-tool"

[tools.git-tool]
scopes = ["global", "project"]
file = "git+https://github.com/example/tool.git//completions/_tool?ref=v1.2.3"
```

`command` 从标准输出读取补全脚本，`file` 从文件来源读取补全脚本。如果同时配置两者，`file` 优先。`pre-command` 适合先生成文件再读取的工具。`check` 默认检查工具名，也可以改成可执行文件名、命令数组，或者设为 `false` 关闭检查。

`disabled = true` 可以禁用内置工具或上层配置里的工具。比如在用户配置或项目配置里只写 `[tools.pnpm] disabled = true`，就可以跳过内置的 `pnpm` 配置。`[settings] output_dir` 可以配置默认输出目录，但仍会被 `ZCS_OUTPUT_DIR` 和 `--output` 覆盖。

同步过程会并发处理注册表里的工具。工具不存在会静默跳过，命令失败、文件读取失败或写入失败会输出 `warn`，并保留已有补全文件。
