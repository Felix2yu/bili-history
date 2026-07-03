<template>
  <div class="border-b border-gray-200 dark:border-gray-700 px-4 py-3">
    <div class="flex flex-wrap items-center gap-3">
      <div v-if="sortOptions && sortOptions.length > 0" class="flex items-center space-x-2">
        <span class="text-xs text-gray-500 dark:text-gray-400">排序:</span>
        <button
          v-for="opt in sortOptions"
          :key="opt.key"
          @click="handleToggleSort(opt.key)"
          class="px-2 py-1 text-xs rounded-md transition-colors"
          :class="sortKey === opt.key
            ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
        >
          {{ opt.label }}
          <span v-if="sortKey === opt.key" class="ml-0.5">{{ sortOrder === 'desc' ? '↓' : '↑' }}</span>
        </button>
      </div>

      <div v-if="showCategoryFilter && allCategories.length > 0" class="w-px h-4 bg-gray-200 dark:bg-gray-700"></div>

      <div v-if="showCategoryFilter && allCategories.length > 0" class="flex items-center space-x-2 flex-wrap">
        <span class="text-xs text-gray-500 dark:text-gray-400">分区:</span>
        <button
          @click="handleSelectCategory('')"
          class="px-2 py-1 text-xs rounded-md transition-colors"
          :class="selectedCategory === ''
            ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
        >
          全部
        </button>
        <button
          v-for="cat in topCategories"
          :key="cat.tname"
          @click="handleSelectCategory(cat.tname)"
          class="px-2 py-1 text-xs rounded-md transition-colors"
          :class="selectedCategory === cat.tname
            ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
        >
          {{ cat.tname }} ({{ cat.count }})
        </button>
        <div v-if="topCategories.length < allCategories.length" class="relative" ref="catDropdownRef">
          <button
            @click.stop="showCatDropdown = !showCatDropdown"
            class="px-2 py-1 text-xs rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          >
            更多...
          </button>
          <div
            v-if="showCatDropdown"
            class="absolute top-full left-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 p-2 max-h-60 overflow-y-auto min-w-[180px]"
          >
            <button
              v-for="cat in restCategories"
              :key="cat.tname"
              @click="handleSelectCategory(cat.tname); showCatDropdown = false"
              class="w-full text-left px-2 py-1 text-xs rounded transition-colors"
              :class="selectedCategory === cat.tname
                ? 'bg-[#fb7299]/10 text-[#fb7299]'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
            >
              {{ cat.tname }} ({{ cat.count }})
            </button>
          </div>
        </div>
      </div>

      <div v-if="showOwnerFilter && allOwners.length > 0" class="w-px h-4 bg-gray-200 dark:bg-gray-700"></div>

      <div v-if="showOwnerFilter && allOwners.length > 0" class="flex items-center space-x-2 flex-wrap">
        <span class="text-xs text-gray-500 dark:text-gray-400">UP主:</span>
        <button
          @click="handleSelectOwner('')"
          class="px-2 py-1 text-xs rounded-md transition-colors"
          :class="selectedOwner === ''
            ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
        >
          全部
        </button>
        <button
          v-for="owner in topOwners"
          :key="owner.name"
          @click="handleSelectOwner(owner.name)"
          class="px-2 py-1 text-xs rounded-md transition-colors max-w-[120px] truncate"
          :class="selectedOwner === owner.name
            ? 'bg-[#fb7299]/10 text-[#fb7299] font-medium'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
          :title="owner.name"
        >
          {{ owner.name }} ({{ owner.count }})
        </button>
        <div v-if="topOwners.length < allOwners.length" class="relative" ref="ownerDropdownRef">
          <button
            @click.stop="showOwnerDropdown = !showOwnerDropdown"
            class="px-2 py-1 text-xs rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          >
            更多...
          </button>
          <div
            v-if="showOwnerDropdown"
            class="absolute top-full left-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 p-2 max-h-60 overflow-y-auto min-w-[180px]"
          >
            <button
              v-for="owner in restOwners"
              :key="owner.name"
              @click="handleSelectOwner(owner.name); showOwnerDropdown = false"
              class="w-full text-left px-2 py-1 text-xs rounded transition-colors"
              :class="selectedOwner === owner.name
                ? 'bg-[#fb7299]/10 text-[#fb7299]'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'"
            >
              {{ owner.name }} ({{ owner.count }})
            </button>
          </div>
        </div>
      </div>

      <slot name="extra"></slot>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  sortOptions: { type: Array, default: () => [] },
  sortKey: { type: String, default: '' },
  sortOrder: { type: String, default: 'desc' },
  showCategoryFilter: { type: Boolean, default: true },
  showOwnerFilter: { type: Boolean, default: true },
  selectedCategory: { type: String, default: '' },
  selectedOwner: { type: String, default: '' },
  allCategories: { type: Array, default: () => [] },
  allOwners: { type: Array, default: () => [] },
  topCount: { type: Number, default: 10 },
})

const emit = defineEmits([
  'update:sortKey',
  'update:sortOrder',
  'update:selectedCategory',
  'update:selectedOwner',
  'sort-change',
  'category-change',
  'owner-change',
])

const showCatDropdown = ref(false)
const showOwnerDropdown = ref(false)
const catDropdownRef = ref(null)
const ownerDropdownRef = ref(null)

const topCategories = computed(() => props.allCategories.slice(0, props.topCount))
const restCategories = computed(() => props.allCategories.slice(props.topCount))
const topOwners = computed(() => props.allOwners.slice(0, props.topCount))
const restOwners = computed(() => props.allOwners.slice(props.topCount))

const handleToggleSort = (key) => {
  let newOrder = 'desc'
  if (props.sortKey === key) {
    newOrder = props.sortOrder === 'desc' ? 'asc' : 'desc'
  } else {
    newOrder = key === 'owner_name' ? 'asc' : 'desc'
  }
  emit('update:sortKey', key)
  emit('update:sortOrder', newOrder)
  emit('sort-change', { key, order: newOrder })
}

const handleSelectCategory = (val) => {
  emit('update:selectedCategory', val)
  emit('category-change', val)
}

const handleSelectOwner = (val) => {
  emit('update:selectedOwner', val)
  emit('owner-change', val)
}

const handleClickOutside = (e) => {
  if (ownerDropdownRef.value && !ownerDropdownRef.value.contains(e.target)) {
    showOwnerDropdown.value = false
  }
  if (catDropdownRef.value && !catDropdownRef.value.contains(e.target)) {
    showCatDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
