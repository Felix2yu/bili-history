<template>
  <div class="relative">
    <!-- 底部弹出式筛选栏 -->
    <VanPopup
      v-model:show="showFilterPopup"
      position="bottom"
      round
      :z-index="2000"
      get-container="body"
      teleport="body"
      :style="{ height: '80%', maxHeight: '640px' }"
      class="flex flex-col overflow-hidden"
    >
      <!-- 固定的抽屉头部 -->
      <div
        class="sticky top-0 z-10 flex shrink-0 items-center justify-between border-b border-gray-100/80 bg-white/95 px-5 py-3.5 backdrop-blur-sm dark:border-gray-800/80 dark:bg-gray-900/95"
      >
        <div class="flex items-center gap-2">
          <div class="h-4 w-1 rounded-full bg-accent" />
          <span class="text-sm font-semibold text-gray-800 dark:text-gray-100">高级筛选</span>
          <span
            v-if="activeFilterCount > 0"
            class="inline-flex h-[18px] min-w-[18px] items-center justify-center rounded-full bg-accent px-1 text-[0.625rem] font-bold text-white"
            >{{ activeFilterCount }}</span
          >
        </div>
        <button
          class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-100/80 transition-colors duration-200 hover:bg-gray-200 active:scale-95 dark:bg-gray-800/80 dark:hover:bg-gray-700"
          @click="closeFilterPopup"
        >
          <svg
            class="h-4 w-4 text-gray-500 dark:text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div class="w-full flex-1 overflow-y-auto px-5 py-4">
        <!-- 活跃筛选标签 -->
        <div v-if="activeFilterCount > 0" class="mb-4 animate-fade-in">
          <div class="mb-2 flex items-center justify-between">
            <span class="text-[0.6875rem] font-medium text-gray-500 dark:text-gray-400"
              >当前筛选</span
            >
            <button
              class="text-[0.6875rem] font-medium text-accent transition-all duration-200 hover:text-accent/80 active:scale-95"
              @click="clearAllFilters"
            >
              清除全部
            </button>
          </div>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-if="business"
              class="inline-flex cursor-pointer items-center gap-1 rounded-full bg-accent/8 px-2.5 py-1 text-[0.6875rem] font-medium text-accent transition-colors duration-200 hover:bg-accent/15 active:scale-95 dark:bg-accent/15 dark:hover:bg-accent/25"
              @click="clearBusiness"
            >
              {{ businessLabel || '条目类型' }}
              <svg class="h-3 w-3 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </span>
            <span
              v-if="date"
              class="inline-flex cursor-pointer items-center gap-1 rounded-full bg-accent/8 px-2.5 py-1 text-[0.6875rem] font-medium text-accent transition-colors duration-200 hover:bg-accent/15 active:scale-95 dark:bg-accent/15 dark:hover:bg-accent/25"
              @click="clearDate"
            >
              日期区间
              <svg class="h-3 w-3 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </span>
            <span
              v-if="category"
              class="inline-flex cursor-pointer items-center gap-1 rounded-full bg-accent/8 px-2.5 py-1 text-[0.6875rem] font-medium text-accent transition-colors duration-200 hover:bg-accent/15 active:scale-95 dark:bg-accent/15 dark:hover:bg-accent/25"
              @click="clearCategory"
            >
              {{ category }}
              <svg class="h-3 w-3 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </span>
          </div>
        </div>

        <!-- 条目类型筛选 -->
        <div class="mb-5">
          <div class="mb-2.5 flex items-center justify-between">
            <h4
              class="flex items-center gap-1.5 text-[0.8125rem] font-semibold text-gray-800 dark:text-gray-200"
            >
              <svg
                class="h-3.5 w-3.5 text-accent"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z"
                />
              </svg>
              条目类型
            </h4>
            <button
              v-if="business"
              class="text-[0.6875rem] font-medium text-accent transition-all duration-200 active:scale-95"
              @click="clearBusiness"
            >
              重置
            </button>
          </div>

          <div class="flex flex-wrap gap-2">
            <div
              v-for="(label, type) in businessTypeMap"
              :key="type"
              class="flex cursor-pointer items-center justify-center rounded-full border px-3.5 py-1.5 transition-all duration-200 active:scale-95"
              :class="
                business === type
                  ? 'border-accent bg-accent text-white shadow-sm shadow-accent/20'
                  : 'border-gray-200/60 bg-white/50 text-gray-600 hover:border-gray-300 hover:bg-white/80 dark:border-gray-700/60 dark:bg-gray-800/50 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:bg-gray-700/50'
              "
              @click="selectBusinessFromPopup(type)"
            >
              <span class="text-[0.75rem] font-medium">{{ label }}</span>
            </div>
          </div>
        </div>

        <!-- 分隔线 -->
        <div class="mb-5 h-px bg-gray-100 dark:bg-gray-800/80" />

        <!-- 日期筛选 -->
        <div class="mb-5">
          <div class="mb-2.5 flex items-center justify-between">
            <h4
              class="flex items-center gap-1.5 text-[0.8125rem] font-semibold text-gray-800 dark:text-gray-200"
            >
              <svg
                class="h-3.5 w-3.5 text-accent"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
              日期区间
            </h4>
            <button
              v-if="date"
              class="text-[0.6875rem] font-medium text-accent transition-all duration-200 active:scale-95"
              @click="clearDate"
            >
              重置
            </button>
          </div>

          <div class="flex items-center gap-2">
            <div class="flex-1">
              <label class="mb-1 block pl-1 text-[0.625rem] text-gray-500 dark:text-gray-400"
                >开始日期</label
              >
              <input
                v-model="startDate"
                type="date"
                class="w-full cursor-pointer rounded-xl border border-gray-200/60 bg-white/50 px-3 py-2 text-[0.75rem] text-gray-800 backdrop-blur-sm transition-all duration-200 focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none dark:border-gray-700/60 dark:bg-gray-800/50 dark:text-gray-200"
                :max="endDate || undefined"
                @change="onDateChange"
              />
            </div>
            <div class="mt-5 flex-none">
              <span class="text-[0.75rem] font-semibold text-accent/60">至</span>
            </div>
            <div class="flex-1">
              <label class="mb-1 block pl-1 text-[0.625rem] text-gray-500 dark:text-gray-400"
                >结束日期</label
              >
              <input
                v-model="endDate"
                type="date"
                class="w-full cursor-pointer rounded-xl border border-gray-200/60 bg-white/50 px-3 py-2 text-[0.75rem] text-gray-800 backdrop-blur-sm transition-all duration-200 focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none dark:border-gray-700/60 dark:bg-gray-800/50 dark:text-gray-200"
                :min="startDate || undefined"
                @change="onDateChange"
              />
            </div>
          </div>
        </div>

        <!-- 分隔线 -->
        <div class="mb-5 h-px bg-gray-100 dark:bg-gray-800/80" />

        <!-- 视频分区筛选 -->
        <div class="pb-4">
          <div class="mb-2.5 flex items-center justify-between">
            <h4
              class="flex items-center gap-1.5 text-[0.8125rem] font-semibold text-gray-800 dark:text-gray-200"
            >
              <svg
                class="h-3.5 w-3.5 text-accent"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                />
              </svg>
              视频分区
            </h4>
            <button
              v-if="category"
              class="text-[0.6875rem] font-medium text-accent transition-all duration-200 active:scale-95"
              @click="clearCategory"
            >
              重置
            </button>
          </div>

          <!-- 分区选择器 -->
          <div
            class="overflow-hidden rounded-2xl border border-gray-200/60 bg-white/50 shadow-sm backdrop-blur-sm dark:border-gray-700/60 dark:bg-gray-800/50"
          >
            <div class="flex h-[200px] flex-row md:h-[240px]">
              <!-- 主分区选择 -->
              <div
                class="w-2/5 scrollbar-thin overflow-y-auto border-r border-gray-100/80 bg-gray-50/50 dark:border-gray-800/60 dark:bg-gray-800/30"
              >
                <div
                  v-for="(categoryItem, index) in videoCategories"
                  :key="categoryItem.text"
                  class="relative cursor-pointer truncate px-3 py-2.5 text-[0.75rem] transition-all duration-200"
                  :class="
                    activeMainCategory === index
                      ? 'bg-white font-medium text-accent shadow-sm dark:bg-gray-900'
                      : 'text-gray-600 hover:bg-white/50 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-700/30 dark:hover:text-gray-200'
                  "
                  @click="activeMainCategory = index"
                >
                  <div
                    v-if="activeMainCategory === index"
                    class="absolute top-1 bottom-1 left-0 w-[3px] rounded-r-full bg-accent"
                  />
                  <span class="block truncate pl-1">{{ categoryItem.text }}</span>
                </div>
              </div>

              <!-- 子分区选择 -->
              <div class="w-3/5 scrollbar-thin overflow-y-auto p-2.5 pl-3">
                <div class="grid grid-cols-2 gap-1.5">
                  <!-- 主分区选项 -->
                  <div
                    class="cursor-pointer truncate rounded-full border px-2 py-1.5 text-center text-[0.6875rem] transition-all duration-200 active:scale-95"
                    :class="
                      category === videoCategories[activeMainCategory]?.text
                        ? 'border-accent bg-accent font-medium text-white shadow-sm shadow-accent/20'
                        : 'border-gray-200/60 bg-white/50 text-gray-700 hover:border-gray-300 hover:bg-white/80 dark:border-gray-700/60 dark:bg-gray-800/50 dark:text-gray-300 dark:hover:border-gray-600 dark:hover:bg-gray-700/50'
                    "
                    @click="
                      selectVideoCategory({
                        text: videoCategories[activeMainCategory]?.text,
                        type: 'main'
                      })
                    "
                  >
                    全部 {{ videoCategories[activeMainCategory]?.text || '频道' }}
                  </div>

                  <!-- 子分区选项 -->
                  <div
                    v-for="subCategory in videoCategories[activeMainCategory]?.children"
                    :key="subCategory.id"
                    class="cursor-pointer truncate rounded-full border px-2 py-1.5 text-center text-[0.6875rem] transition-all duration-200 active:scale-95"
                    :class="
                      category === subCategory.text
                        ? 'border-accent bg-accent font-medium text-white shadow-sm shadow-accent/20'
                        : 'border-gray-200/60 bg-white/50 text-gray-700 hover:border-gray-300 hover:bg-white/80 dark:border-gray-700/60 dark:bg-gray-800/50 dark:text-gray-300 dark:hover:border-gray-600 dark:hover:bg-gray-700/50'
                    "
                    @click="selectVideoCategory(subCategory)"
                  >
                    {{ subCategory.text }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </VanPopup>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { showNotify, Popup as VanPopup } from 'vant';
import 'vant/es/popup/style';
import 'vant/es/notify/style';

const props = defineProps({
  business: {
    type: String,
    default: ''
  },
  businessLabel: {
    type: String,
    default: ''
  },
  date: {
    type: String,
    default: ''
  },
  category: {
    type: String,
    default: ''
  },
  total: {
    type: Number,
    default: 0
  },
  pageSize: {
    type: Number,
    default: 30
  }
});

const emit = defineEmits([
  'update:business',
  'update:businessLabel',
  'update:date',
  'update:category',
  'update:category-type',
  'update:page',
  'update:pageSize',
  'refresh-data'
]);

// 底部弹出筛选栏的显示状态
const showFilterPopup = ref(false);

// 供父组件控制的弹窗开关
const openFilterPopup = () => {
  showFilterPopup.value = true;
};

const closeFilterPopup = () => {
  showFilterPopup.value = false;
};

// 活跃筛选数量
const activeFilterCount = computed(() => {
  let count = 0;
  if (props.business) count++;
  if (props.date) count++;
  if (props.category) count++;
  return count;
});

// 清除全部筛选
const clearAllFilters = () => {
  if (props.business) clearBusiness();
  if (props.date) clearDate();
  if (props.category) clearCategory();
};

// 日期选择相关
const startDate = ref('');
const endDate = ref('');

// 视频分区选择相关
const videoCategories = ref([]);
const activeMainCategory = ref(0);

// 获取视频分类
const fetchVideoCategories = async () => {
  try {
    const { getVideoCategories } = await import('~/utils/api');
    const response = await getVideoCategories();
    if (response.data.status === 'success') {
      videoCategories.value = response.data.data.map((category) => ({
        text: category.name,
        type: 'main',
        children: category.sub_categories.map((sub) => ({
          text: sub.name,
          id: sub.tid,
          type: 'sub'
        }))
      }));
    }
  } catch (error) {
    console.error('获取视频分类失败:', error);
  }
};

// 选择视频分区
const selectVideoCategory = (item) => {
  const isMainName = videoCategories.value.some(
    (cat) => cat.text === item.text && item.type === 'main'
  );

  let categoryText = '';
  let categoryType = ''; // main 或 sub
  if (item.type === 'main' || (item.type === 'sub' && isMainName)) {
    categoryText = item.text;
    categoryType = 'main';
  } else if (item.type === 'sub') {
    categoryText = item.text;
    categoryType = 'sub';
  }

  emit('update:category', categoryText);
  emit('update:category-type', categoryType);
  emit('update:page', 1);

  showNotify({
    type: 'success',
    message: `已筛选分区: ${categoryText || '全部'}`,
    duration: 1000
  });
};

// 监听日期属性变化，解析为开始和结束日期
watch(
  () => props.date,
  (newDate) => {
    if (newDate) {
      const dates = newDate.split(' 至 ');
      if (dates.length === 2) {
        startDate.value = formatDateForInput(dates[0]);
        endDate.value = formatDateForInput(dates[1]);
      }
    } else {
      startDate.value = '';
      endDate.value = '';
    }
  },
  { immediate: true }
);

// 格式化日期为输入框格式 (YYYY-MM-DD)
const formatDateForInput = (dateStr) => {
  try {
    const parts = dateStr.split('/');
    if (parts.length === 3) {
      return `${parts[0]}-${parts[1].padStart(2, '0')}-${parts[2].padStart(2, '0')}`;
    }
    return '';
  } catch {
    return '';
  }
};

// 格式化日期为显示格式 (YYYY-MM-DD)
const formatDateForDisplay = (dateStr) => {
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return '';
    const pad = (n) => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
  } catch (e) {
    console.error('日期格式化错误:', e);
    return '';
  }
};

