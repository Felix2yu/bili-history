<template>
  <div
    class="glass-sidebar flex-shrink-0 flex flex-col h-full transition-all duration-300 ease-in-out z-40"
    :class="isCollapsed ? 'w-16' : 'w-60'"
  >
    <!-- Top Logo -->
    <div class="flex-shrink-0 px-3 py-3 border-b border-white/10 dark:border-white/5">
      <router-link to="/" class="flex items-center justify-center overflow-hidden">
        <img v-if="isCollapsed" src="/logo.svg" class="w-8 h-8 object-contain" alt="Logo" />
        <img v-else src="/logo.png" class="h-8 w-full object-contain" alt="Logo" />
      </router-link>
    </div>

    <!-- Navigation Menu -->
    <nav class="flex-1 overflow-y-auto py-3 px-2 space-y-1">
      <!-- History -->
      <button
        @click="changeContent('history')"
        :title="isCollapsed ? '历史记录' : ''"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'history' && !showRemarks
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">历史记录</span>
      </button>

      <!-- Favorites -->
      <router-link
        to="/favorites"
        :title="isCollapsed ? '我的收藏' : ''"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'favorites'
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">我的收藏</span>
      </router-link>

      <!-- Watch Later -->
      <router-link
        to="/watchlater"
        :title="isCollapsed ? '稍后再看' : ''"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'watchlater'
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">稍后再看</span>
      </router-link>

      <!-- Likes -->
      <router-link
        to="/likes"
        :title="isCollapsed ? '我的点赞' : ''"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'likes'
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">我的点赞</span>
      </router-link>

      <!-- Divider -->
      <div class="py-2 px-3">
        <div class="border-t border-gray-200/20 dark:border-gray-700/30"></div>
      </div>

      <!-- Analytics -->
      <a
        href="/analytics"
        @click.prevent="openAnalytics"
        :title="isCollapsed ? '年度总结' : ''"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 cursor-pointer text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100"
        :class="isCollapsed && 'justify-center px-0'"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">年度总结</span>
      </a>

      <!-- Media Manager -->
      <router-link
        to="/media"
        :title="isCollapsed ? '媒体管理' : ''"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'media'
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">媒体管理</span>
      </router-link>

      <!-- BiliTools -->
      <router-link
        to="/bili-tools"
        :title="isCollapsed ? 'B站助手' : ''"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'bili-tools'
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">B站助手</span>
      </router-link>

      <!-- Scheduler -->
      <router-link
        to="/scheduler"
        :title="isCollapsed ? '计划任务' : ''"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'scheduler'
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">计划任务</span>
      </router-link>

      <!-- Divider -->
      <div class="py-2 px-3">
        <div class="border-t border-gray-200/20 dark:border-gray-700/30"></div>
      </div>

      <!-- Settings -->
      <button
        @click="changeContent('settings')"
        :title="isCollapsed ? '设置' : ''"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200"
        :class="[
          currentContent === 'settings'
            ? 'bg-accent/10 text-accent shadow-accent-glow'
            : 'text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-900 dark:hover:text-gray-100',
          isCollapsed && 'justify-center px-0'
        ]"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <span v-show="!isCollapsed" class="truncate">设置</span>
      </button>

      <!-- Login Status -->
      <button
        @click="handleLoginClick"
        :title="isCollapsed ? (isLoggedIn ? (isPrivacyMode ? '已登录' : userInfo?.uname) : '未登录') : ''"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 text-gray-600 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 cursor-pointer"
        :class="isCollapsed && 'justify-center px-0'"
      >
        <div class="relative flex-shrink-0">
          <img
            v-if="isLoggedIn && userInfo?.face"
            :src="userInfo.face"
            alt="avatar"
            class="w-5 h-5 rounded-full object-cover"
            :class="{ 'blur-md': isPrivacyMode }"
            onerror="this.style.display='none'; this.nextElementSibling.style.display='flex'"
          />
          <div
            v-if="!isLoggedIn || !userInfo?.face"
            class="w-5 h-5 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center"
          >
            <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <span
            v-if="isLoggedIn"
            class="absolute -bottom-0.5 -right-0.5 w-2 h-2 rounded-full bg-green-400 border border-white dark:border-gray-800"
          ></span>
        </div>
        <span v-show="!isCollapsed" class="truncate" :class="{ 'text-green-500': isLoggedIn }">
          <template v-if="isLoggedIn">
            {{ isPrivacyMode ? '已登录' : (userInfo?.uname || '已登录') }}
          </template>
          <template v-else>未登录</template>
        </span>
      </button>
    </nav>

    <!-- Bottom section -->
    <div class="flex-shrink-0 border-t border-white/10 dark:border-white/5 px-3 py-3 space-y-2">
      <!-- Server status -->
      <div v-if="!isCollapsed" class="space-y-1.5 text-[11px]">
        <div class="flex items-center gap-1.5 text-gray-500 dark:text-gray-400">
          <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="serverStatus.isRunning ? 'bg-green-400' : 'bg-red-400'"></span>
          <span class="truncate">服务器 {{ serverStatus.isRunning ? '运行中' : '未连接' }}</span>
        </div>
        <div class="flex items-center gap-1.5 text-gray-500 dark:text-gray-400">
          <span
            class="w-1.5 h-1.5 rounded-full flex-shrink-0"
            :class="{
              'bg-green-400': integrityStatus.status === 'consistent',
              'bg-yellow-400': integrityStatus.status === 'inconsistent',
              'bg-gray-400': integrityStatus.status === 'disabled' || integrityStatus.status === 'unknown'
            }"
          ></span>
          <span
            class="truncate cursor-pointer hover:underline"
            @click="openDataSyncManager('integrity')"
          >
            数据 {{ integrityStatus.status === 'consistent' ? '一致' : integrityStatus.status === 'inconsistent' ? '不一致' : integrityStatus.status === 'disabled' ? '未开启' : '未检查' }}
          </span>
        </div>
        <div class="text-gray-400 dark:text-gray-500 truncate">
          SQLite {{ sqliteVersion?.sqlite_version || '...' }}
        </div>
      </div>

      <!-- Collapse button -->
      <button
        @click="toggleCollapse"
        :title="isCollapsed ? '展开侧边栏' : '收起侧边栏'"
        class="w-full flex items-center justify-center gap-2 px-2 py-2 rounded-xl text-gray-500 dark:text-gray-400 hover:bg-white/10 dark:hover:bg-white/5 hover:text-gray-700 dark:hover:text-gray-200 transition-all duration-200 text-sm"
      >
        <svg
          class="w-4 h-4 flex-shrink-0 transition-transform duration-300"
          :class="{ 'rotate-180': isCollapsed }"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path stroke-linecap="round" stroke-linejoin="round" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
        </svg>
        <span v-show="!isCollapsed" class="truncate text-xs">收起</span>
      </button>
    </div>
  </div>

  <!-- Login Dialog -->
  <LoginDialog
    v-model:show="showLoginDialog"
    @login-success="checkLoginStatus"
  />
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePrivacyStore } from '~/stores/privacy.js'
import { useDarkMode } from '~/stores/darkMode.js'
import { getLoginStatus, logout, getSqliteVersion, checkServerHealth, checkDataIntegrity, getIntegrityCheckConfig } from '~/utils/api'
import { showNotify, showDialog } from 'vant'
import 'vant/es/notify/style'
import 'vant/es/dialog/style'
import LoginDialog from './LoginDialog.vue'

