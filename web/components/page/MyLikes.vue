<template>
  <div class="min-h-screen pb-20 md:pb-0 relative overflow-hidden">
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute -top-32 -right-20 w-96 h-96 bg-accent/5 rounded-full blur-3xl"></div>
      <div class="absolute top-1/4 -left-32 w-80 h-80 bg-accent/3 rounded-full blur-3xl"></div>
    </div>

    <div class="relative py-6">
      <div class="max-w-[1800px] mx-auto sm:px-2 lg:px-8">
        <div class="mb-6">
          <div class="glass-card overflow-hidden">
            <div class="px-5 md:px-6 py-4 md:py-5">
              <div class="flex items-center justify-between flex-wrap gap-4">
                <div class="flex items-center gap-3 md:gap-4">
                  <div class="w-11 h-11 md:w-12 md:h-12 rounded-xl bg-accent flex items-center justify-center shadow-md shadow-accent/20">
                    <svg class="w-5 h-5 md:w-6 md:h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                    </svg>
                  </div>
                  <div>
                    <h1 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">我的点赞</h1>
                    <p class="text-xs md:text-sm text-gray-500 dark:text-gray-400 mt-0.5">记录你在 B 站点赞过的所有精彩视频</p>
                  </div>
                </div>
                <div class="flex items-center gap-3 md:gap-4 flex-wrap">
                  <div class="flex items-center gap-2.5">
                    <div class="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center">
                      <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    </div>
                    <div class="flex flex-col">
                      <span class="text-[0.625rem] text-gray-500 dark:text-gray-400 leading-none">视频总数</span>
                      <span class="text-base md:text-lg font-bold text-gray-900 dark:text-gray-100 leading-tight mt-0.5">{{ totalCount }}</span>
                    </div>
                  </div>
                  <div class="w-px h-8 bg-gray-200 dark:bg-gray-700 hidden md:block"></div>
                  <div class="flex items-center gap-2.5 hidden md:flex">
                    <div class="w-8 h-8 rounded-lg flex items-center justify-center" :class="syncing ? 'bg-accent/10' : 'bg-emerald-500/10 dark:bg-emerald-500/20'">
                      <svg v-if="syncing" class="w-4 h-4 text-accent animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      <svg v-else class="w-4 h-4 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div class="flex flex-col">
                      <span class="text-[0.625rem] text-gray-500 dark:text-gray-400 leading-none">同步状态</span>
                      <span class="text-sm font-semibold leading-tight mt-0.5" :class="syncing ? 'text-accent' : 'text-emerald-500'">{{ syncing ? '同步中...' : '已同步' }}</span>
                    </div>
                  </div>
                  <button
                    @click="syncFromBilibili"
                    :disabled="syncing"
                    class="px-4 py-2 md:px-5 md:py-2.5 text-xs md:text-sm font-medium rounded-xl bg-accent text-white hover:bg-accent/90 transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-md shadow-accent/20 flex items-center gap-1.5 md:gap-2"
                  >
                    <svg class="w-3.5 h-3.5 md:w-4 md:h-4" :class="{'animate-spin': syncing}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    <span>{{ syncing ? '同步中...' : '立即同步' }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="glass-card overflow-hidden">
          <div class="border-b border-glass-border px-5 py-3.5 flex items-center justify-between bg-white/30 dark:bg-white/5">
            <div class="flex items-center gap-2.5">
              <div class="w-1.5 h-5 rounded-full bg-accent"></div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">点赞视频列表</h2>
            </div>
            <div class="flex items-center gap-2">
              <span v-if="syncing" class="text-xs text-accent flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-accent/10">
                <svg class="animate-spin h-3 w-3" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                正在同步数据
              </span>
              <span class="text-xs text-gray-500 dark:text-gray-400 px-2.5 py-1 rounded-full bg-gray-100 dark:bg-gray-800">
                当前显示 {{ videos.length }} / {{ totalCount }} 个视频
              </span>
            </div>
          </div>

          <VideoFilterBar
            v-if="videos.length > 0"
            v-model:sort-key="sortKey"
            v-model:sort-order="sortOrder"
            v-model:selected-category="selectedCategory"
            v-model:selected-owner="selectedOwner"
            :sort-options="sortOptions"
            :all-categories="allCategories"
            :all-owners="allOwners"
            @sort-change="handleSortChange"
            @category-change="handleFilterChange"
            @owner-change="handleFilterChange"
          />

          <div class="p-5">
            <div v-if="loading" class="flex justify-center py-20">
              <div class="inline-flex items-center px-4 py-2 bg-white dark:bg-gray-800 rounded-md shadow text-gray-900 dark:text-gray-100">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-accent" fill="none" viewBox="0 0 24 24">
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
                @click="fetchLocal"
                class="mt-4 px-4 py-2 bg-accent text-white rounded-md hover:bg-accent/90 transition-colors"
              >
                重试
              </button>
            </div>

            <div v-else-if="videos.length === 0" class="text-center py-20">
              <svg class="w-16 h-16 mx-auto text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
              </svg>
              <p class="mt-4 text-gray-500 dark:text-gray-400">暂无点赞视频</p>
            </div>

            <div v-else class="grid gap-3 px-4" style="grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));">
              <VideoGridCard
                v-for="video in videos"
                :key="video.bvid"
                :video="video"
                :time-field="'pubdate'"
                @click="openVideo(video)"
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
                </template>
              </VideoGridCard>
            </div>

            <div v-if="totalPages > 1" class="mt-6 flex justify-center">
              <Pagination
                :current-page="currentPage"
                :total-pages="totalPages"
                :page-size="pageSize"
                @page-change="goToPage"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
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
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAsyncData } from '#imports'
import { showNotify } from 'vant'
import 'vant/es/notify/style'
import { getLikeList, getLikeLocal, syncLikes, getTaskStatus, batchCheckFavoriteStatus } from '~/utils/api'
import VideoGridCard from '../VideoGridCard.vue'
import VideoFilterBar from '../VideoFilterBar.vue'
import Pagination from '../Pagination.vue'
import DownloadDialog from '../DownloadDialog.vue'
import FavoriteDialog from '../FavoriteDialog.vue'

