<template>
  <div class="px-2 sm:px-4 lg:px-6 py-6 space-y-6 max-w-[1600px] mx-auto">
    <!-- 标题 -->
    <h2 class="text-2xl font-bold text-center text-accent">
      数据概览
    </h2>

    <!-- Tab 切换 -->
    <div class="flex justify-center">
      <div class="inline-flex rounded-xl bg-gray-100 dark:bg-gray-800 p-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="switchTab(tab.key)"
          class="px-5 py-2 text-sm font-medium rounded-lg transition-all duration-200"
          :class="activeTab === tab.key
            ? 'bg-white dark:bg-gray-700 text-accent shadow-sm'
            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- 选择弹窗 -->
    <div v-if="showSelector" class="fixed inset-0 z-50 flex items-center justify-center">
      <!-- 遮罩层 -->
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="handleSelectorClose"></div>
      <!-- 弹窗内容 -->
      <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl p-8 max-w-sm w-full mx-4 space-y-6">
        <h3 class="text-xl font-bold text-center text-accent">
          选择概览类型
        </h3>
        <div class="space-y-3">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key; showSelector = false"
            class="w-full flex items-center gap-4 p-4 rounded-xl border-2 transition-all duration-200 hover:border-accent/50 hover:bg-gray-50 dark:hover:bg-gray-700/50 border-gray-200 dark:border-gray-700"
          >
            <div class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 bg-gray-100 dark:bg-gray-700 text-gray-500">
              <svg v-if="tab.key === 'yearly'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <svg v-else-if="tab.key === 'monthly'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              <svg v-else class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <div class="flex-1 text-left">
              <div class="font-medium text-gray-800 dark:text-gray-200">{{ tab.label }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ tab.desc }}</div>
            </div>
          </button>
        </div>
      </div>
    </div>

    <!-- 年度标签：直接渲染 AnimatedAnalytics -->
    <template v-if="activeTab === 'yearly'">
      <AnimatedAnalytics />
    </template>

    <!-- 月度 / 周度标签 -->
    <template v-else>
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
        <van-loading type="spinner" color="var(--accent)" size="36">加载中...</van-loading>
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

            <div class="grid gap-3" style="grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));">
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
    </template>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getWeeklyReport, getMonthlyReport } from '~/utils/api'
import { formatDurationShort } from '~/utils/format'
import ReportSummary from '~/components/ReportSummary.vue'
import ReportVideoCard from '~/components/ReportVideoCard.vue'
import AnimatedAnalytics from '~/components/page/AnimatedAnalytics.vue'

const route = useRoute()

const tabs = [
  { key: 'yearly', label: '年度', desc: '全年观看数据总览与分析' },
  { key: 'monthly', label: '月度', desc: '按月查看观看记录与统计' },
  { key: 'weekly', label: '周度', desc: '按周查看观看记录与统计' },
]

// 如果有 query 参数，直接跳过选择弹窗
const hasInitialTab = !!(route.query.tab)
const showSelector = ref(!hasInitialTab)
const activeTab = ref(hasInitialTab ? (route.query.tab === 'monthly' ? 'monthly' : route.query.tab === 'weekly' ? 'weekly' : 'yearly') : '')
const currentYear = ref(new Date().getFullYear())
const currentWeek = ref(getCurrentWeek())
const currentMonth = ref(new Date().getMonth() + 1)
const reportData = ref(null)
const loading = ref(false)
const error = ref(null)

// ISO 8601 周计算：周一开始，周日结束
function getCurrentWeek() {
  const now = new Date()
  const jan1 = new Date(now.getFullYear(), 0, 1)
  // 找到1月1日之后的第一个周一
  const dayOfWeek = jan1.getDay() || 7 // 转换为 1=周一 ... 7=周日
  const firstMonday = new Date(jan1)
  if (dayOfWeek > 1) {
    firstMonday.setDate(jan1.getDate() + (8 - dayOfWeek))
  }
  // 如果当前日期在第一个周一之前，属于上一年的最后一周
  if (now < firstMonday) {
    // 计算上一年最后一周
    const prevJan1 = new Date(now.getFullYear() - 1, 0, 1)
    const prevDayOfWeek = prevJan1.getDay() || 7
    const prevFirstMonday = new Date(prevJan1)
    if (prevDayOfWeek > 1) {
      prevFirstMonday.setDate(prevJan1.getDate() + (8 - prevDayOfWeek))
    }
    const dec31 = new Date(now.getFullYear() - 1, 11, 31)
    const diffDays = Math.floor((dec31 - prevFirstMonday) / 86400000)
    return Math.max(1, Math.floor(diffDays / 7) + 1)
  }
  const diffDays = Math.floor((now - firstMonday) / 86400000)
  return Math.max(1, Math.floor(diffDays / 7) + 1)
}

function switchTab(key) {
  if (key === activeTab.value) return
  activeTab.value = key
  reportData.value = null
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
  if (activeTab.value === 'yearly' || !activeTab.value) return
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
@reference "../../assets/css/main.css";
.glass-icon-btn {
  @apply w-10 h-10 rounded-full flex items-center justify-center
    bg-white/80 dark:bg-gray-800/80
    border border-gray-200/50 dark:border-gray-700/50
    text-gray-600 dark:text-gray-400
    hover:bg-white dark:hover:bg-gray-700
    hover:text-accent
    transition-all duration-200
    disabled:opacity-50 disabled:cursor-not-allowed;
}
</style>
