<template>
  <ReportPage />
</template>

<script setup>
import ReportPage from '~/components/page/ReportPage.vue';

definePageMeta({
  layout: 'main'
});

useHead({
  title: '数据概览 - BiliHistory'
});
</script>