// 业务类型映射表
const businessTypeMap = {
  '': '全部'
};

// 从弹出窗口选择业务类型
const selectBusinessFromPopup = (type) => {
  emit('update:business', type);
  emit('update:businessLabel', businessTypeMap[type]);
  emit('update:page', 1);

  showNotify({
    type: 'success',
    message: `已切换到${businessTypeMap[type]}`,
    duration: 1000
  });
};

// 应用日期筛选
const applyDateFilter = () => {
  if (startDate.value && endDate.value) {
    const formattedStartDate = formatDateForDisplay(startDate.value);
    const formattedEndDate = formatDateForDisplay(endDate.value);

    if (formattedStartDate && formattedEndDate) {
      const dateRange = `${formattedStartDate} 至 ${formattedEndDate}`;
      emit('update:date', dateRange);
      emit('update:page', 1);

      showNotify({
        type: 'success',
        message: `已筛选日期: ${dateRange}`,
        duration: 1000
      });
    } else {
      showNotify({
        type: 'warning',
        message: '日期格式无效',
        duration: 2000
      });
    }
  } else if (!startDate.value && !endDate.value) {
    emit('update:date', '');
    emit('update:page', 1);
  } else {
    showNotify({
      type: 'warning',
      message: '请同时设置开始和结束日期',
      duration: 2000
    });
  }
};

