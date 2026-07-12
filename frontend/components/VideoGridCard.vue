<template>
  <div
    ref="cardRef"
    class="rounded-md overflow-hidden border transition-all duration-200 relative group"
    :class="cardClasses"
  >
    <div v-if="selectMode" class="absolute top-1.5 left-1.5 z-10">
      <input
        type="checkbox"
        :checked="isSelected"
        @click.stop="$emit('toggle-select', video)"
        class="w-4 h-4 rounded border-gray-300 text-[#fb7299] focus:ring-[#fb7299] bg-white/90"
      />
    </div>

    <div
      class="relative pb-[56.25%] overflow-hidden cursor-pointer group"
      @click="handleClick"
    >
      <img
        v-if="imgLoaded"
        :src="imageUrl"
        :alt="video.title"
        class="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-300 bg-gray-200 dark:bg-gray-700"
        :onerror="errorImageUrl"
      />

      <slot name="cover-top-left" :video="video">
        <div v-if="showCategory && categoryName"
             class="absolute top-1 left-1 bg-[#fb7299]/80 px-1 py-0.5 rounded text-white text-[10px]"
             :class="selectMode ? 'ml-6' : ''">
          {{ categoryName }}
        </div>
      </slot>

      <slot name="cover-top-right" :video="video"></slot>

      <slot name="cover-bottom-left" :video="video"></slot>

      <slot name="cover-bottom-right" :video="video">
        <div v-if="durationText" class="absolute bottom-1 right-1 bg-black/60 px-1 py-0.5 rounded text-white text-[10px]">
          {{ durationText }}
        </div>
      </slot>

      <div v-if="progress !== undefined && progress !== null" class="absolute bottom-0 left-0 w-full">
        <div class="absolute bottom-0 left-0 h-0.5 w-full bg-black/20">
          <div class="h-full bg-[#fb7299] rounded-full" :style="{ width: progressWidth }"></div>
        </div>
      </div>

      <div v-if="!selectMode && showActions" class="absolute right-1.5 top-1.5 z-20 hidden group-hover:flex items-center gap-1">
        <slot name="actions" :video="video"></slot>
      </div>

      <slot name="badges" :video="video"></slot>
    </div>

    <div class="p-2 flex flex-col space-y-1">
      <div
        class="line-clamp-2 text-xs text-gray-900 dark:text-gray-100 font-medium cursor-pointer"
        @click="handleClick"
        :title="video.title"
      >
        <slot name="title" :video="video">
          {{ video.title }}
        </slot>
      </div>

      <slot name="meta-top" :video="video"></slot>

      <div v-if="showOwner && ownerName" class="flex items-center space-x-1">
        <img
          v-if="video.owner_face || video.author_face"
          :src="ownerFaceUrl"
          :alt="ownerName"
          class="w-3.5 h-3.5 rounded-full object-cover cursor-pointer"
          loading="lazy"
          :onerror="ownerErrorUrl"
          @click.stop="handleOwnerClick"
        />
        <span
          class="text-[10px] text-gray-600 dark:text-gray-400 truncate cursor-pointer hover:text-[#fb7299] transition-colors"
          @click.stop="handleOwnerClick"
          :title="ownerName"
        >
          {{ ownerName }}
        </span>
      </div>

      <div class="flex justify-between items-center text-[10px] text-gray-500">
        <slot name="meta-bottom-left" :video="video">
          <span v-if="showViews">{{ viewsText }} 次观看</span>
        </slot>
        <slot name="meta-bottom-right" :video="video">
          <span v-if="showTime && timeText">{{ timeText }}</span>
        </slot>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { normalizeImageUrl } from '~/utils/imageUrl.js'
import { formatDuration } from '~/utils/format'

