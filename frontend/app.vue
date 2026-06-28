<template>
  <div>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

<script setup>
import { useDarkMode } from '~/stores/darkMode'

const darkModeStore = useDarkMode()

onMounted(() => {
  darkModeStore.initDarkMode()
})
</script>
