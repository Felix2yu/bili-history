<template>
  <div class="h-screen overflow-hidden overflow-x-hidden">
    <!-- Main content area -->
    <main class="h-full overflow-y-auto overflow-x-hidden pb-16 md:pb-0">
      <slot />
    </main>

    <!-- Mobile bottom tab bar -->
    <MobileTabBar />

    <!-- Floating drawer sidebar (all screen sizes) -->
    <Teleport to="body">
      <!-- Backdrop -->
      <Transition name="fade">
        <div
          v-if="showSidebar"
          class="fixed inset-0 z-[55] bg-black/30 backdrop-blur-sm md:backdrop-blur"
          @click="showSidebar = false"
        />
      </Transition>

      <!-- Drawer -->
      <Transition name="slide-drawer">
        <div
          v-if="showSidebar"
          class="fixed top-0 left-0 bottom-0 z-[60] glass-sidebar w-64 flex flex-col"
          :class="{ 'pt-[env(safe-area-inset-top)]': true }"
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
provide('showSidebar', showSidebar)

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
