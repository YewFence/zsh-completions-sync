## 配置文件

`zcs` 通过配置文件注册表以确认需要生成哪些工具的补全脚本，以及如何生成。配置文件注册表有高度自定义能力，可以将其他命令行工具的补全脚本整合进来。

配置文件注册表分三个级别，优先级从低到高是内置注册表、用户配置、项目配置。

用户配置优先读取 `~/.config/zsh-completions-sync/registry.toml`，也兼容读取 `~/.config/zsh-completions-sync-registry.toml`。如果两个文件同时存在，只读取前者并输出告警。设置了 `XDG_CONFIG_HOME` 时，用户配置路径会跟随 `XDG_CONFIG_HOME`。

项目配置优先读取当前项目的 `.config/zsh-completions-sync.toml`，也兼容读取 `.zsh-completions-sync.toml`。如果两个文件同时存在，只读取前者并输出告警。

### 配置文件注册表示例

```toml
# 可选的：自定义输出目录
# [settings]
# output_dir = ".completions/zsh"
# 工具名称，建议设定为可执行文件名称
[tools.mise]
# 补全脚本生成范围，global 表示全局生效，project 表示仅在项目里生效
scopes = ["global"]
# 补全脚本生成命令，默认读取该命令的 stdout 作为补全脚本内容
command = ["mise", "completion", "zsh"]

[tools.pnpm]
# 显式关闭工具的补全脚本生成功能
disabled = true

[tools.local-tool]
scopes = ["project"]
# 直接从本地文件读取补全脚本
file = "$PWD/completions/_local-tool"

[tools.installing-tool]
scopes = ["project"]
# 自定义工具可用性检查命令，默认会检查 PATH 里是否有同名可执行文件
check = "installing-tool --version"
# 补全脚本预处理命令，适合那些只能先把补全写到文件的工具。它会在读取补全来源前运行，输出不会被当作补全脚本内容，而是被丢弃掉。
pre-command = ["installing-tool", "completion", "install", "--shell", "zsh"]
# 在 `pre-command` 运行完成后从指定文件读取补全脚本
file = "/usr/share/zsh/vendor-completions/_installing-tool"

[tools.remote-tool]
scopes = ["global"]
# 直接从远程 URL 获取补全脚本
file = "https://example.com/completions/_remote-tool"

[tools.git-tool]
scopes = ["global", "project"]
# 从 Git 仓库里获取补全脚本
file = "git+https://github.com/example/tool.git//completions/_tool?ref=v1.2.3"

[tools.git-tool-alt]
scopes = ["project"]
# 从 Git 仓库里获取补全脚本，Git 表格形式
file = { git = "https://github.com/example/tool.git", path = "completions/_tool", ref = "v1.2.3" }
```
