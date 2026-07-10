<script setup lang="ts">
import { ref, watch } from 'vue';
import CheckIcon from './icons/CheckIcon.vue';
import XIcon from './icons/XIcon.vue';
import { eventBus } from '../composables/useEventBus';

const props = defineProps<{
  filePath?: string;
  note?: any;
}>();

const emit = defineEmits<{
  (e: 'switch-view', view: string): void;
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

const editingContent = ref('');
const editingColor = ref('#FFCDD2');
const currentNote = ref<Note | null>(null);

watch(() => props.note, (newNote) => {
  if (newNote) {
    currentNote.value = newNote;
    editingContent.value = newNote.content || '';
    editingColor.value = newNote.color || '#FFCDD2';
  }
}, { immediate: true });

const goBack = () => {
  emit('switch-view', 'notes');
};

const saveEdit = async () => {
  if (!props.filePath || !currentNote.value) return;

  const updatedNote: Note = {
    ...currentNote.value,
    content: editingContent.value,
    color: editingColor.value
  };

  try {
    // @ts-ignore
    await window.go.main.App.SaveNote(props.filePath, JSON.stringify(updatedNote));
    console.log('笔记已更新');
    eventBus.emit('note-saved', currentNote.value.cfi);
    goBack();
  } catch (err) {
    console.error('更新笔记失败:', err);
  }
};

const cancelEdit = () => {
  goBack();
};
</script>

<template>
  <div class="note-edit-wrapper">
    <div class="note-edit-content" v-if="currentNote">
      <div class="selected-text-box">
        <mark
          :style="{
            backgroundColor: editingColor,
            color: '#000'
          }"
        >
          {{ currentNote.selectedText || '无选中内容' }}
        </mark>
      </div>

      <textarea
        v-model="editingContent"
        class="note-textarea"
        placeholder="点击编辑笔记"
        autofocus
      ></textarea>
    </div>

    <div class="note-edit-footer">
      <button class="footer-btn cancel-btn" @click="cancelEdit">
        <XIcon :size="16" />
        <span>取消</span>
      </button>
      <button class="footer-btn save-btn" @click="saveEdit">
        <CheckIcon :size="16" />
        <span>保存</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.note-edit-wrapper {
  width: 100%;
  height: 100%;
  background: var(--sidebar-bg);
  color: var(--text-color);
  display: flex;
  flex-direction: column;
  user-select: none;
}

.note-edit-content {
  flex: 1;
  padding: 20px 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.selected-text-box {
  font-size: 0.8rem;
  line-height: 1.6;
  word-break: break-word;
  text-align: justify;
  padding: 12px 14px;
  background: rgba(var(--primary-rgb), 0.04);
  border-radius: var(--radius-sm);
  position: relative;
}

.selected-text-box mark {
  padding: 1px 3px;
  border-radius: 3px;
  box-decoration-break: clone;
  -webkit-box-decoration-break: clone;
}

.note-textarea {
  width: 100%;
  flex: 1;
  min-height: 180px;
  padding: 14px 16px;
  border: none !important;
  border-radius: var(--radius-md);
  background: rgba(0, 0, 0, 0.02);
  color: var(--text-primary);
  font-size: 0.9rem;
  font-family: inherit;
  line-height: 1.7;
  resize: none;
  box-sizing: border-box;
  outline: none !important;
  box-shadow: none !important;
}

.note-textarea:focus {
  outline: none !important;
  border: none !important;
  box-shadow: none !important;
}

.note-textarea:active {
  outline: none !important;
  border: none !important;
  box-shadow: none !important;
}

.note-textarea::placeholder {
  color: var(--text-muted);
  font-style: italic;
}

.note-edit-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  gap: 10px;
  background: var(--sidebar-bg);
}

.footer-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.cancel-btn {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

.cancel-btn:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-primary);
}

.save-btn {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(var(--primary-rgb), 0.3);
}

.save-btn:hover {
  opacity: 0.92;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(var(--primary-rgb), 0.4);
}

.save-btn:active {
  transform: translateY(0);
  box-shadow: 0 1px 4px rgba(var(--primary-rgb), 0.3);
}

.note-edit-content::-webkit-scrollbar {
  width: 4px;
}

.note-edit-content::-webkit-scrollbar-track {
  background: transparent;
}

.note-edit-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.note-edit-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>
