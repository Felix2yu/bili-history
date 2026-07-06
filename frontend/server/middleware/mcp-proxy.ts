import { defineEventHandler, getHeaders, setResponseHeader, setResponseStatus, readRawBody } from 'h3'
import http from 'node:http'
import https from 'node:https'
import { URL } from 'node:url'

export default defineEventHandler(async (event) => {
  const path = event.path || ''
  if (!path.startsWith('/mcp')) return

  const config = useRuntimeConfig()
  const backendUrl = config.backendUrl || 'http://localhost:8899'
  const method = event.method || 'GET'
  const target = backendUrl + path

  const parsedUrl = new URL(target)
  const isHttps = parsedUrl.protocol === 'https:'
  const client = isHttps ? https : http
  const reqHeaders = getHeaders(event)

  const options: http.RequestOptions = {
    hostname: parsedUrl.hostname,
    port: parsedUrl.port || (isHttps ? 443 : 80),
    path: parsedUrl.pathname + parsedUrl.search,
    method,
    headers: { ...reqHeaders, host: parsedUrl.host, connection: 'close' },
    family: 4,
  }

  let reqBody: Buffer | null = null
  if (method !== 'GET' && method !== 'HEAD') {
    const raw = await readRawBody(event).catch(() => null)
    if (raw) {
      reqBody = Buffer.from(raw)
      options.headers!['content-length'] = reqBody.length.toString()
    }
  }

  return new Promise((resolve, reject) => {
    const proxyReq = client.request(options, (proxyRes) => {
      const skip = new Set(['transfer-encoding', 'connection'])
      for (const [k, v] of Object.entries(proxyRes.headers)) {
        if (!skip.has(k.toLowerCase()) && v !== undefined) {
          setResponseHeader(event, k, v as any)
        }
      }
      setResponseHeader(event, 'content-type', proxyRes.headers['content-type'] || 'application/json')
      setResponseStatus(event, proxyRes.statusCode || 200)
      const chunks: Buffer[] = []
      proxyRes.on('data', (c) => chunks.push(c))
      proxyRes.on('end', () => resolve(Buffer.concat(chunks)))
      proxyRes.on('error', reject)
    })
    proxyReq.on('error', reject)
    proxyReq.setTimeout(300000, () => proxyReq.destroy(new Error('timeout')))
    if (reqBody) proxyReq.write(reqBody)
    proxyReq.end()
  })
})
