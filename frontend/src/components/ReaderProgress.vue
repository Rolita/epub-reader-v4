<template>
  <div 
    class="reader-progress" 
    :class="{ 'hidden': isHidden, 'visible': visible }" 
    @click="handleClick"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <div class="reader-progress-bar" :style="{ width: progress + '%' }"></div>
    <span class="progress-text">{{ progressText }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  progress: number
  isHidden?: boolean
  visible?: boolean
}>()

const emit = defineEmits<{
  (e: 'click', percent: number): void
  (e: 'mouseenter'): void
  (e: 'mouseleave'): void
}>()

const progressText = computed(() => `${props.progress}%`)

const handleClick = (e: MouseEvent) => {
  const target = e.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  const x = e.clientX - rect.left
  const percent = x / rect.width
  emit('click', percent)
}

const handleMouseEnter = () => {
  emit('mouseenter')
}

const handleMouseLeave = () => {
  emit('mouseleave')
}
</script>

<style scoped>
.reader-progress {
  position: absolute;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  width: 60%;
  height: 6px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 3px;
  cursor: pointer;
  opacity: 0;
  transition: opacity var(--transition-normal);
  display: flex;
  align-items: center;
  pointer-events: none;
}

.reader-progress.visible {
  opacity: 1;
  pointer-events: auto;
}

.reader-progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--primary-color) 0%, var(--accent-color) 100%);
  border-radius: 3px;
  transition: width var(--transition-fast);
}

.progress-text {
  position: absolute;
  right: -50px;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
  white-space: nowrap;
}

.hidden {
  visibility: hidden;
  pointer-events: none;
}
</style>
