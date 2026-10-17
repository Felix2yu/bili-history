<template>
  <div
    ref="cardRef"
    class="group relative mx-auto max-w-2xl cursor-pointer transition-all duration-200 ease-in-out lg:max-w-4xl"
    :class="{
      'glass-card-hover p-3': !isBatchMode || !isSelected,
      'bg-accent/5 ring-2 ring-accent': isBatchMode && isSelected,
      'opacity-50': record.status === 1
    }"
    @click="handleClick"
  >
    <!-- Deleted badge -->
    <div
      v-if="record.status === 1"
      class="absolute top-2 left-2 z-20 rounded bg-red-500/80 px-1.5 py-0.5 text-[0.625rem] text-white"
    >
      已删除
    </div>
    <!-- Article type: full-width cover -->
    <div v-if="record.business === 'article-list' || record.business === 'article'">
      <div class="mb-2">
        <!-- eslint-disable vue/no-v-html -- 标题经 utils/escapeHtml 转义后高亮 -->
        <div
          class="line-clamp-2 text-sm font-medium text-gray-900 dark:text-gray-100"
          v-html="highlightedTitle"
        />
        <!-- eslint-enable vue/no-v-html -->
      </div>
      <div class="relative h-28 w-full overflow-hidden rounded-xl">
        <!-- Action buttons -->
        <div
          v-if="!isBatchMode"
          class="absolute top-2 right-2 z-20 hidden items-center gap-1.5 sm:group-hover:flex"
        >
          <div
            v-if="record.business === 'archive'"
            class="glass-icon-btn !h-7 !w-7"
            title="下载视频"
            @click.stop.prevent="handleDownload"
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
          </div>
          <div
            v-if="record.business === 'archive'"
            class="glass-icon-btn !h-7 !w-7 hover:!bg-pink-500/20 hover:!text-pink-500"
            :class="{ '!bg-pink-500/20 !text-pink-500': isLiked }"
            title="点赞"
            @click.stop="handleLike"
          >
            <svg
              class="h-3.5 w-3.5"
              :fill="isLiked ? 'currentColor' : 'none'"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5"
              />
            </svg>
          </div>
          <div
            class="glass-icon-btn !h-7 !w-7 hover:!bg-red-500/20 hover:!text-red-500"
            title="删除"
            @click.stop="handleDelete"
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
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </div>
        </div>
        <!-- Batch checkbox -->
        <div
          v-if="isBatchMode"
          class="absolute top-2 left-2 z-10"
          @click.stop="$emit('toggle-selection', record)"
        >
          <div
            class="flex h-5 w-5 items-center justify-center rounded-lg border-2 transition-all"
            :class="isSelected ? 'border-accent bg-accent' : 'border-white/80 bg-black/20'"
          >
            <svg
              v-if="isSelected"
              class="h-3 w-3 text-white"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="3"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </div>
        <!-- Downloaded badge -->
        <div
          v-if="isDownloaded && record.business === 'archive'"
          class="absolute top-0 left-0 z-10"
        >
          <div
            class="glass-badge rounded-tl-none rounded-br-lg !border-0 !bg-green-500/90 text-[0.625rem] !text-white"
          >
            <svg
              class="h-3 w-3"
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
            <span>已下载</span>
          </div>
        </div>
        <!-- Favorited badge -->
        <div
          v-if="isVideoFavorited && record.business !== 'live'"
          class="absolute top-0 right-0 z-10"
        >
          <div
            class="glass-badge rounded-tl-lg rounded-br-none !border-0 !bg-amber-500/90 text-[0.625rem] !text-white"
          >
            <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 24 24">
              <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
            </svg>
            <span>已收藏</span>
          </div>
        </div>
        <div
          v-if="record.tag_name || record.tname"
          class="absolute top-1 left-1 rounded bg-accent/80 px-1 py-0.5 text-[0.625rem] text-white"
          :class="isBatchMode ? 'ml-6' : ''"
        >
          {{ record.tag_name || record.tname }}
        </div>
        <img
          v-if="imgLoaded"
          :src="normalizeImageUrl(record.cover || record.covers[0])"
          class="h-full w-full object-cover"
          alt=""
        />
      </div>
      <div class="mt-2 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
        <div
          v-if="record.business !== 'cheese' && record.business !== 'pgc'"
          class="flex items-center gap-2"
          @click.stop
        >
          <img
            :src="normalizeImageUrl(record.author_face)"
            class="h-4 w-4 rounded-full"
            loading="lazy"
            @click="handleAuthorClick"
          />
          <!-- eslint-disable vue/no-v-html -- UP 主名经 utils/escapeHtml 转义 -->
          <span
            class="cursor-pointer transition-colors hover:text-accent"
            @click="handleAuthorClick"
            v-html="highlightedAuthorName"
          />
          <!-- eslint-enable vue/no-v-html -->
        </div>
        <div class="flex items-center gap-2">
          <span
            v-if="record.dt === 1 || record.dt === 3 || record.dt === 5 || record.dt === 7"
            class="text-[0.625rem]"
            >📱</span
          >
          <span v-else-if="record.dt === 2 || record.dt === 33" class="text-[0.625rem]">🖥</span>
          <span v-else-if="record.dt === 4 || record.dt === 6" class="text-[0.625rem]">📟</span>
          <span>{{ formatTimestamp(record.view_at) }}</span>
        </div>
      </div>
    </div>

    <!-- Other types: horizontal layout -->
    <div v-else class="flex gap-3">
      <!-- Cover -->
      <div class="relative h-20 w-32 flex-shrink-0 overflow-hidden rounded-xl sm:h-28 sm:w-40">
        <div
          v-if="isBatchMode"
          class="absolute top-2 left-2 z-10"
          @click.stop="$emit('toggle-selection', record)"
        >
          <div
            class="flex h-5 w-5 items-center justify-center rounded-lg border-2 transition-all"
            :class="isSelected ? 'border-accent bg-accent' : 'border-white/80 bg-black/20'"
          >
            <svg
              v-if="isSelected"
              class="h-3 w-3 text-white"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="3"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </div>
        <div
          v-if="isDownloaded && record.business === 'archive'"
          class="absolute top-0 left-0 z-10"
        >
          <div
            class="glass-badge rounded-tl-none rounded-br-lg !border-0 !bg-green-500/90 text-[0.625rem] !text-white"
          >
            <svg
              class="h-3 w-3"
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
            <span>已下载</span>
          </div>
        </div>
        <div
          v-if="isVideoFavorited && record.business !== 'live'"
          class="absolute top-0 right-0 z-10"
        >
          <div
            class="glass-badge rounded-tl-lg rounded-br-none !border-0 !bg-amber-500/90 text-[0.625rem] !text-white"
          >
            <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 24 24">
              <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
            </svg>
            <span>已收藏</span>
          </div>
        </div>
        <img
          v-if="record.cover"
          :src="normalizeImageUrl(record.cover)"
          class="h-full w-full object-cover"
          alt=""
          loading="lazy"
        />
        <div v-for="(cover, index) in record.covers" v-else :key="index" class="mb-1">
          <img
            :src="normalizeImageUrl(cover)"
            class="h-full w-full object-cover"
            alt=""
            loading="lazy"
          />
        </div>
        <!-- Tag badge on cover -->
        <div
          v-if="record.tag_name || record.tname"
          class="absolute top-1 left-1 rounded bg-accent/80 px-1 py-0.5 text-[0.625rem] text-white"
          :class="isBatchMode ? 'ml-6' : ''"
        >
          {{ record.tag_name || record.tname }}
        </div>
        <!-- Duration & progress -->
        <div
          v-if="
            record.business !== 'article-list' &&
            record.business !== 'article' &&
            record.business !== 'live'
          "
        >
          <div
            class="absolute right-1 bottom-1 rounded-lg bg-black/60 px-1.5 py-0.5 text-[0.625rem] font-medium text-white backdrop-blur-sm"
          >
            <span>{{ formatDuration(record.progress) }}</span>
            <span class="mx-0.5 opacity-50">/</span>
            <span>{{ formatDuration(record.duration) }}</span>
          </div>
          <div class="absolute bottom-0 left-0 h-0.5 w-full bg-black/20">
            <div
              class="h-full rounded-full bg-accent"
              :style="{ width: getProgressWidth(record.progress, record.duration) }"
            />
          </div>
        </div>
      </div>

      <!-- Content -->
      <div class="relative flex min-w-0 flex-1 flex-col justify-between">
        <!-- Action buttons -->
        <div
          v-if="!isBatchMode"
          class="absolute top-0 right-0 z-20 hidden items-center gap-1.5 sm:group-hover:flex"
        >
          <div
            v-if="record.business === 'archive'"
            class="glass-icon-btn !h-7 !w-7"
            title="下载"
            @click.stop.prevent="handleDownload"
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
          </div>
          <div
            v-if="record.business === 'archive'"
            class="glass-icon-btn !h-7 !w-7 hover:!bg-pink-500/20 hover:!text-pink-500"
            :class="{ '!bg-pink-500/20 !text-pink-500': isLiked }"
            title="点赞"
            @click.stop="handleLike"
          >
            <svg
              class="h-3.5 w-3.5"
              :fill="isLiked ? 'currentColor' : 'none'"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5"
              />
            </svg>
          </div>
          <div
            class="glass-icon-btn !h-7 !w-7 hover:!bg-red-500/20 hover:!text-red-500"
            title="删除"
            @click.stop="handleDelete"
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
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </div>
        </div>

        <div>
          <!-- eslint-disable vue/no-v-html -- 同上：标题已转义 -->
          <div
            class="line-clamp-2 text-sm font-medium text-gray-900 dark:text-gray-100"
            v-html="highlightedTitle"
          />
          <!-- eslint-enable vue/no-v-html -->
        </div>

        <div class="mt-1 flex items-center gap-2">
          <!-- Remark -->
          <div class="relative flex-1" @click.stop>
            <div class="flex items-center gap-1">
              <span class="text-[0.625rem] text-accent">备注:</span>
              <input
                v-model="remarkContent"
                type="text"
                placeholder="添加备注..."
                class="min-w-0 flex-1 border-b border-transparent bg-transparent px-1.5 py-0.5 text-[0.625rem] text-accent placeholder-accent/40 transition-colors hover:border-gray-200 focus:border-accent focus:ring-0 dark:hover:border-gray-600"
                @focus="handleRemarkFocus"
                @blur="handleRemarkBlur"
              />
              <span v-if="remarkTime" class="text-[0.625rem] text-gray-400">{{
                formatRemarkTime(remarkTime)
              }}</span>
            </div>
          </div>
        </div>

        <div
          class="mt-1 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400"
        >
          <div v-if="record.business === 'pgc'" class="truncate text-gray-500 dark:text-gray-400">
            {{ record.long_title }}
          </div>
          <div v-else class="flex min-w-0 items-center gap-2" @click.stop>
            <img
              :src="normalizeImageUrl(record.author_face)"
              class="h-4 w-4 flex-shrink-0 rounded-full"
              loading="lazy"
              @click="handleAuthorClick"
            />
            <!-- eslint-disable vue/no-v-html -- 同上：UP 主名已转义 -->
            <span
              class="cursor-pointer truncate transition-colors hover:text-accent"
              @click="handleAuthorClick"
              v-html="highlightedAuthorName"
            />
            <!-- eslint-enable vue/no-v-html -->
          </div>
          <div class="flex flex-shrink-0 items-center gap-2">
            <span
              v-if="record.dt === 1 || record.dt === 3 || record.dt === 5 || record.dt === 7"
              class="text-[0.625rem]"
              >📱</span
            >
            <span v-else-if="record.dt === 2 || record.dt === 33" class="text-[0.625rem]">🖥</span>
            <span v-else-if="record.dt === 4 || record.dt === 6" class="text-[0.625rem]">📟</span>
            <span>{{ formatTimestamp(record.view_at) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Download dialog -->
    <Teleport to="body">
      <DownloadDialog
        v-model:show="showDownloadDialog"
        :video-info="{
          title: record.title,
          author: record.author_name,
          bvid: record.bvid,
          cover: record.cover || record.covers?.[0],
          cid: record.cid
        }"
        :is-batch-download="false"
      />
    </Teleport>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMediaQuery } from '@vueuse/core';
