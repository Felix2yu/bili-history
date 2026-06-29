<script setup>
import { onMounted, onUnmounted } from 'vue'
import { ConfigProvider } from 'vant'
import 'vant/es/notify/style'
import 'vant/es/dialog/style'
import 'vant/es/config-provider/style'
import privacyManager from './utils/privacyManager'
import { useDarkMode } from './store/darkMode'

const { isDarkMode, initDarkMode } = useDarkMode()

const handlePrivacyModeChange = (isEnabled) => {
  console.log('隐私模式状态变化:', isEnabled)
}

onMounted(() => {
  initDarkMode()

  if (localStorage.getItem('apiKey')) {
    localStorage.removeItem('apiKey')
    console.log('已清理localStorage中的API密钥')
  }

  privacyManager.addListener(handlePrivacyModeChange)

  const privacyModeEnabled = privacyManager.isEnabled()
  if (privacyModeEnabled) {
    handlePrivacyModeChange(true)
  }
})
</script>

<template>
  <ConfigProvider :theme="isDarkMode ? 'dark' : 'light'">
    <div class="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 transition-colors duration-300">
      <router-view v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" />
        </Transition>
      </router-view>
    </div>
  </ConfigProvider>
</template>
