# zsh-completions-sync

`zcs` 是一个 zsh 补全脚本管理 CLI，通过简单易懂的 `toml` 配置文件管理多个命令行工具的的 zsh 补全脚本。

> `zcs` 正在早期开发中，不保证向后兼容性。功能可能尚不完备，欢迎尝试和反馈。

## 快速开始

### 安装

#### [Mise](https://mise.jdx.dev/)

```bash
# 仅在当前目录生效，如果需要安装到全局，请加上 -g 参数
mise use github:YewFence/zsh-completions-sync/cmd/zcs
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
# 安装到 GOBIN，需要手动把 GOBIN 加入 PATH
export PATH="$GOBIN:$PATH"
```

### 使用

1. 生成补全脚本

```bash
zcs list --global
# 查看支持的工具
zcs global
# 生成全局生效的补全脚本
```

补全脚本会生成到 `~/.zsh/completions`。

2. 配置 zsh 以加载补全脚本

```bash
echo 'eval "$(zcs init)"' >> ~/.zshrc
source ~/.zshrc
```

Done! 享受愉快的自动补全吧~

3. 可选的：配置自动更新

```zsh
echo 'eval "$(zcs init --global-sync)"' >> ~/.zshrc
```

## 文档

更多说明，请查看[文档站](https://yewfence.github.io/zsh-completions-sync)。或者[文档源码](./docs)

## 许可证

[MIT License](LICENSE)
