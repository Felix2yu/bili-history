<template>
  <div class="space-y-4">
    <!-- 核心统计卡片 -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ summary.total_videos }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">观看视频数</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ formatDurationShort(summary.total_duration) }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">总观看时长</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ summary.unique_days }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">活跃天数</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ summary.unique_authors }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">UP主数量</div>
      </div>
    </div>

    <!-- 日均统计 -->
    <div class="glass-card p-4 flex items-center justify-between text-sm">
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
        <span class="text-gray-600 dark:text-gray-400">日均</span>
      </div>
      <div class="flex gap-6">
        <span class="text-gray-700 dark:text-gray-300">
          <span class="font-semibold text-[#fb7299]">{{ summary.avg_daily_videos?.toFixed(1) }}</span> 个视频
        </span>
        <span class="text-gray-700 dark:text-gray-300">
          <span class="font-semibold text-[#fb7299]">{{ formatDurationShort(summary.avg_daily_duration) }}</span> 观看时长
        </span>
      </div>
    </div>

    <!-- Top 分区 & Top UP主 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <!-- Top 分区 -->
      <div v-if="summary.top_categories?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z" />
          </svg>
          常看分区
        </h4>
        <div class="space-y-2">
          <div
            v-for="(cat, index) in summary.top_categories.slice(0, 5)"
            :key="cat.name"
            class="flex items-center gap-2"
          >
            <span class="text-xs text-gray-400 w-4 text-right">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ cat.name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ cat.count }}次</span>
              </div>
              <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] rounded-full"
                  :style="{ width: `${(cat.count / summary.top_categories[0].count) * 100}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top UP主 -->
      <div v-if="summary.top_authors?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          常看UP主
        </h4>
        <div class="space-y-2">
          <div
            v-for="(author, index) in summary.top_authors.slice(0, 5)"
            :key="author.mid"
            class="flex items-center gap-2"
          >
            <span class="text-xs text-gray-400 w-4 text-right">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ author.name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ author.count }}次</span>
              </div>
              <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] rounded-full"
                  :style="{ width: `${(author.count / summary.top_authors[0].count) * 100}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 设备分布 -->
    <div v-if="summary.device_dist && Object.keys(summary.device_dist).length > 0" class="glass-card p-4">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
        </svg>
        设备分布
      </h4>
      <div class="flex gap-4 flex-wrap">
        <div
          v-for="(count, device) in summary.device_dist"
          :key="device"
          class="flex items-center gap-2 text-sm"
        >
          <span class="text-gray-600 dark:text-gray-400">{{ device }}</span>
          <span class="font-semibold text-[#fb7299]">{{ count }}</span>
          <span class="text-xs text-gray-400">({{ ((count / summary.total_videos) * 100).toFixed(0) }}%)</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { formatDurationShort } from '~/utils/format'

defineProps({
  summary: {
    type: Object,
    required: true,
  },
})
</script>

<style scoped>
.glass-card {
  @apply bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm
    rounded-xl border border-gray-200/50 dark:border-gray-700/50
    shadow-sm;
}
</style>
