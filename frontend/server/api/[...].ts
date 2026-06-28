import { defineEventHandler, proxyRequest, createError, getRequestURL, readBody } from 'h3'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const backendUrl = config.backendUrl || 'http://localhost:8899'

  const reqUrl = getRequestURL(event)
  const path = reqUrl.pathname.replace(/^\/api/, '') + reqUrl.search
  const target = backendUrl + path

  console.log('[API Proxy]', event.method, event.path, '->', target)

  try {
    const result = await proxyRequest(event, target, {
      headers: {
        host: new URL(backendUrl).host,
      },
      fetch: $fetch.native,
      cookieDomainRewrite: '',
    })
    return result
  } catch (err: any) {
    console.error('[API Proxy Error]', target, err.statusCode || err.status, err.message)
    throw createError({
      statusCode: err.statusCode || 502,
      statusMessage: err.statusMessage || 'Bad Gateway',
      message: err.message || 'Proxy error',
    })
  }
})
