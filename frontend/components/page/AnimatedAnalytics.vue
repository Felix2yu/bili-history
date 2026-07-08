<template>
  <div>
    <analytics-layout>
      <!-- 导航栏 -->
      <div class="mb-6">
        <div class="bg-white/5 backdrop-blur-md border-b border-white/10 dark:bg-black/5 dark:border-gray-800/50 rounded-t-xl">
          <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between items-center h-14">
              <h1 class="text-2xl font-bold bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] bg-clip-text text-transparent">
                {{ selectedYear }}年度回顾
              </h1>
              <div class="flex items-center space-x-4">
                <select
                  v-model="selectedYear"
                  class="w-24 bg-white/10 backdrop-blur text-gray-800 dark:text-white border border-white/20 dark:border-gray-700 rounded-lg px-3 py-1 focus:ring-2 focus:ring-[#fb7299] focus:border-transparent transition-colors duration-200"
                  :disabled="loading"
                >
                  <option v-for="year in availableYears" :key="year" :value="year">
                    {{ year }}年
                  </option>
                </select>

                <!-- 强制刷新按钮 -->
                <button
                  @click="handleForceRefresh"
                  :disabled="loading"
                  class="inline-flex items-center text-gray-600 dark:text-gray-300 hover:text-[#fb7299] dark:hover:text-[#fb7299] disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
                >
                  <svg
                    class="w-5 h-5"
                    :class="{'animate-spin': loading}"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <circle
                      v-if="loading"
                      class="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      stroke-width="4"
                    ></circle>
                    <path
                      v-if="loading"
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                    <path
                      v-if="!loading"
                      stroke="currentColor"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                    ></path>
                  </svg>
                  <span class="ml-2">{{ loading ? '加载中' : '强制刷新' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading"
        class="fixed inset-0 flex items-center justify-center z-50 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm"
      >
        <div class="text-center">
          <svg
            class="w-12 h-12 mx-auto mb-4 animate-spin text-[#fb7299]"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <div class="space-y-2">
            <p class="text-lg font-medium text-gray-800 dark:text-gray-200">正在分析{{ selectedYear }}年的观看数据</p>
            <p class="text-sm text-gray-600 dark:text-gray-400">
              {{ loading && viewingData === null ? '首次加载数据可能需要30秒到1分钟，具体加载时间取决于数据量' : '正在从缓存加载数据，预计3-5秒' }}
            </p>
          </div>
        </div>
      </div>

      <!-- 所有分析页面 -->
      <div class="space-y-10">
        <OverviewPage :viewing-data="monthlyStatsData" />
        <TimeAnalysisPage :viewing-data="timeSlotsData" :selected-year="selectedYear" />
        <TimeDistributionPage :viewing-data="weeklyStatsData" />
        <MonthlyPage :viewing-data="monthlyStatsData" />
        <StreakPage :viewing-data="continuityData" :selected-year="selectedYear" />
        <RewatchPage :viewing-data="watchCountsData" />
        <OverallCompletionPage :viewing-data="completionRatesData" />
        <AuthorCompletionPage :viewing-data="authorCompletionData" />
        <TagsPage :viewing-data="tagAnalysisData" />
        <DurationAnalysisPage :viewing-data="durationAnalysisData" />
        <LikesAnalysisPage :viewing-data="likesAnalysisData" />
        <FavoritesAnalysisPage :viewing-data="favoritesAnalysisData" />
        <WatchLaterAnalysisPage :viewing-data="watchLaterAnalysisData" />
      </div>
    </analytics-layout>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { getViewingMonthlyStats, getViewingWeeklyStats, getViewingTimeSlots, getViewingContinuity, getViewingWatchCounts, getViewingCompletionRates, getViewingAuthorCompletion, getViewingTagAnalysis, getViewingDurationAnalysis, getLikesAnalysis, getFavoritesAnalysis, getWatchLaterAnalysis } from '~/utils/api'
import OverviewPage from '../analytics/pages/OverviewPage.vue'
import StreakPage from '../analytics/pages/StreakPage.vue'
import TimeAnalysisPage from '../analytics/pages/TimeAnalysisPage.vue'
import RewatchPage from '../analytics/pages/RewatchPage.vue'
import OverallCompletionPage from '../analytics/pages/OverallCompletionPage.vue'
import AuthorCompletionPage from '../analytics/pages/AuthorCompletionPage.vue'
import TagsPage from '../analytics/pages/TagsPage.vue'
import TimeDistributionPage from '../analytics/pages/TimeDistributionPage.vue'
import MonthlyPage from '../analytics/pages/MonthlyPage.vue'
import DurationAnalysisPage from '../analytics/pages/DurationAnalysisPage.vue'
import LikesAnalysisPage from '../analytics/pages/LikesAnalysisPage.vue'
import FavoritesAnalysisPage from '../analytics/pages/FavoritesAnalysisPage.vue'
import WatchLaterAnalysisPage from '../analytics/pages/WatchLaterAnalysisPage.vue'
import AnalyticsLayout from '../analytics/layout/AnalyticsLayout.vue'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent
} from 'echarts/components'
import { use } from 'echarts/core'
import 'echarts-wordcloud'

// 注册必要的组件
use([
  CanvasRenderer,
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent
])

// 状态
const selectedYear = ref(new Date().getFullYear())
const availableYears = ref([])
const loading = ref(false)
const monthlyStatsData = ref(null)
const weeklyStatsData = ref(null)
const timeSlotsData = ref(null)
const continuityData = ref(null)
const watchCountsData = ref(null)
const completionRatesData = ref(null)
const authorCompletionData = ref(null)
const tagAnalysisData = ref(null)
const durationAnalysisData = ref(null)
const viewingData = ref(null)
const likesAnalysisData = ref(null)
const favoritesAnalysisData = ref(null)
const watchLaterAnalysisData = ref(null)

// 加载所有数据
const loadAllData = async (forceRefresh = false) => {
  if (loading.value) return
  loading.value = true

  try {
    const promises = []

    // 获取可用年份（必须先执行）
    if (!availableYears.value.length || forceRefresh) {
      const yearResponse = await getViewingMonthlyStats(selectedYear.value, true)
      if (yearResponse.data.status === 'success' && yearResponse.data.data.available_years) {
        availableYears.value = yearResponse.data.data.available_years
        if (!availableYears.value.includes(selectedYear.value)) {
          selectedYear.value = availableYears.value[0]
        }
      }
    }

    // 并行加载所有分析数据
    if (!monthlyStatsData.value || forceRefresh) {
      promises.push(getViewingMonthlyStats(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') monthlyStatsData.value = res.data.data
      }))
    }

    if (!timeSlotsData.value || forceRefresh) {
      promises.push(getViewingTimeSlots(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') timeSlotsData.value = res.data.data
      }))
    }

    if (!weeklyStatsData.value || forceRefresh) {
      promises.push(getViewingWeeklyStats(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') weeklyStatsData.value = res.data.data
      }))
    }

    if (!continuityData.value || forceRefresh) {
      promises.push(getViewingContinuity(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') continuityData.value = res.data.data
      }))
    }

    if (!watchCountsData.value || forceRefresh) {
      promises.push(getViewingWatchCounts(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') watchCountsData.value = res.data.data
      }))
    }

    if (!completionRatesData.value || forceRefresh) {
      promises.push(getViewingCompletionRates(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') completionRatesData.value = res.data.data
      }))
    }

    if (!authorCompletionData.value || forceRefresh) {
      promises.push(getViewingAuthorCompletion(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') authorCompletionData.value = res.data.data
      }))
    }

    if (!tagAnalysisData.value || forceRefresh) {
      promises.push(getViewingTagAnalysis(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') tagAnalysisData.value = res.data.data
      }))
    }

    if (!durationAnalysisData.value || forceRefresh) {
      promises.push(getViewingDurationAnalysis(selectedYear.value, true).then(res => {
        if (res.data.status === 'success') durationAnalysisData.value = res.data.data
      }))
    }

    if (!likesAnalysisData.value || forceRefresh) {
      promises.push(getLikesAnalysis().then(res => {
        if (res.data.status === 'success') likesAnalysisData.value = res.data.data
      }))
    }

    if (!favoritesAnalysisData.value || forceRefresh) {
      promises.push(getFavoritesAnalysis().then(res => {
        if (res.data.status === 'success') favoritesAnalysisData.value = res.data.data
      }))
    }

    if (!watchLaterAnalysisData.value || forceRefresh) {
      promises.push(getWatchLaterAnalysis().then(res => {
        if (res.data.status === 'success') watchLaterAnalysisData.value = res.data.data
      }))
    }

    await Promise.all(promises)
  } catch (error) {
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 强制刷新
const handleForceRefresh = async () => {
  if (loading.value) return
  // 清空所有数据
  monthlyStatsData.value = null
  weeklyStatsData.value = null
  timeSlotsData.value = null
  continuityData.value = null
  watchCountsData.value = null
  completionRatesData.value = null
  authorCompletionData.value = null
  tagAnalysisData.value = null
  durationAnalysisData.value = null
  viewingData.value = null
  likesAnalysisData.value = null
  favoritesAnalysisData.value = null
  watchLaterAnalysisData.value = null
  await loadAllData(true)
}

// 监听年份变化
watch(selectedYear, async (newYear) => {
  if (newYear) {
    await loadAllData(true)
  }
})

// 初始化加载
onMounted(async () => {
  await loadAllData()
})
</script>

<style>
/* 过渡效果 */
select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%23fb7299' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 0.5rem center;
  background-repeat: no-repeat;
  background-size: 1.5em 1.5em;
  padding-right: 2.5rem;
}

select:focus {
  outline: none;
  box-shadow: 0 0 0 2px rgba(251, 114, 153, 0.2);
}
</style>
