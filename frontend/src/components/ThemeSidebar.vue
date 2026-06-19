<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useThemeStore } from '../stores/theme';
import SparklesIcon from './icons/SparklesIcon.vue';
import BookOpenIcon from './icons/BookOpenIcon.vue';
import CloudIcon from './icons/CloudIcon.vue';
import MoonIcon from './icons/MoonIcon.vue';
import SunIcon from './icons/SunIcon.vue';
import StarIcon from './icons/StarIcon.vue';
import PlusIcon from './icons/PlusIcon.vue';

const themeStore = useThemeStore();

const emit = defineEmits<{
  (e: 'add-theme'): void
  (e: 'edit-theme', themeId: string): void
}>();

const iconComponents: Record<string, any> = {
  'book-open': BookOpenIcon,
  'cloud': CloudIcon,
  'moon': MoonIcon,
  'sun': SunIcon,
  'star': StarIcon
};

// 拖拽状态
const draggedThemeId = ref<string | null>(null);
const dragOverThemeId = ref<string | null>(null);

// 右键菜单状态
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  themeId: ''
});

const handleContextMenu = (event: MouseEvent, themeId: string) => {
  event.preventDefault();
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    themeId
  };
};

const closeContextMenu = () => {
  contextMenu.value.visible = false;
};

const handleClickOutside = () => {
  closeContextMenu();
};

const handleEditTheme = () => {
  emit('edit-theme', contextMenu.value.themeId);
  closeContextMenu();
};

const handleDeleteTheme = () => {
  const themeId = contextMenu.value.themeId;
  if (themeStore.isDefaultTheme(themeId)) return;
  themeStore.deleteTheme(themeId);
  closeContextMenu();
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});

// 拖拽开始
const handleDragStart = (event: DragEvent, themeId: string) => {
  draggedThemeId.value = themeId;
  if (event.dataTransfer) {
    event.dataTransfer.setData('text/plain', themeId);
    event.dataTransfer.effectAllowed = 'move';
  }
};

// 拖拽进入
const handleDragOver = (event: DragEvent, themeId: string) => {
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move';
  }
  dragOverThemeId.value = themeId;
};

// 拖拽离开
const handleDragLeave = () => {
  dragOverThemeId.value = null;
};

// 放置
const handleDrop = (event: DragEvent, targetThemeId: string) => {
  event.preventDefault();
  if (!draggedThemeId.value || draggedThemeId.value === targetThemeId) {
    draggedThemeId.value = null;
    dragOverThemeId.value = null;
    return;
  }

  // 重新排序主题列表
  const themes = [...themeStore.themes];
  const draggedIndex = themes.findIndex(t => t.id === draggedThemeId.value);
  const targetIndex = themes.findIndex(t => t.id === targetThemeId);

  if (draggedIndex !== -1 && targetIndex !== -1) {
    const [draggedTheme] = themes.splice(draggedIndex, 1);
    themes.splice(targetIndex, 0, draggedTheme);
    
    // 更新主题列表顺序
    themeStore.reorderThemes(themes);
  }

  draggedThemeId.value = null;
  dragOverThemeId.value = null;
};

// 拖拽结束
const handleDragEnd = () => {
  draggedThemeId.value = null;
  dragOverThemeId.value = null;
};
</script>

