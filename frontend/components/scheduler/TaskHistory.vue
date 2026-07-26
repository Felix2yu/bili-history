<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[9999] flex items-center justify-center p-4">
      <Transition name="fade">
        <div v-if="show" class="fixed inset-0 bg-black/60 backdrop-blur-sm" @click="$emit('update:show', false)"></div>
      </Transition>
      <Transition name="scale">
        <div v-if="show" class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] z-10 flex flex-col overflow-hidden">
          <!-- 顶部装饰 -->
          <div class="h-1.5 bg-gradient-to-r from-accent via-accent/70 to-accent/40"></div>

          <!-- 标题栏 -->
          <div class="flex items-center justify-between px-5 md:px-6 py-4 md:py-5 border-b border-gray-100 dark:border-gray-700/50">
            <div class="flex items-center gap-3 min-w-0">
              <div class="w-11 h-11 md:w-12 md:h-12 rounded-xl flex-shrink-0 flex items-center justify-center bg-accent/15">
                <svg class="w-5 h-5 md:w-6 md:h-6 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div class="min-w-0">
                <h3 class="text-base md:text-lg font-semibold text-gray-900 dark:text-white truncate">{{ taskName }}</h3>
                <p class="text-[11px] md:text-xs text-gray-500 dark:text-gray-400">执行历史记录</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button @click="refreshHistory"
                class="inline-flex items-center gap-1.5 px-3 md:px-3.5 py-2 text-xs md:text-sm font-medium text-accent hover:bg-accent/10 rounded-xl transition-colors">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                </svg>
                刷新
              </button>
              <button @click="$emit('update:show', false)"
                class="w-9 h-9 md:w-10 md:h-10 rounded-xl flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <svg class="w-4 h-4 md:w-5 md:h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          <!-- 内容 -->
          <div class="flex-1 overflow-y-auto px-5 md:px-6 py-4 md:py-5">
            <!-- 加载 -->
            <div v-if="loading" class="flex justify-center py-16">
              <div class="relative">
                <div class="w-10 h-10 rounded-full border-2 border-accent/20"></div>
                <div class="absolute inset-0 w-10 h-10 rounded-full border-2 border-accent border-t-transparent animate-spin"></div>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-else-if="!history.length" class="text-center py-16">
              <div class="relative mx-auto w-20 h-20 mb-4">
                <div class="absolute inset-0 rounded-3xl bg-gray-100 dark:bg-gray-700"></div>
                <div class="absolute inset-2 rounded-3xl bg-white dark:bg-gray-800 flex items-center justify-center">
                  <svg class="w-10 h-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
              <p class="text-sm text-gray-500 dark:text-gray-400">暂无执行历史</p>
            </div>

            <!-- 时间线历史列表 -->
            <div v-else class="relative">
              <!-- 时间线竖线 -->
              <div class="absolute left-4 top-2 bottom-2 w-0.5 bg-gray-200 dark:bg-gray-700"></div>

              <div class="space-y-4">
                <TransitionGroup name="timeline">
                  <div v-for="(record, idx) in history" :key="record.execution_id" class="relative pl-12">
                    <!-- 时间线节点 -->
                    <div class="absolute left-0 top-1 w-9 h-9 rounded-full flex items-center justify-center"
                      :class="getNodeBgClass(record.status)">
                      <div class="w-3 h-3 rounded-full" :class="getNodeDotClass(record.status)"></div>
                    </div>

                    <!-- 卡片 -->
                    <div class="group relative rounded-2xl border transition-all duration-300 hover:shadow-md"
                      :class="getCardClass(record.status)">
                      <div class="p-4">
                        <!-- 顶部：状态 + 耗时 -->
                        <div class="flex items-center justify-between mb-2">
                          <div class="flex items-center gap-2">
                            <span class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-lg"
                              :class="getStatusClass(record.status)">
                              <span class="w-1.5 h-1.5 rounded-full" :class="getStatusDotClass(record.status)"></span>
                              {{ statusLabel(record.status) }}
                            </span>
                            <span v-if="idx === 0" class="inline-flex items-center gap-1 px-2 py-0.5 text-[10px] font-medium rounded-lg bg-accent/10 text-accent">
                              最新
                            </span>
                          </div>
                          <div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
                            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            {{ record.duration?.toFixed(1) || 0 }}s
                          </div>
                        </div>

                        <!-- 时间 -->
                        <div class="flex items-center gap-2 text-[11px] md:text-xs text-gray-500 dark:text-gray-400">
                          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
                          </svg>
                          {{ formatTime(record.start_time) }}
                        </div>

                        <!-- 错误按钮 -->
                        <div v-if="record.error" class="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700/50">
                          <button @click="viewError(record)"
                            class="inline-flex items-center gap-1.5 text-xs font-medium text-red-500 hover:text-red-600 dark:hover:text-red-400 transition-colors">
                            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                            </svg>
                            查看错误详情
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </TransitionGroup>
              </div>
            </div>

            <!-- 分页 -->
            <div v-if="total > pageSize" class="mt-6 flex justify-center">
              <div class="flex items-center gap-2">
                <button @click="handlePageChange(currentPage - 1)" :disabled="currentPage <= 1"
                  class="inline-flex items-center gap-1 px-3.5 py-2 text-xs font-medium rounded-xl border border-gray-200 dark:border-gray-600 disabled:opacity-40 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" />
                  </svg>
                  上一页
                </button>
                <span class="px-3 text-xs text-gray-500">{{ currentPage }} / {{ Math.ceil(total / pageSize) }}</span>
                <button @click="handlePageChange(currentPage + 1)" :disabled="currentPage >= Math.ceil(total / pageSize)"
                  class="inline-flex items-center gap-1 px-3.5 py-2 text-xs font-medium rounded-xl border border-gray-200 dark:border-gray-600 disabled:opacity-40 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                  下一页
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>

      <!-- 错误详情弹窗 -->
      <Teleport to="body">
        <Transition name="fade">
          <div v-if="showErrorDialog" class="fixed inset-0 z-[10000] flex items-center justify-center p-4">
            <div class="fixed inset-0 bg-black/60 backdrop-blur-sm" @click="showErrorDialog = false"></div>
            <Transition name="scale">
              <div v-if="showErrorDialog" class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg z-10 overflow-hidden">
                <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-700">
                  <div class="flex items-center gap-2.5">
                    <div class="w-9 h-9 rounded-xl bg-red-500/20 flex items-center justify-center">
                      <svg class="w-4 h-4 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                      </svg>
                    </div>
                    <h3 class="text-sm font-semibold text-gray-900 dark:text-white">错误详情</h3>
                  </div>
                  <button @click="showErrorDialog = false"
                    class="w-8 h-8 rounded-lg flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <div class="p-5">
                  <pre class="text-xs font-mono text-red-600 dark:text-red-400 whitespace-pre-wrap break-all bg-red-50 dark:bg-red-900/20 p-4 rounded-2xl max-h-[50vh] overflow-y-auto">{{ selectedRecord?.error }}</pre>
                </div>
              </div>
            </Transition>
          </div>
        </Transition>
      </Teleport>
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

