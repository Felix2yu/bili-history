<!-- 标签分析页组件 -->
<template>
  <div v-if="viewingData" class="space-y-6">
    <h3 class="text-center text-4xl font-bold text-accent">标签分析</h3>

    <div class="mb-8 space-y-2 text-center text-lg text-gray-600 dark:text-gray-300">
      <div
        v-if="viewingData?.insights?.tag_preference"
        v-html="formatInsightText(viewingData.insights.tag_preference)"
      />
      <div
        v-if="viewingData?.insights?.tag_completion"
        v-html="formatInsightText(viewingData.insights.tag_completion)"
      />
    </div>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- 标签分布图表 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-6 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-4 text-xl font-bold text-accent">观看分布</h4>
        <div class="h-[280px]">
          <v-chart
            ref="distributionChartRef"
            class="h-full w-full"
            :option="tagDistributionOption"
            autoresize
          />
        </div>
      </div>

      <!-- 标签完成率图表 -->
      <div
        class="rounded-xl border border-gray-300/50 bg-white/50 p-6 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
      >
        <h4 class="mb-4 text-xl font-bold text-accent">完成率排行</h4>
        <div class="h-[280px]">
          <v-chart
            ref="completionChartRef"
            class="h-full w-full"
            :option="tagCompletionOption"
            autoresize
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import gsap from 'gsap';
import VChart from 'vue-echarts';
import { useDarkMode } from '~/stores/darkMode';
import { getChartColor } from '~/utils/chartColors';
import { escapeHtml } from '~/utils/escapeHtml';

const props = defineProps({
  viewingData: {
    type: Object,
    required: true
  }
});

const distributionChartRef = ref(null);
const completionChartRef = ref(null);

const tagDistributionOption = computed(() => {
  const { isDarkMode } = useDarkMode();
  const isDark = !!(isDarkMode && isDarkMode.value);
  const axisLabelColor = isDark ? '#bbbbbb' : '#999999';
  const axisLineColor = isDark ? '#888888' : '#666666';
  const splitLineColor = isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.1)';
  const tooltipBg = isDark ? 'rgba(28, 28, 28, 0.9)' : 'rgba(255, 255, 255, 0.95)';
  const tooltipText = isDark ? '#ffffff' : '#111111';

  if (!props.viewingData?.watch_counts?.tag_distribution) return {};

  const data = Object.entries(props.viewingData.watch_counts.tag_distribution || {})
    .sort((a, b) => b[1] - a[1])
    .map(([tag, count]) => ({
      name: tag,
      value: count
    }));

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      },
      backgroundColor: tooltipBg,
      borderColor: 'var(--accent)',
      textStyle: { color: tooltipText }
    },
    grid: {
      top: '3%',
      left: '3%',
      right: '15%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: axisLineColor } },
      axisLabel: { color: axisLabelColor },
      splitLine: { lineStyle: { color: splitLineColor } }
    },
    yAxis: {
      type: 'category',
      data: data.map((item) => item.name),
      axisLine: { lineStyle: { color: axisLineColor } },
      axisLabel: { color: axisLabelColor },
      inverse: true
    },
    series: [
      {
        name: '视频数量',
        type: 'bar',
        data: data.map((item, index) => ({
          value: item.value,
          itemStyle: { color: getChartColor(index) }
        })),
        label: {
          show: true,
          position: 'right',
          color: axisLabelColor,
          formatter: '{c} 个'
        }
      }
    ]
  };
});

const tagCompletionOption = computed(() => {
  const { isDarkMode } = useDarkMode();
  const isDark = !!(isDarkMode && isDarkMode.value);
  const axisLabelColor = isDark ? '#bbbbbb' : '#999999';
  const axisLineColor = isDark ? '#888888' : '#666666';
  const splitLineColor = isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.1)';
  const tooltipBg = isDark ? 'rgba(28, 28, 28, 0.9)' : 'rgba(255, 255, 255, 0.95)';
  const tooltipText = isDark ? '#ffffff' : '#111111';

  if (!props.viewingData?.completion_rates?.tag_completion_rates) return {};

  const data = Object.entries(props.viewingData.completion_rates.tag_completion_rates || {})
    .map(([tag, stats]) => ({
      tag,
      completion: stats.average_completion_rate,
      count: stats.video_count
    }))
    .sort((a, b) => b.completion - a.completion)
    .slice(0, 10);

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: tooltipBg,
      borderColor: 'var(--accent)',
      textStyle: { color: tooltipText },
      formatter: (params) => {
        const data = params[0].data;
        return `${params[0].name}<br/>完成率：${data.value}%<br/>视频数：${data.count}个`;
      }
    },
    grid: {
      top: '3%',
      left: '3%',
      right: '15%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      name: '完成率',
      axisLine: { lineStyle: { color: axisLineColor } },
      axisLabel: {
        color: axisLabelColor,
        formatter: '{value}%'
      },
      splitLine: { lineStyle: { color: splitLineColor } }
    },
    yAxis: {
      type: 'category',
      data: data.map((item) => `${item.tag}(${item.count}个)`),
      axisLine: { lineStyle: { color: axisLineColor } },
      axisLabel: { color: axisLabelColor },
      inverse: true
    },
    series: [
      {
        type: 'bar',
        data: data.map((item, index) => ({
          value: item.completion,
          count: item.count,
          itemStyle: { color: getChartColor(index) }
        })),
        label: {
          show: true,
          position: 'right',
          color: axisLabelColor,
          formatter: (params) => Number(params.value).toFixed(1) + '%'
        }
      }
    ]
  };
});

onMounted(() => {
  const charts = [distributionChartRef.value, completionChartRef.value];
  if (charts.every((chart) => chart)) {
    gsap.from(
      charts.map((chart) => chart.$el),
      {
        opacity: 0,
        y: 20,
        duration: 0.5,
        stagger: 0.1,
        ease: 'power2.out',
        delay: 0.2
      }
    );
  }
});

// 格式化洞察文本，为数字添加颜色
const formatInsightText = (text) => {
  if (!text) return '';
  return escapeHtml(text).replace(/(\d+(\.\d+)?)/g, '<span class="text-accent">$1</span>');
};
</script>
