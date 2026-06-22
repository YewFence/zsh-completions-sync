## 高级用法

### 自定义 `autoload -Uz compinit && compinit` 加载时机

使用

```zsh
zcs init --no-compinit
```

可以关闭自动调用 `compinit` 的功能，此时需要你自行在 `.zshrc` 里调用 `autoload -Uz compinit && compinit`。

### 项目级别工具补全

使用 [mise enter hook](https://mise.jdx.dev/hooks.html#shell-hooks) 以实现进入目录自动动态加载补全脚本的功能。

示例 `mise.toml` 配置
```toml
[hooks.enter]
shell = "zsh"
script = 'eval "$(zcs init --project)"'
```

该命令会运行 `zcs project` 以生成项目级别的补全脚本到当前目录的 `.completions/zsh`，并在进入目录时自动加载。

### 单个工具补全脚本生成

`zcs` 支持针对单个工具生成补全脚本，命令格式是 `zcs global tool` 或 `zcs project tool`，其中 `tool` 是工具名或工具别名。比如：

```zsh
zcs global pnpm
zcs project node
```

### 并发

同步时每个工具会并发处理，单个工具失败只会输出 `warn` 并跳过，不会覆盖已有补全文件，其他工具的补全脚本生成仍然会继续。默认最多同时处理 8 个工具，可以用 `--jobs` 调整。

```sh
zcs global --jobs 4
zcs project --jobs 2
```

### 自定义补全脚本输出目录

输出目录可以通过命令行参数、环境变量或配置文件覆盖。优先级从高到低是 `--output`、`ZCS_OUTPUT_DIR`、`[settings] output_dir`、默认目录。

```sh
zcs global --output ~/.local/share/zsh/completions
ZCS_OUTPUT_DIR=.completions/custom zcs project
```

> 当你自定义里输出脚本目录后，需要在 `.zshrc` 里同步使用 `ZCS_GLOBAL_OUTPUT_DIR` 设置同一个目录，否则补全脚本加载 / 自动更新会失败。
