<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[9999] flex items-center justify-center p-4">
      <Transition name="fade">
        <div
          v-if="show"
          class="fixed inset-0 bg-black/60 backdrop-blur-sm"
          @click="$emit('update:show', false)"
        />
      </Transition>
      <Transition name="scale">
        <div
          v-if="show"
          class="relative z-10 flex max-h-[90vh] w-full max-w-2xl flex-col overflow-hidden rounded-2xl bg-white shadow-2xl dark:bg-gray-800"
        >
          <!-- 顶部装饰 -->
          <div class="h-1.5 bg-accent" />

          <!-- 标题栏 -->
          <div
            class="flex items-center justify-between border-b border-gray-100 px-5 py-4 md:px-6 md:py-5 dark:border-gray-700/50"
          >
            <div class="flex min-w-0 items-center gap-3">
              <div
                class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-accent/15 md:h-12 md:w-12"
              >
                <svg
                  class="h-5 w-5 text-accent md:h-6 md:w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <div class="min-w-0">
                <h3
                  class="truncate text-base font-semibold text-gray-900 md:text-lg dark:text-white"
                >
                  {{ taskName }}
                </h3>
                <p class="text-[0.6875rem] text-gray-500 md:text-xs dark:text-gray-400">
                  执行历史记录
                </p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button
                class="inline-flex items-center gap-1.5 rounded-xl px-3 py-2 text-xs font-medium text-accent transition-colors hover:bg-accent/10 md:px-3.5 md:text-sm"
                @click="refreshHistory"
              >
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99"
                  />
                </svg>
                刷新
              </button>
              <button
                class="flex h-9 w-9 items-center justify-center rounded-xl text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 md:h-10 md:w-10 dark:hover:bg-gray-700 dark:hover:text-gray-300"
                @click="$emit('update:show', false)"
              >
                <svg
                  class="h-4 w-4 md:h-5 md:w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          <!-- 内容 -->
          <div class="flex-1 overflow-y-auto px-5 py-4 md:px-6 md:py-5">
            <!-- 加载 -->
            <div v-if="loading" class="flex justify-center py-16">
              <div class="relative">
                <div class="h-10 w-10 rounded-full border-2 border-accent/20" />
                <div
                  class="absolute inset-0 h-10 w-10 animate-spin rounded-full border-2 border-accent border-t-transparent"
                />
              </div>
            </div>

            <!-- 空状态 -->
            <div v-else-if="!history.length" class="py-16 text-center">
              <div class="relative mx-auto mb-4 h-20 w-20">
                <div class="absolute inset-0 rounded-3xl bg-gray-100 dark:bg-gray-700" />
                <div
                  class="absolute inset-2 flex items-center justify-center rounded-3xl bg-white dark:bg-gray-800"
                >
                  <svg
                    class="h-10 w-10 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
              </div>
              <p class="text-sm text-gray-500 dark:text-gray-400">暂无执行历史</p>
            </div>

            <!-- 时间线历史列表 -->
            <div v-else class="relative">
              <!-- 时间线竖线 -->
              <div class="absolute top-2 bottom-2 left-4 w-0.5 bg-gray-200 dark:bg-gray-700" />

              <div class="space-y-4">
                <TransitionGroup name="timeline">
                  <div
                    v-for="(record, idx) in history"
                    :key="record.execution_id"
                    class="relative pl-12"
                  >
                    <!-- 时间线节点 -->
                    <div
                      class="absolute top-1 left-0 flex h-9 w-9 items-center justify-center rounded-full"
                      :class="getNodeBgClass(record.status)"
                    >
                      <div class="h-3 w-3 rounded-full" :class="getNodeDotClass(record.status)" />
                    </div>

                    <!-- 卡片 -->
                    <div
                      class="group relative rounded-2xl border transition-all duration-300 hover:shadow-md"
                      :class="getCardClass(record.status)"
                    >
                      <div class="p-4">
                        <!-- 顶部：状态 + 耗时 -->
                        <div class="mb-2 flex items-center justify-between">
                          <div class="flex items-center gap-2">
                            <span
                              class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-xs font-medium"
                              :class="getStatusClass(record.status)"
                            >
                              <span
                                class="h-1.5 w-1.5 rounded-full"
                                :class="getStatusDotClass(record.status)"
                              />
                              {{ statusLabel(record.status) }}
                            </span>
                            <span
                              v-if="idx === 0"
                              class="inline-flex items-center gap-1 rounded-lg bg-accent/10 px-2 py-0.5 text-[0.625rem] font-medium text-accent"
                            >
                              最新
                            </span>
                          </div>
                          <div
                            class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400"
                          >
                            <svg
                              class="h-3.5 w-3.5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                              />
                            </svg>
                            {{ record.duration?.toFixed(1) || 0 }}s
                          </div>
                        </div>

                        <!-- 时间 -->
                        <div
                          class="flex items-center gap-2 text-[0.6875rem] text-gray-500 md:text-xs dark:text-gray-400"
                        >
                          <svg
                            class="h-3.5 w-3.5"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            stroke-width="2"
                          >
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5"
                            />
                          </svg>
                          {{ formatTime(record.start_time) }}
                        </div>

                        <!-- 错误按钮 -->
                        <div
                          v-if="record.error"
                          class="mt-3 border-t border-gray-100 pt-3 dark:border-gray-700/50"
                        >
                          <button
                            class="inline-flex items-center gap-1.5 text-xs font-medium text-red-500 transition-colors hover:text-red-600 dark:hover:text-red-400"
                            @click="viewError(record)"
                          >
                            <svg
                              class="h-3.5 w-3.5"
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"
                              />
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
                <button
                  :disabled="currentPage <= 1"
                  class="inline-flex items-center gap-1 rounded-xl border border-gray-200 px-3.5 py-2 text-xs font-medium transition-colors hover:bg-gray-50 disabled:opacity-40 dark:border-gray-600 dark:hover:bg-gray-700"
                  @click="handlePageChange(currentPage - 1)"
                >
                  <svg
                    class="h-3.5 w-3.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M15.75 19.5L8.25 12l7.5-7.5"
                    />
                  </svg>
                  上一页
                </button>
                <span class="px-3 text-xs text-gray-500"
                  >{{ currentPage }} / {{ Math.ceil(total / pageSize) }}</span
                >
                <button
                  :disabled="currentPage >= Math.ceil(total / pageSize)"
                  class="inline-flex items-center gap-1 rounded-xl border border-gray-200 px-3.5 py-2 text-xs font-medium transition-colors hover:bg-gray-50 disabled:opacity-40 dark:border-gray-600 dark:hover:bg-gray-700"
                  @click="handlePageChange(currentPage + 1)"
                >
                  下一页
                  <svg
                    class="h-3.5 w-3.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8.25 4.5l7.5 7.5-7.5 7.5"
                    />
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
          <div
            v-if="showErrorDialog"
            class="fixed inset-0 z-[10000] flex items-center justify-center p-4"
          >
            <div
              class="fixed inset-0 bg-black/60 backdrop-blur-sm"
              @click="showErrorDialog = false"
            />
            <Transition name="scale">
              <div
                v-if="showErrorDialog"
                class="relative z-10 w-full max-w-lg overflow-hidden rounded-2xl bg-white shadow-2xl dark:bg-gray-800"
              >
                <div
                  class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-gray-700"
                >
                  <div class="flex items-center gap-2.5">
                    <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-red-500/20">
                      <svg
                        class="h-4 w-4 text-red-500"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"
                        />
                      </svg>
                    </div>
                    <h3 class="text-sm font-semibold text-gray-900 dark:text-white">错误详情</h3>
                  </div>
                  <button
                    class="flex h-8 w-8 items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
                    @click="showErrorDialog = false"
                  >
                    <svg
                      class="h-4 w-4"
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
                </div>
                <div class="p-5">
                  <pre
                    class="max-h-[50vh] overflow-y-auto rounded-2xl bg-red-50 p-4 font-mono text-xs break-all whitespace-pre-wrap text-red-600 dark:bg-red-900/20 dark:text-red-400"
                    >{{ selectedRecord?.error }}</pre>
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
import { ref, onMounted, watch } from 'vue';
import { showNotify } from 'vant';
import 'vant/es/notify/style';
import { getTaskHistory } from '~/utils/api';

