<template>
  <div class="default-cover" :class="{ compact: compact }">
    <div class="cover-bg"></div>
    <div class="spine"></div>
    <div class="cover-content">
      <div class="top-rule"></div>
      <div class="text-area">
        <div class="title" :class="titleSizeClass">{{ displayTitle }}</div>
        <div v-if="author" class="divider" :class="authorSizeClass"></div>
        <div v-if="author" class="author" :class="authorSizeClass">{{ displayAuthor }}</div>
      </div>
      <div class="bottom-rule"></div>
    </div>
    <div class="corner-accent"></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title?: string
  author?: string
  compact?: boolean
}>()

const displayTitle = computed(() => {
  const t = props.title || '未知书籍'
  return t.length > (props.compact ? 10 : 20) 
    ? t.substring(0, props.compact ? 8 : 18) + '...' 
    : t
})

const displayAuthor = computed(() => {
  const a = props.author || '未知作者'
  return a.length > (props.compact ? 8 : 15) 
    ? a.substring(0, props.compact ? 6 : 13) + '...' 
    : a
})

const titleSizeClass = computed(() => {
  const length = displayTitle.value.length
  if (props.compact) {
    if (length <= 5) return 'size-normal'
    if (length <= 8) return 'size-small'
    return 'size-xs'
  }
  if (length <= 8) return 'size-normal'
  if (length <= 14) return 'size-medium'
  if (length <= 20) return 'size-small'
  return 'size-xs'
})

const authorSizeClass = computed(() => {
  if (props.compact) return 'compact'
  const length = displayTitle.value.length
  if (length <= 8) return 'normal'
  if (length <= 14) return 'normal'
  if (length <= 20) return 'small'
  return 'xs'
})
</script>

<style scoped>
.default-cover {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 4px;
  overflow: hidden;
  box-shadow: 
    0 2px 4px rgba(0,0,0,0.05),
    0 4px 12px rgba(0,0,0,0.04),
    inset 0 0 0 1px rgba(0,0,0,0.06);
}

.cover-bg {
  position: absolute;
  inset: 0;
  background: #F5F0E8;
}

.spine {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 6px;
  background: #8B7355;
}

.cover-content {
  position: relative;
  z-index: 2;
  height: 100%;
  padding: 20px 16px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.top-rule {
  position: absolute;
  top: 20px;
  left: 24px;
  right: 16px;
  height: 1px;
  background: rgba(45, 45, 45, 0.15);
}

.bottom-rule {
  position: absolute;
  bottom: 20px;
  left: 24px;
  right: 16px;
  height: 1px;
  background: rgba(45, 45, 45, 0.1);
}

.text-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-start;
  padding-left: 8px;
}

.title {
  font-family: 'Georgia', 'Times New Roman', 'Noto Serif SC', serif;
  font-weight: 600;
  color: #2D2D2D;
  line-height: 1.5;
  letter-spacing: 0.02em;
  margin-bottom: 8px;
}

.title.size-normal {
  font-size: 20px;
}

.title.size-medium {
  font-size: 16px;
}

.title.size-small {
  font-size: 13px;
}

.title.size-xs {
  font-size: 11px;
}

.divider {
  height: 1px;
  background: rgba(45, 45, 45, 0.25);
  margin-bottom: 10px;
}

.divider.normal {
  width: 32px;
}

.divider.small {
  width: 24px;
}

.divider.xs {
  width: 16px;
}

.divider.compact {
  width: 12px;
  margin-bottom: 4px;
}

.author {
  font-family: 'Georgia', 'Times New Roman', 'Noto Serif SC', serif;
  color: #6B6B6B;
  font-style: italic;
  letter-spacing: 0.01em;
}

.author.normal {
  font-size: 13px;
}

.author.small {
  font-size: 11px;
}

.author.xs {
  font-size: 9px;
}

.author.compact {
  font-size: 7px;
}

.corner-accent {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 40px;
  height: 40px;
  border-bottom-right-radius: 4px;
  background: linear-gradient(
    135deg,
    rgba(139, 115, 85, 0.08) 0%,
    transparent 60%
  );
  pointer-events: none;
}

.default-cover.compact .cover-content {
  padding: 10px 8px;
}

.default-cover.compact .top-rule {
  top: 10px;
  left: 14px;
  right: 8px;
}

.default-cover.compact .bottom-rule {
  bottom: 10px;
  left: 14px;
  right: 8px;
}

.default-cover.compact .text-area {
  padding-left: 4px;
}

.default-cover.compact .title {
  margin-bottom: 4px;
}

.default-cover.compact .spine {
  width: 3px;
}

.default-cover.compact .corner-accent {
  width: 20px;
  height: 20px;
}
</style>