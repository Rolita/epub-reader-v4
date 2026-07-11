<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick, watch } from 'vue';
import BookmarkIcon from './icons/BookmarkIcon.vue';
import EditIcon from './icons/EditIcon.vue';
import TrashIcon from './icons/TrashIcon.vue';
import CheckIcon from './icons/CheckIcon.vue';
import XIcon from './icons/XIcon.vue';
import { getBookmarks, deleteBookmark } from '../composables/useReaderProgress';
import { eventBus } from '../composables/useEventBus';

const props = defineProps<{
  filePath?: string;
  rendition?: any;
}>();

const emit = defineEmits<{
  (e: 'jump', cfi: string): void;
}>();

interface Bookmark {
  cfi: string;
  percentage: number;
  timestamp: number;
  chapterTitle?: string;
  snippet?: string;
}

interface GroupedBookmarks {
  chapterTitle: string;
  bookmarks: Bookmark[];
}

const bookmarks = ref<Bookmark[]>([]);
const isLoading = ref(false);
const editingBookmark = ref<string | null>(null);
const editingSnippet = ref('');
const isClickingOtherBookmark = ref(false);
const editableRefs = ref<Map<string, HTMLElement>>(new Map());
const expandedGroups = ref<Set<string>>(new Set());

const groupedBookmarks = computed<GroupedBookmarks[]>(() => {
  const groups: { [key: string]: Bookmark[] } = {};
  
  bookmarks.value.forEach(bookmark => {
    const key = bookmark.chapterTitle || '未知章节';
    if (!groups[key]) {
      groups[key] = [];
    }
    groups[key].push(bookmark);
  });
  
  return Object.entries(groups).map(([chapterTitle, bookmarks]) => ({
    chapterTitle,
    bookmarks
  }));
});

const toggleGroup = (chapterTitle: string) => {
  if (expandedGroups.value.has(chapterTitle)) {
    expandedGroups.value.delete(chapterTitle);
  } else {
    expandedGroups.value.add(chapterTitle);
  }
};

const isGroupExpanded = (chapterTitle: string) => {
  return expandedGroups.value.has(chapterTitle);
};

const loadBookmarks = async () => {
  if (!props.filePath) {
    console.warn('未提供文件路径，无法加载书签');
    return;
  }

  isLoading.value = true;
  try {
    const result = await getBookmarks(props.filePath);
    if (result) {
      bookmarks.value = result;
      console.log('书签加载成功:', result.length, '个书签');
    } else {
      bookmarks.value = [];
    }
  } catch (err) {
    console.error('加载书签失败:', err);
    bookmarks.value = [];
  } finally {
    isLoading.value = false;
  }
};

const handleBookmarkClick = (bookmark: Bookmark) => {
  if (editingBookmark.value === bookmark.cfi) return;
  isClickingOtherBookmark.value = true;
  setTimeout(() => {
    isClickingOtherBookmark.value = false;
  }, 200);
  emit('jump', bookmark.cfi);
};

const handleDeleteBookmark = async (bookmark: Bookmark) => {
  if (!props.filePath) return;

  try {
    await deleteBookmark(props.filePath, bookmark.cfi);
    bookmarks.value = bookmarks.value.filter(b => b.cfi !== bookmark.cfi);
    console.log('书签已删除');
  } catch (err) {
    console.error('删除书签失败:', err);
  }
};

const startEdit = (bookmark: Bookmark) => {
  editingBookmark.value = bookmark.cfi;
  editingSnippet.value = bookmark.snippet || '';
  nextTick(() => {
    const editableDiv = editableRefs.value.get(bookmark.cfi);
    if (editableDiv) {
      editableDiv.innerText = editingSnippet.value;
      editableDiv.focus();
    }
  });
};

const saveEdit = async (bookmark: Bookmark) => {
  if (!props.filePath) return;
  if (isClickingOtherBookmark.value) {
    editingBookmark.value = null;
    editingSnippet.value = '';
    return;
  }

  const updatedBookmark: Bookmark = {
    ...bookmark,
    snippet: editingSnippet.value
  };

  try {
    // @ts-ignore
    await window.go.main.App.SaveBookmark(props.filePath, JSON.stringify(updatedBookmark));
    const index = bookmarks.value.findIndex(b => b.cfi === bookmark.cfi);
    if (index !== -1) {
      bookmarks.value[index].snippet = editingSnippet.value;
    }
    editingBookmark.value = null;
    console.log('书签已更新');
  } catch (err) {
    console.error('更新书签失败:', err);
  }
};

