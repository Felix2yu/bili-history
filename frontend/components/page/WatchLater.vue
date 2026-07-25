<template>
  <div class="min-h-screen pb-20 md:pb-0 relative overflow-hidden">
    <div class="absolute inset-0 bg-gradient-to-br from-accent/5 via-transparent to-pink-500/5 dark:from-accent/10 dark:via-transparent dark:to-pink-500/10 pointer-events-none"></div>
    <div class="absolute top-20 -left-20 w-80 h-80 bg-accent/10 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute bottom-20 -right-20 w-80 h-80 bg-pink-500/10 rounded-full blur-3xl pointer-events-none"></div>

    <div class="relative py-6">
      <div class="mx-auto sm:px-2 lg:px-8">
        <div class="mb-6">
          <div class="flex items-center justify-between flex-wrap gap-4">
            <div class="flex items-center space-x-4">
              <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-accent to-pink-500 flex items-center justify-center shadow-lg shadow-accent/30">
                <svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div>
                <h1 class="text-2xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 dark:from-gray-100 dark:to-gray-400 bg-clip-text text-transparent">稍后再看</h1>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">保存你想稍后观看的视频</p>
              </div>
            </div>

            <div class="flex items-center gap-3">
              <div class="glass-card px-4 py-2.5 flex items-center gap-3">
                <div class="flex flex-col">
                  <span class="text-[10px] text-gray-500 dark:text-gray-400 uppercase tracking-wider">视频总数</span>
                  <span class="text-lg font-bold text-gray-900 dark:text-gray-100 leading-tight">{{ videos.length }}</span>
                </div>
                <div class="w-px h-8 bg-gray-200 dark:bg-gray-700"></div>
                <div class="flex flex-col">
                  <span class="text-[10px] text-gray-500 dark:text-gray-400 uppercase tracking-wider">当前显示</span>
                  <span class="text-lg font-bold text-accent leading-tight">{{ filteredVideos.length }}</span>
                </div>
                <div v-if="syncing" class="w-px h-8 bg-gray-200 dark:bg-gray-700"></div>
                <div v-if="syncing" class="flex items-center gap-1.5 text-accent">
                  <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  <span class="text-xs font-medium">同步中</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="glass-card overflow-hidden">
          <div class="border-b border-glass-border px-5 py-3.5 flex items-center justify-between bg-white/30 dark:bg-white/5">
            <div class="flex items-center gap-2">
              <svg class="w-5 h-5 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
              <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">视频列表</h2>
            </div>
            <div class="flex items-center space-x-3">
              <button
                v-if="!selectMode && videos.length > 0"
                @click="enterSelectMode"
                class="px-4 py-1.5 text-xs font-medium rounded-xl bg-accent/10 text-accent hover:bg-accent/20 transition-all duration-200 flex items-center gap-1.5"
              >
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                </svg>
                批量管理
              </button>
            </div>
          </div>

          <div v-if="selectMode" class="border-b border-glass-border px-5 py-3 bg-gradient-to-r from-accent/10 to-transparent dark:from-accent/15 dark:to-transparent flex items-center justify-between flex-wrap gap-2">
            <div class="flex items-center space-x-4">
              <label class="flex items-center space-x-2 cursor-pointer group">
                <input
                  type="checkbox"
                  :checked="allFilteredSelected"
                  :indeterminate.prop="someFilteredSelected"
                  @change="toggleSelectAllFiltered"
                  class="w-4 h-4 rounded-lg border-gray-300 dark:border-gray-600 text-accent focus:ring-accent focus:ring-offset-0 transition-colors"
                />
                <span class="text-sm text-gray-700 dark:text-gray-300 font-medium group-hover:text-accent transition-colors">全选当前 ({{ filteredVideos.length }})</span>
              </label>
              <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-accent/15 text-accent text-xs font-semibold rounded-lg">
                <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                已选 {{ selectedBvids.size }} 个
              </span>
            </div>
            <div class="flex items-center space-x-2">
              <button
                @click="batchDeleteSelected"
                :disabled="selectedBvids.size === 0 || deleting"
                class="px-4 py-1.5 text-xs font-semibold rounded-xl bg-gradient-to-r from-red-500 to-red-600 text-white hover:from-red-600 hover:to-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 flex items-center gap-1.5 shadow-md shadow-red-500/20"
              >
                <svg v-if="deleting" class="animate-spin h-3.5 w-3.5" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
                <svg v-else class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                <span>{{ deleting ? '删除中...' : '删除选中' }}</span>
              </button>
              <button
                @click="exitSelectMode"
                :disabled="deleting"
                class="px-4 py-1.5 text-xs font-semibold rounded-xl text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 transition-all duration-200"
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

          <div class="p-6">
            <div v-if="loading" class="flex justify-center py-20">
              <div class="glass-card px-6 py-4 flex items-center gap-3">
                <svg class="animate-spin h-5 w-5 text-accent" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">加载中...</span>
              </div>
            </div>

            <div v-else-if="error" class="text-center py-20">
              <div class="w-20 h-20 mx-auto rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center mb-4">
                <svg class="w-10 h-10 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </div>
              <p class="text-red-500 dark:text-red-400 font-medium">{{ error }}</p>
              <button
                @click="fetchWatchLater"
                class="mt-5 px-5 py-2.5 bg-gradient-to-r from-accent to-pink-500 text-white rounded-xl hover:from-accent/90 hover:to-pink-500/90 transition-all duration-200 font-medium shadow-lg shadow-accent/20"
              >
                重试
              </button>
            </div>

            <div v-else-if="videos.length === 0" class="text-center py-20">
              <div class="w-20 h-20 mx-auto rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
                <svg class="w-10 h-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300">稍后再看列表为空</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">在 B 站添加视频到稍后再看吧</p>
            </div>

            <div v-else-if="filteredVideos.length === 0" class="text-center py-20">
              <div class="w-20 h-20 mx-auto rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
                <svg class="w-10 h-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
              <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300">没有匹配的视频</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">尝试调整筛选条件</p>
            </div>

            <div v-else class="grid gap-4 mx-auto stagger-children" style="grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));">
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
