<template>
  <div class="min-h-screen">
    <!-- 背景装饰 -->
    <div class="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
      <div class="absolute -top-40 -right-40 h-96 w-96 rounded-full bg-accent/10 blur-3xl" />
      <div class="absolute top-1/3 -left-40 h-96 w-96 rounded-full bg-accent/5 blur-3xl" />
    </div>

    <div class="mx-auto max-w-7xl px-4 py-6 md:px-6 md:py-8">
      <!-- 页面标题区 -->
      <div class="mb-6 flex flex-col gap-4 md:mb-8 md:flex-row md:items-center md:justify-between">
        <div class="flex items-center gap-4">
          <div class="relative">
            <div
              class="flex h-12 w-12 items-center justify-center rounded-2xl bg-accent shadow-lg shadow-accent/25 md:h-14 md:w-14"
            >
              <svg
                class="h-6 w-6 text-white md:h-7 md:w-7"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.8"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <span
              class="absolute -top-1 -right-1 h-4 w-4 animate-pulse-soft rounded-full border-2 border-white bg-green-500 dark:border-gray-900"
            />
          </div>
          <div>
            <h1 class="text-xl font-bold tracking-tight text-gray-900 md:text-2xl dark:text-white">
              计划任务
            </h1>
            <p class="mt-0.5 text-xs text-gray-500 md:text-sm dark:text-gray-400">
              自动化调度，让数据同步有条不紊
            </p>
          </div>
        </div>
        <button
          class="group relative inline-flex items-center justify-center gap-2 overflow-hidden rounded-xl px-4 py-2.5 text-sm font-medium text-white transition-all duration-300 md:px-5 md:py-3"
          :class="'bg-accent hover:-translate-y-0.5 hover:shadow-lg hover:shadow-accent/30 active:translate-y-0'"
          @click="openCreateTaskModal"
        >
          <span
            class="absolute inset-0 translate-y-full bg-white/20 transition-transform duration-300 group-hover:translate-y-0"
          />
          <svg
            class="relative h-4 w-4 md:h-5 md:w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          <span class="relative">新建任务</span>
        </button>
      </div>

      <!-- 统计卡片 -->
      <div
        v-if="tasks.length > 0"
        class="mb-6 grid grid-cols-2 gap-3 md:mb-8 md:grid-cols-4 md:gap-4"
      >
        <div
          v-for="stat in stats"
          :key="stat.label"
          class="group relative overflow-hidden rounded-2xl p-4 transition-all duration-300 hover:-translate-y-1 md:p-5"
          :class="stat.bgClass"
        >
          <div
            class="absolute -top-4 -right-4 h-24 w-24 rounded-full opacity-20"
            :class="stat.dotClass"
          />
          <div class="relative">
            <div class="mb-2 flex items-center justify-between">
              <div
                class="flex h-9 w-9 items-center justify-center rounded-xl md:h-10 md:w-10"
                :class="stat.iconBg"
              >
                <!-- eslint-disable-next-line vue/no-v-html -- stat.icon 是本文件内的静态 SVG 字符串，无外部输入 -->
                <svg class="h-4 w-4 md:h-5 md:w-5" :class="stat.iconColor" v-html="stat.icon" />
              </div>
              <span
                v-if="stat.trend"
                class="rounded-full px-2 py-0.5 text-[0.625rem] font-medium md:text-xs"
                :class="stat.trendClass"
                >{{ stat.trend }}</span
              >
            </div>
            <div class="text-2xl font-bold md:text-3xl" :class="stat.valueColor">
              {{ stat.value }}
            </div>
            <div class="mt-0.5 text-[0.6875rem] text-gray-500 md:text-xs dark:text-gray-400">
              {{ stat.label }}
            </div>
          </div>
        </div>
      </div>

      <!-- 搜索和筛选栏 -->
      <div v-if="tasks.length > 0" class="mb-5 flex flex-col gap-3 sm:flex-row md:mb-6">
        <div class="relative flex-1">
          <svg
            class="absolute top-1/2 left-3.5 h-4 w-4 -translate-y-1/2 text-gray-400 md:h-5 md:w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"
            />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索任务名称、端点..."
            class="w-full rounded-xl border border-gray-200 bg-white py-2.5 pr-4 pl-10 text-sm text-gray-900 placeholder-gray-400 transition-all focus:border-accent focus:ring-2 focus:ring-accent/30 focus:outline-none md:py-3 md:pl-12 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
          />
        </div>
        <div class="flex gap-2 overflow-x-auto pb-1 sm:pb-0">
          <button
            v-for="f in filterOptions"
            :key="f.value"
            class="flex-shrink-0 rounded-xl px-3 py-2 text-xs font-medium transition-all duration-200 md:px-4 md:py-2.5 md:text-sm"
            :class="
              activeFilter === f.value
                ? 'bg-accent text-white shadow-md shadow-accent/25'
                : 'border border-gray-200 bg-white text-gray-600 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'
            "
            @click="activeFilter = f.value"
          >
            {{ f.label }}
          </button>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="flex items-center justify-center py-16 md:py-24">
        <div class="flex flex-col items-center gap-4">
          <div class="relative">
            <div class="h-12 w-12 rounded-full border-2 border-accent/20 md:h-16 md:w-16" />
            <div
              class="absolute inset-0 h-12 w-12 animate-spin rounded-full border-2 border-accent border-t-transparent md:h-16 md:w-16"
            />
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400">加载任务中...</p>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else-if="filteredTasks.length === 0 && tasks.length === 0" class="glass-card">
        <div class="relative py-16 text-center md:py-20">
          <div class="relative mx-auto mb-6 h-24 w-24 md:h-32 md:w-32">
            <div class="absolute inset-0 animate-pulse-soft rounded-3xl bg-accent/15" />
            <div
              class="absolute inset-2 flex items-center justify-center rounded-3xl bg-white dark:bg-gray-800"
            >
              <svg
                class="h-12 w-12 text-accent/60 md:h-16 md:w-16"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
          </div>
          <h3 class="mb-2 text-lg font-semibold text-gray-900 md:text-xl dark:text-white">
            {{ searchQuery ? '没有找到匹配的任务' : '开启自动化之旅' }}
          </h3>
          <p class="mx-auto mb-6 max-w-sm text-sm text-gray-500 dark:text-gray-400">
            {{
              searchQuery
                ? '尝试更换搜索关键词或清除筛选条件'
                : '创建您的第一个计划任务，让数据同步自动完成'
            }}
          </p>
          <button
            v-if="!searchQuery"
            class="inline-flex items-center gap-2 rounded-xl bg-accent px-5 py-2.5 text-sm font-medium text-white shadow-lg shadow-accent/25 transition-colors hover:bg-accent/90"
            @click="openCreateTaskModal"
          >
            <svg
              class="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            创建第一个任务
          </button>
        </div>
      </div>

      <!-- 任务列表 -->
      <div v-else class="space-y-3 md:space-y-4">
        <TransitionGroup name="task-list">
          <div
            v-for="task in filteredTasks"
            :key="task.task_id"
            class="group glass-card relative overflow-hidden transition-all duration-300 hover:-translate-y-0.5 hover:shadow-xl"
          >
            <!-- 左侧装饰条 -->
            <div
              class="absolute top-0 bottom-0 left-0 w-1 transition-all duration-300"
              :class="getScheduleTypeColor(task.config?.schedule_type)"
            />

            <div class="p-4 md:p-5">
              <!-- 顶部：任务名称 + 状态 -->
              <div class="mb-4 flex items-start justify-between gap-3">
                <div class="flex min-w-0 flex-1 items-start gap-3">
                  <div
                    class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl transition-transform duration-300 group-hover:scale-105 md:h-11 md:w-11"
                    :class="getScheduleTypeBg(task.config?.schedule_type)"
                  >
                    <svg
                      class="h-5 w-5 md:h-[22px] md:w-[22px]"
                      :class="getScheduleTypeText(task.config?.schedule_type)"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        v-if="task.config?.schedule_type === 'daily'"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z"
                      />
                      <path
                        v-else-if="task.config?.schedule_type === 'chain'"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                      />
                      <path
                        v-else-if="task.config?.schedule_type === 'once'"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                      <path
                        v-else
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <h3
                        class="truncate text-sm font-semibold text-gray-900 md:text-base dark:text-white"
                      >
                        {{ task.config?.name || task.task_id }}
                      </h3>
                      <span
                        v-if="task.sub_tasks && task.sub_tasks.length > 0"
                        class="inline-flex items-center gap-1 rounded-full bg-accent/10 px-2 py-0.5 text-[0.625rem] font-medium text-accent md:text-xs"
                      >
                        <svg
                          class="h-3 w-3"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                          />
                        </svg>
                        {{ task.sub_tasks.length }} 子任务
                      </span>
                    </div>
                    <p
                      class="mt-0.5 truncate font-mono text-[0.6875rem] text-gray-400 md:text-xs dark:text-gray-500"
                    >
                      {{ task.config?.endpoint || '-' }}
                    </p>
                  </div>
                </div>

                <!-- 开关 -->
                <button
                  class="relative h-6 w-11 flex-shrink-0 rounded-full transition-all duration-300 focus:outline-none md:h-7 md:w-12"
                  :class="task.config?.enabled ? 'bg-accent' : 'bg-gray-300 dark:bg-gray-600'"
                  @click="toggleTaskEnabled(task.task_id, !task.config?.enabled)"
                >
                  <span
                    class="absolute top-0.5 left-0.5 flex h-5 w-5 items-center justify-center rounded-full bg-white shadow-md transition-all duration-300 md:h-6 md:w-6"
                    :class="{ 'translate-x-5 md:translate-x-6': task.config?.enabled }"
                  >
                    <svg
                      v-if="task.config?.enabled"
                      class="h-3 w-3 text-accent"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="3"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  </span>
                </button>
              </div>

              <!-- 信息行 -->
              <div class="mb-4 flex flex-wrap items-center gap-2 md:gap-3">
                <span
                  class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-[0.6875rem] font-medium md:text-xs"
                  :class="getScheduleTypeBadge(task.config?.schedule_type)"
                >
                  {{ getScheduleTypeLabel(task.config?.schedule_type) }}
                </span>
                <span
                  class="inline-flex items-center gap-1 text-[0.6875rem] text-gray-500 md:text-xs dark:text-gray-400"
                >
                  <svg
                    class="h-3.5 w-3.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  {{ getScheduleTimeDisplay(task) }}
                </span>
                <span
                  v-if="task.execution?.next_run && task.task_type === 'main'"
                  class="inline-flex items-center gap-1 text-[0.6875rem] text-gray-500 md:text-xs dark:text-gray-400"
                >
                  <svg
                    class="h-3.5 w-3.5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5"
                    />
                  </svg>
                  下次: {{ task.execution.next_run }}
                </span>
              </div>

              <!-- 统计条 + 操作 -->
              <div class="flex items-center justify-between gap-3">
                <!-- 执行统计 -->
                <div class="flex min-w-0 flex-1 items-center gap-3">
                  <template v-if="task.execution?.total_runs > 0">
                    <div class="flex items-center gap-1.5">
                      <div
                        class="h-2 w-2 rounded-full"
                        :class="{
                          'bg-green-500': task.execution.success_rate >= 90,
                          'bg-amber-500':
                            task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'bg-red-500': task.execution.success_rate < 60
                        }"
                      />
                      <span
                        class="text-xs font-semibold"
                        :class="{
                          'text-green-600 dark:text-green-400': task.execution.success_rate >= 90,
                          'text-amber-600 dark:text-amber-400':
                            task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'text-red-600 dark:text-red-400': task.execution.success_rate < 60
                        }"
                      >
                        {{ Math.round(task.execution.success_rate) }}%
                      </span>
                    </div>
                    <div
                      class="h-1.5 max-w-[120px] flex-1 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700"
                    >
                      <div
                        class="h-full rounded-full transition-all duration-500"
                        :class="{
                          'bg-green-500': task.execution.success_rate >= 90,
                          'bg-amber-500':
                            task.execution.success_rate >= 60 && task.execution.success_rate < 90,
                          'bg-red-500': task.execution.success_rate < 60
                        }"
                        :style="{ width: `${Math.round(task.execution.success_rate)}%` }"
                      />
                    </div>
                    <span class="text-[0.625rem] whitespace-nowrap text-gray-400 md:text-xs"
                      >{{ task.execution.total_runs }}次</span
                    >
                  </template>
                  <span v-else class="text-[0.625rem] text-gray-400 md:text-xs">尚未执行</span>
                </div>

                <!-- 操作按钮 -->
                <div class="flex items-center gap-0.5 md:gap-1">
                  <button
                    class="rounded-lg p-2 text-green-600 transition-all duration-200 hover:scale-105 hover:bg-green-50 active:scale-95 md:p-2.5 dark:text-green-400 dark:hover:bg-green-900/20"
                    title="立即执行"
                    @click="executeTask(task.task_id)"
                  >
                    <svg
                      class="h-4 w-4 md:h-[18px] md:w-[18px]"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path d="M8 5v14l11-7z" />
                    </svg>
                  </button>
                  <button
                    class="rounded-lg p-2 text-gray-500 transition-all duration-200 hover:scale-105 hover:bg-gray-100 hover:text-gray-700 active:scale-95 md:p-2.5 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200"
                    title="查看详情"
                    @click="openTaskDetailModal(task.task_id)"
                  >
                    <svg
                      class="h-4 w-4 md:h-[18px] md:w-[18px]"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z"
                      />
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                    </svg>
                  </button>
                  <button
                    class="rounded-lg p-2 text-gray-500 transition-all duration-200 hover:scale-105 hover:bg-gray-100 hover:text-gray-700 active:scale-95 md:p-2.5 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200"
                    title="编辑"
                    @click="openEditTaskModal(task.task_id)"
                  >
                    <svg
                      class="h-4 w-4 md:h-[18px] md:w-[18px]"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10"
                      />
                    </svg>
                  </button>
                  <button
                    class="rounded-lg p-2 text-gray-500 transition-all duration-200 hover:scale-105 hover:bg-gray-100 hover:text-gray-700 active:scale-95 md:p-2.5 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200"
                    title="添加子任务"
                    @click="openCreateSubTaskModal(task.task_id)"
                  >
                    <svg
                      class="h-4 w-4 md:h-[18px] md:w-[18px]"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M12 4.5v15m7.5-7.5h-15"
                      />
                    </svg>
                  </button>
                  <button
                    class="rounded-lg p-2 text-gray-500 transition-all duration-200 hover:scale-105 hover:bg-red-50 hover:text-red-500 active:scale-95 md:p-2.5 dark:text-gray-400 dark:hover:bg-red-900/20"
                    title="删除"
                    @click="confirmDeleteTask(task.task_id)"
                  >
                    <svg
                      class="h-4 w-4 md:h-[18px] md:w-[18px]"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.034-2.09 1.02-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
                      />
                    </svg>
                  </button>
                </div>
              </div>

              <!-- 子任务展开 -->
              <div v-if="task.sub_tasks && task.sub_tasks.length > 0" class="mt-4">
                <button
                  class="flex items-center gap-2 text-[0.6875rem] text-gray-500 transition-colors hover:text-accent md:text-xs dark:text-gray-400"
                  @click="task.isExpanded = !task.isExpanded"
                >
                  <svg
                    class="h-3.5 w-3.5 transition-transform duration-200"
                    :class="{ 'rotate-90': task.isExpanded }"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                  {{ task.isExpanded ? '收起' : '展开' }} {{ task.sub_tasks.length }} 个子任务
                </button>
              </div>
            </div>

            <!-- 子任务列表 -->
            <Transition name="subtask">
              <div
                v-if="task.sub_tasks && task.sub_tasks.length > 0 && task.isExpanded"
                class="border-t border-gray-100 bg-gray-50/50 dark:border-gray-700/50 dark:bg-gray-800/30"
              >
                <div
                  v-for="(sub, idx) in task.sub_tasks"
                  :key="sub.task_id"
                  class="flex items-center justify-between gap-3 px-4 py-3 transition-colors md:px-5"
                  :class="{
                    'border-b border-gray-100 dark:border-gray-700/50':
                      idx < task.sub_tasks.length - 1
                  }"
                >
                  <div class="flex min-w-0 flex-1 items-center gap-2.5">
                    <svg
                      class="h-4 w-4 flex-shrink-0 text-accent/50"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                      />
                    </svg>
                    <div class="min-w-0 flex-1">
                      <span
                        class="block truncate text-xs font-medium text-gray-700 dark:text-gray-300"
                      >
                        {{ sub.config?.name || sub.task_id }}
                      </span>
                      <span class="block truncate text-[0.625rem] text-gray-400"
                        >链式 · 依赖上一任务</span
                      >
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <button
                      class="relative h-5 w-9 rounded-full transition-colors duration-200"
                      :class="sub.config?.enabled ? 'bg-accent' : 'bg-gray-300 dark:bg-gray-600'"
                      @click="toggleTaskEnabled(sub.task_id, !sub.config?.enabled)"
                    >
                      <span
                        class="absolute top-0.5 left-0.5 h-4 w-4 rounded-full bg-white shadow transition-transform duration-200"
                        :class="{ 'translate-x-4': sub.config?.enabled }"
                      />
                    </button>
                    <button
                      class="rounded p-1.5 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-gray-300"
                      @click="openTaskDetailModal(sub.task_id)"
                    >
                      <svg
                        class="h-3.5 w-3.5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z"
                        />
                      </svg>
                    </button>
                    <button
                      class="rounded p-1.5 text-gray-400 transition-colors hover:text-red-500"
                      @click="confirmDeleteTask(sub.task_id, task.task_id)"
                    >
                      <svg
                        class="h-3.5 w-3.5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.034-2.09 1.02-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
                        />
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
    <TaskDetail
      v-model:show="showTaskDetailModal"
      :task="currentTask"
      @view-history="fetchTaskHistory"
      @edit-task="openEditTaskModal"
      @execute-task="executeTask"
      @toggle-enabled="toggleTaskEnabled"
      @delete-task="confirmDeleteTask"
      @refresh="fetchTasks"
    />
    <TaskHistory
      v-model:show="showTaskHistoryModal"
      :task-id="currentTask?.task_id"
      :task-name="currentTask?.config?.name || currentTask?.task_id"
    />
    <TaskForm
      v-model:show="showTaskFormModal"
      :is-editing="isEditing"
      :task-id="currentTask?.task_id"
      :parent-task-id="parentTaskId"
      :tasks="tasks"
      @task-saved="fetchTasks"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useAsyncData } from '#imports';
