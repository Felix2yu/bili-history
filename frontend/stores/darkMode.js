import { defineStore } from 'pinia'

export const useDarkMode = defineStore('darkMode', {
  state: () => ({
    isDarkMode: false,
    darkMode: 'system', // 'system' | 'light' | 'dark'
    initialized: false,
    mediaQuery: null,
    mediaQueryHandler: null,
  }),

  actions: {
    async initDarkMode() {
      if (!process.client || this.initialized) return

      // 尝试从后端API获取配置
      try {
        const { getAppearanceConfig } = await import('~/utils/api')
        const response = await getAppearanceConfig()
        if (response.data && response.data.success) {
          this.darkMode = response.data.dark_mode || 'system'
        }
      } catch (error) {
        console.error('获取外观配置失败，使用默认值:', error)
        this.darkMode = 'system'
      }

      // 同步到 localStorage，供首屏脚本使用
      try { localStorage.setItem('darkMode', this.darkMode) } catch(e) {}

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

      // 同步到 localStorage
      try { localStorage.setItem('darkMode', mode) } catch(e) {}

      // 保存到后端API
      if (process.client) {
        try {
          const { updateAppearanceConfig } = await import('~/utils/api')
          await updateAppearanceConfig({ dark_mode: mode })
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
