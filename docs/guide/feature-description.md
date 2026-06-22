## 功能说明

### 补全脚本生成

`zcs` 会根据[注册表配置文件](#配置文件)里每个工具的配置，运行对应命令或读取对应文件来生成补全脚本到 `~/.zsh/completions`，如果注册表声明了工具但是工具不存在会静默跳过，如果工具存在但是补全生成失败会输出警告，不会覆盖已有补全文件。

### Zsh 加载脚本

`zcs init` 会输出 zsh 脚本，源码位于 [internal/cli/init-snippets](internal/cli/init-snippets)，包含数个部分，可以根据需要自行开关：

| 片段文件 | 功能 | 启用/关闭方式 |
| --- | --- | --- |
| [全局补全目录增加](internal/cli/init-snippets/global.zsh) | 把全局补全目录幂等加入 `fpath` | 始终启用 |
| [补全加载](internal/cli/init-snippets/compinit.zsh)：调用 `compinit` 加载补全脚本。| 默认启用，使用 `--no-compinit` 以关闭 |
| [项目补全加载](internal/cli/init-snippets/project.zsh)：把当前项目的补全目录加入 `fpath` 并调用 `compinit`。| 使用 `--project` 以打开该功能 |
| [项目补全刷新](internal/cli/init-snippets/project-sync.zsh)：生成项目级别的补全脚本。| 使用 `--no-sync` 以关闭 |
| [全局补全刷新](internal/cli/init-snippets/global-sync.zsh)：在 zsh 启动时检查全局补全是否需要更新。| 使用 `--global-sync` 以打开该功能 |

### 自动更新原理

`zcs init --global-sync` 会输出一段脚本，遍历全局补全目录里的 `_tool` 文件，用文件名反推出工具名，再比较对应同名的可执行文件和补全文件的修改时间；只要有一个可执行文件更新于已有补全文件，就会静默运行一次 `zcs global`。它不会读取注册表，也不会发现从未生成过补全的新工具，新工具仍然需要手动运行 `zcs global` 或 `zcs global tool`。

这个设计优先保证平时的 shell 启动速度，而不是追求更新判断的绝对准确性，所以它是一个差不多够用但比较脆弱的启发式检查。刷新判断刻意写成 zsh 脚本，而不是每次启动都进入 Go 程序加载注册表，也不会做工具版本检测/缓存/追踪；所以在发现补全脚本异常时，可以手动运行 `zcs global` 以修复。

> 使用 `mise`、`asdf`、`Volta`、`Nix` 这类命令管理器时要注意初始化顺序。自动刷新片段依赖 zsh 的 `${commands[tool]}` 找到当前可执行文件，如果它在 `mise activate` 之前运行，看到的可能是长期不变的 shim 修改时间，底层工具升级后也不会触发刷新，所以应先激活命令管理器，再运行 `zcs init --global-sync` 输出的片段。

项目级补全目录应该放在全局补全目录之前，这样当前项目的新 shell 会优先使用项目内工具版本生成的补全。
