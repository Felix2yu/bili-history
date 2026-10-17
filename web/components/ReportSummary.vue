<template>
  <div class="space-y-4">
    <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-rose-500">{{ summary.total_videos }}</div>
        <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">观看视频数</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-amber-500">
          {{ formatDurationShort(summary.total_duration) }}
        </div>
        <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">总观看时长</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-emerald-500">{{ summary.unique_days }}</div>
        <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">活跃天数</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-sky-500">{{ summary.unique_authors }}</div>
        <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">UP主数量</div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
      <div class="glass-card flex items-center justify-between p-4 text-sm">
        <div class="flex items-center gap-2">
          <svg
            class="h-4 w-4 text-rose-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
            />
          </svg>
          <span class="text-gray-600 dark:text-gray-400">日均</span>
        </div>
        <div class="flex gap-4">
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-rose-500">{{
              summary.avg_daily_videos?.toFixed(1)
            }}</span>
            个视频
          </span>
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-amber-500">{{
              formatDurationShort(summary.avg_daily_duration)
            }}</span>
          </span>
        </div>
      </div>

      <div
        v-if="summary.completion_stats"
        class="glass-card flex items-center justify-between p-4 text-sm"
      >
        <div class="flex items-center gap-2">
          <svg
            class="h-4 w-4 text-emerald-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <span class="text-gray-600 dark:text-gray-400">完播</span>
        </div>
        <div class="flex gap-4">
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-emerald-500">{{
              summary.completion_stats.finished
            }}</span>
            看完
          </span>
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-sky-500"
              >{{ (summary.completion_stats.avg_rate * 100).toFixed(0) }}%</span
            >
            平均完播率
          </span>
        </div>
      </div>

      <div
        v-if="summary.device_dist && Object.keys(summary.device_dist).length > 0"
        class="glass-card flex items-center justify-between p-4 text-sm"
      >
        <div class="flex items-center gap-2">
          <svg
            class="h-4 w-4 text-sky-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
            />
          </svg>
          <span class="text-gray-600 dark:text-gray-400">设备</span>
        </div>
        <div class="flex flex-wrap gap-3">
          <div
            v-for="(count, device) in summary.device_dist"
            :key="device"
            class="flex items-center gap-1"
          >
            <div
              class="h-2 w-2 rounded-full"
              :style="{ backgroundColor: getDeviceColor(device) }"
            />
            <span class="text-gray-600 dark:text-gray-400">{{ device }}</span>
            <span class="font-semibold" :style="{ color: getDeviceColor(device) }">{{
              count
            }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <div v-if="summary.daily_breakdown?.length" class="glass-card p-4">
        <h4
          class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
        >
          <svg
            class="h-4 w-4 text-sky-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            />
          </svg>
          每日观看
        </h4>
        <div style="height: 100px">
          <div class="flex h-full items-end gap-1">
            <div
              v-for="(day, idx) in summary.daily_breakdown"
              :key="day.date"
              class="group relative flex h-full flex-1 flex-col items-center"
            >
              <div
                class="absolute -top-8 left-1/2 z-10 -translate-x-1/2 rounded bg-gray-800 px-2 py-1 text-[0.625rem] whitespace-nowrap text-white opacity-0 transition-opacity group-hover:opacity-100"
              >
                {{ day.date.slice(5) }}: {{ day.count }}个视频
              </div>
              <div
                class="mt-auto w-full rounded-t transition-all duration-300 hover:opacity-80"
                :style="{
                  height: `${Math.max((day.count / maxDailyCount) * 100, 4)}%`,
                  backgroundColor: getChartColor(idx)
                }"
              />
            </div>
          </div>
        </div>
        <div class="mt-1 flex justify-between text-[0.5625rem] text-gray-400 dark:text-gray-500">
          <span
            v-for="day in summary.daily_breakdown.filter(
              (_, i) =>
                i % Math.max(Math.floor(summary.daily_breakdown.length / 5), 1) === 0 ||
                i === summary.daily_breakdown.length - 1
            )"
            :key="day.date"
            >{{ day.date.slice(8) }}</span
          >
        </div>
      </div>

      <div
        v-if="summary.hour_dist && Object.keys(summary.hour_dist).length > 0"
        class="glass-card p-4"
      >
        <h4
          class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
        >
          <svg
            class="h-4 w-4 text-purple-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          观看时段
        </h4>
        <div style="height: 100px">
          <div class="flex h-full items-end gap-px">
            <div
              v-for="hour in 24"
              :key="hour - 1"
              class="group relative flex h-full flex-1 flex-col items-center"
            >
              <div
                class="absolute -top-8 left-1/2 z-10 -translate-x-1/2 rounded bg-gray-800 px-2 py-1 text-[0.625rem] whitespace-nowrap text-white opacity-0 transition-opacity group-hover:opacity-100"
              >
                {{ hour - 1 }}时: {{ summary.hour_dist[hour - 1] || 0 }}次
              </div>
              <div
                class="mt-auto w-full rounded-t transition-all duration-300"
                :style="{
                  height: `${Math.max(((summary.hour_dist[hour - 1] || 0) / maxHourCount) * 100, 2)}%`,
                  backgroundColor: getTimeSlotColor(hour - 1)
                }"
              />
            </div>
          </div>
        </div>
        <div class="mt-1 flex justify-between text-[0.5625rem] text-gray-400 dark:text-gray-500">
          <span>0时</span><span>6时</span><span>12时</span><span>18时</span><span>23时</span>
        </div>
      </div>
    </div>

    <div v-if="summary.top_categories?.length" class="glass-card p-4">
      <h4
        class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
      >
        <svg
          class="h-4 w-4 text-rose-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z"
          />
        </svg>
        常看分区
      </h4>
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="(cat, index) in summary.top_categories.slice(0, 6)"
          :key="cat.name"
          class="flex items-center gap-2"
        >
          <span class="w-4 text-right text-xs text-gray-400">{{ index + 1 }}</span>
          <div class="min-w-0 flex-1">
            <div class="mb-0.5 flex items-center justify-between">
              <span class="truncate text-sm text-gray-700 dark:text-gray-300">{{ cat.name }}</span>
              <span class="ml-2 text-xs text-gray-500 dark:text-gray-400">{{ cat.count }}次</span>
            </div>
            <div class="h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
              <div
                class="h-full rounded-full"
                :style="{
                  width: `${(cat.count / summary.top_categories[0].count) * 100}%`,
                  backgroundColor: getChartColor(index)
                }"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <div v-if="summary.top_authors?.length" class="glass-card p-4">
        <h4
          class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
        >
          <svg
            class="h-4 w-4 text-sky-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
            />
          </svg>
          常看UP主
        </h4>
        <div class="space-y-2">
          <div
            v-for="(author, index) in summary.top_authors.slice(0, 5)"
            :key="author.mid"
            class="flex items-center gap-2"
          >
            <span class="w-4 text-right text-xs text-gray-400">{{ index + 1 }}</span>
            <div class="min-w-0 flex-1">
              <div class="mb-0.5 flex items-center justify-between">
                <span class="truncate text-sm text-gray-700 dark:text-gray-300">{{
                  author.name
                }}</span>
                <span class="ml-2 text-xs text-gray-500 dark:text-gray-400"
                  >{{ author.count }}次</span
                >
              </div>
              <div class="h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
                <div
                  class="h-full rounded-full"
                  :style="{
                    width: `${(author.count / summary.top_authors[0].count) * 100}%`,
                    backgroundColor: getChartColor(index + 2)
                  }"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="summary.title_keywords?.length" class="glass-card p-4">
        <h4
          class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
        >
          <svg
            class="h-4 w-4 text-amber-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"
            />
          </svg>
          标题热词
        </h4>
        <div class="flex flex-wrap gap-1.5">
          <span
            v-for="(kw, index) in summary.title_keywords.slice(0, 15)"
            :key="kw.word"
            class="inline-flex items-center rounded-full px-2 py-0.5 text-[0.6875rem] font-medium"
            :style="{ backgroundColor: getChartColor(index) + '25', color: getChartColor(index) }"
          >
            {{ kw.word }}
            <span class="ml-1 text-[0.5625rem] opacity-60">{{ kw.count }}</span>
          </span>
        </div>
      </div>

      <div v-if="summary.weekday_dist?.length" class="glass-card p-4">
        <h4
          class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
        >
          <svg
            class="h-4 w-4 text-purple-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
            />
          </svg>
          周内分布
        </h4>
        <div class="space-y-1.5">
          <div
            v-for="(day, idx) in summary.weekday_dist"
            :key="day.name"
            class="flex items-center gap-2"
          >
            <span
              class="w-6 flex-shrink-0 text-right text-[0.625rem] text-gray-500 dark:text-gray-400"
              >{{ day.name }}</span
            >
            <div class="h-4 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
              <div
                class="h-full rounded-full transition-all duration-300"
                :style="{
                  width: `${Math.max((day.count / maxWeekdayCount) * 100, 2)}%`,
                  backgroundColor: getChartColor(idx)
                }"
              />
            </div>
            <span
              class="w-5 flex-shrink-0 text-right text-[0.625rem] text-gray-500 dark:text-gray-400"
              >{{ day.count }}</span
            >
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
      <div v-if="summary.late_night_ratio !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-purple-500">
          {{ (summary.late_night_ratio * 100).toFixed(0) }}%
        </div>
        <div class="mt-0.5 text-[0.625rem] text-gray-500 dark:text-gray-400">深夜观看</div>
      </div>
      <div v-if="summary.favorite_rate !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-amber-500">
          {{ (summary.favorite_rate * 100).toFixed(0) }}%
        </div>
        <div class="mt-0.5 text-[0.625rem] text-gray-500 dark:text-gray-400">收藏率</div>
      </div>
      <div v-if="summary.abandon_rate !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-rose-500">
          {{ (summary.abandon_rate * 100).toFixed(0) }}%
        </div>
        <div class="mt-0.5 text-[0.625rem] text-gray-500 dark:text-gray-400">弃看率</div>
      </div>
      <div v-if="summary.golden_slot_ratio !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-orange-500">
          {{ (summary.golden_slot_ratio * 100).toFixed(0) }}%
        </div>
        <div class="mt-0.5 text-[0.625rem] text-gray-500 dark:text-gray-400">黄金3h集中度</div>
      </div>
      <div v-if="summary.up_diversity !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-cyan-500">
          {{ (summary.up_diversity * 100).toFixed(0) }}%
        </div>
        <div class="mt-0.5 text-[0.625rem] text-gray-500 dark:text-gray-400">UP主多样性</div>
      </div>
      <div v-if="summary.rewatch_stats?.total_rewatched" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-emerald-500">
          {{ summary.rewatch_stats.total_rewatched }}
        </div>
        <div class="mt-0.5 text-[0.625rem] text-gray-500 dark:text-gray-400">重刷次数</div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-if="
          summary.duration_pref?.short + summary.duration_pref?.mid + summary.duration_pref?.long >
          0
        "
        class="glass-card p-4"
      >
        <h4 class="mb-3 text-sm font-semibold text-gray-700 dark:text-gray-300">时长偏好</h4>
        <div class="flex h-4 overflow-hidden rounded-full">
          <div
            class="transition-all"
            :style="{
              width: `${summary.duration_pref.short_ratio * 100}%`,
              backgroundColor: DURATION_COLORS['短视频']
            }"
            title="短视频"
          />
          <div
            class="transition-all"
            :style="{
              width: `${summary.duration_pref.mid_ratio * 100}%`,
              backgroundColor: DURATION_COLORS['中等视频']
            }"
            title="中视频"
          />
          <div
            class="transition-all"
            :style="{
              width: `${summary.duration_pref.long_ratio * 100}%`,
              backgroundColor: DURATION_COLORS['长视频']
            }"
            title="长视频"
          />
        </div>
        <div class="mt-2 flex justify-between text-[0.625rem] text-gray-500 dark:text-gray-400">
          <span class="flex items-center gap-1"
            ><span
              class="inline-block h-2 w-2 rounded-full"
              :style="{ backgroundColor: DURATION_COLORS['短视频'] }"
            />短(&lt;5min) {{ summary.duration_pref.short }}</span
          >
          <span class="flex items-center gap-1"
            ><span
              class="inline-block h-2 w-2 rounded-full"
              :style="{ backgroundColor: DURATION_COLORS['中等视频'] }"
            />中(5-20min) {{ summary.duration_pref.mid }}</span
          >
          <span class="flex items-center gap-1"
            ><span
              class="inline-block h-2 w-2 rounded-full"
              :style="{ backgroundColor: DURATION_COLORS['长视频'] }"
            />长(&gt;20min) {{ summary.duration_pref.long }}</span
          >
        </div>
      </div>

      <div v-if="summary.completion_dist?.length" class="glass-card p-4">
        <h4 class="mb-3 text-sm font-semibold text-gray-700 dark:text-gray-300">完播率分布</h4>
        <div class="space-y-1.5">
          <div
            v-for="(item, idx) in summary.completion_dist"
            :key="item.range"
            class="flex items-center gap-2"
          >
            <span class="w-12 text-right text-[0.625rem] text-gray-500 dark:text-gray-400">{{
              item.range
            }}</span>
            <div class="h-3 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
              <div
                class="h-full rounded-full"
                :style="{
                  width: `${(item.count / maxCompCount) * 100}%`,
                  backgroundColor: getChartColor(idx)
                }"
              />
            </div>
            <span class="w-6 text-[0.625rem] text-gray-500 dark:text-gray-400">{{
              item.count
            }}</span>
          </div>
        </div>
      </div>

      <div v-if="summary.top_time_slots?.length" class="glass-card p-4">
        <h4 class="mb-3 text-sm font-semibold text-gray-700 dark:text-gray-300">活跃时段</h4>
        <div class="space-y-2">
          <div
            v-for="(slot, index) in summary.top_time_slots"
            :key="slot.name"
            class="flex items-center gap-2"
          >
            <span class="w-4 text-right text-xs text-gray-400">{{ index + 1 }}</span>
            <div class="min-w-0 flex-1">
              <div class="mb-0.5 flex items-center justify-between">
                <span class="text-sm text-gray-700 dark:text-gray-300">{{ slot.name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ slot.count }}次</span>
              </div>
              <div class="h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700">
                <div
                  class="h-full rounded-full"
                  :style="{
                    width: `${(slot.count / summary.top_time_slots[0].count) * 100}%`,
                    backgroundColor: getChartColor(index + 4)
                  }"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="summary.rewatch_stats?.rewatched_videos?.length" class="glass-card p-4">
      <h4
        class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
      >
        <svg
          class="h-4 w-4 text-emerald-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        重刷视频
      </h4>
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="(video, index) in summary.rewatch_stats.rewatched_videos"
          :key="video.bvid"
          class="flex items-center gap-3 rounded-lg p-2 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/50"
        >
          <span class="w-4 text-right font-mono text-xs text-gray-400">{{ index + 1 }}</span>
          <img
            v-if="video.cover"
            :src="normalizeImageUrl(video.cover)"
            class="h-10 w-16 flex-shrink-0 rounded object-cover"
            loading="lazy"
          />
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm text-gray-700 dark:text-gray-300">{{ video.title }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ video.author_name }}</div>
          </div>
          <div class="flex-shrink-0 text-right">
            <div class="text-sm font-semibold" :style="{ color: getChartColor(index) }">
              {{ video.count }}次
            </div>
            <div class="text-[0.625rem] text-gray-400">
              {{ formatDurationShort(video.total_duration) }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="summary.top_videos?.length" class="glass-card p-4">
      <h4
        class="mb-3 flex items-center gap-2 text-sm font-semibold text-gray-700 dark:text-gray-300"
      >
        <svg
          class="h-4 w-4 text-amber-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        最长观看
      </h4>
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="(video, index) in summary.top_videos"
          :key="video.bvid"
          class="flex items-center gap-3 rounded-lg p-2 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/50"
        >
          <span class="w-4 text-right font-mono text-xs text-gray-400">{{ index + 1 }}</span>
          <img
            v-if="video.cover"
            :src="normalizeImageUrl(video.cover)"
            class="h-10 w-16 flex-shrink-0 rounded object-cover"
            loading="lazy"
          />
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm text-gray-700 dark:text-gray-300">{{ video.title }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ video.author_name }}</div>
          </div>
          <div class="flex-shrink-0 text-right">
            <div class="text-sm font-semibold" :style="{ color: getChartColor(index) }">
              {{ formatDurationShort(video.duration) }}
            </div>
            <div class="text-[0.625rem] text-gray-400">{{ formatTimestamp(video.view_at) }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { formatDurationShort, formatTimestamp } from '~/utils/format';
import { normalizeImageUrl } from '~/utils/imageUrl';
import {
  getChartColor,
  getTimeSlotColor,
  DEVICE_COLORS,
  DURATION_COLORS
} from '~/utils/chartColors';

const props = defineProps({
  summary: {
    type: Object,
    required: true
  }
});

const getDeviceColor = (device) => {
  return DEVICE_COLORS[device] || '#9CA3AF';
};

const maxDailyCount = computed(() => {
  if (!props.summary.daily_breakdown?.length) return 1;
  return Math.max(...props.summary.daily_breakdown.map((d) => d.count), 1);
});

const maxHourCount = computed(() => {
  if (!props.summary.hour_dist) return 1;
  return Math.max(...Object.values(props.summary.hour_dist), 1);
});

const maxCompCount = computed(() => {
  if (!props.summary.completion_dist?.length) return 1;
  return Math.max(...props.summary.completion_dist.map((d) => d.count), 1);
});

const maxWeekdayCount = computed(() => {
  if (!props.summary.weekday_dist?.length) return 1;
  return Math.max(...props.summary.weekday_dist.map((d) => d.count), 1);
});
</script>

<style scoped>
@reference "../assets/css/main.css";
.glass-card {
  @apply rounded-xl border border-gray-200/50 bg-white/80 shadow-sm backdrop-blur-sm dark:border-gray-700/50 dark:bg-gray-800/80;
}
</style>
