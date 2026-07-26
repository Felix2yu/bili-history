<template>
  <div class="border rounded-lg bg-white overflow-hidden">
    <!-- 头部：头像 + 名称 + 时间 + 动态链接 -->
    <div class="flex items-center px-3 py-2">
      <img v-if="displayFaceUrl" :src="displayFaceUrl" class="w-6 h-6 rounded-full object-cover border" alt="face" />
      <div class="ml-2 min-w-0">
        <div class="text-sm font-medium truncate">{{ item.author_name || `UID ${item.host_mid || ''}` }}</div>
        <div v-if="formattedTime !== '-'" class="text-[11px] text-gray-500 truncate">{{ formattedTime }}</div>
      </div>
      <div class="ml-auto flex items-center space-x-2">
        <button
          v-if="item.id_str"
          type="button"
          class="text-[11px] text-accent hover:underline"
          @click="openLink(opusUrl)"
        >查看动态</button>
        <button
          v-if="item.id_str"
          type="button"
          class="text-[11px] text-red-500 hover:underline"
          @click.stop="handleDelete"
        >删除</button>
      </div>
    </div>

    <!-- 主体：配文/标题内容 + 图片/实况九宫格（如有） -->
    <div class="px-3 pb-3">
      <!-- DYNAMIC_TYPE_DRAW: 展示 opus 标题与摘要 -->
      <template v-if="isDraw">
        <div
          class="text-sm font-semibold text-gray-900 leading-6"
          v-if="drawTitle"
        >{{ drawTitle }}</div>
        <div
          class="mt-1 text-sm text-gray-700 leading-6"
          v-if="drawSummary"
        >
          <span class="whitespace-pre-wrap">
            <template v-for="(seg, i) in parsedSummary" :key="i">
              <span v-if="seg.type==='text'">{{ seg.text }}</span>
              <img
                v-else
                :src="seg.url"
                :alt="seg.name"
                class="emoji emoji-lg inline-block align-text-bottom cursor-zoom-in hover:opacity-90 transition"
                role="button"
                tabindex="0"
                title="Click to preview"
                @click.stop="openPreview('image', seg.url)"
                @keydown.enter.stop="openPreview('image', seg.url)"
              />
            </template>
          </span>
        </div>
      </template>
      <!-- 其他类型：展示 txt（解析表情） -->
      <div
        v-else-if="item.txt"
        role="link"
        tabindex="0"
        @click="openLink(opusUrl)"
        @keydown.enter="openLink(opusUrl)"
        class="text-sm text-gray-800 leading-6 hover:underline cursor-pointer"
      >
        <span class="whitespace-pre-wrap">
          <template v-for="(seg, i) in parsedTxt" :key="'t'+i">
            <span v-if="seg.type==='text'">{{ seg.text }}</span>
            <img
              v-else
              :src="seg.url"
              :alt="seg.name"
              class="emoji emoji-lg inline-block align-text-bottom cursor-zoom-in hover:opacity-90 transition"
              role="button"
              tabindex="0"
              title="Click to preview"
              @click.stop="openPreview('image', seg.url)"
              @keydown.enter.stop="openPreview('image', seg.url)"
            />
          </template>
        </span>
      </div>

      <div v-if="displayMedias.length" class="mt-2 grid gap-1 md:gap-2 grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6">
        <template v-for="(m, idx) in displayMedias" :key="idx">
          <!-- 普通图片 -->
          <div v-if="m.kind==='image'" class="relative rounded-md overflow-hidden hover:opacity-90 cursor-pointer h-28 sm:h-32 md:h-36" role="button" tabindex="0" @click="openPreview('image', m.url)" @keydown.enter="openPreview('image', m.url)">
            <img :src="m.url" class="block w-full h-full object-cover" loading="lazy" />
          </div>
          <!-- 实况照片（悬停播放） -->
          <div v-else class="relative rounded-md overflow-hidden hover:opacity-90 cursor-pointer h-28 sm:h-32 md:h-36"
               role="button" tabindex="0"
               @mouseenter="handleLiveEnter(idx)" @mouseleave="handleLiveLeave(idx)"
               @click="openPreview('video', m.videoUrl, m.coverUrl)" @keydown.enter="openPreview('video', m.videoUrl, m.coverUrl)">
            <video :poster="m.coverUrl" :src="m.videoUrl" muted playsinline loop
                   class="absolute inset-0 w-full h-full object-cover"
                   :ref="el => setLiveRef(idx, el)"
            ></video>
            <!-- 右下角 实况 徽标 -->
            <div class="absolute bottom-1 right-1 px-1.5 py-0.5 bg-black/60 text-white text-[10px] flex items-center rounded">
              <img src="/live.svg" class="w-3 h-3 mr-1 filter invert" alt="live" />
              <span>实况</span>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 预览弹层 -->
    <Teleport to="body">
      <div v-if="showPreview" class="fixed inset-0 z-50 bg-black/80 flex items-center justify-center" @click="closePreview">
        <div class="max-w-[95vw] max-h-[90vh] relative" @click.stop>
          <!-- 左切换按钮 -->
          <button
            v-if="previewType==='image' && previewImages.length > 1"
            class="absolute left-2 top-1/2 -translate-y-1/2 w-10 h-10 rounded-full bg-black/60 text-white flex items-center justify-center hover:bg-black/80 transition z-10"
            @click="prevImage"
          >
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <!-- 加载动画 -->
          <div v-if="previewLoading" class="absolute inset-0 flex items-center justify-center">
            <div class="w-10 h-10 border-4 border-white/30 border-t-white rounded-full animate-spin"></div>
          </div>
          <!-- 图片 -->
          <img v-if="previewType==='image'" :src="previewSrc" class="max-w-[95vw] max-h-[90vh] object-contain rounded-md transition-opacity duration-200" :class="{'opacity-0': previewLoading}" @load="previewLoading = false" />
          <video v-else :src="previewSrc" :poster="previewPoster" controls autoplay loop muted class="max-w-[95vw] max-h-[90vh] rounded-md"></video>
          <!-- 右切换按钮 -->
          <button
            v-if="previewType==='image' && previewImages.length > 1"
            class="absolute right-2 top-1/2 -translate-y-1/2 w-10 h-10 rounded-full bg-black/60 text-white flex items-center justify-center hover:bg-black/80 transition z-10"
            @click="nextImage"
          >
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
          <!-- 关闭按钮 -->
          <button class="absolute -top-3 -right-3 w-8 h-8 rounded-full bg-black/70 text-white flex items-center justify-center hover:bg-black/90" @click="closePreview">×</button>
          <!-- 图片计数 -->
          <div v-if="previewType==='image' && previewImages.length > 1" class="absolute bottom-2 left-1/2 -translate-x-1/2 px-3 py-1 bg-black/60 rounded-full text-white text-xs">
            {{ previewIndex + 1 }} / {{ previewImages.length }}
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { openInBrowser } from '~/utils/openUrl'
import { toStaticUrl } from '~/utils/imageUrl'
import { deleteDynamicItem } from '~/utils/api'

