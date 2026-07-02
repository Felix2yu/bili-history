<template>
  <div class="min-h-screen bg-gray-50/30 dark:bg-gray-900">
    <div class="py-4">
      <div class="max-w-7xl mx-auto px-3">
        <!-- 标题栏 -->
        <div class="flex items-center justify-between mb-4">
          <div>
            <h1 class="text-lg font-semibold text-gray-900 dark:text-gray-100">计划任务</h1>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">管理定时执行的数据同步和分析任务</p>
          </div>
          <button @click="openCreateTaskModal"
            class="flex items-center gap-1.5 px-3 py-1.5 bg-[#fb7299] text-white rounded-lg text-sm font-medium hover:bg-[#fb7299]/90 transition-colors shadow-sm">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            新建任务
          </button>
        </div>

        <!-- 加载状态 -->
        <div v-if="loading" class="flex justify-center items-center py-20">
          <div class="animate-spin rounded-full h-8 w-8 border-2 border-[#fb7299] border-t-transparent"></div>
        </div>

        <!-- 空状态 -->
        <div v-else-if="tasks.length === 0" class="glass-card p-12 text-center">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-[#fb7299]/10 flex items-center justify-center">
            <svg class="w-8 h-8 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h3 class="text-sm font-medium text-gray-900 dark:text-gray-100">暂无计划任务</h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">创建第一个任务来自动化数据同步</p>
        </div>

        <!-- 任务卡片列表 -->
        <div v-else class="space-y-3">
          <div v-for="task in tasks" :key="task.task_id" class="glass-card overflow-hidden">
            <!-- 主任务卡片 -->
            <div class="p-4">
              <!-- 第一行：名称 + 状态 + 操作 -->
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2 min-w-0">
                  <!-- 展开按钮 -->
                  <button v-if="task.sub_tasks && task.sub_tasks.length > 0"
                    @click="task.isExpanded = !task.isExpanded"
                    class="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors flex-shrink-0">
                    <svg class="w-3.5 h-3.5 text-gray-400 transition-transform duration-200"
                      :class="{ 'rotate-90': task.isExpanded }" fill="none" viewBox="0 0 24 24" stroke="currentColor"
                      stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                  <!-- 任务名称 -->
                  <h3 class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                    {{ task.config?.name || task.task_id }}
                  </h3>
                  <!-- 子任务数量 -->
                  <span v-if="task.sub_tasks && task.sub_tasks.length > 0"
                    class="px-1.5 py-0.5 text-[10px] rounded-full bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 flex-shrink-0">
                    {{ task.sub_tasks.length }} 子任务
                  </span>
                </div>
                <!-- 启用/禁用开关 -->
                <button @click="toggleTaskEnabled(task.task_id, !task.config?.enabled)"
                  class="relative w-9 h-5 rounded-full transition-colors duration-200 flex-shrink-0"
                  :class="task.config?.enabled ? 'bg-[#fb7299]' : 'bg-gray-300 dark:bg-gray-600'">
                  <span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform duration-200"
                    :class="{ 'translate-x-4': task.config?.enabled }"></span>
                </button>
              </div>

              <!-- 第二行：调度信息 -->
              <div class="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400 mb-3">
                <!-- 调度类型 -->
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-medium"
                  :class="{
                    'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400': task.config?.schedule_type === 'daily',
                    'bg-purple-50 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400': task.config?.schedule_type === 'chain',
                    'bg-green-50 text-green-600 dark:bg-green-900/30 dark:text-green-400': task.config?.schedule_type === 'once',
                    'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400': task.config?.schedule_type === 'interval'
                  }">
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round"
                      v-if="task.config?.schedule_type === 'daily'" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round"
                      v-else-if="task.config?.schedule_type === 'chain'" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                    <path stroke-linecap="round" stroke-linejoin="round"
                      v-else-if="task.config?.schedule_type === 'once'" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" v-else
                      d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {{ getScheduleTypeLabel(task.config?.schedule_type) }}
                </span>
                <!-- 调度时间 -->
                <span class="flex items-center gap-1">
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {{ getScheduleTimeDisplay(task) }}
                </span>
                <!-- 执行端点 -->
                <span class="flex items-center gap-1 truncate max-w-[200px]">
                  <svg class="w-3 h-3 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round"
                      d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z" />
                  </svg>
                  <span class="truncate">{{ task.config?.endpoint || '-' }}</span>
                </span>
              </div>

              <!-- 第三行：统计 + 操作 -->
              <div class="flex items-center justify-between">
                <!-- 左侧：成功率 -->
                <div class="flex items-center gap-3">
                  <div v-if="task.execution?.total_runs > 0" class="flex items-center gap-1.5">
                    <div class="flex items-center gap-1">
                      <span class="w-1.5 h-1.5 rounded-full"
                        :class="{
                          'bg-green-500': task.execution.success_rate >= 90,
                          'bg-yellow-500': task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'bg-red-500': task.execution.success_rate < 60
                        }"></span>
                      <span class="text-xs font-medium"
                        :class="{
                          'text-green-600 dark:text-green-400': task.execution.success_rate >= 90,
                          'text-yellow-600 dark:text-yellow-400': task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'text-red-600 dark:text-red-400': task.execution.success_rate < 60
                        }">
                        {{ Math.round(task.execution.success_rate) }}%
                      </span>
                    </div>
                    <span class="text-[10px] text-gray-400">·</span>
                    <span class="text-[10px] text-gray-400">{{ task.execution.total_runs }}次执行</span>
                  </div>
                  <span v-else class="text-[10px] text-gray-400">尚未执行</span>
                </div>
                <!-- 右侧：操作按钮 -->
                <div class="flex items-center gap-1">
                  <button @click="executeTask(task.task_id)"
                    class="p-1.5 rounded-lg text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20 transition-colors"
                    title="执行">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z" />
                    </svg>
                  </button>
                  <button @click="openTaskDetailModal(task.task_id)"
                    class="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                    title="详情">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
                    </svg>
                  </button>
                  <button @click="openEditTaskModal(task.task_id)"
                    class="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                    title="编辑">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                    </svg>
                  </button>
                  <button @click="openCreateSubTaskModal(task.task_id)"
                    class="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                    title="添加子任务">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                    </svg>
                  </button>
                  <button @click="confirmDeleteTask(task.task_id)"
                    class="p-1.5 rounded-lg text-gray-400 hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-500 transition-colors"
                    title="删除">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.034-2.09 1.02-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>

            <!-- 子任务列表 -->
            <div v-if="task.sub_tasks && task.sub_tasks.length > 0 && task.isExpanded"
              class="border-t border-gray-100 dark:border-gray-700/50">
              <div v-for="(sub, idx) in task.sub_tasks" :key="sub.task_id"
                class="px-4 py-3 flex items-center justify-between hover:bg-gray-50/50 dark:hover:bg-gray-800/50 transition-colors"
                :class="{ 'border-b border-gray-100 dark:border-gray-700/50': idx < task.sub_tasks.length - 1 }">
                <div class="flex items-center gap-2 min-w-0">
                  <svg class="w-3.5 h-3.5 text-[#fb7299]/60 flex-shrink-0" fill="none" viewBox="0 0 24 24"
                    stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round"
                      d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  </svg>
                  <span class="text-xs font-medium text-gray-700 dark:text-gray-300 truncate">
                    {{ sub.config?.name || sub.task_id }}
                  </span>
                  <span
                    class="px-1.5 py-0.5 text-[10px] rounded bg-purple-50 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400">
                    链式
                  </span>
                  <button @click="toggleTaskEnabled(sub.task_id, !sub.config?.enabled)"
                    class="relative w-7 h-4 rounded-full transition-colors duration-200"
                    :class="sub.config?.enabled ? 'bg-[#fb7299]' : 'bg-gray-300 dark:bg-gray-600'">
                    <span class="absolute top-0.5 left-0.5 w-3 h-3 bg-white rounded-full shadow transition-transform duration-200"
                      :class="{ 'translate-x-3': sub.config?.enabled }"></span>
                  </button>
                </div>
                <div class="flex items-center gap-1">
                  <button @click="openTaskDetailModal(sub.task_id)"
                    class="p-1 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
                    </svg>
                  </button>
                  <button @click="openEditTaskModal(sub.task_id)"
                    class="p-1 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                    </svg>
                  </button>
                  <button @click="confirmDeleteTask(sub.task_id, task.task_id)"
                    class="p-1 rounded text-gray-400 hover:text-red-500 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.034-2.09 1.02-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 任务详情弹窗 -->
  <TaskDetail v-model:show="showTaskDetailModal" :task="currentTask" @view-history="fetchTaskHistory"
    @edit-task="openEditTaskModal" @execute-task="executeTask" @toggle-enabled="toggleTaskEnabled"
    @delete-task="confirmDeleteTask" @refresh="fetchTasks" />

  <!-- 任务历史弹窗 -->
  <TaskHistory v-model:show="showTaskHistoryModal" :task-id="currentTask?.task_id"
    :task-name="currentTask?.config?.name || currentTask?.task_id" />

  <!-- 创建/编辑任务弹窗 -->
  <TaskForm v-model:show="showTaskFormModal" :is-editing="isEditing" :task-id="currentTask?.task_id"
    :parent-task-id="parentTaskId" :tasks="tasks" @task-saved="fetchTasks" />
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAsyncData } from '#imports'
import { showNotify, showDialog } from 'vant'
import 'vant/es/dialog/style'
import 'vant/es/notify/style'
import {
  getAllSchedulerTasks,
  getSchedulerTaskDetail,
  executeSchedulerTask,
  getTaskHistory,
  setTaskEnabled,
  deleteSchedulerTask,
  deleteSubTask
} from '~/utils/api'
import TaskForm from '../scheduler/TaskForm.vue'
import TaskDetail from '../scheduler/TaskDetail.vue'
import TaskHistory from '../scheduler/TaskHistory.vue'

