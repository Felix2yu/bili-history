<template>
  <div class="min-h-screen">
    <!-- 背景装饰 -->
    <div class="fixed inset-0 -z-10 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-96 h-96 bg-accent/10 rounded-full blur-3xl"></div>
      <div class="absolute top-1/3 -left-40 w-96 h-96 bg-accent/5 rounded-full blur-3xl"></div>
    </div>

    <div class="max-w-7xl mx-auto px-4 md:px-6 py-6 md:py-8">
      <!-- 页面标题区 -->
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6 md:mb-8">
        <div class="flex items-center gap-4">
          <div class="relative">
            <div class="w-12 h-12 md:w-14 md:h-14 rounded-2xl bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-lg shadow-accent/25">
              <svg class="w-6 h-6 md:w-7 md:h-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <span class="absolute -top-1 -right-1 w-4 h-4 rounded-full bg-green-500 border-2 border-white dark:border-gray-900 animate-pulse-soft"></span>
          </div>
          <div>
            <h1 class="text-xl md:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">计划任务</h1>
            <p class="text-xs md:text-sm text-gray-500 dark:text-gray-400 mt-0.5">自动化调度，让数据同步有条不紊</p>
          </div>
        </div>
        <button @click="openCreateTaskModal"
          class="group relative inline-flex items-center justify-center gap-2 px-4 md:px-5 py-2.5 md:py-3 rounded-xl font-medium text-sm text-white transition-all duration-300 overflow-hidden"
          :class="'bg-gradient-to-r from-accent to-accent/80 hover:shadow-lg hover:shadow-accent/30 hover:-translate-y-0.5 active:translate-y-0'">
          <span class="absolute inset-0 bg-white/20 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></span>
          <svg class="w-4 h-4 md:w-5 md:h-5 relative" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          <span class="relative">新建任务</span>
        </button>
      </div>

      <!-- 统计卡片 -->
      <div v-if="tasks.length > 0" class="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4 mb-6 md:mb-8">
        <div v-for="stat in stats" :key="stat.label"
          class="group relative overflow-hidden rounded-2xl p-4 md:p-5 transition-all duration-300 hover:-translate-y-1"
          :class="stat.bgClass">
          <div class="absolute -right-4 -top-4 w-24 h-24 rounded-full opacity-20" :class="stat.dotClass"></div>
          <div class="relative">
            <div class="flex items-center justify-between mb-2">
              <div class="w-9 h-9 md:w-10 md:h-10 rounded-xl flex items-center justify-center" :class="stat.iconBg">
                <svg v-html="stat.icon" class="w-4 h-4 md:w-5 md:h-5" :class="stat.iconColor"></svg>
              </div>
              <span v-if="stat.trend" class="text-[10px] md:text-xs font-medium px-2 py-0.5 rounded-full" :class="stat.trendClass">{{ stat.trend }}</span>
            </div>
            <div class="text-2xl md:text-3xl font-bold" :class="stat.valueColor">{{ stat.value }}</div>
            <div class="text-[11px] md:text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ stat.label }}</div>
          </div>
        </div>
      </div>

      <!-- 搜索和筛选栏 -->
      <div v-if="tasks.length > 0" class="flex flex-col sm:flex-row gap-3 mb-5 md:mb-6">
        <div class="relative flex-1">
          <svg class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 md:w-5 md:h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
          </svg>
          <input v-model="searchQuery" type="text" placeholder="搜索任务名称、端点..."
            class="w-full pl-10 md:pl-12 pr-4 py-2.5 md:py-3 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all" />
        </div>
        <div class="flex gap-2 overflow-x-auto pb-1 sm:pb-0">
          <button v-for="f in filterOptions" :key="f.value" @click="activeFilter = f.value"
            class="flex-shrink-0 px-3 md:px-4 py-2 md:py-2.5 text-xs md:text-sm font-medium rounded-xl transition-all duration-200"
            :class="activeFilter === f.value
              ? 'bg-accent text-white shadow-md shadow-accent/25'
              : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 border border-gray-200 dark:border-gray-700'">
            {{ f.label }}
          </button>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="flex justify-center items-center py-16 md:py-24">
        <div class="flex flex-col items-center gap-4">
          <div class="relative">
            <div class="w-12 h-12 md:w-16 md:h-16 rounded-full border-2 border-accent/20"></div>
            <div class="absolute inset-0 w-12 h-12 md:w-16 md:h-16 rounded-full border-2 border-accent border-t-transparent animate-spin"></div>
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400">加载任务中...</p>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else-if="filteredTasks.length === 0 && tasks.length === 0" class="glass-card">
        <div class="relative py-16 md:py-20 text-center">
          <div class="relative mx-auto w-24 h-24 md:w-32 md:h-32 mb-6">
            <div class="absolute inset-0 rounded-3xl bg-gradient-to-br from-accent/20 to-accent/5 animate-pulse-soft"></div>
            <div class="absolute inset-2 rounded-3xl bg-white dark:bg-gray-800 flex items-center justify-center">
              <svg class="w-12 h-12 md:w-16 md:h-16 text-accent/60" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
          <h3 class="text-lg md:text-xl font-semibold text-gray-900 dark:text-white mb-2">
            {{ searchQuery ? '没有找到匹配的任务' : '开启自动化之旅' }}
          </h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm mx-auto mb-6">
            {{ searchQuery ? '尝试更换搜索关键词或清除筛选条件' : '创建您的第一个计划任务，让数据同步自动完成' }}
          </p>
          <button v-if="!searchQuery" @click="openCreateTaskModal"
            class="inline-flex items-center gap-2 px-5 py-2.5 bg-accent text-white text-sm font-medium rounded-xl hover:bg-accent/90 transition-colors shadow-lg shadow-accent/25">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            创建第一个任务
          </button>
        </div>
      </div>

      <!-- 任务列表 -->
      <div v-else class="space-y-3 md:space-y-4">
        <TransitionGroup name="task-list">
          <div v-for="task in filteredTasks" :key="task.task_id"
            class="group relative glass-card overflow-hidden transition-all duration-300 hover:shadow-xl hover:-translate-y-0.5">
            <!-- 左侧装饰条 -->
            <div class="absolute left-0 top-0 bottom-0 w-1 transition-all duration-300"
              :class="getScheduleTypeColor(task.config?.schedule_type)"></div>

            <div class="p-4 md:p-5">
              <!-- 顶部：任务名称 + 状态 -->
              <div class="flex items-start justify-between gap-3 mb-4">
                <div class="flex items-start gap-3 min-w-0 flex-1">
                  <div class="w-10 h-10 md:w-11 md:h-11 rounded-xl flex-shrink-0 flex items-center justify-center transition-transform duration-300 group-hover:scale-105"
                    :class="getScheduleTypeBg(task.config?.schedule_type)">
                    <svg class="w-5 h-5 md:w-[22px] md:h-[22px]" :class="getScheduleTypeText(task.config?.schedule_type)" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        v-if="task.config?.schedule_type === 'daily'"
                        d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round"
                        v-else-if="task.config?.schedule_type === 'chain'"
                        d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                      <path stroke-linecap="round" stroke-linejoin="round"
                        v-else-if="task.config?.schedule_type === 'once'"
                        d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" v-else
                        d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2 flex-wrap">
                      <h3 class="text-sm md:text-base font-semibold text-gray-900 dark:text-white truncate">
                        {{ task.config?.name || task.task_id }}
                      </h3>
                      <span v-if="task.sub_tasks && task.sub_tasks.length > 0"
                        class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] md:text-xs font-medium bg-accent/10 text-accent">
                        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                        </svg>
                        {{ task.sub_tasks.length }} 子任务
                      </span>
                    </div>
                    <p class="text-[11px] md:text-xs text-gray-400 dark:text-gray-500 mt-0.5 font-mono truncate">
                      {{ task.config?.endpoint || '-' }}
                    </p>
                  </div>
                </div>

                <!-- 开关 -->
                <button @click="toggleTaskEnabled(task.task_id, !task.config?.enabled)"
                  class="relative flex-shrink-0 w-11 h-6 md:w-12 md:h-7 rounded-full transition-all duration-300 focus:outline-none"
                  :class="task.config?.enabled ? 'bg-accent' : 'bg-gray-300 dark:bg-gray-600'">
                  <span class="absolute top-0.5 left-0.5 w-5 h-5 md:w-6 md:h-6 bg-white rounded-full shadow-md transition-all duration-300 flex items-center justify-center"
                    :class="{ 'translate-x-5 md:translate-x-6': task.config?.enabled }">
                    <svg v-if="task.config?.enabled" class="w-3 h-3 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  </span>
                </button>
              </div>

              <!-- 信息行 -->
              <div class="flex flex-wrap items-center gap-2 md:gap-3 mb-4">
                <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-[11px] md:text-xs font-medium"
                  :class="getScheduleTypeBadge(task.config?.schedule_type)">
                  {{ getScheduleTypeLabel(task.config?.schedule_type) }}
                </span>
                <span class="inline-flex items-center gap-1 text-[11px] md:text-xs text-gray-500 dark:text-gray-400">
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {{ getScheduleTimeDisplay(task) }}
                </span>
                <span v-if="task.execution?.next_run && task.task_type === 'main'" class="inline-flex items-center gap-1 text-[11px] md:text-xs text-gray-500 dark:text-gray-400">
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
                  </svg>
                  下次: {{ task.execution.next_run }}
                </span>
              </div>

              <!-- 统计条 + 操作 -->
              <div class="flex items-center justify-between gap-3">
                <!-- 执行统计 -->
                <div class="flex items-center gap-3 min-w-0 flex-1">
                  <template v-if="task.execution?.total_runs > 0">
                    <div class="flex items-center gap-1.5">
                      <div class="w-2 h-2 rounded-full"
                        :class="{
                          'bg-green-500': task.execution.success_rate >= 90,
                          'bg-amber-500': task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'bg-red-500': task.execution.success_rate < 60
                        }"></div>
                      <span class="text-xs font-semibold"
                        :class="{
                          'text-green-600 dark:text-green-400': task.execution.success_rate >= 90,
                          'text-amber-600 dark:text-amber-400': task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'text-red-600 dark:text-red-400': task.execution.success_rate < 60
                        }">
                        {{ Math.round(task.execution.success_rate) }}%
                      </span>
                    </div>
                    <div class="h-1.5 flex-1 max-w-[120px] bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                      <div class="h-full rounded-full transition-all duration-500"
                        :class="{
                          'bg-green-500': task.execution.success_rate >= 90,
                          'bg-amber-500': task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'bg-red-500': task.execution.success_rate < 60
                        }"
                        :style="{ width: `${Math.round(task.execution.success_rate)}%` }"></div>
                    </div>
                    <span class="text-[10px] md:text-xs text-gray-400 whitespace-nowrap">{{ task.execution.total_runs }}次</span>
                  </template>
                  <span v-else class="text-[10px] md:text-xs text-gray-400">尚未执行</span>
                </div>

                <!-- 操作按钮 -->
                <div class="flex items-center gap-0.5 md:gap-1">
                  <button @click="executeTask(task.task_id)"
                    class="p-2 md:p-2.5 rounded-lg text-green-600 dark:text-green-400 hover:bg-green-50 dark:hover:bg-green-900/20 transition-all duration-200 hover:scale-105 active:scale-95"
                    title="立即执行">
                    <svg class="w-4 h-4 md:w-[18px] md:h-[18px]" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M8 5v14l11-7z" />
                    </svg>
                  </button>
                  <button @click="openTaskDetailModal(task.task_id)"
                    class="p-2 md:p-2.5 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-gray-200 transition-all duration-200 hover:scale-105 active:scale-95"
                    title="查看详情">
                    <svg class="w-4 h-4 md:w-[18px] md:h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                  </button>
                  <button @click="openEditTaskModal(task.task_id)"
                    class="p-2 md:p-2.5 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-gray-200 transition-all duration-200 hover:scale-105 active:scale-95"
                    title="编辑">
                    <svg class="w-4 h-4 md:w-[18px] md:h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                    </svg>
                  </button>
                  <button @click="openCreateSubTaskModal(task.task_id)"
                    class="p-2 md:p-2.5 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-gray-200 transition-all duration-200 hover:scale-105 active:scale-95"
                    title="添加子任务">
                    <svg class="w-4 h-4 md:w-[18px] md:h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                    </svg>
                  </button>
                  <button @click="confirmDeleteTask(task.task_id)"
                    class="p-2 md:p-2.5 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-500 transition-all duration-200 hover:scale-105 active:scale-95"
                    title="删除">
                    <svg class="w-4 h-4 md:w-[18px] md:h-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.034-2.09 1.02-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                    </svg>
                  </button>
                </div>
              </div>

              <!-- 子任务展开 -->
              <div v-if="task.sub_tasks && task.sub_tasks.length > 0" class="mt-4">
                <button @click="task.isExpanded = !task.isExpanded"
                  class="flex items-center gap-2 text-[11px] md:text-xs text-gray-500 dark:text-gray-400 hover:text-accent transition-colors">
                  <svg class="w-3.5 h-3.5 transition-transform duration-200" :class="{ 'rotate-90': task.isExpanded }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                  {{ task.isExpanded ? '收起' : '展开' }} {{ task.sub_tasks.length }} 个子任务
                </button>
              </div>
            </div>

            <!-- 子任务列表 -->
            <Transition name="subtask">
              <div v-if="task.sub_tasks && task.sub_tasks.length > 0 && task.isExpanded"
                class="border-t border-gray-100 dark:border-gray-700/50 bg-gray-50/50 dark:bg-gray-800/30">
                <div v-for="(sub, idx) in task.sub_tasks" :key="sub.task_id"
                  class="px-4 md:px-5 py-3 flex items-center justify-between gap-3 transition-colors"
                  :class="{ 'border-b border-gray-100 dark:border-gray-700/50': idx < task.sub_tasks.length - 1 }">
                  <div class="flex items-center gap-2.5 min-w-0 flex-1">
                    <svg class="w-4 h-4 text-accent/50 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round"
                        d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                    </svg>
                    <div class="min-w-0 flex-1">
                      <span class="text-xs font-medium text-gray-700 dark:text-gray-300 truncate block">
                        {{ sub.config?.name || sub.task_id }}
                      </span>
                      <span class="text-[10px] text-gray-400 truncate block">链式 · 依赖上一任务</span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <button @click="toggleTaskEnabled(sub.task_id, !sub.config?.enabled)"
                      class="relative w-9 h-5 rounded-full transition-colors duration-200"
                      :class="sub.config?.enabled ? 'bg-accent' : 'bg-gray-300 dark:bg-gray-600'">
                      <span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform duration-200"
                        :class="{ 'translate-x-4': sub.config?.enabled }"></span>
                    </button>
                    <button @click="openTaskDetailModal(sub.task_id)"
                      class="p-1.5 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round"
                          d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                      </svg>
                    </button>
                    <button @click="confirmDeleteTask(sub.task_id, task.task_id)"
                      class="p-1.5 rounded text-gray-400 hover:text-red-500 transition-colors">
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round"
                          d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.034-2.09 1.02-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </Transition>
          </div>
        </TransitionGroup>
      </div>
    </div>

    <!-- 弹窗组件 -->
    <TaskDetail v-model:show="showTaskDetailModal" :task="currentTask" @view-history="fetchTaskHistory"
      @edit-task="openEditTaskModal" @execute-task="executeTask" @toggle-enabled="toggleTaskEnabled"
      @delete-task="confirmDeleteTask" @refresh="fetchTasks" />
    <TaskHistory v-model:show="showTaskHistoryModal" :task-id="currentTask?.task_id"
      :task-name="currentTask?.config?.name || currentTask?.task_id" />
    <TaskForm v-model:show="showTaskFormModal" :is-editing="isEditing" :task-id="currentTask?.task_id"
      :parent-task-id="parentTaskId" :tasks="tasks" @task-saved="fetchTasks" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAsyncData } from '#imports'
