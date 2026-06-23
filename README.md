# zsh-completions-sync

`zcs` 是一个 zsh 补全脚本管理 CLI。很多命令行工具都能生成 zsh 补全脚本，但生成方式、输出位置和更新时机并不统一。`zcs` 用一份 `TOML` 注册表描述这些工具的补全来源，通过直观的 `zcs generate` 命令统一生成补全脚本，并提供自动更新功能。它也可以结合一些 hook 能力管理项目级别特有文件的补全脚本。

> `zcs` 正在早期开发中，不保证向后兼容性。功能可能尚不完备，欢迎尝试和反馈。

## 快速开始

### 安装

#### [Mise](https://mise.jdx.dev/)

```bash
# Github backend
mise use --global github:YewFence/zsh-completions-sync
# Go backend
mise use --global go:github.com/YewFence/zsh-completions-sync/cmd/zcs
```

#### Go

```zsh
go install github.com/YewFence/zsh-completions-sync/cmd/zcs@latest
```

#### 从源码构建

```bash
git clone https://github.com/YewFence/zsh-completions-sync.git
cd zsh-completions-sync
mise trust
mise install
mise run cli:install
```

源码安装会把命令安装到 `GOBIN`，如果当前 shell 找不到 `zcs`，需要把 `GOBIN` 加入 `PATH`。

### 使用

1. 生成补全脚本

```bash
# 查看支持的全局工具
zcs list --scope global

# 生成全局补全脚本
zcs generate
```

补全脚本会生成到 `~/.zsh/completions`。

2. 配置 zsh 以加载补全脚本

```bash
echo 'eval "$(zcs init global)"' >> ~/.zshrc
source ~/.zshrc
```

3. 可选配置自动更新

```zsh
echo 'eval "$(zcs check-update)"' >> ~/.zshrc
```

如果使用 `mise` 等工具管理器，请把自动更新脚本放在命令管理器初始化之后。自动更新的原理详见[高级用法](./docs/advanced-usage.md)文档中的说明。

## 常用命令

完整命令说明请查看 Cobra 自动生成的[命令参考](./docs/reference/zcs.md)。

| 命令 | 说明 |
| --- | --- |
| `zcs list --scope global` | 查看全局补全工具 |
| `zcs list --scope project` | 查看项目补全工具 |
| `zcs generate` | 生成全局补全 |
| `zcs generate --scope project` | 生成项目补全 |
| `zcs generate pnpm` | 只生成指定工具的全局补全 |
| `zcs init global` | 输出全局补全加载脚本 |
| `zcs init project` | 输出项目补全加载脚本 |
| `zcs check-update` | 输出全局补全自动刷新脚本 |

## 文档

更多说明请查看[文档站](https://yewfence.github.io/zsh-completions-sync)，完整命令说明见[命令参考](https://yewfence.github.io/zsh-completions-sync/reference/zcs)，文档源码位于 [docs](./docs)。

## 开发

```bash
mise trust
mise install
mise run check
```

## 许可证

[MIT License](LICENSE)
