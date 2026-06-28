<template>
  <nav
    class="fixed bottom-0 left-0 right-0 z-50 glass-tabbar md:hidden transition-transform duration-300 ease-in-out"
    :class="{ 'translate-y-full': isHidden }"
  >
    <div class="flex items-center justify-around h-14 px-1 safe-area-bottom">
      <router-link
        v-for="tab in tabs"
        :key="tab.route"
        :to="tab.route"
        class="flex flex-col items-center justify-center flex-1 h-full gap-0.5 transition-all duration-200"
        :class="isActive(tab.route)
          ? 'text-accent'
          : 'text-gray-500 dark:text-gray-400'"
      >
        <div
          class="relative flex items-center justify-center w-6 h-6 transition-transform duration-200"
          :class="{ 'scale-110': isActive(tab.route) }"
        >
          <component :is="tab.icon" class="w-5 h-5" />
          <div
            v-if="isActive(tab.route)"
            class="absolute -top-1 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-accent"
          />
        </div>
        <span class="text-[10px] leading-none font-medium">{{ tab.label }}</span>
      </router-link>
    </div>
  </nav>
</template>

<script setup>
import { ref, h, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const isHidden = ref(false)
let lastScrollY = 0
let ticking = false

const onScroll = () => {
  if (ticking) return
  ticking = true

  requestAnimationFrame(() => {
    const scrollContainer = document.querySelector('main')
    if (!scrollContainer) { ticking = false; return }

    const currentScrollY = scrollContainer.scrollTop
    const scrollDelta = currentScrollY - lastScrollY

    if (currentScrollY < 10) {
      isHidden.value = false
    } else if (scrollDelta > 8) {
      isHidden.value = true
    } else if (scrollDelta < -8) {
      isHidden.value = false
    }

    lastScrollY = currentScrollY
    ticking = false
  })
}

onMounted(() => {
  const scrollContainer = document.querySelector('main')
  if (scrollContainer) {
    scrollContainer.addEventListener('scroll', onScroll, { passive: true })
  }
})

onUnmounted(() => {
  const scrollContainer = document.querySelector('main')
  if (scrollContainer) {
    scrollContainer.removeEventListener('scroll', onScroll)
  }
})

const HomeIcon = {
  render() {
    return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.8' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' })
    ])
  }
}

const BookmarkIcon = {
  render() {
    return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.8' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: 'M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z' })
    ])
  }
}

const CubeIcon = {
  render() {
    return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.8' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4' })
    ])
  }
}

const WrenchIcon = {
  render() {
    return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.8' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: 'M13 10V3L4 14h7v7l9-11h-7z' })
    ])
  }
}

const UserIcon = {
  render() {
    return h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.8' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' })
    ])
  }
}

const tabs = [
  { label: '首页', route: '/', icon: HomeIcon },
  { label: '收藏', route: '/favorites', icon: BookmarkIcon },
  { label: '媒体', route: '/media', icon: CubeIcon },
  { label: '工具', route: '/bili-tools', icon: WrenchIcon },
  { label: '我的', route: '/settings', icon: UserIcon },
]

const isActive = (tabRoute) => {
  if (tabRoute === '/') {
    return route.path === '/' || route.path.startsWith('/page/')
  }
  return route.path.startsWith(tabRoute)
}
</script>

<style scoped>
.safe-area-bottom {
  padding-bottom: env(safe-area-inset-bottom, 0px);
}
</style>
