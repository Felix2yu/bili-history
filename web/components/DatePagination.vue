<template>
  <div :class="compact ? '' : 'mx-auto max-w-4xl'">
    <div
      :class="
        compact
          ? 'flex items-center justify-center gap-1 md:gap-2'
          : 'glass-card flex items-center justify-between gap-2 px-4 py-3'
      "
    >
      <!-- 前一天 -->
      <button
        :disabled="!hasPrev"
        :class="
          compact
            ? 'flex shrink-0 items-center gap-1 rounded-lg px-2 py-1.5 text-xs text-gray-600 transition-all duration-200 hover:bg-accent/10 hover:text-accent disabled:cursor-not-allowed disabled:opacity-30 md:text-sm dark:text-gray-400'
            : 'flex shrink-0 items-center gap-1 rounded-xl px-3 py-2 text-sm text-gray-600 transition-all duration-200 hover:bg-accent/10 hover:text-accent disabled:cursor-not-allowed disabled:opacity-30 dark:text-gray-400'
        "
        @click="goPrev"
      >
        <svg
          :class="compact ? 'h-3.5 w-3.5' : 'h-4 w-4'"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        <span :class="compact ? 'hidden md:inline' : 'hidden sm:inline'">前一天</span>
      </button>

      <!-- 日期显示 & 日历按钮 -->
      <div
        :class="
          compact
            ? 'flex min-w-0 items-center justify-center gap-1 md:gap-2'
            : 'flex min-w-0 flex-1 items-center justify-center gap-3'
        "
      >
        <button
          ref="dateBtnRef"
          :class="
            compact
              ? 'flex cursor-pointer items-center gap-1.5 rounded-lg px-2 py-1 text-xs font-medium text-gray-700 transition-all duration-200 hover:bg-accent/10 md:text-sm dark:text-gray-300'
              : 'flex cursor-pointer items-center gap-2 rounded-xl px-3 py-1.5 text-sm font-medium text-gray-700 transition-all duration-200 hover:bg-accent/10 dark:text-gray-300'
          "
          @click="showCalendar = !showCalendar"
        >
          <svg
            :class="compact ? 'h-3.5 w-3.5' : 'h-4 w-4'"
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
          <span :class="compact ? 'hidden sm:inline' : ''">{{ formattedDate }}</span>
          <span :class="compact ? 'sm:hidden' : 'hidden'">{{ shortDate }}</span>
        </button>

        <span
          v-if="recordCount > 0 && !compact"
          class="shrink-0 text-xs text-gray-400 dark:text-gray-500"
        >
          {{ recordCount }} 条
        </span>
      </div>

      <!-- 后一天 -->
      <button
        :disabled="!hasNext"
        :class="
          compact
            ? 'flex shrink-0 items-center gap-1 rounded-lg px-2 py-1.5 text-xs text-gray-600 transition-all duration-200 hover:bg-accent/10 hover:text-accent disabled:cursor-not-allowed disabled:opacity-30 md:text-sm dark:text-gray-400'
            : 'flex shrink-0 items-center gap-1 rounded-xl px-3 py-2 text-sm text-gray-600 transition-all duration-200 hover:bg-accent/10 hover:text-accent disabled:cursor-not-allowed disabled:opacity-30 dark:text-gray-400'
        "
        @click="goNext"
      >
        <span :class="compact ? 'hidden md:inline' : 'hidden sm:inline'">后一天</span>
        <svg
          :class="compact ? 'h-3.5 w-3.5' : 'h-4 w-4'"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
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
        <div class="glass-card w-80 rounded-2xl p-4 shadow-xl">
          <!-- 年月选择栏 -->
          <div class="mb-3 flex items-center justify-between">
            <!-- 年份 -->
            <div class="relative">
              <button
                class="rounded-lg px-2 py-1 text-sm font-medium text-gray-700 transition-colors hover:bg-accent/10 hover:text-accent dark:text-gray-300"
                @click="openYearPicker"
              >
                {{ viewYear }}年
              </button>
              <!-- 年份选择面板 -->
              <div
                v-if="showYearPicker"
                class="glass-card absolute top-full left-0 z-10 mt-1 w-56 rounded-xl p-3 shadow-xl"
              >
                <div class="mb-2 flex items-center justify-between">
                  <button
                    class="rounded p-1 text-gray-400 hover:text-accent"
                    @click="yearPageStart -= 12"
                  >
                    <svg
                      class="h-3.5 w-3.5"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
                    </svg>
                  </button>
                  <span class="text-xs text-gray-400"
                    >{{ yearPageStart }}-{{ yearPageStart + 11 }}</span
                  >
                  <button
                    class="rounded p-1 text-gray-400 hover:text-accent"
                    @click="yearPageStart += 12"
                  >
                    <svg
                      class="h-3.5 w-3.5"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </div>
                <div class="grid grid-cols-4 gap-1">
                  <button
                    v-for="y in yearOptions"
                    :key="y"
                    :class="[
                      'rounded-lg px-2 py-1.5 text-xs transition-all',
                      y === viewYear
                        ? 'bg-accent font-medium text-white'
                        : availableYears.includes(y)
                          ? 'text-gray-700 hover:bg-accent/10 dark:text-gray-300'
                          : 'text-gray-300 dark:text-gray-600'
                    ]"
                    @click="selectYear(y)"
                  >
                    {{ y }}
                  </button>
                </div>
              </div>
            </div>

            <!-- 月份 -->
            <div class="relative">
              <button
                class="rounded-lg px-2 py-1 text-sm font-medium text-gray-700 transition-colors hover:bg-accent/10 hover:text-accent dark:text-gray-300"
                @click="
                  showMonthPicker = !showMonthPicker;
                  showYearPicker = false;
                "
              >
                {{ viewMonth }}月
              </button>
              <!-- 月份选择面板 -->
              <div
                v-if="showMonthPicker"
                class="glass-card absolute top-full left-0 z-10 mt-1 w-52 rounded-xl p-3 shadow-xl"
              >
                <div class="grid grid-cols-4 gap-1">
                  <button
                    v-for="m in 12"
                    :key="m"
                    :class="[
                      'rounded-lg px-2 py-2 text-xs transition-all',
                      m === viewMonth
                        ? 'bg-accent font-medium text-white'
                        : isMonthAvailable(viewYear, m)
                          ? 'text-gray-700 hover:bg-accent/10 dark:text-gray-300'
                          : 'text-gray-300 dark:text-gray-600'
                    ]"
                    @click="selectMonth(m)"
                  >
                    {{ m }}月
                  </button>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-1">
              <button
                class="rounded-lg p-1.5 text-gray-500 transition-all duration-200 hover:bg-accent/10 hover:text-accent dark:text-gray-400"
                @click="prevMonth"
              >
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <button
                :disabled="!canGoNext"
                class="rounded-lg p-1.5 text-gray-500 transition-all duration-200 hover:bg-accent/10 hover:text-accent disabled:cursor-not-allowed disabled:opacity-30 dark:text-gray-400"
                @click="nextMonth"
              >
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </div>
          </div>

          <!-- 星期标题 -->
          <div class="mb-1 grid grid-cols-7">
            <div
              v-for="w in weekHeaders"
              :key="w"
              class="py-1 text-center text-xs font-medium text-gray-400 dark:text-gray-500"
            >
              {{ w }}
            </div>
          </div>

          <!-- 日期网格 -->
          <div class="grid grid-cols-7 gap-y-0.5">
            <div v-for="(cell, i) in calendarCells" :key="i" class="flex justify-center">
              <button
                v-if="cell.day"
                :disabled="!cell.available"
                :class="[
                  'relative flex h-9 w-9 items-center justify-center rounded-lg text-sm transition-all duration-150',
                  cell.dateStr === currentDate
                    ? 'bg-accent font-medium text-white shadow-md'
                    : cell.isToday
                      ? 'font-semibold text-accent'
                      : cell.available
                        ? 'cursor-pointer text-gray-700 hover:bg-accent/10 dark:text-gray-300'
                        : 'cursor-default text-gray-300 dark:text-gray-600'
                ]"
                @click="cell.available && selectDate(cell.dateStr)"
              >
                {{ cell.day }}
              </button>
            </div>
          </div>

          <!-- 回到今天 -->
          <div
            v-if="currentDate !== today"
            class="mt-2 flex justify-center border-t border-gray-100 pt-2 dark:border-gray-700/50"
          >
            <button
              class="rounded-lg px-3 py-1 text-xs text-accent transition-colors duration-200 hover:bg-accent/10 hover:text-accent/80"
              @click="goToday"
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
import { ref, computed, watch } from 'vue';

