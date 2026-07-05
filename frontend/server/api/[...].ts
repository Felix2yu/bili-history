import { defineEventHandler, createError, getHeaders, setResponseHeader, setResponseStatus, readBody } from 'h3'
import http from 'node:http'
import https from 'node:https'
import { URL } from 'node:url'
import { ServerResponse } from 'node:http'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const backendUrl = config.backendUrl || 'http://localhost:8899'

  const method = event.method || 'GET'
  // Use the raw request URL (event.node.req.url) instead of the parsed
  // pathname so that percent-encoded characters like %2F (used in task IDs
  // that contain "/") are preserved when forwarding to the backend.
  const rawReqUrl = event.node.req.url || ''
  const path = rawReqUrl.replace(/^\/api/, '')
  const target = backendUrl + path

  try {
    const parsedUrl = new URL(target)
    const isHttps = parsedUrl.protocol === 'https:'
    const client = isHttps ? https : http

    const reqHeaders = getHeaders(event)

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
        const skipHeaders = new Set(['transfer-encoding', 'connection'])
        for (const [key, value] of Object.entries(proxyRes.headers)) {
          if (!skipHeaders.has(key.toLowerCase()) && value !== undefined) {
            setResponseHeader(event, key, value as any)
          }
        }

        const ct = (proxyRes.headers['content-type'] || 'application/json') as string
        setResponseHeader(event, 'content-type', ct)
        setResponseStatus(event, proxyRes.statusCode || 200)

        // SSE: stream data directly to client without buffering
        if (ct.includes('text/event-stream')) {
          const nodeRes = event.node.res
          proxyRes.on('data', (chunk) => {
            nodeRes.write(chunk)
          })
          proxyRes.on('end', () => {
            nodeRes.end()
            resolve(undefined as any)
          })
          proxyRes.on('error', (err) => {
            reject(err)
          })
          return
        }

        // Non-SSE: buffer entire response (original behavior)
        const chunks: Buffer[] = []
        proxyRes.on('data', (chunk) => chunks.push(chunk))
        proxyRes.on('end', () => {
          const body = Buffer.concat(chunks)
          resolve(body)
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
        })
        reject(err)
      })

      proxyReq.setTimeout(300000, () => {
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
      message: err?.message,
    })

    throw createError({
      statusCode: 502,
      statusMessage: 'Bad Gateway',
      message: err?.message || 'Proxy error',
    })
  }
})
