import { defineStore } from 'pinia'

export const useDarkMode = defineStore('darkMode', {
  state: () => ({
    isDarkMode: false,
    initialized: false,
  }),

  actions: {
    initDarkMode() {
      if (!process.client) return
      const savedMode = localStorage.getItem('darkMode')
      if (savedMode !== null) {
        this.isDarkMode = savedMode === 'true'
      } else {
        this.isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
      }
      this.applyDarkMode()
      this.initialized = true
    },

    applyDarkMode() {
      if (!process.client) return
      if (this.isDarkMode) {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    },

    toggleDarkMode() {
      this.isDarkMode = !this.isDarkMode
      if (process.client) {
        localStorage.setItem('darkMode', this.isDarkMode.toString())
      }
      this.applyDarkMode()
    },
  },
})
