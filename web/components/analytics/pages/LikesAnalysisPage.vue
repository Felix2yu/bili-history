<!-- 点赞分析页组件 -->
<template>
  <div class="space-y-6">
    <h3 class="text-center text-3xl font-bold text-accent">点赞分析</h3>

    <div v-if="viewingData" class="text-center text-base text-gray-600 dark:text-gray-300">
      共点赞了 <span class="font-bold text-accent">{{ viewingData.total_count || 0 }}</span> 个视频
    </div>

    <div v-if="viewingData" class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <!-- Top 创作者 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-3 text-lg font-bold text-accent">最爱点赞的UP主</h4>
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

      <!-- 分类分布 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-3 text-lg font-bold text-accent">点赞分类分布</h4>
        <div class="space-y-2">
          <div
            v-for="(cat, index) in viewingData.category_dist"
            :key="index"
            class="flex items-center justify-between border-b border-gray-200/50 py-1.5 last:border-0 dark:border-gray-700/50"
          >
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ cat.name }}</span>
            <div class="flex items-center">
              <div class="mr-2 h-2 w-24 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
                <div
                  class="h-full rounded-full bg-accent"
                  :style="{ width: `${(cat.count / maxCategoryCount) * 100}%` }"
                />
              </div>
              <span class="text-sm font-medium text-accent">{{ cat.count }}</span>
            </div>
          </div>
          <div
            v-if="!viewingData.category_dist?.length"
            class="py-4 text-center text-sm text-gray-500 dark:text-gray-400"
          >
            暂无数据
          </div>
        </div>
      </div>

      <!-- 时长分布 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-3 text-lg font-bold text-accent">点赞视频时长分布</h4>
        <div class="space-y-2">
          <div
            v-for="(dur, index) in viewingData.duration_dist"
            :key="index"
            class="flex items-center justify-between border-b border-gray-200/50 py-1.5 last:border-0 dark:border-gray-700/50"
          >
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ dur.range }}</span>
            <div class="flex items-center">
              <div class="mr-2 h-2 w-24 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
                <div
                  class="h-full rounded-full bg-accent"
                  :style="{ width: `${(dur.count / maxDurationCount) * 100}%` }"
                />
              </div>
              <span class="text-sm font-medium text-accent">{{ dur.count }}</span>
            </div>
          </div>
          <div
            v-if="!viewingData.duration_dist?.length"
            class="py-4 text-center text-sm text-gray-500 dark:text-gray-400"
          >
            暂无数据
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

const maxCategoryCount = computed(() => {
  if (!props.viewingData?.category_dist?.length) return 1;
  return Math.max(...props.viewingData.category_dist.map((c) => c.count));
});

const maxDurationCount = computed(() => {
  if (!props.viewingData?.duration_dist?.length) return 1;
  return Math.max(...props.viewingData.duration_dist.map((d) => d.count));
});
</script>
