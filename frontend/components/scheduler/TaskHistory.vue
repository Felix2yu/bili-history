<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[9999] flex items-center justify-center">
      <div class="fixed inset-0 bg-black/60" @click="$emit('update:show', false)"></div>
      <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] z-10 flex flex-col mx-4">
        <!-- 标题栏 -->
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-700">
          <div class="flex items-center gap-2.5 min-w-0">
            <div class="w-8 h-8 rounded-xl bg-[#fb7299]/10 flex items-center justify-center flex-shrink-0">
              <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">{{ taskName }}</h3>
              <p class="text-[11px] text-gray-500 dark:text-gray-400">执行历史</p>
            </div>
          </div>
          <div class="flex items-center gap-1.5">
            <button @click="refreshHistory"
              class="px-2.5 py-1 text-xs font-medium text-[#fb7299] hover:bg-[#fb7299]/10 rounded-lg transition-colors">
              刷新
            </button>
            <button @click="$emit('update:show', false)"
              class="w-7 h-7 rounded-lg flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- 内容 -->
        <div class="flex-1 overflow-y-auto px-5 py-4">
          <!-- 加载 -->
          <div v-if="loading" class="flex justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-2 border-[#fb7299] border-t-transparent"></div>
          </div>

          <!-- 空状态 -->
          <div v-else-if="!history.length" class="text-center py-12">
            <div class="w-12 h-12 mx-auto mb-3 rounded-xl bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
              <svg class="w-6 h-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">暂无执行历史</p>
          </div>

          <!-- 历史列表 -->
          <div v-else class="space-y-2">
            <div v-for="record in history" :key="record.execution_id"
              class="p-3 rounded-xl border border-gray-100 dark:border-gray-700 hover:border-gray-200 dark:hover:border-gray-600 transition-colors">
              <div class="flex items-center justify-between mb-1">
                <div class="flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full flex-shrink-0"
                    :class="{
                      'bg-green-500': record.status === 'success',
                      'bg-amber-500': record.status === 'running',
                      'bg-red-500': record.status === 'error'
                    }"></span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">
                    {{ statusLabel(record.status) }}
                  </span>
                </div>
                <span class="text-xs text-gray-400">{{ record.duration?.toFixed(1) || 0 }}s</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-xs text-gray-500 dark:text-gray-400">
                  {{ record.start_time?.replace('T', ' ') }}
                </span>
                <button v-if="record.error" @click="viewError(record)"
                  class="text-xs text-red-500 hover:text-red-600 transition-colors">
                  查看错误
                </button>
              </div>
            </div>
          </div>

          <!-- 分页 -->
          <div v-if="total > pageSize" class="mt-4 flex justify-center">
            <div class="flex items-center gap-2">
              <button @click="handlePageChange(currentPage - 1)" :disabled="currentPage <= 1"
                class="px-3 py-1.5 text-xs rounded-lg border border-gray-200 dark:border-gray-600 disabled:opacity-40 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                上一页
              </button>
              <span class="text-xs text-gray-500">{{ currentPage }} / {{ Math.ceil(total / pageSize) }}</span>
              <button @click="handlePageChange(currentPage + 1)" :disabled="currentPage >= Math.ceil(total / pageSize)"
                class="px-3 py-1.5 text-xs rounded-lg border border-gray-200 dark:border-gray-600 disabled:opacity-40 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                下一页
              </button>
            </div>
          </div>
        </div>

        <!-- 错误详情弹窗 -->
        <Teleport to="body">
          <div v-if="showErrorDialog" class="fixed inset-0 z-[10000] flex items-center justify-center">
            <div class="fixed inset-0 bg-black/60" @click="showErrorDialog = false"></div>
            <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg z-10 mx-4">
              <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-700">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">错误详情</h3>
                <button @click="showErrorDialog = false"
                  class="w-7 h-7 rounded-lg flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div class="p-5">
                <pre class="text-xs font-mono text-red-600 dark:text-red-400 whitespace-pre-wrap bg-red-50 dark:bg-red-900/20 p-3 rounded-xl">{{ selectedRecord?.error }}</pre>
              </div>
            </div>
          </div>
        </Teleport>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { showNotify } from 'vant'
import 'vant/es/notify/style'
import { getTaskHistory } from '~/utils/api'

const props = defineProps({
  show: { type: Boolean, default: false },
  taskId: { type: String, default: '' },
  taskName: { type: String, default: '' }
})

const emit = defineEmits(['update:show'])

const loading = ref(false)
const history = ref([])
const showErrorDialog = ref(false)
const selectedRecord = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const fetchHistory = async () => {
  if (!props.taskId) return
  loading.value = true
  try {
    const response = await getTaskHistory({
      task_id: props.taskId, include_subtasks: false,
      page: currentPage.value, page_size: pageSize.value
    })
    if (response.data?.status === 'success') {
      history.value = response.data.history || []
      total.value = response.data.total || 0
    }
  } catch {
    showNotify({ type: 'danger', message: '获取历史记录失败' })
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page) => {
  currentPage.value = page
  fetchHistory()
}

const refreshHistory = () => fetchHistory()

const viewError = (record) => {
  selectedRecord.value = record
  showErrorDialog.value = true
}

const statusLabel = (status) => {
  return { success: '成功', error: '失败', running: '执行中', pending: '等待中' }[status] || status
}

onMounted(() => { if (props.show && props.taskId) fetchHistory() })
watch(() => props.show, (val) => { if (val && props.taskId) fetchHistory() })
</script>
