<template>
  <Transition name="slide-up">
    <div
      v-if="needRefresh"
      class="fixed bottom-0 left-0 right-0 z-50 p-4 md:p-6"
    >
      <div class="mx-auto max-w-md rounded-2xl border border-white/10 bg-slate-800/90 p-4 shadow-2xl backdrop-blur-xl">
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-blue-500/20">
            <svg class="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
            </svg>
          </div>
          <div class="flex-1">
            <p class="text-sm font-medium text-slate-100">有新版本可用</p>
            <p class="text-xs text-slate-400">点击刷新以获取最新版本</p>
          </div>
          <button
            class="shrink-0 rounded-lg bg-blue-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-600"
            @click="updateApp"
          >
            刷新
          </button>
          <button
            class="shrink-0 rounded-lg p-2 text-slate-400 transition-colors hover:bg-slate-700 hover:text-slate-200"
            @click="dismiss"
          >
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
const { needRefresh, updateServiceWorker } = useNuxtApp().$pwa || {}

function updateApp() {
  updateServiceWorker?.()
}

function dismiss() {
  if (needRefresh) {
    needRefresh.value = false
  }
}
</script>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
