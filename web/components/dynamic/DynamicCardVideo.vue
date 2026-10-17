<template>
  <div class="overflow-hidden rounded-lg border bg-white">
    <!-- 头部：头像 + 名称 + 时间 + 动态链接 -->
    <div class="flex items-center px-3 py-2">
      <img
        v-if="displayFaceUrl"
        :src="displayFaceUrl"
        class="h-6 w-6 rounded-full border object-cover"
        alt="face"
      />
      <div class="ml-2 min-w-0">
        <div class="truncate text-sm font-medium">
          {{ item.author_name || `UID ${item.host_mid || ''}` }}
        </div>
        <div v-if="formattedTime !== '-'" class="truncate text-[0.6875rem] text-gray-500">
          {{ formattedTime }}
        </div>
      </div>
      <div class="ml-auto flex items-center space-x-2">
        <button
          v-if="item.id_str"
          type="button"
          class="text-[0.6875rem] text-accent hover:underline"
          @click="openLink(opusUrl)"
        >
          查看动态
        </button>
        <button
          v-if="item.id_str"
          type="button"
          class="text-[0.6875rem] text-red-500 hover:underline"
          @click.stop="handleDelete"
        >
          删除
        </button>
      </div>
    </div>

    <!-- 主体：封面（更小） + 文本，封面/标题可跳转视频 -->
    <div class="px-3 pb-3">
      <div class="flex items-start">
        <div
          class="relative w-40 flex-shrink-0 cursor-pointer overflow-hidden rounded-md hover:opacity-90 sm:w-48 md:w-56"
          role="link"
          tabindex="0"
          @click="openLink(videoUrl)"
          @keydown.enter="openLink(videoUrl)"
        >
          <div class="w-full" style="aspect-ratio: 16 / 9">
            <img
              v-if="coverUrl"
              :src="coverUrl"
              class="block h-full w-full object-cover"
              alt="cover"
              loading="lazy"
            />
            <div
              v-else
              class="flex h-full w-full items-center justify-center text-xs text-gray-400"
            >
              no cover
            </div>
          </div>
          <div
            v-if="item.bvid"
            class="absolute bottom-1 left-1 rounded bg-black/60 px-1.5 py-0.5 text-[0.625rem] text-white"
          >
            {{ item.bvid }}
          </div>
        </div>
        <div class="ml-3 min-w-0 flex-1">
          <div
            role="link"
            tabindex="0"
            class="line-clamp-2 cursor-pointer text-sm leading-5 font-semibold hover:underline"
            :title="item.title || item.txt || ''"
            @click="openLink(videoUrl)"
            @keydown.enter="openLink(videoUrl)"
          >
            {{ item.title || item.txt || '视频动态' }}
          </div>
          <div v-if="item.desc || item.txt" class="mt-1 line-clamp-2 text-xs text-gray-600">
            {{ item.desc || item.txt }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { openInBrowser } from '~/utils/openUrl';
import { toStaticUrl, normalizeImageUrl } from '~/utils/imageUrl';

const props = defineProps({
  item: { type: Object, required: true },
  faceUrl: { type: String, default: '' }
});

// 优先使用动态中的作者头像，其次使用传入的 faceUrl
const displayFaceUrl = computed(() => {
  const face = props.item?.author_face || props.faceUrl || '';
  if (!face) return '';
  if (face.startsWith('http://') || face.startsWith('https://')) return normalizeImageUrl(face);
  return toStaticUrl(face);
});

const coverUrl = computed(() => {
  const it = props.item || {};
  if (it.cover) return toStaticUrl(it.cover);
  const list = Array.isArray(it.media_locals) ? it.media_locals : [];
  if (list.length) return toStaticUrl(list[0]);
  return '';
});

const formattedTime = computed(() => {
  const ts = props.item?.publish_ts;
  if (!ts) return '-';
  try {
    const d = new Date(ts * 1000);
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
  } catch {
    return String(ts);
  }
});

const opusUrl = computed(() =>
  props.item?.id_str ? `https://www.bilibili.com/opus/${props.item.id_str}` : '#'
);
const videoUrl = computed(() =>
  props.item?.bvid ? `https://www.bilibili.com/video/${props.item.bvid}` : opusUrl.value
);

const openLink = (url) => {
  try {
    openInBrowser(url);
  } catch {
    window.open(url, '_blank');
  }
};
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
