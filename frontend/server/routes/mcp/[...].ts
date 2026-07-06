import { defineEventHandler, createError, getHeaders, setResponseHeader, setResponseStatus, readBody } from 'h3'
import http from 'node:http'
import https from 'node:https'
import { URL } from 'node:url'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const backendUrl = config.backendUrl || 'http://localhost:8899'

  const method = event.method || 'GET'
  const rawReqUrl = event.node.req.url || ''
  const target = backendUrl + rawReqUrl

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

        const chunks: Buffer[] = []
        proxyRes.on('data', (chunk) => chunks.push(chunk))
        proxyRes.on('end', () => resolve(Buffer.concat(chunks)))
        proxyRes.on('error', reject)
      })

      proxyReq.on('error', reject)
      proxyReq.setTimeout(300000, () => proxyReq.destroy(new Error('timeout')))
      if (reqBody) proxyReq.write(reqBody)
      proxyReq.end()
    })
  } catch (err: any) {
    throw createError({ statusCode: 502, statusMessage: 'Bad Gateway', message: err?.message })
  }
})
