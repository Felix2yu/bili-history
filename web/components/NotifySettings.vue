<template>
  <div>
    <div class="border-b border-gray-100 px-5 py-5 md:px-6 dark:border-gray-700/50">
      <div class="flex items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-orange-500">
            <svg
              class="h-5 w-5 text-white"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0M3.124 7.5A8.969 8.969 0 0112 5.25a8.969 8.969 0 018.876 2.25"
              />
            </svg>
          </div>
          <div class="min-w-0">
            <h2 class="truncate text-base font-semibold text-gray-900 md:text-lg dark:text-white">
              通知设置
            </h2>
            <p class="mt-0.5 text-[0.6875rem] text-gray-500 md:text-xs dark:text-gray-400">
              配置 Apprise 通知推送服务
            </p>
          </div>
        </div>
        <div class="flex shrink-0 gap-2">
          <button
            class="inline-flex items-center rounded-xl bg-accent/10 px-3 py-2 text-[0.6875rem] font-medium text-accent transition-colors hover:bg-accent/20 md:text-sm"
            @click="resetConfig"
          >
            <svg
              class="mr-1 h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99"
              />
            </svg>
            重置
          </button>
          <button
            class="inline-flex items-center rounded-xl bg-accent px-4 py-2 text-[0.6875rem] font-medium text-white transition-all hover:shadow-lg hover:shadow-accent/25 md:text-sm"
            @click="saveConfig"
          >
            保存
          </button>
          <button
            class="inline-flex items-center rounded-xl bg-emerald-500 px-4 py-2 text-[0.6875rem] font-medium text-white transition-all hover:shadow-lg hover:shadow-emerald-500/25 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:shadow-none md:text-sm"
            :disabled="!config.enabled || urlList.length === 0"
            @click="testPush"
          >
            测试
          </button>
        </div>
      </div>
    </div>

    <div class="space-y-4 px-5 py-5 md:space-y-5 md:px-6">
      <SettingToggle
        label="启用通知"
        description="关闭后所有通知将不会发送"
        :model-value="config.enabled"
        @update:model-value="config.enabled = $event"
      />

      <div class="pt-2">
        <label
          class="mb-2 block text-[0.75rem] font-medium text-gray-700 md:text-sm dark:text-gray-300"
          >通知地址列表</label
        >
        <textarea
          v-model="config.urls"
          rows="5"
          class="block w-full resize-y rounded-xl border-gray-200 px-3.5 py-2.5 text-[0.6875rem] shadow-sm focus:border-accent focus:ring-accent md:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
          placeholder="每行一个地址，例如：&#10;bark://api.day.app/your-key&#10;tgram://bot-token/chat-id/&#10;smtp://user:pass@host:port/from@example.com/to@example.com&#10;discord://token@id"
        />
        <p class="mt-2 text-[0.6875rem] text-gray-500 md:text-xs dark:text-gray-400">
          完整服务列表请查看
          <a
            href="https://appriseit.com/services/"
            target="_blank"
            rel="noopener noreferrer"
            class="font-medium text-accent hover:underline"
            >Apprise 支持的服务</a
          >
        </p>
      </div>

      <div v-if="urlList.length > 0" class="pt-2">
        <label
          class="mb-2 block text-[0.75rem] font-medium text-gray-700 md:text-sm dark:text-gray-300"
          >已配置服务</label
        >
        <div class="flex flex-wrap gap-2">
          <div
            v-for="(url, idx) in urlList"
            :key="idx"
            class="inline-flex items-center gap-1.5 rounded-lg bg-accent/10 px-2.5 py-1 text-[0.625rem] font-medium text-accent md:text-[0.6875rem]"
          >
            <svg
              class="h-3 w-3"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
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
import { ref, computed, onMounted } from 'vue';
import { showNotify } from 'vant';
import { getNotifyConfig, updateNotifyConfig, testNotifyPush } from '~/utils/api';
import SettingToggle from './SettingToggle.vue';

const config = ref({
  enabled: false,
  urls: ''
});

const urlList = computed(() => {
  if (!config.value.urls) return [];
  return config.value.urls.split('\n').filter((u) => u.trim());
});

const parseServiceName = (url) => {
  try {
    const match = url.match(/^([a-zA-Z0-9]+):\/\//);
    if (match) {
      const name = match[1].toUpperCase();
      return name;
    }
  } catch {
    // URL 无法解析出协议头时，回落「未知服务」
  }
  return '未知服务';
};

const loadConfig = async () => {
  try {
    const response = await getNotifyConfig();
    let configData = null;

    if (response.data && response.data.status === 'success' && response.data.data) {
      configData = response.data.data;
    } else if (response.data && response.data.enabled !== undefined) {
      configData = response.data;
    } else if (response.data && response.data.data) {
      configData = response.data.data;
    }

    if (configData) {
      const urls = Array.isArray(configData.urls)
        ? configData.urls.join('\n')
        : typeof configData.urls === 'string'
          ? configData.urls
          : '';
      config.value = {
        enabled: !!configData.enabled,
        urls
      };
    }
  } catch (error) {
    console.error('获取通知配置失败:', error);
  }
};

const saveConfig = async () => {
  try {
    const urls = urlList.value;
    const response = await updateNotifyConfig({
      enabled: config.value.enabled,
      urls
    });
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: '通知配置已保存' });
    }
  } catch (error) {
    showNotify({
      type: 'danger',
      message: `保存失败：${error.response?.data?.detail || error.message}`
    });
    throw error;
  }
};

const resetConfig = async () => {
  config.value = { enabled: false, urls: '' };
  try {
    await saveConfig();
  } catch {
    // saveConfig 内部已弹出错误提示，此处只需防止未捕获的 Promise 拒绝
  }
};

const testPush = async () => {
  if (!config.value.enabled || urlList.value.length === 0) {
    showNotify({ type: 'warning', message: '请先启用通知并填写推送地址' });
    return;
  }
  try {
    showNotify({ type: 'primary', message: '正在发送测试通知...' });
    const response = await testNotifyPush();
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: response.data.message || '测试通知已发送' });
    } else {
      showNotify({ type: 'danger', message: response.data.message || '测试通知发送失败' });
    }
  } catch (error) {
    showNotify({
      type: 'danger',
      message: `发送失败：${error.response?.data?.detail || error.message || '未知错误'}`
    });
  }
};

onMounted(loadConfig);
</script>