const props = defineProps({
  currentDate: { type: String, required: true },
  availableDates: { type: Array, default: () => [] },
  recordCount: { type: Number, default: 0 },
  compact: { type: Boolean, default: false }
});

const emit = defineEmits(['date-change']);

const showCalendar = ref(false);
const showYearPicker = ref(false);
const showMonthPicker = ref(false);
const dateBtnRef = ref(null);
const weekHeaders = ['一', '二', '三', '四', '五', '六', '日'];

const today = computed(() => {
  const now = new Date();
  return `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}`;
});

const formattedDate = computed(() => {
  const d = parseDate(props.currentDate);
  if (!d) return props.currentDate;
  const weekdayIndex = (d.getDay() + 6) % 7;
  const weekday = weekHeaders[weekdayIndex];
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} 周${weekday}`;
});

const shortDate = computed(() => {
  const d = parseDate(props.currentDate);
  if (!d) return props.currentDate;
  const pad = (n) => String(n).padStart(2, '0');
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
});

const hasPrev = computed(() => {
  const idx = props.availableDates.indexOf(props.currentDate);
  return idx >= 0 && idx < props.availableDates.length - 1;
});

const hasNext = computed(() => {
  const idx = props.availableDates.indexOf(props.currentDate);
  return idx > 0;
});

// 日历视图
const viewDate = ref(parseDate(props.currentDate) || new Date());
const viewYear = computed(() => viewDate.value.getFullYear());
const viewMonth = computed(() => viewDate.value.getMonth() + 1);

const canGoNext = computed(() => {
  const now = new Date();
  return !(
    viewDate.value.getFullYear() === now.getFullYear() &&
    viewDate.value.getMonth() === now.getMonth()
  );
});

// 有数据的年份列表
const availableYears = computed(() => {
  const years = new Set();
  for (const d of props.availableDates) {
    years.add(parseInt(d.substring(0, 4)));
  }
  return [...years].sort((a, b) => b - a);
});

// 年份选择器
const yearPageStart = ref(viewYear.value - 5);
const yearOptions = computed(() => {
  const years = [];
  for (let i = 0; i < 12; i++) {
    years.push(yearPageStart.value + i);
  }
  return years;
});

function openYearPicker() {
  showYearPicker.value = !showYearPicker.value;
  showMonthPicker.value = false;
  if (showYearPicker.value) {
    yearPageStart.value = viewYear.value - 5;
  }
}

function selectYear(y) {
  const d = new Date(viewDate.value);
  d.setFullYear(y);
  viewDate.value = d;
  showYearPicker.value = false;
}

function selectMonth(m) {
  const d = new Date(viewDate.value);
  d.setMonth(m - 1);
  viewDate.value = d;
  showMonthPicker.value = false;
}

function isMonthAvailable(year, month) {
  for (const dateStr of props.availableDates) {
    const y = parseInt(dateStr.substring(0, 4));
    const m = parseInt(dateStr.substring(4, 6));
    if (y === year && m === month) return true;
  }
  return false;
}

const calendarCells = computed(() => {
  const year = viewDate.value.getFullYear();
  const month = viewDate.value.getMonth();
  // getDay(): 0=周日 -> 转换为周一起始: 0=周一,6=周日
  const firstDayOfWeek = (new Date(year, month, 1).getDay() + 6) % 7;
  const daysInMonth = new Date(year, month + 1, 0).getDate();

  const cells = [];
  for (let i = 0; i < firstDayOfWeek; i++) {
    cells.push({ day: null, dateStr: '', available: false, isToday: false });
  }
  for (let d = 1; d <= daysInMonth; d++) {
    const dateStr = `${year}${String(month + 1).padStart(2, '0')}${String(d).padStart(2, '0')}`;
    cells.push({
      day: d,
      dateStr,
      available: props.availableDates.includes(dateStr),
      isToday: dateStr === today.value
    });
  }
  return cells;
});

function prevMonth() {
  const d = new Date(viewDate.value);
  d.setMonth(d.getMonth() - 1);
  viewDate.value = d;
}

function nextMonth() {
  if (!canGoNext.value) return;
  const d = new Date(viewDate.value);
  d.setMonth(d.getMonth() + 1);
  viewDate.value = d;
}

function goPrev() {
  const idx = props.availableDates.indexOf(props.currentDate);
  if (idx >= 0 && idx < props.availableDates.length - 1) {
    emit('date-change', props.availableDates[idx + 1]);
  }
}

function goNext() {
  const idx = props.availableDates.indexOf(props.currentDate);
  if (idx > 0) {
    emit('date-change', props.availableDates[idx - 1]);
  }
}

function selectDate(d) {
  showCalendar.value = false;
  showYearPicker.value = false;
  showMonthPicker.value = false;
  emit('date-change', d);
}

function goToday() {
  viewDate.value = new Date();
  showCalendar.value = false;
  showYearPicker.value = false;
  showMonthPicker.value = false;
  emit('date-change', today.value);
}

function parseDate(str) {
  if (!str || str.length !== 8) return null;
  return new Date(
    parseInt(str.substring(0, 4)),
    parseInt(str.substring(4, 6)) - 1,
    parseInt(str.substring(6, 8))
  );
}

// 关闭日历时重置选择器状态
watch(showCalendar, (v) => {
  if (!v) {
    showYearPicker.value = false;
    showMonthPicker.value = false;
  }
});
</script>