const props = defineProps({
  video: { type: Object, required: true },
  selectMode: { type: Boolean, default: false },
  isSelected: { type: Boolean, default: false },
  showCategory: { type: Boolean, default: true },
  showOwner: { type: Boolean, default: true },
  showViews: { type: Boolean, default: true },
  showTime: { type: Boolean, default: true },
  showActions: { type: Boolean, default: true },
  durationKey: { type: String, default: 'duration' },
  categoryKey: { type: String, default: 'tname' },
  progress: { type: Number, default: null },
  totalDuration: { type: Number, default: null },
  timeField: { type: String, default: 'add_at' },
  timeFormat: { type: String, default: 'date' },
  cardStyle: { type: String, default: 'default' },
  titleKey: { type: String, default: 'title' },
  ownerKey: { type: String, default: 'owner_name' },
  ownerFaceKey: { type: String, default: 'owner_face' },
  coverKey: { type: String, default: 'pic' },
  viewsKey: { type: String, default: 'view' },
})

const emit = defineEmits(['click', 'toggle-select', 'owner-click'])

const cardRef = ref(null)
const imgLoaded = ref(false)
let observer = null

const imageUrl = computed(() => {
  const cover = props.video[props.coverKey]
    || props.video.cover
    || props.video.pic
    || (props.video.covers && props.video.covers[0])
  return normalizeImageUrl(cover)
})
const errorImageUrl = "this.src='https://i0.hdslb.com/bfs/archive/c9e72655b7c9c9c68a30d3275313c501e68427d1.jpg'"
const ownerFaceUrl = computed(() => {
  const face = props.video[props.ownerFaceKey] || props.video.author_face
  return normalizeImageUrl(face)
})
const ownerErrorUrl = "this.src='https://static.hdslb.com/images/member/noface.gif'"

const categoryName = computed(() => {
  return props.video[props.categoryKey]
    || props.video.tag_name
    || props.video.tname
    || ''
})

const ownerName = computed(() => {
  return props.video[props.ownerKey] || props.video.author_name || ''
})

const durationText = computed(() => {
  const dur = props.video[props.durationKey]
  if (dur === undefined || dur === null) return ''
  if (props.progress !== undefined && props.progress !== null && props.totalDuration) {
    return `${formatDuration(props.progress)}/${formatDuration(props.totalDuration)}`
  }
  return formatDuration(dur)
})

const progressWidth = computed(() => {
  if (props.progress === undefined || props.progress === null || !props.totalDuration) return '0%'
  if (props.progress === -1) return '100%'
  if (props.totalDuration === 0) return '0%'
  return `${(props.progress / props.totalDuration) * 100}%`
})

const viewsText = computed(() => {
  const count = props.video[props.viewsKey]
  if (!count) return '0'
  if (count >= 10000) return (count / 10000).toFixed(1) + '万'
  return count.toString()
})

const timeText = computed(() => {
  const timestamp = props.video[props.timeField]
  if (!timestamp) return ''
  if (props.timeFormat === 'date') {
    const date = new Date(timestamp * 1000)
    const pad = (n) => String(n).padStart(2, '0')
    return `${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
  }
  return ''
})

const cardClasses = computed(() => {
  const base = props.cardStyle === 'glass'
    ? 'glass-card-hover'
    : 'bg-white/50 dark:bg-gray-800/50'

  if (props.selectMode) {
    return props.isSelected
      ? `${base} border-[#fb7299] ring-1 ring-[#fb7299]/40`
      : `${base} border-gray-200/50 dark:border-gray-700/50`
  }
  return `${base} border-gray-200/50 dark:border-gray-700/50 hover:border-[#fb7299] hover:shadow-sm`
})

const handleClick = () => {
  if (props.selectMode) {
    emit('toggle-select', props.video)
  } else {
    emit('click', props.video)
  }
}

const handleOwnerClick = () => {
  emit('owner-click', props.video)
}

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          imgLoaded.value = true
          observer.disconnect()
          break
        }
      }
    },
    { rootMargin: '200px' }
  )
  if (cardRef.value) {
    observer.observe(cardRef.value)
  }
})

onBeforeUnmount(() => {
  if (observer) observer.disconnect()
})
</script>
