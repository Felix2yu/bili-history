<template>
  <div class="flex h-screen overflow-hidden">
    <!-- Desktop Sidebar (permanent) -->
    <aside class="hidden md:flex glass-sidebar flex-shrink-0 flex-col h-full w-60 z-40">
      <Sidebar />
    </aside>

    <!-- Main content area -->
    <main class="flex-1 h-full overflow-y-auto overflow-x-hidden pb-16 md:pb-0">
      <slot />
    </main>

    <!-- Mobile bottom tab bar -->
    <MobileTabBar />

    <!-- Mobile drawer sidebar -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showSidebar"
          class="fixed inset-0 z-[55] bg-black/30 backdrop-blur-sm md:hidden"
          @click="showSidebar = false"
        />
      </Transition>
      <Transition name="slide-drawer">
        <div
          v-if="showSidebar"
          class="fixed top-0 left-0 bottom-0 z-[60] glass-sidebar w-64 flex flex-col md:hidden"
        >
          <Sidebar @navigate="showSidebar = false" />
        </div>
      </Transition>
    </Teleport>

    <!-- DataSyncManager modal -->
    <DataSyncManager
      v-model:showModal="showDataSyncModal"
      :initialTab="currentSyncTab"
      @sync-complete="handleSyncComplete"
      @check-complete="handleCheckComplete"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, provide } from 'vue'
import Sidebar from '../Sidebar.vue'
import MobileTabBar from './MobileTabBar.vue'
import DataSyncManager from '../DataSyncManager.vue'

const showSidebar = ref(false)
provide('toggleSidebar', () => { showSidebar.value = !showSidebar.value })

const showDataSyncModal = ref(false)
const currentSyncTab = ref('integrity')

const handleSyncComplete = (result) => { console.log('同步完成:', result) }
const handleCheckComplete = (result) => { console.log('完整性检查完成:', result) }

onMounted(() => {
  window.addEventListener('open-data-sync-manager', handleOpenDataSyncManager)
})
onUnmounted(() => {
  window.removeEventListener('open-data-sync-manager', handleOpenDataSyncManager)
})

const handleOpenDataSyncManager = (event) => {
  if (event.detail && event.detail.tab) {
    currentSyncTab.value = event.detail.tab
  }
  showDataSyncModal.value = true
}
</script>

<style scoped>
.slide-drawer-enter-active,
.slide-drawer-leave-active {
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.slide-drawer-enter-from,
.slide-drawer-leave-to {
  transform: translateX(-100%);
}
</style>
