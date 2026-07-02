<template>
  <div class="p-3 transition-colors duration-200 hover:bg-gray-50 dark:hover:bg-gray-700 md:p-4">
    <div class="flex items-center justify-between mb-2">
      <h3 class="text-[14px] font-medium text-gray-900 dark:text-gray-100 md:text-base">Shoutrrr 通知配置</h3>
      <div class="flex space-x-2">
        <button
          @click="resetConfig"
          class="inline-flex items-center px-3 py-1.5 text-[11px] font-medium text-[#fb7299] md:text-sm bg-[#fb7299]/5 dark:bg-[#fb7299]/10 rounded-lg hover:bg-[#fb7299]/10 dark:hover:bg-[#fb7299]/20"
        >
          <svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          重置
        </button>
        <button
          @click="saveConfig"
          class="inline-flex items-center px-3 py-1.5 text-[11px] font-medium text-white md:text-sm bg-[#fb7299] rounded-lg hover:bg-[#fb7299]/90"
        >保存</button>
        <button
          @click="testPush"
          class="inline-flex items-center px-3 py-1.5 text-[11px] font-medium text-white md:text-sm bg-[#fb7299] rounded-lg hover:bg-[#fb7299]/90"
          :disabled="!config.enabled || urlList.length === 0"
        >测试</button>
      </div>
    </div>

    <div class="mt-2 space-y-3">
      <div class="flex items-center justify-between">
        <div>
          <label class="text-[12px] font-medium text-gray-700 dark:text-gray-300 md:text-sm">启用通知</label>
          <p class="text-[10px] text-gray-500 dark:text-gray-400 md:text-xs">关闭后所有Shoutrrr通知将不会发送</p>
        </div>
        <label class="relative inline-flex shrink-0 cursor-pointer items-center">
          <input type="checkbox" v-model="config.enabled" class="peer sr-only">
          <div class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:translate-x-0 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-[#fb7299] peer-checked:after:translate-x-full peer-checked:after:border-white peer-focus:ring-4 peer-focus:ring-[#fb7299]/20 dark:bg-gray-600"></div>
        </label>
      </div>

      <div>
        <label class="block text-[12px] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-1">通知地址列表</label>
        <textarea
          v-model="config.urls"
          rows="4"
          class="block w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 shadow-sm focus:border-[#fb7299] focus:ring-[#fb7299] text-[11px] md:text-sm"
          placeholder="每行一个地址，例如：&#10;bark://api.day.app/your-key&#10;tgram://bot-token/chat-id/&#10;smtp://user:pass@host:port/from@example.com/to@example.com&#10;discord://token@id"
        ></textarea>
      </div>

      <p class="text-[10px] text-gray-500 dark:text-gray-400 md:text-xs">
        完整服务列表请查看
        <a href="https://containrrr.dev/shoutrrr/services/overview/" target="_blank" rel="noopener noreferrer" class="text-[#fb7299] hover:underline">Shoutrrr 支持的服务</a>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showNotify } from 'vant'
import { getShoutrrrConfig, updateShoutrrrConfig, testShoutrrrPush } from '~/utils/api'

const config = ref({
  enabled: false,
  urls: ''
})

const urlList = computed(() => {
  if (!config.value.urls) return []
  return config.value.urls.split('\n').filter(u => u.trim())
})

const loadConfig = async () => {
  try {
    const response = await getShoutrrrConfig()
    if (response.data) {
      const urls = Array.isArray(response.data.urls) ? response.data.urls.join('\n') : (response.data.urls || '')
      config.value = {
        enabled: response.data.enabled ?? false,
        urls
      }
    }
  } catch (error) {
    console.error('获取Shoutrrr配置失败:', error)
  }
}

const saveConfig = async () => {
  try {
    const urls = urlList.value
    const response = await updateShoutrrrConfig({
      enabled: config.value.enabled,
      urls
    })
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: 'Shoutrrr配置已保存' })
    }
  } catch (error) {
    showNotify({ type: 'danger', message: `保存失败：${error.response?.data?.detail || error.message}` })
    throw error
  }
}

const resetConfig = () => {
  config.value = { enabled: false, urls: '' }
  saveConfig()
}

const testPush = async () => {
  if (!config.value.enabled || urlList.value.length === 0) {
    showNotify({ type: 'warning', message: '请先启用通知并填写推送地址' })
    return
  }
  try {
    showNotify({ type: 'primary', message: '正在发送测试通知...' })
    const response = await testShoutrrrPush()
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: response.data.message || '测试通知已发送' })
    } else {
      showNotify({ type: 'danger', message: response.data.message || '测试通知发送失败' })
    }
  } catch (error) {
    showNotify({ type: 'danger', message: `发送失败：${error.response?.data?.detail || error.message || '未知错误'}` })
  }
}

onMounted(loadConfig)
</script>
