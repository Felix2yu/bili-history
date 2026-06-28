import { defineStore } from 'pinia'

export const usePrivacyStore = defineStore('privacy', {
  state: () => ({
    isPrivacyMode: false,
  }),

  actions: {
    initPrivacyMode() {
      if (!process.client) return
      this.isPrivacyMode = localStorage.getItem('privacyMode') === 'true'
    },

    togglePrivacyMode() {
      this.isPrivacyMode = !this.isPrivacyMode
      if (process.client) {
        localStorage.setItem('privacyMode', this.isPrivacyMode.toString())
      }
    },

    setPrivacyMode(value) {
      this.isPrivacyMode = value
      if (process.client) {
        localStorage.setItem('privacyMode', value.toString())
      }
    },
  },
})