const cancelEdit = () => {
  editingBookmark.value = null;
  editingSnippet.value = '';
};

const handleBookmarkSaved = () => {
  loadBookmarks();
};

const isFirstLoad = ref(true);
const forceExpandChapter = ref<string | null>(null);

const handleSidebarSwitch = (data: any) => {
  if (data && data.view === 'bookmarks' && data.chapterTitle) {
    forceExpandChapter.value = data.chapterTitle;
  }
};

onMounted(() => {
  loadBookmarks();
  eventBus.on('bookmark-saved', handleBookmarkSaved);
  eventBus.on('sidebar-switch', handleSidebarSwitch);
});

watch(bookmarks, (newBookmarks) => {
  if (forceExpandChapter.value) {
    // 强制展开指定章节，其他章节收起
    const targetChapter = forceExpandChapter.value;
    const groups = new Set<string>();
    groups.add(targetChapter);
    expandedGroups.value = groups;
    forceExpandChapter.value = null;
    isFirstLoad.value = false;
  } else if (isFirstLoad.value && newBookmarks.length > 0) {
    // 首次加载到数据时展开所有分组
    const groups = new Set<string>();
    newBookmarks.forEach(bookmark => {
      groups.add(bookmark.chapterTitle || '未知章节');
    });
    expandedGroups.value = groups;
    isFirstLoad.value = false;
  } else if (!isFirstLoad.value) {
    // 后续更新时，只保留当前已有的分组，保持展开状态不变
    const currentGroups = new Set<string>();
    newBookmarks.forEach(bookmark => {
      currentGroups.add(bookmark.chapterTitle || '未知章节');
    });
    
    // 移除不存在的分组，保留已有的分组展开状态
    const newExpanded = new Set<string>();
    expandedGroups.value.forEach(group => {
      if (currentGroups.has(group)) {
        newExpanded.add(group);
      }
    });
    expandedGroups.value = newExpanded;
  }
}, { immediate: true });

onUnmounted(() => {
  eventBus.off('bookmark-saved', handleBookmarkSaved);
  eventBus.off('sidebar-switch', handleSidebarSwitch);
});

defineExpose({
  refresh: loadBookmarks
});

function formatDate(timestamp: number): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diff = now.getTime() - date.getTime();

  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 7) return `${days}天前`;

  return `${date.getMonth() + 1}月${date.getDate()}日`;
}
</script>

