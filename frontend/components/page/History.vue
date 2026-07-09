<template>
  <div class="pb-20 md:pb-0">
    <Navbar
      v-if="currentContent === 'history' && !showRemarks"
      @refresh-data="refreshData"
      v-model:date="date"
      v-model:category="category"
      v-model:business="business"
      v-model:businessLabel="businessLabel"
      :total="total"
      @click-date="show = true"
      :layout="layout"
      @change-layout="layout = $event"
      :is-batch-mode="isBatchMode"
      :show-remarks="showRemarks"
      @toggle-batch-mode="isBatchMode = !isBatchMode"
      @toggle-remarks="showRemarks = !showRemarks"
    />

    <div>
      <div class="mx-auto max-w-7xl sm:px-2 lg:px-8">
        <!-- 日期日历 - 放在顶部 -->
        <div v-if="currentContent === 'history' && !showRemarks && availableDates.length > 0" class="mx-auto mb-4 max-w-4xl mt-4">
          <DatePagination
            :current-date="currentDate"
            :available-dates="availableDates"
            :record-count="recordCount"
            @date-change="handleDateChange"
          />
        </div>

        <div class="">
          <HistoryContent
            v-if="currentContent === 'history' && !showRemarks"
            ref="historyContentRef"
            :selected-year="selectedYear"
            :current-date="currentDate"
            :page="1"
            :page-size="9999"
            @update:total-pages="() => {}"
            @update:total="total = $event"
            @update:record-count="recordCount = $event"
            @update:date="date = $event"
            @update:category="category = $event"
            @update:category-type="categoryType = $event"
            v-model:show="show"
            v-model:showBottom="showBottom"
            :layout="layout"
            :date="date"
            :category="category"
            :category-type="categoryType"
            :business="business"
            :is-batch-mode="isBatchMode"
          />

          <Remarks v-else-if="showRemarks" />
          <Settings v-else-if="currentContent === 'settings'" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Navbar from '../Navbar.vue'
import HistoryContent from '../HistoryContent.vue'
import DatePagination from '../DatePagination.vue'
import Settings from '../Settings.vue'
import Remarks from './Remarks.vue'
import { getHistoryDates } from '~/utils/api'

const props = defineProps({
  defaultShowRemarks: {
    type: Boolean,
    default: false
  }
})

const currentContent = ref('history')
const router = useRouter()
const route = useRoute()

function todayStr() {
  const now = new Date()
  return `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}`
}

const currentDate = ref(route.params.date || todayStr())
const selectedYear = ref(new Date().getFullYear())
const show = ref(false)
const showBottom = ref(false)
const date = ref('')
const total = ref(0)
const recordCount = ref(0)
const category = ref('')
const categoryType = ref('')
const layout = ref(localStorage.getItem('defaultLayout') || 'grid')
const isBatchMode = ref(false)
const showRemarks = ref(props.defaultShowRemarks)
const business = ref('')
const businessLabel = ref('')
const availableDates = ref([])

const historyContentRef = ref(null)

const refreshData = async () => {
  try {
    if (historyContentRef.value?.fetchHistoryByDateRange) {
      await historyContentRef.value.fetchHistoryByDateRange()
    }
    loadAvailableDates()
  } catch (error) {
    console.error('刷新数据失败:', error)
  }
}

const loadAvailableDates = async () => {
  try {
    const res = await getHistoryDates()
    if (res.data && res.data.status === 'success') {
      availableDates.value = res.data.data || []
    }
  } catch (e) {
    console.error('加载日期列表失败:', e)
  }
}

const handleDateChange = (newDate) => {
  router.push(`/date/${newDate}`)
}

watch(
  () => route.path,
  (path) => {
    if (path === '/settings') {
      currentContent.value = 'settings'
      showRemarks.value = false
    } else if (path === '/remarks') {
      currentContent.value = 'history'
      showRemarks.value = true
    } else {
      currentContent.value = 'history'
      showRemarks.value = false
    }
  },
  { immediate: true }
)

watch(
  [() => route.params.date, () => route.path],
  ([newDate, path], [oldDate, oldPath]) => {
    if (path === '/') {
      const today = todayStr()
      if (currentDate.value !== today) {
        currentDate.value = today
        router.replace(`/date/${today}`)
        return
      }
    } else if (newDate && newDate !== currentDate.value) {
      currentDate.value = newDate
    }

    // 路由变化时由 HistoryContent 的 watch(currentDate) 触发刷新
    // 仅在组件已挂载后且非首次加载时手动刷新
    if (historyContentRef.value?.fetchHistoryByDateRange && oldPath !== undefined) {
      nextTick(() => {
        historyContentRef.value?.fetchHistoryByDateRange()
      })
    }
  },
  { immediate: true }
)

onMounted(() => {
  if (props.defaultShowRemarks || route.path === '/remarks') {
    showRemarks.value = true
  }
  loadAvailableDates()
  window.addEventListener('layout-setting-changed', handleLayoutSettingChanged)
})

onUnmounted(() => {
  window.removeEventListener('layout-setting-changed', handleLayoutSettingChanged)
})

const handleLayoutSettingChanged = (event) => {
  if (event.detail && typeof event.detail.layout === 'string') {
    layout.value = event.detail.layout
  }
}

watch(layout, (newLayout) => {
  localStorage.setItem('defaultLayout', newLayout)
  try {
    window.dispatchEvent(new CustomEvent('layout-changed', { detail: { layout: newLayout } }))
  } catch (error) {
    console.error('触发布局变更事件失败:', error)
  }
})
</script>

<style scoped>
@keyframes bounce-x {
  0%, 100% { transform: translateX(0); }
  50% { transform: translateX(4px); }
}
.animate-bounce-x { animation: bounce-x 1.5s infinite; }
</style>
