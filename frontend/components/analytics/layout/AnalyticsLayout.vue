<!-- 年度总结通用布局组件 -->
<template>
  <div class="absolute inset-0 bg-gradient-to-b from-[#fef6f9] to-[#fff9fa] dark:from-gray-900 dark:to-gray-950 overflow-hidden">
    <!-- 背景装饰 -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -right-40 w-96 h-96 bg-gradient-to-br from-accent/10 to-[#fc9b7a]/10 rounded-full blur-3xl animate-float"></div>
      <div class="absolute -bottom-20 -left-20 w-80 h-80 bg-gradient-to-tr from-[#fc9b7a]/10 to-accent/10 rounded-full blur-3xl animate-float-delay"></div>
      <div class="absolute top-1/3 left-1/4 w-64 h-64 bg-accent/5 rounded-full blur-3xl animate-float" style="animation-delay: -2s"></div>
    </div>

    <div class="relative min-h-screen flex items-center justify-center py-8 pb-20 md:pb-8">
      <div class="max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8">
        <slot></slot>
      </div>
    </div>
  </div>
</template>

<script setup>
</script>

<style>
@keyframes float {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-20px, 20px); }
}

@keyframes float-delay {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, -20px); }
}

.animate-float {
  animation: float 8s ease-in-out infinite;
}

.animate-float-delay {
  animation: float 8s ease-in-out infinite;
  animation-delay: -4s;
}
</style>