import { showDialog, showNotify } from 'vant';
import {
  batchDeleteHistory,
  updateVideoRemark,
  deleteBilibiliHistory,
  toggleLike
} from '~/utils/api';
import 'vant/es/dialog/style';
import 'vant/es/popup/style';
import 'vant/es/field/style';
import DownloadDialog from './DownloadDialog.vue';
import { openInBrowser } from '~/utils/openUrl.js';
import { normalizeImageUrl } from '~/utils/imageUrl.js';
import { saveHistoryRecord } from '~/utils/historyRecordStore.js';
import { formatDuration, formatTimestamp } from '~/utils/format';
import { escapeHtml, highlightText } from '~/utils/escapeHtml';

const router = useRouter();
const isSmallScreen = useMediaQuery('(max-width: 639px)');

const props = defineProps({
  record: { type: Object, required: true },
  searchKeyword: { type: String, default: '' },
  searchType: { type: String, default: 'title' },
  isBatchMode: { type: Boolean, default: false },
  isSelected: { type: Boolean, default: false },
  remarkData: { type: Object, default: () => ({}) },
  isDownloaded: { type: Boolean, default: false },
  isVideoFavorited: { type: Boolean, default: false }
});

const emit = defineEmits(['toggle-selection', 'refresh-data', 'remark-updated']);

