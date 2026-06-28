<template>
  <div class="min-h-screen bg-gray-50/30 dark:bg-gray-900">
    <div class="py-6">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="bg-white dark:bg-gray-800 rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
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
            </div>
          </div>

          <div v-if="!loading && videos.length > 0" class="border-b border-gray-200 dark:border-gray-700 px-4 py-3">
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

              <div class="w-px h-4 bg-gray-200 dark:bg-gray-700"></div>

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
              <div
                v-for="video in filteredVideos"
                :key="video.bvid"
                class="bg-white/50 dark:bg-gray-800/50 rounded-md overflow-hidden border border-gray-200/50 dark:border-gray-700/50 hover:border-[#fb7299] hover:shadow-sm transition-all duration-200 relative group"
              >
                <div class="relative pb-[56.25%] overflow-hidden cursor-pointer group" @click="openVideo(video)">
                  <img
                    :src="video.pic"
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
                  <div class="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      @click.stop="removeVideo(video)"
                      class="bg-red-500/80 hover:bg-red-500 text-white rounded-full p-1"
                      title="从稍后再看中移除"
                    >
                      <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                </div>

                <div class="p-2 flex flex-col space-y-1">
                  <div class="line-clamp-2 text-xs text-gray-900 dark:text-gray-100 font-medium cursor-pointer" @click="openVideo(video)">
                    {{ video.title }}
                  </div>
                  <div class="flex items-center space-x-1">
                    <img
                      :src="video.owner_face"
                      :alt="video.owner_name"
                      class="w-3.5 h-3.5 rounded-full object-cover"
                      loading="lazy"
                      onerror="this.src='https://i1.hdslb.com/bfs/face/1b6f746be0d0c8324e01e618c5e85e113a8b38be.jpg'"
                    />
                    <span class="text-[10px] text-gray-600 dark:text-gray-400 truncate">{{ video.owner_name }}</span>
                  </div>
                  <div class="flex justify-between items-center text-[10px] text-gray-500">
                    <span>{{ formatViews(video.view) }} 次观看</span>
                    <span>{{ formatTime(video.add_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { showNotify } from 'vant'
import 'vant/es/notify/style'
import { getWatchLaterList, removeFromWatchLater, getWatchLaterLocal } from '~/utils/api'

const loading = ref(false)
const syncing = ref(false)
const error = ref('')
const videos = ref([])

const sortKey = ref('add_at')
const sortOrder = ref('desc')
const selectedOwner = ref('')
const selectedCategory = ref('')
const showOwnerDropdown = ref(false)
const showCatDropdown = ref(false)
const ownerDropdownRef = ref(null)
const catDropdownRef = ref(null)

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

const topOwners = computed(() => allOwners.value.slice(0, 10))
const restOwners = computed(() => allOwners.value.slice(10))

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

const topCategories = computed(() => allCategories.value.slice(0, 10))
const restCategories = computed(() => allCategories.value.slice(10))

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

function toggleSort(key) {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
  } else {
    sortKey.value = key
    sortOrder.value = key === 'owner_name' ? 'asc' : 'desc'
  }
}

function handleClickOutside(e) {
  if (ownerDropdownRef.value && !ownerDropdownRef.value.contains(e.target)) {
    showOwnerDropdown.value = false
  }
  if (catDropdownRef.value && !catDropdownRef.value.contains(e.target)) {
    showCatDropdown.value = false
  }
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside)
  await fetchLocal()
  syncFromBilibili()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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
}

async function removeVideo(video) {
  try {
    const response = await removeFromWatchLater(video.bvid)
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: '已移除' })
      videos.value = videos.value.filter(v => v.bvid !== video.bvid)
    } else {
      showNotify({ type: 'danger', message: response.data.message || '移除失败' })
    }
  } catch (e) {
    showNotify({ type: 'danger', message: '移除失败: ' + e.message })
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
