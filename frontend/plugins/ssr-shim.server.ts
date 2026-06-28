export default defineNuxtPlugin(() => {
  if (import.meta.server) {
    if (!globalThis.localStorage) {
      globalThis.localStorage = {
        getItem: () => null,
        setItem: () => {},
        removeItem: () => {},
        clear: () => {},
        key: () => null,
        length: 0,
      } as Storage
    }

    if (!globalThis.window) {
      globalThis.window = {
        location: {
          reload: () => {},
          href: '',
          hostname: 'localhost',
          protocol: 'http:',
          port: '3000',
        },
        matchMedia: () => ({ matches: false }),
        dispatchEvent: () => {},
        URL: {
          createObjectURL: () => '',
          revokeObjectURL: () => {},
        },
        __TAURI_INVOKE__: undefined,
        __TAURI__: undefined,
      } as any
    }

    if (!globalThis.document) {
      globalThis.document = {
        createElement: () => ({
          click: () => {},
          setAttribute: () => {},
          appendChild: () => {},
          removeChild: () => {},
          href: '',
        }),
        documentElement: {
          classList: {
            add: () => {},
            remove: () => {},
          },
        },
        body: {
          appendChild: () => {},
          removeChild: () => {},
        },
      } as any
    }

    if (!globalThis.navigator) {
      globalThis.navigator = {
        userAgent: 'SSR',
      } as Navigator
    }
  }
})
