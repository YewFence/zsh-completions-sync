---
layout: home

hero:
  name: 'zcs'
  text: 'zsh-completions-sync'
  tagline: '用一份注册表同步全局和项目级 zsh 补全脚本。'
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/getting-started
    - theme: alt
      text: 命令参考
      link: /reference/zcs
    - theme: alt
      text: GitHub
      link: https://github.com/YewFence/zsh-completions-sync

features:
  - title: 全局补全
    details: 运行 zcs generate，把稳定工具的补全写入 ~/.zsh/completions。
  - title: 项目补全
    details: 运行 zcs generate --scope project，把项目内工具补全写入 .completions/zsh。
  - title: 注册表配置
    details: 使用 TOML 描述命令、文件、远程 URL 或 Git 仓库中的补全来源。
---