const loading = ref(false)
const syncing = ref(false)
const error = ref('')
const videos = ref([])
const totalCount = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)

const sortKey = ref('fetch_time')
const sortOrder = ref('desc')
const selectedOwner = ref('')
const selectedCategory = ref('')

const allOwners = ref([])
const allCategories = ref([])

const showDownloadDialog = ref(false)
const showFavoriteDialog = ref(false)
const currentVideo = ref({})
const favoriteStatus = ref({})

const sortOptions = [
  { key: 'pubdate', label: '发布时间' },
  { key: 'fetch_time', label: '同步时间' },
  { key: 'duration', label: '时长' },
  { key: 'view', label: '播放量' },
]

const totalPages = computed(() => Math.ceil(totalCount.value / pageSize.value))

function handleSortChange() {
  currentPage.value = 1
  fetchLocal()
}

function handleFilterChange() {
  currentPage.value = 1
  fetchLocal()
}

function goToPage(page) {
  currentPage.value = page
  fetchLocal()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const { data: initialData } = await useAsyncData('likes-initial', async () => {
  try {
    const response = await getLikeLocal({
      page: 1,
      size: 50,
      sort: 'fetch_time',
      order: 'desc'
    })
    if (response.data.status === 'success') {
      return {
        videos: response.data.data.list || [],
        totalCount: response.data.data.total || 0
      }
    }
    return { videos: [], totalCount: 0 }
  } catch (error) {
    console.error('SSR 获取点赞列表失败:', error)
    return { videos: [], totalCount: 0 }
  }
})

if (initialData.value) {
  videos.value = initialData.value.videos
  totalCount.value = initialData.value.totalCount
}

onMounted(async () => {
  if (videos.value.length === 0) {
    await fetchLocal()
  }
  loadFilterOptions()
})

async function fetchLocal() {
  loading.value = true
  error.value = ''
  try {
    const response = await getLikeLocal({
      page: currentPage.value,
      size: pageSize.value,
      sort: sortKey.value,
      order: sortOrder.value,
      category: selectedCategory.value || undefined,
      owner: selectedOwner.value || undefined,
    })
    if (response.data.status === 'success') {
      videos.value = response.data.data.list || []
      totalCount.value = response.data.data.total || 0
    } else {
      error.value = response.data.message || '获取点赞数据失败'
    }
  } catch (e) {
    error.value = '请求失败: ' + (e.message || '未知错误')
  } finally {
    loading.value = false
  }
  checkFavoriteStatus()
}

async function loadFilterOptions() {
  try {
    const response = await getLikeLocal({ page: 1, size: 500, sort: 'pubdate', order: 'desc' })
    if (response.data.status === 'success') {
      const list = response.data.data.list || []
      const ownerMap = {}
      const catMap = {}
      for (const v of list) {
        const owner = v.owner_name || '未知'
        ownerMap[owner] = (ownerMap[owner] || 0) + 1
        const cat = v.tname || '未知分区'
        catMap[cat] = (catMap[cat] || 0) + 1
      }
      allOwners.value = Object.entries(ownerMap)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count)
      allCategories.value = Object.entries(catMap)
        .map(([tname, count]) => ({ tname, count }))
        .sort((a, b) => b.count - a.count)
    }
  } catch (e) {
    console.warn('加载筛选选项失败:', e)
  }
}

async function syncFromBilibili() {
  syncing.value = true
  try {
    const response = await syncLikes()
    if (response.data.status === 'success') {
      const data = response.data.data || {}
      if (data.task_id) {
        const result = await pollTaskStatus(data.task_id)
        if (result.status === 'success' || result.status === 'completed') {
          const extra = parseResultJson(result.result)
          const total = extra?.total || 0
          showNotify({ type: 'success', message: `同步完成：共 ${total} 条点赞记录` })
        } else {
          showNotify({ type: 'warning', message: result.error || result.message || '同步失败' })
        }
      } else {
        const total = data.total || 0
        showNotify({ type: 'success', message: `同步完成：共 ${total} 条点赞记录` })
      }
      await fetchLocal()
      loadFilterOptions()
    } else {
      showNotify({ type: 'warning', message: response.data.message || '同步失败' })
    }
  } catch (e) {
    showNotify({ type: 'danger', message: '同步失败: ' + (e.message || '未知错误') })
  } finally {
    syncing.value = false
  }
}

function parseResultJson(result) {
  if (!result) return null
  if (typeof result === 'object') return result
  try {
    return JSON.parse(result)
  } catch (e) {
    return null
  }
}

async function pollTaskStatus(taskId, maxAttempts = 60, interval = 2000) {
  for (let i = 0; i < maxAttempts; i++) {
    try {
      const response = await getTaskStatus(taskId)
      const result = response.data.data || response.data
      if (result.status === 'success' || result.status === 'failed' || result.status === 'completed') {
        return result
      }
    } catch (e) {
      console.warn('查询任务状态失败:', e)
    }
    await new Promise(resolve => setTimeout(resolve, interval))
  }
  return { status: 'timeout', message: '任务执行超时，请稍后查看结果' }
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
