<template>
  <div class="flex h-screen overflow-hidden">
    <!-- Desktop Sidebar -->
    <Sidebar class="hidden md:flex" />

    <!-- Main content area -->
    <main class="flex-1 overflow-y-auto pb-16 md:pb-0">
      <router-view v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" />
        </Transition>
      </router-view>
    </main>

    <!-- Mobile bottom tab bar -->
    <MobileTabBar />
  </div>

  <!-- DataSyncManager modal - global level -->
  <DataSyncManager
    v-model:showModal="showDataSyncModal"
    :initialTab="currentSyncTab"
    @sync-complete="handleSyncComplete"
    @check-complete="handleCheckComplete"
  />
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import Sidebar from '../Sidebar.vue'
import MobileTabBar from './MobileTabBar.vue'
import DataSyncManager from '../DataSyncManager.vue'

const showDataSyncModal = ref(false)
const currentSyncTab = ref('integrity')

const handleSyncComplete = (result) => {
  console.log('同步完成:', result)
}

const handleCheckComplete = (result) => {
  console.log('完整性检查完成:', result)
}

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
