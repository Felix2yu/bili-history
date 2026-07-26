<template>
  <div class="space-y-6 md:space-y-8" v-if="viewingData">
    <div class="text-center">
      <h3 class="text-2xl md:text-3xl font-bold text-gray-900 dark:text-white">
        年度观看数据
      </h3>
      <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">{{ viewingData.year }}年数据概览</p>
    </div>

    <div class="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
      <div class="relative overflow-hidden rounded-2xl p-4 md:p-5 bg-rose-50 dark:from-rose-900/20 dark:to-rose-900/10 border border-rose-100 dark:border-rose-900/30">
        <div class="absolute -right-3 -top-3 w-16 h-16 rounded-full bg-rose-200/50 dark:bg-rose-800/30"></div>
        <div class="relative">
          <div class="flex items-center justify-between mb-2">
            <div class="w-9 h-9 rounded-xl bg-rose-500 flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
            </div>
          </div>
          <div class="text-2xl md:text-3xl font-bold text-rose-600 dark:text-rose-400">{{ viewingData.total_videos || 0 }}</div>
          <div class="text-[11px] md:text-xs text-rose-500/70 dark:text-rose-400/70 mt-0.5">观看视频数</div>
        </div>
      </div>

      <div class="relative overflow-hidden rounded-2xl p-4 md:p-5 bg-amber-50 dark:from-amber-900/20 dark:to-amber-900/10 border border-amber-100 dark:border-amber-900/30">
        <div class="absolute -right-3 -top-3 w-16 h-16 rounded-full bg-amber-200/50 dark:bg-amber-800/30"></div>
        <div class="relative">
          <div class="flex items-center justify-between mb-2">
            <div class="w-9 h-9 rounded-xl bg-amber-500 flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
          <div class="text-2xl md:text-3xl font-bold text-amber-600 dark:text-amber-400">{{ formatDurationShort(viewingData.total_duration || 0) }}</div>
          <div class="text-[11px] md:text-xs text-amber-500/70 dark:text-amber-400/70 mt-0.5">总观看时长</div>
        </div>
      </div>

      <div class="relative overflow-hidden rounded-2xl p-4 md:p-5 bg-emerald-50 dark:from-emerald-900/20 dark:to-emerald-900/10 border border-emerald-100 dark:border-emerald-900/30">
        <div class="absolute -right-3 -top-3 w-16 h-16 rounded-full bg-emerald-200/50 dark:bg-emerald-800/30"></div>
        <div class="relative">
          <div class="flex items-center justify-between mb-2">
            <div class="w-9 h-9 rounded-xl bg-emerald-500 flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
          </div>
          <div class="text-2xl md:text-3xl font-bold text-emerald-600 dark:text-emerald-400">{{ viewingData.active_days || 0 }}</div>
          <div class="text-[11px] md:text-xs text-emerald-500/70 dark:text-emerald-400/70 mt-0.5">活跃天数</div>
        </div>
      </div>

      <div class="relative overflow-hidden rounded-2xl p-4 md:p-5 bg-sky-50 dark:from-sky-900/20 dark:to-sky-900/10 border border-sky-100 dark:border-sky-900/30">
        <div class="absolute -right-3 -top-3 w-16 h-16 rounded-full bg-sky-200/50 dark:bg-sky-800/30"></div>
        <div class="relative">
          <div class="flex items-center justify-between mb-2">
            <div class="w-9 h-9 rounded-xl bg-sky-500 flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
            </div>
          </div>
          <div class="text-2xl md:text-3xl font-bold text-sky-600 dark:text-sky-400">{{ viewingData.unique_authors || 0 }}</div>
          <div class="text-[11px] md:text-xs text-sky-500/70 dark:text-sky-400/70 mt-0.5">UP主数量</div>
        </div>
      </div>
    </div>

    <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
      <div class="flex items-center gap-2 mb-3">
        <div class="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center">
          <svg class="w-4 h-4 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
        </div>
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white">数据洞察</h4>
      </div>
      <div class="text-sm text-gray-600 dark:text-gray-300 space-y-2">
        <div v-if="viewingData?.insights?.overall_activity" v-html="formatInsightText(viewingData.insights.overall_activity)"></div>
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
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-5">
      <div v-if="viewingData.tag_ranking?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <div class="w-8 h-8 rounded-lg bg-rose-100 dark:bg-rose-900/30 flex items-center justify-center">
            <svg class="w-4 h-4 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z" />
            </svg>
          </div>
          常看分区
        </h4>
        <div class="space-y-2.5">
          <div v-for="(tag, index) in viewingData.tag_ranking.slice(0, 6)" :key="tag.tag_name" class="flex items-center gap-3">
            <span class="w-6 h-6 rounded-lg flex items-center justify-center text-xs font-bold text-white shrink-0" :style="{ backgroundColor: getChartColor(index) }">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ tag.tag_name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2 shrink-0">{{ tag.count }}次</span>
              </div>
              <div class="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all duration-500" :style="{ width: `${(tag.count / viewingData.tag_ranking[0].count) * 100}%`, backgroundColor: getChartColor(index) }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="viewingData.author_ranking?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <div class="w-8 h-8 rounded-lg bg-sky-100 dark:bg-sky-900/30 flex items-center justify-center">
            <svg class="w-4 h-4 text-sky-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          常看UP主
        </h4>
        <div class="space-y-2.5">
          <div v-for="(author, index) in viewingData.author_ranking.slice(0, 5)" :key="author.author_mid" class="flex items-center gap-3">
            <span class="w-6 h-6 rounded-lg flex items-center justify-center text-xs font-bold text-white shrink-0" :style="{ backgroundColor: getChartColor(index + 2) }">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ author.author_name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2 shrink-0">{{ author.count }}次</span>
              </div>
              <div class="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all duration-500" :style="{ width: `${(author.count / viewingData.author_ranking[0].count) * 100}%`, backgroundColor: getChartColor(index + 2) }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 md:gap-5">
      <div v-if="deviceEntries.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">设备分布</h4>
        <div class="space-y-2.5">
          <div v-for="[device, count] in deviceEntries" :key="device" class="flex items-center gap-2.5">
            <div class="w-3 h-3 rounded-full shrink-0" :style="{ backgroundColor: getDeviceColor(device) }"></div>
            <span class="text-sm text-gray-700 dark:text-gray-300 flex-1">{{ device }}</span>
            <span class="text-xs font-semibold text-gray-600 dark:text-gray-400">{{ count }}</span>
          </div>
        </div>
      </div>

      <div v-if="viewingData.title_keywords?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">标题热词</h4>
        <div class="flex flex-wrap gap-1.5">
          <span
            v-for="(kw, index) in viewingData.title_keywords.slice(0, 15)"
            :key="kw.word"
            class="inline-flex items-center px-2.5 py-1 rounded-full text-[11px] font-medium"
            :style="{
              backgroundColor: getChartColor(index) + '20',
              color: getChartColor(index)
            }"
          >
            {{ kw.word }}
            <span class="ml-1 text-[10px] opacity-60">{{ kw.count }}</span>
          </span>
        </div>
      </div>

      <div v-if="viewingData.duration_preference" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">时长偏好</h4>
        <div class="flex h-5 rounded-full overflow-hidden mb-3">
          <div class="transition-all" :style="{ width: `${viewingData.duration_preference.short_ratio * 100}%`, backgroundColor: DURATION_COLORS['短视频'] }" title="短视频"></div>
          <div class="transition-all" :style="{ width: `${viewingData.duration_preference.mid_ratio * 100}%`, backgroundColor: DURATION_COLORS['中等视频'] }" title="中视频"></div>
          <div class="transition-all" :style="{ width: `${viewingData.duration_preference.long_ratio * 100}%`, backgroundColor: DURATION_COLORS['长视频'] }" title="长视频"></div>
        </div>
        <div class="space-y-1.5">
          <div class="flex items-center justify-between text-[11px]">
            <span class="flex items-center gap-1.5 text-gray-600 dark:text-gray-400"><span class="w-2.5 h-2.5 rounded-full inline-block" :style="{ backgroundColor: DURATION_COLORS['短视频'] }"></span>短视频(&lt;5min)</span>
            <span class="font-semibold text-gray-700 dark:text-gray-300">{{ viewingData.duration_preference.short }}</span>
          </div>
          <div class="flex items-center justify-between text-[11px]">
            <span class="flex items-center gap-1.5 text-gray-600 dark:text-gray-400"><span class="w-2.5 h-2.5 rounded-full inline-block" :style="{ backgroundColor: DURATION_COLORS['中等视频'] }"></span>中视频(5-20min)</span>
            <span class="font-semibold text-gray-700 dark:text-gray-300">{{ viewingData.duration_preference.mid }}</span>
          </div>
          <div class="flex items-center justify-between text-[11px]">
            <span class="flex items-center gap-1.5 text-gray-600 dark:text-gray-400"><span class="w-2.5 h-2.5 rounded-full inline-block" :style="{ backgroundColor: DURATION_COLORS['长视频'] }"></span>长视频(&gt;20min)</span>
            <span class="font-semibold text-gray-700 dark:text-gray-300">{{ viewingData.duration_preference.long }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="viewingData.weekday_distribution?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
      <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <div class="w-8 h-8 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center">
          <svg class="w-4 h-4 text-purple-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        周内分布
      </h4>
      <div class="space-y-2">
        <div v-for="(day, index) in viewingData.weekday_distribution" :key="day.name" class="flex items-center gap-3">
          <span class="text-xs text-gray-500 dark:text-gray-400 w-10 text-right shrink-0">{{ day.name }}</span>
          <div class="flex-1 h-6 bg-gray-100 dark:bg-gray-700 rounded-lg overflow-hidden">
            <div class="h-full rounded-lg transition-all duration-500 flex items-center justify-end pr-2" :style="{ width: `${Math.max((day.count / maxWeekdayCount) * 100, 8)}%`, backgroundColor: getChartColor(index % 7) }">
              <span v-if="day.count > 0" class="text-[10px] font-semibold text-white">{{ day.count }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="viewingData.top_videos?.length" class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
      <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <div class="w-8 h-8 rounded-lg bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center">
          <svg class="w-4 h-4 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        最长观看
      </h4>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <div
          v-for="(video, index) in viewingData.top_videos"
          :key="video.bvid"
          class="flex items-center gap-3 p-3 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors border border-gray-100 dark:border-gray-700/30"
        >
          <span class="w-6 h-6 rounded-lg flex items-center justify-center text-xs font-bold text-white shrink-0" :style="{ backgroundColor: getChartColor(index) }">{{ index + 1 }}</span>
          <img
            v-if="video.cover"
            :src="normalizeImageUrl(video.cover)"
            class="w-16 h-10 rounded-lg object-cover flex-shrink-0"
            loading="lazy"
          />
          <div class="flex-1 min-w-0">
            <div class="text-sm text-gray-700 dark:text-gray-300 truncate font-medium">{{ video.title }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ video.author_name }}</div>
          </div>
          <div class="text-right flex-shrink-0">
            <div class="text-sm font-bold" :style="{ color: getChartColor(index) }">{{ formatDurationShort(video.duration) }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl p-5 border border-gray-100 dark:border-gray-700/50 shadow-sm">
      <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <div class="w-8 h-8 rounded-lg bg-teal-100 dark:bg-teal-900/30 flex items-center justify-center">
          <svg class="w-4 h-4 text-teal-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
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
import { getChartColor, DEVICE_COLORS, DURATION_COLORS } from '~/utils/chartColors'

const props = defineProps({
  viewingData: {
    type: Object,
    required: true
  }
})

const viewingBehaviorData = ref(null)

const formatInsightText = (text) => {
  if (!text) return ''
  return text.replace(/(\d+(\.\d+)?)/g, '<span class="text-accent font-semibold">$1</span>')
}

const deviceEntries = computed(() => {
  if (!props.viewingData?.device_distribution) return []
  return Object.entries(props.viewingData.device_distribution).sort((a, b) => b[1] - a[1])
})

const getDeviceColor = (device) => {
  return DEVICE_COLORS[device] || '#9CA3AF'
}

const maxWeekdayCount = computed(() => {
  if (!props.viewingData?.weekday_distribution?.length) return 1
  return Math.max(...props.viewingData.weekday_distribution.map(d => d.count), 1)
})

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

watch(() => props.viewingData?.year, (newYear) => {
  if (newYear) {
    fetchViewingBehavior(newYear)
  }
}, { immediate: true })
</script>
