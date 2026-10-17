<!-- 年度总结通用布局组件 -->
<template>
  <div class="overflow-hidden rounded-xl bg-gray-50/30 dark:from-gray-900 dark:to-gray-950">
    <!-- 背景装饰 -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute -top-40 -right-40 h-96 w-96 animate-float rounded-full bg-accent/10 blur-3xl"
      />
      <div
        class="animate-float-delay absolute -bottom-20 -left-20 h-80 w-80 rounded-full bg-accent/10 blur-3xl"
      />
    </div>

    <div class="relative py-6 pb-16 md:pb-6">
      <div class="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8">
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup></script>

<style>
@keyframes float {
  0%,
  100% {
    transform: translate(0, 0);
  }
  50% {
    transform: translate(-20px, 20px);
  }
}

@keyframes float-delay {
  0%,
  100% {
    transform: translate(0, 0);
  }
  50% {
    transform: translate(20px, -20px);
  }
}

.animate-float {
  animation: float 8s ease-in-out infinite;
}

.animate-float-delay {
  animation: float 8s ease-in-out infinite;
  animation-delay: -4s;
}
</style>