<template>
  <div class="theme-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><SparklesIcon :size="22" style="margin-right: 8px; vertical-align: middle;" />主题设置</h2>
    </div>
    
    <div class="theme-content">
      <div class="theme-list">
        <button
          v-for="theme in themeStore.themes"
          :key="theme.id"
          :class="[
            'theme-item', 
            { 
              active: themeStore.currentTheme === theme.id,
              dragging: draggedThemeId === theme.id,
              'drag-over': dragOverThemeId === theme.id && draggedThemeId !== theme.id
            }
          ]"
          draggable="true"
          @dragstart="handleDragStart($event, theme.id)"
          @dragover="handleDragOver($event, theme.id)"
          @dragleave="handleDragLeave"
          @drop="handleDrop($event, theme.id)"
          @dragend="handleDragEnd"
          @click="themeStore.setTheme(theme.id)"
          @contextmenu="handleContextMenu($event, theme.id)"
        >
          <div class="theme-icon-wrapper">
            <component :is="iconComponents[theme.icon]" :size="22" class="theme-icon" />
          </div>
          <span class="theme-name">{{ theme.name }}</span>
          <div 
            class="theme-indicator" 
            :style="{ background: `linear-gradient(135deg, ${theme.primary} 0%, ${theme.accent} 100%)` }"
          ></div>
        </button>
        
        <!-- 添加主题按钮 -->
        <button class="theme-item add-theme-btn" @click="emit('add-theme')">
          <div class="theme-icon-wrapper">
            <PlusIcon :size="22" class="theme-icon" />
          </div>
          <span class="theme-name">添加主题</span>
        </button>
      </div>
    </div>

    <!-- 右键菜单 -->
    <div 
      v-if="contextMenu.visible" 
      class="context-menu" 
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      @click.stop
    >
      <button class="context-menu-item" @click="handleEditTheme">
        编辑主题
      </button>
      <button 
        v-if="!themeStore.isDefaultTheme(contextMenu.themeId)" 
        class="context-menu-item danger" 
        @click="handleDeleteTheme"
      >
        删除主题
      </button>
    </div>
  </div>
</template>

<style scoped>
/* 主题容器 */
.theme-wrapper {
  width: 100%;
  height: 100%;
  background: var(--sidebar-bg);
  color: var(--text-color);
  display: flex;
  flex-direction: column;
  user-select: none;
}

/* 侧边栏头部 */
.sidebar-header {
  padding: 28px 20px;
  border-bottom: 1px solid var(--border-color);
  background: linear-gradient(180deg, var(--primary-light) 0%, transparent 100%);
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  display: flex;
  align-items: center;
}

/* 主题内容区 */
.theme-content {
  flex: 1;
  padding: 20px 18px;
  overflow-y: auto;
}

/* 主题列表 */
.theme-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 主题项 */
.theme-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border: 2px solid transparent;
  border-radius: var(--radius-lg);
  background: var(--bg-color);
  transition: all var(--transition-normal);
  box-shadow: var(--shadow-sm);
  text-align: left;
}

.theme-item:hover {
  background: var(--primary-light);
  transform: translateX(4px);
}

.theme-item.active {
  border-color: var(--primary-color);
  background: var(--primary-light);
}

.theme-item.dragging {
  opacity: 0.5;
  transform: scale(0.98);
}

.theme-item.drag-over {
  background: var(--primary-light);
  border: 2px dashed var(--primary-color);
}

/* 主题图标容器 */
.theme-icon-wrapper {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  background: rgba(0, 0, 0, 0.05);
}

.theme-icon {
  color: var(--text-primary);
}

/* 主题名称 */
.theme-name {
  flex: 1;
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 500;
  letter-spacing: -0.01em;
}

/* 主题指示器 */
.theme-indicator {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-md);
  border: 1.5px solid var(--border-color);
  box-shadow: var(--shadow-sm);
}

/* 添加主题按钮 */
.add-theme-btn {
  border: 2px dashed var(--border-color);
  background: transparent;
}

.add-theme-btn:hover {
  border-color: var(--primary-color);
  background: var(--primary-light);
}

.add-theme-btn .theme-icon-wrapper {
  background: var(--primary-light);
}

/* 滚动条美化 */
.theme-content::-webkit-scrollbar {
  width: 4px;
}

.theme-content::-webkit-scrollbar-track {
  background: transparent;
}

.theme-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.theme-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

/* 右键菜单 */
.context-menu {
  position: fixed;
  z-index: 1000;
  min-width: 140px;
  padding: 4px 0;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.context-menu-item {
  display: block;
  width: 100%;
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 0.85rem;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s;
}

.context-menu-item:hover {
  background: var(--primary-light);
  color: var(--primary-color);
}

.context-menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger-color, #EF4444);
}
</style>