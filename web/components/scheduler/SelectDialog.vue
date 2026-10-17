<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="show" class="fixed inset-0 z-[10000] flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50" @click="$emit('update:show', false)" />
        <div
          class="relative flex max-h-[80vh] w-[90%] max-w-lg flex-col overflow-hidden rounded-xl bg-white shadow-2xl dark:bg-gray-800"
        >
          <div
            class="flex flex-shrink-0 items-center justify-between border-b border-gray-100 px-4 py-3 dark:border-gray-700"
          >
            <span class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ title }}</span>
            <button
              class="rounded-lg p-1 transition-colors hover:bg-gray-100 dark:hover:bg-gray-700"
              @click="$emit('update:show', false)"
            >
              <svg
                class="h-4 w-4 text-gray-500 dark:text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <div class="flex flex-1 flex-col overflow-hidden p-4">
            <div class="mb-4 flex flex-shrink-0 items-center space-x-2">
              <div class="relative flex-1">
                <input
                  v-model="searchQuery"
                  type="text"
                  :placeholder="searchPlaceholder"
                  class="w-full rounded-md border border-gray-300 bg-white py-1.5 pr-3 pl-8 text-sm text-gray-900 focus:border-accent focus:ring-1 focus:ring-accent focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
                />
                <svg
                  class="absolute top-2 left-2.5 h-4 w-4 text-gray-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                  />
                </svg>
                <button
                  v-if="searchQuery"
                  class="absolute top-2 right-2 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
                  @click="resetSearch"
                >
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>

              <select
                v-if="showMethodFilter"
                v-model="methodFilter"
                class="rounded-md border border-gray-300 bg-white py-1.5 pr-8 pl-2 text-sm text-gray-900 focus:border-accent focus:ring-1 focus:ring-accent focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
              >
                <option value="ALL">全部</option>
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
              </select>
            </div>

            <div class="custom-scrollbar flex-1 overflow-y-auto pr-2">
              <div v-for="(groupItems, group) in filteredGroupedItems" :key="group" class="mb-4">
                <div
                  class="mb-2 flex cursor-pointer items-center justify-between rounded-lg bg-gray-100 p-2 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700"
                  @click="toggleGroup(group)"
                >
                  <div class="flex items-center space-x-2">
                    <span class="text-sm font-medium text-gray-800 dark:text-gray-100">{{
                      group || '未分类'
                    }}</span>
                    <span class="text-xs text-gray-600 dark:text-gray-400"
                      >({{ groupItems.length }}个)</span
                    >
                  </div>
                  <svg
                    class="h-4 w-4 transform text-gray-600 transition-transform duration-200"
                    :class="{ 'rotate-90': expandedGroups[group] }"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>

                <div
                  v-show="expandedGroups[group]"
                  class="space-y-1 border-l-2 border-gray-100 pl-2 dark:border-gray-700"
                >
                  <div
                    v-for="item in groupItems"
                    :key="item.id"
                    class="cursor-pointer rounded-md p-2 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700"
                    :class="{ 'bg-accent/10 hover:bg-accent/20': isSelected(item) }"
                    @click="selectItem(item)"
                  >
                    <div class="flex items-start justify-between">
                      <div class="flex-1">
                        <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                          {{ item.name || '' }}
                        </div>
                        <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                          {{ item.description || '' }}
                        </div>
                      </div>
                      <div v-if="item.method" class="ml-3">
                        <span
                          class="rounded-full px-2 py-0.5 text-xs font-medium"
                          :class="{
                            'bg-blue-100 text-blue-800': item.method === 'GET',
                            'bg-green-100 text-green-800': item.method === 'POST',
                            'bg-yellow-100 text-yellow-800': item.method === 'PUT',
                            'bg-red-100 text-red-800': item.method === 'DELETE'
                          }"
                        >
                          {{ item.method }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue';

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    required: true
  },
  items: {
    type: Array,
    default: () => []
  },
  groupBy: {
    type: String,
    default: 'tags'
  },
  selected: {
    type: [String, Array],
    default: null
  },
  multiple: {
    type: Boolean,
    default: false
  },
  showMethodFilter: {
    type: Boolean,
    default: false
  },
  searchPlaceholder: {
    type: String,
    default: '搜索...'
  },
  descriptionField: {
    type: String,
    default: 'description'
  }
});

const emit = defineEmits(['update:show', 'select', 'update:selected']);

const searchQuery = ref('');
const methodFilter = ref('ALL');
const expandedGroups = ref({});
const selectedItems = ref(props.multiple ? [] : null);

watch(
  () => props.selected,
  (newVal) => {
    selectedItems.value = props.multiple ? (Array.isArray(newVal) ? newVal : []) : newVal;
  },
  { immediate: true }
);

const formatTaskItem = (task) => {
  const descField = props.descriptionField;
  const formattedTask = {
    id: task.id || task.operationId || task.path,
    path: task.path || '',
    name: task.name || task.summary || task.path,
    description: task[descField] || task.description || task.path,
    method: task.method || 'GET',
    tags: Array.isArray(task.tags) ? task.tags : [],
    enabled: task.enabled,
    execution: task.execution || {},
    config: task.config || {},
    parent_id: task.parent_id,
    sub_tasks: task.sub_tasks || [],
    sequence_number: task.sequence_number
  };
  return formattedTask;
};

const filteredGroupedItems = computed(() => {
  let filtered = props.items.map(formatTaskItem);

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      (item) =>
        (item.name && item.name.toLowerCase().includes(query)) ||
        (item.description && item.description.toLowerCase().includes(query)) ||
        (item.path && item.path.toLowerCase().includes(query))
    );
  }

  if (props.showMethodFilter && methodFilter.value !== 'ALL') {
    filtered = filtered.filter((item) => item.method === methodFilter.value);
  }

  const grouped = {};
  filtered.forEach((item) => {
    if (props.groupBy === 'tags' && Array.isArray(item.tags) && item.tags.length > 0) {
      item.tags.forEach((tag) => {
        if (!grouped[tag]) {
          grouped[tag] = [];
        }
        grouped[tag].push(item);
      });
    } else {
      const group = item[props.groupBy] || '未分类';
      if (!grouped[group]) {
        grouped[group] = [];
      }
      grouped[group].push(item);
    }
  });

  return grouped;
});

const toggleGroup = (group) => {
  expandedGroups.value[group] = !expandedGroups.value[group];
};

const resetSearch = () => {
  searchQuery.value = '';
  methodFilter.value = 'ALL';
  Object.keys(filteredGroupedItems.value).forEach((group) => {
    expandedGroups.value[group] = true;
  });
};

const selectItem = (item) => {
  if (props.multiple) {
    const index = selectedItems.value.indexOf(item.id);
    if (index === -1) {
      selectedItems.value.push(item.id);
    } else {
      selectedItems.value.splice(index, 1);
    }
  } else {
    selectedItems.value = item.id;
    emit('update:show', false);
  }
  emit('select', item);
  emit('update:selected', selectedItems.value);
};

const isSelected = (item) => {
  if (props.multiple) {
    return selectedItems.value.includes(item.id);
  }
  return selectedItems.value === item.id;
};

watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      resetSearch();
    }
  }
);
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: var(--accent) #f3f4f6;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: #f3f4f6;
  border-radius: 2px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: var(--accent);
  border-radius: 2px;
}
</style>