const cardRef = ref(null);
const imgLoaded = ref(false);
let observer = null;

const remarkContent = ref('');
const originalRemark = ref('');
const remarkTime = ref(null);

const highlightedTitle = computed(() => {
  if (props.searchType === 'all' || props.searchType === 'title') {
    return highlightText(props.record.title, props.searchKeyword, 'text-accent font-medium');
  }
  return escapeHtml(props.record.title);
});

const highlightedAuthorName = computed(() => {
  if (props.searchType === 'all' || props.searchType === 'author') {
    return highlightText(props.record.author_name, props.searchKeyword, 'text-accent font-medium');
  }
  return escapeHtml(props.record.author_name);
});

const isLiked = ref(false);

const handleLike = async () => {
  try {
    const newLikeState = !isLiked.value;
    const response = await toggleLike(props.record.bvid, newLikeState);
    if (response.data.status === 'success') {
      isLiked.value = newLikeState;
      showNotify({ type: 'success', message: newLikeState ? '已点赞' : '已取消点赞' });
    } else {
      throw new Error(response.data.message || '操作失败');
    }
  } catch (error) {
    showNotify({
      type: 'danger',
      message: error.response?.data?.detail || error.message || '点赞操作失败'
    });
  }
};

const handleClick = () => {
  if (props.isBatchMode) {
    emit('toggle-selection', props.record);
  } else if (isSmallScreen.value) {
    navigateToActionPage();
  } else {
    handleContentClick();
  }
};

