<template>
  <div class="min-h-screen bg-gray-50/30 dark:bg-gray-900 pb-20 md:pb-0">
    <div class="py-6">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="glass-card overflow-hidden">
          <div class="border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between">
            <h2 class="text-lg font-medium text-gray-900 dark:text-gray-100">稍后再看</h2>
            <div class="flex items-center space-x-3">
              <span v-if="syncing" class="text-xs text-[#fb7299] flex items-center">
                <svg class="animate-spin -ml-1 mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                同步中...
              </span>
              <span class="text-sm text-gray-500 dark:text-gray-400">共 {{ filteredVideos.length }} / {{ videos.length }} 个视频</span>
              <button
                v-if="!selectMode && videos.length > 0"
                @click="enterSelectMode"
                class="px-3 py-1 text-xs rounded-md border border-[#fb7299] text-[#fb7299] hover:bg-[#fb7299]/10 transition-colors"
              >
                批量管理
              </button>
            </div>
          </div>

          <div v-if="selectMode" class="border-b border-gray-200 dark:border-gray-700 px-4 py-2 bg-[#fb7299]/5 flex items-center justify-between flex-wrap gap-2">
            <div class="flex items-center space-x-3">
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input
                  type="checkbox"
                  :checked="allFilteredSelected"
                  :indeterminate.prop="someFilteredSelected"
                  @change="toggleSelectAllFiltered"
                  class="w-3.5 h-3.5 rounded border-gray-300 text-[#fb7299] focus:ring-[#fb7299]"
                />
                <span class="text-xs text-gray-700 dark:text-gray-300">全选当前 ({{ filteredVideos.length }})</span>
              </label>
              <span class="text-xs text-gray-500 dark:text-gray-400">已选 {{ selectedBvids.size }} 个</span>
            </div>
            <div class="flex items-center space-x-2">
              <button
                @click="batchDeleteSelected"
                :disabled="selectedBvids.size === 0 || deleting"
                class="px-3 py-1 text-xs rounded-md bg-red-500 text-white hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center space-x-1"
              >
                <svg v-if="deleting" class="animate-spin h-3 w-3" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
                <span>{{ deleting ? '删除中...' : '删除选中' }}</span>
              </button>
              <button
                @click="exitSelectMode"
                :disabled="deleting"
                class="px-3 py-1 text-xs rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors"
              >
                取消
              </button>
            </div>
          </div>

          <VideoFilterBar
            v-if="!loading && videos.length > 0"
            v-model:sort-key="sortKey"
            v-model:sort-order="sortOrder"
            v-model:selected-category="selectedCategory"
            v-model:selected-owner="selectedOwner"
            :sort-options="sortOptions"
            :all-categories="allCategories"
            :all-owners="allOwners"
          />

          <div class="p-5">
            <div v-if="loading" class="flex justify-center py-20">
              <div class="inline-flex items-center px-4 py-2 bg-white dark:bg-gray-800 rounded-md shadow text-gray-900 dark:text-gray-100">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-[#fb7299]" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>加载中...</span>
              </div>
            </div>

            <div v-else-if="error" class="text-center py-20">
              <svg class="w-16 h-16 mx-auto text-red-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <p class="mt-4 text-red-500">{{ error }}</p>
              <button
                @click="fetchWatchLater"
                class="mt-4 px-4 py-2 bg-[#fb7299] text-white rounded-md hover:bg-[#fb7299]/90 transition-colors"
              >
                重试
              </button>
            </div>

            <div v-else-if="videos.length === 0" class="text-center py-20">
              <svg class="w-16 h-16 mx-auto text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="mt-4 text-gray-500 dark:text-gray-400">稍后再看列表为空</p>
            </div>

            <div v-else-if="filteredVideos.length === 0" class="text-center py-20">
              <p class="text-gray-500 dark:text-gray-400">没有匹配的视频</p>
            </div>

            <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
              <VideoGridCard
                v-for="video in filteredVideos"
                :key="video.bvid"
                :video="video"
                :select-mode="selectMode"
                :is-selected="isSelected(video.bvid)"
                :time-field="'add_at'"
                @click="openVideo(video)"
                @toggle-select="toggleSelect(video.bvid)"
              >
                <template #actions="{ video: v }">
                  <div
                    class="glass-icon-btn !w-6 !h-6"
                    @click.stop.prevent="handleDownload(v)"
                    title="下载"
                  >
                    <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                    </svg>
                  </div>
                  <div
                    class="glass-icon-btn !w-6 !h-6 hover:!bg-red-500/20 hover:!text-red-500"
                    @click.stop="confirmDeleteOne(v)"
                    title="从稍后再看移除"
                  >
                    <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </div>
                </template>
              </VideoGridCard>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="confirmDialog.show" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="cancelConfirm">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-sm w-full mx-4 p-4">
        <div class="flex items-start space-x-3">
          <div class="flex-shrink-0 w-9 h-9 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
            <svg class="w-5 h-5 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div class="flex-1">
            <h3 class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ confirmDialog.title }}</h3>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ confirmDialog.message }}</p>
          </div>
        </div>
        <div class="mt-4 flex justify-end space-x-2">
          <button
            @click="cancelConfirm"
            :disabled="deleting"
            class="px-3 py-1.5 text-xs rounded-md text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50 transition-colors"
          >
            取消
          </button>
          <button
            @click="executeConfirmedDelete"
            :disabled="deleting"
            class="px-3 py-1.5 text-xs rounded-md text-white bg-red-500 hover:bg-red-600 disabled:opacity-50 transition-colors flex items-center space-x-1"
          >
            <svg v-if="deleting" class="animate-spin h-3 w-3" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
            </svg>
            <span>{{ deleting ? '删除中...' : '确认删除' }}</span>
          </button>
        </div>
      </div>
    </div>

    <div v-if="toast.show" class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 px-4 py-2 rounded-md shadow-lg text-sm text-white" :class="toast.type === 'error' ? 'bg-red-500' : 'bg-green-500'">
      {{ toast.message }}
    </div>

    <DownloadDialog
      v-model:show="showDownloadDialog"
      :video-info="currentVideo"
      @download-complete="handleDownloadComplete"
    />

    <FavoriteDialog
      v-model="showFavoriteDialog"
      :video-info="currentVideo"
      @favorite-done="handleFavoriteDone"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useAsyncData } from '#imports'
