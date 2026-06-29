<template>
  <div class="flex flex-col h-full">
    <!-- Header with logo and close -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-white/10 dark:border-white/5">
      <router-link to="/" @click="emit('navigate')" class="flex items-center gap-2">
        <img src="/logo.png" class="h-7 object-contain" alt="Logo" />
      </router-link>
      <button @click="emit('navigate')" class="w-8 h-8 rounded-full flex items-center justify-center hover:bg-white/10 dark:hover:bg-white/5 transition-colors text-gray-500 md:hidden">
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto py-3 px-2 space-y-1">
      <button
        @click="changeContent('history')"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'history' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>历史记录</span>
      </button>

      <router-link to="/favorites" @click="emit('navigate')"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'favorites' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
        </svg>
        <span>我的收藏</span>
      </router-link>

      <router-link to="/watchlater" @click="emit('navigate')"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'watchlater' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>稍后再看</span>
      </router-link>

      <router-link to="/likes" @click="emit('navigate')"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'likes' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
        </svg>
        <span>我的点赞</span>
      </router-link>

      <div class="py-2 px-3"><div class="border-t border-gray-200/20 dark:border-gray-700/30"></div></div>

      <a href="/analytics" @click.prevent="openAnalytics(); emit('navigate')"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 cursor-pointer text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
        <span>年度总结</span>
      </a>

      <router-link to="/media" @click="emit('navigate')"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'media' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
        <span>媒体管理</span>
      </router-link>

      <router-link to="/bili-tools" @click="emit('navigate')"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'bili-tools' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        <span>B站助手</span>
      </router-link>

      <router-link to="/scheduler" @click="emit('navigate')"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'scheduler' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>计划任务</span>
      </router-link>

      <div class="py-2 px-3"><div class="border-t border-gray-200/20 dark:border-gray-700/30"></div></div>

      <button @click="changeContent('settings')"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="currentContent === 'settings' ? 'bg-accent/10 text-accent' : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <span>设置</span>
      </button>
    </nav>

    <!-- Bottom: user info -->
    <div class="flex-shrink-0 border-t border-white/10 dark:border-white/5 px-3 py-3">
      <button @click="handleLoginClick"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 cursor-pointer"
      >
        <div class="relative flex-shrink-0">
          <img v-if="isLoggedIn && userInfo?.face" :src="userInfo.face" alt="avatar"
            class="w-6 h-6 rounded-full object-cover" :class="{ 'blur-md': isPrivacyMode }"
            onerror="this.style.display='none'; this.nextElementSibling.style.display='flex'"
          />
          <div v-if="!isLoggedIn || !userInfo?.face" class="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
            <svg class="w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <span v-if="isLoggedIn" class="absolute -bottom-0.5 -right-0.5 w-2 h-2 rounded-full bg-green-400 border-2 border-white dark:border-gray-800"></span>
        </div>
        <span class="truncate" :class="{ 'text-green-500': isLoggedIn }">
          {{ isLoggedIn ? (isPrivacyMode ? '已登录' : (userInfo?.uname || '已登录')) : '未登录' }}
        </span>
      </button>

      <div class="mt-2 px-3 text-[11px] text-gray-400 dark:text-gray-500 space-y-1">
        <div class="flex items-center gap-1.5">
          <span class="w-1.5 h-1.5 rounded-full" :class="serverStatus.isRunning ? 'bg-green-400' : 'bg-red-400'"></span>
          <span>服务器 {{ serverStatus.isRunning ? '运行中' : '未连接' }}</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-1.5 h-1.5 rounded-full" :class="{ 'bg-green-400': integrityStatus.status === 'consistent', 'bg-yellow-400': integrityStatus.status === 'inconsistent', 'bg-gray-400': integrityStatus.status !== 'consistent' && integrityStatus.status !== 'inconsistent' }"></span>
          <span class="cursor-pointer hover:underline" @click="openDataSyncManager('integrity')">
            数据 {{ integrityStatus.status === 'consistent' ? '一致' : integrityStatus.status === 'inconsistent' ? '不一致' : '未检查' }}
          </span>
        </div>
      </div>
    </div>
  </div>

  <LoginDialog v-model:show="showLoginDialog" @login-success="checkLoginStatus" />
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePrivacyStore } from '@/store/privacy.js'
import { useDarkMode } from '@/store/darkMode.js'
import { getLoginStatus, logout, checkServerHealth, checkDataIntegrity, getIntegrityCheckConfig } from '@/api/api'
import { showNotify, showDialog } from 'vant'
import 'vant/es/notify/style'
import 'vant/es/dialog/style'
import LoginDialog from './LoginDialog.vue'