// 处理日期变化
const onDateChange = () => {
  applyDateFilter();
};

// 清除分区
const clearCategory = () => {
  emit('update:category', '');
  emit('update:page', 1);

  showNotify({
    type: 'success',
    message: '已清除分区筛选',
    duration: 1000
  });
};

// 清除日期
const clearDate = () => {
  emit('update:date', '');
  emit('update:page', 1);

  showNotify({
    type: 'success',
    message: '已清除日期筛选',
    duration: 1000
  });
};

// 清除业务类型
const clearBusiness = () => {
  emit('update:business', '');
  emit('update:businessLabel', '');
  emit('update:page', 1);

  showNotify({
    type: 'success',
    message: '已清除业务类型筛选',
    duration: 1000
  });
};

// 组件挂载时获取视频分类
onMounted(() => {
  fetchVideoCategories();
});

// 暴露控制方法，便于导航栏触发筛选面板
defineExpose({
  openFilterPopup,
  closeFilterPopup
});
</script>

<style scoped>
/* 确保日期输入框在移动设备上正常工作 */
input[type='date'] {
  -webkit-appearance: none;
  appearance: none;
  position: relative;
}

input[type='date']::-webkit-calendar-picker-indicator {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
}

/* 自定义滚动条 */
.scrollbar-thin::-webkit-scrollbar {
  width: 4px;
}

.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}

.scrollbar-thin::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 2px;
}

.dark .scrollbar-thin::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
}
</style>
