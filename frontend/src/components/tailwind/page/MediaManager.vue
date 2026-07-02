<template>
  <div class="min-h-screen bg-gray-50/30 dark:bg-gray-900">
    <div class="py-6">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <!-- 主内容卡片 -->
        <div class="glass-card overflow-hidden">
          <!-- 标签导航 - pill style -->
          <div class="px-3 py-2">
            <nav class="flex gap-1 overflow-x-auto" aria-label="媒体管理选项卡">
              <button
                @click="activeTab = 'dynamic'"
                class="py-2 px-3 rounded-xl text-sm font-medium flex items-center gap-2 transition-all duration-200 whitespace-nowrap"
                :class="activeTab === 'dynamic'
                  ? 'bg-accent/10 text-accent'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-white/10 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300'"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                <span>动态下载</span>
              </button>

              <button
                @click="activeTab = 'video-download'"
                class="py-2 px-3 rounded-xl text-sm font-medium flex items-center gap-2 transition-all duration-200 whitespace-nowrap"
                :class="activeTab === 'video-download'
                  ? 'bg-accent/10 text-accent'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-white/10 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300'"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                <span>视频下载</span>
              </button>

              <button
                @click="activeTab = 'downloaded'"
                class="py-2 px-3 rounded-xl text-sm font-medium flex items-center gap-2 transition-all duration-200 whitespace-nowrap"
                :class="activeTab === 'downloaded'
                  ? 'bg-accent/10 text-accent'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-white/10 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300'"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                <span>已下载</span>
              </button>

              <button
                @click="activeTab = 'images'"
                class="py-2 px-3 rounded-xl text-sm font-medium flex items-center gap-2 transition-all duration-200 whitespace-nowrap"
                :class="activeTab === 'images'
                  ? 'bg-accent/10 text-accent'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-white/10 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300'"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <span>图片管理</span>
              </button>

              <button
                @click="activeTab = 'details'"
                class="py-2 px-3 rounded-xl text-sm font-medium flex items-center gap-2 transition-all duration-200 whitespace-nowrap"
                :class="activeTab === 'details'
                  ? 'bg-accent/10 text-accent'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-white/10 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300'"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>视频详情</span>
              </button>
            </nav>
          </div>

          <!-- 内容区域 -->
          <div class="transition-all duration-300 p-5">
            <!-- 动态下载 -->
            <div v-if="activeTab === 'dynamic'" class="animate-fadeIn">
              <DynamicDownloader />
            </div>

            <!-- 视频下载 -->
            <div v-if="activeTab === 'video-download'" class="animate-fadeIn">
              <VideoDownloader />
            </div>

            <!-- 已下载 -->
            <div v-if="activeTab === 'downloaded'" class="animate-fadeIn">
              <Downloads />
            </div>

            <!-- 图片管理 -->
            <div v-if="activeTab === 'images'" class="animate-fadeIn">
              <Images />
            </div>

            <!-- 视频详情管理 -->
            <div v-if="activeTab === 'details'" class="animate-fadeIn">
              <VideoDetailsManager />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import Images from './Images.vue'
import VideoDetailsManager from './VideoDetailsManager.vue'
import DynamicDownloader from './DynamicDownloader.vue'
import VideoDownloader from './VideoDownloader.vue'
import Downloads from './Downloads.vue'

const route = useRoute()

// 当前激活的标签
const activeTab = ref('dynamic')

// 监听路由变化以更新激活的标签
watch(
  () => route.query.tab,
  (tab) => {
    if (tab && ['images', 'dynamic', 'details', 'video-download', 'downloaded'].includes(tab)) {
      activeTab.value = tab
    }
  },
  { immediate: true }
)

// 组件挂载时根据URL初始化标签
onMounted(() => {
  const { tab } = route.query
  if (tab && ['images', 'dynamic', 'details', 'video-download', 'downloaded'].includes(tab)) {
    activeTab.value = tab
  }
})
</script>

<style scoped>
.animate-fadeIn {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>