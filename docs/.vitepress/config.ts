import { defineConfig } from 'vitepress'

export default defineConfig({
  base: '/zsh-completions-sync/',
  lang: 'zh-CN',
  title: 'zcs',
  description: 'Synchronize zsh completion scripts into global and project-local completion directories.',

  themeConfig: {
    nav: [
      { text: 'Shell 补全', link: '/guide/completion' },
      { text: 'GitHub', link: 'https://github.com/YewFence/zsh-completions-sync' }
    ],

    sidebar: [
      {
        text: '指南',
        items: [
          { text: 'Shell 补全', link: '/guide/completion' }
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
