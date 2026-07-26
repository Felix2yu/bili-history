<template>
  <Teleport to="body">
    <div v-if="show && task" class="fixed inset-0 z-[9999] flex items-center justify-center p-4">
      <Transition name="fade">
        <div v-if="show" class="fixed inset-0 bg-black/60 backdrop-blur-sm" @click="$emit('update:show', false)"></div>
      </Transition>
      <Transition name="scale">
        <div v-if="show" class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] z-10 flex flex-col overflow-hidden">
          <!-- 顶部装饰 -->
          <div class="h-1.5 bg-accent"></div>

          <!-- 标题栏 -->
          <div class="flex items-center justify-between px-5 md:px-6 py-4 md:py-5 border-b border-gray-100 dark:border-gray-700/50">
            <div class="flex items-center gap-3 min-w-0">
              <div class="w-11 h-11 md:w-12 md:h-12 rounded-xl flex-shrink-0 flex items-center justify-center"
                :class="getScheduleTypeBg(task.config?.schedule_type)">
                <svg class="w-5 h-5 md:w-6 md:h-6" :class="getScheduleTypeText(task.config?.schedule_type)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round"
                    v-if="task.config?.schedule_type === 'daily'"
                    d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round"
                    v-else-if="task.config?.schedule_type === 'chain'"
                    d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  <path stroke-linecap="round" stroke-linejoin="round"
                    v-else-if="task.config?.schedule_type === 'once'"
                    d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" v-else
                    d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div class="min-w-0">
                <h3 class="text-base md:text-lg font-semibold text-gray-900 dark:text-white truncate">{{ task.config?.name }}</h3>
                <p class="text-[11px] md:text-xs text-gray-500 dark:text-gray-400 font-mono">{{ task.task_id }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button @click="$emit('view-history', task.task_id)"
                class="px-3 md:px-3.5 py-1.5 md:py-2 text-xs md:text-sm font-medium text-white bg-accent rounded-xl hover:shadow-md hover:shadow-accent/25 transition-all">
                历史
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
          <div class="flex-1 overflow-y-auto px-5 md:px-6 py-4 md:py-5 space-y-5">
            <!-- 状态栏 -->
            <div class="flex flex-wrap items-center gap-2">
              <span v-if="task.execution?.status"
                class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs md:text-sm font-medium rounded-xl"
                :class="statusClass">
                <span class="w-1.5 h-1.5 rounded-full animate-pulse" :class="statusDotClass"></span>
                {{ statusLabel }}
              </span>
              <span class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs md:text-sm font-medium rounded-xl"
                :class="task.config?.enabled
                  ? 'bg-green-500/10 text-green-600 dark:text-green-400'
                  : 'bg-gray-500/10 text-gray-500 dark:text-gray-400'">
                {{ task.config?.enabled ? '已启用' : '已禁用' }}
              </span>
            </div>

            <!-- 执行统计卡片 -->
            <div class="grid grid-cols-3 gap-2 md:gap-3">
              <div class="relative overflow-hidden p-3 md:p-4 rounded-2xl"
                :class="successRateBgClass">
                <div class="text-xl md:text-2xl font-bold" :class="successRateTextClass">{{ Math.round(executionInfo.successRate) }}%</div>
                <div class="text-[10px] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">成功率</div>
              </div>
              <div class="relative overflow-hidden p-3 md:p-4 rounded-2xl bg-gray-50 dark:bg-gray-700/50">
                <div class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">{{ executionInfo.totalRuns }}</div>
                <div class="text-[10px] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">总执行</div>
              </div>
              <div class="relative overflow-hidden p-3 md:p-4 rounded-2xl bg-gray-50 dark:bg-gray-700/50">
                <div class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white">{{ executionInfo.avgDuration.toFixed(1) }}<span class="text-xs md:text-sm text-gray-500">s</span></div>
                <div class="text-[10px] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">平均耗时</div>
              </div>
            </div>

            <div class="flex items-center gap-2 text-[11px] md:text-xs">
              <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
              <span class="text-green-600 dark:text-green-400 font-medium">{{ executionInfo.successRuns }}</span>
              <span class="text-gray-400">成功</span>
              <span class="text-gray-300 dark:text-gray-600">·</span>
              <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
              <span class="text-red-600 dark:text-red-400 font-medium">{{ executionInfo.failRuns }}</span>
              <span class="text-gray-400">失败</span>
            </div>

            <!-- 基本信息 -->
            <div>
              <h4 class="text-[11px] md:text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">基本信息</h4>
              <div class="grid grid-cols-2 gap-3 md:gap-4">
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">API 端点</p>
                  <p class="text-xs md:text-sm text-gray-800 dark:text-gray-200 font-mono truncate" :title="task.config?.endpoint">{{ task.config?.endpoint }}</p>
                </div>
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">请求方法</p>
                  <p class="text-xs md:text-sm font-semibold" :class="methodClass">{{ task.config?.method }}</p>
                </div>
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">最后修改</p>
                  <p class="text-xs md:text-sm text-gray-800 dark:text-gray-200">{{ task.last_modified?.replace('T', ' ') }}</p>
                </div>
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">优先级</p>
                  <p class="text-xs md:text-sm text-gray-800 dark:text-gray-200">{{ task.config?.priority || 0 }}</p>
                </div>
              </div>
            </div>

            <!-- 调度信息 -->
            <div>
              <h4 class="text-[11px] md:text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">调度信息</h4>
              <div class="grid grid-cols-2 gap-3 md:gap-4">
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">调度类型</p>
                  <p class="text-xs md:text-sm text-gray-800 dark:text-gray-200 font-medium">{{ scheduleTypeLabel }}</p>
                </div>
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">执行时间</p>
                  <p class="text-xs md:text-sm text-gray-800 dark:text-gray-200 font-medium">
                    <template v-if="task.task_type === 'main'">
                      {{ task.config?.schedule_type === 'interval' ? unitLabel : (task.config?.schedule_time || '未设置') }}
                    </template>
                    <template v-else>依赖主任务</template>
                  </p>
                </div>
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">上次执行</p>
                  <p class="text-xs md:text-sm text-gray-800 dark:text-gray-200">{{ task.execution?.last_run?.replace('T', ' ') || '从未执行' }}</p>
                </div>
                <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-700/30">
                  <p class="text-[10px] md:text-xs text-gray-400 mb-1">下次执行</p>
                  <p class="text-xs md:text-sm text-gray-800 dark:text-gray-200">
                    <template v-if="task.task_type === 'main'">{{ task.execution?.next_run || '未排定' }}</template>
                    <template v-else>依赖主任务</template>
                  </p>
                </div>
              </div>
            </div>

            <!-- 依赖任务 -->
            <div v-if="task.depends_on">
              <h4 class="text-[11px] md:text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">依赖任务</h4>
              <div class="inline-flex items-center gap-2 px-3 py-2 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                </svg>
                <span class="text-xs md:text-sm font-medium">{{ task.depends_on.name }}</span>
              </div>
            </div>

            <!-- 最近错误 -->
            <div v-if="task.execution?.last_error"
              class="relative overflow-hidden p-4 rounded-2xl bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-900/30">
              <div class="flex items-start gap-3">
                <div class="w-8 h-8 rounded-lg bg-red-500/20 flex items-center justify-center flex-shrink-0">
                  <svg class="w-4 h-4 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                  </svg>
                </div>
                <div class="min-w-0 flex-1">
                  <h4 class="text-xs md:text-sm font-semibold text-red-600 dark:text-red-400 mb-1">最近错误</h4>
                  <p class="text-[11px] md:text-xs text-red-600 dark:text-red-400 whitespace-pre-wrap font-mono break-all">{{ task.execution.last_error }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- 底部操作 -->
          <div class="flex items-center justify-between px-5 md:px-6 py-3 md:py-4 border-t border-gray-100 dark:border-gray-700/50 bg-gray-50/50 dark:bg-gray-800/50">
            <button @click="$emit('delete-task', task.task_id)"
              class="inline-flex items-center gap-1.5 px-3 md:px-3.5 py-2 text-xs md:text-sm font-medium text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-xl transition-colors">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round"
                  d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.034-2.09 1.02-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
              </svg>
              删除
            </button>
            <div class="flex items-center gap-2">
              <button @click="$emit('toggle-enabled', task.task_id, !task.config?.enabled)"
                class="inline-flex items-center gap-1.5 px-3 md:px-3.5 py-2 text-xs md:text-sm font-medium rounded-xl transition-colors"
                :class="task.config?.enabled
                  ? 'text-amber-600 hover:bg-amber-50 dark:hover:bg-amber-900/20'
                  : 'text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20'">
                {{ task.config?.enabled ? '禁用' : '启用' }}
              </button>
              <button @click="$emit('edit-task', task.task_id)"
                class="inline-flex items-center gap-1.5 px-3 md:px-4 py-2 text-xs md:text-sm font-medium text-white bg-accent rounded-xl hover:shadow-md hover:shadow-accent/25 transition-all">
                编辑
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  task: { type: Object, default: null }
})

