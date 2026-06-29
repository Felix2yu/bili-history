// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  modules: [
    '@pinia/nuxt',
    '@vueuse/nuxt',
    // '@vite-pwa/nuxt',
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
        { name: 'viewport', content: 'width=device-width, initial-scale=1.0, viewport-fit=cover, user-scalable=no' },
        { name: 'theme-color', content: '#1e293b' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-status-bar-style', content: 'black-translucent' },
        { name: 'apple-mobile-web-app-title', content: 'Bili历史' },
        { name: 'description', content: '获取和管理B站历史记录' },
        { name: 'format-detection', content: 'telephone=no' },
        { property: 'og:title', content: 'Bilibili 历史记录管理' },
        { property: 'og:description', content: '获取和管理B站历史记录' },
        { property: 'og:type', content: 'website' },
        { property: 'og:image', content: '/icons/icon-512x512.png' },
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/logo.ico' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/icons/apple-touch-icon.png' },
        { rel: 'apple-touch-icon', sizes: '152x152', href: '/icons/icon-192x192.png' },
        { rel: 'apple-touch-icon', href: '/icons/icon-192x192.png' },
        { rel: 'mask-icon', href: '/logo.svg', color: '#1e293b' },
      ],
    },
  },

  pwa: {
    registerType: 'autoUpdate',
    manifest: {
      id: '/',
      name: 'Bilibili 历史记录管理',
      short_name: 'Bili历史',
      description: '获取和管理B站历史记录',
      theme_color: '#1e293b',
      background_color: '#0f172a',
      display: 'standalone',
      orientation: 'portrait-primary',
      scope: '/',
      start_url: '/',
      categories: ['entertainment', 'utilities'],
      icons: [
        {
          src: '/icons/icon-192x192.png',
          sizes: '192x192',
          type: 'image/png',
        },
        {
          src: '/icons/icon-512x512.png',
          sizes: '512x512',
          type: 'image/png',
        },
        {
          src: '/icons/maskable-icon-512x512.png',
          sizes: '512x512',
          type: 'image/png',
          purpose: 'maskable',
        },
      ],
    },
    workbox: {
      navigateFallback: '/',
      navigateFallbackDenylist: [/^\/api\//, /^\/_nuxt\//],
      runtimeCaching: [
        {
          urlPattern: /^\/api\/.*/i,
          handler: 'NetworkOnly',
          options: {
            cacheName: 'api-cache',
          },
        },
        {
          urlPattern: /\.(png|jpg|jpeg|svg|gif|webp|ico)$/i,
          handler: 'CacheFirst',
          options: {
            cacheName: 'image-cache',
            expiration: {
              maxEntries: 100,
              maxAgeSeconds: 60 * 60 * 24 * 30,
            },
          },
        },
        {
          urlPattern: /\.(css|js)$/i,
          handler: 'StaleWhileRevalidate',
          options: {
            cacheName: 'static-cache',
            expiration: {
              maxEntries: 60,
              maxAgeSeconds: 60 * 60 * 24 * 7,
            },
          },
        },
        {
          urlPattern: /^https:\/\/i[0-9]\.hdslb\.com\/.*/i,
          handler: 'CacheFirst',
          options: {
            cacheName: 'bilibili-image-cache',
            expiration: {
              maxEntries: 200,
              maxAgeSeconds: 60 * 60 * 24 * 7,
            },
          },
        },
      ],
    },
    client: {
      installPrompt: true,
    },
    devOptions: {
      enabled: false,
    },
  },

  vite: {
    define: {
      __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false',
    },
  },
})
