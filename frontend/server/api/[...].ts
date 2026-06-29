import { defineEventHandler, createError, getRequestURL, getHeaders, setResponseHeader, readBody } from 'h3'
import http from 'node:http'
import https from 'node:https'
import { URL } from 'node:url'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const backendUrl = config.backendUrl || 'http://localhost:8899'

  const reqUrl = getRequestURL(event)
  const path = reqUrl.pathname.replace(/^\/api/, '') + reqUrl.search
  const target = backendUrl + path

  console.log('[API Proxy]', event.method, event.path, '->', target)

  try {
    const parsedUrl = new URL(target)
    const isHttps = parsedUrl.protocol === 'https:'
    const client = isHttps ? https : http

    const reqHeaders = getHeaders(event)
    const method = event.method || 'GET'

    const options: http.RequestOptions = {
      hostname: parsedUrl.hostname,
      port: parsedUrl.port || (isHttps ? 443 : 80),
      path: parsedUrl.pathname + parsedUrl.search,
      method,
      headers: {
        ...reqHeaders,
        host: parsedUrl.host,
        connection: 'close',
      },
      family: 4,
    }

    let reqBody: Buffer | null = null
    if (method !== 'GET' && method !== 'HEAD') {
      try {
        const body = await readBody(event)
        if (body !== undefined && body !== null) {
          if (typeof body === 'string') {
            reqBody = Buffer.from(body)
          } else if (Buffer.isBuffer(body)) {
            reqBody = body
          } else {
            reqBody = Buffer.from(JSON.stringify(body))
            if (!options.headers['content-type']) {
              options.headers['content-type'] = 'application/json'
            }
          }
          options.headers['content-length'] = reqBody.length.toString()
        }
      } catch {
        reqBody = null
      }
    }

    return await new Promise((resolve, reject) => {
      const proxyReq = client.request(options, (proxyRes) => {
        const respHeaders: Record<string, string | string[]> = {}
        for (const [key, value] of Object.entries(proxyRes.headers)) {
          if (key.toLowerCase() === 'transfer-encoding') continue
          if (key.toLowerCase() === 'connection') continue
          if (key.toLowerCase() === 'content-length') continue
          if (value !== undefined) {
            respHeaders[key] = value
            setResponseHeader(event, key, value as any)
          }
        }

        const chunks: Buffer[] = []
        proxyRes.on('data', (chunk) => chunks.push(chunk))
        proxyRes.on('end', () => {
          const body = Buffer.concat(chunks)
          const contentType = proxyRes.headers['content-type'] || ''
          if (contentType.includes('application/json')) {
            try {
              resolve(JSON.parse(body.toString()))
            } catch {
              resolve(body.toString())
            }
          } else if (contentType.includes('text/')) {
            resolve(body.toString())
          } else {
            resolve(body)
          }
        })
        proxyRes.on('error', (err) => {
          reject(err)
        })
      })

      proxyReq.on('error', (err: any) => {
        console.error('[API Proxy Request Error]', {
          target,
          method,
          code: err.code,
          message: err.message,
          hostname: parsedUrl.hostname,
          port: parsedUrl.port,
        })
        reject(err)
      })

      proxyReq.setTimeout(30000, () => {
        proxyReq.destroy(new Error('Request timeout'))
      })

      if (reqBody) {
        proxyReq.write(reqBody)
      }
      proxyReq.end()
    })
  } catch (err: any) {
    console.error('[API Proxy Error]', {
      target,
      method: event.method,
      name: err?.name,
      message: err?.message,
      code: err?.code,
      cause: err?.cause?.message || err?.cause,
    })

    throw createError({
      statusCode: 502,
      statusMessage: 'Bad Gateway',
      message: err?.message || 'Proxy error',
    })
  }
})
