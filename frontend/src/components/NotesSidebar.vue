<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick, watch } from 'vue';
import NoteIcon from './icons/NoteIcon.vue';
import EditIcon from './icons/EditIcon.vue';
import TrashIcon from './icons/TrashIcon.vue';
import CheckIcon from './icons/CheckIcon.vue';
import XIcon from './icons/XIcon.vue';
import SearchIcon from './icons/SearchIcon.vue';
import FullscreenIcon from './icons/FullscreenIcon.vue';
import { getNotes, deleteNote } from '../composables/useReaderProgress';
import { eventBus } from '../composables/useEventBus';

const props = defineProps<{
  filePath?: string;
  rendition?: any;
}>();

const emit = defineEmits<{
  (e: 'jump', cfi: string): void;
  (e: 'edit-note', note: any): void;
}>();

interface Note {
  cfi: string;
  percentage: number;
  timestamp: number;
  chapterTitle?: string;
  content: string;
  selectedText?: string;
  color?: string;
}

interface GroupedNotes {
  chapterTitle: string;
  notes: Note[];
}

const notes = ref<Note[]>([]);
const isLoading = ref(false);
const editingNote = ref<string | null>(null);
const editingContent = ref('');
const editingColor = ref('#FFCDD2');
const isClickingOtherNote = ref(false);
const editableRefs = ref<Map<string, HTMLElement>>(new Map());
const searchQuery = ref('');
const expandedGroups = ref<Set<string>>(new Set());

const colorOptions = [
  '#FFCDD2', '#FFE0B2', '#FFF9C4', '#C8E6C9', 
  '#B3E5FC', '#E1BEE7', '#FFAB91', '#CE93D8'
];

const filteredNotes = computed<Note[]>(() => {
  if (!searchQuery.value.trim()) {
    return notes.value;
  }
  const query = searchQuery.value.toLowerCase();
  return notes.value.filter(note => 
    (note.content && note.content.toLowerCase().includes(query)) ||
    (note.selectedText && note.selectedText.toLowerCase().includes(query)) ||
    (note.chapterTitle && note.chapterTitle.toLowerCase().includes(query))
  );
});