import { showNotify, showDialog } from 'vant'
import 'vant/es/dialog/style'
import 'vant/es/notify/style'
import {
  getAllSchedulerTasks,
  getSchedulerTaskDetail,
  executeSchedulerTask,
  getTaskHistory,
  setTaskEnabled,
  deleteSchedulerTask,
  deleteSubTask
} from '~/utils/api'
import TaskForm from '../scheduler/TaskForm.vue'
import TaskDetail from '../scheduler/TaskDetail.vue'
import TaskHistory from '../scheduler/TaskHistory.vue'

const loading = ref(false)
const tasks = ref([])
const currentTask = ref(null)
const showTaskFormModal = ref(false)
const showTaskDetailModal = ref(false)
const showTaskHistoryModal = ref(false)
const isEditing = ref(false)
const parentTaskId = ref(null)
const searchQuery = ref('')
const activeFilter = ref('all')

const filterOptions = [
  { value: 'all', label: '全部' },
  { value: 'enabled', label: '已启用' },
  { value: 'disabled', label: '已禁用' },
  { value: 'daily', label: '每日' },
  { value: 'interval', label: '间隔' },
]

const stats = computed(() => {
  const total = tasks.value.length
  const enabled = tasks.value.filter(t => t.config?.enabled).length
  const avgSuccess = total > 0
    ? Math.round(tasks.value.reduce((sum, t) => sum + (t.execution?.success_rate || 0), 0) / total)
    : 0
  const totalRuns = tasks.value.reduce((sum, t) => sum + (t.execution?.total_runs || 0), 0)

  return [
    {
      label: '任务总数',
      value: total,
      icon: '<path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />',
      iconBg: 'bg-accent/15',
      iconColor: 'text-accent',
      bgClass: 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700',
      dotClass: 'bg-accent',
      valueColor: 'text-gray-900 dark:text-white',
    },
    {
      label: '运行中',
      value: enabled,
      icon: '<path stroke-linecap="round" stroke-linejoin="round" d="M5.636 5.636a9 9 0 1012.728 0M12 3v9" />',
      iconBg: 'bg-green-500/15',
      iconColor: 'text-green-600 dark:text-green-400',
      bgClass: 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700',
      dotClass: 'bg-green-500',
      valueColor: 'text-gray-900 dark:text-white',
      trend: `${enabled > 0 ? Math.round(enabled / total * 100) : 0}%`,
      trendClass: 'bg-green-500/10 text-green-600 dark:text-green-400',
    },
    {
      label: '平均成功率',
      value: avgSuccess + '%',
      icon: '<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />',
      iconBg: 'bg-blue-500/15',
      iconColor: 'text-blue-600 dark:text-blue-400',
      bgClass: 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700',
      dotClass: 'bg-blue-500',
      valueColor: avgSuccess >= 90 ? 'text-green-600 dark:text-green-400' : (avgSuccess >= 60 ? 'text-amber-600 dark:text-amber-400' : 'text-red-600 dark:text-red-400'),
    },
    {
      label: '累计执行',
      value: totalRuns,
      icon: '<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />',
      iconBg: 'bg-amber-500/15',
      iconColor: 'text-amber-600 dark:text-amber-400',
      bgClass: 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700',
      dotClass: 'bg-amber-500',
      valueColor: 'text-gray-900 dark:text-white',
    },
  ]
})