import { showNotify, showDialog } from 'vant';
import 'vant/es/dialog/style';
import 'vant/es/notify/style';
import {
  getAllSchedulerTasks,
  getSchedulerTaskDetail,
  executeSchedulerTask,
  getTaskHistory,
  setTaskEnabled,
  deleteSchedulerTask,
  deleteSubTask
} from '~/utils/api';
import TaskForm from '../scheduler/TaskForm.vue';
import TaskDetail from '../scheduler/TaskDetail.vue';
import TaskHistory from '../scheduler/TaskHistory.vue';

const loading = ref(false);
const tasks = ref([]);
const currentTask = ref(null);
const showTaskFormModal = ref(false);
const showTaskDetailModal = ref(false);
const showTaskHistoryModal = ref(false);
const isEditing = ref(false);
const parentTaskId = ref(null);
const searchQuery = ref('');
const activeFilter = ref('all');

const filterOptions = [
  { value: 'all', label: '全部' },
  { value: 'enabled', label: '已启用' },
  { value: 'disabled', label: '已禁用' },
  { value: 'daily', label: '每日' },
  { value: 'interval', label: '间隔' }
];

const stats = computed(() => {
  const total = tasks.value.length;
  const enabled = tasks.value.filter((t) => t.config?.enabled).length;
  const avgSuccess =
    total > 0
      ? Math.round(
          tasks.value.reduce((sum, t) => sum + (t.execution?.success_rate || 0), 0) / total
        )
      : 0;
  const totalRuns = tasks.value.reduce((sum, t) => sum + (t.execution?.total_runs || 0), 0);

  return [
    {
      label: '任务总数',
      value: total,
      icon: '<path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />',
      iconBg: 'bg-accent/15',
      iconColor: 'text-accent',
      bgClass: 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700',
      dotClass: 'bg-accent',
      valueColor: 'text-gray-900 dark:text-white'
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
      trend: `${enabled > 0 ? Math.round((enabled / total) * 100) : 0}%`,
      trendClass: 'bg-green-500/10 text-green-600 dark:text-green-400'
    },
    {
      label: '平均成功率',
      value: avgSuccess + '%',
      icon: '<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />',
      iconBg: 'bg-blue-500/15',
      iconColor: 'text-blue-600 dark:text-blue-400',
      bgClass: 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700',
      dotClass: 'bg-blue-500',
      valueColor:
        avgSuccess >= 90
          ? 'text-green-600 dark:text-green-400'
          : avgSuccess >= 60
            ? 'text-amber-600 dark:text-amber-400'
            : 'text-red-600 dark:text-red-400'
    },
    {
      label: '累计执行',
      value: totalRuns,
      icon: '<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />',
      iconBg: 'bg-amber-500/15',
      iconColor: 'text-amber-600 dark:text-amber-400',
      bgClass: 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700',
      dotClass: 'bg-amber-500',
      valueColor: 'text-gray-900 dark:text-white'
    }
  ];
});

