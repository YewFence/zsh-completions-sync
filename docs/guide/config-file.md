# 配置文件

`zcs` 通过注册表确认需要生成哪些工具的补全脚本，以及如何生成这些脚本。

注册表使用 TOML 格式。每个工具放在 `[tools.name]` 表中，`name` 通常建议使用可执行文件名。

## 配置位置

注册表分为三层，优先级从低到高依次是内置注册表、用户配置、项目配置。

用户配置优先读取 `~/.config/zsh-completions-sync/registry.toml`，也兼容读取 `~/.config/zsh-completions-sync-registry.toml`。如果两个文件同时存在，只读取前者并输出告警。设置了 `XDG_CONFIG_HOME` 时，用户配置路径会跟随 `XDG_CONFIG_HOME`。

项目配置优先读取当前项目的 `.config/zsh-completions-sync.toml`，也兼容读取 `.zsh-completions-sync.toml`。如果两个文件同时存在，只读取前者并输出告警。

## 最小配置

下面的配置会注册 `mise` 的全局补全。

```toml
[tools.mise]
scopes = ["global"]
command = ["mise", "completion", "zsh"]
```

`scopes` 支持 `global` 和 `project`。`command` 是生成补全脚本的命令，`zcs` 会读取它的标准输出作为补全内容。

## 禁用工具

如果内置注册表里有你不想同步的工具，可以在用户配置或项目配置里禁用它。

```toml
[tools.pnpm]
disabled = true
```

禁用后的工具不会出现在生成列表里。`zcs list` 仍然会显示它的禁用状态，方便确认覆盖是否生效。

## 本地文件来源

如果补全脚本已经存在于本地文件，可以用 `file` 读取。

```toml
[tools.local-tool]
scopes = ["project"]
file = "$PWD/completions/_local-tool"
```

本地路径会展开环境变量，也支持 `~`。

## 预处理命令

有些工具不能把补全脚本直接输出到标准输出，只能先写入固定文件。可以用 `pre-command` 先运行安装或生成动作，再从 `file` 读取结果。

```toml
[tools.installing-tool]
scopes = ["project"]
check = ["installing-tool", "--version"]
pre-command = ["installing-tool", "completion", "install", "--shell", "zsh"]
file = "/usr/share/zsh/vendor-completions/_installing-tool"
```

`pre-command` 的输出不会被当作补全脚本内容。

## 远程文件来源

`file` 可以直接指向远程 URL。

```toml
[tools.remote-tool]
scopes = ["global"]
file = "https://example.com/completions/_remote-tool"
```

远程读取使用 HTTP GET，并带有超时限制。

## Git 文件来源

如果补全脚本位于 Git 仓库中，可以使用 `git+` 字符串形式。

```toml
[tools.git-tool]
scopes = ["global", "project"]
file = "git+https://github.com/example/tool.git//completions/_tool?ref=v1.2.3"
```

也可以使用表格形式。

```toml
[tools.git-tool-alt]
scopes = ["project"]
file = { git = "https://github.com/example/tool.git", path = "completions/_tool", ref = "v1.2.3" }
```

`ref` 可以是分支、标签或提交。如果省略 `ref`，会读取远程仓库默认分支。

## 自定义检查命令

默认情况下，`zcs` 会检查 `PATH` 中是否存在与工具名同名的可执行文件。可以用字符串形式把 `check` 改成其他可执行文件名，也可以用数组形式运行检查命令。

```toml
[tools.node]
scopes = ["project"]
check = "node"
command = ["pnpm", "completion", "zsh"]
```

如果工具不需要检查，可以把 `check` 设置为 `false`。

```toml
[tools.generated]
scopes = ["global"]
check = false
file = "~/.local/share/generated-completions/_generated"
```

## 自定义输出目录

配置文件可以通过 `[settings] output_dir` 设置默认输出目录。

```toml
[settings]
output_dir = ".completions/zsh"
```

命令行参数 `--output` 和环境变量 `ZCS_OUTPUT_DIR` 的优先级更高，详细说明见[高级用法](./advanced-usage#自定义输出目录)。

## 相关文档

- [高级用法](./advanced-usage.md) 如果你想了解更多使用方式，包括项目级补全、单个工具生成、并发控制、自定义输出目录和自动更新原理，请查看高级用法文档。
