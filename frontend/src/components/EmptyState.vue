<template>
  <div class="empty-layout">
    <div class="empty-content">
      <div class="empty-icon-wrapper">
        <div class="empty-icon">
          <div class="book-spine"></div>
          <div class="book-cover">
            <div class="book-face">
              <div class="eye eye-left"></div>
              <div class="eye eye-right"></div>
            </div>
            <div class="zzz">z<span>z</span>z</div>
          </div>
        </div>
      </div>
      <h2 class="empty-title">EPUB Reader</h2>
      <p class="empty-desc">点击左侧 <span class="highlight">+</span> 创建书架，导入 EPUB 开始阅读。</p>
      <button class="empty-action-btn" @click="$emit('action')">
        开始阅读 
      </button>
    </div>
    <div class="empty-decoration">
      <!-- 背景渐变层 -->
      <div class="bg-gradient"></div>
      
      <!-- 动态光斑 -->
      <div class="light-spot light-spot-1"></div>
      <div class="light-spot light-spot-2"></div>
      <div class="light-spot light-spot-3"></div>
      
      <!-- 流动光带 -->
      <div class="light-stream light-stream-1"></div>
      <div class="light-stream light-stream-2"></div>
      
      <!-- 几何网格 -->
      <div class="geometric-grid"></div>
      
      <!-- 浮动几何形状 -->
      <div class="floating-shape shape-1"></div>
      <div class="floating-shape shape-2"></div>
      <div class="floating-shape shape-3"></div>
      <div class="floating-shape shape-4"></div>
      
      <!-- 粒子效果 -->
      <div class="particle-container">
        <div v-for="i in 20" :key="i" class="particle" :style="{ 
          left: Math.random() * 100 + '%', 
          top: Math.random() * 100 + '%',
          animationDelay: Math.random() * 10 + 's',
          animationDuration: (8 + Math.random() * 12) + 's',
          width: (2 + Math.random() * 4) + 'px',
          height: (2 + Math.random() * 4) + 'px'
        }"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  (e: 'action'): void
}>()
</script>

<style scoped>
.empty-layout {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  user-select: none;
  position: relative;
  overflow: hidden;
}

.empty-content {
  text-align: center;
  z-index: 1;
  position: relative;
}

.empty-icon-wrapper {
  margin-bottom: 32px;
  perspective: 600px;
}

.empty-icon {
  position: relative;
  width: 120px;
  height: 90px;
  margin: 0 auto;
  transform-style: preserve-3d;
  transform: rotateX(-15deg) rotateY(-15deg);
}

.book-spine {
  position: absolute;
  left: 0;
  top: 0;
  width: 24px;
  height: 100%;
  background: linear-gradient(180deg, var(--primary-color) 0%, var(--accent-color) 100%);
  border-radius: 6px 0 0 6px;
  box-shadow: inset -4px 0 12px rgba(0, 0, 0, 0.2);
}

.book-cover {
  position: absolute;
  left: 24px;
  top: 0;
  width: 96px;
  height: 100%;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15) 0%, rgba(99, 102, 241, 0.25) 50%, rgba(99, 102, 241, 0.35) 100%);
  border-radius: 0 6px 6px 0;
  box-shadow: 8px 8px 24px rgba(99, 102, 241, 0.2),
              4px 4px 12px rgba(99, 102, 241, 0.1);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 10px;
}

.book-face {
  position: relative;
  width: 40px;
  height: 24px;
}

.eye {
  position: absolute;
  width: 12px;
  height: 12px;
  background: var(--primary-color);
  border-radius: 50%;
  top: 0;
  animation: blink 4s infinite;
}

.eye-left {
  left: 2px;
}

.eye-right {
  right: 2px;
}

@keyframes blink {
  0%, 45%, 55%, 100% {
    transform: scaleY(1);
  }
  50% {
    transform: scaleY(0.1);
  }
}

.zzz {
  position: absolute;
  right: -35px;
  top: 15px;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--primary-color);
  opacity: 0.7;
  animation: float 3s ease-in-out infinite;
}

.zzz span {
  font-size: 0.9rem;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(-10deg);
  }
  50% {
    transform: translateY(-8px) rotate(-10deg);
  }
}

.empty-title {
  font-size: 1.8rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
  letter-spacing: -0.02em;
}

.empty-desc {
  font-size: 0.95rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin-bottom: 28px;
}

.empty-desc .highlight {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  background: var(--primary-light);
  color: var(--primary-color);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  font-weight: 600;
  font-size: 1.1rem;
  margin: 0 4px;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.empty-desc .highlight:hover {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  transform: scale(1.05);
}

.empty-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 28px;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.3);
}

.empty-action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(99, 102, 241, 0.4);
}

.empty-action-btn:active {
  transform: translateY(0);
}

.empty-action-btn .arrow {
  transition: transform 0.3s ease;
}

.empty-action-btn:hover .arrow {
  transform: translateX(4px);
}

.empty-decoration {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  overflow: hidden;
  z-index: 0;
}

/* 背景渐变层 */
.bg-gradient {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    radial-gradient(ellipse at 20% 20%, color-mix(in srgb, var(--primary-color) 12%, transparent) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 80%, color-mix(in srgb, var(--accent-color) 8%, transparent) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, color-mix(in srgb, var(--primary-color) 5%, transparent) 0%, transparent 70%);
}

/* 动态光斑 */
.light-spot {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.6;
}

.light-spot-1 {
  width: 400px;
  height: 400px;
  top: -150px;
  right: -100px;
  background: radial-gradient(circle, color-mix(in srgb, var(--primary-color) 25%, transparent) 0%, color-mix(in srgb, var(--primary-color) 8%, transparent) 50%, transparent 70%);
  animation: pulse-glow-1 8s ease-in-out infinite;
}