const loading = ref(false)
const tasks = ref([])
const currentTask = ref(null)
const showTaskFormModal = ref(false)
const showTaskDetailModal = ref(false)
const showTaskHistoryModal = ref(false)
const isEditing = ref(false)
const parentTaskId = ref(null)

const getScheduleTypeLabel = (type) => {
  const map = { daily: '每日', chain: '链式', once: '一次性', interval: '间隔' }
  return map[type] || type || '-'
}

const getScheduleTimeDisplay = (task) => {
  if (task.config?.schedule_type === 'chain') return '依赖主任务'
  if (task.config?.schedule_type === 'interval') {
    const unit = { minutes: '分钟', hours: '小时', days: '天', months: '月', years: '年' }
    return `${task.config?.interval_value || '-'} ${unit[task.config?.interval_unit] || ''}`
  }
  return task.config?.schedule_time || '-'
}

const fetchTasks = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const response = await getAllSchedulerTasks({ include_subtasks: true, detail_level: 'full' })
    if (response.data?.status === 'success') {
      tasks.value = (response.data.tasks || []).map(t => ({ ...t, isExpanded: true }))
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取任务列表失败' })
  } finally {
    loading.value = false
  }
}

const executeTask = async (taskId) => {
  try {
    const response = await executeSchedulerTask(taskId, { wait_for_completion: false })
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: '任务执行已启动' })
      fetchTasks()
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '执行失败: ' + (error.response?.data?.message || error.message) })
  }
}