const navigateToActionPage = () => {
  saveHistoryRecord(props.record);
  router.push({
    name: 'VideoActions',
    params: { bvid: props.record.bvid, viewAt: String(props.record.view_at) }
  });
};

const handleContentClick = async () => {
  let url = '';
  switch (props.record.business) {
    case 'archive':
      url = `https://www.bilibili.com/video/${props.record.bvid}`;
      break;
    case 'article':
      url = `https://www.bilibili.com/read/cv${props.record.oid}`;
      break;
    case 'article-list':
      url = `https://www.bilibili.com/read/readlist/rl${props.record.oid}`;
      break;
    case 'live':
      url = `https://live.bilibili.com/${props.record.oid}`;
      break;
    case 'pgc':
      url = props.record.uri || `https://www.bilibili.com/bangumi/play/ep${props.record.epid}`;
      break;
    case 'cheese':
      url = props.record.uri || `https://www.bilibili.com/cheese/play/ep${props.record.epid}`;
      break;
    default:
      return;
  }
  if (url) await openInBrowser(url);
};

const handleAuthorClick = async () => {
  await openInBrowser(`https://space.bilibili.com/${props.record.author_mid}`);
};

const getProgressWidth = (progress, duration) => {
  if (progress === -1) return '100%';
  if (duration === 0) return '0%';
  return `${(progress / duration) * 100}%`;
};

