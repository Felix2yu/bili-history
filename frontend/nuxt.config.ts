// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  modules: [
    '@pinia/nuxt',
    '@vueuse/nuxt',
  ],

  css: [
    '~/assets/css/main.css',
    'animate.css',
    'vant/lib/index.css',
  ],

  postcss: {
    plugins: {
      tailwindcss: {},
      autoprefixer: {},
    },
  },

  runtimeConfig: {
    public: {
      defaultBackendUrl: process.env.NUXT_PUBLIC_DEFAULT_BACKEND_URL || '/api',
    },
    backendUrl: process.env.NUXT_BACKEND_URL || 'http://localhost:8899',
  },

  nitro: {
    devProxy: {
      '/api': {
        target: process.env.NUXT_BACKEND_URL || 'http://localhost:8899',
        changeOrigin: true,
        pathRewrite: { '^/api': '' },
      },
    },
  },

  app: {
    head: {
      title: 'Bilibili 历史记录管理',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1.0' },
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/logo.ico' },
      ],
    },
  },

  vite: {
    define: {
      __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false',
    },
  },
})
