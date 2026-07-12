<template>
  <ReportPage />
</template>

<script setup>
definePageMeta({
  layout: 'main',
})

import ReportPage from '~/components/page/ReportPage.vue'

useHead({
  title: '周报月报 - BiliHistory',
})
</script>
