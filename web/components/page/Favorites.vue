<!-- 收藏夹页面 -->
<template>
  <div class="relative min-h-screen overflow-hidden pb-20 md:pb-0">
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute -top-32 -right-20 h-96 w-96 rounded-full bg-accent/5 blur-3xl" />
      <div class="absolute top-1/4 -left-32 h-80 w-80 rounded-full bg-accent/3 blur-3xl" />
    </div>

    <div class="relative py-6">
      <div class="mx-auto max-w-[1800px] sm:px-2 lg:px-8">
        <!-- 页面标题区 -->
        <div v-if="!showFolderContents" class="mb-6">
          <div class="glass-card overflow-hidden">
            <div class="px-5 py-4 md:px-6 md:py-5">
              <div class="flex flex-wrap items-center justify-between gap-4">
                <div class="flex items-center gap-3 md:gap-4">
                  <div
                    class="flex h-11 w-11 items-center justify-center rounded-xl bg-accent shadow-md shadow-accent/20 md:h-12 md:w-12"
                  >
                    <svg
                      class="h-5 w-5 text-white md:h-6 md:w-6"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
                      />
                    </svg>
                  </div>
                  <div>
                    <h1 class="text-xl font-bold text-gray-900 md:text-2xl dark:text-white">
                      我的收藏
                    </h1>
                    <p class="mt-0.5 text-xs text-gray-500 md:text-sm dark:text-gray-400">
                      共
                      <span class="font-medium text-gray-700 dark:text-gray-300">{{
                        totalItems
                      }}</span>
                      个收藏夹
                    </p>
                  </div>
                </div>
                <nav
                  class="inline-flex rounded-xl bg-gray-100 p-1 dark:bg-gray-800"
                  aria-label="收藏夹选项卡"
                >
                  <button
                    class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium transition-all duration-200 md:px-4 md:py-2 md:text-sm"
                    :class="
                      activeTab === 'created'
                        ? 'bg-white text-accent shadow-sm dark:bg-gray-700'
                        : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                    "
                    @click="activeTab = 'created'"
                  >
                    <svg
                      class="h-3.5 w-3.5 md:h-4 md:w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="1.8"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
                      />
                    </svg>
                    <span class="hidden sm:inline">我创建的</span>
                    <span class="sm:hidden">创建</span>
                  </button>
                  <button
                    class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium transition-all duration-200 md:px-4 md:py-2 md:text-sm"
                    :class="
                      activeTab === 'collected'
                        ? 'bg-white text-accent shadow-sm dark:bg-gray-700'
                        : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                    "
                    @click="activeTab = 'collected'"
                  >
                    <svg
                      class="h-3.5 w-3.5 md:h-4 md:w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="1.8"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
                      />
                    </svg>
                    <span class="hidden sm:inline">我收藏的</span>
                    <span class="sm:hidden">收藏</span>
                  </button>
                </nav>
              </div>
            </div>
          </div>
        </div>

        <!-- 主内容卡片 -->
        <div class="glass-card overflow-hidden">
          <!-- 文件夹内容标题栏 -->
          <div v-if="showFolderContents" class="border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between px-4 py-3">
              <div class="flex items-center space-x-4">
                <button
                  class="rounded-full p-1 transition-colors hover:bg-gray-100 dark:hover:bg-gray-700"
                  @click="backToFolderList"
                >
                  <svg
                    class="h-5 w-5 text-gray-600 dark:text-gray-300"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M10 19l-7-7m0 0l7-7m-7 7h18"
                    />
                  </svg>
                </button>
                <h2 class="truncate text-lg font-medium">
                  {{ currentFolder?.title || '收藏夹内容' }}
                </h2>
              </div>
              <div class="flex items-center space-x-2">
                <button
                  class="flex items-center rounded-md bg-gray-100 px-3 py-1.5 text-xs text-gray-700 transition-colors hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
                  :disabled="fetchingAll"
                  @click="fetchAllContents"
                >
                  <svg
                    class="mr-1 h-3.5 w-3.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                    />
                  </svg>
                  <span>同步到本地</span>
                </button>
              </div>
            </div>
          </div>

          <!-- 内容区域 -->
          <div class="p-5 transition-all duration-300">
            <!-- 全局提示信息 -->
            <div
              v-if="showFavoritesTip"
              class="mb-4 rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-700 dark:border-amber-700 dark:bg-amber-900/20 dark:text-amber-300"
            >
              <div class="flex items-start">
                <svg
                  class="mt-0.5 mr-2 h-5 w-5 flex-shrink-0 text-amber-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
                <div class="flex-1">
                  <p class="mt-1">
                    用户收藏夹往往非常庞大，解析时很容易触发反爬机制。如遇该问题请稍等片刻后重试。（emmm，如果视频太多的话还是建议逐个收藏夹下载……）
                  </p>
                </div>
                <button
                  class="ml-2 flex-shrink-0 text-amber-500 hover:text-amber-700"
                  @click="dismissFavoritesTip"
                >
                  <svg
                    class="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>

            <!-- 收藏夹列表 -->
            <div v-if="!showFolderContents" class="animate-fadeIn">
              <!-- 收藏夹列表显示区域 -->
              <div v-if="initialLoading || loading" class="flex justify-center py-20">
                <div class="glass-card flex items-center gap-3 px-5 py-4">
                  <div
                    class="h-5 w-5 animate-spin rounded-full border-2 border-accent border-t-transparent"
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300">加载中...</span>
                </div>
              </div>

              <div v-else-if="favorites.length === 0" class="py-20 text-center">
                <svg
                  class="mx-auto h-16 w-16 text-gray-300"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
                  />
                </svg>
                <p class="mt-4 text-gray-500 dark:text-gray-400">暂无收藏夹</p>
                <!-- 在线收藏夹（需要登录） -->
                <template
                  v-if="(activeTab === 'created' || activeTab === 'collected') && !isLoggedIn"
                >
                  <p class="mt-2 text-sm text-gray-400">您需要登录B站账号才能查看收藏夹</p>
                  <button
                    class="btn-press mt-4 rounded-xl bg-accent px-5 py-2.5 text-sm font-medium text-white shadow-lg shadow-accent/20 transition-all hover:bg-accent/90"
                    @click="openLoginDialog"
                  >
                    登录账号
                  </button>
                </template>
                <!-- 已登录但没有收藏夹 -->
                <template
                  v-else-if="(activeTab === 'created' || activeTab === 'collected') && isLoggedIn"
                >
                  <p class="mt-2 text-sm text-gray-400">
                    {{
                      activeTab === 'created' ? '您还没有创建过收藏夹' : '您还没有收藏任何收藏夹'
                    }}
                  </p>
                </template>
              </div>

              <div
                v-else
                class="stagger-children grid gap-3 px-4"
                style="grid-template-columns: repeat(auto-fill, minmax(240px, 1fr))"
              >
                <!-- 收藏夹卡片 -->
                <div
                  v-for="folder in favorites"
                  :key="folder.id || folder.media_id"
                  class="glass-card-hover flex flex-col overflow-hidden"
                >
                  <!-- 封面图 -->
                  <div class="relative aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700">
                    <img
                      :src="normalizeImageUrl(folder.cover)"
                      :alt="folder.title"
                      class="h-full w-full object-cover transition-transform duration-300 hover:scale-105"
                      loading="lazy"
                      @click="viewFolderContents(folder)"
                    />
                    <button
                      class="absolute top-2 right-2 flex h-7 w-7 items-center justify-center rounded-lg bg-black/40 text-white transition-colors hover:bg-black/60"
                      title="下载收藏夹中的视频"
                      @click.stop="startDownloadFolder(folder)"
                    >
                      <svg
                        class="h-3.5 w-3.5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                        />
                      </svg>
                    </button>
                    <div
                      class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent"
                    />
                    <div class="absolute right-0 bottom-0 left-0 p-3">
                      <p class="truncate text-sm font-semibold text-white drop-shadow-sm">
                        {{ folder.title }}
                      </p>
                      <div class="mt-1 flex items-center">
                        <span class="text-xs font-medium text-white/85"
                          >{{ folder.media_count }}个内容</span
                        >
                      </div>
                    </div>
                  </div>

                  <!-- 收藏夹信息 -->
                  <div class="flex flex-1 flex-col p-3">
                    <div class="flex items-start justify-between">
                      <div class="flex-1">
                        <h3
                          class="cursor-pointer font-medium text-gray-900 transition-colors hover:text-accent dark:text-gray-100"
                          @click="viewFolderContents(folder)"
                        >
                          {{ folder.title }}
                        </h3>
                        <p
                          v-if="folder.intro"
                          class="mt-1 line-clamp-2 text-xs text-gray-500 dark:text-gray-400"
                        >
                          {{ folder.intro }}
                        </p>
                        <p
                          v-if="activeTab === 'collected' && folder.upper?.name"
                          class="mt-1 text-xs text-gray-400 dark:text-gray-500"
                        >
                          UP主: {{ folder.upper.name }}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 分页控件 -->
              <div v-if="favorites.length > 0 && totalPages > 1" class="mt-6 flex justify-center">
                <Pagination
                  :current-page="currentPage"
                  :total-pages="totalPages"
                  @page-change="handlePageChange"
                />
              </div>
            </div>

            <!-- 收藏夹内容 -->
            <div v-if="showFolderContents" class="animate-fadeIn">
              <div v-if="loadingContents" class="flex justify-center py-20">
                <div
                  class="inline-flex items-center rounded-md bg-white px-4 py-2 text-gray-900 shadow dark:bg-gray-800 dark:text-gray-100"
                >
                  <svg
                    class="mr-3 -ml-1 h-5 w-5 animate-spin text-accent"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
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
                  <span>加载中...</span>
                </div>
              </div>

              <div v-else-if="folderContents.length === 0" class="py-10 text-center">
                <p class="text-gray-500 dark:text-gray-400">该收藏夹暂无内容</p>
              </div>

              <div v-else>
                <!-- 收藏夹操作栏 -->
                <div
                  class="mb-4 flex flex-wrap items-center justify-between rounded-lg bg-white/70 p-3 shadow-sm dark:bg-gray-800/70"
                >
                  <div class="flex items-center space-x-3">
                    <div class="text-sm text-gray-700 dark:text-gray-300">
                      共 {{ contentsTotalItems }} 个内容
                    </div>
                    <div v-if="invalidVideosCount > 0" class="text-sm text-red-500">
                      ({{ invalidVideosCount }} 个失效)
                    </div>
                  </div>

                  <div class="mt-2 flex items-center space-x-3 sm:mt-0">
                    <button
                      class="flex items-center rounded-md bg-blue-50 px-3 py-1.5 text-xs text-blue-600 transition-colors hover:bg-blue-100"
                      @click="startDownloadFolder(currentFolder)"
                    >
                      <svg
                        class="mr-1 h-3.5 w-3.5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                        />
                      </svg>
                      <span>下载收藏夹</span>
                    </button>
                  </div>
                </div>

                <!-- 内容列表 - 网格布局 -->
                <div
                  class="stagger-children grid gap-3 px-4"
                  style="grid-template-columns: repeat(auto-fill, minmax(220px, 1fr))"
                >
                  <div
                    v-for="item in folderContents"
                    :key="item.id || item.bvid"
                    class="glass-card-hover group relative overflow-hidden"
                  >
                    <!-- 视频封面 -->
                    <div
                      class="group relative cursor-pointer overflow-hidden pb-[56.25%]"
                      @click="openVideo(item)"
                    >
                      <img
                        :src="normalizeImageUrl(getVideoImage(item))"
                        :alt="getVideoTitle(item)"
                        class="absolute inset-0 h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                        loading="lazy"
                        onerror="
                          this.src =
                            '/api/images/proxy?url=https%3A%2F%2Fi0.hdslb.com%2Fbfs%2Farchive%2Fc9e72655b7c9c9c68a30d3275313c501e68427d1.jpg'
                        "
                      />

                      <!-- 视频时长标签 -->
                      <div
                        class="absolute right-1 bottom-1 rounded bg-black/60 px-1 py-0.5 text-[0.625rem] text-white"
                      >
                        {{ formatDuration(item.duration) }}
                      </div>

                      <!-- 分区标签 -->
                      <div
                        v-if="item.tname"
                        class="absolute top-1 left-1 rounded bg-accent/80 px-1 py-0.5 text-[0.625rem] text-white"
                      >
                        {{ item.tname }}
                      </div>
                    </div>

                    <!-- 视频信息 -->
                    <div class="flex flex-col space-y-1 p-2">
                      <!-- 标题 -->
                      <div
                        class="line-clamp-2 cursor-pointer text-xs font-medium text-gray-900 dark:text-gray-100"
                        @click="openVideo(item)"
                      >
                        {{ getVideoTitle(item) }}
                      </div>

                      <!-- 作者信息 -->
                      <div class="flex items-center space-x-1">
                        <img
                          :src="normalizeImageUrl(getAuthorFace(item))"
                          :alt="getAuthorName(item)"
                          class="h-3.5 w-3.5 cursor-pointer rounded-full object-cover"
                          loading="lazy"
                          onerror="this.style.display = 'none'"
                          @click.stop="openAuthorPage(item)"
                        />
                        <span
                          class="cursor-pointer truncate text-[0.625rem] text-gray-600 hover:text-accent dark:text-gray-400"
                          @click.stop="openAuthorPage(item)"
                        >
                          {{ getAuthorName(item) }}
                        </span>
                      </div>

                      <!-- 收藏时间（合集无收藏时间，不显示） -->
                      <div
                        v-if="item.fav_time"
                        class="flex items-center justify-between text-[0.625rem] text-gray-500"
                      >
                        <div class="flex items-center space-x-1">
                          <svg
                            class="h-2.5 w-2.5"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                          >
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              stroke-width="2"
                              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                            />
                          </svg>
                          <span>收藏于: {{ formatTime(item.fav_time) }}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 内容分页（合集不分页，B站API返回全部数据） -->
                <div
                  v-if="contentsTotalPages > 1 && !isSeasonFolder"
                  class="mt-6 flex justify-center"
                >
                  <Pagination
                    :current-page="contentsPage"
                    :total-pages="contentsTotalPages"
                    @page-change="handleContentsPageChange"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 登录弹窗 -->
    <LoginDialog v-model:show="showLoginDialog" @login-success="onLoginSuccess" />

    <!-- 全屏加载遮罩 -->
    <div v-if="fetchingAll" class="glass-overlay flex items-center justify-center">
      <div class="glass-modal w-full max-w-xs animate-scale-in p-6 text-center">
        <div
          class="mx-auto mb-4 h-10 w-10 animate-spin rounded-full border-3 border-accent border-t-transparent"
        />
        <h3 class="mb-2 text-lg font-medium text-gray-900 dark:text-gray-100">
          正在获取全部收藏内容
        </h3>
        <p class="mb-3 text-sm text-gray-500 dark:text-gray-400">请耐心等待，这可能需要一些时间</p>
        <div class="mb-3 h-2 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
          <div
            class="h-2 rounded-full bg-accent transition-all duration-300"
            :style="{ width: `${fetchProgress}%` }"
          />
        </div>
        <p class="text-sm text-gray-600 dark:text-gray-300">
          已获取 {{ currentFetchPage }} / {{ totalFetchPages }} 页
        </p>
      </div>
    </div>

    <!-- 使用DownloadDialog组件 -->
    <DownloadDialog
      v-model:show="showDownloadDialog"
      :video-info="favoriteDownloadInfo"
      @download-complete="handleDownloadComplete"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch, onUnmounted } from 'vue';
