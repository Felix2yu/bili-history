<!-- 稍后再看分析页组件 -->
<template>
  <div class="space-y-6">
    <h3 class="text-3xl font-bold text-center bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent">
      稍后再看分析
    </h3>

    <div v-if="viewingData" class="text-base text-center text-gray-600 dark:text-gray-300">
      稍后再看列表中有 <span class="text-accent font-bold">{{ viewingData.total_count || 0 }}</span> 个视频
    </div>

    <div v-if="viewingData" class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- 分类分布 -->
      <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">分类分布</h4>
        <div class="space-y-2">
          <div v-for="(cat, index) in viewingData.category_dist" :key="index"
            class="flex items-center justify-between py-1.5 border-b border-gray-200/50 dark:border-gray-700/50 last:border-0">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ cat.name }}</span>
            <div class="flex items-center">
              <div class="w-24 h-2 bg-gray-200 dark:bg-gray-700 rounded-full mr-2 overflow-hidden">
                <div class="h-full bg-gradient-to-r from-accent to-accent/70 rounded-full"
                  :style="{ width: `${(cat.count / maxCategoryCount) * 100}%` }"></div>
              </div>
              <span class="text-sm font-medium text-accent">{{ cat.count }}</span>
            </div>
          </div>
          <div v-if="!viewingData.category_dist?.length" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            暂无数据
          </div>
        </div>
      </div>

      <!-- 时长分布 -->
      <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">时长分布</h4>
        <div class="space-y-2">
          <div v-for="(dur, index) in viewingData.duration_dist" :key="index"
            class="flex items-center justify-between py-1.5 border-b border-gray-200/50 dark:border-gray-700/50 last:border-0">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ dur.range }}</span>
            <div class="flex items-center">
              <div class="w-24 h-2 bg-gray-200 dark:bg-gray-700 rounded-full mr-2 overflow-hidden">
                <div class="h-full bg-gradient-to-r from-[#fc9b7a] to-accent rounded-full"
                  :style="{ width: `${(dur.count / maxDurationCount) * 100}%` }"></div>
              </div>
              <span class="text-sm font-medium text-accent">{{ dur.count }}</span>
            </div>
          </div>
          <div v-if="!viewingData.duration_dist?.length" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            暂无数据
          </div>
        </div>
      </div>

      <!-- 最早的条目 -->
      <div class="lg:col-span-2 bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">积灰最久的视频</h4>
        <div class="space-y-2">
          <div v-for="(item, index) in viewingData.oldest_items" :key="index"
            class="flex items-center justify-between py-2 border-b border-gray-200/50 dark:border-gray-700/50 last:border-0">
            <div class="flex items-center flex-1 min-w-0">
              <span class="w-6 h-6 rounded-full bg-gradient-to-r from-accent to-accent/70 flex items-center justify-center text-white text-xs font-bold mr-2 flex-shrink-0">
                {{ index + 1 }}
              </span>
              <div class="min-w-0">
                <div class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ item.title }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.owner_name }}</div>
              </div>
            </div>
            <span class="text-sm font-medium text-accent ml-2 whitespace-nowrap">{{ item.days_ago }}天前</span>
          </div>
          <div v-if="!viewingData.oldest_items?.length" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            暂无数据
          </div>
        </div>
      </div>

    </div>

    <div v-else class="text-center text-gray-500 dark:text-gray-400 py-8">
      加载中...
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  viewingData: {
    type: Object,
    default: null
  }
})

const maxCategoryCount = computed(() => {
  if (!props.viewingData?.category_dist?.length) return 1
  return Math.max(...props.viewingData.category_dist.map(c => c.count))
})

const maxDurationCount = computed(() => {
  if (!props.viewingData?.duration_dist?.length) return 1
  return Math.max(...props.viewingData.duration_dist.map(d => d.count))
})
</script>
