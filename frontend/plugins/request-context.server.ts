import { instance } from '~/utils/api'

export default defineNuxtPlugin(() => {
  const event = useRequestEvent()
  const cookie = event?.node?.req?.headers?.cookie
  if (!cookie) return

  instance.interceptors.request.use((config) => {
    if (!config.headers.Cookie) {
      config.headers.Cookie = cookie
    }
    return config
  })
})

