import { ref, computed } from 'vue'

const DEFAULT_FALLBACK_URL = 'http://localhost:8899'

export const useApiBase = () => {
  const config = useRuntimeConfig()

  const getBaseUrl = () => {
    if (import.meta.server) {
      return config.backendUrl || config.public.defaultBackendUrl || DEFAULT_FALLBACK_URL
    }
    if (process.client) {
      const stored = localStorage.getItem('baseUrl')
      if (stored) return stored
    }
    return config.public.defaultBackendUrl || DEFAULT_FALLBACK_URL
  }

  const baseUrl = ref(getBaseUrl())

  const serverUrls = computed(() => {
    const urls = [
      'http://127.0.0.1:8899',
      'http://localhost:8899',
      'http://0.0.0.0:8899',
    ]
    const defaultUrl = config.public.defaultBackendUrl
    if (defaultUrl && !urls.includes(defaultUrl)) {
      urls.unshift(defaultUrl)
    }
    return urls
  })

  const setBaseUrl = (url: string) => {
    if (process.client) {
      localStorage.setItem('baseUrl', url)
    }
    baseUrl.value = url
    if (process.client) {
      window.location.reload()
    }
  }

  const resetBaseUrl = () => {
    if (process.client) {
      localStorage.removeItem('baseUrl')
    }
    baseUrl.value = config.public.defaultBackendUrl || DEFAULT_FALLBACK_URL
    if (process.client) {
      window.location.reload()
    }
  }

  return {
    baseUrl,
    serverUrls,
    setBaseUrl,
    resetBaseUrl,
  }
}

export const useApiFetch = (request: any, opts: any = {}) => {
  const { baseUrl } = useApiBase()
  const config = useRuntimeConfig()

  const apiBase = import.meta.server
    ? (config.backendUrl || config.public.defaultBackendUrl)
    : baseUrl.value

  return useFetch(request, {
    baseURL: apiBase,
    ...opts,
  })
}

export const useApi$fetch = () => {
  const { baseUrl } = useApiBase()
  const config = useRuntimeConfig()

  const apiBase = import.meta.server
    ? (config.backendUrl || config.public.defaultBackendUrl)
    : baseUrl.value

  return $fetch.create({
    baseURL: apiBase,
  })
}
