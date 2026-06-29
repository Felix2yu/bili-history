<template>
  <div class="sticky top-0 z-40 pt-[env(safe-area-inset-top)]">
    <nav class="glass border-b border-glass-border">
      <div class="mx-auto transition-all duration-300 ease-in-out" :class="{'max-w-4xl': layout === 'list', 'max-w-6xl': layout === 'grid'}">
        <div class="flex items-center justify-between px-3 py-2.5 gap-3">
          <!-- Left: action buttons -->
          <div class="flex items-center gap-1.5">
            <!-- Hamburger menu (mobile only) -->
            <button @click="toggleSidebar" class="glass-icon-btn md:hidden" title="菜单">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <!-- Refresh button -->
            <button
              @click="handleUpdate"
              :disabled="isUpdating"
              class="glass-icon-btn"
              :class="{ 'active': isUpdating }"
              :title="syncDeleted ? '当前模式：同步已删除记录' : '当前模式：不同步已删除记录'"
            >
              <svg v-if="isUpdating" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>

            <!-- Dark mode toggle -->
            <button @click="toggleDarkMode" class="glass-icon-btn" :title="isDarkMode ? '浅色模式' : '深色模式'">
              <svg v-if="isDarkMode" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
              </svg>
            </button>

            <!-- Privacy mode -->
            <button @click="togglePrivacyMode" class="glass-icon-btn" :class="{ 'active': isPrivacyMode }" :title="isPrivacyMode ? '关闭隐私' : '隐私模式'">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" :stroke="isPrivacyMode ? '#fb7299' : 'currentColor'" stroke-width="2">
                <path v-if="!isPrivacyMode" stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path v-if="!isPrivacyMode" stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                <path v-if="isPrivacyMode" stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
              </svg>
            </button>
          </div>

          <!-- Center: SearchBar -->
          <div class="flex-1 max-w-xl mx-2">
            <SearchBar />
          </div>

          <!-- Right: layout/filter/batch -->
          <div class="flex items-center gap-1.5">
            <!-- Layout toggle (desktop only) -->
            <button
              @click="$emit('change-layout', layout === 'list' ? 'grid' : 'list')"
              class="hidden sm:flex glass-icon-btn"
              :class="{ 'active': layout === 'grid' }"
              :title="layout === 'list' ? '网格视图' : '列表视图'"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path v-if="layout === 'list'" stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
                <path v-else stroke-linecap="round" stroke-linejoin="round" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
            </button>

            <!-- Filter button -->
            <button
              @click="openFilterPanel"
              class="glass-icon-btn"
              :class="{ 'active': isFilterActive }"
              :title="isFilterActive ? '筛选中' : '筛选'"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
              </svg>
            </button>

            <!-- Batch mode -->
            <button
              @click="$emit('toggle-batch-mode')"
              class="glass-icon-btn"
              :class="{ 'active': isBatchMode }"
              :title="isBatchMode ? '点击取消' : '批量操作'"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" :stroke="isBatchMode ? '#fb7299' : 'currentColor'" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
              </svg>
            </button>

            <!-- Settings (mobile only) -->
            <button
              @click="$router.push('/settings')"
              class="flex sm:hidden glass-icon-btn"
              title="设置"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Filter area -->
        <div class="mx-auto transition-all duration-300 ease-in-out" :class="{'max-w-4xl': layout === 'list', 'max-w-6xl': layout === 'grid'}">
          <FilterDropdown
            ref="filterDropdownRef"
            :business="business"
            :businessLabel="businessLabel"
            :date="date"
            :category="category"
            :total="total"
            @update:business="$emit('update:business', $event)"
            @update:businessLabel="$emit('update:businessLabel', $event)"
            @update:date="$emit('update:date', $event)"
            @update:category="$emit('update:category', $event)"
            @click-date="$emit('click-date')"
            @click-category="$emit('click-category')"
            @refresh-data="$emit('refresh-data')"
          />
        </div>
      </div>
    </nav>
  </div>
</template>

<script setup>
import SearchBar from './SearchBar.vue'
import FilterDropdown from './FilterDropdown.vue'
import { ref, watch, computed, inject } from 'vue'
import { showNotify } from 'vant'
import { usePrivacyStore } from '~/stores/privacy.js'
import { useDarkMode } from '~/stores/darkMode.js'
import 'vant/es/notify/style'

const toggleSidebar = inject('toggleSidebar', () => {})

const { isPrivacyMode, togglePrivacyMode } = usePrivacyStore()
const { isDarkMode, toggleDarkMode } = useDarkMode()

const props = defineProps({
  date: { type: String, default: '' },
  total: { type: Number, default: 0 },
  category: { type: String, default: '' },
  layout: { type: String, default: 'list' },
  isBatchMode: { type: Boolean, default: false },
  businessLabel: { type: String, default: '' },
  business: { type: String, default: '' }
})

const emit = defineEmits([
  'click-date', 'click-category', 'click-business',
  'change-layout', 'update:date', 'update:category',
  'update:business', 'update:businessLabel', 'refresh-data', 'toggle-batch-mode'
])

const isUpdating = ref(false)
const syncDeleted = ref(localStorage.getItem('syncDeleted') === 'true')
const filterDropdownRef = ref(null)

const isFilterActive = computed(() => Boolean(props.date || props.category || props.business))

watch(() => localStorage.getItem('syncDeleted'), (newVal) => {
  syncDeleted.value = newVal === 'true'
})

const openFilterPanel = () => {
  if (filterDropdownRef.value?.openFilterPopup) {
    filterDropdownRef.value.openFilterPopup()
  }
}

const handleUpdate = async () => {
  if (isUpdating.value) return
  isUpdating.value = true
  try {
    emit('refresh-data')
  } catch (error) {
    console.error('更新失败:', error)
    showNotify({
      type: 'danger',
      message: error.response?.data?.message || error.message || '更新失败，请稍后重试',
      duration: 3500,
    })
  } finally {
    setTimeout(() => { isUpdating.value = false }, 2000)
  }
}
</script>