const toggleTaskEnabled = async (taskId, enabled) => {
  try {
    const response = await setTaskEnabled(taskId, enabled)
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: enabled ? '已启用' : '已禁用' })
      fetchTasks()
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '操作失败' })
  }
}

const openTaskDetailModal = async (taskId) => {
  try {
    const response = await getSchedulerTaskDetail(taskId)
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      currentTask.value = response.data.tasks[0]
      showTaskDetailModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取详情失败' })
  }
}

const fetchTaskHistory = async (taskId) => {
  try {
    const response = await getTaskHistory({ task_id: taskId, include_subtasks: true, page: 1, page_size: 20 })
    if (response.data?.status === 'success') {
      showTaskHistoryModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取历史失败' })
  }
}

const openEditTaskModal = async (taskId) => {
  try {
    isEditing.value = true
    parentTaskId.value = null
    const response = await getSchedulerTaskDetail(taskId)
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      currentTask.value = response.data.tasks[0]
      showTaskFormModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取任务详情失败' })
  }
}

const openCreateSubTaskModal = async (taskId) => {
  try {
    isEditing.value = false
    const response = await getSchedulerTaskDetail(taskId)
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      const mainTask = response.data.tasks[0]
      parentTaskId.value = taskId
      if (mainTask.sub_tasks?.length > 0) {
        const lastSub = mainTask.sub_tasks[mainTask.sub_tasks.length - 1]
        currentTask.value = { ...mainTask, depends_on: { task_id: lastSub.task_id, name: lastSub.config?.name || lastSub.task_id } }
      } else {
        currentTask.value = mainTask
      }
      showTaskFormModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取主任务详情失败' })
  }
}

const openCreateTaskModal = () => {
  isEditing.value = false
  parentTaskId.value = null
  currentTask.value = null
  showTaskFormModal.value = true
}

const confirmDeleteTask = (taskId, parentTaskId = null) => {
  showDialog({
    title: '确认删除',
    message: parentTaskId ? '确定删除此子任务吗？' : '确定删除此任务及其所有子任务吗？',
    showCancelButton: true,
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    confirmButtonColor: '#fb7299',
  }).then(() => deleteTask(taskId, parentTaskId))
}

const deleteTask = async (taskId, parentTaskId = null) => {
  try {
    const response = parentTaskId ? await deleteSubTask(parentTaskId, taskId) : await deleteSchedulerTask(taskId)
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: '删除成功' })
      showTaskDetailModal.value = false
      fetchTasks()
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '删除失败' })
  }
}

const { data: initialData } = await useAsyncData('scheduler-initial', async () => {
  try {
    const response = await getAllSchedulerTasks({ include_subtasks: true, detail_level: 'full' })
    if (response.data?.status === 'success') {
      return { tasks: (response.data.tasks || []).map(t => ({ ...t, isExpanded: true })) }
    }
  } catch {}
  return { tasks: [] }
})

if (initialData.value?.tasks) tasks.value = initialData.value.tasks

onMounted(() => { if (tasks.value.length === 0) fetchTasks() })
</script>

<style scoped>
:deep(.van-dialog__content) {
  max-height: 70vh;
  overflow-y: auto;
}
</style>
