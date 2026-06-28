<template>
  <div
    class="mx-auto max-w-2xl cursor-pointer transition-all duration-200 ease-in-out lg:max-w-4xl relative group"
    :class="{
      'glass-card-hover p-3': !isBatchMode || !isSelected,
      'ring-2 ring-accent bg-accent/5': isBatchMode && isSelected
    }"
    @click="handleClick"
  >
    <!-- Article type: full-width cover -->
    <div v-if="record.business === 'article-list' || record.business === 'article'">
      <div class="mb-2">
        <div class="line-clamp-2 text-gray-900 dark:text-gray-100 text-sm font-medium"
          v-html="isPrivacyMode ? '******' : highlightedTitle"
          :class="{ 'blur-sm': isPrivacyMode }"></div>
      </div>
      <div class="relative h-28 w-full overflow-hidden rounded-xl">
        <!-- Action buttons -->
        <div v-if="!isBatchMode" class="absolute right-2 top-2 z-20 hidden sm:group-hover:flex items-center gap-1.5">
          <div v-if="record.business === 'archive'" class="glass-icon-btn !w-7 !h-7" @click.stop.prevent="handleDownload" title="下载视频">
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
          </div>
          <div v-if="record.business !== 'live'" class="glass-icon-btn !w-7 !h-7" @click.stop.prevent="handleFavorite" title="收藏">
            <svg class="w-3.5 h-3.5" :class="isVideoFavorited ? 'text-yellow-400' : ''" :fill="isVideoFavorited ? 'currentColor' : 'none'" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" /></svg>
          </div>
          <div class="glass-icon-btn !w-7 !h-7 hover:!bg-red-500/20 hover:!text-red-500" @click.stop="handleDelete" title="删除">
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
          </div>
        </div>
        <!-- Batch checkbox -->
        <div v-if="isBatchMode" class="absolute left-2 top-2 z-10" @click.stop="$emit('toggle-selection', record)">
          <div class="w-5 h-5 rounded-lg border-2 flex items-center justify-center transition-all" :class="isSelected ? 'bg-accent border-accent' : 'border-white/80 bg-black/20'">
            <svg v-if="isSelected" class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" /></svg>
          </div>
        </div>
        <!-- Downloaded badge -->
        <div v-if="isDownloaded && record.business === 'archive'" class="absolute left-0 top-0 z-10">
          <div class="glass-badge !bg-green-500/90 !text-white !border-0 rounded-br-lg rounded-tl-none text-[10px]">
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
            <span>已下载</span>
          </div>
        </div>
        <!-- Favorited badge -->
        <div v-if="isVideoFavorited && record.business !== 'live'" class="absolute right-0 top-0 z-10">
          <div class="glass-badge !bg-amber-500/90 !text-white !border-0 rounded-tl-lg rounded-br-none text-[10px]">
            <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24"><path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" /></svg>
            <span>已收藏</span>
          </div>
        </div>
        <img :src="normalizeImageUrl(record.cover || record.covers[0])" class="h-full w-full object-cover" :class="{ 'blur-md': isPrivacyMode }" alt="" />
      </div>
      <div class="mt-2 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
        <div v-if="record.business !== 'cheese' && record.business !== 'pgc'" class="flex items-center gap-2" @click.stop>
          <img :src="normalizeImageUrl(record.author_face)" class="w-4 h-4 rounded-full" :class="{ 'blur-md': isPrivacyMode }" @click="handleAuthorClick" />
          <span class="cursor-pointer hover:text-accent transition-colors" @click="handleAuthorClick" v-html="isPrivacyMode ? '******' : highlightedAuthorName"></span>
        </div>
        <div class="flex items-center gap-2">
          <span v-if="record.dt === 1 || record.dt === 3 || record.dt === 5 || record.dt === 7" class="text-[10px]">📱</span>
          <span v-else-if="record.dt === 2 || record.dt === 33" class="text-[10px]">🖥</span>
          <span v-else-if="record.dt === 4 || record.dt === 6" class="text-[10px]">📟</span>
          <span>{{ formatTimestamp(record.view_at) }}</span>
        </div>
      </div>
    </div>

    <!-- Other types: horizontal layout -->
    <div v-else class="flex gap-3">
      <!-- Cover -->
      <div class="relative h-20 w-32 flex-shrink-0 overflow-hidden rounded-xl sm:h-28 sm:w-40">
        <div v-if="isBatchMode" class="absolute left-2 top-2 z-10" @click.stop="$emit('toggle-selection', record)">
          <div class="w-5 h-5 rounded-lg border-2 flex items-center justify-center transition-all" :class="isSelected ? 'bg-accent border-accent' : 'border-white/80 bg-black/20'">
            <svg v-if="isSelected" class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" /></svg>
          </div>
        </div>
        <div v-if="isDownloaded && record.business === 'archive'" class="absolute left-0 top-0 z-10">
          <div class="glass-badge !bg-green-500/90 !text-white !border-0 rounded-br-lg rounded-tl-none text-[10px]">
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
            <span>已下载</span>
          </div>
        </div>
        <div v-if="isVideoFavorited && record.business !== 'live'" class="absolute right-0 top-0 z-10">
          <div class="glass-badge !bg-amber-500/90 !text-white !border-0 rounded-tl-lg rounded-br-none text-[10px]">
            <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24"><path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" /></svg>
            <span>已收藏</span>
          </div>
        </div>
        <img v-if="record.cover" :src="normalizeImageUrl(record.cover)" class="h-full w-full object-cover" :class="{ 'blur-md': isPrivacyMode }" alt="" />
        <div v-else v-for="(cover, index) in record.covers" :key="index" class="mb-1">
          <img :src="normalizeImageUrl(cover)" class="h-full w-full object-cover" :class="{ 'blur-md': isPrivacyMode }" alt="" />
        </div>
        <!-- Duration & progress -->
        <div v-if="record.business !== 'article-list' && record.business !== 'article' && record.business !== 'live'">
          <div class="absolute bottom-1 right-1 rounded-lg bg-black/60 backdrop-blur-sm px-1.5 py-0.5 text-[10px] font-medium text-white">
            <span>{{ formatDuration(record.progress) }}</span>
            <span class="opacity-50 mx-0.5">/</span>
            <span>{{ formatDuration(record.duration) }}</span>
          </div>
          <div class="absolute bottom-0 left-0 h-0.5 w-full bg-black/20">
            <div class="h-full bg-accent rounded-full" :style="{ width: getProgressWidth(record.progress, record.duration) }"></div>
          </div>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0 flex flex-col justify-between relative">
        <!-- Action buttons -->
        <div v-if="!isBatchMode" class="absolute right-0 top-0 z-20 hidden sm:group-hover:flex items-center gap-1.5">
          <div v-if="record.business === 'archive'" class="glass-icon-btn !w-7 !h-7" @click.stop.prevent="handleDownload" title="下载">
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
          </div>
          <div v-if="record.business !== 'live'" class="glass-icon-btn !w-7 !h-7" @click.stop.prevent="handleFavorite" title="收藏">
            <svg class="w-3.5 h-3.5" :class="isVideoFavorited ? 'text-yellow-400' : ''" :fill="isVideoFavorited ? 'currentColor' : 'none'" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" /></svg>
          </div>
          <div class="glass-icon-btn !w-7 !h-7 hover:!bg-red-500/20 hover:!text-red-500" @click.stop="handleDelete" title="删除">
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
          </div>
        </div>

        <div>
          <div class="line-clamp-2 text-sm font-medium text-gray-900 dark:text-gray-100"
            v-html="isPrivacyMode ? '******' : highlightedTitle"
            :class="{ 'blur-sm': isPrivacyMode }"></div>
        </div>

        <div class="flex items-center gap-2 mt-1">
          <span v-if="record.business !== 'pgc'" class="glass-tag text-[10px]">{{ record.tag_name }}</span>
          <!-- Remark -->
          <div class="flex-1 relative" @click.stop>
            <div class="flex items-center gap-1">
              <span class="text-[10px] text-accent">备注:</span>
              <input type="text" v-model="remarkContent" @focus="handleRemarkFocus" @blur="handleRemarkBlur"
                placeholder="添加备注..." :disabled="isPrivacyMode"
                class="flex-1 min-w-0 px-1.5 py-0.5 text-[10px] text-accent bg-transparent border-b border-transparent hover:border-gray-200 dark:hover:border-gray-600 focus:border-accent focus:ring-0 transition-colors placeholder-accent/40"
                :class="{ 'blur-sm': isPrivacyMode }" />
              <span v-if="remarkTime" class="text-[10px] text-gray-400">{{ formatRemarkTime(remarkTime) }}</span>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mt-1">
          <div v-if="record.business === 'pgc'" class="text-gray-500 dark:text-gray-400 truncate">{{ record.long_title }}</div>
          <div v-else class="flex items-center gap-2 min-w-0" @click.stop>
            <img :src="normalizeImageUrl(record.author_face)" class="w-4 h-4 rounded-full flex-shrink-0" :class="{ 'blur-md': isPrivacyMode }" @click="handleAuthorClick" />
            <span class="cursor-pointer hover:text-accent transition-colors truncate" @click="handleAuthorClick" v-html="isPrivacyMode ? '******' : highlightedAuthorName"></span>
          </div>
          <div class="flex items-center gap-2 flex-shrink-0">
            <span v-if="record.dt === 1 || record.dt === 3 || record.dt === 5 || record.dt === 7" class="text-[10px]">📱</span>
            <span v-else-if="record.dt === 2 || record.dt === 33" class="text-[10px]">🖥</span>
            <span v-else-if="record.dt === 4 || record.dt === 6" class="text-[10px]">📟</span>
            <span>{{ formatTimestamp(record.view_at) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Download dialog -->
    <Teleport to="body">
      <DownloadDialog v-model:show="showDownloadDialog" :video-info="{ title: record.title, author: record.author_name, bvid: record.bvid, cover: record.cover || record.covers?.[0], cid: record.cid }" :is-batch-download="false" />
    </Teleport>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMediaQuery } from '@vueuse/core'
import { usePrivacyStore } from '../../store/privacy'
import { showDialog, showNotify } from 'vant'
import { batchDeleteHistory, updateVideoRemark, deleteBilibiliHistory } from '../../api/api'
import 'vant/es/dialog/style'
import 'vant/es/popup/style'
import 'vant/es/field/style'
import DownloadDialog from './DownloadDialog.vue'
import { openInBrowser } from '@/utils/openUrl.js'
import { normalizeImageUrl } from '@/utils/imageUrl.js'
import { saveHistoryRecord } from '@/utils/historyRecordStore.js'

const { isPrivacyMode } = usePrivacyStore()
const router = useRouter()
const isSmallScreen = useMediaQuery('(max-width: 639px)')

const props = defineProps({
  record: { type: Object, required: true },
  searchKeyword: { type: String, default: '' },
  searchType: { type: String, default: 'title' },
  isBatchMode: { type: Boolean, default: false },
  isSelected: { type: Boolean, default: false },
  remarkData: { type: Object, default: () => ({}) },
  isDownloaded: { type: Boolean, default: false },
  isVideoFavorited: { type: Boolean, default: false }
})

const emit = defineEmits(['toggle-selection', 'refresh-data', 'remark-updated', 'favorite'])

const remarkContent = ref('')
const originalRemark = ref('')
const remarkTime = ref(null)

const highlightText = (text) => {
  if (!props.searchKeyword || !text) return text
  const keywords = props.searchKeyword.split(/\s+/).filter(k => k)
  let highlightedText = text
  keywords.forEach(keyword => {
    const regex = new RegExp(keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
    highlightedText = highlightedText.replace(regex, match => `<span class="text-accent font-medium">${match}</span>`)
  })
  return highlightedText
}

const highlightedTitle = computed(() => {
  if (!props.searchKeyword) return props.record.title
  if (props.searchType === 'all' || props.searchType === 'title') return highlightText(props.record.title)
  return props.record.title
})

const highlightedAuthorName = computed(() => {
  if (!props.searchKeyword) return props.record.author_name
  if (props.searchType === 'all' || props.searchType === 'author') return highlightText(props.record.author_name)
  return props.record.author_name
})

const handleClick = () => {
  if (props.isBatchMode) {
    emit('toggle-selection', props.record)
  } else if (isSmallScreen.value) {
    navigateToActionPage()
  } else {
    handleContentClick()
  }
}

const navigateToActionPage = () => {
  saveHistoryRecord(props.record)
  router.push({ name: 'VideoActions', params: { bvid: props.record.bvid, viewAt: String(props.record.view_at) } })
}

const handleContentClick = async () => {
  let url = ''
  switch (props.record.business) {
    case 'archive': url = `https://www.bilibili.com/video/${props.record.bvid}`; break
    case 'article': url = `https://www.bilibili.com/read/cv${props.record.oid}`; break
    case 'article-list': url = `https://www.bilibili.com/read/readlist/rl${props.record.oid}`; break
    case 'live': url = `https://live.bilibili.com/${props.record.oid}`; break
    case 'pgc': url = props.record.uri || `https://www.bilibili.com/bangumi/play/ep${props.record.epid}`; break
    case 'cheese': url = props.record.uri || `https://www.bilibili.com/cheese/play/ep${props.record.epid}`; break
    default: return
  }
  if (url) await openInBrowser(url)
}

const handleAuthorClick = async () => {
  await openInBrowser(`https://space.bilibili.com/${props.record.author_mid}`)
}

const formatTimestamp = (timestamp) => {
  if (!timestamp) return '时间未知'
  try {
    const date = new Date(timestamp * 1000)
    const now = new Date()
    if (isNaN(date.getTime())) return '时间未知'
    const isToday = now.toDateString() === date.toDateString()
    const isYesterday = new Date(now.setDate(now.getDate() - 1)).toDateString() === date.toDateString()
    const timeString = date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    if (isToday) return timeString
    if (isYesterday) return `昨天 ${timeString}`
    if (now.getFullYear() === date.getFullYear()) return `${date.getMonth() + 1}-${date.getDate()} ${timeString}`
    return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()} ${timeString}`
  } catch (error) { return '时间未知' }
}

const formatDuration = (seconds) => {
  if (seconds === -1) return '已看完'
  return `${String(Math.floor(seconds / 60)).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`
}

const getProgressWidth = (progress, duration) => {
  if (progress === -1) return '100%'
  if (duration === 0) return '0%'
  return `${(progress / duration) * 100}%`
}

const handleDelete = async () => {
  try {
    const syncDeleteToBilibili = localStorage.getItem('syncDeleteToBilibili') === 'true'
    await showDialog({
      title: '确认删除',
      message: syncDeleteToBilibili
        ? '确定要删除这条记录吗？此操作将同时删除B站服务器上的历史记录，不可恢复。'
        : '确定要删除这条记录吗？此操作不可恢复。',
      showCancelButton: true, confirmButtonText: '确认删除', cancelButtonText: '取消', confirmButtonColor: '#fb7299'
    })
    if (syncDeleteToBilibili) {
      try {
        let kid = ''
        const business = props.record.business || 'archive'
        switch (business) {
          case 'archive': kid = `${business}_${props.record.oid}`; break
          case 'live': kid = `${business}_${props.record.oid}`; break
          case 'article': kid = `${business}_${props.record.oid}`; break
          case 'pgc': kid = `${business}_${props.record.oid || props.record.ssid}`; break
          case 'article-list': kid = `${business}_${props.record.oid || props.record.rlid}`; break
          default: kid = `${business}_${props.record.oid || props.record.bvid}`; break
        }
        if (kid) {
          const biliResponse = await deleteBilibiliHistory(kid, true)
          if (biliResponse.data.status !== 'success') throw new Error(biliResponse.data.message || '删除B站历史记录失败')
        }
      } catch (error) { console.error('B站历史记录删除失败:', error) }
    }
    const response = await batchDeleteHistory([{ bvid: props.record.bvid, view_at: props.record.view_at }])
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: response.data.message + (syncDeleteToBilibili ? '，并已同步删除B站历史记录' : '') })
      emit('refresh-data')
    } else { throw new Error(response.data.message || '删除失败') }
  } catch (error) {
    if (error.toString().includes('cancel')) return
    showNotify({ type: 'danger', message: error.response?.data?.detail || error.message || '删除失败' })
  }
}

const formatRemarkTime = (timestamp) => {
  if (!timestamp) return ''
  return new Date(timestamp * 1000).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const initRemark = () => {
  const key = `${props.record.bvid}_${props.record.view_at}`
  const data = props.remarkData[key]
  if (data) {
    remarkContent.value = data.remark || ''
    remarkTime.value = data.remark_time || null
    originalRemark.value = remarkContent.value
  } else {
    remarkContent.value = ''
    remarkTime.value = null
    originalRemark.value = ''
  }
}

const handleRemarkBlur = async () => {
  if (remarkContent.value === originalRemark.value) return
  try {
    const response = await updateVideoRemark(props.record.bvid, props.record.view_at, remarkContent.value)
    if (response.data.status === 'success') {
      if (remarkContent.value) showNotify({ type: 'success', message: '备注已保存' })
      originalRemark.value = remarkContent.value
      remarkTime.value = response.data.data.remark_time
      emit('remark-updated', { bvid: props.record.bvid, view_at: props.record.view_at, remark: remarkContent.value, remark_time: response.data.data.remark_time })
    }
  } catch (error) {
    showNotify({ type: 'danger', message: `保存备注失败：${error.message}` })
    remarkContent.value = originalRemark.value
  }
}

onMounted(() => { initRemark() })
watch(() => props.remarkData, () => { initRemark() }, { deep: true })

const showDownloadDialog = ref(false)
const handleDownload = () => { showDownloadDialog.value = true }
const handleFavorite = () => { emit('favorite', props.record) }
</script>