const filteredTasks = computed(() => {
  let result = tasks.value;

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    result = result.filter(
      (t) =>
        (t.config?.name || '').toLowerCase().includes(q) ||
        (t.config?.endpoint || '').toLowerCase().includes(q) ||
        (t.task_id || '').toLowerCase().includes(q)
    );
  }

  if (activeFilter.value === 'enabled') {
    result = result.filter((t) => t.config?.enabled);
  } else if (activeFilter.value === 'disabled') {
    result = result.filter((t) => !t.config?.enabled);
  } else if (activeFilter.value !== 'all') {
    result = result.filter((t) => t.config?.schedule_type === activeFilter.value);
  }

  return result;
});

const getScheduleTypeLabel = (type) => {
  const map = { daily: '每日执行', chain: '链式任务', once: '一次性', interval: '间隔执行' };
  return map[type] || type || '-';
};

const getScheduleTypeColor = (type) => {
  const map = {
    daily: 'bg-blue-500',
    chain: 'bg-purple-500',
    once: 'bg-green-500',
    interval: 'bg-amber-500'
  };
  return map[type] || 'bg-gray-500';
};

const getScheduleTypeBg = (type) => {
  const map = {
    daily: 'bg-blue-500/15',
    chain: 'bg-purple-500/15',
    once: 'bg-green-500/15',
    interval: 'bg-amber-500/15'
  };
  return map[type] || 'bg-gray-500/15';
};

