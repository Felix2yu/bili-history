<!-- 数据概览页组件 -->
<template>
  <div class="space-y-12">
    <h3 class="text-3xl font-bold text-center bg-gradient-to-r from-[#fb7299] to-[#fc9b7a] bg-clip-text text-transparent">
      年度观看数据
    </h3>

    <div class="text-base text-center text-gray-600 dark:text-gray-300 space-y-3">
      <!-- 总体活动总结（放在最前面） -->
      <div v-if="viewingData?.insights?.overall_activity"
        v-html="formatInsightText(viewingData.insights.overall_activity)"
      >
      </div>

      <!-- 按指定顺序合并所有总结，用逗号分隔 -->
      <div>
        <span v-if="viewingBehaviorData?.report?.total_summary" v-html="formatInsightText(viewingBehaviorData.report.total_summary)"></span>
        <span v-if="viewingBehaviorData?.report?.total_summary && viewingBehaviorData?.report?.category_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.category_summary" v-html="formatInsightText(viewingBehaviorData.report.category_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary) && viewingBehaviorData?.report?.device_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.device_summary" v-html="formatInsightText(viewingBehaviorData.report.device_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary || viewingBehaviorData?.report?.device_summary) && viewingBehaviorData?.report?.up_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.up_summary" v-html="formatInsightText(viewingBehaviorData.report.up_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary || viewingBehaviorData?.report?.device_summary || viewingBehaviorData?.report?.up_summary) && viewingBehaviorData?.report?.time_slot_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.time_slot_summary" v-html="formatInsightText(viewingBehaviorData.report.time_slot_summary)"></span>
        <span v-if="(viewingBehaviorData?.report?.total_summary || viewingBehaviorData?.report?.category_summary || viewingBehaviorData?.report?.device_summary || viewingBehaviorData?.report?.up_summary || viewingBehaviorData?.report?.time_slot_summary) && viewingBehaviorData?.report?.late_night_summary">, </span>
        <span v-if="viewingBehaviorData?.report?.late_night_summary" v-html="formatInsightText(viewingBehaviorData.report.late_night_summary)"></span>
      </div>
    </div>

    <!-- 年度观看热力图 -->
    <div class="space-y-8">
      <HtmlHeatmap :year="viewingData?.year" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { getViewingBehavior } from '~/utils/api'
import HtmlHeatmap from './HtmlHeatmap.vue'

const props = defineProps({
  viewingData: {
    type: Object,
    required: true
  }
})

const viewingBehaviorData = ref(null)

// 格式化洞察文本，为数字添加颜色
const formatInsightText = (text) => {
  if (!text) return '';
  return text.replace(/(\d+(\.\d+)?)/g, '<span class="text-[#fb7299]">$1</span>')
}

// 获取观看行为数据
const fetchViewingBehavior = async (year) => {
  if (!year) return

  try {
    const response = await getViewingBehavior(year, true)
    if (response.data && response.data.status === 'success') {
      viewingBehaviorData.value = response.data.data
    }
  } catch (error) {
    console.error('获取观看行为数据失败:', error)
  }
}

onMounted(() => {
  if (props.viewingData?.year) {
    fetchViewingBehavior(props.viewingData.year)
  }
})

// 监听年份变化
watch(() => props.viewingData?.year, (newYear) => {
  if (newYear) {
    fetchViewingBehavior(newYear)
  }
}, { immediate: true })
</script>
