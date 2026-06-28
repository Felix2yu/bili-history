import { defineEventHandler, proxyRequest } from 'h3'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const backendUrl = config.backendUrl || 'http://localhost:8899'

  const path = event.path.replace(/^\/api/, '')
  const target = backendUrl + path

  return proxyRequest(event, target, {
    headers: {
      host: new URL(backendUrl).host,
    },
  })
})