const getScheduleTypeText = (type) => {
  const map = {
    daily: 'text-blue-600 dark:text-blue-400',
    chain: 'text-purple-600 dark:text-purple-400',
    once: 'text-green-600 dark:text-green-400',
    interval: 'text-amber-600 dark:text-amber-400'
  };
  return map[type] || 'text-gray-600 dark:text-gray-400';
};

const getScheduleTypeBadge = (type) => {
  const map = {
    daily: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
    chain: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
    once: 'bg-green-500/10 text-green-600 dark:text-green-400',
    interval: 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
  };
  return map[type] || 'bg-gray-500/10 text-gray-600 dark:text-gray-400';
};

const getScheduleTimeDisplay = (task) => {
  if (task.config?.schedule_type === 'chain') return '依赖主任务';
  if (task.config?.schedule_type === 'interval') {
    const unit = { minutes: '分钟', hours: '小时', days: '天', months: '月', years: '年' };
    return `每 ${task.config?.interval_value || '-'} ${unit[task.config?.interval_unit] || ''}`;
  }
  return task.config?.schedule_time || '-';
};

const fetchTasks = async () => {
  if (loading.value) return;
  loading.value = true;
  try {
    const response = await getAllSchedulerTasks({ include_subtasks: true, detail_level: 'full' });
    if (response.data?.status === 'success') {
      tasks.value = (response.data.tasks || []).map((t) => ({ ...t, isExpanded: false }));
    }
  } catch {
    showNotify({ type: 'danger', message: '获取任务列表失败' });
  } finally {
    loading.value = false;
  }
};