<template>
  <div class="bookmarks-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><BookmarkIcon :size="22" style="margin-right: 8px;" />书签</h2>
    </div>

    <div class="bookmarks-content">
      <div v-if="isLoading" class="empty-state">
        <BookmarkIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>加载中...</span>
      </div>

      <div v-else-if="bookmarks.length === 0" class="empty-state">
        <BookmarkIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>暂无书签</span>
        <p style="font-size: 0.8rem; color: var(--text-muted); margin-top: 8px;">在阅读时添加书签</p>
      </div>

      <div v-else class="bookmarks-groups">
        <div
          v-for="group in groupedBookmarks"
          :key="group.chapterTitle"
          class="bookmark-group"
        >
          <div class="group-header" @click="toggleGroup(group.chapterTitle)">
            <div class="group-header-left">
              <span class="expand-icon" :class="{ 'expanded': isGroupExpanded(group.chapterTitle) }">
                ▶
              </span>
              <span class="group-title">{{ group.chapterTitle }}</span>
            </div>
            <span class="group-count">{{ group.bookmarks.length }}</span>
          </div>
          
          <div class="group-bookmarks" v-show="isGroupExpanded(group.chapterTitle)">
            <div
              v-for="bookmark in group.bookmarks"
              :key="bookmark.cfi"
              class="bookmark-card"
              :class="{ 'editing': editingBookmark === bookmark.cfi }"
              @click="handleBookmarkClick(bookmark)"
            >
              <div 
                v-if="editingBookmark !== bookmark.cfi"
                class="bookmark-snippet"
                @click="(e) => { if (editingBookmark === bookmark.cfi) e.stopPropagation() }"
                v-html="(bookmark.snippet || '无预览内容').replace(/\n/g, '<br>')"
              >
              </div>
              <div 
                v-else
                class="bookmark-snippet editing"
                contenteditable="true"
                :ref="(el) => { if (el) editableRefs.set(bookmark.cfi, el as HTMLElement) }"
                @input="(e) => { editingSnippet = (e.target as HTMLElement).innerText }"
                @keydown.enter.ctrl="saveEdit(bookmark)"
                @keydown.escape="cancelEdit"
                @click.stop
              >
              </div>
              
              <div class="bookmark-footer">
                <span class="bookmark-date">{{ formatDate(bookmark.timestamp) }}</span>
                <div class="bookmark-actions" v-if="editingBookmark !== bookmark.cfi">
                  <button class="action-btn edit-btn" @click.stop="startEdit(bookmark)" title="编辑">
                    <EditIcon :size="14" />
                  </button>
                  <button class="action-btn delete-btn" @click.stop="handleDeleteBookmark(bookmark)" title="删除">
                    <TrashIcon :size="14" />
                  </button>
                </div>
                <div class="bookmark-actions editing-actions" v-else>
                  <button class="action-btn confirm-btn" @click.stop="saveEdit(bookmark)" title="保存">
                    <CheckIcon :size="14" />
                  </button>
                  <button class="action-btn cancel-btn" @click.stop="cancelEdit" title="取消">
                    <XIcon :size="14" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bookmarks-wrapper {
  width: 100%;
  height: 100%;
  background: var(--sidebar-bg);
  color: var(--text-color);
  display: flex;
  flex-direction: column;
  user-select: none;
}

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

.bookmarks-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

.bookmarks-groups {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.bookmark-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  cursor: pointer;
}

.group-header-left {
  display: flex;
  align-items: center;
  gap: 6px;
}

.expand-icon {
  font-size: 0.6rem;
  color: var(--text-muted);
  transition: transform 0.2s ease;
  transform: rotate(0deg);
}

.expand-icon.expanded {
  transform: rotate(90deg);
}

.group-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.group-count {
  font-size: 0.7rem;
  color: var(--text-muted);
  background: var(--border-color);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
}

.group-bookmarks {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bookmark-card {
  background: var(--bg-color);
  border-radius: var(--radius-md);
  padding: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}

.bookmark-card:hover {
  background: rgba(0, 0, 0, 0.03);
  border-color: var(--border-color);
}

.bookmark-card.editing {
  background: rgba(0, 0, 0, 0.05);
  border-color: var(--primary-color);
}



.bookmark-snippet {
  font-size: 0.8rem;
  color: var(--text-secondary);
  line-height: 1.5;
  word-break: break-word;
  cursor: pointer;
  outline: none;
}

.bookmark-snippet.editing {
  cursor: text;
  outline: none;
}

.bookmark-snippet.editing:focus {
  outline: none;
}

.bookmark-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}

.bookmark-date {
  font-size: 0.7rem;
  color: var(--text-muted);
}

.bookmark-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.bookmark-card:hover .bookmark-actions {
  opacity: 1;
}

.bookmark-card.editing .bookmark-actions {
  opacity: 1;
}

.action-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: var(--text-muted);
  transition: all var(--transition-fast);
}

.action-btn:hover {
  background: var(--border-color);
  color: var(--text-secondary);
}

.delete-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  color: var(--error-color);
}

.confirm-btn {
  color: var(--success-color);
}

.confirm-btn:hover {
  background: rgba(34, 197, 94, 0.1);
}

.cancel-btn {
  color: var(--text-secondary);
}

.cancel-btn:hover {
  background: rgba(0, 0, 0, 0.1);
}

.bookmarks-content::-webkit-scrollbar {
  width: 4px;
}

.bookmarks-content::-webkit-scrollbar-track {
  background: transparent;
}

.bookmarks-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.bookmarks-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>