const handleDelete = async () => {
  try {
    const { getSyncConfig } = await import('~/utils/api');
    const syncConfigRes = await getSyncConfig();
    const syncDeleteToBilibili = syncConfigRes.data?.sync_delete_to_bilibili || false;
    await showDialog({
      title: '确认删除',
      message: syncDeleteToBilibili
        ? '确定要删除这条记录吗？此操作将同时删除B站服务器上的历史记录，不可恢复。'
        : '确定要删除这条记录吗？此操作不可恢复。',
      showCancelButton: true,
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      confirmButtonColor: 'var(--accent)'
    });
    if (syncDeleteToBilibili) {
      try {
        let kid = '';
        const business = props.record.business || 'archive';
        switch (business) {
          case 'archive':
            kid = `${business}_${props.record.oid}`;
            break;
          case 'live':
            kid = `${business}_${props.record.oid}`;
            break;
          case 'article':
            kid = `${business}_${props.record.oid}`;
            break;
          case 'pgc':
            kid = `${business}_${props.record.oid || props.record.ssid}`;
            break;
          case 'article-list':
            kid = `${business}_${props.record.oid || props.record.rlid}`;
            break;
          default:
            kid = `${business}_${props.record.oid || props.record.bvid}`;
            break;
        }
        if (kid) {
          const biliResponse = await deleteBilibiliHistory(kid, true);
          if (biliResponse.data.status !== 'success')
            throw new Error(biliResponse.data.message || '删除B站历史记录失败');
        }
      } catch (error) {
        console.error('B站历史记录删除失败:', error);
      }
    }
    const response = await batchDeleteHistory([
      { bvid: props.record.bvid, view_at: props.record.view_at }
    ]);
    if (response.data.status === 'success') {
      showNotify({
        type: 'success',
        message: response.data.message + (syncDeleteToBilibili ? '，并已同步删除B站历史记录' : '')
      });
      emit('refresh-data');
    } else {
      throw new Error(response.data.message || '删除失败');
    }
  } catch (error) {
    if (error.toString().includes('cancel')) return;
    showNotify({
      type: 'danger',
      message: error.response?.data?.detail || error.message || '删除失败'
    });
  }
};

const formatRemarkTime = (timestamp) => {
  if (!timestamp) return '';
  return new Date(timestamp * 1000).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
};

const initRemark = () => {
  const key = `${props.record.bvid}_${props.record.view_at}`;
  const data = props.remarkData[key];
  if (data) {
    remarkContent.value = data.remark || '';
    remarkTime.value = data.remark_time || null;
    originalRemark.value = remarkContent.value;
  } else {
    remarkContent.value = '';
    remarkTime.value = null;
    originalRemark.value = '';
  }
};

const handleRemarkBlur = async () => {
  if (remarkContent.value === originalRemark.value) return;
  try {
    const response = await updateVideoRemark(
      props.record.bvid,
      props.record.view_at,
      remarkContent.value
    );
    if (response.data.status === 'success') {
      if (remarkContent.value) showNotify({ type: 'success', message: '备注已保存' });
      originalRemark.value = remarkContent.value;
      remarkTime.value = response.data.data.remark_time;
      emit('remark-updated', {
        bvid: props.record.bvid,
        view_at: props.record.view_at,
        remark: remarkContent.value,
        remark_time: response.data.data.remark_time
      });
    }
  } catch (error) {
    showNotify({ type: 'danger', message: `保存备注失败：${error.message}` });
    remarkContent.value = originalRemark.value;
  }
};

onMounted(() => {
  initRemark();
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          imgLoaded.value = true;
          observer.disconnect();
          break;
        }
      }
    },
    { rootMargin: '200px' }
  );
  if (cardRef.value) {
    observer.observe(cardRef.value);
  }
});

onBeforeUnmount(() => {
  if (observer) observer.disconnect();
});
watch(
  () => props.remarkData,
  () => {
    initRemark();
  },
  { deep: true }
);

const showDownloadDialog = ref(false);
const handleDownload = () => {
  showDownloadDialog.value = true;
};
</script>