const props = defineProps({
  item: { type: Object, required: true },
  faceUrl: { type: String, default: '' }
})

const emit = defineEmits(['deleted'])

// 优先使用动态中的作者头像，其次使用传入的 faceUrl
const displayFaceUrl = computed(() => {
  const face = props.item?.author_face || props.faceUrl || ''
  if (!face) return ''
  if (face.startsWith('http://') || face.startsWith('https://')) return face
  return toStaticUrl(face)
})

const handleDelete = async () => {
  if (!props.item?.id_str) return
  if (!confirm('确定要删除这条动态吗？')) return
  try {
    await deleteDynamicItem(props.item.id_str)
    emit('deleted', props.item.id_str)
  } catch (e) {
    alert('删除失败: ' + (e.message || e))
  }
}

// 是否图文动态
const isDraw = computed(() => String(props.item?.type || '') === 'DYNAMIC_TYPE_DRAW')
const drawTitle = computed(() => props.item?.opus_title || '')
const drawSummary = computed(() => props.item?.opus_summary_text || '')

// 归一化扩展名检查
const isVideoPath = (p) => /\.mp4$/i.test(String(p || '').replace(/\\/g, '/'))
const isImagePath = (p) => /\.(png|jpe?g|gif|webp)$/i.test(String(p || '').replace(/\\/g, '/'))
const getNameFromPath = (p) => {
  const filename = String(p || '').split(/[/\\]/).pop() || ''
  const decoded = decodeURIComponent(filename)
  return decoded.replace(/\.[^.]+$/, '')
}

// 从摘要中提取 [xxx] 表情名集合
const extractEmojiNames = (text) => {
  const set = new Set()
  if (!text) return set
  const re = /\[([^\[\]]+?)\]/g
  let m
  while ((m = re.exec(text)) !== null) {
    set.add(m[1])
  }
  return set
}

const emojiNamesFromSummary = computed(() => extractEmojiNames(drawSummary.value))
const emojiNamesFromTxt = computed(() => extractEmojiNames(props.item?.txt || ''))
const allEmojiNames = computed(() => new Set([...emojiNamesFromSummary.value, ...emojiNamesFromTxt.value]))

// 构建 emoji 名称到图片URL的映射（仅使用摘要中出现过的表情名）
const emojiMap = computed(() => {
  const map = {}
  const ml = Array.isArray(props.item?.media_locals) ? props.item.media_locals : []
  for (const p of ml) {
    if (!isImagePath(p)) continue
    const name = getNameFromPath(p)
    if (/^live_/i.test(name)) continue
    if (allEmojiNames.value.has(name)) {
      map[name] = toStaticUrl(p)
    }
  }
  return map
})

// 将摘要解析为文本/emoji 片段
const parseWithEmoji = (text) => {
  const map = emojiMap.value || {}
  if (!text) return []
  const result = []
  const re = /\[([^\[\]]+?)\]/g
  let last = 0
  let m
  while ((m = re.exec(text)) !== null) {
    const idx = m.index
    const raw = m[0]
    const name = m[1]
    if (idx > last) {
      result.push({ type: 'text', text: text.slice(last, idx) })
    }
    if (map[name]) {
      result.push({ type: 'emoji', name, url: map[name] })
    } else {
      result.push({ type: 'text', text: raw })
    }
    last = idx + raw.length
  }
  if (last < text.length) {
    result.push({ type: 'text', text: text.slice(last) })
  }
  return result
}

