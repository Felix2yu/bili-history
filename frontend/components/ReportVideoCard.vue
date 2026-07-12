<template>
  <div class="glass-card p-3 flex gap-3 hover:shadow-md transition-shadow duration-200">
    <!-- 封面 -->
    <div class="relative w-32 sm:w-40 flex-shrink-0 aspect-video rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-800">
      <img
        v-if="video.cover"
        :src="video.cover"
        :alt="video.title"
        class="w-full h-full object-cover"
        loading="lazy"
      />
      <div v-else class="w-full h-full flex items-center justify-center text-gray-400">
        <svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z" />
        </svg>
      </div>
      <!-- 时长角标 -->
      <div class="absolute bottom-1 right-1 bg-black/70 text-white text-[10px] px-1.5 py-0.5 rounded">
        {{ formatDuration(video.duration) }}
      </div>
    </div>

    <!-- 信息 -->
    <div class="flex-1 min-w-0 flex flex-col justify-between">
      <div>
        <!-- 标题 -->
        <h4 class="text-sm font-medium text-gray-800 dark:text-gray-200 line-clamp-2 leading-snug">
          {{ video.title }}
        </h4>

        <!-- UP主 -->
        <div class="flex items-center gap-1.5 mt-1.5">
          <img
            v-if="video.author_face"
            :src="video.author_face"
            class="w-4 h-4 rounded-full object-cover"
            loading="lazy"
          />
          <span class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ video.author_name }}</span>
        </div>

        <!-- 标签行 -->
        <div class="flex items-center gap-1.5 mt-1.5 flex-wrap">
          <span
            v-if="video.main_category || video.tag_name"
            class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-[#fb7299]/10 text-[#fb7299]"
          >
            {{ video.main_category || video.tag_name }}
          </span>
          <span class="text-[10px] text-gray-400 dark:text-gray-500">
            {{ getDeviceName(video.dt) }}
          </span>
          <span class="text-[10px] text-gray-400 dark:text-gray-500">
            {{ getBusinessType(video.business) }}
          </span>
        </div>
      </div>

      <!-- 底部信息 -->
      <div class="flex items-center justify-between mt-2">
        <div class="flex items-center gap-3 text-[11px] text-gray-400 dark:text-gray-500">
          <span>{{ formatTimestamp(video.view_at) }}</span>
        </div>

        <!-- 完播率 -->
        <div class="flex items-center gap-1.5">
          <div class="w-16 h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="completionRate >= 90 ? 'bg-green-500' : completionRate >= 50 ? 'bg-[#fb7299]' : 'bg-gray-400'"
              :style="{ width: `${Math.min(completionRate, 100)}%` }"
            ></div>
          </div>
          <span class="text-[10px] text-gray-400 dark:text-gray-500">{{ completionRateText }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { formatDuration, formatTimestamp, getBusinessType } from '~/utils/format'

const props = defineProps({
  video: {
    type: Object,
    required: true,
  },
})

const completionRate = computed(() => {
  if (props.video.duration <= 0) return 0
  if (props.video.progress === -1) return 100
  return Math.round((props.video.progress / props.video.duration) * 100)
})

const completionRateText = computed(() => {
  if (props.video.progress === -1) return '看完'
  if (completionRate.value >= 90) return '看完'
  return `${completionRate.value}%`
})

function getDeviceName(dt) {
  switch (dt) {
    case 1: case 3: case 5: case 7: case 33: return '手机'
    case 2: return '电脑'
    case 4: case 6: return '平板'
    default: return '其他'
  }
}
</script>

<style scoped>
.glass-card {
  @apply bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm
    rounded-xl border border-gray-200/50 dark:border-gray-700/50
    shadow-sm;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
