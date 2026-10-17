<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[9999] flex items-center justify-center">
      <div class="fixed inset-0 bg-black/60" @click="cancel" />
      <div
        class="relative z-10 mx-4 flex max-h-[90vh] w-full max-w-lg flex-col rounded-2xl bg-white shadow-2xl dark:bg-gray-800"
      >
        <!-- 标题栏 -->
        <div
          class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-gray-700"
        >
          <div class="flex items-center gap-2.5">
            <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-accent/10">
              <svg
                class="h-4 w-4 text-accent"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  v-if="isEditing"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10"
                />
                <path v-else stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
              </svg>
            </div>
            <div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">
                {{ isEditing ? '编辑任务' : parentTaskId ? '添加子任务' : '新建任务' }}
              </h3>
              <p class="text-[0.6875rem] text-gray-500 dark:text-gray-400">
                {{
                  isEditing ? '修改任务配置' : parentTaskId ? '创建链式子任务' : '配置定时执行任务'
                }}
              </p>
            </div>
          </div>
          <button
            class="flex h-7 w-7 items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
            @click="cancel"
          >
            <svg
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- 表单内容 -->
        <form class="flex-1 space-y-4 overflow-y-auto px-5 py-4" @submit.prevent="submitForm">
          <!-- 基本信息 -->
          <div>
            <h4
              class="mb-2 text-xs font-semibold tracking-wider text-gray-500 uppercase dark:text-gray-400"
            >
              基本信息
            </h4>
            <div class="space-y-3">
              <div>
                <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                  >任务 ID</label
                >
                <input
                  v-model="form.task_id"
                  type="text"
                  :disabled="isEditing"
                  class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
                  placeholder="选择端点后自动填充"
                  required
                />
              </div>
              <div>
                <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                  >任务名称 <span class="text-accent">*</span></label
                >
                <input
                  v-model="form.name"
                  type="text"
                  class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
                  placeholder="例如：获取B站历史记录"
                  required
                />
              </div>
            </div>
          </div>

          <!-- API 设置 -->
          <div>
            <h4
              class="mb-2 text-xs font-semibold tracking-wider text-gray-500 uppercase dark:text-gray-400"
            >
              API 设置
            </h4>
            <div class="space-y-3">
              <div>
                <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                  >API 端点 <span class="text-accent">*</span></label
                >
                <button
                  type="button"
                  class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-left text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
                  @click="showApiSelector = true"
                >
                  {{ form.endpoint || '选择 API 端点...' }}
                </button>
              </div>
              <div>
                <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                  >请求方法</label
                >
                <input
                  v-model="form.method"
                  type="text"
                  disabled
                  class="w-full rounded-xl border border-gray-200 bg-gray-100 px-3 py-2 text-sm text-gray-500 dark:border-gray-600 dark:bg-gray-700/30 dark:text-gray-400"
                />
              </div>
            </div>
          </div>

          <!-- 热门视频清理设置 -->
          <div v-if="isPopularCleanupEndpoint">
            <h4
              class="mb-2 text-xs font-semibold tracking-wider text-gray-500 uppercase dark:text-gray-400"
            >
              热门视频清理
            </h4>
            <div>
              <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                >清理年份</label
              >
              <select
                v-model="popularCleanupYear"
                :disabled="popularCleanupYearsLoading"
                class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
              >
                <option :value="null">全部年份</option>
                <option v-for="year in popularCleanupYearOptions" :key="year" :value="year">
                  {{ year }}
                </option>
              </select>
              <p v-if="popularCleanupDefaultYear" class="mt-1 text-[0.6875rem] text-gray-400">
                建议：{{ popularCleanupDefaultYear }}
              </p>
            </div>
          </div>

          <!-- 调度设置 -->
          <div>
            <h4
              class="mb-2 text-xs font-semibold tracking-wider text-gray-500 uppercase dark:text-gray-400"
            >
              调度设置
            </h4>
            <div class="space-y-3">
              <div>
                <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                  >调度类型 <span class="text-accent">*</span></label
                >
                <select
                  v-model="form.schedule_type"
                  :disabled="!!parentTaskId"
                  class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
                  required
                >
                  <option v-if="!parentTaskId" value="daily">每日执行</option>
                  <option v-if="!parentTaskId" value="interval">间隔执行</option>
                  <option value="chain">链式依赖</option>
                </select>
              </div>
              <div v-if="form.schedule_type === 'daily' && !parentTaskId">
                <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                  >执行时间 <span class="text-accent">*</span></label
                >
                <input
                  v-model="form.schedule_time"
                  type="time"
                  class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
                  required
                />
              </div>
              <div
                v-if="form.schedule_type === 'interval' && !parentTaskId"
                class="grid grid-cols-2 gap-3"
              >
                <div>
                  <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                    >间隔 <span class="text-accent">*</span></label
                  >
                  <input
                    v-model.number="form.interval"
                    type="number"
                    min="1"
                    class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
                    required
                  />
                </div>
                <div>
                  <label class="mb-1 block text-xs font-medium text-gray-700 dark:text-gray-300"
                    >单位 <span class="text-accent">*</span></label
                  >
                  <select
                    v-model="form.unit"
                    class="w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
                    required
                  >
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
            <h4
              class="mb-2 text-xs font-semibold tracking-wider text-gray-500 uppercase dark:text-gray-400"
            >
              依赖任务
            </h4>
            <button
              type="button"
              :disabled="!!parentTaskId || isEditing"
              class="min-h-[2.5rem] w-full rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 text-left text-sm text-gray-900 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700/50 dark:text-gray-100"
              @click="showDependencySelector = true"
            >
              <div v-if="form.depends_on.length === 0" class="text-gray-400">
                {{ parentTaskId ? '自动依赖父任务' : isEditing ? '不可修改' : '选择依赖任务...' }}
              </div>
              <div v-else class="flex flex-wrap gap-1">
                <span
                  v-for="depId in form.depends_on"
                  :key="depId"
                  class="inline-flex items-center gap-1 rounded-full bg-accent/10 px-2 py-0.5 text-xs text-accent"
                >
                  {{ getTaskName(depId) }}
                  <button
                    v-if="!parentTaskId && !isEditing"
                    type="button"
                    class="hover:text-accent/70"
                    @click.stop="removeTask(depId)"
                  >
                    <svg
                      class="h-3 w-3"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </span>
              </div>
            </button>
          </div>
        </form>

        <!-- 底部按钮 -->
        <div
          class="flex items-center justify-end gap-2 border-t border-gray-100 px-5 py-3 dark:border-gray-700"
        >
          <button
            type="button"
            class="rounded-xl bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
            @click="cancel"
          >
            取消
          </button>
          <button
            type="submit"
            class="rounded-xl bg-accent px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-accent/90"
            @click="submitForm"
          >
            {{ isEditing ? '保存修改' : '创建任务' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- API选择弹窗 -->
  <select-dialog
    v-model:show="showApiSelector"
    v-model:selected="selectedEndpoint"
    title="选择API端点"
    :items="availableEndpoints"
    :show-method-filter="true"
    search-placeholder="搜索API端点..."
    group-by="tags"
    id-field="id"
    name-field="name"
    description-field="description"
    @select="handleApiSelect"
  />

  <!-- 依赖任务选择弹窗 -->
  <select-dialog
    v-model:show="showDependencySelector"
    v-model:selected="form.depends_on"
    title="选择依赖任务"
    :items="availableEndpoints"
    :multiple="false"
    :show-method-filter="true"
    search-placeholder="搜索任务..."
    group-by="tags"
    id-field="operationId"
    name-field="summary"
    description-field="path"
  />
</template>

<script setup>
import { ref, computed, reactive, watch } from 'vue';
import { showNotify } from 'vant';
import 'vant/es/notify/style';
import SelectDialog from './SelectDialog.vue';
import {
  createSchedulerTask,
  updateSchedulerTask,
  getSchedulerTaskDetail,
  getAvailableEndpoints,
  addSubTask,
  getPopularCleanupYears
} from '~/utils/api';

const props = defineProps({
  show: { type: Boolean, default: false },
  isEditing: { type: Boolean, default: false },
  taskId: { type: String, default: '' },
  parentTaskId: { type: String, default: '' },
  tasks: { type: Array, default: () => [] },
  currentTask: { type: Object, default: () => null }
});

const emit = defineEmits(['update:show', 'task-saved']);

const form = reactive({
  task_id: '',
  name: '',
  endpoint: '',
  method: '',
  params: {},
  schedule_type: 'daily',
  schedule_time: '00:00',
  interval: 1,
  unit: 'hours',
  depends_on: [],
  enabled: true,
  sub_tasks: []
});

const showApiSelector = ref(false);
const showDependencySelector = ref(false);
const selectedEndpoint = ref('');
const availableEndpoints = ref([]);

const POPULAR_CLEANUP_ENDPOINT_PATH = '/bilibili/popular/cleanup';
const popularCleanupYears = ref([]);
const popularCleanupDefaultYear = ref(null);
const popularCleanupYearsLoading = ref(false);
const popularCleanupYearTouched = ref(false);

const parseEndpointUrlParts = (endpoint) => {
  const raw = (endpoint || '').trim();
  if (!raw) return { path: '', searchParams: new URLSearchParams() };
  try {
    const url = new URL(raw, 'http://placeholder');
    return { path: url.pathname || '', searchParams: url.searchParams || new URLSearchParams() };
  } catch {
    const [pathPart, queryPart] = raw.split('?');
    return { path: (pathPart || raw).trim(), searchParams: new URLSearchParams(queryPart || '') };
  }
};

const normalizeEndpointPath = (path) => {
  let normalized = (path || '').trim();
  if (!normalized) return '';
  if (!normalized.startsWith('/')) normalized = `/${normalized}`;
  return normalized.replace(/\/+$/, '') || '/';
};

const isPopularCleanupEndpoint = computed(() => {
  const normalized = normalizeEndpointPath(parseEndpointUrlParts(form.endpoint).path);
  return (
    normalized === POPULAR_CLEANUP_ENDPOINT_PATH ||
    normalized.endsWith(POPULAR_CLEANUP_ENDPOINT_PATH)
  );
});

const popularCleanupYear = computed({
  get: () => form.params?.year ?? null,
  set: (year) => {
    popularCleanupYearTouched.value = true;
    if (year === null || year === undefined || year === '') {
      delete form.params.year;
      return;
    }
    form.params.year = year;
  }
});

const popularCleanupYearOptions = computed(() => {
  const years = Array.isArray(popularCleanupYears.value) ? [...popularCleanupYears.value] : [];
  const current = form.params?.year;
  const n = typeof current === 'string' ? Number.parseInt(current, 10) : current;
  if (Number.isFinite(n) && !years.includes(n)) years.push(n);
  return years.filter((y) => Number.isFinite(y)).sort((a, b) => b - a);
});

const fetchPopularCleanupYears = async () => {
  if (popularCleanupYearsLoading.value) return;
  popularCleanupYearsLoading.value = true;
  try {
    const response = await getPopularCleanupYears();
    if (response.data?.status === 'success') {
      popularCleanupYears.value = Array.isArray(response.data.data) ? response.data.data : [];
      popularCleanupDefaultYear.value = response.data.default_year ?? null;
      if (
        isPopularCleanupEndpoint.value &&
        !props.isEditing &&
        !popularCleanupYearTouched.value &&
        !('year' in form.params) &&
        popularCleanupDefaultYear.value !== null
      ) {
        form.params.year = popularCleanupDefaultYear.value;
      }
    }
  } catch {
    // 可选年份列表获取失败时保持为空，仅影响默认年份预填
  }
  popularCleanupYearsLoading.value = false;
};

const getTaskName = (taskId) => {
  const task = props.tasks.find((t) => t.task_id === taskId);
  return task?.config?.name || taskId;
};

const removeTask = (taskId) => {
  form.depends_on = form.depends_on.filter((id) => id !== taskId);
};

const handleApiSelect = (endpoint) => {
  if (!endpoint) return;
  let endpointPath = endpoint.path || '';
  if (!endpointPath && endpoint.operationId) {
    const found = availableEndpoints.value.find((e) => e.operationId === endpoint.operationId);
    if (found && found.path) {
      endpointPath = found.path;
    }
  }
  if (!endpointPath && endpoint.id) {
    const found = availableEndpoints.value.find(
      (e) => e.operationId === endpoint.id || e.path === endpoint.id
    );
    if (found && found.path) {
      endpointPath = found.path;
    }
  }
  if (endpointPath && !endpointPath.startsWith('/')) {
    endpointPath = '/' + endpointPath;
  }
  form.endpoint = endpointPath;
  form.method = endpoint.method || 'GET';
  form.task_id = endpoint.operationId || endpoint.id || '';
  if (!form.name.trim()) {
    form.name = endpoint.name || endpoint.summary || '';
  }
};

const loadEndpoints = async () => {
  try {
    const response = await getAvailableEndpoints();
    if (response.data?.status === 'success' && Array.isArray(response.data.endpoints)) {
      availableEndpoints.value = response.data.endpoints;
    }
  } catch {
    // 端点目录获取失败时下拉保持为空，用户仍可手动输入
  }
};

const loadTaskData = async () => {
  if (!props.taskId) return;
  try {
    const response = await getSchedulerTaskDetail(props.taskId);
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      const task = response.data.tasks[0];
      form.task_id = task.task_id || '';
      form.name = task.config?.name || '';
      form.endpoint = task.config?.endpoint || '';
      form.method = task.config?.method || 'GET';
      form.params = task.config?.params || {};
      form.schedule_type = task.config?.schedule_type || 'daily';
      form.schedule_time = task.config?.schedule_time || '00:00';
      form.interval = task.config?.interval_value || 1;
      form.unit = task.config?.interval_unit || 'hours';
      form.depends_on = task.config?.depends_on || [];
      form.enabled = task.config?.enabled !== false;
    }
  } catch {
    // 任务详情加载失败时表单保持空白，由用户手动填写
  }
};

const validateForm = () => {
  if (!form.task_id.trim()) {
    showNotify({ type: 'warning', message: '请输入任务ID' });
    return false;
  }
  if (!form.name.trim()) {
    showNotify({ type: 'warning', message: '请输入任务名称' });
    return false;
  }
  if (!form.endpoint.trim()) {
    showNotify({ type: 'warning', message: '请选择API端点' });
    return false;
  }
  return true;
};

const submitForm = async () => {
  if (!validateForm()) return;

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
  };

  try {
    let response;
    if (props.isEditing) {
      response = await updateSchedulerTask(props.taskId, taskData);
    } else if (props.parentTaskId) {
      response = await addSubTask(props.parentTaskId, taskData);
    } else {
      response = await createSchedulerTask(taskData);
    }

    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: props.isEditing ? '修改成功' : '创建成功' });
      emit('update:show', false);
      emit('task-saved');
    } else {
      showNotify({ type: 'danger', message: response.data?.message || '操作失败' });
    }
  } catch (error) {
    showNotify({
      type: 'danger',
      message: error.response?.data?.message || error.message || '操作失败'
    });
  }
};

const cancel = () => {
  emit('update:show', false);
};

watch(
  () => props.show,
  async (val) => {
    if (val) {
      await loadEndpoints();
      if (props.isEditing && props.taskId) {
        await loadTaskData();
      } else {
        Object.assign(form, {
          task_id: '',
          name: '',
          endpoint: '',
          method: '',
          params: {},
          schedule_type: props.parentTaskId ? 'chain' : 'daily',
          schedule_time: '00:00',
          interval: 1,
          unit: 'hours',
          depends_on: props.parentTaskId ? [] : [],
          enabled: true
        });
      }
      if (isPopularCleanupEndpoint.value) fetchPopularCleanupYears();
    }
  }
);
</script>