defineEmits(['update:show', 'view-history', 'edit-task', 'execute-task', 'toggle-enabled', 'delete-task', 'refresh'])

const executionInfo = computed(() => {
  const e = props.task?.execution || {}
  return {
    successRate: typeof e.success_rate === 'number' ? e.success_rate : 0,
    avgDuration: typeof e.avg_duration === 'number' ? e.avg_duration : 0,
    totalRuns: typeof e.total_runs === 'number' ? e.total_runs : 0,
    successRuns: typeof e.success_runs === 'number' ? e.success_runs : 0,
    failRuns: typeof e.fail_runs === 'number' ? e.fail_runs : 0
  }
})

const scheduleTypeLabel = computed(() => {
  const map = { daily: '每日执行', chain: '链式任务', once: '一次性', interval: '间隔执行' }
  return map[props.task?.config?.schedule_type] || props.task?.config?.schedule_type || '-'
})

const unitLabel = computed(() => {
  const map = { minutes: '分钟', hours: '小时', days: '天', months: '月', years: '年' }
  return `${props.task?.config?.interval_value || '-'} ${map[props.task?.config?.interval_unit] || ''}`
})

const statusLabel = computed(() => {
  const map = { success: '成功', running: '执行中', error: '失败', pending: '等待中' }
  return map[props.task?.execution?.status] || props.task?.execution?.status || '-'
})

