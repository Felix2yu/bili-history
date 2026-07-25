<template>
  <div class="min-h-screen bg-gray-50/30 dark:bg-gray-900 pb-20 md:pb-0">
    <div class="py-6">
      <div class="mx-auto sm:px-2 lg:px-8">
        <div class="glass-card overflow-hidden">
          <div class="border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between">
            <h2 class="text-lg font-medium text-gray-900 dark:text-gray-100">我的点赞</h2>
            <div class="flex items-center space-x-3">
              <span v-if="syncing" class="text-xs text-[#fb7299] flex items-center">
                <svg class="animate-spin -ml-1 mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                同步中...
              </span>
              <button
                @click="syncFromBilibili"
                :disabled="syncing"
                class="px-3 py-1 text-xs rounded-md bg-[#fb7299]/10 text-[#fb7299] hover:bg-[#fb7299]/20 transition-colors disabled:opacity-50"
              >
                同步
              </button>
              <span class="text-sm text-gray-500 dark:text-gray-400">
                共 {{ totalCount }} 个视频
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
                @click="fetchLocal"
                class="mt-4 px-4 py-2 bg-[#fb7299] text-white rounded-md hover:bg-[#fb7299]/90 transition-colors"
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

            <div v-else class="grid gap-4 px-4 mx-auto" style="grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));">
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
import { getLikeList, getLikeLocal, syncLikes, batchCheckFavoriteStatus } from '~/utils/api'
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
      const total = response.data.data?.total || 0
      showNotify({ type: 'success', message: `同步完成：共 ${total} 条点赞记录` })
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
