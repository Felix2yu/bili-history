<template>
  <div class="max-w-5xl mx-auto px-4 py-6 space-y-6">
    <!-- 标题 -->
    <h2 class="text-2xl font-bold text-center bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] bg-clip-text text-transparent">
      周报 / 月报
    </h2>

    <!-- Tab 切换 -->
    <div class="flex justify-center">
      <div class="inline-flex rounded-xl bg-gray-100 dark:bg-gray-800 p-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="px-5 py-2 text-sm font-medium rounded-lg transition-all duration-200"
          :class="activeTab === tab.key
            ? 'bg-white dark:bg-gray-700 text-[#fb7299] shadow-sm'
            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- 时间选择器 -->
    <div class="flex items-center justify-center gap-4">
      <button
        @click="navigateTime(-1)"
        class="glass-icon-btn"
        :disabled="loading"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <div class="text-center min-w-[200px]">
        <div v-if="activeTab === 'weekly'" class="text-lg font-semibold text-gray-800 dark:text-gray-200">
          {{ currentYear }} 第 {{ currentWeek }} 周
        </div>
        <div v-else class="text-lg font-semibold text-gray-800 dark:text-gray-200">
          {{ currentYear }}年{{ currentMonth }}月
        </div>
        <div v-if="reportData" class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
          <template v-if="activeTab === 'weekly'">
            {{ reportData.start_date }} ~ {{ reportData.end_date }}
          </template>
          <template v-else>
            共 {{ reportData.summary?.unique_days || 0 }} 天有观看记录
          </template>
        </div>
      </div>

      <button
        @click="navigateTime(1)"
        class="glass-icon-btn"
        :disabled="loading"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex justify-center py-16">
      <van-loading type="spinner" color="#fb7299" size="36">加载中...</van-loading>
    </div>

    <!-- 空状态 -->
    <div v-else-if="reportData && reportData.videos?.length === 0" class="text-center py-16">
      <div class="text-gray-400 dark:text-gray-500 text-lg mb-2">该{{ activeTab === 'weekly' ? '周' : '月' }}暂无观看记录</div>
      <div class="text-gray-400 dark:text-gray-500 text-sm">试试切换到其他时间</div>
    </div>

    <!-- 报告内容 -->
    <template v-else-if="reportData && reportData.videos?.length > 0">
      <!-- 汇总统计 -->
      <ReportSummary :summary="reportData.summary" />

      <!-- 视频列表 -->
      <div class="space-y-6">
        <div v-for="(group, date) in groupedVideos" :key="date" class="space-y-3">
          <div class="flex items-center gap-2">
            <div class="h-px flex-1 bg-gray-200 dark:bg-gray-700"></div>
            <span class="text-sm font-medium text-gray-500 dark:text-gray-400 whitespace-nowrap">
              {{ formatDateGroup(date) }}
              <span class="text-xs text-gray-400 dark:text-gray-500 ml-1">
                ({{ group.length }}个视频 · {{ formatDayDuration(date) }})
              </span>
            </span>
            <div class="h-px flex-1 bg-gray-200 dark:bg-gray-700"></div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
            <ReportVideoCard
              v-for="video in group"
              :key="`${video.bvid}-${video.view_at}`"
              :video="video"
            />
          </div>
        </div>
      </div>
    </template>

    <!-- 错误状态 -->
    <div v-else-if="error" class="text-center py-16">
      <div class="text-red-500 text-lg mb-2">加载失败</div>
      <div class="text-gray-500 text-sm">{{ error }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { getWeeklyReport, getMonthlyReport, getAvailableYears } from '~/utils/api'
import { formatDurationShort } from '~/utils/format'
import ReportSummary from '~/components/ReportSummary.vue'
import ReportVideoCard from '~/components/ReportVideoCard.vue'

const tabs = [
  { key: 'weekly', label: '周报' },
  { key: 'monthly', label: '月报' },
]

const activeTab = ref('weekly')
const currentYear = ref(new Date().getFullYear())
const currentWeek = ref(getCurrentWeek())
const currentMonth = ref(new Date().getMonth() + 1)
const reportData = ref(null)
const loading = ref(false)
const error = ref(null)

function getCurrentWeek() {
  const now = new Date()
  const jan4 = new Date(now.getFullYear(), 0, 4)
  const weekday = jan4.getDay() || 7 // Sunday = 7
  const week1Monday = new Date(jan4)
  week1Monday.setDate(jan4.getDate() - weekday + 1)
  const diffDays = Math.floor((now - week1Monday) / 86400000)
  return Math.max(1, Math.floor(diffDays / 7) + 1)
}

const groupedVideos = computed(() => {
  if (!reportData.value?.videos) return {}
  const groups = {}
  for (const video of reportData.value.videos) {
    const date = new Date(video.view_at * 1000)
    const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
    if (!groups[key]) groups[key] = []
    groups[key].push(video)
  }
  return groups
})

function formatDateGroup(dateStr) {
  const date = new Date(dateStr + 'T00:00:00')
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  const month = date.getMonth() + 1
  const day = date.getDate()
  const weekday = weekdays[date.getDay()]
  return `${month}月${day}日 ${weekday}`
}

function formatDayDuration(dateStr) {
  const group = groupedVideos.value[dateStr]
  if (!group) return ''
  const total = group.reduce((sum, v) => sum + (v.duration > 0 ? v.duration : 0), 0)
  return formatDurationShort(total)
}

async function fetchReport() {
  loading.value = true
  error.value = null
  reportData.value = null

  try {
    let response
    if (activeTab.value === 'weekly') {
      response = await getWeeklyReport(currentYear.value, currentWeek.value)
    } else {
      response = await getMonthlyReport(currentYear.value, currentMonth.value)
    }

    if (response.data?.status === 'success') {
      reportData.value = response.data.data
    } else {
      error.value = response.data?.message || '获取数据失败'
    }
  } catch (err) {
    error.value = err.message || '网络错误'
  } finally {
    loading.value = false
  }
}

function navigateTime(direction) {
  if (activeTab.value === 'weekly') {
    currentWeek.value += direction
    if (currentWeek.value > 53) {
      currentWeek.value = 1
      currentYear.value++
    } else if (currentWeek.value < 1) {
      currentWeek.value = 53
      currentYear.value--
    }
  } else {
    currentMonth.value += direction
    if (currentMonth.value > 12) {
      currentMonth.value = 1
      currentYear.value++
    } else if (currentMonth.value < 1) {
      currentMonth.value = 12
      currentYear.value--
    }
  }
}

watch([activeTab, currentYear, currentWeek, currentMonth], () => {
  fetchReport()
})

onMounted(() => {
  fetchReport()
})
</script>

<style scoped>
.glass-icon-btn {
  @apply w-10 h-10 rounded-full flex items-center justify-center
    bg-white/80 dark:bg-gray-800/80
    border border-gray-200/50 dark:border-gray-700/50
    text-gray-600 dark:text-gray-400
    hover:bg-white dark:hover:bg-gray-700
    hover:text-[#fb7299]
    transition-all duration-200
    disabled:opacity-50 disabled:cursor-not-allowed;
}
</style>
