<template>
  <ReportPage />
</template>

<script setup>
definePageMeta({
  layout: 'main',
})

import ReportPage from '~/components/page/ReportPage.vue'

useHead({
  title: '数据概览 - BiliHistory',
})
</script>
