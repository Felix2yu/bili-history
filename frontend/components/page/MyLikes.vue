<template>
  <div class="min-h-screen bg-gray-50/30 dark:bg-gray-900 pb-20 md:pb-0">
    <div class="py-6">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
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

          <div v-if="videos.length > 0" class="border-b border-gray-200 dark:border-gray-700 px-4 py-3">
            <div class="flex flex-wrap items-center gap-3">
              <div class="flex items-center space-x-2">
                <span class="text-xs text-gray-500 dark:text-gray-400">排序:</span>
                <button
                  v-for="opt in sortOptions"
                  :key="opt.key"
                  @click="toggleSort(opt.key)"
                  class="px-2 py-1 text-xs rounded-md transition-colors"
                  :class="sortKey === opt.key
                    ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
                >
                  {{ opt.label }}
                  <span v-if="sortKey === opt.key" class="ml-0.5">{{ sortOrder === 'desc' ? '↓' : '↑' }}</span>
                </button>
              </div>

              <div class="w-px h-4 bg-gray-200 dark:bg-gray-700"></div>

              <div class="flex items-center space-x-2 flex-wrap">
                <span class="text-xs text-gray-500 dark:text-gray-400">分区:</span>
                <button
                  @click="selectedCategory = ''"
                  class="px-2 py-1 text-xs rounded-md transition-colors"
                  :class="selectedCategory === ''
                    ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
                >
                  全部
                </button>
                <button
                  v-for="cat in topCategories"
                  :key="cat.tname"
                  @click="selectedCategory = cat.tname"
                  class="px-2 py-1 text-xs rounded-md transition-colors"
                  :class="selectedCategory === cat.tname
                    ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
                >
                  {{ cat.tname }} ({{ cat.count }})
                </button>
                <div v-if="topCategories.length < allCategories.length" class="relative" ref="catDropdownRef">
                  <button
                    @click.stop="showCatDropdown = !showCatDropdown"
                    class="px-2 py-1 text-xs rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                  >
                    更多...
                  </button>
                  <div
                    v-if="showCatDropdown"
                    class="absolute top-full left-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 p-2 max-h-60 overflow-y-auto min-w-[180px]"
                  >
                    <button
                      v-for="cat in restCategories"
                      :key="cat.tname"
                      @click="selectedCategory = cat.tname; showCatDropdown = false"
                      class="w-full text-left px-2 py-1 text-xs rounded transition-colors"
                      :class="selectedCategory === cat.tname
                        ? 'bg-[#fb7299]/10 text-[#fb7299]'
                        : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
                    >
                      {{ cat.tname }} ({{ cat.count }})
                    </button>
                  </div>
                </div>
              </div>

              <div class="w-px h-4 bg-gray-200 dark:border-gray-700"></div>

              <div class="flex items-center space-x-2 flex-wrap">
                <span class="text-xs text-gray-500 dark:text-gray-400">UP主:</span>
                <button
                  @click="selectedOwner = ''"
                  class="px-2 py-1 text-xs rounded-md transition-colors"
                  :class="selectedOwner === ''
                    ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
                >
                  全部
                </button>
                <button
                  v-for="owner in topOwners"
                  :key="owner.name"
                  @click="selectedOwner = owner.name"
                  class="px-2 py-1 text-xs rounded-md transition-colors max-w-[120px] truncate"
                  :class="selectedOwner === owner.name
                    ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
                  :title="owner.name"
                >
                  {{ owner.name }} ({{ owner.count }})
                </button>
                <div v-if="topOwners.length < allOwners.length" class="relative" ref="ownerDropdownRef">
                  <button
                    @click.stop="showOwnerDropdown = !showOwnerDropdown"
                    class="px-2 py-1 text-xs rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                  >
                    更多...
                  </button>
                  <div
                    v-if="showOwnerDropdown"
                    class="absolute top-full left-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 p-2 max-h-60 overflow-y-auto min-w-[180px]"
                  >
                    <button
                      v-for="owner in restOwners"
                      :key="owner.name"
                      @click="selectedOwner = owner.name; showOwnerDropdown = false"
                      class="w-full text-left px-2 py-1 text-xs rounded transition-colors"
                      :class="selectedOwner === owner.name
                        ? 'bg-[#fb7299]/10 text-[#fb7299]'
                        : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
                    >
                      {{ owner.name }} ({{ owner.count }})
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

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

            <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
              <div
                v-for="video in videos"
                :key="video.bvid"
                class="bg-white/50 dark:bg-gray-800/50 rounded-md overflow-hidden border border-gray-200/50 dark:border-gray-700/50 hover:border-[#fb7299] hover:shadow-sm transition-all duration-200 relative group"
              >
                <div class="relative pb-[56.25%] overflow-hidden cursor-pointer group" @click="openVideo(video)">
                  <img
                    :src="normalizeImageUrl(video.pic)"
                    :alt="video.title"
                    class="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                    loading="lazy"
                    onerror="this.src='https://i0.hdslb.com/bfs/archive/c9e72655b7c9c9c68a30d3275313c501e68427d1.jpg'"
                  />
                  <div class="absolute bottom-1 right-1 bg-black/60 px-1 py-0.5 rounded text-white text-[10px]">
                    {{ formatDuration(video.duration) }}
                  </div>
                  <div v-if="video.tname" class="absolute top-1 left-1 bg-[#fb7299]/80 px-1 py-0.5 rounded text-white text-[10px]">
                    {{ video.tname }}
                  </div>
                </div>

                <div class="p-2 flex flex-col space-y-1">
                  <div class="line-clamp-2 text-xs text-gray-900 dark:text-gray-100 font-medium cursor-pointer" @click="openVideo(video)">
                    {{ video.title }}
                  </div>
                  <div class="flex items-center space-x-1">
                    <img
                      :src="normalizeImageUrl(video.owner_face)"
                      :alt="video.owner_name"
                      class="w-3.5 h-3.5 rounded-full object-cover"
                      loading="lazy"
                      onerror="this.src='https://i1.hdslb.com/bfs/face/1b6f746be0d0c8324e01e618c5e85e113a8b38be.jpg'"
                    />
                    <span class="text-[10px] text-gray-600 dark:text-gray-400 truncate">{{ video.owner_name }}</span>
                  </div>
                  <div class="flex justify-between items-center text-[10px] text-gray-500">
                    <span>{{ formatViews(video.view) }} 次观看</span>
                    <span>{{ formatTime(video.pubdate) }}</span>
                  </div>
                </div>
              </div>
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
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useAsyncData } from '#imports'
import { showNotify } from 'vant'
import 'vant/es/notify/style'
import { getLikeList, getLikeLocal } from '~/utils/api'
import { normalizeImageUrl } from '~/utils/imageUrl.js'

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
const showOwnerDropdown = ref(false)
const showCatDropdown = ref(false)
const ownerDropdownRef = ref(null)
const catDropdownRef = ref(null)

