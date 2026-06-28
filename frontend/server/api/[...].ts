import { defineEventHandler, createError, getRequestURL, readBody, getHeaders, setResponseHeaders } from 'h3'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const backendUrl = config.backendUrl || 'http://localhost:8899'

  const reqUrl = getRequestURL(event)
  const path = reqUrl.pathname.replace(/^\/api/, '') + reqUrl.search
  const target = backendUrl + path

  console.log('[API Proxy]', event.method, event.path, '->', target)

  try {
    const reqHeaders = getHeaders(event)
    const method = event.method || 'GET'

    let body: any = undefined
    if (method !== 'GET' && method !== 'HEAD') {
      try {
        body = await readBody(event)
      } catch {
        body = undefined
      }
    }

    const response = await $fetch.raw(target, {
      method,
      headers: {
        ...reqHeaders,
        host: new URL(backendUrl).host,
      },
      body,
      credentials: 'include',
      onResponse({ response }) {
        const respHeaders: Record<string, string> = {}
        for (const [key, value] of response.headers.entries()) {
          if (key.toLowerCase() === 'transfer-encoding') continue
          if (key.toLowerCase() === 'connection') continue
          respHeaders[key] = value
        }
        setResponseHeaders(event, respHeaders)
      },
    })

    return response._data
  } catch (err: any) {
    console.error('[API Proxy Error]', {
      target,
      method: event.method,
      name: err?.name,
      message: err?.message,
      statusCode: err?.statusCode || err?.status || err?.response?.status,
      statusText: err?.statusText || err?.response?.statusText,
      cause: err?.cause?.message || err?.cause,
      code: err?.code,
    })

    const status = err?.statusCode || err?.status || err?.response?.status || 502
    const statusMessage = err?.statusText || err?.response?.statusText || 'Bad Gateway'

    if (err?.response?._data) {
      return err.response._data
    }

    throw createError({
      statusCode: status,
      statusMessage,
      message: err?.message || 'Proxy error',
    })
  }
})