import { useAsyncData } from '#imports';
import { showNotify } from 'vant';
import 'vant/es/notify/style';
import 'vant/es/dialog/style';
import Pagination from '../Pagination.vue';
import LoginDialog from '../LoginDialog.vue';
import DownloadDialog from '../DownloadDialog.vue';
import {
  getCreatedFavoriteFolders,
  getCollectedFavoriteFolders,
  getLocalCollectedFolders,
  getFavoriteContents,
  getLocalFavoriteContents,
  getOnlineFavoriteContents,
  getLoginStatus
} from '~/utils/api';
import { openInBrowser } from '~/utils/openUrl.js';
import { normalizeImageUrl } from '~/utils/imageUrl.js';
import { formatDuration } from '~/utils/format';

// 状态变量
const loading = ref(false);
const favorites = ref([]);
const activeTab = ref('created');
const currentPage = ref(1);
const pageSize = ref(40);
const totalItems = ref(0);
const searchKeyword = ref('');
const showFavoritesTip = ref(localStorage.getItem('favorites_tip_dismissed') !== 'true');
const initialLoading = ref(true);

function dismissFavoritesTip() {
  showFavoritesTip.value = false;
  localStorage.setItem('favorites_tip_dismissed', 'true');
}

// 收藏夹内容状态
const showFolderContents = ref(false);
const currentFolder = ref(null);
const folderContents = ref([]);
const loadingContents = ref(false);
const contentsPage = ref(1);
const contentsPageSize = ref(40);
const contentsTotalItems = ref(0);

