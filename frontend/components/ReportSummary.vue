<template>
  <div class="space-y-4">
    <!-- 核心统计卡片 -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ summary.total_videos }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">观看视频数</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ formatDurationShort(summary.total_duration) }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">总观看时长</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ summary.unique_days }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">活跃天数</div>
      </div>
      <div class="glass-card p-4 text-center">
        <div class="text-2xl font-bold text-[#fb7299]">{{ summary.unique_authors }}</div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">UP主数量</div>
      </div>
    </div>

    <!-- 日均 + 完播 + 设备分布 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
      <div class="glass-card p-4 flex items-center justify-between text-sm">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
          <span class="text-gray-600 dark:text-gray-400">日均</span>
        </div>
        <div class="flex gap-4">
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-[#fb7299]">{{ summary.avg_daily_videos?.toFixed(1) }}</span> 个视频
          </span>
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-[#fb7299]">{{ formatDurationShort(summary.avg_daily_duration) }}</span>
          </span>
        </div>
      </div>

      <div v-if="summary.completion_stats" class="glass-card p-4 flex items-center justify-between text-sm">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-gray-600 dark:text-gray-400">完播</span>
        </div>
        <div class="flex gap-4">
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-green-500">{{ summary.completion_stats.finished }}</span> 看完
          </span>
          <span class="text-gray-700 dark:text-gray-300">
            <span class="font-semibold text-[#fb7299]">{{ (summary.completion_stats.avg_rate * 100).toFixed(0) }}%</span> 平均完播率
          </span>
        </div>
      </div>

      <!-- 设备分布（紧凑条形） -->
      <div v-if="summary.device_dist && Object.keys(summary.device_dist).length > 0" class="glass-card p-4 flex items-center justify-between text-sm">
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          <span class="text-gray-600 dark:text-gray-400">设备</span>
        </div>
        <div class="flex gap-3 flex-wrap">
          <div
            v-for="(count, device) in summary.device_dist"
            :key="device"
            class="flex items-center gap-1"
          >
            <div
              class="w-2 h-2 rounded-full"
              :class="device === '手机' ? 'bg-[#fb7299]' : device === '电脑' ? 'bg-blue-500' : device === '平板' ? 'bg-green-500' : 'bg-gray-400'"
            ></div>
            <span class="text-gray-600 dark:text-gray-400">{{ device }}</span>
            <span class="font-semibold text-[#fb7299]">{{ count }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 每日观看 + 时段分布（并排） -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- 每日观看分布 (柱状图) -->
      <div v-if="summary.daily_breakdown?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          每日观看
        </h4>
        <div class="relative" style="height: 100px;">
          <div class="absolute inset-0 flex items-end gap-1">
            <div
              v-for="day in summary.daily_breakdown"
              :key="day.date"
              class="flex-1 flex flex-col items-center group relative"
              style="height: 100%;"
            >
              <div class="absolute -top-8 left-1/2 -translate-x-1/2 bg-gray-800 text-white text-[10px] px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-10">
                {{ day.date.slice(5) }}: {{ day.count }}个视频
              </div>
              <div class="w-full mt-auto bg-gradient-to-t from-[#fb7299] to-[#fc9b7a] rounded-t transition-all duration-300 hover:opacity-80"
                   :style="{ height: `${Math.max((day.count / maxDailyCount) * 100, 4)}%` }"
              ></div>
              <span class="text-[9px] text-gray-400 dark:text-gray-500 mt-1 flex-shrink-0">{{ day.date.slice(8) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 时段分布 -->
      <div v-if="summary.hour_dist && Object.keys(summary.hour_dist).length > 0" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          观看时段
        </h4>
        <div class="relative" style="height: 100px;">
          <div class="absolute inset-0 flex items-end gap-px">
            <div
              v-for="hour in 24"
              :key="hour - 1"
              class="flex-1 flex flex-col items-center group relative"
              style="height: 100%;"
            >
              <div class="absolute -top-8 left-1/2 -translate-x-1/2 bg-gray-800 text-white text-[10px] px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-10">
                {{ hour - 1 }}时: {{ summary.hour_dist[hour - 1] || 0 }}次
              </div>
              <div
                class="w-full mt-auto rounded-t transition-all duration-300"
                :class="(hour - 1) >= 22 || (hour - 1) < 6 ? 'bg-purple-400' : 'bg-[#fb7299]'"
                :style="{ height: `${Math.max(((summary.hour_dist[hour - 1] || 0) / maxHourCount) * 100, 2)}%` }"
              ></div>
            </div>
          </div>
        </div>
        <div class="flex justify-between text-[9px] text-gray-400 dark:text-gray-500 mt-1">
          <span>0时</span>
          <span>6时</span>
          <span>12时</span>
          <span>18时</span>
          <span>23时</span>
        </div>
      </div>
    </div>

    <!-- Top 分区 -->
    <div v-if="summary.top_categories?.length" class="glass-card p-4">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z" />
        </svg>
        常看分区
      </h4>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
        <div
          v-for="(cat, index) in summary.top_categories.slice(0, 6)"
          :key="cat.name"
          class="flex items-center gap-2"
        >
          <span class="text-xs text-gray-400 w-4 text-right">{{ index + 1 }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between mb-0.5">
              <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ cat.name }}</span>
              <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ cat.count }}次</span>
            </div>
            <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="h-full bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] rounded-full"
                :style="{ width: `${(cat.count / summary.top_categories[0].count) * 100}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 常看UP主 + 标题热词 + 周内分布 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 常看UP主 -->
      <div v-if="summary.top_authors?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          常看UP主
        </h4>
        <div class="space-y-2">
          <div
            v-for="(author, index) in summary.top_authors.slice(0, 5)"
            :key="author.mid"
            class="flex items-center gap-2"
          >
            <span class="text-xs text-gray-400 w-4 text-right">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ author.name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ author.count }}次</span>
              </div>
              <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] rounded-full"
                  :style="{ width: `${(author.count / summary.top_authors[0].count) * 100}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 标题热词 -->
      <div v-if="summary.title_keywords?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
          </svg>
          标题热词
        </h4>
        <div class="flex flex-wrap gap-1.5">
          <span
            v-for="(kw, index) in summary.title_keywords.slice(0, 12)"
            :key="kw.word"
            class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium transition-colors"
            :class="index < 3 ? 'bg-[#fb7299]/15 text-[#fb7299]' : index < 6 ? 'bg-[#fb7299]/10 text-[#fb7299]/80' : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'"
          >
            {{ kw.word }}
            <span class="ml-1 text-[9px] opacity-60">{{ kw.count }}</span>
          </span>
        </div>
      </div>

      <!-- 周内分布 -->
      <div v-if="summary.weekday_dist?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          周内分布
        </h4>
        <div class="flex items-end gap-2 h-20">
          <div
            v-for="day in summary.weekday_dist"
            :key="day.name"
            class="flex-1 flex flex-col items-center gap-1 group relative"
            style="height: 100%;"
          >
            <div class="absolute -top-6 left-1/2 -translate-x-1/2 bg-gray-800 text-white text-[10px] px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-10">
              {{ day.name }}: {{ day.count }}
            </div>
            <div class="w-full mt-auto bg-gradient-to-t from-[#fb7299] to-[#fc9b7a] rounded-t transition-all duration-300 hover:opacity-80"
                 :style="{ height: `${Math.max((day.count / maxWeekdayCount) * 100, 4)}%` }"
            ></div>
            <span class="text-[10px] text-gray-500 dark:text-gray-400">{{ day.name }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 额外分析维度 -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
      <!-- 深夜占比 -->
      <div v-if="summary.late_night_ratio !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-purple-500">{{ (summary.late_night_ratio * 100).toFixed(0) }}%</div>
        <div class="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5">深夜观看</div>
      </div>
      <!-- 收藏率 -->
      <div v-if="summary.favorite_rate !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-amber-500">{{ (summary.favorite_rate * 100).toFixed(0) }}%</div>
        <div class="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5">收藏率</div>
      </div>
      <!-- 弃看率 -->
      <div v-if="summary.abandon_rate !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-red-400">{{ (summary.abandon_rate * 100).toFixed(0) }}%</div>
        <div class="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5">弃看率</div>
      </div>
      <!-- 黄金时段集中度 -->
      <div v-if="summary.golden_slot_ratio !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-orange-400">{{ (summary.golden_slot_ratio * 100).toFixed(0) }}%</div>
        <div class="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5">黄金3h集中度</div>
      </div>
      <!-- UP主多样性 -->
      <div v-if="summary.up_diversity !== undefined" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-cyan-500">{{ (summary.up_diversity * 100).toFixed(0) }}%</div>
        <div class="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5">UP主多样性</div>
      </div>
      <!-- 重刷次数 -->
      <div v-if="summary.rewatch_stats?.total_rewatched" class="glass-card p-3 text-center">
        <div class="text-lg font-bold text-green-500">{{ summary.rewatch_stats.total_rewatched }}</div>
        <div class="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5">重刷次数</div>
      </div>
    </div>

    <!-- 时长偏好 + 完播率分布 + 活跃时段 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <!-- 时长偏好 -->
      <div v-if="summary.duration_pref?.short + summary.duration_pref?.mid + summary.duration_pref?.long > 0" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">时长偏好</h4>
        <div class="flex h-4 rounded-full overflow-hidden">
          <div class="bg-green-400 transition-all" :style="{ width: `${summary.duration_pref.short_ratio * 100}%` }" title="短视频"></div>
          <div class="bg-[#fb7299] transition-all" :style="{ width: `${summary.duration_pref.mid_ratio * 100}%` }" title="中视频"></div>
          <div class="bg-blue-500 transition-all" :style="{ width: `${summary.duration_pref.long_ratio * 100}%` }" title="长视频"></div>
        </div>
        <div class="flex justify-between text-[10px] text-gray-500 dark:text-gray-400 mt-2">
          <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-green-400 inline-block"></span>短(&lt;5min) {{ summary.duration_pref.short }}</span>
          <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-[#fb7299] inline-block"></span>中(5-20min) {{ summary.duration_pref.mid }}</span>
          <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-blue-500 inline-block"></span>长(&gt;20min) {{ summary.duration_pref.long }}</span>
        </div>
      </div>

      <!-- 完播率分布 -->
      <div v-if="summary.completion_dist?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">完播率分布</h4>
        <div class="space-y-1.5">
          <div v-for="item in summary.completion_dist" :key="item.range" class="flex items-center gap-2">
            <span class="text-[10px] text-gray-500 dark:text-gray-400 w-12 text-right">{{ item.range }}</span>
            <div class="flex-1 h-3 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div class="h-full bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] rounded-full" :style="{ width: `${(item.count / maxCompCount) * 100}%` }"></div>
            </div>
            <span class="text-[10px] text-gray-500 dark:text-gray-400 w-6">{{ item.count }}</span>
          </div>
        </div>
      </div>

      <!-- 活跃时段 -->
      <div v-if="summary.top_time_slots?.length" class="glass-card p-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">活跃时段</h4>
        <div class="space-y-2">
          <div v-for="(slot, index) in summary.top_time_slots" :key="slot.name" class="flex items-center gap-2">
            <span class="text-xs text-gray-400 w-4 text-right">{{ index + 1 }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <span class="text-sm text-gray-700 dark:text-gray-300">{{ slot.name }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ slot.count }}次</span>
              </div>
              <div class="h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div class="h-full bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] rounded-full" :style="{ width: `${(slot.count / summary.top_time_slots[0].count) * 100}%` }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 重刷视频列表 -->
    <div v-if="summary.rewatch_stats?.rewatched_videos?.length" class="glass-card p-4">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        重刷视频
      </h4>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
        <div
          v-for="(video, index) in summary.rewatch_stats.rewatched_videos"
          :key="video.bvid"
          class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
        >
          <span class="text-xs text-gray-400 w-4 text-right font-mono">{{ index + 1 }}</span>
          <img
            v-if="video.cover"
            :src="normalizeImageUrl(video.cover)"
            class="w-16 h-10 rounded object-cover flex-shrink-0"
            loading="lazy"
          />
          <div class="flex-1 min-w-0">
            <div class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ video.title }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ video.author_name }}</div>
          </div>
          <div class="text-right flex-shrink-0">
            <div class="text-sm font-semibold text-green-500">{{ video.count }}次</div>
            <div class="text-[10px] text-gray-400">{{ formatDurationShort(video.total_duration) }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 最长观看视频 -->
    <div v-if="summary.top_videos?.length" class="glass-card p-4">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4 text-[#fb7299]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        最长观看
      </h4>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
        <div
          v-for="(video, index) in summary.top_videos"
          :key="video.bvid"
          class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
        >
          <span class="text-xs text-gray-400 w-4 text-right font-mono">{{ index + 1 }}</span>
          <img
            v-if="video.cover"
            :src="normalizeImageUrl(video.cover)"
            class="w-16 h-10 rounded object-cover flex-shrink-0"
            loading="lazy"
          />
          <div class="flex-1 min-w-0">
            <div class="text-sm text-gray-700 dark:text-gray-300 truncate">{{ video.title }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ video.author_name }}</div>
          </div>
          <div class="text-right flex-shrink-0">
            <div class="text-sm font-semibold text-[#fb7299]">{{ formatDurationShort(video.duration) }}</div>
            <div class="text-[10px] text-gray-400">{{ formatTimestamp(video.view_at) }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { formatDurationShort, formatTimestamp } from '~/utils/format'
import { normalizeImageUrl } from '~/utils/imageUrl'

const props = defineProps({
  summary: {
    type: Object,
    required: true,
  },
})

const maxDailyCount = computed(() => {
  if (!props.summary.daily_breakdown?.length) return 1
  return Math.max(...props.summary.daily_breakdown.map(d => d.count), 1)
})

const maxHourCount = computed(() => {
  if (!props.summary.hour_dist) return 1
  return Math.max(...Object.values(props.summary.hour_dist), 1)
})

const maxCompCount = computed(() => {
  if (!props.summary.completion_dist?.length) return 1
  return Math.max(...props.summary.completion_dist.map(d => d.count), 1)
})

const maxWeekdayCount = computed(() => {
  if (!props.summary.weekday_dist?.length) return 1
  return Math.max(...props.summary.weekday_dist.map(d => d.count), 1)
})
</script>

<style scoped>
.glass-card {
  @apply bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm
    rounded-xl border border-gray-200/50 dark:border-gray-700/50
    shadow-sm;
}
</style>
