# 贡献指南

项目使用 [mise](https://mise.jdx.dev/) 安装开发依赖，并通过 mise task 组织常用命令。

## 初始化环境

```bash
mise trust && mise install
```

## 代码检查

```bash
mise run check
```

这个命令会运行格式检查、静态检查、构建和 lint。

## 运行命令行工具

```bash
mise run cli
```

## 文档站开发

```bash
mise tasks | rg docs
```

常用任务包括安装文档依赖、启动开发服务器、构建文档站和预览构建结果。

## 更新 GitHub Action

```bash
mise run action:update
```

## 查看全部任务

```bash
mise tasks
```
