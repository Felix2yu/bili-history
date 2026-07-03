<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50" v-if="showModal">
    <div class="bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-100 p-6 rounded-lg shadow-lg max-w-2xl w-full max-h-[80vh] overflow-y-auto border border-gray-200 dark:border-gray-700">
      <!-- 头部标题 -->
      <div class="flex justify-between items-center mb-4">
        <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100">{{ currentTab === 'sync' ? '数据状态' : '数据完整性检查' }}</h2>
        <button @click="closeModal" class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- 导航标签 -->
      <div class="flex border-b mb-4">
        <button
          @click="currentTab = 'sync'"
          class="py-2 px-4 font-medium text-sm transition-colors duration-200"
          :class="currentTab === 'sync' ? 'text-pink-500 border-b-2 border-pink-500' : 'text-gray-600 hover:text-pink-400'"
        >
          数据状态
        </button>
        <button
          @click="currentTab = 'integrity'"
          class="py-2 px-4 font-medium text-sm transition-colors duration-200"
          :class="currentTab === 'integrity' ? 'text-pink-500 border-b-2 border-pink-500' : 'text-gray-600 hover:text-pink-400'"
        >
          数据完整性检查
        </button>
      </div>

      <!-- 数据状态面板 -->
      <div v-if="currentTab === 'sync'">
        <div class="mb-4">
          <p class="text-sm text-gray-600 mb-2">查看当前数据库状态。</p>

          <button
            @click="startSync"
            class="w-full bg-pink-600 hover:bg-pink-700 text-white font-medium py-2 px-4 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-pink-500 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="isSyncing"
          >
            <span v-if="isSyncing" class="flex items-center justify-center">
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              检查中...
            </span>
            <span v-else>刷新状态</span>
          </button>
        </div>

        <!-- 同步结果显示 -->
        <div v-if="syncResult" class="mt-6 border-t pt-4">
          <h3 class="font-medium text-gray-900 mb-2">数据库状态</h3>
          <div class="bg-gray-50 dark:bg-gray-900 p-3 rounded-md border border-gray-200 dark:border-gray-700">
            <div class="grid grid-cols-2 gap-2 mb-3">
              <div class="text-sm">
                <span class="text-gray-500">检查时间：</span>
                <span class="text-gray-900 dark:text-gray-100">{{ formatDateTime(syncResult.timestamp) }}</span>
              </div>
              <div class="text-sm">
                <span class="text-gray-500">数据库记录：</span>
                <span class="text-gray-900 dark:text-gray-100 font-medium">{{ syncResult.total_db_records }}</span>
              </div>
            </div>
            <div class="text-sm p-2 rounded-md bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300">
              {{ syncResult.message }}
            </div>
          </div>
        </div>
      </div>

      <!-- 数据完整性检查面板 -->
      <div v-if="currentTab === 'integrity'">
        <!-- 报告概览卡片 -->
        <div v-if="reportData" class="mb-6">
          <div class="bg-gradient-to-r from-pink-50 to-rose-50 dark:from-pink-900/20 dark:to-rose-900/20 p-4 rounded-lg border border-pink-100 dark:border-pink-800/30 mb-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-500 dark:text-gray-400">数据库总记录</p>
                <p class="text-2xl font-bold text-pink-600 dark:text-pink-400">{{ reportData.total_records?.toLocaleString() }}</p>
              </div>
              <div class="text-right">
                <p class="text-sm text-gray-500 dark:text-gray-400">数据年份</p>
                <p class="text-2xl font-bold text-gray-800 dark:text-gray-200">{{ reportData.years?.length || 0 }}</p>
              </div>
            </div>
          </div>

          <!-- 各年份柱状图 -->
          <div v-if="reportData.years && reportData.years.length" class="bg-white dark:bg-gray-800 p-4 rounded-lg border border-gray-200 dark:border-gray-700">
            <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">各年份记录分布</h4>
            <div class="space-y-2">
              <div v-for="year in reportData.years" :key="year.year" class="flex items-center gap-3">
                <span class="text-xs font-mono text-gray-500 w-10 text-right">{{ year.year }}</span>
                <div class="flex-1 bg-gray-100 dark:bg-gray-700 rounded-full h-5 overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="year.count === reportData.max_year_count ? 'bg-pink-500' : 'bg-pink-300 dark:bg-pink-600'"
                    :style="{ width: reportData.max_year_count ? (year.count / reportData.max_year_count * 100) + '%' : '0%' }"
                  ></div>
                </div>
                <span class="text-xs font-medium text-gray-600 dark:text-gray-400 w-20 text-right">{{ year.count?.toLocaleString() }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="mb-4">
          <button
            @click="startCheck"
            class="w-full bg-pink-600 hover:bg-pink-700 text-white font-medium py-2 px-4 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-pink-500 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="isChecking"
          >
            <span v-if="isChecking" class="flex items-center justify-center">
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              检查中...
            </span>
            <span v-else>刷新数据</span>
          </button>
        </div>

        <!-- 检查结果提示 -->
        <div v-if="checkResult" class="text-sm p-2 rounded-md" :class="checkResult.difference === 0 ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300'">
          {{ checkResult.message }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { syncData, getSyncResult, checkDataIntegrity, getIntegrityReport } from '~/utils/api'
import { showNotify } from 'vant'
import 'vant/es/notify/style'

// 定义Props
const props = defineProps({
  showModal: {
    type: Boolean,
    default: false
  },
  initialTab: {
    type: String,
    default: 'integrity'
  }
})

// 定义事件
const emit = defineEmits(['update:showModal', 'sync-complete', 'check-complete'])

// 状态变量
const currentTab = ref(props.initialTab)
const isSyncing = ref(false)
const isChecking = ref(false)
const syncResult = ref(null)
const checkResult = ref(null)
const reportData = ref(null)

// 格式化日期时间
const formatDateTime = (dateTimeString) => {
  if (!dateTimeString) return ''
  const date = new Date(dateTimeString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 关闭模态框
const closeModal = () => {
  emit('update:showModal', false)
}

// 获取上次同步结果
const fetchSyncResult = async () => {
  try {
    const response = await getSyncResult()
    if (response.data && response.data.success) {
      syncResult.value = response.data
    }
  } catch (error) {
    console.error('获取状态失败:', error)
  }
}

// 开始同步数据
const startSync = async () => {
  isSyncing.value = true
  try {
    const response = await syncData()

    if (response.data.success) {
      syncResult.value = response.data
      showNotify({ type: 'success', message: response.data.message || '状态刷新完成' })
      emit('sync-complete', response.data)
    } else {
      showNotify({ type: 'danger', message: response.data.message || '操作失败' })
    }
  } catch (error) {
    console.error('获取状态失败:', error)
    showNotify({ type: 'danger', message: error.response?.data?.detail || '获取状态失败' })
  } finally {
    isSyncing.value = false
  }
}

// 开始数据完整性检查
const startCheck = async () => {
  isChecking.value = true
  try {
    const response = await checkDataIntegrity()

    if (response.data.success) {
      checkResult.value = response.data

      if (response.data.message && response.data.message.includes('配置中禁用')) {
        showNotify({
          type: 'warning',
          message: '数据完整性校验已在配置中禁用，但已强制执行检查'
        })
      } else {
        showNotify({ type: 'success', message: '数据完整性检查完成' })
      }

      emit('check-complete', response.data)
      await fetchIntegrityReport()
    } else {
      showNotify({ type: 'danger', message: response.data.message || '检查失败' })
    }
  } catch (error) {
    console.error('数据完整性检查失败:', error)
    showNotify({ type: 'danger', message: error.response?.data?.detail || '数据完整性检查失败' })
  } finally {
    isChecking.value = false
  }
}

// 获取完整性报告内容
const fetchIntegrityReport = async () => {
  try {
    const response = await getIntegrityReport()
    if (response.data && response.data.data) {
      reportData.value = response.data.data
    }
  } catch (error) {
    console.error('获取报告失败:', error)
  }
}

// 监听模态框状态变化
watch(() => props.showModal, async (newVal) => {
  if (newVal) {
    await fetchIntegrityReport()
    await fetchSyncResult()
  }
})

// 监听initialTab变化
watch(() => props.initialTab, (newVal) => {
  currentTab.value = newVal
})

// 组件挂载时
onMounted(async () => {
  if (props.showModal) {
    await fetchIntegrityReport()
    await fetchSyncResult()
  }
})
</script>