const parsedSummary = computed(() => {
  const text = drawSummary.value || ''
  return parseWithEmoji(text)
})

const parsedTxt = computed(() => parseWithEmoji(props.item?.txt || ''))

// 判断是否为完整URL
const isFullUrl = (p) => p && (p.startsWith('http://') || p.startsWith('https://'))

// 获取图片URL（完整URL直接使用，否则转换为本地静态URL）
const getImageUrl = (p) => isFullUrl(p) ? p : toStaticUrl(p)

// 构造展示媒体：普通图片 + 实况照片（由 live_media_locals 配对 png+mp4）
const displayMedias = computed(() => {
  const medias = []
  const ml = Array.isArray(props.item?.media_locals) ? props.item.media_locals : []
  for (const p of ml) {
    if (isImagePath(p) && !/live_/i.test(p)) {
      const name = getNameFromPath(p)
      // 过滤掉作为emoji使用的图片
      if (!allEmojiNames.value.has(name)) {
        medias.push({ kind: 'image', url: getImageUrl(p) })
      }
    }
  }

  const live = Array.isArray(props.item?.live_media_locals) ? props.item.live_media_locals : []
  if (live.length) {
    const covers = live.filter(isImagePath)
    const videos = live.filter(isVideoPath)
    const n = Math.min(covers.length, videos.length, 9)
    for (let i = 0; i < n; i++) {
      medias.push({ kind: 'live', coverUrl: getImageUrl(covers[i]), videoUrl: getImageUrl(videos[i]) })
    }
  }

  return medias.slice(0, 12) // 限制单条最多展示若干项
})

// 网格固定为每行最多6列（随断点 3/4/5/6 列）

const formattedTime = computed(() => {
  const ts = props.item?.publish_ts
  if (!ts) return '-'
  try {
    const d = new Date(ts * 1000)
    const pad = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  } catch {
    return String(ts)
  }
})

const opusUrl = computed(() => props.item?.id_str ? `https://www.bilibili.com/opus/${props.item.id_str}` : '#')

const openLink = (url) => {
  try { openInBrowser(url) } catch { window.open(url, '_blank') }
}

// 实况播放控制
const liveRefs = ref({})
const setLiveRef = (idx, el) => {
  if (!liveRefs.value) liveRefs.value = {}
  if (el) liveRefs.value[idx] = el
  else delete liveRefs.value[idx]
}
const handleLiveEnter = (idx) => {
  const v = liveRefs.value?.[idx]
  if (v) {
    try { v.play() } catch (e) {}
  }
}
const handleLiveLeave = (idx) => {
  const v = liveRefs.value?.[idx]
  if (v) {
    try { v.pause(); v.currentTime = 0 } catch (e) {}
  }
}

// 预览逻辑
const showPreview = ref(false)
const previewType = ref('image')
const previewSrc = ref('')
const previewPoster = ref('')
const previewImages = ref([])
const previewIndex = ref(0)
const previewLoading = ref(false)

const openPreview = (type, src, poster = '') => {
  previewType.value = type
  previewPoster.value = poster
  previewLoading.value = true

  // 收集所有图片用于切换
  if (type === 'image') {
    const allImages = displayMedias.value
      .filter(m => m.kind === 'image')
      .map(m => m.url)
    previewImages.value = allImages
    previewIndex.value = allImages.indexOf(src)
    if (previewIndex.value === -1) previewIndex.value = 0
    previewSrc.value = allImages[previewIndex.value] || src
  } else {
    previewImages.value = []
    previewIndex.value = 0
    previewSrc.value = src
  }

  showPreview.value = true
}

const switchImage = (newSrc) => {
  previewLoading.value = true
  previewSrc.value = newSrc
}

const prevImage = () => {
  if (previewImages.value.length <= 1) return
  const newIndex = (previewIndex.value - 1 + previewImages.value.length) % previewImages.value.length
  previewIndex.value = newIndex
  switchImage(previewImages.value[newIndex])
}

const nextImage = () => {
  if (previewImages.value.length <= 1) return
  const newIndex = (previewIndex.value + 1) % previewImages.value.length
  previewIndex.value = newIndex
  switchImage(previewImages.value[newIndex])
}

const closePreview = () => { showPreview.value = false }

// 键盘左右键切换图片
const handleKeydown = (e) => {
  if (!showPreview.value) return
  if (e.key === 'ArrowLeft') prevImage()
  else if (e.key === 'ArrowRight') nextImage()
  else if (e.key === 'Escape') closePreview()
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
/* 使emoji图片与文字同高 */
.emoji {
  height: 1em;
  width: 1em;
  margin: 0 2px;
}
.emoji-lg {
  height: 52px;
  width: 52px;
}
</style>