const allOwners = ref([])
const allCategories = ref([])

const sortOptions = [
  { key: 'pubdate', label: '发布时间' },
  { key: 'fetch_time', label: '同步时间' },
  { key: 'duration', label: '时长' },
  { key: 'view', label: '播放量' },
]

const topOwners = computed(() => allOwners.value.slice(0, 10))
const restOwners = computed(() => allOwners.value.slice(10))

const topCategories = computed(() => allCategories.value.slice(0, 10))
const restCategories = computed(() => allCategories.value.slice(10))

const totalPages = computed(() => Math.ceil(totalCount.value / pageSize.value))

function toggleSort(key) {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
  } else {
    sortKey.value = key
    sortOrder.value = key === 'owner_name' ? 'asc' : 'desc'
  }
  currentPage.value = 1
  fetchLocal()
}

function goToPage(page) {
  currentPage.value = page
  fetchLocal()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function handleClickOutside(e) {
  if (ownerDropdownRef.value && !ownerDropdownRef.value.contains(e.target)) {
    showOwnerDropdown.value = false
  }
  if (catDropdownRef.value && !catDropdownRef.value.contains(e.target)) {
    showCatDropdown.value = false
  }
}

// SSR: 初始数据在服务端获取
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

// 从 SSR 数据初始化组件状态
if (initialData.value) {
  videos.value = initialData.value.videos
  totalCount.value = initialData.value.totalCount
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside)
  if (videos.value.length === 0) {
    await fetchLocal()
  }
  loadFilterOptions()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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
    const response = await getLikeList()
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: `同步完成：新增 ${response.data.data.new}，更新 ${response.data.data.updated}` })
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

function openVideo(video) {
  if (video.link) {
    window.open(video.link, '_blank')
  }
}

function formatDuration(seconds) {
  if (!seconds) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

function formatViews(count) {
  if (!count) return '0'
  if (count >= 10000) return (count / 10000).toFixed(1) + '万'
  return count.toString()
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}
</script>
