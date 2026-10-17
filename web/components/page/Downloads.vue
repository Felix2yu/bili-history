<template>
  <div class="overflow-y-auto">
    <!-- 搜索框 -->
    <div class="mb-6">
      <SimpleSearchBar
        v-model="searchTerm"
        placeholder="搜索已下载的视频或目录路径..."
        class="w-full"
        @search="loadDownloadedVideos"
      />
    </div>

    <!-- 主要内容 -->
    <div>
      <!-- 下载数据加载中的占位 -->
      <div v-if="isLoading" class="flex flex-col items-center justify-center py-12">
        <svg class="mb-4 h-8 w-8 animate-spin text-accent" fill="none" viewBox="0 0 24 24">
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <p class="text-gray-500 dark:text-gray-400">加载中，请稍候...</p>
      </div>

      <!-- 没有数据时的显示 -->
      <div
        v-else-if="!downloads.videos || downloads.videos.length === 0"
        class="flex flex-col items-center justify-center py-12"
      >
        <svg
          class="mb-4 h-16 w-16 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"
          />
        </svg>
        <p class="mb-2 text-xl font-medium text-gray-600 dark:text-gray-300">暂无下载记录</p>
        <p class="mb-4 text-center text-gray-500 dark:text-gray-400">
          你还没有下载任何视频，在浏览历史记录时可以点击"下载"按钮下载视频。
        </p>
      </div>

      <!-- 下载视频列表 -->
      <div v-else>
        <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">
          共找到 {{ downloads.total }} 个下载记录，当前第 {{ downloads.page }} 页，共
          {{ downloads.pages }} 页
        </p>

        <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
          <div
            v-for="(video, index) in downloads.videos"
            :key="index"
            class="group relative overflow-hidden rounded-md border border-gray-200/50 bg-white/50 backdrop-blur-sm transition-all duration-200 hover:border-accent hover:shadow-sm dark:border-gray-700/50 dark:bg-gray-800/50"
          >
            <!-- 视频封面 -->
            <div
              class="group relative cursor-pointer overflow-hidden pb-[56.25%]"
              @click="handleVideoClick(video)"
            >
              <img
                :src="
                  normalizeImageUrl(video.cover) ||
                  '/api/images/proxy?url=https%3A%2F%2Fi0.hdslb.com%2Fbfs%2Farchive%2Fc9e72655b7c9c9c68a30d3275313c501e68427d1.jpg'
                "
                :alt="video.title"
                class="absolute inset-0 h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                loading="lazy"
                onerror="
                  this.src =
                    '/api/images/proxy?url=https%3A%2F%2Fi0.hdslb.com%2Fbfs%2Farchive%2Fc9e72655b7c9c9c68a30d3275313c501e68427d1.jpg'
                "
              />

              <!-- 播放按钮覆盖 -->
              <div
                class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity group-hover:opacity-100"
              >
                <button
                  v-if="video.files && video.files.length > 0 && !video.files[0].is_audio_only"
                  class="flex h-8 w-8 items-center justify-center rounded-full bg-accent/80 text-white backdrop-blur-sm"
                  @click.stop="playVideo(video.files[0].file_path)"
                >
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                    />
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </button>
              </div>

              <!-- 文件信息标签 -->
              <div
                class="absolute right-1 bottom-1 flex items-center gap-1 rounded bg-black/60 px-1 py-0.5 text-[0.625rem] text-white backdrop-blur-sm"
              >
                <span
                  v-if="isUnsupportedFormat(video)"
                  class="rounded bg-amber-500/80 px-1 text-[0.5625rem]"
                  >不支持播放</span
                >
                <div v-if="video.files && video.files.length > 0">
                  {{ video.files[0].size_mb.toFixed(1) }} MB
                </div>
              </div>

              <!-- 删除按钮 -->
              <div
                class="absolute top-1.5 right-1.5 z-20 hidden h-6 w-6 cursor-pointer items-center justify-center rounded-md bg-[#7d7c75]/60 backdrop-blur-sm transition-all duration-200 group-hover:flex hover:bg-[#7d7c75]/80"
                @click.stop="handleDeleteVideo(video)"
              >
                <svg
                  class="h-3.5 w-3.5 text-white"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </div>

              <!-- 多文件角标 -->
              <div
                v-if="video.files && video.files.length > 1"
                class="absolute top-1 left-1 rounded bg-accent px-1 py-0.5 text-[0.625rem] text-white"
              >
                {{ video.files.length }}
              </div>
            </div>

            <!-- 视频信息 -->
            <div class="flex flex-col space-y-1 p-2">
              <!-- 标题 -->
              <div
                class="line-clamp-1 cursor-pointer text-xs font-medium text-gray-900 dark:text-gray-100"
                @click="handleVideoClick(video)"
              >
                {{ video.title }}
              </div>

              <!-- 作者信息 -->
              <div class="flex items-center space-x-1">
                <img
                  v-if="normalizeImageUrl(video.author_face)"
                  :src="normalizeImageUrl(video.author_face)"
                  :alt="video.author_name"
                  class="h-3.5 w-3.5 cursor-pointer rounded-full object-cover"
                  loading="lazy"
                  onerror="this.style.display = 'none'"
                  @click.stop="handleAuthorClick(video)"
                />
                <span
                  class="cursor-pointer truncate text-[0.625rem] text-gray-600 hover:text-accent dark:text-gray-400"
                  @click.stop="handleAuthorClick(video)"
                >
                  {{ video.author_name || '未知UP主' }}
                </span>
              </div>

              <!-- 下载时间 -->
              <div
                class="flex items-center justify-between text-[0.625rem] text-gray-500 dark:text-gray-400"
              >
                <div class="flex items-center space-x-1">
                  <svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                  <span>{{ video.download_date }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 分页 -->
        <div class="mt-8 flex justify-center">
          <Pagination
            :current-page="currentPage"
            :total-pages="downloads.pages"
            @page-change="handlePageChange"
          />
        </div>
      </div>
    </div>

    <!-- 视频播放对话框 -->
    <VideoPlayerDialog v-model:show="showVideoPlayer" :video-path="currentVideoPath" />

    <!-- 删除确认对话框 -->
    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center">
        <!-- 遮罩层 -->
        <div class="fixed inset-0 bg-black/50" @click="showDeleteConfirm = false" />

        <!-- 弹窗内容 -->
        <div
          class="relative z-10 w-[500px] rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800"
        >
          <h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-gray-100">确认删除视频</h3>

          <p class="mb-4 text-gray-600 dark:text-gray-400">
            确定要删除以下视频吗？此操作不可恢复。
          </p>

          <p class="mb-2 font-medium text-gray-800 dark:text-gray-100">
            {{ currentVideo?.title || '未知视频' }}
          </p>

          <!-- 显示CID和目录信息 -->
          <div class="mb-3 text-sm text-gray-500">
            <p v-if="currentVideo?.cid">CID: {{ currentVideo?.cid }}</p>

            <!-- 目录信息显示 -->
            <div class="mt-2">
              <p class="mb-1 text-gray-600 dark:text-gray-400">目录路径:</p>
              <div
                class="rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-sm break-all text-gray-700 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-300"
              >
                {{ getVideoDirectory(currentVideo) || '无法获取目录路径' }}
              </div>
            </div>
          </div>

          <!-- 删除选项 -->
          <div class="mb-4">
            <label class="flex cursor-pointer items-center space-x-2 select-none">
              <input
                v-model="deleteDirectory"
                type="checkbox"
                class="h-4 w-4 rounded border-gray-300 text-accent focus:ring-accent dark:border-gray-600"
              />
              <span>同时删除整个目录（包含所有相关文件）</span>
            </label>
          </div>

          <!-- 视频来源提示 (针对收藏夹视频) -->
          <div
            v-if="!currentVideo?.cid && getVideoDirectory(currentVideo)"
            class="mb-4 rounded-md border border-amber-200 bg-amber-50 p-2 dark:border-amber-800 dark:bg-amber-900/20"
          >
            <p class="text-sm text-amber-700">
              <span class="font-medium">提示：</span>
              该视频可能来自收藏夹批量下载，将使用目录路径进行删除。
            </p>
          </div>

          <!-- 按钮 -->
          <div class="mt-6 flex justify-end space-x-3">
            <button
              class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700"
              @click="showDeleteConfirm = false"
            >
              取消
            </button>
            <button
              class="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
              :disabled="isDeleting || (!currentVideo?.cid && !getVideoDirectory(currentVideo))"
              @click="confirmDeleteVideo"
            >
              {{ isDeleting ? '删除中...' : '确认删除' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useAsyncData } from '#imports';
import Pagination from '../Pagination.vue';
import VideoPlayerDialog from '../VideoPlayerDialog.vue';
import { getDownloadedVideos, deleteDownloadedVideo } from '~/utils/api';
import { showNotify } from 'vant';
import 'vant/es/notify/style';
import SimpleSearchBar from '../SimpleSearchBar.vue';
import { openInBrowser } from '~/utils/openUrl.js';
import { normalizeImageUrl } from '~/utils/imageUrl.js';

// 定义组件选项
defineOptions({
  name: 'Downloads'
});

// 状态变量
const isLoading = ref(true);
const downloads = ref({
  videos: [],
  total: 0,
  page: 1,
  limit: 20,
  pages: 1
});
const searchTerm = ref('');
const currentPage = ref(1);

// 视频播放相关
const showVideoPlayer = ref(false);
const currentVideoPath = ref('');

// 删除相关
const showDeleteConfirm = ref(false);
const currentVideo = ref(null);
const deleteDirectory = ref(true);
const isDeleting = ref(false);

// 加载已下载的视频
const loadDownloadedVideos = async () => {
  try {
    isLoading.value = true;
    const response = await getDownloadedVideos(searchTerm.value, currentPage.value, 20);

    if (response.data && response.data.status === 'success' && response.data.data) {
      downloads.value = {
        videos: response.data.data.videos || [],
        total: response.data.data.total || 0,
        page: response.data.data.page || 1,
        limit: response.data.data.limit || 20,
        pages: response.data.data.pages || 1
      };
    } else {
      console.error('获取下载视频失败:', response.data?.message || '未知错误');
    }
  } catch (error) {
    console.error('请求获取下载视频列表出错:', error);
  } finally {
    isLoading.value = false;
  }
};

// 处理页码变化
const handlePageChange = (page) => {
  currentPage.value = page;
  loadDownloadedVideos();
};

// 播放视频
const playVideo = (filePath) => {
  // 先关闭播放器并重置路径
  showVideoPlayer.value = false;
  currentVideoPath.value = '';

  // 在短暂延迟后重新设置路径并打开播放器
  setTimeout(() => {
    currentVideoPath.value = filePath;
    showVideoPlayer.value = true;
  }, 50);
};

// 处理删除视频
const handleDeleteVideo = (video) => {
  currentVideo.value = video;
  showDeleteConfirm.value = true;
  deleteDirectory.value = true;
};

// 获取视频目录
const getVideoDirectory = (video) => {
  // 如果视频对象已经有目录属性，直接使用
  if (video.directory) return video.directory;

  // 否则，尝试从第一个文件路径中提取目录
  if (video.files && video.files.length > 0) {
    const filePath = video.files[0].file_path;
    if (filePath) {
      // 提取目录路径 (去掉文件名)
      const lastSlashIndex = Math.max(filePath.lastIndexOf('/'), filePath.lastIndexOf('\\'));
      if (lastSlashIndex !== -1) {
        return filePath.substring(0, lastSlashIndex);
      }
    }
  }
  return null;
};

// 确认删除视频
const confirmDeleteVideo = async () => {
  try {
    isDeleting.value = true;

    // 确定目录路径
    const directory = getVideoDirectory(currentVideo.value);

    // 自动判断是否使用CID
    // 如果有CID则使用CID，否则使用目录路径
    const cid = currentVideo.value?.cid || null;

    // 如果既没有CID也没有目录路径，则无法删除
    if (!cid && !directory) {
      showNotify({
        type: 'warning',
        message: '无法获取视频信息，删除失败'
      });
      isDeleting.value = false;
      return;
    }

    const response = await deleteDownloadedVideo(cid, deleteDirectory.value, directory);

    if (response.data && response.data.status === 'success') {
      showNotify({
        type: 'success',
        message: response.data.message
      });

      // 关闭对话框
      showDeleteConfirm.value = false;

      // 重新加载列表
      await loadDownloadedVideos();
    } else {
      throw new Error(response.data?.message || '删除视频失败');
    }
  } catch (error) {
    showNotify({
      type: 'danger',
      message: error.response?.data?.message || error.message || '删除视频失败'
    });
  } finally {
    isDeleting.value = false;
  }
};

// 点击视频跳转到B站
const handleVideoClick = async (video) => {
  // 构建视频链接，使用bvid而不是cid
  const url = `https://www.bilibili.com/video/${video.bvid}`;
  // 在系统默认浏览器中打开
  await openInBrowser(url);
};

// 点击作者头像或名称跳转到作者页面
const handleAuthorClick = async (video) => {
  if (video.author_mid) {
    const url = `https://space.bilibili.com/${video.author_mid}`;
    await openInBrowser(url);
  }
};

// 检查是否为浏览器不支持的格式（MKV 容器、AV1 编码）
const isUnsupportedFormat = (video) => {
  if (!video.files || !video.files.length) return false;
  const fileName = video.files[0].file_path || '';
  const lower = fileName.toLowerCase();
  // MKV 容器在大多数浏览器中无法直接播放
  return lower.endsWith('.mkv');
};

// SSR: 初始数据在服务端获取
const { data: initialData } = await useAsyncData('downloads-initial', async () => {
  try {
    const response = await getDownloadedVideos('', 1, 20);
    if (response.data?.status === 'success' && response.data?.data) {
      return { downloads: response.data.data };
    }
    return { downloads: { videos: [], total: 0, page: 1, limit: 20, pages: 1 } };
  } catch (error) {
    console.error('SSR 获取下载列表失败:', error);
    return { downloads: { videos: [], total: 0, page: 1, limit: 20, pages: 1 } };
  }
});

// 从 SSR 数据初始化组件状态
if (initialData.value?.downloads) {
  downloads.value = initialData.value.downloads;
  isLoading.value = false;
}

// 组件挂载时加载数据
onMounted(() => {
  if (!downloads.value?.videos || downloads.value.videos.length === 0) {
    loadDownloadedVideos();
  }
});
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.group:hover .group-hover\:block {
  display: block;
}
</style>
