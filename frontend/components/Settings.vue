<template>
  <div class="min-h-screen pb-20 md:pb-0">
    <!-- 背景装饰 -->
    <div class="fixed inset-0 -z-10 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-96 h-96 bg-accent/10 rounded-full blur-3xl"></div>
      <div class="absolute top-1/2 -left-40 w-96 h-96 bg-accent/5 rounded-full blur-3xl"></div>
    </div>

    <div class="max-w-7xl mx-auto px-4 md:px-6 py-6 md:py-8">
      <!-- 页面标题区 -->
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6 md:mb-8">
        <div class="flex items-center gap-4">
          <div class="relative">
            <div class="w-12 h-12 md:w-14 md:h-14 rounded-2xl bg-accent flex items-center justify-center shadow-lg shadow-accent/25">
              <svg class="w-6 h-6 md:w-7 md:h-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
          </div>
          <div>
            <h1 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">设置</h1>
            <p class="text-xs md:text-sm text-gray-500 dark:text-gray-400 mt-0.5">个性化您的体验，管理数据与通知</p>
          </div>
        </div>
      </div>

      <div class="flex flex-col lg:flex-row gap-6">
        <!-- 侧边栏导航 (桌面端) -->
        <nav class="hidden lg:block w-56 flex-shrink-0">
          <div class="sticky top-6 space-y-1">
            <button
              v-for="(tab, index) in settingTabs"
              :key="index"
              @click="activeTab = tab.key"
              class="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200"
              :class="activeTab === tab.key
                ? 'bg-accent/10 text-accent shadow-sm'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200'"
            >
              <div class="w-5 h-5" v-html="tab.icon"></div>
              <span>{{ tab.label }}</span>
            </button>
          </div>
        </nav>

        <!-- 移动端标签栏 -->
        <div class="lg:hidden mb-2">
          <div class="bg-gray-100 dark:bg-gray-800/60 rounded-2xl p-1.5">
            <nav class="flex gap-1">
              <button
                v-for="(tab, index) in settingTabs"
                :key="index"
                @click="activeTab = tab.key"
                class="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-xl text-[0.75rem] font-medium transition-all duration-200 whitespace-nowrap"
                :class="activeTab === tab.key
                  ? 'bg-white dark:bg-gray-700 text-accent shadow-sm'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
              >
                <div class="w-4 h-4" v-html="tab.icon"></div>
                <span>{{ tab.label }}</span>
              </button>
            </nav>
          </div>
        </div>

        <!-- 内容区域 -->
        <div class="flex-1 min-w-0 space-y-5 md:space-y-6">
          <!-- 基础设置 -->
          <Transition name="fade" mode="out-in">
            <div v-if="activeTab === 'basic'" key="basic" class="space-y-5 md:space-y-6">
              <!-- 外观设置卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <div class="px-5 md:px-6 py-5 border-b border-gray-100 dark:border-gray-700/50">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 rounded-xl bg-accent flex items-center justify-center">
                      <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M7.217 10.907a2.25 2.25 0 100 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186l9.566-5.314m-9.566 7.5l9.566 5.314m0 0a2.25 2.25 0 103.935 2.186 2.25 2.25 0 00-3.935-2.186zm0-12.814a2.25 2.25 0 103.933-2.185 2.25 2.25 0 00-3.933 2.185z" />
                      </svg>
                    </div>
                    <div>
                      <h2 class="text-base md:text-lg font-semibold text-gray-900 dark:text-white">外观设置</h2>
                      <p class="text-[0.6875rem] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">定制您的视觉体验</p>
                    </div>
                  </div>
                </div>

                <!-- 深色模式 -->
                <div class="px-5 md:px-6 py-4 border-b border-gray-50 dark:border-gray-700/30">
                  <div class="flex items-start justify-between gap-4">
                    <div class="min-w-0 flex-1">
                      <h3 class="text-[0.8125rem] md:text-sm font-medium text-gray-900 dark:text-gray-100">深色模式</h3>
                      <p class="text-[0.6875rem] text-gray-500 dark:text-gray-400 mt-0.5 md:text-xs">切换应用的显示主题，跟随系统将自动匹配系统设置</p>
                    </div>
                    <div class="flex bg-gray-100 dark:bg-gray-700/60 rounded-xl p-0.5 shrink-0">
                      <button
                        v-for="mode in darkModeOptions"
                        :key="mode.value"
                        @click="darkMode = mode.value; handleDarkModeChange()"
                        class="px-2.5 md:px-3 py-1.5 text-[0.6875rem] md:text-xs font-medium rounded-lg transition-all duration-200"
                        :class="darkMode === mode.value
                          ? 'bg-white dark:bg-gray-600 text-accent shadow-sm'
                          : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
                      >
                        {{ mode.label }}
                      </button>
                    </div>
                  </div>
                </div>

                <!-- 主题色 -->
                <div class="px-5 md:px-6 py-4 md:py-5">
                  <div class="mb-4">
                    <h3 class="text-[0.8125rem] md:text-sm font-medium text-gray-900 dark:text-gray-100">主题色</h3>
                    <p class="text-[0.6875rem] text-gray-500 dark:text-gray-400 mt-0.5 md:text-xs">选择您喜欢的主题色彩，打造个性化界面</p>
                  </div>
                  <div class="flex flex-wrap gap-2.5 md:gap-3">
                    <button
                      v-for="theme in themePresets"
                      :key="theme.id"
                      @click="handleThemeColorChange(theme.id)"
                      class="relative group flex flex-col items-center gap-1.5 p-2 rounded-xl transition-all duration-200"
                      :class="themeColor === theme.id
                        ? 'bg-accent/10 ring-2 ring-accent ring-offset-2 ring-offset-white dark:ring-offset-gray-800'
                        : 'hover:bg-gray-50 dark:hover:bg-gray-700/50'"
                    >
                      <div
                        class="w-9 h-9 md:w-10 md:h-10 rounded-full shadow-md transition-all duration-200 group-hover:scale-110 group-hover:shadow-lg flex items-center justify-center"
                        :style="{ backgroundColor: theme.color }"
                      >
                        <svg v-if="themeColor === theme.id" class="w-4 h-4 md:w-5 md:h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                        </svg>
                      </div>
                      <span class="text-[0.625rem] md:text-[0.6875rem] text-gray-600 dark:text-gray-400 font-medium">{{ theme.name }}</span>
                    </button>
                  </div>
                </div>

                <!-- 字体大小 -->
                <div class="px-5 md:px-6 py-4 border-b border-gray-50 dark:border-gray-700/30">
                  <div class="flex items-start justify-between gap-4">
                    <div class="min-w-0 flex-1">
                      <h3 class="text-[0.8125rem] md:text-sm font-medium text-gray-900 dark:text-gray-100">字体大小</h3>
                      <p class="text-[0.6875rem] text-gray-500 dark:text-gray-400 mt-0.5 md:text-xs">调整页面文字大小，提升阅读体验</p>
                    </div>
                    <div class="flex bg-gray-100 dark:bg-gray-700/60 rounded-xl p-0.5 shrink-0">
                      <button
                        v-for="preset in fontSizePresets"
                        :key="preset.id"
                        @click="handleFontSizeChange(preset.id)"
                        class="px-2.5 md:px-3 py-1.5 text-[0.6875rem] md:text-xs font-medium rounded-lg transition-all duration-200"
                        :class="fontSize === preset.id
                          ? 'bg-white dark:bg-gray-600 text-accent shadow-sm'
                          : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
                      >
                        {{ preset.label }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 通用设置卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <div class="px-5 md:px-6 py-5 border-b border-gray-100 dark:border-gray-700/50">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 rounded-xl bg-blue-500 flex items-center justify-center">
                      <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 11-3 0m3 0a1.5 1.5 0 10-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-9.75 0h9.75" />
                      </svg>
                    </div>
                    <div>
                      <h2 class="text-base md:text-lg font-semibold text-gray-900 dark:text-white">通用设置</h2>
                      <p class="text-[0.6875rem] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">界面显示与数据同步选项</p>
                    </div>
                  </div>
                </div>

                <SettingToggle label="使用本地图片源" description="选择使用本地图片源或在线图片源，本地图片源适合离线访问" :modelValue="useLocalImages" @update:modelValue="useLocalImages = $event; handleImageSourceChange()" />
                <SettingToggle label="侧边栏显示" description="设置是否默认显示侧边栏，关闭后侧边栏将自动收起" :modelValue="showSidebar" @update:modelValue="showSidebar = $event; handleSidebarChange()" />
                <SettingToggle label="首页默认网格布局" description="设置历史记录页面的默认展示方式，开启为网格视图，关闭为列表视图" :modelValue="isGridLayout" @update:modelValue="isGridLayout = $event; handleLayoutChange()" />
                <SettingToggle label="同步已删除记录" description="开启后将同步已删除的历史记录，建议仅在需要恢复记录时开启" :modelValue="syncDeleted" @update:modelValue="syncDeleted = $event; handleSyncDeletedChange()" />
                <SettingToggle label="同步删除B站历史记录" description="开启后删除本地历史记录时，同时删除B站服务器上的对应记录" :modelValue="syncDeleteToBilibili" @update:modelValue="syncDeleteToBilibili = $event; handleSyncDeleteToBilibiliChange()" />
                <SettingToggle label="启动时数据完整性校验" description="开启后每次启动应用时都会进行数据完整性校验，关闭可加快启动速度" :modelValue="checkIntegrityOnStartup" @update:modelValue="checkIntegrityOnStartup = $event; handleIntegrityCheckChange()" />
              </div>

              <!-- MCP服务卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <div class="px-5 md:px-6 py-5 border-b border-gray-100 dark:border-gray-700/50">
                  <div class="flex items-center justify-between gap-4">
                    <div class="flex items-center gap-3 min-w-0">
                      <div class="w-10 h-10 rounded-xl bg-purple-500 flex items-center justify-center shrink-0">
                        <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 12l3 3m0 0l3-3m-3 3V6" />
                        </svg>
                      </div>
                      <div class="min-w-0">
                        <h2 class="text-base md:text-lg font-semibold text-gray-900 dark:text-white truncate">MCP服务</h2>
                        <p class="text-[0.6875rem] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">允许AI客户端通过MCP协议读取本地历史记录</p>
                      </div>
                    </div>
                    <label
                      class="relative inline-flex shrink-0 items-center"
                      :class="(isMcpConfigLoading || isMcpConfigSaving) ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'"
                    >
                      <input
                        type="checkbox"
                        v-model="mcpConfig.enabled"
                        class="peer sr-only"
                        :disabled="isMcpConfigLoading || isMcpConfigSaving"
                        @change="handleMcpEnabledChange"
                      >
                      <div class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:translate-x-0 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-accent peer-checked:after:translate-x-full peer-checked:after:border-white dark:bg-gray-600"></div>
                    </label>
                  </div>
                </div>

                <div v-if="mcpConfigAvailable" class="px-5 md:px-6 py-5 space-y-4 md:space-y-5">
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label class="mb-1.5 block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm">MCP URL</label>
                      <div class="flex gap-2">
                        <input
                          :value="mcpUrl"
                          readonly
                          type="text"
                          class="block min-w-0 flex-1 rounded-xl border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-900 text-[0.6875rem] md:text-sm focus:border-accent focus:ring-accent dark:text-gray-100"
                        />
                        <button
                          @click="copyText(mcpUrl, 'MCP URL')"
                          class="inline-flex shrink-0 items-center justify-center rounded-xl bg-accent/10 px-3 text-[0.6875rem] font-medium text-accent hover:bg-accent/20 md:text-sm transition-colors"
                          title="复制MCP URL"
                        >
                          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                          </svg>
                        </button>
                      </div>
                    </div>

                    <div>
                      <label class="mb-1.5 block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm">Bearer Token</label>
                      <div class="flex gap-2">
                        <input
                          :value="mcpConfig.token"
                          readonly
                          type="password"
                          class="block min-w-0 flex-1 rounded-xl border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-900 text-[0.6875rem] md:text-sm focus:border-accent focus:ring-accent dark:text-gray-100"
                          placeholder="未配置Token"
                        />
                        <button
                          @click="copyText(mcpConfig.token, 'Bearer Token')"
                          :disabled="!mcpConfig.token"
                          class="inline-flex shrink-0 items-center justify-center rounded-xl bg-accent/10 px-3 text-[0.6875rem] font-medium text-accent hover:bg-accent/20 disabled:opacity-50 disabled:cursor-not-allowed md:text-sm transition-colors"
                          title="复制Bearer Token"
                        >
                          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>

                  <div>
                    <div class="mb-1.5 flex items-center justify-between gap-2">
                      <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm">AI连接提示词</label>
                      <button
                        @click="copyText(mcpConnectionPrompt, 'AI连接提示词')"
                        class="inline-flex shrink-0 items-center rounded-xl bg-accent/10 px-3 py-1.5 text-[0.6875rem] font-medium text-accent hover:bg-accent/20 md:text-sm transition-colors"
                      >
                        <svg class="mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                        复制提示词
                      </button>
                    </div>
                    <textarea
                      :value="mcpConnectionPrompt"
                      readonly
                      rows="8"
                      class="block w-full resize-y rounded-xl border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-900 text-[0.6875rem] leading-5 md:text-sm focus:border-accent focus:ring-accent dark:text-gray-100"
                    ></textarea>
                  </div>

                  <div>
                    <div class="mb-1.5 flex items-center justify-between gap-2">
                      <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm">配套 Skill</label>
                      <button
                        @click="copyText(mcpConfig.skill_content, '配套 Skill')"
                        :disabled="!mcpConfig.skill_content"
                        class="inline-flex shrink-0 items-center rounded-xl bg-accent/10 px-3 py-1.5 text-[0.6875rem] font-medium text-accent hover:bg-accent/20 disabled:opacity-50 disabled:cursor-not-allowed md:text-sm transition-colors"
                      >
                        <svg class="mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                        复制 Skill
                      </button>
                    </div>
                    <textarea
                      :value="mcpConfig.skill_content || '后端未找到配套 Skill 文件'"
                      readonly
                      rows="10"
                      class="block w-full resize-y rounded-xl border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-900 text-[0.6875rem] leading-5 md:text-sm focus:border-accent focus:ring-accent dark:text-gray-100"
                    ></textarea>
                  </div>
                </div>

                <p v-else-if="!isMcpConfigLoading" class="px-5 md:px-6 py-4 text-[0.6875rem] md:text-sm text-amber-600 dark:text-amber-300">
                  配置详情加载失败，请确保后端已更新并重启。
                </p>
              </div>

              <!-- 通知设置卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <ShoutrrrSettings />
              </div>

            </div>

            <!-- 数据管理 -->
            <div v-if="activeTab === 'data'" key="data" class="space-y-5 md:space-y-6">
              <!-- 数据导出卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <div class="px-5 md:px-6 py-5 border-b border-gray-100 dark:border-gray-700/50">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 rounded-xl bg-emerald-500 flex items-center justify-center">
                      <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                      </svg>
                    </div>
                    <div>
                      <h2 class="text-base md:text-lg font-semibold text-gray-900 dark:text-white">数据导出</h2>
                      <p class="text-[0.6875rem] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">导出历史记录数据到Excel文件</p>
                    </div>
                  </div>
                </div>

                <div class="px-5 md:px-6 py-5">
                  <div class="bg-gray-50 dark:bg-gray-900/50 border border-gray-100 dark:border-gray-700/50 p-4 md:p-5 rounded-2xl">
                    <div class="flex flex-wrap items-end gap-4">
                      <div class="w-full sm:w-32">
                        <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-1.5">年份</label>
                        <select
                          v-model="exportOptions.year"
                          class="block w-full rounded-xl border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 shadow-sm focus:border-accent focus:ring-accent text-[0.6875rem] md:text-sm py-2"
                        >
                          <option v-for="year in availableYears" :key="year" :value="year">
                            {{ year }}年
                          </option>
                        </select>
                      </div>

                      <div class="w-full sm:w-40">
                        <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-1.5">导出类型</label>
                        <select
                          v-model="exportOptions.exportType"
                          class="block w-full rounded-xl border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 shadow-sm focus:border-accent focus:ring-accent text-[0.6875rem] md:text-sm py-2"
                        >
                          <option value="year">全年数据</option>
                          <option value="month">按月份</option>
                          <option value="dateRange">按日期范围</option>
                        </select>
                      </div>

                      <div v-if="exportOptions.exportType === 'month'" class="w-full sm:w-28">
                        <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-1.5">月份</label>
                        <select
                          v-model="exportOptions.month"
                          class="block w-full rounded-xl border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 shadow-sm focus:border-accent focus:ring-accent text-[0.6875rem] md:text-sm py-2"
                        >
                          <option v-for="month in 12" :key="month" :value="month">
                            {{ month }}月
                          </option>
                        </select>
                      </div>

                      <template v-if="exportOptions.exportType === 'dateRange'">
                        <div class="w-full sm:w-40">
                          <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-1.5">开始日期</label>
                          <input
                            type="date"
                            v-model="exportOptions.startDate"
                            class="block w-full rounded-xl border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 shadow-sm focus:border-accent focus:ring-accent text-[0.6875rem] md:text-sm py-2"
                          />
                        </div>

                        <div class="w-full sm:w-40">
                          <label class="block text-[0.75rem] font-medium text-gray-700 dark:text-gray-300 md:text-sm mb-1.5">结束日期</label>
                          <input
                            type="date"
                            v-model="exportOptions.endDate"
                            class="block w-full rounded-xl border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 shadow-sm focus:border-accent focus:ring-accent text-[0.6875rem] md:text-sm py-2"
                          />
                        </div>
                      </template>

                      <button
                        @click="exportAndDownloadExcel"
                        :disabled="isExporting"
                        class="inline-flex items-center px-5 py-2.5 text-[0.6875rem] font-medium text-white md:text-sm bg-accent rounded-xl hover:shadow-lg hover:shadow-accent/25 disabled:opacity-50 transition-all"
                      >
                        <svg v-if="isExporting" class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
                          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        <svg v-else class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M12 10.5v6m3-3H9m4.06-7.19l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" />
                        </svg>
                        {{ isExporting ? '导出中...' : '导出Excel' }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 数据库下载卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <div class="px-5 md:px-6 py-5">
                  <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div class="flex items-start gap-3">
                      <div class="w-10 h-10 rounded-xl bg-blue-500 flex items-center justify-center shrink-0">
                        <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
                        </svg>
                      </div>
                      <div>
                        <h3 class="text-[0.8125rem] md:text-sm font-medium text-gray-900 dark:text-gray-100">数据库下载</h3>
                        <p class="text-[0.6875rem] text-gray-500 dark:text-gray-400 mt-0.5 md:text-xs">下载完整的SQLite数据库文件，包含所有历史记录数据</p>
                      </div>
                    </div>
                    <button
                      @click="downloadSqlite"
                      class="inline-flex w-full justify-center items-center sm:w-auto px-5 py-2.5 text-[0.6875rem] font-medium text-white md:text-sm bg-blue-500 rounded-xl hover:shadow-lg hover:shadow-blue-500/25 transition-all"
                    >
                      <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
                      </svg>
                      下载SQLite数据库
                    </button>
                  </div>
                </div>
              </div>

              <!-- 危险操作卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-red-100 dark:border-red-900/30 overflow-hidden">
                <div class="px-5 md:px-6 py-5">
                  <div class="flex items-center gap-3 mb-4">
                    <div class="w-10 h-10 rounded-xl bg-red-500 flex items-center justify-center">
                      <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
                      </svg>
                    </div>
                    <div>
                      <h2 class="text-base md:text-lg font-semibold text-red-600 dark:text-red-400">危险操作</h2>
                      <p class="text-[0.6875rem] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">此操作不可逆，请谨慎操作</p>
                    </div>
                  </div>
                  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 border border-red-100 dark:border-red-900/40 rounded-2xl bg-red-50/50 dark:bg-red-900/10">
                    <div>
                      <h4 class="text-sm font-medium text-red-700 dark:text-red-300">数据库重置</h4>
                      <p class="text-xs text-red-600 dark:text-red-300/80 mt-0.5">删除现有数据库并重新导入数据（此操作不可逆）</p>
                    </div>
                    <button
                      @click="handleResetDatabase"
                      class="inline-flex items-center px-5 py-2.5 text-[0.6875rem] font-medium text-white md:text-sm bg-red-500 rounded-xl hover:shadow-lg hover:shadow-red-500/25 transition-all"
                    >
                      <svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                      重置数据库
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- 关于页面 -->
            <div v-if="activeTab === 'about'" key="about" class="space-y-5 md:space-y-6">
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <div class="px-5 md:px-6 py-8 md:py-10 text-center relative overflow-hidden">
                  <div class="relative">
                    <img src="/logo.svg" class="w-16 h-16 md:w-20 md:h-20 mx-auto mb-4 object-contain drop-shadow-xl" alt="拾帧集" />
                    <h1 class="text-xl md:text-2xl font-bold">
                      <span class="text-accent">BiliHistory</span>
                    </h1>
                    <p class="text-xs md:text-sm text-gray-500 dark:text-gray-400 mt-2">哔哩哔哩历史记录管理与分析工具</p>
                  </div>
                </div>

                <div class="px-5 md:px-6 py-5 space-y-5 md:space-y-6 border-t border-gray-100 dark:border-gray-700/50">
                  <div>
                    <h2 class="mb-3 flex items-center text-[0.8125rem] font-medium text-gray-800 dark:text-gray-100 md:mb-4 md:text-base">
                      <svg class="mr-1.5 h-4 w-4 text-accent md:mr-2 md:h-5 md:w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      项目简介
                    </h2>
                    <p class="text-[0.75rem] text-gray-600 dark:text-gray-300 md:text-sm leading-relaxed">
                      此项目是一个哔哩哔哩历史记录管理与分析工具，帮助用户更好地管理和分析自己的B站观看历史。基于Vue 3构建，通过现代的界面设计提供强大的功能，包括历史记录查询、视频下载、数据分析等多项功能。
                    </p>
                  </div>
                </div>
              </div>

              <!-- 技术致谢卡片 -->
              <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700/50 overflow-hidden">
                <div class="px-5 md:px-6 py-5">
                  <h2 class="mb-4 flex items-center text-[0.8125rem] font-medium text-gray-800 dark:text-gray-100 md:text-base">
                    <svg class="mr-1.5 h-4 w-4 text-accent md:mr-2 md:h-5 md:w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                    技术致谢
                  </h2>
                  <ul class="space-y-2.5 text-[0.75rem] text-gray-600 dark:text-gray-300 md:text-sm list-none">
                    <li class="flex items-start gap-2">
                      <span class="text-accent mt-0.5">•</span>
                      <a href="https://github.com/2977094657/BiliHistoryFrontend" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline">BiliHistoryFrontend</a>
                      <span class="text-gray-500 dark:text-gray-400">- 原始前端项目</span>
                    </li>
                    <li class="flex items-start gap-2">
                      <span class="text-accent mt-0.5">•</span>
                      <a href="https://github.com/2977094657/BilibiliHistoryFetcher" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline">BilibiliHistoryFetcher</a>
                      <span class="text-gray-500 dark:text-gray-400">- 原始后端项目</span>
                    </li>
                    <li class="flex items-start gap-2">
                      <span class="text-accent mt-0.5">•</span>
                      <a href="https://github.com/SocialSisterYi/bilibili-API-collect" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline">bilibili-API-collect</a>
                      <span class="text-gray-500 dark:text-gray-400">- 没有它就没有这个项目</span>
                    </li>
                    <li class="flex items-start gap-2">
                      <span class="text-accent mt-0.5">•</span>
                      <a href="https://github.com/Felix2yu/bili-dl" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline">bili-dl</a>
                      <span class="text-gray-500 dark:text-gray-400">- B站视频下载库</span>
                    </li>
                    <li class="flex items-start gap-2">
                      <span class="text-accent mt-0.5">•</span>
                      <a href="https://github.com/zhw2590582/ArtPlayer" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline">ArtPlayer</a>
                      <span class="text-gray-500 dark:text-gray-400">- 强大且灵活的HTML5视频播放器</span>
                    </li>
                    <li class="flex items-start gap-2">
                      <span class="text-accent mt-0.5">•</span>
                      <a href="https://www.aicu.cc/" target="_blank" rel="noopener noreferrer" class="text-accent hover:underline">aicu.cc</a>
                      <span class="text-gray-500 dark:text-gray-400">- 第三方B站用户评论API</span>
                    </li>
                    <li class="flex items-start gap-2">
                      <span class="text-accent mt-0.5">•</span>
                      <span>所有贡献者</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { showNotify } from 'vant'
import 'vant/es/notify/style'
import 'vant/es/toast/style'
import 'vant/es/dialog/style'
import {
  exportHistory,
  downloadDatabase,
  resetDatabase,
  getAvailableYears,
  importSqliteData,
  getMcpConfig,
  updateMcpConfig,
  getIntegrityCheckConfig,
  updateIntegrityCheckConfig,
  getSyncConfig,
  updateSyncConfig,
  getAppearanceConfig,
  updateAppearanceConfig
} from '~/utils/api'
import ShoutrrrSettings from './ShoutrrrSettings.vue'
import SettingToggle from './SettingToggle.vue'
import { showDialog } from 'vant'
import { useRoute } from 'vue-router'
import { THEME_PRESETS, FONT_SIZE_PRESETS, useDarkMode } from '~/stores/darkMode'

import { storeToRefs } from 'pinia'

// 设置选项卡
const settingTabs = [
  {
    key: 'basic',
    label: '基础设置',
    icon: '<svg class="text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2m-2-4h.01M17 16h.01" /></svg>'
  },
  {
    key: 'data',
    label: '数据管理',
    icon: '<svg class="text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" /></svg>'
  },
  {
    key: 'about',
    label: '关于',
    icon: '<svg class="text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>'
  }
]

const route = useRoute()
const activeTab = ref('basic')

// 监听路由参数变化，切换标签页
watch(() => route.query.tab, (newTab) => {
  if (newTab && settingTabs.some(tab => tab.key === newTab)) {
    activeTab.value = newTab
  }
}, { immediate: true })

const availableYears = ref([])
const isExporting = ref(false)

// 导出选项
const exportOptions = ref({
  year: new Date().getFullYear(),
  month: null,
  startDate: '',
  endDate: '',
  exportType: 'year' // 默认导出全年数据
})
const useLocalImages = ref(localStorage.getItem('useLocalImages') === 'true')
const DEFAULT_MCP_CONFIG = {
  enabled: false,
  path: '/mcp',
  auth_enabled: true,
  token: '',
  token_configured: false,
  max_page_size: 100,
  server_url: '',
  mcp_url: '',
  skill_content: '',
  restart_required: false
}
const mcpConfig = ref({ ...DEFAULT_MCP_CONFIG })
const mcpConfigAvailable = ref(false)
const isMcpConfigLoading = ref(true)
const isMcpConfigSaving = ref(false)

const trimTrailingSlash = (value) => (value || '').trim().replace(/\/+$/, '')

const normalizeMcpPath = (path) => {
  let normalizedPath = (path || '/mcp').trim()
  if (!normalizedPath.startsWith('/')) {
    normalizedPath = `/${normalizedPath}`
  }
  if (normalizedPath.length > 1) {
    normalizedPath = normalizedPath.replace(/\/+$/, '')
  }
  return normalizedPath || '/mcp'
}

const mcpUrl = computed(() => {
  const baseUrl = trimTrailingSlash(typeof window !== 'undefined' ? window.location.origin : '')
  const path = normalizeMcpPath(mcpConfig.value.path)
  return path === '/' ? `${baseUrl}/` : `${baseUrl}${path}/`
})

const mcpConnectionPrompt = computed(() => {
  const authLine = mcpConfig.value.auth_enabled
    ? `Authorization: Bearer ${mcpConfig.value.token || '<token>'}`
    : 'Authorization: not required'

  return [
    '请通过 MCP Streamable HTTP 连接我的 BilibiliHistoryFetcher 只读服务。',
    '',
    `MCP URL: ${mcpUrl.value}`,
    authLine,
    '',
    '连接后请先读取以下 Resources：',
    '- bili://project/overview',
    '- bili://project/data-status',
    '- bili://project/tool-guide',
    '',
    '使用规则：',
    '- 这是只读 MCP，不要请求同步、下载、删除、登录、重置数据库或修改配置。',
    '- 查询明细时必须分页，优先使用统计/摘要工具，再按需读取 records。',
    '- 观看历史属于隐私数据，只在当前任务需要时读取。'
  ].join('\n')
})

// 同步已删除记录
const syncDeleted = ref(false)

// 处理同步已删除记录变更
const handleSyncDeletedChange = async () => {
  try {
    const response = await updateSyncConfig({ sync_deleted: syncDeleted.value })
    if (response.data && response.data.success) {
      showNotify({
        type: 'success',
        message: syncDeleted.value ? '已开启同步已删除记录' : '已关闭同步已删除记录'
      })
    } else {
      throw new Error(response.data?.message || '更新配置失败')
    }
  } catch (error) {
    console.error('更新同步配置失败:', error)
    showNotify({
      type: 'danger',
      message: `更新配置失败: ${error.message}`
    })
    // 恢复原值
    syncDeleted.value = !syncDeleted.value
  }
}

// 同步删除B站历史记录
const syncDeleteToBilibili = ref(false)

// 处理同步删除B站历史记录变更
const handleSyncDeleteToBilibiliChange = async () => {
  try {
    const response = await updateSyncConfig({ sync_delete_to_bilibili: syncDeleteToBilibili.value })
    if (response.data && response.data.success) {
      showNotify({
        type: 'success',
        message: syncDeleteToBilibili.value ? '已开启同步删除B站历史记录' : '已关闭同步删除B站历史记录'
      })
    } else {
      throw new Error(response.data?.message || '更新配置失败')
    }
  } catch (error) {
    console.error('更新同步配置失败:', error)
    showNotify({
      type: 'danger',
      message: `更新配置失败: ${error.message}`
    })
    // 恢复原值
    syncDeleteToBilibili.value = !syncDeleteToBilibili.value
  }
}

// 深色模式设置
const appearanceStore = useDarkMode()
const { darkMode, themeColor, fontSize } = storeToRefs(appearanceStore)
const darkModeOptions = [
  { value: 'system', label: '跟随系统' },
  { value: 'light', label: '浅色' },
  { value: 'dark', label: '深色' },
]

// 主题色设置
const themePresets = THEME_PRESETS

// 字体大小设置
const fontSizePresets = FONT_SIZE_PRESETS

const handleFontSizeChange = async (sizeId) => {
  const previousSize = fontSize.value
  try {
    await appearanceStore.setFontSize(sizeId)
    const sizeLabels = { small: '小号字体', default: '默认字体', large: '大号字体' }
    showNotify({
      type: 'success',
      message: `已切换到${sizeLabels[sizeId]}`
    })
  } catch (error) {
    console.error('更新字体大小失败:', error)
    showNotify({
      type: 'danger',
      message: `更新字体大小失败: ${error.message}`
    })
    await appearanceStore.setFontSize(previousSize)
  }
}

// 处理深色模式变更
const handleDarkModeChange = async () => {
  const previousMode = darkMode.value
  try {
    await appearanceStore.setDarkMode(darkMode.value)
    const modeLabels = { system: '跟随系统', light: '浅色模式', dark: '深色模式' }
    showNotify({
      type: 'success',
      message: `已切换到${modeLabels[darkMode.value]}`
    })
  } catch (error) {
    console.error('更新外观配置失败:', error)
    showNotify({
      type: 'danger',
      message: `更新配置失败: ${error.message}`
    })
    await appearanceStore.setDarkMode(previousMode)
  }
}

// 处理主题色变更
const handleThemeColorChange = async (themeId) => {
  const previousTheme = themeColor.value
  try {
    await appearanceStore.setThemeColor(themeId)
    const theme = themePresets.find(t => t.id === themeId)
    showNotify({
      type: 'success',
      message: `已切换到${theme?.name || '自定义'}主题`
    })
  } catch (error) {
    console.error('更新主题色失败:', error)
    showNotify({
      type: 'danger',
      message: `更新主题色失败: ${error.message}`
    })
    await appearanceStore.setThemeColor(previousTheme)
  }
}

// 启动时数据完整性校验
const checkIntegrityOnStartup = ref(true)

// 处理数据完整性校验设置变更
const handleIntegrityCheckChange = async () => {
  try {
    const response = await updateIntegrityCheckConfig(checkIntegrityOnStartup.value)
    if (response.data && response.data.success) {
      showNotify({
        type: 'success',
        message: checkIntegrityOnStartup.value ? '已开启启动时数据完整性校验' : '已关闭启动时数据完整性校验'
      })
    } else {
      throw new Error(response.data?.message || '更新配置失败')
    }
  } catch (error) {
    console.error('更新数据完整性校验配置失败:', error)
    showNotify({
      type: 'danger',
      message: `更新配置失败: ${error.message}`
    })
    // 恢复原值
    checkIntegrityOnStartup.value = !checkIntegrityOnStartup.value
  }
}

// 首页默认布局设置 - 网格布局或列表布局
const isGridLayout = ref(localStorage.getItem('defaultLayout') === 'list' ? false : true) // 默认为网格视图

// 处理布局变更
const handleLayoutChange = () => {
  // 更新localStorage，保存用户选择的布局模式
  const newLayout = isGridLayout.value ? 'grid' : 'list'
  localStorage.setItem('defaultLayout', newLayout)

  // 触发全局事件，通知其他组件更新布局
  try {
    const event = new CustomEvent('layout-setting-changed', {
      detail: { layout: newLayout }
    })
    window.dispatchEvent(event)
    console.log('已触发布局设置更新事件:', newLayout)
  } catch (error) {
    console.error('触发布局设置更新事件失败:', error)
  }

  showNotify({
    type: 'success',
    message: `已切换到${isGridLayout.value ? '网格' : '列表'}视图`
  })
}

// 侧边栏显示设置
const showSidebar = ref(localStorage.getItem('showSidebar') !== 'false') // 默认为true

// 处理侧边栏设置变更
const handleSidebarChange = () => {
  localStorage.setItem('showSidebar', showSidebar.value.toString())

  // 触发全局事件，通知侧边栏组件更新设置
  try {
    const event = new CustomEvent('sidebar-setting-changed', {
      detail: { showSidebar: showSidebar.value }
    })
    window.dispatchEvent(event)
    console.log('已触发侧边栏设置更新事件:', showSidebar.value)
  } catch (error) {
    console.error('触发侧边栏设置更新事件失败:', error)
  }

  showNotify({
    type: 'success',
    message: `已${showSidebar.value ? '启用' : '禁用'}侧边栏显示`
  })
}

onMounted(async () => {
  console.log('Settings组件开始挂载')

  try {

    // 监听侧边栏切换事件
    window.addEventListener('sidebar-toggle-changed', handleSidebarToggleEvent)

    // 监听布局切换事件
    window.addEventListener('layout-changed', handleLayoutChangedEvent)

    // 获取可用年份数据
    await getAvailableYears().then(response => {
      if (response.data.status === 'success') {
        availableYears.value = response.data.data
        if (availableYears.value.length > 0) {
          exportOptions.value.year = availableYears.value[0]
        }
      }
    }).catch(error => {
      console.error('获取可用年份失败:', error)
    })

    await Promise.all([
      (async () => {
        console.log('开始初始化MCP配置')
        await initMcpConfig()
        console.log('MCP配置初始化完成')
      })(),
      (async () => {
        console.log('开始获取可用年份')
        try {
          const response = await getAvailableYears()
          console.log('获取年份响应:', response.data)
          if (response.data.status === 'success') {
            availableYears.value = response.data.data.sort((a, b) => b - a)
            if (availableYears.value.length > 0) {
              // 设置导出选项的年份
              exportOptions.value.year = availableYears.value[0]
            }
            console.log('获取可用年份成功:', availableYears.value)
          } else {
            throw new Error(response.data.message || '获取年份列表失败')
          }
        } catch (error) {
          console.error('获取可用年份失败:', error)
          showNotify({
            type: 'danger',
            message: '获取年份列表失败'
          })
          // 设置当前年份作为默认值
          const currentYear = new Date().getFullYear()
          availableYears.value = [currentYear]

          // 重置导出选项
          exportOptions.value = {
            year: currentYear,
            month: null,
            startDate: '',
            endDate: '',
            exportType: 'year'
          }
        }
      })(),
      (async () => {
        console.log('开始获取数据完整性校验配置')
        try {
          const response = await getIntegrityCheckConfig()
          if (response.data && response.data.success) {
            checkIntegrityOnStartup.value = response.data.check_on_startup
            console.log('数据完整性校验配置获取成功:', checkIntegrityOnStartup.value)
          } else {
            throw new Error(response.data?.message || '获取配置失败')
          }
        } catch (error) {
          console.error('获取数据完整性校验配置失败:', error)
          // 使用默认值
          checkIntegrityOnStartup.value = true
        }
        console.log('数据完整性校验配置获取完成')
      })(),
      (async () => {
        console.log('开始获取同步删除配置')
        try {
          const response = await getSyncConfig()
          if (response.data && response.data.success) {
            syncDeleted.value = response.data.sync_deleted
            syncDeleteToBilibili.value = response.data.sync_delete_to_bilibili
            console.log('同步删除配置获取成功:', response.data)
          } else {
            throw new Error(response.data?.message || '获取配置失败')
          }
        } catch (error) {
          console.error('获取同步删除配置失败:', error)
          // 使用默认值
          syncDeleted.value = false
          syncDeleteToBilibili.value = false
        }
        console.log('同步删除配置获取完成')
      })(),
      (async () => {
        console.log('同步外观配置')
        try {
          await appearanceStore.initDarkMode()
        } catch (error) {
          console.error('外观配置初始化失败:', error)
        }
        console.log('外观配置同步完成')
      })()
    ])
    console.log('Settings组件初始化完成')
  } catch (error) {
    console.error('Settings组件初始化失败:', error)
  }
})

// 导出并下载Excel
const exportAndDownloadExcel = async () => {
  if (isExporting.value) return

  try {
    // 准备导出参数
    let exportParams = {}

    // 根据导出类型设置参数
    switch (exportOptions.value.exportType) {
      case 'month':
        // 检查月份是否选择
        if (!exportOptions.value.month) {
          showNotify({
            type: 'danger',
            message: '请选择要导出的月份'
          })
          return
        }
        exportParams = {
          year: exportOptions.value.year,
          month: exportOptions.value.month
        }
        break

      case 'dateRange':
        // 验证日期范围
        if (!exportOptions.value.startDate || !exportOptions.value.endDate) {
          showNotify({
            type: 'danger',
            message: '请选择完整的日期范围'
          })
          return
        }

        const startDate = new Date(exportOptions.value.startDate)
        const endDate = new Date(exportOptions.value.endDate)
        if (startDate > endDate) {
          showNotify({
            type: 'danger',
            message: '开始日期不能晚于结束日期'
          })
          return
        }

        // 只传递日期范围参数，不传递年份参数
        exportParams = {
          start_date: exportOptions.value.startDate,
          end_date: exportOptions.value.endDate
        }
        break

      case 'year':
      default:
        // 全年数据，只需要year参数
        exportParams = {
          year: exportOptions.value.year
        }
        break
    }

    isExporting.value = true
    showNotify({
      type: 'primary',
      message: '正在导出数据...'
    })

    console.log('导出选项:', exportParams)
    const response = await exportHistory(exportParams)
    console.log('导出响应:', response)

    if (response.success) {
      showNotify({
        type: 'success',
        message: '导出成功，下载已开始'
      })
    } else {
      throw new Error('导出失败')
    }
  } catch (error) {
    console.error('导出错误:', error)
    let errorMessage = error.message

    // 尝试获取服务器返回的错误信息
    if (error.response && error.response.data) {
      if (error.response.data.detail) {
        errorMessage = error.response.data.detail
      } else if (typeof error.response.data === 'string') {
        errorMessage = error.response.data
      }
    }

    showNotify({
      type: 'danger',
      message: `操作失败：${errorMessage}`
    })
  } finally {
    isExporting.value = false
  }
}

// 下载SQLite数据库
const downloadSqlite = async () => {
  try {
    await downloadDatabase()
  } catch (error) {
    showNotify({
      type: 'danger',
      message: `下载失败：${error.message}`
    })
  }
}

// 处理数据库重置
const handleResetDatabase = () => {
  showDialog({
    title: '危险操作确认',
    message: '此操作将删除现有数据库并重新导入数据。此操作不可逆，确定要继续吗？',
    showCancelButton: true,
    confirmButtonText: '确定重置',
    cancelButtonText: '取消',
    confirmButtonColor: '#dc2626'
  }).then(async (result) => {
    if (result === 'confirm') {
      try {
        showNotify({
          type: 'warning',
          message: '正在重置数据库...'
        })

        // 重置数据库
        const resetResponse = await resetDatabase()
        if (resetResponse.data.status === 'success') {
          showNotify({
            type: 'success',
            message: '数据库已重置，正在重新导入数据...'
          })

          // 重新导入数据
          try {
            const importResponse = await importSqliteData()
            if (importResponse.data.status === 'success') {
              showNotify({
                type: 'success',
                message: '数据导入完成，页面即将刷新'
              })
              // 等待1秒后刷新页面，确保用户看到成功提示
              setTimeout(() => {
                window.location.reload()
              }, 1000)
            } else {
              throw new Error(importResponse.data.message || '数据导入失败')
            }
          } catch (importError) {
            showNotify({
              type: 'danger',
              message: `数据导入失败：${importError.message}`
            })
          }
        }
      } catch (error) {
        showNotify({
          type: 'danger',
          message: `重置失败：${error.message}`
        })
      }
    }
  })
}

// 处理图片源变更
const handleImageSourceChange = () => {
  localStorage.setItem('useLocalImages', useLocalImages.value.toString())
  showNotify({
    type: 'success',
    message: `已${useLocalImages.value ? '启用' : '禁用'}本地图片源`
  })
  // 刷新页面以应用新设置
  window.location.reload()
}

// 初始化MCP配置
const initMcpConfig = async () => {
  isMcpConfigLoading.value = true
  try {
    const response = await getMcpConfig()
    if (response.data && response.data.status === 'success') {
      mcpConfig.value = {
        ...DEFAULT_MCP_CONFIG,
        ...response.data,
        path: normalizeMcpPath(response.data.path)
      }
      mcpConfigAvailable.value = true
    } else {
      throw new Error(response.data?.message || '获取MCP配置失败')
    }
  } catch (error) {
    console.error('获取MCP配置失败:', error)
    mcpConfig.value = { ...DEFAULT_MCP_CONFIG }
    mcpConfigAvailable.value = false
  } finally {
    isMcpConfigLoading.value = false
  }
}

// 保存MCP开关配置
const handleMcpEnabledChange = async () => {
  if (isMcpConfigSaving.value) return

  const nextEnabled = mcpConfig.value.enabled
  isMcpConfigSaving.value = true

  try {
    const response = await updateMcpConfig({ enabled: nextEnabled })
    if (response.data && response.data.status === 'success') {
      mcpConfig.value = {
        ...mcpConfig.value,
        ...response.data,
        path: normalizeMcpPath(response.data.path)
      }
      mcpConfigAvailable.value = true
      showNotify({
        type: response.data.restart_required ? 'warning' : 'success',
        message: response.data.restart_required ? 'MCP配置已保存，重启后端后生效' : 'MCP配置已保存，已立即生效'
      })
    } else {
      throw new Error(response.data?.message || '保存MCP配置失败')
    }
  } catch (error) {
    console.error('保存MCP配置失败:', error)
    mcpConfig.value.enabled = !nextEnabled
    showNotify({
      type: 'danger',
      message: `保存MCP配置失败：${error.message || '未知错误'}`
    })
  } finally {
    isMcpConfigSaving.value = false
  }
}

const fallbackCopyText = (text) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  const copied = document.execCommand('copy')
  document.body.removeChild(textarea)
  return copied
}

// 复制MCP连接信息
const copyText = async (text, label = '内容') => {
  if (!text) {
    showNotify({
      type: 'warning',
      message: `${label}为空，无法复制`
    })
    return
  }

  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
    } else if (!fallbackCopyText(text)) {
      throw new Error('复制失败，请手动复制')
    }
    showNotify({
      type: 'success',
      message: `${label}已复制`
    })
  } catch (error) {
    console.error('复制失败:', error)
    showNotify({
      type: 'danger',
      message: error.message || '复制失败，请手动复制'
    })
  }
}

// 处理侧边栏切换事件
const handleSidebarToggleEvent = (event) => {
  if (event.detail && typeof event.detail.showSidebar === 'boolean') {
    showSidebar.value = event.detail.showSidebar
  }
}

// 在script setup部分添加卸载功能
onUnmounted(() => {
  // 移除事件监听
  window.removeEventListener('sidebar-toggle-changed', handleSidebarToggleEvent)
  window.removeEventListener('layout-changed', handleLayoutChangedEvent)
})

// 处理布局变更事件 - 从首页接收的布局变化
const handleLayoutChangedEvent = (event) => {
  if (event.detail && typeof event.detail.layout === 'string') {
    isGridLayout.value = event.detail.layout === 'grid'
  }
}
</script>
