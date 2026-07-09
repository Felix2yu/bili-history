<template>
  <div class="mx-auto max-w-4xl">
    <div class="glass-card px-4 py-3 flex items-center justify-between gap-2">
      <!-- 前一天 -->
      <button
        @click="goPrev"
        :disabled="!hasPrev"
        class="flex shrink-0 items-center gap-1 px-3 py-2 rounded-xl text-sm text-gray-600 dark:text-gray-400 hover:bg-accent/10 hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-200"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        <span class="hidden sm:inline">前一天</span>
      </button>

      <!-- 日期显示 & 日历按钮 -->
      <div class="flex min-w-0 flex-1 items-center justify-center gap-3">
        <button
          ref="dateBtnRef"
          @click="showCalendar = !showCalendar"
          class="flex items-center gap-2 px-3 py-1.5 rounded-xl text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-accent/10 transition-all duration-200 cursor-pointer"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <span>{{ formattedDate }}</span>
        </button>

        <!-- 记录数 -->
        <span v-if="recordCount > 0" class="text-xs text-gray-400 dark:text-gray-500 shrink-0">
          {{ recordCount }} 条
        </span>
      </div>

      <!-- 后一天 -->
      <button
        @click="goNext"
        :disabled="!hasNext"
        class="flex shrink-0 items-center gap-1 px-3 py-2 rounded-xl text-sm text-gray-600 dark:text-gray-400 hover:bg-accent/10 hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-200"
      >
        <span class="hidden sm:inline">后一天</span>
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <!-- 日历弹出层 -->
    <Teleport to="body">
      <div
        v-if="showCalendar"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/30"
        @click.self="showCalendar = false"
      >
        <div class="glass-card rounded-2xl p-4 w-80 shadow-xl">
          <!-- 月份导航 -->
          <div class="flex items-center justify-between mb-3">
            <button
              @click="prevMonth"
              class="p-1.5 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-accent/10 hover:text-accent transition-all duration-200"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
              </svg>
            </button>

            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ viewYear }}年{{ viewMonth }}月
            </span>

            <button
              @click="nextMonth"
              :disabled="!canGoNext"
              class="p-1.5 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-accent/10 hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-200"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
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
                  'w-9 h-9 rounded-lg text-sm flex items-center justify-center transition-all duration-150 relative',
                  cell.dateStr === currentDate
                    ? 'bg-accent text-white font-medium shadow-md'
                    : cell.isToday
                      ? 'text-accent font-semibold'
                      : cell.available
                        ? 'text-gray-700 dark:text-gray-300 hover:bg-accent/10 cursor-pointer'
                        : 'text-gray-300 dark:text-gray-600 cursor-default'
                ]"
              >
                {{ cell.day }}
              </button>
            </div>
          </div>

          <!-- 回到今天 -->
          <div v-if="currentDate !== today" class="flex justify-center mt-2 pt-2 border-t border-gray-100 dark:border-gray-700/50">
            <button
              @click="goToday"
              class="text-xs text-accent hover:text-accent/80 transition-colors duration-200 px-3 py-1 rounded-lg hover:bg-accent/10"
            >
              回到今天
            </button>
          </div>
        </div>
      </div>
    </Teleport>
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

const showCalendar = ref(false)
const dateBtnRef = ref(null)
const weekHeaders = ['日', '一', '二', '三', '四', '五', '六']

const today = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}`
})

const formattedDate = computed(() => {
  const d = parseDate(props.currentDate)
  if (!d) return props.currentDate
  const weekday = weekHeaders[d.getDay()]
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} 周${weekday}`
})

const hasPrev = computed(() => {
  const idx = props.availableDates.indexOf(props.currentDate)
  return idx >= 0 && idx < props.availableDates.length - 1
})

const hasNext = computed(() => {
  const idx = props.availableDates.indexOf(props.currentDate)
  return idx > 0
})

// 日历视图月份
const viewDate = ref(parseDate(props.currentDate) || new Date())
const viewYear = computed(() => viewDate.value.getFullYear())
const viewMonth = computed(() => viewDate.value.getMonth() + 1)

const canGoNext = computed(() => {
  const now = new Date()
  return !(viewDate.value.getFullYear() === now.getFullYear() && viewDate.value.getMonth() === now.getMonth())
})

const calendarCells = computed(() => {
  const year = viewDate.value.getFullYear()
  const month = viewDate.value.getMonth()
  const firstDay = new Date(year, month, 1).getDay()
  const daysInMonth = new Date(year, month + 1, 0).getDate()

  const cells = []
  for (let i = 0; i < firstDay; i++) {
    cells.push({ day: null, dateStr: '', available: false, isToday: false })
  }
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

function goPrev() {
  const idx = props.availableDates.indexOf(props.currentDate)
  if (idx >= 0 && idx < props.availableDates.length - 1) {
    emit('date-change', props.availableDates[idx + 1])
  }
}

function goNext() {
  const idx = props.availableDates.indexOf(props.currentDate)
  if (idx > 0) {
    emit('date-change', props.availableDates[idx - 1])
  }
}

function selectDate(d) {
  showCalendar.value = false
  emit('date-change', d)
}

function goToday() {
  viewDate.value = new Date()
  showCalendar.value = false
  emit('date-change', today.value)
}

function parseDate(str) {
  if (!str || str.length !== 8) return null
  return new Date(parseInt(str.substring(0, 4)), parseInt(str.substring(4, 6)) - 1, parseInt(str.substring(6, 8)))
}
</script>
