<!-- 收藏分析页组件 -->
<template>
  <div class="space-y-6">
    <h3 class="text-3xl font-bold text-center bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent">
      收藏分析
    </h3>

    <div v-if="viewingData" class="text-base text-center text-gray-600 dark:text-gray-300">
      共收藏了 <span class="text-accent font-bold">{{ viewingData.total_count || 0 }}</span> 个视频，
      分布在 <span class="text-[#fc9b7a] font-bold">{{ viewingData.folder_count || 0 }}</span> 个收藏夹中
    </div>

    <div v-if="viewingData" class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- 收藏夹分布 -->
      <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">收藏夹分布</h4>
        <div class="space-y-2">
          <div v-for="(folder, index) in viewingData.folder_dist" :key="index"
            class="flex items-center justify-between py-1.5 border-b border-gray-200/50 dark:border-gray-700/50 last:border-0">
            <div class="flex items-center flex-1 min-w-0">
              <span class="w-6 h-6 rounded-full bg-gradient-to-r from-accent to-accent/70 flex items-center justify-center text-white text-xs font-bold mr-2 flex-shrink-0">
                {{ index + 1 }}
              </span>
              <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ folder.title || '未命名收藏夹' }}</span>
            </div>
            <div class="flex items-center ml-2">
              <div class="w-20 h-2 bg-gray-200 dark:bg-gray-700 rounded-full mr-2 overflow-hidden">
                <div class="h-full bg-gradient-to-r from-accent to-accent/70 rounded-full"
                  :style="{ width: `${(folder.count / maxFolderCount) * 100}%` }"></div>
              </div>
              <span class="text-sm font-medium text-accent">{{ folder.count }}</span>
            </div>
          </div>
          <div v-if="!viewingData.folder_dist?.length" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            暂无数据
          </div>
        </div>
      </div>

      <!-- Top 创作者 -->
      <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">最爱收藏的UP主</h4>
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

      <!-- 数据概览 -->
      <div class="lg:col-span-2 bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
        <h4 class="text-lg font-bold bg-gradient-to-r from-accent to-accent/70 bg-clip-text text-transparent mb-3">数据概览</h4>
        <div class="grid grid-cols-1 gap-4">
          <div class="text-center">
            <div class="text-2xl font-bold text-[#fc9b7a]">{{ Math.round(viewingData.avg_collect_count || 0) }}</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">平均收藏数</div>
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

const maxFolderCount = computed(() => {
  if (!props.viewingData?.folder_dist?.length) return 1
  return Math.max(...props.viewingData.folder_dist.map(f => f.count))
})
</script>
