<!-- 收藏分析页组件 -->
<template>
  <div class="space-y-6">
    <h3 class="text-center text-3xl font-bold text-accent">收藏分析</h3>

    <div v-if="viewingData" class="text-center text-base text-gray-600 dark:text-gray-300">
      共收藏了
      <span class="font-bold text-accent">{{ viewingData.total_count || 0 }}</span> 个视频， 分布在
      <span class="font-bold text-[#fc9b7a]">{{ viewingData.folder_count || 0 }}</span> 个收藏夹中
    </div>

    <div v-if="viewingData" class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <!-- 收藏夹分布 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-3 text-lg font-bold text-accent">收藏夹分布</h4>
        <div class="space-y-2">
          <div
            v-for="(folder, index) in viewingData.folder_dist"
            :key="index"
            class="flex items-center justify-between border-b border-gray-200/50 py-1.5 last:border-0 dark:border-gray-700/50"
          >
            <div class="flex min-w-0 flex-1 items-center">
              <span
                class="mr-2 flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-accent text-xs font-bold text-white"
              >
                {{ index + 1 }}
              </span>
              <span class="truncate text-sm text-gray-700 dark:text-gray-300">{{
                folder.title || '未命名收藏夹'
              }}</span>
            </div>
            <div class="ml-2 flex items-center">
              <div class="mr-2 h-2 w-20 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
                <div
                  class="h-full rounded-full bg-accent"
                  :style="{ width: `${(folder.count / maxFolderCount) * 100}%` }"
                />
              </div>
              <span class="text-sm font-medium text-accent">{{ folder.count }}</span>
            </div>
          </div>
          <div
            v-if="!viewingData.folder_dist?.length"
            class="py-4 text-center text-sm text-gray-500 dark:text-gray-400"
          >
            暂无数据
          </div>
        </div>
      </div>

      <!-- Top 创作者 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-3 text-lg font-bold text-accent">最爱收藏的UP主</h4>
        <div class="space-y-2">
          <div
            v-for="(creator, index) in viewingData.top_creators"
            :key="index"
            class="flex items-center justify-between border-b border-gray-200/50 py-1.5 last:border-0 dark:border-gray-700/50"
          >
            <div class="flex items-center">
              <span
                class="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-accent text-xs font-bold text-white"
              >
                {{ index + 1 }}
              </span>
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ creator.name }}</span>
            </div>
            <span class="text-sm font-medium text-accent">{{ creator.count }} 个</span>
          </div>
          <div
            v-if="!viewingData.top_creators?.length"
            class="py-4 text-center text-sm text-gray-500 dark:text-gray-400"
          >
            暂无数据
          </div>
        </div>
      </div>

      <!-- 数据概览 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm lg:col-span-2 dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-3 text-lg font-bold text-accent">数据概览</h4>
        <div class="grid grid-cols-1 gap-4">
          <div class="text-center">
            <div class="text-2xl font-bold text-[#fc9b7a]">
              {{ Math.round(viewingData.avg_collect_count || 0) }}
            </div>
            <div class="text-sm text-gray-500 dark:text-gray-400">平均收藏数</div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="py-8 text-center text-gray-500 dark:text-gray-400">加载中...</div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  viewingData: {
    type: Object,
    default: null
  }
});

const maxFolderCount = computed(() => {
  if (!props.viewingData?.folder_dist?.length) return 1;
  return Math.max(...props.viewingData.folder_dist.map((f) => f.count));
});
</script>
