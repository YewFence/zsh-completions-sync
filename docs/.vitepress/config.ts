import { defineConfig } from 'vitepress'

export default defineConfig({
  base: '/zsh-completions-sync/',
  lang: 'zh-CN',
  title: 'zcs',
  description: 'Synchronize zsh completion scripts into global and project-local completion directories.',

  themeConfig: {
    nav: [
      { text: '快速开始', link: '/guide/getting-started' },
      { text: '配置文件', link: '/guide/config-file' },
      { text: '命令参考', link: '/reference/zcs' },
      { text: 'GitHub', link: 'https://github.com/YewFence/zsh-completions-sync' }
    ],

    sidebar: [
      {
        text: '指南',
        items: [
          { text: '快速开始', link: '/guide/getting-started' },
          { text: '核心概念', link: '/guide/concepts' },
          { text: '配置文件', link: '/guide/config-file' },
          { text: '高级用法', link: '/guide/advanced-usage' },
          { text: 'zcs 命令补全', link: '/guide/zcs-completion' }
        ]
      },
      {
        text: '参考',
        items: [
          { text: 'zcs', link: '/reference/zcs' },
          { text: 'check-update', link: '/reference/zcs_check-update' },
          { text: 'generate', link: '/reference/zcs_generate' },
          { text: 'init', link: '/reference/zcs_init' },
          { text: 'init global', link: '/reference/zcs_init_global' },
          { text: 'init project', link: '/reference/zcs_init_project' },
          { text: 'list', link: '/reference/zcs_list' },
          { text: 'version', link: '/reference/zcs_version' }
        ]
      },
      {
        text: '开发',
        items: [
          { text: '贡献指南', link: '/guide/contributing' }
        ]
      }
    ],

    search: {
      provider: 'local'
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/YewFence/zsh-completions-sync' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © YewFence'
    },

    docFooter: {
      prev: '上一页',
      next: '下一页'
    },

    outline: {
      label: '本页目录'
    },

    lastUpdated: {
      text: '最后更新'
    }
  }
})