const groupedNotes = computed<GroupedNotes[]>(() => {
  const groups: { [key: string]: Note[] } = {};
  
  filteredNotes.value.forEach(note => {
    const key = note.chapterTitle || '未知章节';
    if (!groups[key]) {
      groups[key] = [];
    }
    groups[key].push(note);
  });
  
  return Object.entries(groups).map(([chapterTitle, notes]) => ({
    chapterTitle,
    notes
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

const loadNotes = async () => {
  if (!props.filePath) {
    console.warn('未提供文件路径，无法加载笔记');
    return;
  }

  isLoading.value = true;
  try {
    const result = await getNotes(props.filePath);
    if (result) {
      notes.value = result;
      console.log('笔记加载成功:', result.length, '个笔记');
    } else {
      notes.value = [];
    }
  } catch (err) {
    console.error('加载笔记失败:', err);
    notes.value = [];
  } finally {
    isLoading.value = false;
  }
};

const handleNoteClick = (note: Note) => {
  if (editingNote.value === note.cfi) return;
  isClickingOtherNote.value = true;
  setTimeout(() => {
    isClickingOtherNote.value = false;
  }, 200);
  emit('jump', note.cfi);
};

const handleDeleteNote = async (note: Note) => {
  if (!props.filePath) return;

  try {
    await deleteNote(props.filePath, note.cfi);
    notes.value = notes.value.filter(n => n.cfi !== note.cfi);
    console.log('笔记已删除');
    // 发送带 CFI 的事件，让阅读器只更新这一个笔记
    eventBus.emit('note-saved', note.cfi);
  } catch (err) {
    console.error('删除笔记失败:', err);
  }
};

const startEdit = (note: Note) => {
  editingNote.value = note.cfi;
  editingContent.value = note.content || '';
  editingColor.value = note.color || '#FFCDD2';
  nextTick(() => {
    const editableDiv = editableRefs.value.get(note.cfi);
    if (editableDiv) {
      editableDiv.innerText = editingContent.value;
      editableDiv.focus();
    }
  });
};

const saveEdit = async (note: Note) => {
  if (!props.filePath) return;
  if (isClickingOtherNote.value) {
    editingNote.value = null;
    editingContent.value = '';
    return;
  }

  const updatedNote: Note = {
    ...note,
    content: editingContent.value,
    color: editingColor.value
  };

  try {
    // @ts-ignore
    await window.go.main.App.SaveNote(props.filePath, JSON.stringify(updatedNote));
    const index = notes.value.findIndex(n => n.cfi === note.cfi);
    if (index !== -1) {
      notes.value[index].content = editingContent.value;
      notes.value[index].color = editingColor.value;
    }
    editingNote.value = null;
    console.log('笔记已更新');
    // 发送带 CFI 的事件，让阅读器只更新这一个笔记
    eventBus.emit('note-saved', note.cfi);
  } catch (err) {
    console.error('更新笔记失败:', err);
  }
};

const cancelEdit = () => {
  editingNote.value = null;
  editingContent.value = '';
};

const handleNoteSaved = () => {
  console.log('NotesSidebar: 收到 note-saved 事件，开始刷新笔记');
  loadNotes();
};

const isFirstLoad = ref(true);

onMounted(() => {
  console.log('NotesSidebar: 组件挂载，filePath:', props.filePath);
  loadNotes();
  eventBus.on('note-saved', handleNoteSaved);
});

watch(notes, (newNotes) => {
  if (isFirstLoad.value && newNotes.length > 0) {
    // 首次加载到数据时展开所有分组
    const groups = new Set<string>();
    newNotes.forEach(note => {
      groups.add(note.chapterTitle || '未知章节');
    });
    expandedGroups.value = groups;
    isFirstLoad.value = false;
  } else if (!isFirstLoad.value) {
    // 后续更新时，只保留当前已有的分组，保持展开状态不变
    const currentGroups = new Set<string>();
    newNotes.forEach(note => {
      currentGroups.add(note.chapterTitle || '未知章节');
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
  eventBus.off('note-saved', handleNoteSaved);
});

defineExpose({
  refresh: loadNotes
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
  <div class="notes-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><NoteIcon :size="22" style="margin-right: 8px;" />笔记</h2>
    </div>
    
    <div class="search-container" v-if="notes.length > 0">
      <div class="search-input-wrapper">
        <SearchIcon :size="16" class="search-input-icon" />
        <input 
          v-model="searchQuery" 
          type="text" 
          placeholder="搜索笔记..." 
          class="search-input"
        />
        <button v-if="searchQuery" class="search-clear-btn" @click="searchQuery = ''">
          <XIcon :size="14" />
        </button>
      </div>
    </div>

    <div class="notes-content">
      <div v-if="isLoading" class="empty-state">
        <NoteIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>加载中...</span>
      </div>

      <div v-else-if="notes.length === 0" class="empty-state">
        <NoteIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>暂无笔记</span>
        <p style="font-size: 0.8rem; color: var(--text-muted); margin-top: 8px;">在阅读时选中文本添加笔记</p>
      </div>
      
      <div v-else-if="filteredNotes.length === 0" class="empty-state">
        <NoteIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>未找到匹配的笔记</span>
      </div>

      <div v-else class="notes-groups">
        <div
          v-for="group in groupedNotes"
          :key="group.chapterTitle"
          class="note-group"
        >
          <div class="group-header" @click="toggleGroup(group.chapterTitle)">
            <div class="group-header-left">
              <span class="expand-icon" :class="{ 'expanded': isGroupExpanded(group.chapterTitle) }">
                ▶
              </span>
              <span class="group-title">{{ group.chapterTitle }}</span>
            </div>
            <span class="group-count">{{ group.notes.length }}</span>
          </div>
          
          <div class="group-notes" v-show="isGroupExpanded(group.chapterTitle)">
            <div
              v-for="note in group.notes"
              :key="note.cfi"
              class="note-card"
              :class="{ 'editing': editingNote === note.cfi }"
              @click="handleNoteClick(note)"
            >
              <div class="note-selected-text">
                <mark 
                  :style="{ 
                    backgroundColor: note.color || '#FFCDD2', 
                    color: '#000',
                    opacity: 0.8
                  }"
                >
                  {{ note.selectedText || '无选中内容' }}
                </mark>
              </div>
              
              <div 
                v-if="editingNote !== note.cfi"
                class="note-content"
                @click="(e) => { if (editingNote === note.cfi) e.stopPropagation() }"
                v-html="(note.content || '点击编辑笔记').replace(/\n/g, '<br>')"
              >
              </div>
              <div 
                v-else
                class="note-content editing"
                contenteditable="true"
                :ref="(el) => { if (el) editableRefs.set(note.cfi, el as HTMLElement) }"
                @input="(e) => { editingContent = (e.target as HTMLElement).innerText }"
                @keydown.enter.ctrl="saveEdit(note)"
                @keydown.escape="cancelEdit"
                @click.stop
              >
              </div>
              
              <div class="note-footer">
                <span v-if="editingNote !== note.cfi" class="note-date">{{ formatDate(note.timestamp) }}</span>
                <span v-else class="note-date-placeholder"></span>
                <div class="note-actions" v-if="editingNote !== note.cfi">
                  <button class="action-btn sidebar-edit-btn" @click.stop="emit('edit-note', note)" title="在侧边栏编辑">
                    <FullscreenIcon :size="14" />
                  </button>
                  <button class="action-btn edit-btn" @click.stop="startEdit(note)" title="编辑">
                    <EditIcon :size="14" />
                  </button>
                  <button class="action-btn delete-btn" @click.stop="handleDeleteNote(note)" title="删除">
                    <TrashIcon :size="14" />
                  </button>
                </div>
                <div class="note-actions editing-actions" v-else>
                  <div class="color-picker">
                    <button 
                      v-for="color in colorOptions" 
                      :key="color"
                      class="color-btn"
                      :style="{ backgroundColor: color }"
                      :class="{ active: editingColor === color }"
                      @click.stop="editingColor = color"
                      :title="color"
                    />
                  </div>
                  <button class="action-btn confirm-btn" @click.stop="saveEdit(note)" title="保存">
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
.notes-wrapper {
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

.search-container {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}

.search-input-wrapper {
  position: relative;
}

.search-input-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.search-input {
  width: 100%;
  padding: 10px 36px 10px 32px;
  border: none;
  border-radius: var(--radius-md);
  background: rgba(var(--primary-rgb), 0.06);
  color: var(--text-primary);
  font-size: 0.85rem;
  transition: all var(--transition-fast);
  box-sizing: border-box;
}

.search-input:hover {
  background: rgba(var(--primary-rgb), 0.1);
}

.search-input:focus {
  background: rgba(var(--primary-rgb), 0.12);
  box-shadow: 0 0 0 2px rgba(var(--primary-rgb), 0.2);
  outline: none;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-clear-btn {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.search-clear-btn:hover {
  color: var(--text-secondary);
  background: var(--border-color);
}

.notes-content {
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

.notes-groups {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.note-group {
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

.group-notes {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.note-card {
  background: var(--bg-color);
  border-radius: var(--radius-md);
  padding: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}

.note-card:hover {
  background: rgba(0, 0, 0, 0.03);
  border-color: var(--border-color);
}

.note-card.editing {
  background: rgba(0, 0, 0, 0.05);
  border-color: var(--primary-color);
}

.note-selected-text {
  font-size: 0.8rem;
  line-height: 1.4;
  word-break: break-word;
  margin-bottom: 8px;
  text-align: justify;
}

.note-selected-text mark {
  padding: 0 2px;
  border-radius: 2px;
  opacity: 0.8;
}

.note-content {
  font-size: 0.9rem;
  color: var(--text-secondary);
  line-height: 1.5;
  word-break: break-word;
  cursor: pointer;
  outline: none;
  min-height: 24px;
}

.note-content.editing {
  cursor: text;
  outline: none;
}

.note-content.editing:focus {
  outline: none;
}

.note-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}

.note-date {
  font-size: 0.7rem;
  color: var(--text-muted);
  white-space: nowrap;
}

.note-date-placeholder {
  font-size: 0.7rem;
  flex: 1;
}

.note-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.note-card:hover .note-actions {
  opacity: 1;
}

.note-card.editing .note-actions {
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

.sidebar-edit-btn:hover {
  background: rgba(var(--primary-rgb), 0.1);
  color: var(--primary-color);
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

.color-picker {
  display: flex;
  gap: 4px;
  margin-right: 4px;
}

.color-btn {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.color-btn:hover {
  transform: scale(1.1);
}

.color-btn.active {
  border-color: var(--text-primary);
  box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.1);
}

.notes-content::-webkit-scrollbar {
  width: 4px;
}

.notes-content::-webkit-scrollbar-track {
  background: transparent;
}

.notes-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.notes-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>