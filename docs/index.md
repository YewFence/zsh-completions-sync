---
layout: home

hero:
  name: 'zcs'
  text: 'zsh-completions-sync'
  tagline: '把全局和项目级 zsh 补全脚本分开同步。'
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/completion
    - theme: alt
      text: GitHub
      link: https://github.com/YewFence/zsh-completions-sync

features:
  - title: 全局补全
    details: 运行 zcs global，把稳定工具补全写入 ~/.zsh/completions。
  - title: 项目补全
    details: 运行 zcs project，把项目内工具补全写入 .completions/zsh。
  - title: 并发同步
    details: 多个工具并发生成，失败项只输出告警，不覆盖已有补全。
---