const statusClass = computed(() => {
  const map = {
    success: 'bg-green-500/10 text-green-600 dark:text-green-400',
    running: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
    error: 'bg-red-500/10 text-red-600 dark:text-red-400',
    pending: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
  }
  return map[props.task?.execution?.status] || 'bg-gray-500/10 text-gray-500'
})

const statusDotClass = computed(() => {
  const map = {
    success: 'bg-green-500',
    running: 'bg-amber-500',
    error: 'bg-red-500',
    pending: 'bg-blue-500',
  }
  return map[props.task?.execution?.status] || 'bg-gray-500'
})

const successRateBgClass = computed(() => {
  const rate = executionInfo.value.successRate
  if (rate >= 90) return 'bg-green-500/10'
  if (rate >= 60) return 'bg-amber-500/10'
  return 'bg-red-500/10'
})

const successRateTextClass = computed(() => {
  const rate = executionInfo.value.successRate
  if (rate >= 90) return 'text-green-600 dark:text-green-400'
  if (rate >= 60) return 'text-amber-600 dark:text-amber-400'
  return 'text-red-600 dark:text-red-400'
})

const methodClass = computed(() => {
  const map = {
    GET: 'text-green-600 dark:text-green-400',
    POST: 'text-blue-600 dark:text-blue-400',
    PUT: 'text-amber-600 dark:text-amber-400',
    DELETE: 'text-red-600 dark:text-red-400',
    PATCH: 'text-purple-600 dark:text-purple-400',
  }
  return map[props.task?.config?.method] || 'text-gray-600 dark:text-gray-400'
})

const getScheduleTypeBg = (type) => {
  const map = {
    daily: 'bg-blue-500/15',
    chain: 'bg-purple-500/15',
    once: 'bg-green-500/15',
    interval: 'bg-amber-500/15',
  }
  return map[type] || 'bg-gray-500/15'
}

const getScheduleTypeText = (type) => {
  const map = {
    daily: 'text-blue-600 dark:text-blue-400',
    chain: 'text-purple-600 dark:text-purple-400',
    once: 'text-green-600 dark:text-green-400',
    interval: 'text-amber-600 dark:text-amber-400',
  }
  return map[type] || 'text-gray-600 dark:text-gray-400'
}
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
</style>
