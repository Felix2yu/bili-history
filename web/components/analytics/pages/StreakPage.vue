<!-- 连续观看记录页组件 -->
<template>
  <div v-if="viewingData" class="space-y-6">
    <h3 class="text-center text-4xl font-bold text-accent">连续观看记录</h3>

    <div
      v-if="viewingData?.insights?.continuity"
      class="mb-8 text-center text-lg text-gray-600 dark:text-gray-300"
      v-html="formatInsightText(viewingData.insights.continuity)"
    />

    <div v-if="viewingData?.viewing_continuity" class="grid grid-cols-2 gap-8">
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-6 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <div class="streak-number text-4xl font-bold text-accent">
          {{ viewingData.viewing_continuity.max_streak || 0 }}天
        </div>
        <div class="mt-2 text-lg text-gray-600 dark:text-gray-400">最长连续观看</div>
        <div
          v-if="viewingData.viewing_continuity.longest_streak_period"
          class="mt-4 text-sm text-gray-500"
        >
          {{ viewingData.viewing_continuity.longest_streak_period.start }} 至
          {{ viewingData.viewing_continuity.longest_streak_period.end }}
        </div>
      </div>
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-6 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <div class="streak-number text-4xl font-bold text-accent">
          {{ viewingData.viewing_continuity.current_streak || 0 }}天
        </div>
        <div class="mt-2 text-lg text-gray-600 dark:text-gray-400">当前连续观看</div>
        <div
          v-if="viewingData.viewing_continuity.current_streak_start"
          class="mt-4 text-sm text-gray-500"
        >
          开始于 {{ viewingData.viewing_continuity.current_streak_start }}
        </div>
      </div>
    </div>

    <!-- 数据加载中或无数据时的提示 -->
    <div v-else class="text-center text-gray-500 dark:text-gray-400">正在加载连续观看数据...</div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import gsap from 'gsap';
import { escapeHtml } from '~/utils/escapeHtml';

// viewingData 仅模板使用（script 内不引用），故不赋值给变量
defineProps({
  viewingData: {
    type: Object,
    required: true
  }
});

onMounted(() => {
  // 数字动画效果
  const streakNumbers = document.querySelectorAll('.streak-number');
  streakNumbers.forEach((el) => {
    const finalValue = parseInt(el.textContent);
    gsap.fromTo(
      el,
      { textContent: 0 },
      {
        duration: 2,
        textContent: finalValue,
        snap: { textContent: 1 },
        ease: 'power1.out'
      }
    );
  });
});

// 格式化洞察文本，为数字添加颜色
const formatInsightText = (text) => {
  if (!text) return '';
  return escapeHtml(text).replace(/(\d+(\.\d+)?)/g, '<span class="text-accent">$1</span>');
};
</script>
