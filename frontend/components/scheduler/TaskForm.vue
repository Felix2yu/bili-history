<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[9999] flex items-center justify-center">
      <div class="fixed inset-0 bg-black/60" @click="cancel"></div>
      <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] z-10 flex flex-col mx-4">
        <!-- 标题栏 -->
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-700">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-xl bg-accent/10 flex items-center justify-center">
              <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round"
                  v-if="isEditing" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                <path stroke-linecap="round" stroke-linejoin="round" v-else d="M12 4v16m8-8H4" />
              </svg>
            </div>
            <div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">
                {{ isEditing ? '编辑任务' : (parentTaskId ? '添加子任务' : '新建任务') }}
              </h3>
              <p class="text-[11px] text-gray-500 dark:text-gray-400">
                {{ isEditing ? '修改任务配置' : (parentTaskId ? '创建链式子任务' : '配置定时执行任务') }}
              </p>
            </div>
          </div>
          <button @click="cancel"
            class="w-7 h-7 rounded-lg flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- 表单内容 -->
        <form @submit.prevent="submitForm" class="flex-1 overflow-y-auto px-5 py-4 space-y-4">
          <!-- 基本信息 -->
          <div>
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">基本信息</h4>
            <div class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">任务 ID</label>
                <input v-model="form.task_id" type="text" :disabled="isEditing"
                  class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent disabled:opacity-50 transition-all"
                  placeholder="选择端点后自动填充" required />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">任务名称 <span class="text-accent">*</span></label>
                <input v-model="form.name" type="text"
                  class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all"
                  placeholder="例如：获取B站历史记录" required />
              </div>
            </div>
          </div>

          <!-- API 设置 -->
          <div>
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">API 设置</h4>
            <div class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">API 端点 <span class="text-accent">*</span></label>
                <button type="button" @click="showApiSelector = true"
                  class="w-full text-left px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all">
                  {{ form.endpoint || '选择 API 端点...' }}
                </button>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">请求方法</label>
                <input v-model="form.method" type="text" disabled
                  class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-100 dark:bg-gray-700/30 text-gray-500 dark:text-gray-400" />
              </div>
            </div>
          </div>

          <!-- 热门视频清理设置 -->
          <div v-if="isPopularCleanupEndpoint">
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">热门视频清理</h4>
            <div>
              <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">清理年份</label>
              <select v-model="popularCleanupYear" :disabled="popularCleanupYearsLoading"
                class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all disabled:opacity-50">
                <option :value="null">全部年份</option>
                <option v-for="year in popularCleanupYearOptions" :key="year" :value="year">{{ year }}</option>
              </select>
              <p v-if="popularCleanupDefaultYear" class="text-[11px] text-gray-400 mt-1">建议：{{ popularCleanupDefaultYear }}</p>
            </div>
          </div>

          <!-- 调度设置 -->
          <div>
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">调度设置</h4>
            <div class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">调度类型 <span class="text-accent">*</span></label>
                <select v-model="form.schedule_type" :disabled="!!parentTaskId"
                  class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all disabled:opacity-50" required>
                  <option v-if="!parentTaskId" value="daily">每日执行</option>
                  <option v-if="!parentTaskId" value="interval">间隔执行</option>
                  <option value="chain">链式依赖</option>
                </select>
              </div>
              <div v-if="form.schedule_type === 'daily' && !parentTaskId">
                <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">执行时间 <span class="text-accent">*</span></label>
                <input v-model="form.schedule_time" type="time"
                  class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all" required />
              </div>
              <div v-if="form.schedule_type === 'interval' && !parentTaskId" class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">间隔 <span class="text-accent">*</span></label>
                  <input v-model.number="form.interval" type="number" min="1"
                    class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all" required />
                </div>
                <div>
                  <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">单位 <span class="text-accent">*</span></label>
                  <select v-model="form.unit"
                    class="w-full px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all" required>
                    <option value="minutes">分钟</option>
                    <option value="hours">小时</option>
                    <option value="days">天</option>
                    <option value="months">月</option>
                    <option value="years">年</option>
                  </select>
                </div>
              </div>
            </div>
          </div>

          <!-- 依赖任务 -->
          <div>
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">依赖任务</h4>
            <button type="button" @click="showDependencySelector = true" :disabled="!!parentTaskId || isEditing"
              class="w-full text-left px-3 py-2 text-sm rounded-xl border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all disabled:opacity-50 min-h-[2.5rem]">
              <div v-if="form.depends_on.length === 0" class="text-gray-400">
                {{ parentTaskId ? '自动依赖父任务' : (isEditing ? '不可修改' : '选择依赖任务...') }}
              </div>
              <div v-else class="flex flex-wrap gap-1">
                <span v-for="taskId in form.depends_on" :key="taskId"
                  class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-accent/10 text-accent">
                  {{ getTaskName(taskId) }}
                  <button type="button" @click.stop="removeTask(taskId)" v-if="!parentTaskId && !isEditing"
                    class="hover:text-accent/70">
                    <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </span>
              </div>
            </button>
          </div>
        </form>

        <!-- 底部按钮 -->
        <div class="flex items-center justify-end gap-2 px-5 py-3 border-t border-gray-100 dark:border-gray-700">
          <button type="button" @click="cancel"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-xl hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors">
            取消
          </button>
          <button type="submit" @click="submitForm"
            class="px-4 py-2 text-sm font-medium text-white bg-accent rounded-xl hover:bg-accent/90 transition-colors shadow-sm">
            {{ isEditing ? '保存修改' : '创建任务' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- API选择弹窗 -->
  <select-dialog v-model:show="showApiSelector" title="选择API端点" :items="availableEndpoints"
    v-model:selected="selectedEndpoint" :show-method-filter="true" search-placeholder="搜索API端点..." group-by="tags"
    id-field="id" name-field="name" description-field="description" @select="handleApiSelect" />

  <!-- 依赖任务选择弹窗 -->
  <select-dialog v-model:show="showDependencySelector" title="选择依赖任务" :items="availableEndpoints"
    v-model:selected="form.depends_on" :multiple="false" :show-method-filter="true" search-placeholder="搜索任务..."
    group-by="tags" id-field="operationId" name-field="summary" description-field="path" />
</template>

<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { showNotify } from 'vant'
import 'vant/es/notify/style'
import SelectDialog from './SelectDialog.vue'
import {
  createSchedulerTask,
  updateSchedulerTask,
  getSchedulerTaskDetail,
  getAvailableEndpoints,
  addSubTask,
  getPopularCleanupYears
} from '~/utils/api'

const props = defineProps({
  show: { type: Boolean, default: false },
  isEditing: { type: Boolean, default: false },
  taskId: { type: String, default: '' },
  parentTaskId: { type: String, default: '' },
  tasks: { type: Array, default: () => [] },
  currentTask: { type: Object, default: () => null }
})

const emit = defineEmits(['update:show', 'task-saved'])

const form = reactive({
  task_id: '', name: '', endpoint: '', method: '', params: {},
  schedule_type: 'daily', schedule_time: '00:00', interval: 1, unit: 'hours',
  depends_on: [], enabled: true, sub_tasks: []
})

const showApiSelector = ref(false)
const showDependencySelector = ref(false)
const selectedEndpoint = ref('')
const availableEndpoints = ref([])

const POPULAR_CLEANUP_ENDPOINT_PATH = '/bilibili/popular/cleanup'
const popularCleanupYears = ref([])
const popularCleanupDefaultYear = ref(null)
const popularCleanupYearsMessage = ref('')
const popularCleanupYearsLoading = ref(false)
const popularCleanupYearTouched = ref(false)

const parseEndpointUrlParts = (endpoint) => {
  const raw = (endpoint || '').trim()
  if (!raw) return { path: '', searchParams: new URLSearchParams() }
  try {
    const url = new URL(raw, 'http://placeholder')
    return { path: url.pathname || '', searchParams: url.searchParams || new URLSearchParams() }
  } catch {
    const [pathPart, queryPart] = raw.split('?')
    return { path: (pathPart || raw).trim(), searchParams: new URLSearchParams(queryPart || '') }
  }
}

const normalizeEndpointPath = (path) => {
  let normalized = (path || '').trim()
  if (!normalized) return ''
  if (!normalized.startsWith('/')) normalized = `/${normalized}`
  return normalized.replace(/\/+$/, '') || '/'
}

const isPopularCleanupEndpoint = computed(() => {
  const normalized = normalizeEndpointPath(parseEndpointUrlParts(form.endpoint).path)
  return normalized === POPULAR_CLEANUP_ENDPOINT_PATH || normalized.endsWith(POPULAR_CLEANUP_ENDPOINT_PATH)
})

const popularCleanupYearFromEndpointQuery = computed(() => {
  const year = parseEndpointUrlParts(form.endpoint).searchParams.get('year')
  if (!year) return null
  const n = Number.parseInt(year, 10)
  return Number.isFinite(n) ? n : null
})

const popularCleanupYear = computed({
  get: () => form.params?.year ?? null,
  set: (year) => {
    popularCleanupYearTouched.value = true
    if (year === null || year === undefined || year === '') {
      delete form.params.year
      return
    }
    form.params.year = year
  }
})

const popularCleanupYearOptions = computed(() => {
  const years = Array.isArray(popularCleanupYears.value) ? [...popularCleanupYears.value] : []
  const current = form.params?.year
  const n = typeof current === 'string' ? Number.parseInt(current, 10) : current
  if (Number.isFinite(n) && !years.includes(n)) years.push(n)
  return years.filter(y => Number.isFinite(y)).sort((a, b) => b - a)
})

const fetchPopularCleanupYears = async () => {
  if (popularCleanupYearsLoading.value) return
  popularCleanupYearsLoading.value = true
  try {
    const response = await getPopularCleanupYears()
    if (response.data?.status === 'success') {
      popularCleanupYears.value = Array.isArray(response.data.data) ? response.data.data : []
      popularCleanupDefaultYear.value = response.data.default_year ?? null
      if (isPopularCleanupEndpoint.value && !props.isEditing && !popularCleanupYearTouched.value
        && !('year' in form.params) && popularCleanupDefaultYear.value !== null) {
        form.params.year = popularCleanupDefaultYear.value
      }
    }
  } catch { }
  popularCleanupYearsLoading.value = false
}

const getTaskName = (taskId) => {
  const task = props.tasks.find(t => t.task_id === taskId)
  return task?.config?.name || taskId
}

const removeTask = (taskId) => {
  form.depends_on = form.depends_on.filter(id => id !== taskId)
}

const handleApiSelect = (endpoint) => {
  if (!endpoint) return
  let endpointPath = endpoint.path || ''
  if (!endpointPath && endpoint.operationId) {
    const found = availableEndpoints.value.find(e => e.operationId === endpoint.operationId)
    if (found && found.path) {
      endpointPath = found.path
    }
  }
  if (!endpointPath && endpoint.id) {
    const found = availableEndpoints.value.find(e => e.operationId === endpoint.id || e.path === endpoint.id)
    if (found && found.path) {
      endpointPath = found.path
    }
  }
  if (endpointPath && !endpointPath.startsWith('/')) {
    endpointPath = '/' + endpointPath
  }
  form.endpoint = endpointPath
  form.method = endpoint.method || 'GET'
  form.task_id = endpoint.operationId || endpoint.id || ''
  if (!form.name.trim()) {
    form.name = endpoint.name || endpoint.summary || ''
  }
}

const loadEndpoints = async () => {
  try {
    const response = await getAvailableEndpoints()
    if (response.data?.status === 'success' && Array.isArray(response.data.endpoints)) {
      availableEndpoints.value = response.data.endpoints
    }
  } catch { }
}

const loadTaskData = async () => {
  if (!props.taskId) return
  try {
    const response = await getSchedulerTaskDetail(props.taskId)
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      const task = response.data.tasks[0]
      form.task_id = task.task_id || ''
      form.name = task.config?.name || ''
      form.endpoint = task.config?.endpoint || ''
      form.method = task.config?.method || 'GET'
      form.params = task.config?.params || {}
      form.schedule_type = task.config?.schedule_type || 'daily'
      form.schedule_time = task.config?.schedule_time || '00:00'
      form.interval = task.config?.interval_value || 1
      form.unit = task.config?.interval_unit || 'hours'
      form.depends_on = task.config?.depends_on || []
      form.enabled = task.config?.enabled !== false
    }
  } catch { }
}

const validateForm = () => {
  if (!form.task_id.trim()) { showNotify({ type: 'warning', message: '请输入任务ID' }); return false }
  if (!form.name.trim()) { showNotify({ type: 'warning', message: '请输入任务名称' }); return false }
  if (!form.endpoint.trim()) { showNotify({ type: 'warning', message: '请选择API端点' }); return false }
  return true
}

const submitForm = async () => {
  if (!validateForm()) return

  const taskData = {
    task_id: form.task_id.trim(),
    name: form.name.trim(),
    endpoint: form.endpoint.trim(),
    method: form.method,
    params: { ...form.params },
    schedule_type: form.schedule_type,
    schedule_time: form.schedule_time,
    interval_value: form.interval,
    interval_unit: form.unit,
    depends_on: form.depends_on,
    enabled: form.enabled
  }

  try {
    let response
    if (props.isEditing) {
      response = await updateSchedulerTask(props.taskId, taskData)
    } else if (props.parentTaskId) {
      response = await addSubTask(props.parentTaskId, taskData)
    } else {
      response = await createSchedulerTask(taskData)
    }

    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: props.isEditing ? '修改成功' : '创建成功' })
      emit('update:show', false)
      emit('task-saved')
    } else {
      showNotify({ type: 'danger', message: response.data?.message || '操作失败' })
    }
  } catch (error) {
    showNotify({ type: 'danger', message: error.response?.data?.message || error.message || '操作失败' })
  }
}

const cancel = () => {
  emit('update:show', false)
}

watch(() => props.show, async (val) => {
  if (val) {
    await loadEndpoints()
    if (props.isEditing && props.taskId) {
      await loadTaskData()
    } else {
      Object.assign(form, {
        task_id: '', name: '', endpoint: '', method: '', params: {},
        schedule_type: props.parentTaskId ? 'chain' : 'daily',
        schedule_time: '00:00', interval: 1, unit: 'hours',
        depends_on: props.parentTaskId ? [] : [], enabled: true
      })
    }
    if (isPopularCleanupEndpoint.value) fetchPopularCleanupYears()
  }
})
</script>