defineEmits(['update:show'])

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
  return { success: '执行成功', error: '执行失败', running: '执行中', pending: '等待中' }[status] || status
}

const formatTime = (time) => {
  if (!time) return '-'
  return time.replace('T', ' ')
}

const getStatusClass = (status) => {
  const map = {
    success: 'bg-green-500/10 text-green-600 dark:text-green-400',
    running: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
    error: 'bg-red-500/10 text-red-600 dark:text-red-400',
    pending: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
  }
  return map[status] || 'bg-gray-500/10 text-gray-500'
}

const getStatusDotClass = (status) => {
  const map = {
    success: 'bg-green-500',
    running: 'bg-amber-500',
    error: 'bg-red-500',
    pending: 'bg-blue-500',
  }
  return map[status] || 'bg-gray-500'
}

const getNodeBgClass = (status) => {
  const map = {
    success: 'bg-green-500/20',
    running: 'bg-amber-500/20',
    error: 'bg-red-500/20',
    pending: 'bg-blue-500/20',
  }
  return map[status] || 'bg-gray-500/20'
}

const getNodeDotClass = (status) => {
  const map = {
    success: 'bg-green-500',
    running: 'bg-amber-500',
    error: 'bg-red-500',
    pending: 'bg-blue-500',
  }
  return map[status] || 'bg-gray-500'
}

const getCardClass = (status) => {
  const map = {
    success: 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700',
    running: 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700',
    error: 'bg-red-50/50 dark:bg-red-900/10 border-red-100 dark:border-red-900/30',
    pending: 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700',
  }
  return map[status] || 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700'
}

onMounted(() => { if (props.show && props.taskId) fetchHistory() })
watch(() => props.show, (val) => { if (val && props.taskId) fetchHistory() })
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.scale-enter-active,
.scale-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.scale-enter-from,
.scale-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(10px);
}

.timeline-enter-active,
.timeline-leave-active {
  transition: all 0.3s ease;
}
.timeline-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}
.timeline-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
