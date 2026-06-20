<template>
  <div class="loading-overlay">
    <div class="loader simple-lines">
      <div class="line line-1"></div>
      <div class="line line-2"></div>
      <div class="line line-3"></div>
    </div>
    <p class="loading-text">正在打开书籍...</p>
    <p class="loading-subtext">请稍候，精彩内容即将呈现</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  text?: string
  subtext?: string
}>()
</script>

<style scoped>
.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: 
    radial-gradient(
      circle at center, 
      rgba(var(--primary-color-rgb, 91, 76, 219), 0.06) 0%, 
      var(--bg-color) 55%, 
      rgba(0, 0, 0, 0.02) 100%
    );
  z-index: 1000;
  animation: fadeIn 0.5s ease-in-out;
  will-change: opacity;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.loader.simple-lines {
  position: relative;
  width: 60px;
  height: 30px;
  margin-bottom: 32px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transform: translateZ(0);
}

.line {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  box-shadow: 0 0 8px currentColor;
  transform: translateZ(0);
  will-change: transform, opacity;
  animation: bounce 1.3s ease-in-out infinite;
}

.line-1 {
  background-color: var(--primary-color);
  color: var(--primary-color);
  animation-delay: 0s;
}
.line-2 {
  background-color: var(--accent-color, #ff6b6b);
  color: var(--accent-color, #ff6b6b);
  animation-delay: 0.25s;
}
.line-3 {
  background-color: var(--text-primary);
  color: var(--text-primary);
  animation-delay: 0.5s;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0) translateZ(0);
    opacity: 0.6;
  }
  50% {
    transform: translateY(-12px) translateZ(0);
    opacity: 1;
  }
}

.loading-text {
  font-size: 1.2rem;
  color: var(--text-primary);
  font-weight: 500;
  letter-spacing: 2px;
  margin-bottom: 8px;
  text-shadow: 0 0 20px rgba(var(--primary-color-rgb, 91, 76, 219), 0.3);
  animation: textPulse 2s ease-in-out infinite;
  transform: translateZ(0);
}

@keyframes textPulse {
  0%, 100% {
    opacity: 0.8;
    text-shadow: 0 0 20px rgba(var(--primary-color-rgb, 91, 76, 219), 0.3);
  }
  50% {
    opacity: 1;
    text-shadow: 0 0 30px rgba(var(--primary-color-rgb, 91, 76, 219), 0.5),
                 0 0 40px rgba(var(--primary-color-rgb, 91, 76, 219), 0.3);
  }
}

.loading-subtext {
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 400;
  letter-spacing: 1px;
  opacity: 0.7;
}
</style>