const filteredTasks = computed(() => {
  let result = tasks.value

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim()
    result = result.filter(t =>
      (t.config?.name || '').toLowerCase().includes(q) ||
      (t.config?.endpoint || '').toLowerCase().includes(q) ||
      (t.task_id || '').toLowerCase().includes(q)
    )
  }

  if (activeFilter.value === 'enabled') {
    result = result.filter(t => t.config?.enabled)
  } else if (activeFilter.value === 'disabled') {
    result = result.filter(t => !t.config?.enabled)
  } else if (activeFilter.value !== 'all') {
    result = result.filter(t => t.config?.schedule_type === activeFilter.value)
  }

  return result
})

const getScheduleTypeLabel = (type) => {
  const map = { daily: '每日执行', chain: '链式任务', once: '一次性', interval: '间隔执行' }
  return map[type] || type || '-'
}

const getScheduleTypeColor = (type) => {
  const map = {
    daily: 'bg-blue-500',
    chain: 'bg-purple-500',
    once: 'bg-green-500',
    interval: 'bg-amber-500',
  }
  return map[type] || 'bg-gray-500'
}

const getScheduleTypeBg = (type) => {
  const map = {
    daily: 'bg-blue-500/15',
    chain: 'bg-purple-500/15',
    once: 'bg-green-500/15',
    interval: 'bg-amber-500/15',
  }
  return map[type] || 'bg-gray-500/15'
}

