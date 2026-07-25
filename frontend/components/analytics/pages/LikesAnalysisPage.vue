<!-- 点赞分析页组件 -->
<template>
  <div class="space-y-6">
    <h3 class="text-3xl font-bold text-center bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent">
      点赞分析
    </h3>

    <div v-if="viewingData" class="text-base text-center text-gray-600 dark:text-gray-300">
      共点赞了 <span class="text-accent font-bold">{{ viewingData.total_count || 0 }}</span> 个视频
    </div>

    <div v-if="viewingData" class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- Top 创作者 -->
      <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">最爱点赞的UP主</h4>
        <div class="space-y-2">
          <div v-for="(creator, index) in viewingData.top_creators" :key="index"
            class="flex items-center justify-between py-1.5 border-b border-gray-200/50 dark:border-gray-700/50 last:border-0">
            <div class="flex items-center">
              <span class="w-6 h-6 rounded-full bg-gradient-to-r from-accent to-accent/70 flex items-center justify-center text-white text-xs font-bold mr-2">
                {{ index + 1 }}
              </span>
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ creator.name }}</span>
            </div>
            <span class="text-sm font-medium text-accent">{{ creator.count }} 个</span>
          </div>
          <div v-if="!viewingData.top_creators?.length" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            暂无数据
          </div>
        </div>
      </div>

      <!-- 分类分布 -->
      <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">点赞分类分布</h4>
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
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">点赞视频时长分布</h4>
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