const { isDarkMode } = useDarkMode()
const route = useRoute()
const router = useRouter()

const currentRoute = computed(() => route.path)

const currentContent = computed(() => {
  const path = route.path
  if (path.startsWith('/search')) return 'search'
  if (path.startsWith('/analytics')) return 'analytics'
  if (path.startsWith('/settings')) return 'settings'
  if (path.startsWith('/images')) return 'images'
  if (path.startsWith('/scheduler')) return 'scheduler'
  if (path.startsWith('/downloads')) return 'downloads'
  if (path.startsWith('/media')) return 'media'
  if (path.startsWith('/favorites')) return 'favorites'
  if (path.startsWith('/watchlater')) return 'watchlater'
  if (path.startsWith('/likes')) return 'likes'
  if (path.startsWith('/bili-tools')) return 'bili-tools'
  return 'history'
})

const props = defineProps({
  showRemarks: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['change-content', 'update:showRemarks'])

const changeContent = (content) => {
  if (content === 'history') {
    emit('change-content', content)
    emit('update:showRemarks', false)
    if (route.path !== '/' && !route.path.startsWith('/page/')) {
      router.push('/')
    }
  } else if (content === 'settings') {
    emit('change-content', content)
    emit('update:showRemarks', false)
    router.push('/settings')
  }
}

const openAnalytics = () => {
  const resolved = router.resolve({ path: '/analytics' })
  window.open(resolved.href, '_blank')
}

const isHistoryPage = computed(() => {
  return currentRoute.value === '/' || currentRoute.value.startsWith('/page/')
})

const { isPrivacyMode } = usePrivacyStore()

// Sidebar collapse state
const isCollapsed = ref(false)
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
  localStorage.setItem('sidebarCollapsed', isCollapsed.value.toString())

  if (isCollapsed.value) {
    localStorage.setItem('showSidebar', 'false')
    try {
      window.dispatchEvent(new CustomEvent('sidebar-toggle-changed', { detail: { showSidebar: false } }))
    } catch (error) {
      console.error('触发侧边栏切换事件失败:', error)
    }
  } else {
    localStorage.setItem('showSidebar', 'true')
    try {
      window.dispatchEvent(new CustomEvent('sidebar-toggle-changed', { detail: { showSidebar: true } }))
    } catch (error) {
      console.error('触发侧边栏切换事件失败:', error)
    }
  }
}

