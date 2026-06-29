<template>
  <div class="overflow-x-hidden">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <PwaUpdatePrompt />
    <PwaInstallPrompt />
  </div>
</template>

<script setup>
import { useDarkMode } from '~/stores/darkMode'

const darkModeStore = useDarkMode()

onMounted(() => {
  darkModeStore.initDarkMode()
})
</script>
