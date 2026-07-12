<!-- UP主完成率分析页组件 -->
<template>
  <div class="space-y-4" v-if="viewingData">
    <div class="max-w-7xl w-full mx-auto px-2 py-4">
        <div class="space-y-4">
          <h3 class="text-3xl font-bold text-center bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] bg-clip-text text-transparent">
            UP主完成率分析
          </h3>

          <div class="text-sm text-center text-gray-800 dark:text-gray-200 mb-2 space-y-1 px-4">
            <div v-if="viewingData?.insights?.most_valuable_author" v-html="formatInsightText(viewingData.insights.most_valuable_author)">
            </div>
            <div v-if="viewingData?.insights?.highest_completion_author" v-html="formatInsightText(viewingData.insights.highest_completion_author)">
            </div>
            <div v-if="viewingData?.insights?.potential_discovery" v-html="formatInsightText(viewingData.insights.potential_discovery)">
            </div>
            <div v-if="viewingData?.insights?.viewing_behavior_summary" v-html="formatInsightText(viewingData.insights.viewing_behavior_summary)">
            </div>
          </div>

          <!-- 图表容器 -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
            <!-- 最喜欢的UP主 -->
            <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
              <h4 class="text-base font-bold text-center bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] bg-clip-text text-transparent mb-3">
                最喜欢的UP主
              </h4>
              <div class="space-y-2">
                <div v-for="(item, index) in topFavoriteAuthors" :key="item.name" class="flex items-center gap-2">
                  <span class="text-xs text-gray-400 w-4 text-right flex-shrink-0">{{ index + 1 }}</span>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between mb-0.5">
                      <span class="text-sm text-gray-700 dark:text-gray-300 truncate cursor-pointer hover:text-[#fb7299]" @click="handleAuthorClick(item.mid)">{{ item.name }}</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 ml-2 flex-shrink-0">{{ item.score.toFixed(1) }}分</span>
                    </div>
                    <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                      <div class="h-full bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] rounded-full" :style="{ width: `${(item.score / topScore) * 100}%` }"></div>
                    </div>
                    <div class="flex gap-3 mt-0.5 text-[10px] text-gray-400 dark:text-gray-500">
                      <span>完播 {{ item.completion }}%</span>
                      <span>视频 {{ item.videos }}个</span>
                      <span>喜爱 {{ item.loyalty.toFixed(0) }}</span>
                    </div>
                  </div>
                </div>
                <div v-if="!topFavoriteAuthors.length" class="text-center text-gray-400 dark:text-gray-500 py-4 text-sm">暂无数据</div>
              </div>
            </div>

            <!-- 观看最多的UP主 -->
            <div class="bg-white/50 dark:bg-white/5 backdrop-blur-sm rounded-xl p-4 border border-gray-300/50 dark:border-gray-500/50">
              <h4 class="text-base font-bold text-center bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] bg-clip-text text-transparent mb-3">
                观看最多的UP主
              </h4>
              <div class="space-y-2">
                <div v-for="(item, index) in topWatchedAuthors" :key="item.name" class="flex items-center gap-2">
                  <span class="text-xs text-gray-400 w-4 text-right flex-shrink-0">{{ index + 1 }}</span>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between mb-0.5">
                      <span class="text-sm text-gray-700 dark:text-gray-300 truncate cursor-pointer hover:text-[#fb7299]" @click="handleAuthorClick(item.mid)">{{ item.name }}</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 ml-2 flex-shrink-0">{{ item.videos }}个视频</span>
                    </div>
                    <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                      <div class="h-full bg-gradient-to-r from-[#40a9ff] to-[#69c0ff] rounded-full" :style="{ width: `${(item.videos / topVideoCount) * 100}%` }"></div>
                    </div>
                    <div class="flex gap-3 mt-0.5 text-[10px] text-gray-400 dark:text-gray-500">
                      <span>完播 {{ item.completion }}%</span>
                      <span>完整观看 {{ item.fullyWatched }}个</span>
                    </div>
                  </div>
                </div>
                <div v-if="!topWatchedAuthors.length" class="text-center text-gray-400 dark:text-gray-500 py-4 text-sm">暂无数据</div>
              </div>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  viewingData: {
    type: Object,
    required: true
  }
})

// 最喜欢的UP主数据
const topFavoriteAuthors = computed(() => {
  if (!props.viewingData?.completion_rates?.most_valuable_authors) return []
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
    }))
})

const topScore = computed(() => {
  if (!topFavoriteAuthors.value.length) return 1
  return topFavoriteAuthors.value[0].score
})

// 观看最多的UP主数据
const topWatchedAuthors = computed(() => {
  if (!props.viewingData?.completion_rates?.most_watched_authors) return []
  return Object.entries(props.viewingData.completion_rates.most_watched_authors)
    .sort((a, b) => b[1].video_count - a[1].video_count)
    .slice(0, 8)
    .map(([name, stats]) => ({
      name,
      mid: stats.author_mid,
      videos: stats.video_count,
      completion: stats.average_completion_rate?.toFixed(1) || '0',
      fullyWatched: stats.fully_watched
    }))
})

const topVideoCount = computed(() => {
  if (!topWatchedAuthors.value.length) return 1
  return topWatchedAuthors.value[0].videos
})

const handleAuthorClick = (mid) => {
  if (mid) window.open(`https://space.bilibili.com/${mid}`, '_blank')
}

const formatInsightText = (text) => {
  if (!text) return '';
  return text.replace(/(\d+(\.\d+)?)/g, '<span class="text-[#fb7299]">$1</span>')
}
</script>
