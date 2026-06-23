# zcs 命令补全

`zcs` 使用 Cobra 提供自身命令补全，可以输出适用于 zsh 的补全脚本。

```zsh
zcs completion zsh
```

这个命令只负责输出 `zcs` 自身的补全脚本，不会自动修改 `.zshrc`，也不会自动写入补全目录。

## 通过注册表管理

直接使用

```zsh
zcs generate
```

`zcs` 内置注册表已经包含了自己的补全脚本，所以会自动生成 `zcs` 的补全到全局目录。