// 登录弹窗状态
const showLoginDialog = ref(false);

// 已移除修复功能相关弹窗

// 登录状态
const isLoggedIn = ref(false);
const checkingLoginStatus = ref(false);

// 获取全部收藏夹内容相关状态
const fetchingAll = ref(false);
const currentFetchPage = ref(0);
const totalFetchPages = ref(0);
const fetchProgress = ref(0);
const allFetchedContents = ref([]);

// 已移除修复状态与结果

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(totalItems.value / pageSize.value);
});

// 计算内容总页数
const contentsTotalPages = computed(() => {
  return Math.ceil(contentsTotalItems.value / contentsPageSize.value);
});

// 判断当前是否为合集（B站合集API不分页，返回全部数据）
const isSeasonFolder = computed(() => {
  return currentFolder.value?.type === 21;
});

// 监听活动标签变化
watch(activeTab, () => {
  currentPage.value = 1;
  fetchFavorites();
});

// 非阻塞加载初始数据（仅读取本地数据，不调用远程B站API避免超时）
const { data: initialData } = useAsyncData('favorites-initial', async () => {
  try {
    const loginResponse = await getLoginStatus();
    const loggedIn =
      loginResponse.data && loginResponse.data.code === 0 && loginResponse.data.data.isLogin;

    if (!loggedIn) {
      return { isLoggedIn: false, favorites: [], totalItems: 0 };
    }

    const response = await getLocalFavoriteFolders({ page: 1, size: 50 });
    const favorites = response.data.status === 'success' ? response.data.data.list || [] : [];
    const totalItems = response.data.status === 'success' ? response.data.data.count || 0 : 0;

    return { isLoggedIn: true, favorites, totalItems };
  } catch (error) {
    console.error('获取初始收藏夹数据失败:', error);
    return { isLoggedIn: false, favorites: [], totalItems: 0 };
  }
});