const getScheduleTypeText = (type) => {
  const map = {
    daily: 'text-blue-600 dark:text-blue-400',
    chain: 'text-purple-600 dark:text-purple-400',
    once: 'text-green-600 dark:text-green-400',
    interval: 'text-amber-600 dark:text-amber-400',
  }
  return map[type] || 'text-gray-600 dark:text-gray-400'
}

const getScheduleTypeBadge = (type) => {
  const map = {
    daily: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
    chain: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
    once: 'bg-green-500/10 text-green-600 dark:text-green-400',
    interval: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
  }
  return map[type] || 'bg-gray-500/10 text-gray-600 dark:text-gray-400'
}

const getScheduleTimeDisplay = (task) => {
  if (task.config?.schedule_type === 'chain') return '依赖主任务'
  if (task.config?.schedule_type === 'interval') {
    const unit = { minutes: '分钟', hours: '小时', days: '天', months: '月', years: '年' }
    return `每 ${task.config?.interval_value || '-'} ${unit[task.config?.interval_unit] || ''}`
  }
  return task.config?.schedule_time || '-'
}

const fetchTasks = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const response = await getAllSchedulerTasks({ include_subtasks: true, detail_level: 'full' })
    if (response.data?.status === 'success') {
      tasks.value = (response.data.tasks || []).map(t => ({ ...t, isExpanded: false }))
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取任务列表失败' })
  } finally {
    loading.value = false
  }
}

