<template>
  <div class="relative mx-auto max-w-xl">
    <div class="flex w-full h-9 items-center rounded-full glass-input pl-3 gap-2">
      <!-- Search icon -->
      <svg class="w-4 h-4 text-gray-400 dark:text-gray-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>

      <!-- Search type selector -->
      <div class="relative flex-shrink-0">
        <button
          @click="showTypeDropdown = !showTypeDropdown"
          class="flex items-center gap-1 text-accent text-xs font-medium whitespace-nowrap hover:bg-accent/10 px-2 py-1 rounded-lg transition-colors"
        >
          {{ getTypeLabel(searchType) }}
          <svg class="w-3 h-3 transition-transform" :class="{ 'rotate-180': showTypeDropdown }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <!-- Dropdown -->
        <Transition name="slide-down">
          <div
            v-if="showTypeDropdown"
            class="absolute top-full left-0 mt-1 glass-card rounded-xl py-1 min-w-[100px] z-50"
          >
            <button
              v-for="option in searchTypeOptions"
              :key="option.value"
              @click="onSearchTypeChange(option.value)"
              class="w-full text-left px-3 py-2 text-sm transition-colors hover:bg-accent/10"
              :class="searchType === option.value ? 'text-accent font-medium' : 'text-gray-700 dark:text-gray-300'"
            >
              {{ option.label }}
            </button>
          </div>
        </Transition>
      </div>

      <!-- Divider -->
      <div class="w-px h-4 bg-gray-200 dark:bg-gray-600 flex-shrink-0"></div>

      <!-- Input -->
      <input
        ref="inputRef"
        v-model="searchQuery"
        @keyup.enter="handleSearch"
        @focus="showTypeDropdown = false"
        type="search"
        :placeholder="getPlaceholder"
        class="flex-1 min-w-0 h-full border-none bg-transparent text-sm text-gray-700 dark:text-gray-200 focus:outline-none placeholder-gray-400 dark:placeholder-gray-500"
      />
    </div>

    <!-- Click outside to close dropdown -->
    <div v-if="showTypeDropdown" class="fixed inset-0 z-40" @click="showTypeDropdown = false"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getAvailableYears } from '../../api/api.js'

const props = defineProps({
  initialYear: { type: Number, default: () => new Date().getFullYear() },
  initialKeyword: { type: String, default: '' },
  initialSearchType: { type: String, default: 'all' }
})

const emit = defineEmits(['year-change', 'search'])
const router = useRouter()
const route = useRoute()

const searchQuery = ref(props.initialKeyword)
const searchType = ref(props.initialSearchType)
const showTypeDropdown = ref(false)
const inputRef = ref(null)

const searchTypeOptions = [
  { value: 'all', label: '全部' },
  { value: 'title', label: '标题' },
  { value: 'author', label: 'UP主' },
  { value: 'tag', label: '分区' },
  { value: 'remark', label: '备注' }
]

const getTypeLabel = (value) => {
  const option = searchTypeOptions.find(opt => opt.value === value)
  return option ? option.label : '全部'
}

const onSearchTypeChange = (value) => {
  searchType.value = value
  showTypeDropdown.value = false
  if (route.name === 'Search' && searchQuery.value.trim()) {
    handleSearch()
  }
}

const selectedYear = ref(props.initialYear)
const availableYears = ref([props.initialYear])

const getPlaceholder = computed(() => {
  switch (searchType.value) {
    case 'title': return '视频标题/oid'
    case 'author': return 'UP主名称'
    case 'tag': return '分区名称'
    case 'remark': return '备注内容'
    default: return '输入关键词搜索'
  }
})

const fetchAvailableYears = async () => {
  try {
    const response = await getAvailableYears()
    if (response.data.status === 'success') {
      availableYears.value = response.data.data
      if (!availableYears.value.includes(selectedYear.value)) {
        selectedYear.value = availableYears.value[0]
        emit('year-change', selectedYear.value)
      }
    }
  } catch (error) {
    console.error('获取可用年份失败:', error)
  }
}

const handleSearch = () => {
  if (searchQuery.value.trim()) {
    if (route.name === 'Search') {
      emit('search', { keyword: searchQuery.value.trim(), type: searchType.value })
    } else {
      router.push({
        name: 'Search',
        params: { keyword: searchQuery.value.trim() },
        query: { type: searchType.value }
      })
      searchQuery.value = ''
    }
  }
}

watch(() => props.initialKeyword, (newKeyword) => { searchQuery.value = newKeyword })
watch(() => props.initialYear, (newYear) => { if (newYear !== selectedYear.value) selectedYear.value = newYear })
watch(() => props.initialSearchType, (newType) => { if (newType !== searchType.value) searchType.value = newType })

onMounted(async () => { await fetchAvailableYears() })
</script>

<style scoped>
input[type="search"]::-webkit-search-decoration,
input[type="search"]::-webkit-search-cancel-button,
input[type="search"]::-webkit-search-results-button,
input[type="search"]::-webkit-search-results-decoration {
  display: none;
}
input:focus { box-shadow: none !important; outline: none !important; }
</style>