// initialData 异步解析完成后初始化组件状态
watch(initialData, (val) => {
  if (val) {
    isLoggedIn.value = val.isLoggedIn;
    favorites.value = val.favorites;
    totalItems.value = val.totalItems;
    initialLoading.value = false;
  }
});

// 组件挂载时加载数据
onMounted(() => {
  if (initialData.value) {
    initialLoading.value = false;
    if (!initialData.value.isLoggedIn) {
      checkLoginStatus();
    }
    if (initialData.value.isLoggedIn && initialData.value.favorites.length === 0) {
      fetchFavorites();
    }
  } else {
    checkLoginStatus();
  }

  setTimeout(() => {
    initialLoading.value = false;
  }, 8000);

  // 添加全局登录状态变化的监听
  window.addEventListener('login-status-changed', handleLoginStatusChange);
});

// 组件卸载时移除事件监听
onUnmounted(() => {
  window.removeEventListener('login-status-changed', handleLoginStatusChange);
});

// 处理登录状态变化事件
function handleLoginStatusChange(event) {
  console.log('收藏页面收到登录状态变化事件:', event.detail);
  if (event.detail && typeof event.detail.isLoggedIn !== 'undefined') {
    isLoggedIn.value = event.detail.isLoggedIn;
    if (isLoggedIn.value) {
      // 如果登录状态变为已登录，重新获取收藏夹
      fetchFavorites();
    }
  } else {
    // 如果事件没有包含登录状态信息，则重新检查
    checkLoginStatus();
  }
}

