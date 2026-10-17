<!-- UP主完成率分析页组件 -->
<template>
  <div v-if="viewingData" class="space-y-4">
    <div class="mx-auto w-full max-w-7xl px-2 py-4">
      <div class="space-y-4">
        <h3 class="text-center text-3xl font-bold text-accent">UP主完成率分析</h3>

        <div class="mb-2 space-y-1 px-4 text-center text-sm text-gray-800 dark:text-gray-200">
          <div
            v-if="viewingData?.insights?.most_valuable_author"
            v-html="formatInsightText(viewingData.insights.most_valuable_author)"
          />
          <div
            v-if="viewingData?.insights?.highest_completion_author"
            v-html="formatInsightText(viewingData.insights.highest_completion_author)"
          />
          <div
            v-if="viewingData?.insights?.potential_discovery"
            v-html="formatInsightText(viewingData.insights.potential_discovery)"
          />
          <div
            v-if="viewingData?.insights?.viewing_behavior_summary"
            v-html="formatInsightText(viewingData.insights.viewing_behavior_summary)"
          />
        </div>

        <!-- 图表容器 -->
        <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
          <!-- 最喜欢的UP主 -->
          <div
            class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
          >
            <h4 class="mb-3 text-center text-base font-bold text-accent">最喜欢的UP主</h4>
            <div class="space-y-2">
              <div
                v-for="(item, index) in topFavoriteAuthors"
                :key="item.name"
                class="flex items-center gap-2"
              >
                <span class="w-4 flex-shrink-0 text-right text-xs text-gray-400">{{
                  index + 1
                }}</span>
                <div class="min-w-0 flex-1">
                  <div class="mb-0.5 flex items-center justify-between">
                    <span
                      class="cursor-pointer truncate text-sm text-gray-700 hover:text-accent dark:text-gray-300"
                      @click="handleAuthorClick(item.mid)"
                      >{{ item.name }}</span
                    >
                    <span class="ml-2 flex-shrink-0 text-xs text-gray-500 dark:text-gray-400"
                      >{{ item.score.toFixed(1) }}分</span
                    >
                  </div>
                  <div class="h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
                    <div
                      class="h-full rounded-full bg-accent"
                      :style="{ width: `${(item.score / topScore) * 100}%` }"
                    />
                  </div>
                  <div class="mt-0.5 flex gap-3 text-[0.625rem] text-gray-400 dark:text-gray-500">
                    <span>完播 {{ item.completion }}%</span>
                    <span>视频 {{ item.videos }}个</span>
                    <span>喜爱 {{ item.loyalty.toFixed(0) }}</span>
                  </div>
                </div>
              </div>
              <div
                v-if="!topFavoriteAuthors.length"
                class="py-4 text-center text-sm text-gray-400 dark:text-gray-500"
              >
                暂无数据
              </div>
            </div>
          </div>

          <!-- 观看最多的UP主 -->
          <div
            class="rounded-xl border border-gray-300/50 bg-white/50 p-4 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
          >
            <h4 class="mb-3 text-center text-base font-bold text-accent">观看最多的UP主</h4>
            <div class="space-y-2">
              <div
                v-for="(item, index) in topWatchedAuthors"
                :key="item.name"
                class="flex items-center gap-2"
              >
                <span class="w-4 flex-shrink-0 text-right text-xs text-gray-400">{{
                  index + 1
                }}</span>
                <div class="min-w-0 flex-1">
                  <div class="mb-0.5 flex items-center justify-between">
                    <span
                      class="cursor-pointer truncate text-sm text-gray-700 hover:text-accent dark:text-gray-300"
                      @click="handleAuthorClick(item.mid)"
                      >{{ item.name }}</span
                    >
                    <span class="ml-2 flex-shrink-0 text-xs text-gray-500 dark:text-gray-400"
                      >{{ item.videos }}个视频</span
                    >
                  </div>
                  <div class="h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
                    <div
                      class="h-full rounded-full bg-blue-500"
                      :style="{ width: `${(item.videos / topVideoCount) * 100}%` }"
                    />
                  </div>
                  <div class="mt-0.5 flex gap-3 text-[0.625rem] text-gray-400 dark:text-gray-500">
                    <span>完播 {{ item.completion }}%</span>
                    <span>完整观看 {{ item.fullyWatched }}个</span>
                  </div>
                </div>
              </div>
              <div
                v-if="!topWatchedAuthors.length"
                class="py-4 text-center text-sm text-gray-400 dark:text-gray-500"
              >
                暂无数据
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { escapeHtml } from '~/utils/escapeHtml';

const props = defineProps({
  viewingData: {
    type: Object,
    required: true
  }
});

// 最喜欢的UP主数据
const topFavoriteAuthors = computed(() => {
  if (!props.viewingData?.completion_rates?.most_valuable_authors) return [];
  return Object.entries(props.viewingData.completion_rates.most_valuable_authors)
    .sort((a, b) => b[1].comprehensive_score - a[1].comprehensive_score)
    .slice(0, 8)
    .map(([name, stats]) => ({
      name,
      mid: stats.author_mid,
      score: stats.comprehensive_score,
      completion: stats.average_completion_rate?.toFixed(1) || '0',
      videos: stats.video_count,
      loyalty: stats.loyalty_score
    }));
});

const topScore = computed(() => {
  if (!topFavoriteAuthors.value.length) return 1;
  return topFavoriteAuthors.value[0].score;
});

// 观看最多的UP主数据
const topWatchedAuthors = computed(() => {
  if (!props.viewingData?.completion_rates?.most_watched_authors) return [];
  return Object.entries(props.viewingData.completion_rates.most_watched_authors)
    .sort((a, b) => b[1].video_count - a[1].video_count)
    .slice(0, 8)
    .map(([name, stats]) => ({
      name,
      mid: stats.author_mid,
      videos: stats.video_count,
      completion: stats.average_completion_rate?.toFixed(1) || '0',
      fullyWatched: stats.fully_watched
    }));
});

const topVideoCount = computed(() => {
  if (!topWatchedAuthors.value.length) return 1;
  return topWatchedAuthors.value[0].videos;
});

const handleAuthorClick = (mid) => {
  if (mid) window.open(`https://space.bilibili.com/${mid}`, '_blank');
};

const formatInsightText = (text) => {
  if (!text) return '';
  return escapeHtml(text).replace(/(\d+(\.\d+)?)/g, '<span class="text-accent">$1</span>');
};
</script>
