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
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => {},
        requestAnimationFrame: (cb: Function) => setTimeout(cb, 0),
        cancelAnimationFrame: (id: number) => clearTimeout(id),
        innerWidth: 1024,
        innerHeight: 768,
        performance: { now: () => Date.now() },
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
          style: {},
          classList: { add: () => {}, remove: () => {}, toggle: () => {} },
        }),
        getElementById: () => null,
        createDocumentFragment: () => ({
          appendChild: () => {},
          append: () => {},
        }),
        createRange: () => ({
          selectNodeContents: () => {},
          deleteContents: () => {},
          extractContents: () => ({ childNodes: [] }),
          createContextualFragment: () => ({ nodeType: 11, childNodes: [] }),
        }),
        documentElement: {
          classList: {
            add: () => {},
            remove: () => {},
          },
          style: {},
        },
        head: {
          appendChild: () => {},
          removeChild: () => {},
        },
        body: {
          appendChild: () => {},
          removeChild: () => {},
        },
        addEventListener: () => {},
        removeEventListener: () => {},
        querySelector: () => null,
        querySelectorAll: () => [],
        createElementNS: () => ({
          setAttribute: () => {},
        }),
        createTextNode: () => ({}),
        createComment: () => ({}),
      } as any
    }

    if (!globalThis.navigator) {
      globalThis.navigator = {
        userAgent: 'SSR',
      } as Navigator
    }
  }
})