const executeTask = async (taskId) => {
  try {
    const response = await executeSchedulerTask(taskId, { wait_for_completion: false });
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: '任务执行已启动' });
      fetchTasks();
    }
  } catch (error) {
    showNotify({
      type: 'danger',
      message: '执行失败: ' + (error.response?.data?.message || error.message)
    });
  }
};

const toggleTaskEnabled = async (taskId, enabled) => {
  try {
    const response = await setTaskEnabled(taskId, enabled);
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: enabled ? '已启用' : '已禁用' });
      fetchTasks();
    }
  } catch {
    showNotify({ type: 'danger', message: '操作失败' });
  }
};

const openTaskDetailModal = async (taskId) => {
  try {
    const response = await getSchedulerTaskDetail(taskId);
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      currentTask.value = response.data.tasks[0];
      showTaskDetailModal.value = true;
    }
  } catch {
    showNotify({ type: 'danger', message: '获取详情失败' });
  }
};

const fetchTaskHistory = async (taskId) => {
  try {
    const response = await getTaskHistory({
      task_id: taskId,
      include_subtasks: true,
      page: 1,
      page_size: 20
    });
    if (response.data?.status === 'success') {
      showTaskHistoryModal.value = true;
    }
  } catch {
    showNotify({ type: 'danger', message: '获取历史失败' });
  }
};

