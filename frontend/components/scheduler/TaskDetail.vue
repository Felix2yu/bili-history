<template>
  <Teleport to="body">
    <div v-if="show && task" class="fixed inset-0 z-[9999] flex items-center justify-center">
      <div class="fixed inset-0 bg-black/60" @click="$emit('update:show', false)"></div>
      <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] z-10 flex flex-col mx-4">
        <!-- 标题栏 -->
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-700">
          <div class="flex items-center gap-2.5 min-w-0">
            <div class="w-8 h-8 rounded-xl bg-[#fb7299]/10 flex items-center justify-center flex-shrink-0">
              <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
              </svg>
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">{{ task.config?.name }}</h3>
              <p class="text-[11px] text-gray-500 dark:text-gray-400">{{ task.task_id }}</p>
            </div>
          </div>
          <div class="flex items-center gap-1.5 flex-shrink-0">
            <button @click="$emit('view-history', task.task_id)"
              class="px-2.5 py-1 text-xs font-medium text-white bg-[#fb7299] rounded-lg hover:bg-[#fb7299]/90 transition-colors">
              历史
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
        <div class="flex-1 overflow-y-auto px-5 py-4 space-y-4">
          <!-- 状态栏 -->
          <div class="flex items-center gap-2">
            <span v-if="task.execution?.status"
              class="px-2 py-0.5 text-xs font-medium rounded-lg"
              :class="{
                'bg-green-50 text-green-600 dark:bg-green-900/30 dark:text-green-400': task.execution.status === 'success',
                'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400': task.execution.status === 'running',
                'bg-red-50 text-red-600 dark:bg-red-900/30 dark:text-red-400': task.execution.status === 'error'
              }">
              {{ statusLabel }}
            </span>
            <span class="px-2 py-0.5 text-xs font-medium rounded-lg"
              :class="task.config?.enabled ? 'bg-green-50 text-green-600 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'">
              {{ task.config?.enabled ? '已启用' : '已禁用' }}
            </span>
          </div>

          <!-- 基本信息 -->
          <div>
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">基本信息</h4>
            <div class="grid grid-cols-2 gap-3">
              <div><p class="text-[11px] text-gray-400 mb-0.5">API 端点</p>
                <p class="text-sm text-gray-800 dark:text-gray-100 font-mono truncate">{{ task.config?.endpoint }}</p></div>
              <div><p class="text-[11px] text-gray-400 mb-0.5">请求方法</p>
                <p class="text-sm text-gray-800 dark:text-gray-100">{{ task.config?.method }}</p></div>
              <div><p class="text-[11px] text-gray-400 mb-0.5">最后修改</p>
                <p class="text-sm text-gray-800 dark:text-gray-100">{{ task.last_modified?.replace('T', ' ') }}</p></div>
              <div><p class="text-[11px] text-gray-400 mb-0.5">优先级</p>
                <p class="text-sm text-gray-800 dark:text-gray-100">{{ task.config?.priority || 0 }}</p></div>
            </div>
          </div>

          <!-- 调度信息 -->
          <div>
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">调度信息</h4>
            <div class="grid grid-cols-2 gap-3">
              <div><p class="text-[11px] text-gray-400 mb-0.5">调度类型</p>
                <p class="text-sm text-gray-800 dark:text-gray-100">{{ scheduleTypeLabel }}</p></div>
              <div><p class="text-[11px] text-gray-400 mb-0.5">执行时间</p>
                <p class="text-sm text-gray-800 dark:text-gray-100">
                  <template v-if="task.task_type === 'main'">
                    {{ task.config?.schedule_type === 'interval'
                      ? `${task.config?.interval_value || '-'} ${unitLabel}`
                      : (task.config?.schedule_time || '未设置') }}
                  </template>
                  <template v-else>依赖于主任务</template>
                </p>
              </div>
              <div><p class="text-[11px] text-gray-400 mb-0.5">上次执行</p>
                <p class="text-sm text-gray-800 dark:text-gray-100">{{ task.execution?.last_run?.replace('T', ' ') || '从未执行' }}</p></div>
              <div><p class="text-[11px] text-gray-400 mb-0.5">下次执行</p>
                <p class="text-sm text-gray-800 dark:text-gray-100">
                  <template v-if="task.task_type === 'main'">{{ task.execution?.next_run || '未排定' }}</template>
                  <template v-else>依赖于主任务</template>
                </p>
              </div>
            </div>
          </div>

          <!-- 执行统计 -->
          <div>
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">执行统计</h4>
            <div class="grid grid-cols-3 gap-3">
              <div class="text-center p-2 rounded-xl bg-gray-50 dark:bg-gray-700/50">
                <div class="text-lg font-bold"
                  :class="{
                    'text-green-600 dark:text-green-400': executionInfo.successRate >= 90,
                    'text-amber-600 dark:text-amber-400': executionInfo.successRate >= 60 && executionInfo.successRate < 90,
                    'text-red-600 dark:text-red-400': executionInfo.successRate < 60
                  }">{{ Math.round(executionInfo.successRate) }}%</div>
                <div class="text-[10px] text-gray-400">成功率</div>
              </div>
              <div class="text-center p-2 rounded-xl bg-gray-50 dark:bg-gray-700/50">
                <div class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ executionInfo.totalRuns }}</div>
                <div class="text-[10px] text-gray-400">总执行</div>
              </div>
              <div class="text-center p-2 rounded-xl bg-gray-50 dark:bg-gray-700/50">
                <div class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ executionInfo.avgDuration.toFixed(1) }}s</div>
                <div class="text-[10px] text-gray-400">平均耗时</div>
              </div>
            </div>
            <div class="mt-2 flex items-center gap-2 text-xs text-gray-500">
              <span class="text-green-600">{{ executionInfo.successRuns }} 成功</span>
              <span class="text-gray-300">·</span>
              <span class="text-red-600">{{ executionInfo.failRuns }} 失败</span>
            </div>
          </div>

          <!-- 依赖任务 -->
          <div v-if="task.depends_on">
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">依赖任务</h4>
            <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-medium bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
              <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
              {{ task.depends_on.name }}
            </span>
          </div>

          <!-- 最近错误 -->
          <div v-if="task.execution?.last_error"
            class="p-3 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-900/30">
            <h4 class="text-xs font-semibold text-red-600 mb-1">最近错误</h4>
            <p class="text-xs text-red-600 dark:text-red-400 whitespace-pre-wrap font-mono">{{ task.execution.last_error }}</p>
          </div>
        </div>

        <!-- 底部操作 -->
        <div class="flex items-center justify-between px-5 py-3 border-t border-gray-100 dark:border-gray-700">
          <button @click="$emit('delete-task', task.task_id)"
            class="px-3 py-1.5 text-xs font-medium text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors">
            删除任务
          </button>
          <div class="flex items-center gap-2">
            <button @click="$emit('toggle-enabled', task.task_id, !task.config?.enabled)"
              class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors"
              :class="task.config?.enabled ? 'text-amber-600 hover:bg-amber-50 dark:hover:bg-amber-900/20' : 'text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20'">
              {{ task.config?.enabled ? '禁用' : '启用' }}
            </button>
            <button @click="$emit('edit-task', task.task_id)"
              class="px-3 py-1.5 text-xs font-medium text-[#fb7299] hover:bg-[#fb7299]/10 rounded-lg transition-colors">
              编辑
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  task: { type: Object, default: null }
})

const emit = defineEmits(['update:show', 'view-history', 'edit-task', 'execute-task', 'toggle-enabled', 'delete-task'])

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
  const map = { daily: '每日', chain: '链式任务', once: '一次性', interval: '间隔执行' }
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
</script>
