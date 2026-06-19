<template>
  <aside 
    class="sidebar-container" 
    :class="{ 'is-collapsed': isCollapsed }"
    :style="!isCollapsed ? { width: settingsStore.sidebarWidth + 'px' } : {}"
  >
    <component 
      :is="currentComponent" 
      v-bind="componentProps" 
      @switch-view="switchView"
      @jump="handleJump"
      @open-shelf="handleOpenShelf"
      @add-theme="handleAddTheme"
      @edit-theme="handleEditTheme"
      @sync-complete="handleSyncComplete"
      @show-toast="handleShowToast"
    />
    <!-- 拖拽调整柄 -->
    <div 
      class="resize-handle" 
      @mousedown="startResize"
      title="拖动调整宽度"
    />
  </aside>
</template>

<script setup lang="ts">
import { ref, shallowRef } from 'vue';
import ShelfSidebar from './ShelfSidebar.vue';
import TocSidebar from './TocSidebar.vue';
import LayoutSidebar from './LayoutSidebar.vue';
import ThemeSidebar from './ThemeSidebar.vue';
import WebDavSidebar from './WebDavSidebar.vue';
import { useSettingsStore } from '../stores/settings';

const settingsStore = useSettingsStore();

// 使用 any 类型解决组件事件类型不兼容问题
const currentComponent = shallowRef<any>(ShelfSidebar);
const componentProps = ref({});
const isCollapsed = ref(false);

// 拖拽调整宽度
const MIN_WIDTH = 240;
const MAX_WIDTH = 300;

const startResize = (e: MouseEvent) => {
  e.preventDefault();
  const startX = e.clientX;
  const startWidth = settingsStore.sidebarWidth;

  document.body.classList.add('resizing-sidebar');

  const onMouseMove = (e: MouseEvent) => {
    const delta = e.clientX - startX;
    const newWidth = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, startWidth + delta));
    settingsStore.sidebarWidth = newWidth;
  };

  const onMouseUp = () => {
    document.body.classList.remove('resizing-sidebar');
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp);
  };

  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', onMouseUp);
};

// 切换显示的逻辑
const switchView = (viewName: string) => {
  if (viewName === 'shelf') currentComponent.value = ShelfSidebar;
  if (viewName === 'toc') currentComponent.value = TocSidebar;
  if (viewName === 'layout') currentComponent.value = LayoutSidebar;
  if (viewName === 'theme') currentComponent.value = ThemeSidebar;
  if (viewName === 'webdav') currentComponent.value = WebDavSidebar;
};

// 处理目录跳转，传递给父组件
const handleJump = (href: string) => {
  emit('jump', href);
};

// 处理打开书架
const handleOpenShelf = (shelfId: string, shelfName: string) => {
  emit('open-shelf', shelfId, shelfName);
};

// 处理添加主题
const handleAddTheme = () => {
  emit('add-theme');
};

// 处理编辑主题
const handleEditTheme = (themeId: string) => {
  emit('edit-theme', themeId);
};

// 处理同步完成
const handleSyncComplete = () => {
  emit('sync-complete');
};

// 处理显示提示气泡
const handleShowToast = (message: string, type: 'success' | 'error') => {
  emit('show-toast', message, type);
};

// 暴露方法供父组件调用
defineExpose({
  switchView
});

const emit = defineEmits<{
  (e: 'jump', href: string): void
  (e: 'open-shelf', shelfId: string, shelfName: string): void
  (e: 'add-theme'): void
  (e: 'edit-theme', themeId: string): void
  (e: 'sync-complete'): void
  (e: 'show-toast', message: string, type: 'success' | 'error'): void
}>();
</script>

<style scoped>
.sidebar-container {
  position: relative;
  width: 240px;
  height: 100%;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--border-color);
  flex-shrink: 0;
  box-shadow: 2px 0 20px rgba(0, 0, 0, 0.04);
}

.sidebar-container.is-collapsed {
  width: 60px;
}

/* 拖拽调整柄 */
.resize-handle {
  position: absolute;
  top: 0;
  right: -3px;
  width: 6px;
  height: 100%;
  cursor: col-resize;
  z-index: 10;
  background: transparent;
  transition: background 0.2s;
}

.resize-handle:hover {
  background: var(--primary-color);
  opacity: 0.4;
}
</style>