const emit = defineEmits(['navigate'])
const route = useRoute()
const router = useRouter()
const { isDarkMode } = useDarkMode()
const { isPrivacyMode } = usePrivacyStore()

const currentContent = computed(() => {
  const p = route.path
  if (p.startsWith('/search')) return 'search'
  if (p.startsWith('/analytics')) return 'analytics'
  if (p.startsWith('/settings')) return 'settings'
  if (p.startsWith('/images')) return 'images'
  if (p.startsWith('/scheduler')) return 'scheduler'
  if (p.startsWith('/downloads')) return 'downloads'
  if (p.startsWith('/media')) return 'media'
  if (p.startsWith('/favorites')) return 'favorites'
  if (p.startsWith('/watchlater')) return 'watchlater'
  if (p.startsWith('/likes')) return 'likes'
  if (p.startsWith('/bili-tools')) return 'bili-tools'
  return 'history'
})

const changeContent = (content) => {
  emit('navigate')
  if (content === 'history') {
    if (route.path !== '/' && !route.path.startsWith('/page/')) router.push('/')
  } else if (content === 'settings') {
    router.push('/settings')
  }
}

const openAnalytics = () => {
  window.open(router.resolve({ path: '/analytics' }).href, '_blank')
}

const isLoggedIn = ref(false)
const userInfo = ref(null)
const showLoginDialog = ref(false)

const checkLoginStatus = async () => {
  try {
    const response = await getLoginStatus()
    if (response.data && response.data.code === 0) {
      isLoggedIn.value = response.data.data.isLogin
      if (isLoggedIn.value) userInfo.value = response.data.data
    }
  } catch { isLoggedIn.value = false; userInfo.value = null }
}

const handleLoginClick = () => {
  if (!isLoggedIn.value) {
    showLoginDialog.value = true
  } else {
    showDialog({
      title: '确认退出', message: '确定要退出登录吗？',
      showCancelButton: true, confirmButtonText: '确认', cancelButtonText: '取消', confirmButtonColor: '#fb7299'
    }).then(() => handleLogout()).catch(() => {})
  }
}

const handleLogout = async () => {
  try {
    const response = await logout()
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: '已成功退出登录' })
      isLoggedIn.value = false; userInfo.value = null
      setTimeout(() => { window.location.reload() }, 1000)
    }
  } catch (error) {
    showNotify({ type: 'danger', message: `退出登录失败: ${error.message}` })
  }
}

const serverStatus = ref({ isRunning: false })
const integrityStatus = ref({ status: 'unknown' })

const checkServerHealthStatus = async () => {
  try {
    const response = await checkServerHealth()
    serverStatus.value.isRunning = response.data?.status === 'running'
  } catch { serverStatus.value.isRunning = false }
}

const fetchIntegrityStatus = async () => {
  try {
    const configResponse = await getIntegrityCheckConfig()
    if (configResponse.data?.success && !configResponse.data.check_on_startup) {
      integrityStatus.value = { status: 'disabled' }; return
    }
    const response = await checkDataIntegrity('output/bilibili_history.db', 'output/history_by_date', false)
    if (response.data?.success) {
      integrityStatus.value = {
        status: response.data.difference === 0 ? 'consistent' : 'inconsistent'
      }
    }
  } catch {}
}

const openDataSyncManager = (tab) => {
  window.dispatchEvent(new CustomEvent('open-data-sync-manager', { detail: { tab: tab || 'integrity' } }))
}

let healthCheckTimer = null
onMounted(async () => {
  checkLoginStatus()
  checkServerHealthStatus()
  await fetchIntegrityStatus()
  healthCheckTimer = setInterval(checkServerHealthStatus, 30000)
  window.addEventListener('login-status-changed', (e) => {
    if (e.detail?.isLoggedIn) { isLoggedIn.value = true; if (e.detail.userInfo) userInfo.value = e.detail.userInfo; else checkLoginStatus() }
    else checkLoginStatus()
  })
})
onUnmounted(() => { if (healthCheckTimer) clearInterval(healthCheckTimer) })
</script>
