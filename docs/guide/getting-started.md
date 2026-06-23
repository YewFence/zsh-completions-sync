# 快速开始

`zcs` 会把多个命令行工具的 zsh 补全脚本同步到统一目录，再输出一段 zsh 初始化脚本来加载这些补全。

它默认区分全局补全和项目补全。稳定、全局安装的工具适合放在全局补全里，项目内版本变化频繁的工具适合放在项目补全里。

## 安装

### 使用 mise

如果只想在当前项目使用 `zcs`，运行下面的命令。

```bash
mise use github:YewFence/zsh-completions-sync/cmd/zcs
```

如果希望 `zcs` 全局可用，给 `mise use` 加上 `-g`。

```bash
mise use -g github:YewFence/zsh-completions-sync/cmd/zcs
```

### 使用 Go

```zsh
go install github.com/YewFence/zsh-completions-sync/cmd/zcs@latest
```

确认安装目录已经加入 `PATH` 后，可以运行下面的命令检查版本。

```zsh
zcs version
```

### 从源码构建

```bash
git clone https://github.com/YewFence/zsh-completions-sync.git
cd zsh-completions-sync
mise trust
mise install
mise run cli:install
```

源码安装会把命令安装到 `GOBIN`，如果当前 shell 找不到 `zcs`，需要把 `GOBIN` 加入 `PATH`。

```zsh
export PATH="$GOBIN:$PATH"
```

## 生成全局补全

先查看当前注册表里启用的全局工具。

```zsh
zcs list --scope global
```

再生成补全脚本。

```zsh
zcs generate
```

默认情况下，全局补全脚本会写入 `~/.zsh/completions`。

## 加载补全

把 `zcs init global` 输出的初始化脚本加入 `.zshrc`。

```zsh
echo 'eval "$(zcs init global)"' >> ~/.zshrc
source ~/.zshrc
```

重新打开 shell 后，zsh 会把 `~/.zsh/completions` 加入 `fpath` 并调用 `compinit`。

## 自动刷新全局补全

如果想在工具升级后自动刷新已有补全，可以把 `zcs check-update` 加入 `.zshrc`。

```zsh
echo 'eval "$(zcs check-update)"' >> ~/.zshrc
```

使用 `mise`、`asdf`、`Volta`、`Nix` 等命令管理器时，应该先完成命令管理器初始化，再运行 `zcs check-update` 输出的脚本。否则 zsh 可能只能看到长期不变的 shim 文件，底层工具升级后不会触发刷新。

## 下一步

如果你只想使用内置工具注册表，读完本页就可以开始使用。想了解全局补全和项目补全的差异，可以继续阅读[核心概念](./concepts)。想添加自己的工具，可以直接看[配置文件](./config-file)。
