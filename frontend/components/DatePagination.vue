<template>
  <div class="mx-auto max-w-4xl">
    <div class="glass-card px-4 py-3 flex items-center justify-between gap-2">
      <!-- Previous day -->
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

      <!-- Date display & picker -->
      <div class="flex min-w-0 flex-1 items-center justify-center gap-3">
        <button
          ref="dateBtnRef"
          @click="showPicker = true"
          class="flex items-center gap-2 px-3 py-1.5 rounded-xl text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-accent/10 transition-all duration-200 cursor-pointer"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <span>{{ formattedDate }}</span>
        </button>

        <!-- Record count -->
        <span v-if="recordCount > 0" class="text-xs text-gray-400 dark:text-gray-500 shrink-0">
          {{ recordCount }} 条
        </span>
      </div>

      <!-- Next day -->
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

    <!-- Date picker overlay -->
    <Teleport to="body">
      <div
        v-if="showPicker"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/30"
        @click.self="showPicker = false"
      >
        <div class="glass-card rounded-2xl p-4 w-72 max-h-96 overflow-y-auto shadow-xl">
          <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3 text-center">选择日期</div>

          <!-- Year groups -->
          <div v-for="year in groupedDates" :key="year.year" class="mb-3">
            <div class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1 px-1">{{ year.year }}</div>
            <div class="grid grid-cols-4 gap-1">
              <button
                v-for="d in year.dates"
                :key="d"
                @click="selectDate(d)"
                :class="[
                  'px-2 py-1.5 rounded-lg text-xs transition-all duration-150',
                  d === currentDate
                    ? 'bg-accent text-white font-medium'
                    : 'text-gray-600 dark:text-gray-400 hover:bg-accent/10'
                ]"
              >
                {{ formatShortDate(d) }}
              </button>
            </div>
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

const showPicker = ref(false)
const dateBtnRef = ref(null)

const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']

const formattedDate = computed(() => {
  const d = parseDate(props.currentDate)
  if (!d) return props.currentDate
  const weekday = WEEKDAYS[d.getDay()]
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

const groupedDates = computed(() => {
  const groups = {}
  for (const d of props.availableDates) {
    const year = d.substring(0, 4)
    if (!groups[year]) groups[year] = []
    groups[year].push(d)
  }
  return Object.entries(groups)
    .sort(([a], [b]) => b.localeCompare(a))
    .map(([year, dates]) => ({ year, dates }))
})

function parseDate(str) {
  if (!str || str.length !== 8) return null
  return new Date(parseInt(str.substring(0, 4)), parseInt(str.substring(4, 6)) - 1, parseInt(str.substring(6, 8)))
}

function formatShortDate(str) {
  const d = parseDate(str)
  if (!d) return str
  return `${d.getMonth() + 1}/${d.getDate()}`
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
  showPicker.value = false
  emit('date-change', d)
}
</script>