const openEditTaskModal = async (taskId) => {
  try {
    isEditing.value = true;
    parentTaskId.value = null;
    const response = await getSchedulerTaskDetail(taskId);
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      currentTask.value = response.data.tasks[0];
      showTaskFormModal.value = true;
    }
  } catch {
    showNotify({ type: 'danger', message: '获取任务详情失败' });
  }
};

const openCreateSubTaskModal = async (taskId) => {
  try {
    isEditing.value = false;
    const response = await getSchedulerTaskDetail(taskId);
    if (response.data?.status === 'success' && response.data.tasks?.length > 0) {
      const mainTask = response.data.tasks[0];
      parentTaskId.value = taskId;
      if (mainTask.sub_tasks?.length > 0) {
        const lastSub = mainTask.sub_tasks[mainTask.sub_tasks.length - 1];
        currentTask.value = {
          ...mainTask,
          depends_on: { task_id: lastSub.task_id, name: lastSub.config?.name || lastSub.task_id }
        };
      } else {
        currentTask.value = mainTask;
      }
      showTaskFormModal.value = true;
    }
  } catch {
    showNotify({ type: 'danger', message: '获取主任务详情失败' });
  }
};

const openCreateTaskModal = () => {
  isEditing.value = false;
  parentTaskId.value = null;
  currentTask.value = null;
  showTaskFormModal.value = true;
};

