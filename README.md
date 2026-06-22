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
go run ./cmd/zcs list
```

## 使用

全局补全会生成到 `~/.zsh/completions`。

```sh
zcs global
```

也可以只同步指定工具。

```sh
zcs global cargo
zcs project pnpm uv
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

同步时每个工具会并发处理，单个工具失败只会输出 `warn` 并跳过，不会覆盖已有补全文件。默认最多同时处理 8 个工具，可以用 `--jobs` 调整。

```sh
zcs global --jobs 4
zcs project --jobs 2
```

输出目录可以通过命令行参数、环境变量或配置文件覆盖。优先级从高到低是 `--output`、`ZCS_OUTPUT_DIR`、`[settings] output_dir`、默认目录。

```sh
zcs global --output ~/.local/share/zsh/completions
ZCS_OUTPUT_DIR=.completions/custom zcs project
```

## Zsh 加载示例

`zcs init` 会输出可以按需加入 `.zshrc` 的脚本片段。默认只加入全局补全目录，并包含 `compinit`。

```zsh
zcs init
```

如果希望 zsh 启动时自动刷新已经生成过的全局补全，可以使用 `--global-sync`。这段脚本会遍历全局补全目录里的 `_tool` 文件，用文件名反推出工具名，再比较对应可执行文件和补全文件的修改时间；只要有一个可执行文件更新于已有补全文件，就会静默运行一次 `zcs global`。它不会读取注册表，也不会发现从未生成过补全的新工具，新工具仍然需要手动运行 `zcs global` 或 `zcs global tool`。

```zsh
zcs init --global-sync
```

这个设计优先保证平时的 shell 启动速度，而不是追求更新判断的绝对准确性，所以它是一个差不多够用但比较脆弱的启发式检查。刷新判断刻意写成 zsh 脚本，而不是每次启动都进入 Go 程序加载注册表，也不会做版本检测或依赖追踪；如果你需要更可靠的更新，仍然推荐自己定时手动运行 `zcs global`。

如果你用自定义全局输出目录，需要在 `.zshrc` 里同步设置同一个目录，否则纯 zsh 加载和刷新片段看不到配置文件里的 `[settings] output_dir`。建议使用只影响全局初始化片段的 `ZCS_GLOBAL_OUTPUT_DIR`，也可以继续使用 `ZCS_OUTPUT_DIR`。

```zsh
export ZCS_GLOBAL_OUTPUT_DIR=$HOME/.local/share/zsh/completions
eval "$(zcs init --global-sync)"
```

使用 `mise`、`asdf`、`Volta`、`Nix` 这类命令管理器时要注意初始化顺序。自动刷新片段依赖 zsh 的 `${commands[tool]}` 找到当前可执行文件，如果它在 `mise activate` 之前运行，看到的可能是长期不变的 shim 修改时间，底层工具升级后也不会触发刷新，所以应先激活命令管理器，再运行 `zcs init --global-sync` 输出的片段。

如果希望同时加载项目级补全，可以使用 `--project`。这个模式会在脚本片段前面调用 `zcs project`，因为项目补全需要先生成再加入 `fpath`。

```zsh
zcs init --project
```

如果你已经用别的机制刷新项目补全，或者想自己管理 `compinit`，可以关掉对应输出。

```zsh
zcs init --project --no-sync
zcs init --no-compinit
```

项目级补全目录应该放在全局补全目录之前，这样当前项目的新 shell 会优先使用项目内工具版本生成的补全。

## 配置层级

注册表分三个级别，优先级从低到高是内置注册表、用户配置、项目配置。

用户配置优先读取 `~/.config/zsh-completions-sync/registry.toml`，也兼容读取 `~/.config/zsh-completions-sync-registry.toml`。如果两个文件同时存在，只读取前者并输出告警。设置了 `XDG_CONFIG_HOME` 时，用户配置路径会跟随 `XDG_CONFIG_HOME`。

项目配置优先读取当前项目的 `.config/zsh-completions-sync.toml`，也兼容读取 `.zsh-completions-sync.toml`。如果两个文件同时存在，只读取前者并输出告警。

## 注册表示例

```toml
[tools.mise]
scopes = ["global"]
command = ["mise", "completion", "zsh"]

[tools.pnpm]
disabled = true

[settings]
output_dir = ".completions/zsh"

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

`disabled = true` 可以禁用内置工具或上层配置里的工具。比如在用户配置或项目配置里只写 `[tools.pnpm] disabled = true`，就可以跳过内置的 `pnpm` 配置。`[settings] output_dir` 可以配置默认输出目录，但仍会被 `ZCS_OUTPUT_DIR` 和 `--output` 覆盖。

## 开发

```sh
mise run test
mise run build
mise run docs:dev
```

## 许可证

[MIT License](LICENSE)
