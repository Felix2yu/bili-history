<template>
  <div class="relative min-h-screen overflow-hidden">
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute -top-32 -right-20 h-96 w-96 rounded-full bg-accent/5 blur-3xl" />
      <div class="absolute top-1/4 -left-32 h-80 w-80 rounded-full bg-accent/3 blur-3xl" />
    </div>

    <div class="relative z-10">
      <Navbar
        v-if="currentContent === 'history' && !showRemarks"
        v-model:date="date"
        v-model:category="category"
        v-model:business="business"
        v-model:business-label="businessLabel"
        :total="total"
        :layout="layout"
        :is-batch-mode="isBatchMode"
        :show-remarks="showRemarks"
        @refresh-data="refreshData"
        @click-date="show = true"
        @change-layout="layout = $event"
        @toggle-batch-mode="isBatchMode = !isBatchMode"
        @toggle-remarks="showRemarks = !showRemarks"
      />

      <div
        v-if="currentContent === 'history' && !showRemarks"
        class="mx-auto max-w-[1800px] px-3 py-4 transition-all duration-300 ease-in-out sm:px-4 md:py-6 lg:px-8"
      >
        <div
          class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white/70 shadow-sm backdrop-blur-md dark:border-gray-700/50 dark:bg-gray-800/60"
        >
          <div class="absolute inset-0 bg-accent/[0.03]" />
          <div class="absolute -top-10 -right-10 h-40 w-40 rounded-full bg-accent/10 blur-2xl" />
          <div class="absolute -bottom-10 -left-10 h-32 w-32 rounded-full bg-accent/5 blur-2xl" />

          <div class="relative p-3 md:p-4">
            <div class="flex items-center justify-between gap-2 md:gap-4">
              <div class="flex min-w-0 items-center gap-2 md:gap-3">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-accent shadow-md shadow-accent/20 md:h-10 md:w-10"
                >
                  <svg
                    class="h-4 w-4 text-white md:h-5 md:w-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                </div>
                <div class="min-w-0">
                  <div
                    class="text-[0.625rem] font-medium tracking-wider text-gray-500 uppercase md:text-xs dark:text-gray-400"
                  >
                    今日浏览
                  </div>
                  <div class="truncate text-sm font-bold text-gray-900 md:text-lg dark:text-white">
                    {{ currentDate.slice(0, 4) }}年{{ Number(currentDate.slice(4, 6)) }}月{{
                      Number(currentDate.slice(6, 8))
                    }}日
                  </div>
                </div>
              </div>

              <DatePagination
                :current-date="currentDate"
                :available-dates="availableDates"
                :record-count="recordCount"
                :compact="true"
                @date-change="handleDateChange"
              />

              <div class="flex shrink-0 items-center gap-3 md:gap-4">
                <div class="text-right">
                  <div class="text-[0.625rem] text-gray-500 md:text-xs dark:text-gray-400">
                    当前筛选
                  </div>
                  <div class="text-sm font-bold text-accent md:text-base">
                    {{ total }}
                    <span
                      class="text-[0.625rem] font-normal text-gray-500 md:text-xs dark:text-gray-400"
                      >条</span
                    >
                  </div>
                </div>
                <div class="hidden h-8 w-px bg-gray-200 md:block dark:bg-gray-700/50" />
                <div class="hidden text-right md:block">
                  <div class="text-[0.625rem] text-gray-500 md:text-xs dark:text-gray-400">
                    今日总计
                  </div>
                  <div class="text-sm font-bold text-gray-900 md:text-base dark:text-white">
                    {{ recordCount }}
                    <span
                      class="text-[0.625rem] font-normal text-gray-500 md:text-xs dark:text-gray-400"
                      >条</span
                    >
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mx-auto max-w-[1800px] px-3 pb-6 sm:px-4 md:pb-8 lg:px-8">
        <div class="transition-all duration-300 ease-in-out">
          <div>
            <HistoryContent
              v-if="currentContent === 'history' && !showRemarks"
              ref="historyContentRef"
              v-model:show="show"
              v-model:show-bottom="showBottom"
              :selected-year="selectedYear"
              :current-date="currentDate"
              :page="1"
              :page-size="9999"
              :layout="layout"
              :date="date"
              :category="category"
              :category-type="categoryType"
              :business="business"
              :is-batch-mode="isBatchMode"
              @update:total-pages="() => {}"
              @update:total="total = $event"
              @update:record-count="recordCount = $event"
              @update:date="date = $event"
              @update:category="category = $event"
              @update:category-type="categoryType = $event"
            />

            <Remarks v-else-if="showRemarks" />
            <Settings v-else-if="currentContent === 'settings'" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted, nextTick } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import Navbar from '../Navbar.vue';