// 检查登录状态
async function checkLoginStatus() {
  checkingLoginStatus.value = true;
  try {
    const response = await getLoginStatus();
    console.log('获取登录状态响应:', response.data);
    if (response.data && response.data.code === 0) {
      isLoggedIn.value = response.data.data.isLogin;
      console.log('登录状态:', isLoggedIn.value);
    } else {
      console.warn('登录状态响应异常:', response.data);
      isLoggedIn.value = false;
    }
  } catch (error) {
    console.error('获取登录状态失败:', error);
    isLoggedIn.value = false;
  } finally {
    checkingLoginStatus.value = false;
  }
}

// 获取收藏夹列表
async function fetchFavorites() {
  loading.value = true;
  favorites.value = [];

  try {
    let response;

    if (activeTab.value === 'created') {
      response = await getLocalFavoriteFolders({
        page: currentPage.value,
        size: pageSize.value
      });
      if (
        !response ||
        response.data.status !== 'success' ||
        !response.data.data.list ||
        response.data.data.list.length === 0
      ) {
        if (response && response.data.status !== 'success') {
          console.warn('本地收藏夹读取失败，尝试从B站在线获取:', response.data.message);
        }
        response = await getCreatedFavoriteFolders({
          keyword: searchKeyword.value || undefined
        });
      }
    } else if (activeTab.value === 'collected') {
      response = await getLocalCollectedFolders({
        page: currentPage.value,
        size: pageSize.value
      });
      if (
        !response ||
        response.data.status !== 'success' ||
        !response.data.data.list ||
        response.data.data.list.length === 0
      ) {
        if (response && response.data.status !== 'success') {
          console.warn('本地收藏夹读取失败，尝试从B站在线获取:', response.data.message);
        }
        response = await getCollectedFavoriteFolders({
          pn: currentPage.value,
          ps: pageSize.value,
          keyword: searchKeyword.value || undefined
        });
      }
    }

    if (response.data.status === 'success') {
      favorites.value = response.data.data.list || [];
      totalItems.value = response.data.data.count || 0;

      // 如果收藏夹没有封面，使用第一个视频的封面
      for (const folder of favorites.value) {
        if (!folder.cover || folder.cover.includes('nocover')) {
          // 预加载第一个视频的封面
          preloadFirstVideoCover(folder);
        }
      }
    } else {
      showNotify({ type: 'danger', message: response.data.message || '获取收藏夹失败' });
    }
  } catch (error) {
    console.error('获取收藏夹出错:', error);
    showNotify({ type: 'danger', message: '获取收藏夹出错: ' + (error.message || '未知错误') });
  } finally {
    loading.value = false;
  }
}

