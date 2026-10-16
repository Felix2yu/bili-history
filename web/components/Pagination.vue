<template>
  <div class="mx-auto max-w-4xl">
    <div class="glass-card px-4 py-3 flex items-center justify-between gap-2">
      <!-- Previous -->
      <button
        @click="handlePageChange(currentPage - 1)"
        :disabled="currentPage === 1"
        class="flex shrink-0 items-center gap-1 px-3 py-2 rounded-xl text-sm text-gray-600 dark:text-gray-400 hover:bg-accent/10 hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-200"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        <span class="hidden sm:inline">上一页</span>
      </button>

      <!-- Center controls -->
      <div class="flex min-w-0 flex-1 items-center justify-center gap-3 sm:gap-4">
        <!-- Page size -->
        <div v-if="showPageSizeControl" class="flex shrink-0 items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
          <span>每页</span>
          <input
            type="number"
            v-model="pageSizeInput"
            @input="handlePageSizeChange"
            @blur="handlePageSizeBlur"
            min="10"
            max="100"
            class="h-7 w-14 rounded-lg glass-input px-2 text-center text-xs text-gray-700 dark:text-gray-300 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
          />
          <span>条</span>
        </div>

        <!-- Page number -->
        <div class="flex shrink-0 items-center gap-1.5 text-gray-700 dark:text-gray-300">
          <input
            ref="pageInput"
            type="number"
            v-model="currentPageInput"
            @keyup.enter="handleJumpPage"
            @blur="handleJumpPage"
            @focus="handleFocus"
            min="1"
            :max="totalPages"
            class="h-8 w-14 rounded-lg glass-input px-2 text-center text-sm text-gray-700 dark:text-gray-300 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
          />
          <span class="text-gray-400 dark:text-gray-500">/ {{ totalPages }}</span>
        </div>
      </div>

      <!-- Next -->
      <button
        @click="handlePageChange(currentPage + 1)"
        :disabled="currentPage === totalPages"
        class="flex shrink-0 items-center gap-1 px-3 py-2 rounded-xl text-sm text-gray-600 dark:text-gray-400 hover:bg-accent/10 hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-200"
      >
        <span class="hidden sm:inline">下一页</span>
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const props = defineProps({
  currentPage: { type: Number, required: true },
  totalPages: { type: Number, required: true },
  pageSize: { type: Number, default: 0 },
  useRouting: { type: Boolean, default: false },
})

const emit = defineEmits(['page-change', 'update:page-size'])
const router = useRouter()
const route = useRoute()

const currentPageInput = ref(props.currentPage.toString())
const pageSizeInput = ref(props.pageSize > 0 ? props.pageSize.toString() : '')
const showPageSizeControl = ref(props.pageSize > 0)

watch(() => props.currentPage, (newPage) => { currentPageInput.value = newPage.toString() })
watch(() => props.pageSize, (newPageSize) => {
  showPageSizeControl.value = newPageSize > 0
  pageSizeInput.value = newPageSize > 0 ? newPageSize.toString() : ''
}, { immediate: true })

const handlePageChange = (newPage) => {
  if (newPage >= 1 && newPage <= props.totalPages) {
    if (props.useRouting) {
      if (route.name && (route.name === 'Search' || route.name === 'SearchPage')) {
        router.push(newPage === 1 ? `/search/${route.params.keyword}` : `/search/${route.params.keyword}/page/${newPage}`)
      } else {
        router.push(newPage === 1 ? '/' : `/page/${newPage}`)
      }
    } else {
      emit('page-change', newPage)
    }
  }
}

const handleFocus = (event) => { event.target.select() }

const handleJumpPage = () => {
  const targetPage = parseInt(currentPageInput.value)
  if (!isNaN(targetPage) && targetPage >= 1 && targetPage <= props.totalPages) {
    if (targetPage !== props.currentPage) handlePageChange(targetPage)
  } else {
    currentPageInput.value = props.currentPage.toString()
  }
}

const normalizePageSize = (value) => {
  if (isNaN(value) || value < 10) return 10
  if (value > 100) return 100
  return value
}

const updatePageSize = (value) => {
  const normalizedValue = normalizePageSize(value)
  pageSizeInput.value = normalizedValue.toString()
  emit('update:page-size', normalizedValue)
}

const handlePageSizeChange = () => {
  const value = parseInt(pageSizeInput.value, 10)
  if (!isNaN(value) && value >= 10 && value <= 100) emit('update:page-size', value)
}

const handlePageSizeBlur = () => {
  const value = parseInt(pageSizeInput.value, 10)
  updatePageSize(value)
}
</script>
