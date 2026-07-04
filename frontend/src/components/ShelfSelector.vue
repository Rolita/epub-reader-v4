<script setup lang="ts">
import { computed } from 'vue'
import { useLibraryStore } from '../stores/library'
import BookIcon from './icons/BookIcon.vue'
import CustomModalEx from './CustomModalEx.vue'

interface ShelfSelectorProps {
  visible: boolean
  epubPath: string
}

const props = defineProps<ShelfSelectorProps>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', shelfId: string, epubPath: string): void
  (e: 'import-only', shelfId: string, epubPath: string): void
}>()

const store = useLibraryStore()

const filteredShelves = computed(() => {
  return store.shelves
})

const selectShelf = (shelfId: string) => {
  emit('select', shelfId, props.epubPath)
}

const handleClose = () => {
  emit('close')
}
</script>

<template>
  <CustomModalEx
    :visible="visible"
    title="选择书架"
    :show-footer="true"
    :show-cancel="true"
    cancel-text="取消"
    :max-width="'420px'"
    @cancel="handleClose"
  >
    <p class="epub-path">
      即将导入：{{ epubPath.split(/[\\/]/).pop() }}
    </p>
    
    <div class="shelf-list">
      <div
        v-for="shelf in filteredShelves"
        :key="shelf.id"
        class="shelf-item"
      >
        <div class="shelf-info" @click="selectShelf(shelf.id)">
          <BookIcon :size="24" />
          <span class="shelf-name">{{ shelf.name }}</span>
        </div>
        <button 
          class="import-only-btn" 
          @click.stop="emit('import-only', shelf.id, props.epubPath)"
        >
          仅导入
        </button>
      </div>
    </div>
  </CustomModalEx>
</template>

<style scoped>
.epub-path {
  margin: 0 0 16px 0;
  font-size: 13px;
  color: var(--text-secondary-color);
  word-break: break-all;
}

.shelf-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 280px;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.shelf-list::-webkit-scrollbar {
  width: 6px;
}

.shelf-list::-webkit-scrollbar-track {
  background: transparent;
}

.shelf-list::-webkit-scrollbar-thumb {
  background: var(--border-color, #dcdfe6);
  border-radius: 3px;
}

.shelf-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.shelf-item:hover {
  background: var(--primary-light, rgba(74, 144, 217, 0.08));
}

.shelf-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.shelf-info svg {
  flex-shrink: 0;
}

.shelf-name {
  font-size: 14px;
  color: var(--text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.import-only-btn {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary-color);
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.import-only-btn:hover {
  background: var(--primary-color);
  border-color: var(--primary-color);
  color: white;
}
</style>
