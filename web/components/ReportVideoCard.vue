<template>
  <div
    ref="cardRef"
    class="glass-card flex gap-3 p-3 transition-shadow duration-200 hover:shadow-md"
  >
    <!-- 封面 -->
    <div
      class="relative aspect-video w-32 flex-shrink-0 overflow-hidden rounded-lg bg-gray-100 sm:w-40 dark:bg-gray-800"
    >
      <img
        v-if="video.cover && imgLoaded"
        :src="coverUrl"
        :alt="video.title"
        class="h-full w-full object-cover"
      />
      <div
        v-else-if="!video.cover"
        class="flex h-full w-full items-center justify-center text-gray-400"
      >
        <svg
          class="h-8 w-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z"
          />
        </svg>
      </div>
      <!-- 时长角标 -->
      <div
        class="absolute right-1 bottom-1 rounded bg-black/70 px-1.5 py-0.5 text-[0.625rem] text-white"
      >
        {{ formatDuration(video.duration) }}
      </div>
    </div>

    <!-- 信息 -->
    <div class="flex min-w-0 flex-1 flex-col justify-between">
      <div>
        <!-- 标题 -->
        <h4 class="line-clamp-2 text-sm leading-snug font-medium text-gray-800 dark:text-gray-200">
          {{ video.title }}
        </h4>

        <!-- UP主 -->
        <div class="mt-1.5 flex items-center gap-1.5">
          <img
            v-if="video.author_face && faceLoaded"
            :src="authorFaceUrl"
            class="h-4 w-4 rounded-full object-cover"
          />
          <span class="truncate text-xs text-gray-500 dark:text-gray-400">{{
            video.author_name
          }}</span>
        </div>

        <!-- 标签行 -->
        <div class="mt-1.5 flex flex-wrap items-center gap-1.5">
          <span
            v-if="video.main_category || video.tag_name"
            class="inline-flex items-center rounded bg-accent/10 px-1.5 py-0.5 text-[0.625rem] font-medium text-accent"
          >
            {{ video.main_category || video.tag_name }}
          </span>
          <span class="text-[0.625rem] text-gray-400 dark:text-gray-500">
            {{ getDeviceName(video.dt) }}
          </span>
          <span class="text-[0.625rem] text-gray-400 dark:text-gray-500">
            {{ getBusinessType(video.business) }}
          </span>
        </div>
      </div>

      <!-- 底部信息 -->
      <div class="mt-2 flex items-center justify-between">
        <div class="flex items-center gap-3 text-[0.6875rem] text-gray-400 dark:text-gray-500">
          <span>{{ formatTimestamp(video.view_at) }}</span>
        </div>

        <!-- 完播率 -->
        <div class="flex items-center gap-1.5">
          <div class="h-1.5 w-16 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="
                completionRate >= 90
                  ? 'bg-green-500'
                  : completionRate >= 50
                    ? 'bg-accent'
                    : 'bg-gray-400'
              "
              :style="{ width: `${Math.min(completionRate, 100)}%` }"
            />
          </div>
          <span class="text-[0.625rem] text-gray-400 dark:text-gray-500">{{
            completionRateText
          }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import { formatDuration, formatTimestamp, getBusinessType } from '~/utils/format';
import { normalizeImageUrl } from '~/utils/imageUrl';

const props = defineProps({
  video: {
    type: Object,
    required: true
  }
});

const cardRef = ref(null);
const imgLoaded = ref(false);
const faceLoaded = ref(false);
let observer = null;

const completionRate = computed(() => {
  if (props.video.duration <= 0) return 0;
  if (props.video.progress === -1) return 100;
  return Math.round((props.video.progress / props.video.duration) * 100);
});

const completionRateText = computed(() => {
  if (props.video.progress === -1) return '看完';
  if (completionRate.value >= 90) return '看完';
  return `${completionRate.value}%`;
});

const coverUrl = computed(() => normalizeImageUrl(props.video.cover));
const authorFaceUrl = computed(() => normalizeImageUrl(props.video.author_face));

function getDeviceName(dt) {
  switch (dt) {
    case 1:
    case 3:
    case 5:
    case 7:
    case 33:
      return '手机';
    case 2:
      return '电脑';
    case 4:
    case 6:
      return '平板';
    default:
      return '其他';
  }
}

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          imgLoaded.value = true;
          faceLoaded.value = true;
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
</script>

<style scoped>
@reference "../assets/css/main.css";
.glass-card {
  @apply rounded-xl border border-gray-200/50 bg-white/80 shadow-sm backdrop-blur-sm dark:border-gray-700/50 dark:bg-gray-800/80;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
