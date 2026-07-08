<template>
  <div class="space-y-6">
    <!-- 图例 -->
    <div class="flex justify-center items-center text-sm text-gray-500 dark:text-gray-400 space-x-6">
      <div class="flex items-center">
        <span class="inline-block w-3 h-3 rounded-sm bg-[#FFECF1] dark:bg-[#4B1F2C] mr-1"></span>
        <span>1-10</span>
      </div>
      <div class="flex items-center">
        <span class="inline-block w-3 h-3 rounded-sm bg-[#FFB3CA] dark:bg-[#7A2D47] mr-1"></span>
        <span>11-50</span>
      </div>
      <div class="flex items-center">
        <span class="inline-block w-3 h-3 rounded-sm bg-[#FF8CB0] dark:bg-[#B3476A] mr-1"></span>
        <span>51-100</span>
      </div>
      <div class="flex items-center">
        <span class="inline-block w-3 h-3 rounded-sm bg-[#FF6699] dark:bg-[#E35C8B] mr-1"></span>
        <span>101-200</span>
      </div>
      <div class="flex items-center">
        <span class="inline-block w-3 h-3 rounded-sm bg-[#E84B85] dark:bg-[#FF7FA8] mr-1"></span>
        <span>201+</span>
      </div>
    </div>

    <!-- 热力图主体 -->
    <div class="overflow-x-auto">
      <div class="inline-flex gap-0.5">
        <!-- 星期标签 -->
        <div class="flex flex-col gap-0.5 mr-1 text-xs text-gray-500 dark:text-gray-400">
          <div v-for="(day, i) in dayLabels" :key="i" class="h-[14px] leading-[14px]">
            {{ i % 2 === 1 ? day : '' }}
          </div>
        </div>

        <!-- 月份和日期格子 -->
        <div v-for="(week, wi) in weeks" :key="wi" class="flex flex-col gap-0.5">
          <!-- 月份标签（仅第一行显示） -->
          <div class="h-[14px] leading-[14px] text-xs text-gray-500 dark:text-gray-400 text-center">
            {{ week.monthLabel }}
          </div>
          <!-- 日期格子 -->
          <div
            v-for="(cell, ci) in week.cells"
            :key="ci"
            class="w-[14px] h-[14px] rounded-[2px] cursor-pointer transition-all duration-150 relative group"
            :class="cell ? getCellClass(cell.count) : 'bg-gray-100 dark:bg-gray-800'"
            @mouseenter="showTooltip($event, cell)"
            @mouseleave="hideTooltip"
          >
            <!-- Tooltip -->
            <div
              v-if="tooltip.visible && tooltip.cell === cell"
              class="absolute z-50 px-3 py-2 text-sm rounded-lg shadow-lg pointer-events-none whitespace-nowrap"
              :class="isDarkMode ? 'bg-gray-800 text-white border border-gray-600' : 'bg-white text-gray-800 border border-gray-200'"
              :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
            >
              <div class="font-medium">{{ cell.date }}</div>
              <div class="text-[#fb7299]">{{ cell.count }} 个视频</div>
              <div v-if="cell.duration > 0" class="text-gray-500 dark:text-gray-400">
                观看时长 {{ formatDuration(cell.duration) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useDarkMode } from '~/stores/darkMode'
import { getHeatmapData, getDailyStats } from '~/utils/api'

const props = defineProps({
  year: {
    type: Number,
    required: true
  }
})

const { isDarkMode } = useDarkMode()
const heatmapData = ref({})
const dailyStats = ref({})
const tooltip = ref({ visible: false, cell: null, x: 0, y: 0 })

const dayLabels = ['日', '一', '二', '三', '四', '五', '六']

// 计算年份的所有周
const weeks = computed(() => {
  const year = props.year
  if (!year) return []

  const startDate = new Date(Date.UTC(year, 0, 1))
  const endDate = new Date(Date.UTC(year, 11, 31))
  const result = []

  let current = new Date(startDate)
  // 调整到第一个周日
  const dayOfWeek = current.getDay()
  if (dayOfWeek !== 0) {
    current = new Date(current.getTime() - (dayOfWeek * 24 * 60 * 60 * 1000))
  }

  let currentWeek = { monthLabel: '', cells: [] }
  let lastMonth = -1

  while (current <= endDate || currentWeek.cells.length > 0) {
    const dateStr = current.toISOString().split('T')[0]
    const month = current.getMonth()

    // 新月份开始，添加月份标签
    if (month !== lastMonth && current <= endDate) {
      if (currentWeek.cells.length > 0) {
        result.push(currentWeek)
      }
      currentWeek = { monthLabel: `${month + 1}月`, cells: [] }
      lastMonth = month
    }

    if (current <= endDate) {
      const count = heatmapData.value[dateStr] || 0
      const stats = dailyStats.value[dateStr] || {}
      currentWeek.cells.push({
        date: dateStr,
        count,
        duration: stats.total_duration || 0,
        weekday: current.getDay()
      })
    } else {
      // 填充最后一周的空位
      if (currentWeek.cells.length > 0) {
        currentWeek.cells.push(null)
      }
    }

    // 周日或到达结束日期时，保存当前周
    if ((current.getDay() === 6 || current >= endDate) && currentWeek.cells.length > 0) {
      result.push(currentWeek)
      currentWeek = { monthLabel: '', cells: [] }
    }

    current = new Date(current.getTime() + 24 * 60 * 60 * 1000)
  }

  return result
})

// 获取单元格颜色类
const getCellClass = (count) => {
  if (count === 0) return 'bg-gray-100 dark:bg-gray-800'
  if (count <= 10) return isDarkMode.value ? 'bg-[#4B1F2C]' : 'bg-[#FFECF1]'
  if (count <= 50) return isDarkMode.value ? 'bg-[#7A2D47]' : 'bg-[#FFB3CA]'
  if (count <= 100) return isDarkMode.value ? 'bg-[#B3476A]' : 'bg-[#FF8CB0]'
  if (count <= 200) return isDarkMode.value ? 'bg-[#E35C8B]' : 'bg-[#FF6699]'
  return isDarkMode.value ? 'bg-[#FF7FA8]' : 'bg-[#E84B85]'
}

// 格式化时长
const formatDuration = (seconds) => {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}小时${String(m).padStart(2, '0')}分`
  if (m > 0) return `${m}分${String(s).padStart(2, '0')}秒`
  return `${s}秒`
}

// 显示 tooltip
const showTooltip = (event, cell) => {
  if (!cell) return
  const rect = event.target.getBoundingClientRect()
  tooltip.value = {
    visible: true,
    cell,
    x: rect.left + rect.width / 2,
    y: rect.top - 10
  }
}

// 隐藏 tooltip
const hideTooltip = () => {
  tooltip.value = { visible: false, cell: null, x: 0, y: 0 }
}

// 获取热力图数据
const fetchHeatmapData = async () => {
  try {
    const response = await getHeatmapData(props.year)
    if (response.data.status === 'success') {
      heatmapData.value = response.data.data.data || {}
    }
  } catch (error) {
    console.error('获取热力图数据失败:', error)
  }
}

// 获取每日统计（用于显示观看时长）
const fetchDailyStats = async () => {
  try {
    // 从年度分析数据中获取每日统计
    // 这里我们可以复用现有的数据，或者单独获取
    const stats = {}
    // 遍历每天获取统计（如果需要的话）
    // 为了避免大量请求，我们可以先不获取，只显示数量
    dailyStats.value = stats
  } catch (error) {
    console.error('获取每日统计失败:', error)
  }
}

onMounted(() => {
  fetchHeatmapData()
})

watch(() => props.year, () => {
  fetchHeatmapData()
})
</script>
