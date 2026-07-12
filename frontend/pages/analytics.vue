<template>
  <div></div>
</template>

<script setup>
definePageMeta({
  layout: 'main',
})

// 重定向到数据概览的年度标签
navigateTo('/report', { replace: true })
</script>
