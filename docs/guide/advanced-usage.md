# 高级用法

本页收集更细的使用方式，包括项目级补全、单个工具生成、并发控制、自定义输出目录和自动更新原理。

## 自定义 compinit 加载时机

默认情况下，`zcs init global` 和 `zcs init project` 输出的脚本会调用 `autoload -Uz compinit && compinit`。

如果你已经在 `.zshrc` 中统一管理 `compinit`，可以关闭这段输出。

```zsh
zcs init global --no-compinit
zcs init project --no-compinit
```

关闭后需要自行确保 zsh 在合适的位置调用 `compinit`。

## 项目级补全

项目级补全可以配合 [mise enter hook](https://mise.jdx.dev/hooks.html#shell-hooks) 使用，让 shell 进入项目目录时自动生成并加载项目补全。

示例 `mise.toml` 配置如下。

```toml
[hooks.enter]
shell = "zsh"
script = 'eval "$(zcs init project)"'
```

这段脚本会运行 `zcs generate --scope project`，把项目级补全写入当前目录的 `.completions/zsh`，并在进入目录时加载它。

如果你只想加载项目补全目录，不想在进入目录时自动生成，可以加上 `--no-sync`。

```toml
[hooks.enter]
shell = "zsh"
script = 'eval "$(zcs init project --no-sync)"'
```

## 生成单个工具补全

`zcs generate` 可以只处理指定工具。工具名需要存在于当前注册表，并且启用于当前 scope。

```zsh
zcs generate pnpm
zcs generate --scope project uv
```

如果指定的工具不存在，命令会报错并指出缺失的工具名。

## 控制并发数

同步时每个工具会并发处理。默认最多同时处理 8 个工具，可以用 `--jobs` 调整。

```zsh
zcs generate --jobs 4
zcs generate -s project --jobs 2
```

单个工具失败只会输出 `warn` 并跳过，不会覆盖已有补全文件，其他工具的补全生成仍然会继续。

## 自定义输出目录

输出目录可以通过命令行参数、环境变量或配置文件覆盖。优先级从高到低依次是 `--output`、`ZCS_OUTPUT_DIR`、`[settings] output_dir`、默认目录。

```zsh
zcs generate --output ~/.local/share/zsh/completions
ZCS_OUTPUT_DIR=.completions/custom zcs generate --scope project
```

如果自定义了全局输出目录，需要在 `.zshrc` 中同步设置 `ZCS_GLOBAL_OUTPUT_DIR`，否则 `zcs init global` 和 `zcs check-update` 输出的加载脚本仍然会使用默认目录。

```zsh
export ZCS_GLOBAL_OUTPUT_DIR="$HOME/.local/share/zsh/completions"
eval "$(zcs init global)"
eval "$(zcs check-update)"
```

## 自动更新原理

`zcs check-update` 会输出一段 zsh 脚本。它会遍历全局补全目录里的 `_tool` 文件，用文件名反推出工具名，再比较对应同名可执行文件和补全文件的修改时间。

只要有一个可执行文件比已有补全文件更新，这段脚本就会静默运行一次 `zcs generate`。

这个机制不会读取注册表，也不会发现从未生成过补全的新工具。新工具仍然需要手动运行 `zcs generate` 或 `zcs generate tool`。

它是轻量级启发式检查，优先降低 shell 启动成本。因此在发现补全脚本异常时，可以手动运行下面的命令修复。

```zsh
zcs generate
```

使用 `mise`、`asdf`、`Volta`、`Nix` 这类命令管理器时要注意初始化顺序。自动刷新片段依赖 zsh 的 `${commands[tool]}` 找到当前可执行文件，如果它在命令管理器激活之前运行，看到的可能是长期不变的 shim 修改时间，底层工具升级后也不会触发刷新。

## 相关文档

- [核心概念](./concepts.md) 如果你想了解 `zcs` 的核心流程和相关概念，请查看核心概念文档。
- [配置文件](./config-file.md) 如果你想了解注册表的配置方式和更多字段说明，请查看配置文件文档。