import { showNotify } from 'vant'
import 'vant/es/notify/style'
import { getWatchLaterList, getWatchLaterLocal, removeFromWatchLater, batchRemoveFromWatchLater, favoriteResource, batchCheckFavoriteStatus } from '~/utils/api'
import { normalizeImageUrl } from '~/utils/imageUrl.js'
import VideoGridCard from '../VideoGridCard.vue'
import VideoFilterBar from '../VideoFilterBar.vue'
import DownloadDialog from '../DownloadDialog.vue'
import FavoriteDialog from '../FavoriteDialog.vue'

let imageObserver = null

function initImageObserver() {
  if (imageObserver) return
  imageObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target
        const src = img.dataset.src
        if (src) {
          img.src = src
          img.removeAttribute('data-src')
        }
        imageObserver.unobserve(img)
      }
    })
  }, { rootMargin: '200px' })
}

function observeImages() {
  nextTick(() => {
    if (!imageObserver) initImageObserver()
    document.querySelectorAll('img[data-src]').forEach(img => {
      imageObserver.observe(img)
    })
  })
}

const loading = ref(false)
const syncing = ref(false)
const deleting = ref(false)
const error = ref('')
const videos = ref([])

const sortKey = ref('add_at')
const sortOrder = ref('desc')
const selectedOwner = ref('')
const selectedCategory = ref('')

const selectMode = ref(false)
const selectedBvids = ref(new Set())
const confirmDialog = ref({ show: false, title: '', message: '', action: null })
const toast = ref({ show: false, message: '', type: 'success', timer: null })

const showDownloadDialog = ref(false)
const showFavoriteDialog = ref(false)
const currentVideo = ref({})
const favoriteStatus = ref({})

const sortOptions = [
  { key: 'add_at', label: '加入时间' },
  { key: 'duration', label: '时长' },
  { key: 'owner_name', label: '发布者' },
]

