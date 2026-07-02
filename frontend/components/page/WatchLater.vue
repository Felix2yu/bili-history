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

          <!-- 批量选择操作栏 -->
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
                class="bg-white/50 dark:bg-gray-800/50 rounded-md overflow-hidden border transition-all duration-200 relative group"
                :class="selectMode
                  ? (isSelected(video.bvid) ? 'border-[#fb7299] ring-1 ring-[#fb7299]/40' : 'border-gray-200/50 dark:border-gray-700/50')
                  : 'border-gray-200/50 dark:border-gray-700/50 hover:border-[#fb7299] hover:shadow-sm'"
              >
                <!-- 选择模式 checkbox 覆盖层 -->
                <div v-if="selectMode" class="absolute top-1.5 left-1.5 z-10">
                  <input
                    type="checkbox"
                    :checked="isSelected(video.bvid)"
                    @click.stop="toggleSelect(video.bvid)"
                    class="w-4 h-4 rounded border-gray-300 text-[#fb7299] focus:ring-[#fb7299] bg-white/90"
                  />
                </div>

                <div class="relative pb-[56.25%] overflow-hidden cursor-pointer group" @click="selectMode ? toggleSelect(video.bvid) : openVideo(video)">
                  <img
                    :data-src="getProxyImageUrl(video.pic)"
                    :alt="video.title"
                    class="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-300 bg-gray-200 dark:bg-gray-700"
                    loading="lazy"
                    onerror="this.src='https://i0.hdslb.com/bfs/archive/c9e72655b7c9c9c68a30d3275313c501e68427d1.jpg'"
                  />
                  <div class="absolute bottom-1 right-1 bg-black/60 px-1 py-0.5 rounded text-white text-[10px]">
                    {{ formatDuration(video.duration) }}
                  </div>
                  <div v-if="video.tname" class="absolute top-1 left-1 bg-[#fb7299]/80 px-1 py-0.5 rounded text-white text-[10px]" :class="selectMode ? 'ml-6' : ''">
                    {{ video.tname }}
                  </div>
                </div>

                <div class="p-2 flex flex-col space-y-1">
                  <div class="line-clamp-2 text-xs text-gray-900 dark:text-gray-100 font-medium cursor-pointer" @click="selectMode ? toggleSelect(video.bvid) : openVideo(video)">
                    {{ video.title }}
                  </div>
                  <div class="flex items-center space-x-1">
                    <img
                      :src="getProxyImageUrl(video.owner_face)"
                      :alt="video.owner_name"
                      class="w-3.5 h-3.5 rounded-full object-cover"
                      onerror="this.src='https://static.hdslb.com/images/member/noface.gif'"
                    />
                    <span class="text-[10px] text-gray-600 dark:text-gray-400 truncate">{{ video.owner_name }}</span>
                  </div>
                  <div class="flex justify-between items-center text-[10px] text-gray-500">
                    <span>{{ formatViews(video.view) }} 次观看</span>
                    <span>{{ formatTime(video.add_at) }}</span>
                  </div>
                </div>

                <!-- 单个删除按钮（非选择模式下显示） -->
                <button
                  v-if="!selectMode"
                  @click.stop="confirmDeleteOne(video)"
                  :disabled="deleting"
                  class="absolute top-1.5 right-1.5 z-10 w-6 h-6 rounded-full bg-black/60 text-white opacity-0 group-hover:opacity-100 hover:bg-red-500 transition-all flex items-center justify-center disabled:opacity-30"
                  title="从稍后再看移除"
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
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

    <!-- 操作结果提示 -->
    <div v-if="toast.show" class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 px-4 py-2 rounded-md shadow-lg text-sm text-white" :class="toast.type === 'error' ? 'bg-red-500' : 'bg-green-500'">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useAsyncData } from '#imports'
import { getWatchLaterList, getWatchLaterLocal, removeFromWatchLater, batchRemoveFromWatchLater } from '~/utils/api'
import { getProxyImageUrl } from '~/utils/imageUrl.js'

// 图片懒加载
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
const showOwnerDropdown = ref(false)
const showCatDropdown = ref(false)
const ownerDropdownRef = ref(null)
const catDropdownRef = ref(null)

// 批量选择 / 删除状态
const selectMode = ref(false)
const selectedBvids = ref(new Set())
// 确认弹窗状态
const confirmDialog = ref({ show: false, title: '', message: '', action: null })
// toast 提示
const toast = ref({ show: false, message: '', type: 'success', timer: null })

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

// ---- 批量选择 & 删除 ----
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
    // 取消当前筛选范围内的全选
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

function handleClickOutside(e) {
  if (ownerDropdownRef.value && !ownerDropdownRef.value.contains(e.target)) {
    showOwnerDropdown.value = false
  }
  if (catDropdownRef.value && !catDropdownRef.value.contains(e.target)) {
    showCatDropdown.value = false
  }
}

// SSR: 初始数据在服务端获取
const { data: initialData } = await useAsyncData('watchlater-initial', async () => {
  try {
    const response = await getWatchLaterLocal({ size: 500 })
    if (response.data.status === 'success') {
      return { videos: response.data.data.list || [] }
    }
    return { videos: [] }
  } catch (error) {
    console.error('SSR 获取稍后再看失败:', error)
    return { videos: [] }
  }
})

// 从 SSR 数据初始化组件状态
if (initialData.value?.videos?.length > 0) {
  videos.value = initialData.value.videos
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside)
  initImageObserver()
  if (videos.value.length === 0) {
    await fetchLocal()
  }
  observeImages()
  syncFromBilibili()
})

// 监听视频列表变化，重新观察图片
watch(videos, () => { observeImages() })

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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
