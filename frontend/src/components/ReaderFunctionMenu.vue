<template>
  <div
    class="reader-function-menu"
    :class="{ 'visible': visible || isPinned }"
    @mouseleave="handleMouseLeave"
  >
    <!-- 固定按钮 -->
    <button
      @click.stop="emit('toggle-pin')"
      class="fullscreen-toggle-btn"
      :class="{ 'pinned': isPinned }"
      :title="isPinned ? '取消固定' : '固定显示'"
    >
      <svg v-if="!isPinned" stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 24 24" height="24" width="24" xmlns="http://www.w3.org/2000/svg">
        <path fill="none" d="M0 0h24v24H0z"></path>
        <path d="M14 4v5c0 1.12.37 2.16 1 3H9c.65-.86 1-1.9 1-3V4h4m3-2H7c-.55 0-1 .45-1 1s.45 1 1 1h1v5c0 1.66-1.34 3-3 3v2h5.97v7l1 1 1-1v-7H19v-2c-1.66 0-3-1.34-3-3V4h1c.55 0 1-.45 1-1s-.45-1-1-1z"></path>
      </svg>
      <svg v-else stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 24 24" height="24" width="24" xmlns="http://www.w3.org/2000/svg">
        <path fill="none" d="M0 0h24v24H0z"></path>
        <path fill-rule="evenodd" d="M16 9V4h1c.55 0 1-.45 1-1s-.45-1-1-1H7c-.55 0-1 .45-1 1s.45 1 1 1h1v5c0 1.66-1.34 3-3 3v2h5.97v7l1 1 1-1v-7H19v-2c-1.66 0-3-1.34-3-3z"></path>
      </svg>
    </button>

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

    <!-- 复制按钮 -->
    <button
      @click="emit('copySelected')"
      class="fullscreen-toggle-btn"
      title="复制选中内容"
    >
      <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
      </svg>
    </button>

    <!-- 笔记按钮 -->
    <button
      @click="emit('saveNote')"
      class="fullscreen-toggle-btn"
      title="保存笔记"
    >
      <NoteIcon :size="24" />
    </button>
  </div>
</template>

<script setup lang="ts">
import FullscreenIcon from './icons/FullscreenIcon.vue';
import FullscreenExitIcon from './icons/FullscreenExitIcon.vue';
import BookmarkIcon from './icons/BookmarkIcon.vue';
import NoteIcon from './icons/NoteIcon.vue';
import { saveBookmark } from '../composables/useReaderProgress';
import { eventBus } from '../composables/useEventBus';

const props = defineProps<{
  visible: boolean;
  isFullscreen: boolean;
  isLoading: boolean;
  filePath: string;
  rendition: any;
  isPinned: boolean;
}>();

const emit = defineEmits<{
  (e: 'enterFullscreen'): void;
  (e: 'exitFullscreen'): void;
  (e: 'mouseleave'): void;
  (e: 'bookmarkSaved'): void;
  (e: 'copySelected'): void;
  (e: 'saveNote'): void;
  (e: 'toggle-pin'): void;
  (e: 'switch-sidebar', viewName: string): void;
}>();

const handleMouseLeave = () => {
  if (!props.isPinned) {
    emit('mouseleave');
  }
};

const handleSaveBookmark = async () => {
  if (!props.rendition) {
    console.error('无法保存书签：rendition 未就绪');
    return;
  }

  const result = await saveBookmark(props.rendition, props.filePath);
  if (result) {
    console.log('书签保存成功:', result);
    emit('bookmarkSaved');
    emit('switch-sidebar', 'bookmarks');
    eventBus.emit('sidebar-switch', { view: 'bookmarks', chapterTitle: result.chapterTitle });
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
  transition: opacity 0.3s ease;
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

.fullscreen-toggle-btn.pinned {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
}
</style>