const props = defineProps({
  show: { type: Boolean, default: false },
  taskId: { type: String, default: '' },
  taskName: { type: String, default: '' }
});

defineEmits(['update:show']);

const loading = ref(false);
const history = ref([]);
const showErrorDialog = ref(false);
const selectedRecord = ref(null);
const currentPage = ref(1);
const pageSize = ref(20);
const total = ref(0);

const fetchHistory = async () => {
  if (!props.taskId) return;
  loading.value = true;
  try {
    const response = await getTaskHistory({
      task_id: props.taskId,
      include_subtasks: false,
      page: currentPage.value,
      page_size: pageSize.value
    });
    if (response.data?.status === 'success') {
      history.value = response.data.history || [];
      total.value = response.data.total || 0;
    }
  } catch {
    showNotify({ type: 'danger', message: '获取历史记录失败' });
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page) => {
  currentPage.value = page;
  fetchHistory();
};

const refreshHistory = () => fetchHistory();

const viewError = (record) => {
  selectedRecord.value = record;
  showErrorDialog.value = true;
};

const statusLabel = (status) => {
  return (
    { success: '执行成功', error: '执行失败', running: '执行中', pending: '等待中' }[status] ||
    status
  );
};

const formatTime = (time) => {
  if (!time) return '-';
  return time.replace('T', ' ');
};

const getStatusClass = (status) => {
  const map = {
    success: 'bg-green-500/10 text-green-600 dark:text-green-400',
    running: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
    error: 'bg-red-500/10 text-red-600 dark:text-red-400',
    pending: 'bg-blue-500/10 text-blue-600 dark:text-blue-400'
  };
  return map[status] || 'bg-gray-500/10 text-gray-500';
};

const getStatusDotClass = (status) => {
  const map = {
    success: 'bg-green-500',
    running: 'bg-amber-500',
    error: 'bg-red-500',
    pending: 'bg-blue-500'
  };
  return map[status] || 'bg-gray-500';
};

const getNodeBgClass = (status) => {
  const map = {
    success: 'bg-green-500/20',
    running: 'bg-amber-500/20',
    error: 'bg-red-500/20',
    pending: 'bg-blue-500/20'
  };
  return map[status] || 'bg-gray-500/20';
};

const getNodeDotClass = (status) => {
  const map = {
    success: 'bg-green-500',
    running: 'bg-amber-500',
    error: 'bg-red-500',
    pending: 'bg-blue-500'
  };
  return map[status] || 'bg-gray-500';
};

const getCardClass = (status) => {
  const map = {
    success: 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700',
    running: 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700',
    error: 'bg-red-50/50 dark:bg-red-900/10 border-red-100 dark:border-red-900/30',
    pending: 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700'
  };
  return map[status] || 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700';
};

onMounted(() => {
  if (props.show && props.taskId) fetchHistory();
});
watch(
  () => props.show,
  (val) => {
    if (val && props.taskId) fetchHistory();
  }
);
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
