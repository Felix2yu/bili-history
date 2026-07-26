import { ref, computed } from 'vue'

const DEFAULT_FALLBACK_URL = '/api'

export const useApiBase = () => {
  const config = useRuntimeConfig()

  const getBaseUrl = () => {
    const stored = localStorage.getItem('baseUrl')
    if (stored) return stored
    return config.public.defaultBackendUrl || DEFAULT_FALLBACK_URL
  }

  const baseUrl = ref(getBaseUrl())

  const serverUrls = computed(() => {
    const urls = [
      '/api',
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
    localStorage.setItem('baseUrl', url)
    baseUrl.value = url
    window.location.reload()
  }

  const resetBaseUrl = () => {
    localStorage.removeItem('baseUrl')
    baseUrl.value = config.public.defaultBackendUrl || DEFAULT_FALLBACK_URL
    window.location.reload()
  }

  return {
    baseUrl,
    serverUrls,
    setBaseUrl,
    resetBaseUrl,
  }
}