const confirmDeleteTask = (taskId, parentTaskId = null) => {
  showDialog({
    title: '确认删除',
    message: parentTaskId ? '确定删除此子任务吗？' : '确定删除此任务及其所有子任务吗？',
    showCancelButton: true,
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    confirmButtonColor: 'var(--accent)'
  }).then(() => deleteTask(taskId, parentTaskId));
};

const deleteTask = async (taskId, parentTaskId = null) => {
  try {
    const response = parentTaskId
      ? await deleteSubTask(parentTaskId, taskId)
      : await deleteSchedulerTask(taskId);
    if (response.data?.status === 'success') {
      showNotify({ type: 'success', message: '删除成功' });
      showTaskDetailModal.value = false;
      fetchTasks();
    }
  } catch {
    showNotify({ type: 'danger', message: '删除失败' });
  }
};

const { data: initialData } = await useAsyncData('scheduler-initial', async () => {
  try {
    const response = await getAllSchedulerTasks({ include_subtasks: true, detail_level: 'full' });
    if (response.data?.status === 'success') {
      return { tasks: (response.data.tasks || []).map((t) => ({ ...t, isExpanded: false })) };
    }
  } catch {
    // SSR 预取失败时回落空列表，挂载后由 fetchTasks 重试
  }
  return { tasks: [] };
});

if (initialData.value?.tasks) tasks.value = initialData.value.tasks;

onMounted(() => {
  if (tasks.value.length === 0) fetchTasks();
});
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