const allOwners = computed(() => {
  const map = {}
  for (const v of videos.value) {
    const name = v.owner_name || '未知'
    map[name] = (map[name] || 0) + 1
  }
  return Object.entries(map)
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
})

const allCategories = computed(() => {
  const map = {}
  for (const v of videos.value) {
    const name = v.tname || '未知分区'
    map[name] = (map[name] || 0) + 1
  }
  return Object.entries(map)
    .map(([tname, count]) => ({ tname, count }))
    .sort((a, b) => b.count - a.count)
})

const filteredVideos = computed(() => {
  let list = [...videos.value]
  if (selectedOwner.value) {
    list = list.filter(v => v.owner_name === selectedOwner.value)
  }
  if (selectedCategory.value) {
    list = list.filter(v => (v.tname || '未知分区') === selectedCategory.value)
  }
  list.sort((a, b) => {
    let va = a[sortKey.value]
    let vb = b[sortKey.value]
    if (sortKey.value === 'owner_name') {
      va = (va || '').toLowerCase()
      vb = (vb || '').toLowerCase()
      return sortOrder.value === 'asc' ? va.localeCompare(vb) : vb.localeCompare(va)
    }
    va = va || 0
    vb = vb || 0
    return sortOrder.value === 'asc' ? va - vb : vb - va
  })
  return list
})

const filteredBvids = computed(() => filteredVideos.value.map(v => v.bvid))
const allFilteredSelected = computed(() =>
  filteredBvids.value.length > 0 && filteredBvids.value.every(b => selectedBvids.value.has(b))
)
const someFilteredSelected = computed(() =>
  !allFilteredSelected.value && filteredBvids.value.some(b => selectedBvids.value.has(b))
)

function isSelected(bvid) {
  return selectedBvids.value.has(bvid)
}

function toggleSelect(bvid) {
  const next = new Set(selectedBvids.value)
  if (next.has(bvid)) {
    next.delete(bvid)
  } else {
    next.add(bvid)
  }
  selectedBvids.value = next
}

function toggleSelectAllFiltered() {
  const next = new Set(selectedBvids.value)
  if (allFilteredSelected.value) {
    filteredBvids.value.forEach(b => next.delete(b))
  } else {
    filteredBvids.value.forEach(b => next.add(b))
  }
  selectedBvids.value = next
}

function enterSelectMode() {
  selectMode.value = true
  selectedBvids.value = new Set()
}

function exitSelectMode() {
  selectMode.value = false
  selectedBvids.value = new Set()
}

function showToast(message, type = 'success') {
  if (toast.value.timer) clearTimeout(toast.value.timer)
  toast.value = { show: true, message, type, timer: setTimeout(() => { toast.value.show = false }, 3000) }
}

function confirmDeleteOne(video) {
  confirmDialog.value = {
    show: true,
    title: '从稍后再看移除',
    message: `确定要从稍后再看移除「${video.title}」吗？该操作会同时从 B 站稍后再看列表删除。`,
    action: async () => {
      deleting.value = true
      try {
        const response = await removeFromWatchLater(video.bvid)
        if (response.data.status === 'success') {
          videos.value = videos.value.filter(v => v.bvid !== video.bvid)
          showToast('已移除')
        } else {
          showToast(response.data.message || '删除失败', 'error')
        }
      } catch (e) {
        showToast('请求失败: ' + (e.message || '未知错误'), 'error')
      } finally {
        deleting.value = false
      }
    }
  }
}

function batchDeleteSelected() {
  const count = selectedBvids.value.size
  if (count === 0) return
  confirmDialog.value = {
    show: true,
    title: `批量删除 ${count} 个视频`,
    message: `确定要从稍后再看移除选中的 ${count} 个视频吗？该操作会同时从 B 站稍后再看列表删除，且不可撤销。`,
    action: async () => {
      deleting.value = true
      try {
        const bvids = Array.from(selectedBvids.value)
        const response = await batchRemoveFromWatchLater(bvids)
        if (response.data.status === 'success') {
          const data = response.data.data || {}
          const failedBvids = new Set(
            (data.results || []).filter(r => !r.success).map(r => r.bvid)
          )
          videos.value = videos.value.filter(v => !selectedBvids.value.has(v.bvid) || failedBvids.has(v.bvid))
          selectedBvids.value = failedBvids
          const success = data.success || 0
          const failed = data.failed || 0
          if (failed === 0) {
            showToast(`成功删除 ${success} 个视频`)
            exitSelectMode()
          } else {
            showToast(`成功 ${success} 个，失败 ${failed} 个`, 'error')
          }
        } else {
          showToast(response.data.message || '批量删除失败', 'error')
        }
      } catch (e) {
        showToast('请求失败: ' + (e.message || '未知错误'), 'error')
      } finally {
        deleting.value = false
      }
    }
  }
}