// SQLite version info
const sqliteVersion = ref({
  sqlite_version: '',
  user_version: 0,
  database_settings: { journal_mode: '', synchronous: 0, legacy_format: null },
  database_file: { exists: false, size_bytes: 0, size_mb: 0, path: '' }
})

// Login state
const isLoggedIn = ref(false)
const userInfo = ref(null)
const showLoginDialog = ref(false)

const checkLoginStatus = async () => {
  try {
    const response = await getLoginStatus()
    if (response.data && response.data.code === 0) {
      isLoggedIn.value = response.data.data.isLogin
      if (isLoggedIn.value) {
        userInfo.value = response.data.data
      }
    }
  } catch (error) {
    console.error('获取登录状态失败:', error)
    isLoggedIn.value = false
    userInfo.value = null
  }
}

const handleLoginClick = () => {
  if (!isLoggedIn.value) {
    showLoginDialog.value = true
  } else {
    showDialog({
      title: '确认退出',
      message: '确定要退出登录吗？',
      showCancelButton: true,
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      confirmButtonColor: '#fb7299'
    }).then(() => {
      handleLogout()
    }).catch(() => {})
  }
}

const handleLogout = async () => {
  try {
    const response = await logout()
    if (response.data.status === 'success') {
      showNotify({ type: 'success', message: '已成功退出登录' })
      isLoggedIn.value = false
      userInfo.value = null
      setTimeout(() => { window.location.reload() }, 1000)
    } else {
      throw new Error(response.data.message || '退出登录失败')
    }
  } catch (error) {
    console.error('退出登录失败:', error)
    showNotify({
      type: 'danger',
      message: error.response?.status === 500 ? '服务器错误,请稍后重试' : `退出登录失败: ${error.message}`
    })
  }
}

