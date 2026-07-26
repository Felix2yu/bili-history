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
    <div class="overflow-x-auto" ref="containerRef">
      <div class="flex gap-0.5 w-full">
        <!-- 星期标签 -->
        <div class="flex flex-col gap-0.5 mr-1 text-xs text-gray-500 dark:text-gray-400 flex-shrink-0">
          <div v-for="(day, i) in dayLabels" :key="i" class="h-[14px] leading-[14px]">
            {{ i % 2 === 1 ? day : '' }}
          </div>
        </div>

        <!-- 月份和日期格子 -->
        <div v-for="(week, wi) in weeks" :key="wi" class="flex flex-col gap-0.5 flex-1 min-w-[14px]">
          <!-- 月份标签（仅在月份开始的周显示） -->
          <div class="h-[14px] leading-[14px] text-xs text-gray-500 dark:text-gray-400 text-center whitespace-nowrap">
            {{ week.monthLabel }}
          </div>
          <!-- 日期格子 -->
          <div
            v-for="(cell, ci) in week.cells"
            :key="ci"
            class="h-[14px] rounded-[2px] cursor-pointer transition-all duration-150 w-full"
            :class="cell ? getCellClass(cell.count) : 'bg-gray-100 dark:bg-gray-800'"
            @mouseenter="showTooltip($event, cell)"
            @mouseleave="hideTooltip"
          ></div>
        </div>
      </div>
    </div>

    <!-- Tooltip (fixed positioning) -->
    <Teleport to="body">
      <div
        v-if="tooltip.visible && tooltip.cell"
        class="fixed z-[9999] px-3 py-2 text-sm rounded-lg shadow-lg pointer-events-none whitespace-nowrap"
        :class="isDarkMode ? 'bg-gray-800 text-white border border-gray-600' : 'bg-white text-gray-800 border border-gray-200'"
        :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px', transform: 'translate(-50%, -100%)' }"
      >
        <div class="font-medium">{{ tooltip.cell.date }}</div>
        <div class="text-accent">{{ tooltip.cell.count }} 个视频</div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useDarkMode } from '~/stores/darkMode'
import { getHeatmapData } from '~/utils/api'

const props = defineProps({
  year: {
    type: Number,
    required: true
  }
})

const { isDarkMode } = useDarkMode()
const containerRef = ref(null)
const heatmapData = ref({})
const tooltip = ref({ visible: false, cell: null, x: 0, y: 0 })

const dayLabels = ['一', '二', '三', '四', '五', '六', '日']

// 计算年份的所有周
const weeks = computed(() => {
  const year = props.year
  if (!year) return []

  const result = []
  // 从1月1日开始
  const startDate = new Date(Date.UTC(year, 0, 1))
  const endDate = new Date(Date.UTC(year, 11, 31))

  // 找到第一个周一 (getDay: 0=周日,1=周一...6=周六)
  let current = new Date(startDate)
  const firstDayOfWeek = current.getDay()
  // 转换为周一起始: 0=周一,6=周日
  const daysToMonday = (firstDayOfWeek + 6) % 7
  if (daysToMonday !== 0) {
    // 回退到上一个周一
    current = new Date(current.getTime() - (daysToMonday * 24 * 60 * 60 * 1000))
  }

  let lastMonth = -1

  while (current <= endDate) {
    const weekCells = []
    let monthLabel = ''

    // 填充一周的7天
    for (let dayOfWeek = 0; dayOfWeek < 7; dayOfWeek++) {
      const dateStr = current.toISOString().split('T')[0]
      const month = current.getMonth()

      // 如果是新的一周的第一天且月份变了，记录月份标签
      if (dayOfWeek === 0 && month !== lastMonth) {
        monthLabel = `${month + 1}月`
        lastMonth = month
      }

      if (current <= endDate) {
        const count = heatmapData.value[dateStr] || 0
        weekCells.push({
          date: dateStr,
          count,
          weekday: dayOfWeek
        })
      } else {
        weekCells.push(null)
      }

      current = new Date(current.getTime() + 24 * 60 * 60 * 1000)
    }

    result.push({ monthLabel, cells: weekCells })
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

// 显示 tooltip
const showTooltip = (event, cell) => {
  if (!cell) return
  const rect = event.target.getBoundingClientRect()
  tooltip.value = {
    visible: true,
    cell,
    x: rect.left + rect.width / 2,
    y: rect.top - 8
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

onMounted(() => {
  fetchHeatmapData()
})

watch(() => props.year, () => {
  fetchHeatmapData()
})
</script>