const executeTask = async (taskId) => {
  try {
    const response = await executeSchedulerTask(taskId, { wait_for_completion: false })
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: '任务执行已启动' })
      fetchTasks()
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '执行失败: ' + (error.response?.data?.message || error.message) })
  }
}

const toggleTaskEnabled = async (taskId, enabled) => {
  try {
    const response = await setTaskEnabled(taskId, enabled)
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: enabled ? '已启用' : '已禁用' })
      fetchTasks()
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '操作失败' })
  }
}

const openTaskDetailModal = async (taskId) => {
  try {
    const response = await getSchedulerTaskDetail(taskId)
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      currentTask.value = response.data.tasks[0]
      showTaskDetailModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取详情失败' })
  }
}

const fetchTaskHistory = async (taskId) => {
  try {
    const response = await getTaskHistory({ task_id: taskId, include_subtasks: true, page: 1, page_size: 20 })
    if (response.data?.status === 'success') {
      showTaskHistoryModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取历史失败' })
  }
}

const openEditTaskModal = async (taskId) => {
  try {
    isEditing.value = true
    parentTaskId.value = null
    const response = await getSchedulerTaskDetail(taskId)
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      currentTask.value = response.data.tasks[0]
      showTaskFormModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取任务详情失败' })
  }
}

const openCreateSubTaskModal = async (taskId) => {
  try {
    isEditing.value = false
    const response = await getSchedulerTaskDetail(taskId)
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      const mainTask = response.data.tasks[0]
      parentTaskId.value = taskId
      if (mainTask.sub_tasks?.length > 0) {
        const lastSub = mainTask.sub_tasks[mainTask.sub_tasks.length - 1]
        currentTask.value = { ...mainTask, depends_on: { task_id: lastSub.task_id, name: lastSub.config?.name || lastSub.task_id } }
      } else {
        currentTask.value = mainTask
      }
      showTaskFormModal.value = true
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '获取主任务详情失败' })
  }
}