const fetchSqliteVersion = async () => {
  try {
    const response = await getSqliteVersion()
    if (response.data.status === 'success') {
      sqliteVersion.value = response.data.data
    }
  } catch (error) {
    console.error('获取SQLite版本失败:', error)
    sqliteVersion.value = null
  }
}

// Server status
const serverStatus = ref({ isRunning: false, timestamp: '', schedulerStatus: '' })
const integrityStatus = ref({ status: 'unknown', difference: 0, lastCheck: null })

const checkServerHealthStatus = async () => {
  try {
    const response = await checkServerHealth()
    if (response.data && response.data.status === 'running') {
      serverStatus.value = {
        isRunning: true,
        timestamp: response.data.timestamp,
        schedulerStatus: response.data.scheduler_status
      }
      return true
    }
    serverStatus.value.isRunning = false
    return false
  } catch (error) {
    console.error('服务器健康检查失败:', error)
    serverStatus.value.isRunning = false
    return false
  }
}

const fetchIntegrityStatus = async () => {
  try {
    const configResponse = await getIntegrityCheckConfig()
    if (configResponse.data && configResponse.data.success) {
      if (!configResponse.data.check_on_startup) {
        integrityStatus.value = { status: 'disabled', difference: 0, lastCheck: new Date().toISOString() }
        return
      }
    }
    const response = await checkDataIntegrity('output/bilibili_history.db', 'output/history_by_date', false)
    if (response.data && response.data.success) {
      if (response.data.message && response.data.message.includes('数据完整性校验已在配置中禁用')) {
        integrityStatus.value = { status: 'disabled', difference: 0, lastCheck: response.data.timestamp }
      } else {
        integrityStatus.value = {
          status: response.data.difference === 0 ? 'consistent' : 'inconsistent',
          difference: response.data.difference || 0,
          lastCheck: response.data.timestamp
        }
      }
    }
  } catch (error) {
    console.error('获取完整性状态失败:', error)
  }
}

const openDataSyncManager = (tab = null) => {
  window.dispatchEvent(new CustomEvent('open-data-sync-manager', { detail: { tab: tab || 'integrity' } }))
}

const healthCheckTimer = ref(null)

const setupPeriodicHealthCheck = () => {
  if (healthCheckTimer.value) clearInterval(healthCheckTimer.value)
  healthCheckTimer.value = setInterval(async () => {
    await checkServerHealthStatus()
  }, 30000)
}

// Event handlers
const handleLoginStatusChange = (event) => {
  if (event.detail && event.detail.isLoggedIn) {
    isLoggedIn.value = true
    if (event.detail.userInfo) {
      userInfo.value = event.detail.userInfo
    } else {
      checkLoginStatus()
    }
  } else {
    checkLoginStatus()
  }
}

const handleSidebarSettingChange = (event) => {
  if (event.detail && typeof event.detail.showSidebar === 'boolean') {
    if (!event.detail.showSidebar) {
      isCollapsed.value = true
    }
  }
}

onMounted(async () => {
  const showSidebar = localStorage.getItem('showSidebar') !== 'false'
  if (!showSidebar) {
    isCollapsed.value = true
  } else {
    isCollapsed.value = localStorage.getItem('sidebarCollapsed') === 'true'
  }

  checkLoginStatus()
  await fetchSqliteVersion()
  setupPeriodicHealthCheck()
  checkServerHealthStatus()
  await fetchIntegrityStatus()

  window.addEventListener('login-status-changed', handleLoginStatusChange)
  window.addEventListener('sidebar-setting-changed', handleSidebarSettingChange)
})

onUnmounted(() => {
  window.removeEventListener('login-status-changed', handleLoginStatusChange)
  window.removeEventListener('sidebar-setting-changed', handleSidebarSettingChange)
  if (healthCheckTimer.value) {
    clearInterval(healthCheckTimer.value)
    healthCheckTimer.value = null
  }
})
</script>