function cancelConfirm() {
  if (deleting.value) return
  confirmDialog.value = { show: false, title: '', message: '', action: null }
}

async function executeConfirmedDelete() {
  const action = confirmDialog.value.action
  if (!action) return
  try {
    await action()
  } finally {
    if (!deleting.value) {
      confirmDialog.value = { show: false, title: '', message: '', action: null }
    }
  }
}

const { data: initialData } = await useAsyncData('watchlater-initial', async () => {
  try {
    const response = await getWatchLaterList()
    if (response.data.status === 'success') {
      return { videos: response.data.data.list || [] }
    }
    return { videos: [] }
  } catch (error) {
    console.error('SSR 获取稍后再看失败:', error)
    return { videos: [] }
  }
})

if (initialData.value?.videos?.length > 0) {
  videos.value = initialData.value.videos
}

onMounted(async () => {
  initImageObserver()
  if (videos.value.length === 0) {
    await fetchLocal()
  }
  observeImages()
  syncFromBilibili()
})

watch(videos, () => { observeImages() })

onUnmounted(() => {
  if (imageObserver) {
    imageObserver.disconnect()
    imageObserver = null
  }
})

async function fetchWatchLater() {
  loading.value = true
  error.value = ''
  videos.value = []
  try {
    const response = await getWatchLaterList()
    if (response.data.status === 'success') {
      videos.value = response.data.data.list || []
    } else {
      error.value = response.data.message || '获取稍后再看列表失败'
    }
  } catch (e) {
    error.value = '请求失败: ' + (e.message || '未知错误')
  } finally {
    loading.value = false
  }
}

async function fetchLocal() {
  loading.value = true
  try {
    const response = await getWatchLaterLocal({ size: 500 })
    if (response.data.status === 'success') {
      const list = response.data.data.list || []
      if (list.length > 0) {
        videos.value = list
      }
    }
  } catch (e) {
    console.warn('读取本地数据库失败:', e)
  } finally {
    loading.value = false
  }
  checkFavoriteStatus()
}

async function syncFromBilibili() {
  syncing.value = true
  try {
    const response = await getWatchLaterList()
    if (response.data.status === 'success') {
      videos.value = response.data.data.list || []
    }
  } catch (e) {
    console.warn('后台同步失败:', e)
  } finally {
    syncing.value = false
  }
  checkFavoriteStatus()
}

function isVideoFavorited(aid) {
  if (!aid) return false
  const oidStr = String(aid)
  return Object.keys(favoriteStatus.value).some(key => {
    return String(key) === oidStr && favoriteStatus.value[key].is_favorited
  })
}

async function checkFavoriteStatus() {
  try {
    const aids = videos.value
      .filter(v => v.aid)
      .map(v => parseInt(v.aid, 10))
    if (aids.length === 0) return
    const response = await batchCheckFavoriteStatus({ aids })
    if (response.data.status === 'success') {
      const list = response.data.data || []
      const statusMap = {}
      for (const item of list) {
        statusMap[String(item.aid)] = item
      }
      favoriteStatus.value = statusMap
    }
  } catch (e) {
    console.warn('检查收藏状态失败:', e)
  }
}

function handleFavoriteDone() {
  checkFavoriteStatus()
}

function handleDownload(video) {
  currentVideo.value = {
    title: video.title,
    author: video.owner_name,
    bvid: video.bvid,
    cover: video.pic,
    cid: video.cid || 0,
  }
  showDownloadDialog.value = true
}

function handleDownloadComplete() {
  showNotify({ type: 'success', message: '下载完成' })
}

function openVideo(video) {
  if (video.link) {
    window.open(video.link, '_blank')
  }
}
</script>
