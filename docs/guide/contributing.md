## 开发

直接依赖： [mise](https://mise.jdx.dev/)

该项目通过 mise 安装开发依赖，使用 mise task 组织常用命令

### 初始化环境

```bash
mise trust && mise install
```

### 代码标准

```bash
mise run check
```

检查通过

### 运行 CLI 以测试功能

```bash
mise run cli
```

### 文档站开发

```bash
mise tasks | rg docs
```

### 更新 Github Action

```bash
mise run action:update
```

## 更多开发功能

```bash
mise tasks
```
