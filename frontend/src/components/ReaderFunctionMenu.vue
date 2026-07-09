<template>
  <div
    class="reader-function-menu"
    :class="{ 'visible': visible }"
    @mouseleave="emit('mouseleave')"
  >
    <!-- 全屏按钮 -->
    <button
      v-if="!isLoading && !isFullscreen"
      @click="emit('enterFullscreen')"
      class="fullscreen-toggle-btn"
      title="全屏阅读 (F11)"
    >
      <FullscreenIcon :size="24" />
    </button>

    <!-- 退出全屏按钮 -->
    <button
      v-if="!isLoading && isFullscreen"
      @click="emit('exitFullscreen')"
      class="fullscreen-toggle-btn"
      title="退出全屏 (Esc)"
    >
      <FullscreenExitIcon :size="24" />
    </button>

    <!-- 保存书签按钮 -->
    <button
      @click="handleSaveBookmark"
      class="fullscreen-toggle-btn"
      title="保存书签"
    >
      <BookmarkIcon :size="24" />
    </button>
  </div>
</template>

<script setup lang="ts">
import FullscreenIcon from './icons/FullscreenIcon.vue';
import FullscreenExitIcon from './icons/FullscreenExitIcon.vue';
import BookmarkIcon from './icons/BookmarkIcon.vue';
import { saveBookmark } from '../composables/useReaderProgress';

const props = defineProps<{
  visible: boolean;
  isFullscreen: boolean;
  isLoading: boolean;
  filePath: string;
  rendition: any;
}>();

const emit = defineEmits<{
  (e: 'enterFullscreen'): void;
  (e: 'exitFullscreen'): void;
  (e: 'mouseleave'): void;
  (e: 'bookmarkSaved'): void;
}>();

const handleSaveBookmark = async () => {
  if (!props.rendition) {
    console.error('无法保存书签：rendition 未就绪');
    return;
  }

  const result = await saveBookmark(props.rendition, props.filePath);
  if (result) {
    console.log('书签保存成功:', result);
    emit('bookmarkSaved');
  } else {
    console.error('书签保存失败');
  }
};
</script>

<style scoped>
.reader-function-menu {
  position: absolute;
  top: 0;
  right: 0;
  height: 100%; /* Make sure it covers the full height of the parent */
  width: auto; /* Let width be determined by content */
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 10px; /* Adjust padding as needed, instead of padding-top */
  background: transparent; /* 初始透明 */
  transition: background 0.3s ease, opacity 0.3s ease;
  opacity: 0; /* 默认隐藏 */
  pointer-events: none; /* 默认不响应事件 */
  z-index: 10;
}

.reader-function-menu.visible {
  opacity: 1;
  pointer-events: auto; /* 可见时响应事件 */
}

.fullscreen-toggle-btn {
  width: 44px;
  height: 44px;
  background: rgba(0, 0, 0, 0.05); /* Change to match previous design */
  border: 1px solid rgba(0, 0, 0, 0.08); /* Change to match previous design */
  border-radius: var(--radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  color: var(--text-color);
  transition: all var(--transition-normal);
  margin: 10px 15px;
}

.fullscreen-toggle-btn:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  transform: scale(1.05);
}
</style>