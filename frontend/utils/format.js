/**
 * 格式化时长（秒 → mm:ss）
 * @param {number} seconds 秒数，-1 表示"已看完"
 */
export const formatDuration = (seconds) => {
  if (seconds === -1) return '已看完'
  const minutes = String(Math.floor(seconds / 60)).padStart(2, '0')
  const secs = String(seconds % 60).padStart(2, '0')
  return `${minutes}:${secs}`
}

/**
 * 格式化时长简写（秒 → Xh Xm / Xm Xs / Xs）
 * @param {number} seconds 秒数
 */
export const formatDurationShort = (seconds) => {
  if (seconds < 0) return '0s'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

/**
 * 格式化 Unix 时间戳为相对时间或绝对时间
 * @param {number} timestamp Unix 时间戳（秒）
 */
export const formatTimestamp = (timestamp) => {
  if (!timestamp) return '时间未知'

  const date = new Date(timestamp * 1000)
  const now = new Date()

  if (isNaN(date.getTime())) return '时间未知'

  const isToday = now.toDateString() === date.toDateString()
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  const isYesterday = yesterday.toDateString() === date.toDateString()
  const timeString = date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })

  const pad = (n) => String(n).padStart(2, '0')
  const mmdd = `${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
  const yyyymmdd = `${date.getFullYear()}-${mmdd}`

  if (isToday) {
    return timeString
  } else if (isYesterday) {
    return `昨天 ${timeString}`
  } else if (now.getFullYear() === date.getFullYear()) {
    return `${mmdd} ${timeString}`
  } else {
    return `${yyyymmdd} ${timeString}`
  }
}

/**
 * 格式化业务类型
 * @param {string} business 业务类型代码
 */
export const getBusinessType = (business) => {
  const types = {
    archive: '稿件',
    cheese: '课堂',
    pgc: '电影',
    live: '直播',
    'article-list': '专栏',
    'article': '专栏',
  }
  return types[business] || business || '其他'
}
