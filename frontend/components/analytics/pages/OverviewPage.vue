<!-- 数据概览页组件 -->
<template>
  <div class="space-y-12" v-if="viewingData">
    <h3 class="text-3xl font-bold text-center bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent">
      年度观看数据
    </h3>

    <!-- 核心统计卡片 -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 text-center border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <div class="text-2xl font-bold text-accent">{{ viewingData.total_videos || 0 }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">观看视频数</div>
      </div>
      <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 text-center border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <div class="text-2xl font-bold text-accent">{{ formatDurationShort(viewingData.total_duration || 0) }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">总观看时长</div>
      </div>
      <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 text-center border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <div class="text-2xl font-bold text-accent">{{ viewingData.active_days || 0 }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">活跃天数</div>
      </div>
      <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 text-center border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <div class="text-2xl font-bold text-accent">{{ viewingData.unique_authors || 0 }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">UP主数量</div>
      </div>
    </div>

    <div class="text-base text-center text-gray-600 dark:text-gray-300 space-y-3">
      <!-- 总体活动总结（放在最前面） -->
      <div v-if="viewingData?.insights?.overall_activity"
        v-html="formatInsightText(viewingData.insights.overall_activity)"
      >
      </div>

      <!-- 按指定顺序合并所有总结，用逗号分隔 -->
      <div>
        <span v-if="viewingBehaviorData?.report?.total_summary" v-html="formatInsightText(viewingBehaviorData.report.total_summary)"></span>
        <span v-if="viewingBehaviorData?.report?.total_summary && viewingBehaviorData?.report?.category_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.category_summary" v-html="formatInsightText(viewingBehaviorData.report.category_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary) && viewingBehaviorData?.report?.device_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.device_summary" v-html="formatInsightText(viewingBehaviorData.report.device_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary || viewingBehaviorData?.report?.device_summary) && viewingBehaviorData?.report?.up_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.up_summary" v-html="formatInsightText(viewingBehaviorData.report.up_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary || viewingBehaviorData?.report?.device_summary || viewingBehaviorData?.report?.up_summary) && viewingBehaviorData?.report?.time_slot_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.time_slot_summary" v-html="formatInsightText(viewingBehaviorData.report.time_slot_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary || viewingBehaviorData?.report?.device_summary || viewingBehaviorData?.report?.up_summary || viewingBehaviorData?.report?.time_slot_summary) && viewingBehaviorData?.report?.late_night_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.late_night_summary" v-html="formatInsightText(viewingBehaviorData.report.late_night_summary)"></span>
      </div>
    </div>

    <!-- 常看分区 + 常看UP主 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- 常看分区 Top 6 -->
      <div v-if="viewingData.tag_ranking?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z" />
          </svg>
          常看分区
        </h4>
        <div class="space-y-2">
          <div v-for="(tag, index) in viewingData.tag_ranking.slice(0, 6)" :key="tag.tag_name" class="flex items-center gap-2">
            <span class="text-xs text-gray-400 w-4 text-right">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ tag.tag_name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ tag.count }}次</span>
              </div>
              <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div class="h-full bg-gradient-to-r from-accent to-accent/70 rounded-full" :style="{ width: `${(tag.count / viewingData.tag_ranking[0].count) * 100}%` }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 常看UP主 Top 5 -->
      <div v-if="viewingData.author_ranking?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          常看UP主
        </h4>
        <div class="space-y-2">
          <div v-for="(author, index) in viewingData.author_ranking.slice(0, 5)" :key="author.author_mid" class="flex items-center gap-2">
            <span class="text-xs text-gray-400 w-4 text-right">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ author.author_name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ author.count }}次</span>
              </div>
              <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div class="h-full bg-gradient-to-r from-accent to-accent/70 rounded-full" :style="{ width: `${(author.count / viewingData.author_ranking[0].count) * 100}%` }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 设备分布 + 标题热词 + 时长偏好 -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <!-- 设备分布 -->
      <div v-if="deviceEntries.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">设备分布</h4>
        <div class="space-y-2">
          <div v-for="[device, count] in deviceEntries" :key="device" class="flex items-center gap-2">
            <div class="w-2 h-2 rounded-full flex-shrink-0" :class="device === '手机' ? 'bg-accent' : device === '电脑' ? 'bg-blue-500' : device === '平板' ? 'bg-green-500' : 'bg-gray-400'"></div>
            <span class="text-sm text-gray-700 dark:text-gray-300 flex-1">{{ device }}</span>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ count }}</span>
          </div>
        </div>
      </div>

      <!-- 标题热词 -->
      <div v-if="viewingData.title_keywords?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">标题热词</h4>
        <div class="flex flex-wrap gap-1.5">
          <span
            v-for="(kw, index) in viewingData.title_keywords.slice(0, 15)"
            :key="kw.word"
            class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium transition-colors"
            :class="index < 3 ? 'bg-accent/15 text-accent' : index < 6 ? 'bg-accent/10 text-accent/80' : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'"
          >
            {{ kw.word }}
            <span class="ml-1 text-[9px] opacity-60">{{ kw.count }}</span>
          </span>
        </div>
      </div>

      <!-- 时长偏好 -->
      <div v-if="viewingData.duration_preference" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">时长偏好</h4>
        <div class="flex h-4 rounded-full overflow-hidden mb-2">
          <div class="bg-green-400 transition-all" :style="{ width: `${viewingData.duration_preference.short_ratio * 100}%` }" title="短视频"></div>
          <div class="bg-accent transition-all" :style="{ width: `${viewingData.duration_preference.mid_ratio * 100}%` }" title="中视频"></div>
          <div class="bg-blue-500 transition-all" :style="{ width: `${viewingData.duration_preference.long_ratio * 100}%` }" title="长视频"></div>
        </div>
        <div class="flex justify-between text-[10px] text-gray-500 dark:text-gray-400">
          <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-green-400 inline-block"></span>短(&lt;5min) {{ viewingData.duration_preference.short }}</span>
          <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-accent inline-block"></span>中(5-20min) {{ viewingData.duration_preference.mid }}</span>
          <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-blue-500 inline-block"></span>长(&gt;20min) {{ viewingData.duration_preference.long }}</span>
        </div>
      </div>
    </div>

    <!-- 周内分布 -->
    <div v-if="viewingData.weekday_distribution?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        周内分布
      </h4>
      <div class="space-y-1.5">
        <div v-for="day in viewingData.weekday_distribution" :key="day.name" class="flex items-center gap-2">
          <span class="text-[10px] text-gray-500 dark:text-gray-400 w-6 text-right flex-shrink-0">{{ day.name }}</span>
          <div class="flex-1 h-4 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
            <div class="h-full bg-gradient-to-r from-accent to-accent/70 rounded-full transition-all duration-300" :style="{ width: `${Math.max((day.count / maxWeekdayCount) * 100, 2)}%` }"></div>
          </div>
          <span class="text-[10px] text-gray-500 dark:text-gray-400 w-5 text-right flex-shrink-0">{{ day.count }}</span>
        </div>
      </div>
    </div>

    <!-- 最长观看视频 -->
    <div v-if="viewingData.top_videos?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        最长观看
      </h4>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
        <div
          v-for="(video, index) in viewingData.top_videos"
          :key="video.bvid"
          class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
        >
          <span class="text-xs text-gray-400 w-4 text-right font-mono">{{ index + 1 }}</span>
          <img
            v-if="video.cover"
            :src="normalizeImageUrl(video.cover)"
            class="w-16 h-10 rounded object-cover flex-shrink-0"
            loading="lazy"
          />
          <div class="flex-1 min-w-0">
            <div class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ video.title }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ video.author_name }}</div>
          </div>
          <div class="text-right flex-shrink-0">
            <div class="text-sm font-semibold text-accent">{{ formatDurationShort(video.duration) }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 年度观看热力图 -->
    <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-xl p-4 border border-gray-200/50 dark:border-gray-700/50 shadow-sm">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        年度观看热力图
      </h4>
      <HtmlHeatmap :year="viewingData?.year" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { getViewingBehavior } from '~/utils/api'
import { formatDurationShort } from '~/utils/format'
import { normalizeImageUrl } from '~/utils/imageUrl'
import HtmlHeatmap from './HtmlHeatmap.vue'

const props = defineProps({
  viewingData: {
    type: Object,
    required: true
  }
})

const viewingBehaviorData = ref(null)

// 格式化洞察文本，为数字添加颜色
const formatInsightText = (text) => {
  if (!text) return '';
  return text.replace(/(\d+(\.\d+)?)/g, '<span class="text-accent">$1</span>')
}

// 设备分布条目
const deviceEntries = computed(() => {
  if (!props.viewingData?.device_distribution) return []
  return Object.entries(props.viewingData.device_distribution).sort((a, b) => b[1] - a[1])
})

// 周内分布最大值
const maxWeekdayCount = computed(() => {
  if (!props.viewingData?.weekday_distribution?.length) return 1
  return Math.max(...props.viewingData.weekday_distribution.map(d => d.count), 1)
})

// 获取观看行为数据
const fetchViewingBehavior = async (year) => {
  if (!year) return

  try {
    const response = await getViewingBehavior(year, true)
    if (response.data && response.data.status === 'success') {
      viewingBehaviorData.value = response.data.data
    }
  } catch (error) {
    console.error('获取观看行为数据失败:', error)
  }
}

onMounted(() => {
  if (props.viewingData?.year) {
    fetchViewingBehavior(props.viewingData.year)
  }
})

// 监听年份变化
watch(() => props.viewingData?.year, (newYear) => {
  if (newYear) {
    fetchViewingBehavior(newYear)
  }
}, { immediate: true })
</script>
