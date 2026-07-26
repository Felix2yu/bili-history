<template>
  <div class="px-5 md:px-6 py-4 border-b border-gray-50 dark:border-gray-700/30 last:border-b-0 transition-colors duration-200 hover:bg-gray-50/50 dark:hover:bg-gray-700/20">
    <div class="flex items-start justify-between gap-4">
      <div class="min-w-0 flex-1">
        <h3 class="text-[0.8125rem] md:text-sm font-medium text-gray-900 dark:text-gray-100">{{ label }}</h3>
        <p v-if="description" class="text-[0.6875rem] text-gray-500 dark:text-gray-400 mt-0.5 md:text-xs">{{ description }}</p>
      </div>
      <label class="relative inline-flex shrink-0 cursor-pointer items-center" :class="{ 'cursor-not-allowed opacity-60': disabled }">
        <input
          type="checkbox"
          :checked="modelValue"
          class="peer sr-only"
          :disabled="disabled"
          @change="$emit('update:modelValue', $event.target.checked)"
        >
        <div class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:translate-x-0 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-accent peer-checked:after:translate-x-full peer-checked:after:border-white peer-focus:ring-4 peer-focus:ring-accent/20 dark:bg-gray-600"></div>
      </label>
    </div>
  </div>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  description: { type: String, default: '' },
  modelValue: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false }
})

defineEmits(['update:modelValue'])
</script>