import HistoryContent from '../HistoryContent.vue';
import DatePagination from '../DatePagination.vue';
import Settings from '../Settings.vue';
import Remarks from './Remarks.vue';
import { getHistoryDates } from '~/utils/api';

const props = defineProps({
  defaultShowRemarks: {
    type: Boolean,
    default: false
  }
});

const currentContent = ref('history');
const router = useRouter();
const route = useRoute();

function todayStr() {
  const now = new Date();
  return `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}`;
}

const currentDate = ref(route.params.date || todayStr());
const selectedYear = ref(new Date().getFullYear());
const show = ref(false);
const showBottom = ref(false);
const date = ref('');
const total = ref(0);
const recordCount = ref(0);
const category = ref('');
const categoryType = ref('');
const layout = ref(localStorage.getItem('defaultLayout') || 'grid');
const isBatchMode = ref(false);
const showRemarks = ref(props.defaultShowRemarks);
const business = ref('');
const businessLabel = ref('');
const availableDates = ref([]);

const historyContentRef = ref(null);

const refreshData = async () => {
  try {
    if (historyContentRef.value?.fetchHistoryByDateRange) {
      await historyContentRef.value.fetchHistoryByDateRange();
    }
    loadAvailableDates();
  } catch (error) {
    console.error('刷新数据失败:', error);
  }
};

const loadAvailableDates = async () => {
  try {
    const res = await getHistoryDates();
    if (res.data && res.data.status === 'success') {
      availableDates.value = res.data.data || [];
    }
  } catch (e) {
    console.error('加载日期列表失败:', e);
  }
};

const handleDateChange = (newDate) => {
  router.push(`/date/${newDate}`);
};

watch(
  () => route.path,
  (path) => {
    if (path === '/settings') {
      currentContent.value = 'settings';
      showRemarks.value = false;
    } else if (path === '/remarks') {
      currentContent.value = 'history';
      showRemarks.value = true;
    } else {
      currentContent.value = 'history';
      showRemarks.value = false;
    }
  },
  { immediate: true }
);

watch(
  [() => route.params.date, () => route.path],
  ([newDate, path], [, oldPath]) => {
    if (path === '/') {
      const today = todayStr();
      if (currentDate.value !== today) {
        currentDate.value = today;
        router.replace(`/date/${today}`);
        return;
      }
    } else if (newDate && newDate !== currentDate.value) {
      currentDate.value = newDate;
    }

    // 路由变化时由 HistoryContent 的 watch(currentDate) 触发刷新
    // 仅在组件已挂载后且非首次加载时手动刷新
    if (historyContentRef.value?.fetchHistoryByDateRange && oldPath !== undefined) {
      nextTick(() => {
        historyContentRef.value?.fetchHistoryByDateRange();
      });
    }
  },
  { immediate: true }
);

onMounted(() => {
  if (props.defaultShowRemarks || route.path === '/remarks') {
    showRemarks.value = true;
  }
  loadAvailableDates();
  window.addEventListener('layout-setting-changed', handleLayoutSettingChanged);
});

onUnmounted(() => {
  window.removeEventListener('layout-setting-changed', handleLayoutSettingChanged);
});

const handleLayoutSettingChanged = (event) => {
  if (event.detail && typeof event.detail.layout === 'string') {
    layout.value = event.detail.layout;
  }
};

watch(layout, (newLayout) => {
  localStorage.setItem('defaultLayout', newLayout);
  try {
    window.dispatchEvent(new CustomEvent('layout-changed', { detail: { layout: newLayout } }));
  } catch (error) {
    console.error('触发布局变更事件失败:', error);
  }
});
</script>

<style scoped>
@keyframes bounce-x {
  0%,
  100% {
    transform: translateX(0);
  }
  50% {
    transform: translateX(4px);
  }
}
.animate-bounce-x {
  animation: bounce-x 1.5s infinite;
}
</style>
