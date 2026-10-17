<!-- 视频时长分析页组件 -->
<template>
  <div v-if="viewingData" class="space-y-6">
    <h3 class="text-center text-4xl font-bold text-accent">视频时长分析</h3>

    <div
      v-if="viewingData?.insights?.duration_preference"
      class="mb-8 text-center text-lg text-gray-600 dark:text-gray-300"
      v-html="formatInsightText(viewingData.insights.duration_preference)"
    />

    <!-- 时长分布图表 -->
    <div
      class="rounded-xl border border-gray-300/50 bg-white/50 p-6 backdrop-blur-sm dark:border-gray-500/50 dark:bg-white/5"
    >
      <h4 class="mb-4 text-xl font-bold text-accent">时长分布</h4>
      <div class="h-[300px]">
        <v-chart
          ref="chartRef"
          class="h-full w-full"
          :option="durationDistributionOption"
          autoresize
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import gsap from 'gsap';
import VChart from 'vue-echarts';
import { useDarkMode } from '~/stores/darkMode';
import { DURATION_COLORS } from '~/utils/chartColors';
import { escapeHtml } from '~/utils/escapeHtml';

const props = defineProps({
  viewingData: {
    type: Object,
    required: true
  }
});

const chartRef = ref(null);
const { isDarkMode } = useDarkMode();

const durationDistributionOption = computed(() => {
  if (!props.viewingData?.duration_correlation) return {};

  const periods = ['凌晨', '上午', '下午', '晚上'];
  const types = ['短视频', '中等视频', '长视频'];
  const data = periods
    .map((period) => {
      const periodData = props.viewingData.duration_correlation[period];
      return types.map((type) => ({
        period,
        type,
        value: periodData[type].video_count,
        avg_duration: Math.round(periodData[type].avg_duration / 60)
      }));
    })
    .flat();

  const typeColors = {
    短视频: DURATION_COLORS['短视频'],
    中等视频: DURATION_COLORS['中等视频'],
    长视频: DURATION_COLORS['长视频']
  };

  const isDark = !!(isDarkMode && isDarkMode.value);
  const legendTextColor = isDark ? '#bbbbbb' : '#666';
  const axisLabelColor = isDark ? '#bbbbbb' : '#999';
  const axisLineColor = isDark ? '#888888' : '#666';
  const splitLineColor = isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.1)';
  const tooltipBg = isDark ? 'rgba(28, 28, 28, 0.9)' : 'rgba(255, 255, 255, 0.95)';
  const tooltipText = isDark ? '#ffffff' : '#111111';

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: tooltipBg,
      borderColor: 'var(--accent)',
      textStyle: { color: tooltipText },
      formatter: (params) => {
        const period = params[0].axisValue;
        let result = `${period}<br/>`;
        params.forEach((param) => {
          const data = param.data;
          result += `${param.seriesName}：${data.value}个 (平均${data.avg_duration}分钟)<br/>`;
        });
        return result;
      }
    },
    legend: {
      data: types,
      top: 0,
      textStyle: { color: legendTextColor },
      itemStyle: {
        borderWidth: 0
      }
    },
    grid: {
      top: '15%',
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: periods,
      axisLine: { lineStyle: { color: axisLineColor } },
      axisLabel: { color: axisLabelColor }
    },
    yAxis: {
      type: 'value',
      name: '视频数量',
      axisLine: { lineStyle: { color: axisLineColor } },
      axisLabel: { color: axisLabelColor },
      splitLine: { lineStyle: { color: splitLineColor } }
    },
    series: types.map((type) => ({
      name: type,
      type: 'bar',
      stack: 'total',
      emphasis: {
        focus: 'series'
      },
      itemStyle: {
        color: typeColors[type]
      },
      data: periods.map((period) => {
        return data.find((d) => d.period === period && d.type === type);
      })
    }))
  };
});

onMounted(() => {
  if (chartRef.value) {
    gsap.from(chartRef.value.$el, {
      opacity: 0,
      y: 20,
      duration: 0.5,
      ease: 'power2.out',
      delay: 0.2
    });
  }
});

// 格式化洞察文本，为数字添加颜色
const formatInsightText = (text) => {
  if (!text) return '';
  return escapeHtml(text).replace(/(\d+(\.\d+)?)/g, '<span class="text-accent">$1</span>');
};
</script>
