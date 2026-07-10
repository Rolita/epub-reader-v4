<template>
  <aside 
    class="sidebar-container" 
    :class="{ 'is-collapsed': isCollapsed }"
    :style="!isCollapsed ? { width: settingsStore.sidebarWidth + 'px' } : {}"
  >
    <keep-alive>
      <component 
        ref="currentComponentRef"
        :is="currentComponent" 
        v-bind="componentProps" 
        :key="componentKey"
        @switch-view="switchView"
      @jump="handleJump"
      @preview="handlePreview"
      @open-shelf="handleOpenShelf"
      @add-theme="handleAddTheme"
      @edit-theme="handleEditTheme"
      @sync-complete="handleSyncComplete"
      @show-toast="handleShowToast"
      @edit-note="handleEditNote"
      />
    </keep-alive>
    <!-- 拖拽调整柄 -->
    <div 
      class="resize-handle" 
      @mousedown="startResize"
      title="拖动调整宽度"
    />
  </aside>
</template>

<script setup lang="ts">
import { ref, shallowRef, watch } from 'vue';
import ShelfSidebar from './ShelfSidebar.vue';
import TocSidebar from './TocSidebar.vue';
import LayoutSidebar from './LayoutSidebar.vue';
import ThemeSidebar from './ThemeSidebar.vue';
import WebDavSidebar from './WebDavSidebar.vue';
import BookshelfLayoutSidebar from './BookshelfLayoutSidebar.vue';
import IllustrationSidebar from './IllustrationSidebar.vue';
import BookmarksSidebar from './BookmarksSidebar.vue';
import TranslateSidebar from './TranslateSidebar.vue';
import SearchSidebar from './SearchSidebar.vue';
import NotesSidebar from './NotesSidebar.vue';
import NoteEditSidebar from './NoteEditSidebar.vue';
import { useSettingsStore } from '../stores/settings';

const settingsStore = useSettingsStore();

const props = defineProps<{
  hasActiveBook?: boolean;
  bookTitle?: string;
  filePath?: string;
  searchInBook?: (keyword: string) => Promise<Array<{ chapter: string; snippet: string; href: string; cfi: string; page: number }>>;
  highlightSearchKeyword?: (keyword: string) => void;
  clearSearchHighlight?: () => void;
}>();

// 使用 any 类型解决组件事件类型不兼容问题
const currentComponent = shallowRef<any>(ShelfSidebar);
const currentComponentRef = ref<any>(null);
const componentProps = ref({});
const isCollapsed = ref(false);
const componentKey = ref('shelf');
const editingNoteData = ref<any>(null);

watch(() => [props.hasActiveBook, props.bookTitle, props.searchInBook, props.highlightSearchKeyword, props.clearSearchHighlight], () => {
  if (currentComponent.value === SearchSidebar) {
    componentProps.value = {
      hasActiveBook: props.hasActiveBook,
      bookTitle: props.bookTitle,
      searchInBook: props.searchInBook,
      highlightSearchKeyword: props.highlightSearchKeyword,
      clearSearchHighlight: props.clearSearchHighlight
    };
  }
});

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
  if (viewName === 'shelf') {
    currentComponent.value = ShelfSidebar;
    componentKey.value = 'shelf';
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'toc') {
    currentComponent.value = TocSidebar;
    componentKey.value = 'toc';
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'layout') {
    currentComponent.value = LayoutSidebar;
    componentKey.value = 'layout';
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'theme') {
    currentComponent.value = ThemeSidebar;
    componentKey.value = 'theme';
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'webdav') {
    currentComponent.value = WebDavSidebar;
    componentKey.value = 'webdav';
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'bookshelf-layout') {
    currentComponent.value = BookshelfLayoutSidebar;
    componentKey.value = 'bookshelf-layout';
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'illustration') {
    currentComponent.value = IllustrationSidebar;
    componentKey.value = 'illustration';
    settingsStore.showIllustrationSidebar = true;
  }
  if (viewName === 'bookmarks') {
    currentComponent.value = BookmarksSidebar;
    componentKey.value = `bookmarks-${props.filePath}`;
    componentProps.value = {
      filePath: props.filePath
    };
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'translate') {
    currentComponent.value = TranslateSidebar;
    componentKey.value = 'translate';
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'search') {
    currentComponent.value = SearchSidebar;
    componentKey.value = `search-${props.filePath}`;
    componentProps.value = {
      hasActiveBook: props.hasActiveBook,
      bookTitle: props.bookTitle,
      filePath: props.filePath,
      searchInBook: props.searchInBook,
      highlightSearchKeyword: props.highlightSearchKeyword,
      clearSearchHighlight: props.clearSearchHighlight
    };
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'notes') {
    currentComponent.value = NotesSidebar;
    componentKey.value = `notes-${props.filePath}`;
    componentProps.value = {
      filePath: props.filePath
    };
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'note-edit') {
    currentComponent.value = NoteEditSidebar;
    componentKey.value = `note-edit-${editingNoteData.value?.cfi || ''}`;
    componentProps.value = {
      filePath: props.filePath,
      note: editingNoteData.value
    };
    settingsStore.showIllustrationSidebar = false;
  }
  if (viewName === 'none') {
    settingsStore.showIllustrationSidebar = false;
  }
};

// 处理目录跳转，传递给父组件
const handleJump = (payload: any) => {
  emit('jump', payload);
};

// 处理图片预览，传递给父组件
const handlePreview = (payload: { src: string; alt: string }) => {
  emit('preview', payload);
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

// 处理在侧边栏中编辑笔记
const handleEditNote = (note: any) => {
  editingNoteData.value = note;
  switchView('note-edit');
};

const refreshBookmarks = () => {
  if (currentComponent.value === BookmarksSidebar && currentComponentRef.value) {
    currentComponentRef.value.refresh();
  }
};

const refreshNotes = () => {
  if (currentComponent.value === NotesSidebar && currentComponentRef.value) {
    currentComponentRef.value.refresh();
  }
};

// 暴露方法供父组件调用
defineExpose({
  switchView,
  refreshBookmarks,
  refreshNotes
});

const emit = defineEmits<{
  (e: 'jump', payload: any): void
  (e: 'preview', payload: { src: string; alt: string }): void
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