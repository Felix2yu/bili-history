<template>
  <div class="border-b border-gray-200 px-4 py-3 dark:border-gray-700">
    <div class="flex flex-wrap items-center gap-3">
      <div v-if="sortOptions && sortOptions.length > 0" class="flex items-center space-x-2">
        <span class="text-xs text-gray-500 dark:text-gray-400">排序:</span>
        <button
          v-for="opt in sortOptions"
          :key="opt.key"
          class="rounded-md px-2 py-1 text-xs transition-colors"
          :class="
            sortKey === opt.key
              ? 'bg-accent/10 font-medium text-accent'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
          "
          @click="handleToggleSort(opt.key)"
        >
          {{ opt.label }}
          <span v-if="sortKey === opt.key" class="ml-0.5">{{
            sortOrder === 'desc' ? '↓' : '↑'
          }}</span>
        </button>
      </div>

      <div
        v-if="showCategoryFilter && allCategories.length > 0"
        class="h-4 w-px bg-gray-200 dark:bg-gray-700"
      />

      <div
        v-if="showCategoryFilter && allCategories.length > 0"
        class="flex flex-wrap items-center space-x-2"
      >
        <span class="text-xs text-gray-500 dark:text-gray-400">分区:</span>
        <button
          class="rounded-md px-2 py-1 text-xs transition-colors"
          :class="
            selectedCategory === ''
              ? 'bg-accent/10 font-medium text-accent'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
          "
          @click="handleSelectCategory('')"
        >
          全部
        </button>
        <button
          v-for="cat in topCategories"
          :key="cat.tname"
          class="rounded-md px-2 py-1 text-xs transition-colors"
          :class="
            selectedCategory === cat.tname
              ? 'bg-accent/10 font-medium text-accent'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
          "
          @click="handleSelectCategory(cat.tname)"
        >
          {{ cat.tname }} ({{ cat.count }})
        </button>
        <div
          v-if="topCategories.length < allCategories.length"
          ref="catDropdownRef"
          class="relative"
        >
          <button
            class="rounded-md px-2 py-1 text-xs text-gray-600 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
            @click.stop="showCatDropdown = !showCatDropdown"
          >
            更多...
          </button>
          <div
            v-if="showCatDropdown"
            class="absolute top-full left-0 z-10 mt-1 max-h-60 min-w-[180px] overflow-y-auto rounded-lg border border-gray-200 bg-white p-2 shadow-lg dark:border-gray-700 dark:bg-gray-800"
          >
            <button
              v-for="cat in restCategories"
              :key="cat.tname"
              class="w-full rounded px-2 py-1 text-left text-xs transition-colors"
              :class="
                selectedCategory === cat.tname
                  ? 'bg-accent/10 text-accent'
                  : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
              "
              @click="
                handleSelectCategory(cat.tname);
                showCatDropdown = false;
              "
            >
              {{ cat.tname }} ({{ cat.count }})
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="showOwnerFilter && allOwners.length > 0"
        class="h-4 w-px bg-gray-200 dark:bg-gray-700"
      />

      <div
        v-if="showOwnerFilter && allOwners.length > 0"
        class="flex flex-wrap items-center space-x-2"
      >
        <span class="text-xs text-gray-500 dark:text-gray-400">UP主:</span>
        <button
          class="rounded-md px-2 py-1 text-xs transition-colors"
          :class="
            selectedOwner === ''
              ? 'bg-accent/10 font-medium text-accent'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
          "
          @click="handleSelectOwner('')"
        >
          全部
        </button>
        <button
          v-for="owner in topOwners"
          :key="owner.name"
          class="max-w-[120px] truncate rounded-md px-2 py-1 text-xs transition-colors"
          :class="
            selectedOwner === owner.name
              ? 'bg-accent/10 font-medium text-accent'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
          "
          :title="owner.name"
          @click="handleSelectOwner(owner.name)"
        >
          {{ owner.name }} ({{ owner.count }})
        </button>
        <div v-if="topOwners.length < allOwners.length" ref="ownerDropdownRef" class="relative">
          <button
            class="rounded-md px-2 py-1 text-xs text-gray-600 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700"
            @click.stop="showOwnerDropdown = !showOwnerDropdown"
          >
            更多...
          </button>
          <div
            v-if="showOwnerDropdown"
            class="absolute top-full left-0 z-10 mt-1 max-h-60 min-w-[180px] overflow-y-auto rounded-lg border border-gray-200 bg-white p-2 shadow-lg dark:border-gray-700 dark:bg-gray-800"
          >
            <button
              v-for="owner in restOwners"
              :key="owner.name"
              class="w-full rounded px-2 py-1 text-left text-xs transition-colors"
              :class="
                selectedOwner === owner.name
                  ? 'bg-accent/10 text-accent'
                  : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
              "
              @click="
                handleSelectOwner(owner.name);
                showOwnerDropdown = false;
              "
            >
              {{ owner.name }} ({{ owner.count }})
            </button>
          </div>
        </div>
      </div>

      <slot name="extra" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';

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
  topCount: { type: Number, default: 10 }
});

const emit = defineEmits([
  'update:sortKey',
  'update:sortOrder',
  'update:selectedCategory',
  'update:selectedOwner',
  'sort-change',
  'category-change',
  'owner-change'
]);

const showCatDropdown = ref(false);
const showOwnerDropdown = ref(false);
const catDropdownRef = ref(null);
const ownerDropdownRef = ref(null);

const topCategories = computed(() => props.allCategories.slice(0, props.topCount));
const restCategories = computed(() => props.allCategories.slice(props.topCount));
const topOwners = computed(() => props.allOwners.slice(0, props.topCount));
const restOwners = computed(() => props.allOwners.slice(props.topCount));

const handleToggleSort = (key) => {
  let newOrder = 'desc';
  if (props.sortKey === key) {
    newOrder = props.sortOrder === 'desc' ? 'asc' : 'desc';
  } else {
    newOrder = key === 'owner_name' ? 'asc' : 'desc';
  }
  emit('update:sortKey', key);
  emit('update:sortOrder', newOrder);
  emit('sort-change', { key, order: newOrder });
};

const handleSelectCategory = (val) => {
  emit('update:selectedCategory', val);
  emit('category-change', val);
};

const handleSelectOwner = (val) => {
  emit('update:selectedOwner', val);
  emit('owner-change', val);
};

const handleClickOutside = (e) => {
  if (ownerDropdownRef.value && !ownerDropdownRef.value.contains(e.target)) {
    showOwnerDropdown.value = false;
  }
  if (catDropdownRef.value && !catDropdownRef.value.contains(e.target)) {
    showCatDropdown.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>