.light-spot-2 {
  width: 300px;
  height: 300px;
  bottom: -100px;
  left: -80px;
  background: radial-gradient(circle, color-mix(in srgb, var(--accent-color) 20%, transparent) 0%, color-mix(in srgb, var(--accent-color) 6%, transparent) 50%, transparent 70%);
  animation: pulse-glow-2 10s ease-in-out infinite;
}

.light-spot-3 {
  width: 250px;
  height: 250px;
  top: 40%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: radial-gradient(circle, color-mix(in srgb, var(--primary-color) 15%, transparent) 0%, transparent 60%);
  animation: pulse-glow-3 6s ease-in-out infinite;
}

@keyframes pulse-glow-1 {
  0%, 100% {
    transform: scale(1);
    opacity: 0.6;
  }
  50% {
    transform: scale(1.1);
    opacity: 0.8;
  }
}

@keyframes pulse-glow-2 {
  0%, 100% {
    transform: scale(1) rotate(0deg);
    opacity: 0.5;
  }
  50% {
    transform: scale(1.2) rotate(5deg);
    opacity: 0.7;
  }
}

@keyframes pulse-glow-3 {
  0%, 100% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.2;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.3);
    opacity: 0.35;
  }
}

/* 流动光带 */
.light-stream {
  position: absolute;
  width: 200px;
  height: 4px;
  background: linear-gradient(90deg, transparent 0%, color-mix(in srgb, var(--primary-color) 20%, transparent) 50%, transparent 100%);
  filter: blur(2px);
  border-radius: 2px;
}

.light-stream-1 {
  top: 30%;
  left: -200px;
  animation: stream-flow-1 15s linear infinite;
  transform: rotate(-15deg);
}

.light-stream-2 {
  bottom: 40%;
  right: -200px;
  animation: stream-flow-2 18s linear infinite;
  transform: rotate(20deg);
}

@keyframes stream-flow-1 {
  0% {
    transform: translateX(0) rotate(-15deg);
  }
  100% {
    transform: translateX(calc(100vw + 400px)) rotate(-15deg);
  }
}

@keyframes stream-flow-2 {
  0% {
    transform: translateX(0) rotate(20deg);
  }
  100% {
    transform: translateX(calc(-100vw - 400px)) rotate(20deg);
  }
}

/* 几何网格 */
.geometric-grid {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: 
    linear-gradient(color-mix(in srgb, var(--primary-color) 2%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--primary-color) 2%, transparent) 1px, transparent 1px);
  background-size: 60px 60px;
  animation: grid-move 20s linear infinite;
}

@keyframes grid-move {
  0% {
    transform: translate(0, 0);
  }
  100% {
    transform: translate(60px, 60px);
  }
}

/* 浮动几何形状 */
.floating-shape {
  position: absolute;
  opacity: 0.08;
}

.shape-1 {
  width: 80px;
  height: 80px;
  top: 25%;
  left: 15%;
  background: linear-gradient(135deg, color-mix(in srgb, var(--primary-color) 30%, transparent) 0%, transparent 50%);
  border-radius: 20px;
  transform: rotate(45deg);
  animation: float-shape-1 12s ease-in-out infinite;
}

.shape-2 {
  width: 60px;
  height: 60px;
  bottom: 30%;
  right: 20%;
  border: 2px solid color-mix(in srgb, var(--accent-color) 20%, transparent);
  border-radius: 50%;
  animation: float-shape-2 10s ease-in-out infinite;
}

.shape-3 {
  width: 100px;
  height: 40px;
  top: 60%;
  left: 25%;
  background: linear-gradient(90deg, color-mix(in srgb, var(--primary-color) 25%, transparent) 0%, transparent 100%);
  border-radius: 20px;
  transform: rotate(-10deg);
  animation: float-shape-3 14s ease-in-out infinite;
}

.shape-4 {
  width: 50px;
  height: 50px;
  bottom: 20%;
  left: 18%;
  background: color-mix(in srgb, var(--primary-color) 15%, transparent);
  clip-path: polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%);
  animation: float-shape-4 8s ease-in-out infinite;
}

@keyframes float-shape-1 {
  0%, 100% {
    transform: rotate(45deg) translateY(0) translateX(0);
  }
  25% {
    transform: rotate(45deg) translateY(-15px) translateX(10px);
  }
  50% {
    transform: rotate(45deg) translateY(-8px) translateX(-5px);
  }
  75% {
    transform: rotate(45deg) translateY(-20px) translateX(5px);
  }
}

@keyframes float-shape-2 {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  25% {
    transform: translateY(-12px) rotate(90deg);
  }
  50% {
    transform: translateY(-5px) rotate(180deg);
  }
  75% {
    transform: translateY(-18px) rotate(270deg);
  }
}

@keyframes float-shape-3 {
  0%, 100% {
    transform: rotate(-10deg) translateY(0);
  }
  50% {
    transform: rotate(-10deg) translateY(-25px);
  }
}

@keyframes float-shape-4 {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-18px) rotate(180deg);
  }
}

/* 粒子容器 */
.particle-container {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.particle {
  position: absolute;
  background: color-mix(in srgb, var(--primary-color) 30%, transparent);
  border-radius: 50%;
  animation: particle-float 10s ease-in-out infinite;
}

@keyframes particle-float {
  0%, 100% {
    transform: translateY(0) translateX(0) scale(1);
    opacity: 0;
  }
  10% {
    opacity: 0.6;
  }
  50% {
    transform: translateY(-100px) translateX(30px) scale(1.2);
    opacity: 0.4;
  }
  90% {
    opacity: 0.2;
  }
}

.empty-emoji {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-top: 12px;
}
</style>