// 预加载收藏夹的第一个视频封面
async function preloadFirstVideoCover(folder) {
  try {
    const folderId = folder.id || folder.media_id;
    if (!folderId) return;

    const response = await getLocalFavoriteContents({
      media_id: folderId,
      page: 1,
      size: 20
    });

    if (response.data.status === 'success') {
      const contents = response.data.data?.list || [];
      for (const item of contents) {
        if (item.cover) {
          folder.cover = item.cover;
          return;
        }
      }
    }
  } catch (error) {
    console.error('获取封面出错:', error);
  }
}

// 查看收藏夹内容
async function viewFolderContents(folder) {
  currentFolder.value = folder;
  showFolderContents.value = true;
  contentsPage.value = 1;
  folderContents.value = [];

  await loadContents();
}

// 返回到收藏夹列表
function backToFolderList() {
  showFolderContents.value = false;
  currentFolder.value = null;
  folderContents.value = [];
}

// 加载收藏夹内容
async function loadContents() {
  if (!currentFolder.value) return;

  loadingContents.value = true;
  folderContents.value = [];

  try {
    let response;
    const folderId = currentFolder.value.media_id || currentFolder.value.id;
    const isSeason = currentFolder.value.type === 21;

    // 优先读本地
    response = await getLocalFavoriteContents({
      media_id: folderId,
      page: contentsPage.value,
      size: contentsPageSize.value
    });

    // 本地没有数据时，从B站在线获取
    if (
      !response ||
      response.data.status !== 'success' ||
      !response.data.data.list ||
      response.data.data.list.length === 0
    ) {
      const params = {
        pn: contentsPage.value,
        ps: contentsPageSize.value
      };
      if (isSeason) {
        params.season_id = folderId;
      } else {
        params.media_id = folderId;
      }
      response = await getOnlineFavoriteContents(params);
    }

    if (response.data.status === 'success') {
      // 更新收藏夹信息
      if (response.data.data && response.data.data.info) {
        const info = response.data.data.info;
        // 更新当前展示的收藏夹信息
        currentFolder.value.title = info.title || currentFolder.value.title;
        currentFolder.value.cover = info.cover || currentFolder.value.cover;
        currentFolder.value.intro = info.intro || currentFolder.value.intro;
        currentFolder.value.media_count = info.media_count || currentFolder.value.media_count;

        // 更新UP主信息
        if (info.upper) {
          currentFolder.value.upper = info.upper;
        }
      }

      // 确保我们能够正确处理不同的数据结构
      if (response.data.data && response.data.data.list) {
        folderContents.value = response.data.data.list;
        contentsTotalItems.value = response.data.data.total || currentFolder.value.media_count || 0;
      } else if (response.data.data && response.data.data.medias) {
        folderContents.value = response.data.data.medias;
        contentsTotalItems.value = currentFolder.value.media_count || 0;
      } else if (response.data.data && Array.isArray(response.data.data)) {
        folderContents.value = response.data.data;
        contentsTotalItems.value = response.data.total || currentFolder.value.media_count || 0;
      } else {
        console.warn('收藏夹内容数据结构异常:', response.data);
        folderContents.value = [];
        showNotify({ type: 'warning', message: '收藏夹数据结构异常，无法显示内容' });
      }

      // 确保至少更新了folderContents
      if (folderContents.value.length === 0) {
        console.warn('无法从响应中提取内容数据');
        showNotify({ type: 'warning', message: '无法从响应中提取内容数据' });
      }
    } else {
      console.error('收藏夹请求失败:', response.data);
      showNotify({ type: 'danger', message: response.data.message || '获取收藏夹内容失败' });
    }
  } catch (error) {
    console.error('获取收藏夹内容出错:', error);
    showNotify({ type: 'danger', message: '获取收藏夹内容出错: ' + (error.message || '未知错误') });
  } finally {
    loadingContents.value = false;
  }

  // 返回加载的内容，便于调用者使用
  return folderContents.value;
}

// 打开视频
async function openVideo(video) {
  // 使用BV号或视频ID打开视频，跳转到B站
  const videoId = video.bvid || video.id;
  if (videoId) {
    // 在系统默认浏览器中打开B站视频链接
    await openInBrowser(`https://www.bilibili.com/video/${videoId}`);
  }
}

// 处理搜索
// 处理分页变化
function handlePageChange(page) {
  currentPage.value = page;
  fetchFavorites();
}

// 处理内容分页变化
async function handleContentsPageChange(page) {
  console.log(`切换到第${page}页内容`);
  contentsPage.value = page;

  try {
    // 等待内容加载完成
    await loadContents();
  } catch (error) {
    console.error('分页处理出错:', error);
    showNotify({
      type: 'danger',
      message: '分页处理出错: ' + (error.message || '未知错误')
    });
  }
}

// 打开登录对话框
function openLoginDialog() {
  showLoginDialog.value = true;
}

// 登录成功回调
function onLoginSuccess() {
  isLoggedIn.value = true;
  showNotify({ type: 'success', message: '登录成功，正在获取收藏夹数据' });
  fetchFavorites();
}

