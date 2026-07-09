<template>
  <div class="glass-card mx-3 sm:mx-0 px-4 py-3">
    <!-- 月份导航 + 日期显示 -->
    <div class="flex items-center justify-between mb-3">
      <button
        @click="prevMonth"
        class="p-2 rounded-xl text-gray-500 dark:text-gray-400 hover:bg-accent/10 hover:text-accent transition-all duration-200"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <div class="flex items-center gap-2">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ currentYear }}年{{ currentMonth }}月
        </span>
        <span v-if="recordCount > 0" class="text-xs text-gray-400 dark:text-gray-500">
          {{ recordCount }} 条
        </span>
      </div>

      <button
        @click="nextMonth"
        :disabled="!canGoNext"
        class="p-2 rounded-xl text-gray-500 dark:text-gray-400 hover:bg-accent/10 hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-200"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <!-- 星期标题 -->
    <div class="grid grid-cols-7 mb-1">
      <div
        v-for="w in weekHeaders"
        :key="w"
        class="text-center text-xs font-medium text-gray-400 dark:text-gray-500 py-1"
      >
        {{ w }}
      </div>
    </div>

    <!-- 日期网格 -->
    <div class="grid grid-cols-7 gap-y-0.5">
      <div v-for="(cell, i) in calendarCells" :key="i" class="flex justify-center">
        <button
          v-if="cell.day"
          @click="cell.available && selectDate(cell.dateStr)"
          :disabled="!cell.available"
          :class="[
            'w-8 h-8 sm:w-9 sm:h-9 rounded-lg text-xs sm:text-sm flex items-center justify-center transition-all duration-150 relative',
            cell.isToday && cell.dateStr !== currentDate
              ? 'text-accent font-semibold'
              : cell.dateStr === currentDate
                ? 'bg-accent text-white font-medium shadow-md'
                : cell.available
                  ? 'text-gray-700 dark:text-gray-300 hover:bg-accent/10 cursor-pointer'
                  : 'text-gray-300 dark:text-gray-600 cursor-default'
          ]"
        >
          {{ cell.day }}
          <!-- 有数据的日期显示小圆点 -->
          <span
            v-if="cell.available && cell.dateStr !== currentDate"
            class="absolute bottom-0.5 w-1 h-1 rounded-full bg-accent/60"
          />
        </button>
      </div>
    </div>

    <!-- 快捷操作 -->
    <div v-if="currentDate !== today" class="flex justify-center mt-2 pt-2 border-t border-gray-100 dark:border-gray-700/50">
      <button
        @click="goToday"
        class="text-xs text-accent hover:text-accent/80 transition-colors duration-200 px-3 py-1 rounded-lg hover:bg-accent/10"
      >
        回到今天
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  currentDate: { type: String, required: true },
  availableDates: { type: Array, default: () => [] },
  recordCount: { type: Number, default: 0 },
})

const emit = defineEmits(['date-change'])

const weekHeaders = ['日', '一', '二', '三', '四', '五', '六']

const today = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}`
})

// 当前显示的月份（从 currentDate 解析）
const viewDate = ref(parseDate(props.currentDate) || new Date())

const currentYear = computed(() => viewDate.value.getFullYear())
const currentMonth = computed(() => viewDate.value.getMonth() + 1)

const canGoNext = computed(() => {
  const now = new Date()
  return !(viewDate.value.getFullYear() === now.getFullYear() && viewDate.value.getMonth() === now.getMonth())
})

// 日历网格数据
const calendarCells = computed(() => {
  const year = viewDate.value.getFullYear()
  const month = viewDate.value.getMonth()
  const firstDay = new Date(year, month, 1).getDay()
  const daysInMonth = new Date(year, month + 1, 0).getDate()

  const cells = []
  // 前面的空白
  for (let i = 0; i < firstDay; i++) {
    cells.push({ day: null, dateStr: '', available: false, isToday: false })
  }
  // 日期
  for (let d = 1; d <= daysInMonth; d++) {
    const dateStr = `${year}${String(month + 1).padStart(2, '0')}${String(d).padStart(2, '0')}`
    cells.push({
      day: d,
      dateStr,
      available: props.availableDates.includes(dateStr),
      isToday: dateStr === today.value,
    })
  }
  return cells
})

function prevMonth() {
  const d = new Date(viewDate.value)
  d.setMonth(d.getMonth() - 1)
  viewDate.value = d
}

function nextMonth() {
  if (!canGoNext.value) return
  const d = new Date(viewDate.value)
  d.setMonth(d.getMonth() + 1)
  viewDate.value = d
}

function selectDate(dateStr) {
  emit('date-change', dateStr)
}

function goToday() {
  viewDate.value = new Date()
  emit('date-change', today.value)
}

function parseDate(str) {
  if (!str || str.length !== 8) return null
  return new Date(parseInt(str.substring(0, 4)), parseInt(str.substring(4, 6)) - 1, parseInt(str.substring(6, 8)))
}
</script>
