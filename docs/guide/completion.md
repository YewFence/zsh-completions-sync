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

## Zsh 加载

把项目补全目录放在全局补全目录前面，当前项目的新 shell 就会优先使用项目内补全。

```zsh
zcs project

fpath=(
  "$PWD/.completions/zsh"
  "$HOME/.zsh/completions"
  $fpath
)

autoload -Uz compinit
compinit
```

## 注册表位置

用户配置优先读取 `~/.config/zsh-completions-sync/registry.toml`，也兼容读取 `~/.config/zsh-completions-sync-registry.toml`。设置了 `XDG_CONFIG_HOME` 时，用户配置路径会跟随它。

项目配置优先读取当前项目的 `.config/zsh-completions-sync.toml`，也兼容读取 `.zsh-completions-sync.toml`。

## 注册表示例

```toml
[tools.mise]
scopes = ["global"]
command = ["mise", "completion", "zsh"]

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

同步过程会并发处理注册表里的工具。工具不存在会静默跳过，命令失败、文件读取失败或写入失败会输出 `warn`，并保留已有补全文件。