// 格式化时间戳为可读格式
function formatTime(timestamp) {
  if (!timestamp) return '未知';

  const date = new Date(timestamp * 1000);
  const pad = (n) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

// 获取作者头像
function getAuthorFace(item) {
  // 首先检查upper对象
  if (item.upper && item.upper.face) {
    return item.upper.face;
  }
  // 然后检查creator_face属性（本地数据结构）
  else if (item.creator_face) {
    return item.creator_face;
  }
  // 再检查upper_mid关联的信息（本地数据可能有的另一种形式）
  else if (item.upper_mid && typeof item.upper_mid === 'number') {
    return '';
  } else {
    return '';
  }
}

// 获取作者名称
function getAuthorName(item) {
  // 首先检查upper对象
  if (item.upper && item.upper.name) {
    return item.upper.name;
  }
  // 然后检查creator_name属性（本地数据结构）
  else if (item.creator_name) {
    return item.creator_name;
  }
  // 最后返回默认名称
  else {
    return '未知UP主';
  }
}

// 获取视频是否有修复信息
// 已移除修复信息获取逻辑

// 获取全部收藏夹内容的函数，每页请求间隔1秒
async function fetchAllContents() {
  if (!currentFolder.value || fetchingAll.value) return;

  fetchingAll.value = true;
  allFetchedContents.value = [];

  try {
    // 获取收藏夹信息以确定总页数
    const folderId = currentFolder.value.media_id || currentFolder.value.id;
    let firstPageResponse;
    let fetchApi;
    let processResponse;

    console.log('开始获取全部收藏内容，收藏夹ID:', folderId, '每页大小:', contentsPageSize.value);

    fetchApi = (page) => {
      console.log(`请求收藏夹第${page}页, 参数:`, {
        media_id: folderId,
        page: page,
        size: contentsPageSize.value
      });
      return getFavoriteContents({
        media_id: folderId,
        page: page,
        size: contentsPageSize.value
      });
    };

    processResponse = (response, page) => {
      console.log(`处理收藏夹第${page}页响应:`, response.data);
      if (response.data.status === 'success') {
        // 更新收藏夹信息
        if (response.data.data && response.data.data.info && page === 1) {
          const info = response.data.data.info;
          currentFolder.value.title = info.title || currentFolder.value.title;
          currentFolder.value.cover = info.cover || currentFolder.value.cover;
          currentFolder.value.intro = info.intro || currentFolder.value.intro;
          currentFolder.value.media_count = info.media_count || currentFolder.value.media_count;

          if (info.upper) {
            currentFolder.value.upper = info.upper;
          }
        }

        // 处理多种可能的数据结构
        if (response.data.data && response.data.data.list) {
          return {
            contents: response.data.data.list,
            total: response.data.data.total || currentFolder.value.media_count || 0
          };
        } else if (response.data.data && response.data.data.medias) {
          return {
            contents: response.data.data.medias,
            total: currentFolder.value.media_count || 0
          };
        } else if (response.data.data && Array.isArray(response.data.data)) {
          console.log(`第${page}页: 找到数组数据，数量:`, response.data.data.length);
          return {
            contents: response.data.data,
            total: response.data.total || currentFolder.value.media_count || 0
          };
        } else {
          console.warn(`第${page}页: 数据结构异常:`, response.data);
          return { contents: [], total: currentFolder.value.media_count || 0 };
        }
      }
      console.error(`第${page}页: 请求失败:`, response.data);
      return { contents: [], total: 0 };
    };

    // 第一页请求，获取总数量信息
    console.log('获取第1页数据...');
    firstPageResponse = await fetchApi(1);
    const result = processResponse(firstPageResponse, 1);
    const total = result.total;

    console.log('首页数据获取完成，总数据条目:', total);

    // 计算总页数
    totalFetchPages.value = Math.ceil(total / contentsPageSize.value);
    console.log('计算出总页数:', totalFetchPages.value);

    currentFetchPage.value = 1;
    fetchProgress.value = (1 / totalFetchPages.value) * 100;

    // 添加第一页数据
    allFetchedContents.value = [...allFetchedContents.value, ...result.contents];

    // 如果只有一页，则完成
    if (totalFetchPages.value <= 1) {
      showNotify({ type: 'success', message: '收藏夹内容获取完成！' });
      fetchingAll.value = false;
      folderContents.value = allFetchedContents.value;
      return;
    }

    // 依次请求后续页面
    for (let page = 2; page <= totalFetchPages.value; page++) {
      console.log(`等待1秒后获取第${page}页数据...`);
      // 等待1秒
      await new Promise((resolve) => setTimeout(resolve, 1000));

      try {
        console.log(`开始获取第${page}页数据`);
        const response = await fetchApi(page);
        const pageResult = processResponse(response, page);

        // 添加本页数据
        allFetchedContents.value = [...allFetchedContents.value, ...pageResult.contents];

        // 更新进度
        currentFetchPage.value = page;
        fetchProgress.value = (page / totalFetchPages.value) * 100;
        console.log(`第${page}页数据获取完成，当前进度: ${fetchProgress.value.toFixed(2)}%`);
      } catch (error) {
        console.error(`获取第${page}页出错:`, error);
        showNotify({
          type: 'warning',
          message: `获取第${page}页出错，将继续获取下一页: ${error.message}`
        });
      }
    }

    // 完成后更新数据并通知
    folderContents.value = allFetchedContents.value;
    console.log('所有页面获取完成，总共获取到', allFetchedContents.value.length, '条内容');
    showNotify({
      type: 'success',
      message: `已获取全部${allFetchedContents.value.length}个收藏内容！`
    });
  } catch (error) {
    console.error('获取全部收藏夹内容出错:', error);
    showNotify({ type: 'danger', message: '获取收藏夹内容出错: ' + (error.message || '未知错误') });
  } finally {
    fetchingAll.value = false;
  }
}

// 获取视频封面，优先使用修复后的封面
function getVideoImage(video) {
  // 检查video对象是否存在
  if (!video) return '';

  console.log('获取视频封面:', video.bvid || video.avid);

  // 返回原始封面
  return video.cover || '';
}

// 获取视频标题，优先使用修复后的标题
function getVideoTitle(video) {
  // 检查video对象是否存在
  if (!video) return '未知标题';

  console.log('获取视频标题:', video.bvid || video.avid);

  // 返回原始标题
  return video.title || '未知标题';
}

// 判断视频是否正在修复
// 已移除修复中状态判断

// 打开UP主页面
function openAuthorPage(video) {
  let upId = null;

  // 已移除从修复结果提取UP主ID逻辑

  // 然后检查视频本身的数据
  if (!upId && video.upper_mid) {
    upId = video.upper_mid;
  } else if (!upId && video.upper && video.upper.mid) {
    upId = video.upper.mid;
  }

  // 如果找到UP主ID，跳转到B站UP主页面
  if (upId) {
    window.open(`https://space.bilibili.com/${upId}`, '_blank');
  } else {
    showNotify({ type: 'warning', message: '无法获取UP主信息' });
  }
}

// 下载相关状态
const showDownloadDialog = ref(false);
const favoriteDownloadInfo = ref({
  title: '',
  author: '',
  bvid: '',
  cover: '',
  cid: 0
});

// 计算无效视频数量
const invalidVideosCount = computed(() => 0);

// 开始下载收藏夹
async function startDownloadFolder(folder) {
  if (!folder) return;

  // 检查登录状态
  if (!isLoggedIn.value) {
    showNotify({ type: 'warning', message: '请先登录B站账号' });
    showLoginDialog.value = true;
    return;
  }

  // 获取完整的收藏夹视频总数
  try {
    // 发起一次API请求获取视频总数，仅获取第一页第一条
    const response = await getFavoriteContents({
      media_id: folder.id || folder.media_id,
      pn: 1,
      ps: 1
    });

    if (response.data && response.data.status === 'success' && response.data.data) {
      // 更新收藏夹信息
      if (response.data.data.info) {
        folder.media_count =
          response.data.data.info.media_count || response.data.data.total || folder.media_count;
      } else if (response.data.data.total) {
        folder.media_count = response.data.data.total;
      }

      console.log(`获取到收藏夹[${folder.title}]视频总数: ${folder.media_count}`);
    }
  } catch (error) {
    console.error('获取收藏夹信息失败:', error);
  }

  // 设置要下载的收藏夹信息
  favoriteDownloadInfo.value = {
    title: `收藏夹: ${folder.title || '未命名收藏夹'}`,
    author: folder.upper?.name || folder.creator_name || '未知创建者',
    bvid: `fid_${(folder.id || folder.media_id || '').toString()}`, // 使用特殊格式标识这是收藏夹ID
    cover: folder.cover || '',
    cid: 0,
    // 添加额外信息供下载组件使用
    is_favorite_folder: true,
    user_id: (folder.mid || folder.creator_mid || '').toString(),
    fid: (folder.id || folder.media_id || '').toString(),
    // 添加视频总数信息，帮助下载对话框显示正确的总数
    total_videos: folder.media_count || 0
  };

  // 打开下载对话框
  showDownloadDialog.value = true;
}

// 处理下载完成
function handleDownloadComplete() {
  showNotify({ type: 'success', message: '收藏夹下载完成' });
}
</script>

<style scoped>
.animate-fadeIn {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
