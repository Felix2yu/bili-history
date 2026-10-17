<template>
  <div class="relative" :style="containerStyle">
    <button
      ref="triggerBtn"
      type="button"
      :class="[
        'custom-dropdown-trigger flex w-full items-center justify-between rounded-md border border-gray-300 bg-white px-2 py-1.5 text-xs text-gray-800 transition-colors duration-200 hover:bg-gray-50 focus:border-accent focus:ring focus:ring-accent/20 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700',
        customClass,
        'overflow-hidden whitespace-nowrap'
      ]"
      @click.stop="toggleDropdown"
    >
      <slot name="trigger-content">
        <span class="mr-1 truncate">{{ selectedText }}</span>
      </slot>
      <svg
        class="h-3 w-3 flex-shrink-0 text-accent"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <!-- 下拉菜单 -->
    <div
      v-if="showDropdown"
      class="ring-opacity-5 fixed z-50 mt-1 max-w-[90vw] rounded-lg border border-gray-200 bg-white shadow-lg ring-1 ring-black focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:ring-white/10"
      :style="dropdownStyle"
      @click.stop
    >
      <div class="max-h-60 overflow-auto py-1">
        <slot>
          <button
            v-for="(option, index) in options"
            :key="index"
            class="flex w-full items-center px-2 py-1 text-left text-xs whitespace-nowrap transition-colors hover:bg-accent/5 hover:text-accent"
            :class="{ 'bg-accent/5 font-medium text-accent': modelValue === option.value }"
            @click.stop="selectOption(option.value)"
          >
            {{ option.label }}
          </button>
        </slot>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: ''
  },
  options: {
    type: Array,
    default: () => []
  },
  selectedText: {
    type: String,
    default: '请选择'
  },
  customClass: {
    type: String,
    default: ''
  },
  minWidth: {
    type: Number,
    default: 120
  },
  useFixedWidth: {
    type: Boolean,
    default: false
  },
  buttonWidth: {
    type: [String, Number],
    default: null
  }
});

const emit = defineEmits(['update:modelValue', 'change']);

// 计算容器样式（包括宽度控制）
const containerStyle = computed(() => {
  if (props.buttonWidth) {
    return {
      width: typeof props.buttonWidth === 'number' ? `${props.buttonWidth}px` : props.buttonWidth
    };
  }
  return {};
});

const triggerBtn = ref(null);
const showDropdown = ref(false);
const dropdownStyle = ref({});

// 切换下拉菜单
const toggleDropdown = () => {
  showDropdown.value = !showDropdown.value;

  if (showDropdown.value) {
    // 计算位置
    calculateDropdownPosition();
  }
};

// 选择选项
const selectOption = (value) => {
  emit('update:modelValue', value);
  emit('change', value);
  showDropdown.value = false;
};

// 计算下拉菜单位置
const calculateDropdownPosition = () => {
  setTimeout(() => {
    if (triggerBtn.value) {
      const rect = triggerBtn.value.getBoundingClientRect();
      // 使用按钮自身的宽度或指定的最小宽度，不自动放大
      const width = props.useFixedWidth ? rect.width : Math.max(rect.width, props.minWidth || 80);

      // 获取视口宽度，确保菜单不会超出边界
      const viewportWidth = window.innerWidth;
      let left = rect.left;

      // 如果下拉菜单右边缘超出视口，调整位置
      if (left + width > viewportWidth - 10) {
        left = Math.max(10, viewportWidth - width - 10);
      }

      dropdownStyle.value = {
        top: `${rect.bottom + window.scrollY}px`,
        left: `${left}px`,
        width: `${width}px`,
        maxWidth: viewportWidth - 20 + 'px' // 确保不会超过视口宽度
      };
    }
  }, 0);
};

// 点击外部关闭下拉菜单
const closeDropdown = (event) => {
  if (triggerBtn.value && !triggerBtn.value.contains(event.target)) {
    showDropdown.value = false;
  }
};

// 处理窗口大小变化
const handleResize = () => {
  if (showDropdown.value) {
    calculateDropdownPosition();
  }
};

// 处理滚动事件
const handleScroll = () => {
  if (showDropdown.value) {
    calculateDropdownPosition();
  }
};

onMounted(() => {
  document.addEventListener('click', closeDropdown);
  window.addEventListener('resize', handleResize);
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown);
  window.removeEventListener('resize', handleResize);
  window.removeEventListener('scroll', handleScroll);
});
</script>

<style scoped>
/* 任何自定义样式 */
</style>