const openCreateTaskModal = () => {
  isEditing.value = false
  parentTaskId.value = null
  currentTask.value = null
  showTaskFormModal.value = true
}

const confirmDeleteTask = (taskId, parentTaskId = null) => {
  showDialog({
    title: '确认删除',
    message: parentTaskId ? '确定删除此子任务吗？' : '确定删除此任务及其所有子任务吗？',
    showCancelButton: true,
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    confirmButtonColor: 'var(--accent)',
  }).then(() => deleteTask(taskId, parentTaskId))
}

const deleteTask = async (taskId, parentTaskId = null) => {
  try {
    const response = parentTaskId ? await deleteSubTask(parentTaskId, taskId) : await deleteSchedulerTask(taskId)
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: '删除成功' })
      showTaskDetailModal.value = false
      fetchTasks()
    }
  } catch (error) {
    showNotify({ type: 'danger', message: '删除失败' })
  }
}

const { data: initialData } = await useAsyncData('scheduler-initial', async () => {
  try {
    const response = await getAllSchedulerTasks({ include_subtasks: true, detail_level: 'full' })
    if (response.data?.status === 'success') {
      return { tasks: (response.data.tasks || []).map(t => ({ ...t, isExpanded: false })) }
    }
  } catch {}
  return { tasks: [] }
})

if (initialData.value?.tasks) tasks.value = initialData.value.tasks

onMounted(() => { if (tasks.value.length === 0) fetchTasks() })
</script>

<style scoped>
.task-list-enter-active,
.task-list-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}
.task-list-enter-from {
  opacity: 0;
  transform: translateY(20px);
}
.task-list-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
.task-list-move {
  transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.subtask-enter-active,
.subtask-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}
.subtask-enter-from,
.subtask-leave-to {
  opacity: 0;
  max-height: 0;
}
.subtask-enter-to,
.subtask-leave-from {
  max-height: 1000px;
}
</style>
