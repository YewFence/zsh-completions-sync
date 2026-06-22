# zsh-completions-sync

`zcs` 是一个小型 zsh 补全脚本同步工具，用来把全局工具补全和项目工具补全分开放置，避免所有补全脚本都混在 `~/.zsh/completions` 里。

## 安装

当前项目是 Go CLI，可以从源码构建。

```sh
mise run build
./bin/zcs version
```

也可以直接运行。

```sh
go run . list
```

## 使用

全局补全会生成到 `~/.zsh/completions`。

```sh
zcs global
```

项目补全会生成到当前项目的 `.completions/zsh`。

```sh
zcs project
```

查看合并后的注册表。

```sh
zcs list
zcs list --scope global
zcs list --scope project
```

同步时每个工具会并发处理，单个工具失败只会输出 `warn` 并跳过，不会覆盖已有补全文件。

## Zsh 加载示例

项目级补全目录应该放在全局补全目录之前，这样当前项目的新 shell 会优先使用项目内工具版本生成的补全。

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

## 配置层级

注册表分三个级别，优先级从低到高是内置注册表、用户配置、项目配置。

用户配置优先读取 `~/.config/zsh-completions-sync/registry.toml`，也兼容读取 `~/.config/zsh-completions-sync-registry.toml`。如果两个文件同时存在，只读取前者并输出告警。设置了 `XDG_CONFIG_HOME` 时，用户配置路径会跟随 `XDG_CONFIG_HOME`。

项目配置优先读取当前项目的 `.config/zsh-completions-sync.toml`，也兼容读取 `.zsh-completions-sync.toml`。如果两个文件同时存在，只读取前者并输出告警。

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

[tools.git-tool-alt]
scopes = ["project"]
file = { git = "https://github.com/example/tool.git", path = "completions/_tool", ref = "v1.2.3" }
```

`scopes` 决定工具在哪些命令里生效。`command` 表示运行命令并从标准输出读取补全脚本，`file` 表示直接读取补全脚本，支持本地路径、`file://`、HTTP、HTTPS、`git+仓库//路径?ref=版本` 和 Git 表格形式。如果同时配置了 `file` 和 `command`，会优先使用 `file`。

`pre-command` 会在读取补全来源前运行，适合那些只能先把补全写到文件的工具。`check` 用来判断工具是否可用，没有配置时默认检查工具名本身，可以配置成字符串、命令数组，或者配置为 `false` 关闭检查。

## 开发

```sh
mise run test
mise run build
mise run docs:dev
```

## 许可证

[MIT License](LICENSE)
