import withNuxt from './.nuxt/eslint.config.mjs';

// .nuxt/eslint.config.mjs 由 @nuxt/eslint 模块在 nuxt prepare 时生成，
// 已带 Nuxt 自动导入 globals 与 vue/ts 推荐规则；这里只做增量定制。
// .nuxt / .output / dist / node_modules 已被 .gitignore 覆盖
// （withNuxt 内置 eslint-config-flat-gitignore），显式再声明一遍，
// 防止在非 git 上下文里跑 eslint 时把生成产物当源码检查。
export default withNuxt(
  {
    ignores: ['.output/**', '.nuxt/**', 'dist/**', 'node_modules/**']
  },
  {
    // analytics 各页的洞察文本统一走 utils/escapeHtml 的 escapeHtml 转义后
    // 再包 <span>（formatInsightText / OverallCompletionPage 的分段），
    // v-html 内容不含任何未转义输入；该目录外如仍需裸 v-html，规则保持告警。
    files: ['components/analytics/**/*.vue'],
    rules: {
      'vue/no-v-html': 'off'
    }
  },
  {
    // 项目命名风格：Navbar / Pagination / Settings / Sidebar 等顶层组件
    // 使用单词名（Nuxt 官方示例同样如此），强制 multi-word 只会带来
    // 大面积无收益改名；layouts/pages/app 已被 Nuxt 配置自动豁免。
    rules: {
      'vue/multi-word-component-names': 'off',
      // 与 prettier 的 Vue 输出对齐：void 元素（img/input/br）同样自闭合
      // （<img />）。该规则默认 void: 'never'，与 prettier 互改对方产出，
      // 两者永远无法同时满足；对齐后 eslint --fix 与 prettier --write 幂等共存。
      'vue/html-self-closing': [
        'warn',
        {
          html: { normal: 'always', void: 'always', component: 'always' },
          svg: 'always',
          math: 'always'
        }
      ]
    }
  }
);
