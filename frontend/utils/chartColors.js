export const CHART_COLORS = [
  '#F87171',
  '#FB923C',
  '#FBBF24',
  '#A3E635',
  '#34D399',
  '#22D3EE',
  '#60A5FA',
  '#A78BFA',
  '#F472B6',
  '#FB7185',
  '#4ADE80',
  '#38BDF8'
]

export const getChartColor = (index) => {
  return CHART_COLORS[index % CHART_COLORS.length]
}

export const getChartColors = (count) => {
  const colors = []
  for (let i = 0; i < count; i++) {
    colors.push(getChartColor(i))
  }
  return colors
}

export const CATEGORY_COLORS = {
  生活: '#F87171',
  娱乐: '#FB923C',
  影视: '#FBBF24',
  动画: '#A3E635',
  游戏: '#34D399',
  科技: '#22D3EE',
  知识: '#60A5FA',
  音乐: '#A78BFA',
  舞蹈: '#F472B6',
  时尚: '#FB7185',
  美食: '#4ADE80',
  动物: '#38BDF8'
}

export const TIME_SLOT_COLORS = {
  凌晨: '#7A8BFA',
  上午: '#7AFC8C',
  下午: '#FC9B7A',
  晚上: 'var(--accent)',
  深夜: '#7A8BFA'
}

export const getTimeSlotColor = (hour) => {
  const h = typeof hour === 'string' ? parseInt(hour.replace('时', '')) : hour
  if (h >= 6 && h < 12) return TIME_SLOT_COLORS['上午']
  if (h >= 12 && h < 18) return TIME_SLOT_COLORS['下午']
  if (h >= 18 && h < 24) return TIME_SLOT_COLORS['晚上']
  return TIME_SLOT_COLORS['凌晨']
}

export const DEVICE_COLORS = {
  '手机': 'var(--accent)',
  '电脑': '#60A5FA',
  '平板': '#34D399',
  '电视': '#FBBF24',
  '未知': '#9CA3AF'
}

export const DURATION_COLORS = {
  '短视频': '#34D399',
  '中等视频': 'var(--accent)',
  '长视频': '#60A5FA'
}
