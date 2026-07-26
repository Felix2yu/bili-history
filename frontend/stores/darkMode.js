import { defineStore } from 'pinia'

export const THEME_PRESETS = [
  { id: 'sakura', name: '樱花粉', color: '#fb7299' },
  { id: 'ocean', name: '海洋蓝', color: '#3b82f6' },
  { id: 'forest', name: '森林绿', color: '#10b981' },
  { id: 'amber', name: '琥珀橙', color: '#f59e0b' },
  { id: 'lavender', name: '薰衣草', color: '#8b5cf6' },
  { id: 'mint', name: '薄荷青', color: '#14b8a6' },
  { id: 'midnight', name: '暗夜金', color: '#eab308' },
  { id: 'crimson', name: '烈焰红', color: '#ef4444' },
]

export const useDarkMode = defineStore('darkMode', {
  state: () => ({
    isDarkMode: false,
    darkMode: 'system',
    themeColor: 'sakura',
    initialized: false,
    mediaQuery: null,
    mediaQueryHandler: null,
  }),

  getters: {
    currentTheme: (state) => {
      return THEME_PRESETS.find(t => t.id === state.themeColor) || THEME_PRESETS[0]
    },
  },

  actions: {
    async initDarkMode() {
      if (!process.client || this.initialized) return

      try {
        const { getAppearanceConfig } = await import('~/utils/api')
        const response = await getAppearanceConfig()
        if (response.data && response.data.success) {
          this.darkMode = response.data.dark_mode || 'system'
          this.themeColor = response.data.theme_color || 'sakura'
        }
      } catch (error) {
        console.error('获取外观配置失败，使用默认值:', error)
        this.darkMode = 'system'
        this.themeColor = 'sakura'
      }

      try {
        const savedTheme = localStorage.getItem('themeColor')
        if (savedTheme && THEME_PRESETS.some(t => t.id === savedTheme)) {
          this.themeColor = savedTheme
        }
        localStorage.setItem('darkMode', this.darkMode)
        localStorage.setItem('themeColor', this.themeColor)
      } catch(e) {}

      this.applyThemeColor()
      this.applyDarkMode()
      this.setupSystemListener()
      this.initialized = true
    },

    setupSystemListener() {
      if (!process.client) return

      this.mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      this.mediaQueryHandler = () => {
        if (this.darkMode === 'system') {
          this.applyDarkMode()
        }
      }
      this.mediaQuery.addEventListener('change', this.mediaQueryHandler)
    },

    applyThemeColor() {
      if (!process.client) return
      document.documentElement.setAttribute('data-theme', this.themeColor)
    },

    applyDarkMode() {
      if (!process.client) return

      let shouldBeDark = false
      if (this.darkMode === 'system') {
        shouldBeDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      } else {
        shouldBeDark = this.darkMode === 'dark'
      }

      this.isDarkMode = shouldBeDark
      if (shouldBeDark) {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    },

    async setDarkMode(mode) {
      if (!['system', 'light', 'dark'].includes(mode)) return

      this.darkMode = mode
      this.applyDarkMode()

      try { localStorage.setItem('darkMode', mode) } catch(e) {}

      if (process.client) {
        try {
          const { updateAppearanceConfig } = await import('~/utils/api')
          await updateAppearanceConfig({ dark_mode: mode, theme_color: this.themeColor })
        } catch (error) {
          console.error('保存外观配置失败:', error)
        }
      }
    },

    async setThemeColor(themeId) {
      if (!THEME_PRESETS.some(t => t.id === themeId)) return

      this.themeColor = themeId
      this.applyThemeColor()

      try { localStorage.setItem('themeColor', themeId) } catch(e) {}

      if (process.client) {
        try {
          const { updateAppearanceConfig } = await import('~/utils/api')
          await updateAppearanceConfig({ dark_mode: this.darkMode, theme_color: themeId })
        } catch (error) {
          console.error('保存外观配置失败:', error)
        }
      }
    },

    toggleDarkMode() {
      const modes = ['light', 'dark']
      const currentIndex = modes.indexOf(this.darkMode === 'system'
        ? (this.isDarkMode ? 'dark' : 'light')
        : this.darkMode)
      const nextMode = modes[(currentIndex + 1) % modes.length]
      this.setDarkMode(nextMode)
    },
  },
})
