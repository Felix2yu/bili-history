<template>
  <div>
    <div class="px-5 md:px-6 py-5 border-b border-gray-100 dark:border-gray-700/50">
      <div class="flex items-center justify-between gap-4">
        <div class="flex items-center gap-3 min-w-0">
          <div class="w-10 h-10 rounded-xl bg-orange-500 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0M3.124 7.5A8.969 8.969 0 0112 5.25a8.969 8.969 0 018.876 2.25" />
            </svg>
          </div>
          <div class="min-w-0">
            <h2 class="text-base md:text-lg font-semibold text-gray-900 dark:text-white truncate">通知设置</h2>
            <p class="text-[0.6875rem] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">配置 Shoutrrr 通知推送服务</p>
          </div>
        </div>
        <div class="flex gap-2 shrink-0">
          <button
            @click="resetConfig"
            class="inline-flex items-center px-3 py-2 text-[0.6875rem] font-medium text-accent md:text-sm bg-accent/10 rounded-xl hover:bg-accent/20 transition-colors"
          >
            <svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
            </svg>
            重置
          </button>
          <button
            @click="saveConfig"
            class="inline-flex items-center px-4 py-2 text-[0.6875rem] font-medium text-white md:text-sm bg-accent rounded-xl hover:shadow-lg hover:shadow-accent/25 transition-all"
          >保存</button>
          <button
            @click="testPush"
            class="inline-flex items-center px-4 py-2 text-[0.6875rem] font-medium text-white md:text-sm bg-emerald-500 rounded-xl hover:shadow-lg hover:shadow-emerald-500/25 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:shadow-none transition-all"
            :disabled="!config.enabled || urlList.length === 0"
          >测试</button>
        </div>
      </div>
    </div>

    <div class="px-5 md:px-6 py-5 space-y-4 md:space-y-5">
      <SettingToggle
        label="启用通知"
        description="关闭后所有 Shoutrrr 通知将不会发送"
        :modelValue="config.enabled"
        @update:modelValue="config.enabled = $event"
      />

      <div class="pt-2">
        <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-2">通知地址列表</label>
        <textarea
          v-model="config.urls"
          rows="5"
          class="block w-full rounded-xl border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 shadow-sm focus:border-accent focus:ring-accent text-[0.6875rem] md:text-sm resize-y"
          placeholder="每行一个地址，例如：&#10;bark://api.day.app/your-key&#10;tgram://bot-token/chat-id/&#10;smtp://user:pass@host:port/from@example.com/to@example.com&#10;discord://token@id"
        ></textarea>
        <p class="text-[0.6875rem] text-gray-500 dark:text-gray-400 md:text-xs mt-2">
          完整服务列表请查看
          <a href="https://containrrr.dev/shoutrrr/services/overview/" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline font-medium">Shoutrrr 支持的服务</a>
        </p>
      </div>

      <div v-if="urlList.length > 0" class="pt-2">
        <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-2">已配置服务</label>
        <div class="flex flex-wrap gap-2">
          <div
            v-for="(url, idx) in urlList"
            :key="idx"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-accent/10 text-accent text-[0.625rem] md:text-[0.6875rem] font-medium"
          >
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            {{ parseServiceName(url) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showNotify } from 'vant'
import { getShoutrrrConfig, updateShoutrrrConfig, testShoutrrrPush } from '~/utils/api'
import SettingToggle from './SettingToggle.vue'

const config = ref({
  enabled: false,
  urls: ''
})

const urlList = computed(() => {
  if (!config.value.urls) return []
  return config.value.urls.split('\n').filter(u => u.trim())
})

const parseServiceName = (url) => {
  try {
    const match = url.match(/^([a-zA-Z0-9]+):\/\//)
    if (match) {
      const name = match[1].toUpperCase()
      return name
    }
  } catch (e) {}
  return '未知服务'
}

const loadConfig = async () => {
  try {
    const response = await getShoutrrrConfig()
    let configData = null

    if (response.data && response.data.status === 'success' && response.data.data) {
      configData = response.data.data
    } else if (response.data && response.data.enabled !== undefined) {
      configData = response.data
    } else if (response.data && response.data.data) {
      configData = response.data.data
    }

    if (configData) {
      const urls = Array.isArray(configData.urls)
        ? configData.urls.join('\n')
        : (typeof configData.urls === 'string' ? configData.urls : '')
      config.value = {
        enabled: !!configData.enabled,
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
      showNotify({ type: 'success', message: '通知配置已保存' })
    }
  } catch (error) {
    showNotify({ type: 'danger', message: `保存失败：${error.response?.data?.detail || error.message}` })
    throw error
  }
}

const resetConfig = async () => {
  config.value = { enabled: false, urls: '' }
  try {
    await saveConfig()
  } catch (e) {}
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
