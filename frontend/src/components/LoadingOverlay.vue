<template>
  <div class="loading-overlay">
    <div class="loading-ring">
      <div class="ring-outer"></div>
      <div class="ring-inner"></div>
    </div>
    <p class="loading-text">正在打开书籍...</p>
    <p class="loading-subtext">请稍候，精彩内容即将呈现</p>
  </div>
</template>

<script setup lang="ts">
// 独立的加载动画组件
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
  background: radial-gradient(circle at center, var(--bg-color) 0%, rgba(0, 0, 0, 0.02) 100%);
  z-index: 1000;
  animation: fadeIn 0.5s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* 渐变圆环加载器 */
.loading-ring {
  position: relative;
  width: 80px;
  height: 80px;
  margin-bottom: 32px;
}

/* 外环 - 渐变旋转 */
.ring-outer {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: conic-gradient(
    from 0deg,
    transparent 0%,
    var(--primary-color) 30%,
    var(--accent-color, #ff6b6b) 70%,
    transparent 100%
  );
  mask: radial-gradient(farthest-side, transparent calc(100% - 4px), black calc(100% - 3px));
  -webkit-mask: radial-gradient(farthest-side, transparent calc(100% - 4px), black calc(100% - 3px));
  box-shadow: 
    0 0 25px rgba(var(--primary-color-rgb, 91, 76, 219), 0.4),
    inset 0 0 25px rgba(var(--primary-color-rgb, 91, 76, 219), 0.15);
  animation: rotate 1.2s linear infinite;
}

/* 内环 - 反向渐变旋转 */
.ring-inner {
  position: absolute;
  top: 10px;
  left: 10px;
  right: 10px;
  bottom: 10px;
  border-radius: 50%;
  background: conic-gradient(
    from 180deg,
    transparent 0%,
    var(--accent-color, #ff6b6b) 40%,
    var(--primary-color) 80%,
    transparent 100%
  );
  mask: radial-gradient(farthest-side, transparent calc(100% - 3px), black calc(100% - 2px));
  -webkit-mask: radial-gradient(farthest-side, transparent calc(100% - 3px), black calc(100% - 2px));
  box-shadow: 
    0 0 20px rgba(var(--accent-color-rgb, 255, 107, 107), 0.5),
    inset 0 0 20px rgba(var(--accent-color-rgb, 255, 107, 107), 0.2);
  animation: rotate-reverse 0.8s linear infinite;
}

/* 正向旋转 */
@keyframes rotate {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* 反向旋转 */
@keyframes rotate-reverse {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(-360deg);
  }
}

/* 加载文字 */
.loading-text {
  font-size: 1.2rem;
  color: var(--text-primary);
  font-weight: 500;
  letter-spacing: 2px;
  margin-bottom: 8px;
  text-shadow: 0 0 20px rgba(var(--primary-color-rgb, 91, 76, 219), 0.3);
  animation: textPulse 2s ease-in-out infinite;
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