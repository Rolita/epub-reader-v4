<script setup lang="ts">
import { ref } from 'vue'
import { useLibraryStore } from '../stores/library'
import BookIcon from './icons/BookIcon.vue'
import CustomModal from './CustomModal.vue'
import CustomModalEx from './CustomModalEx.vue'

const emit = defineEmits<{
  (e: 'switch-view', viewName: string): void
  (e: 'open-shelf', shelfId: string, shelfName: string): void
}>()

const store = useLibraryStore()

// 确认弹窗状态
const confirmModal = ref({
  visible: false,
  title: '',
  message: '',
  type: 'info' as 'info' | 'warning' | 'success' | 'error',
  showCancel: true,
  onConfirm: () => {}
})

const showConfirmModal = (
  title: string,
  message: string,
  type: 'info' | 'warning' | 'success' | 'error' = 'info',
  options?: { showCancel?: boolean; onConfirm?: () => void }
) => {
  confirmModal.value = {
    visible: true,
    title,
    message,
    type,
    showCancel: options?.showCancel ?? true,
    onConfirm: options?.onConfirm || (() => {})
  }
}

const handleConfirm = () => {
  confirmModal.value.onConfirm()
  confirmModal.value.visible = false
}

const handleCancel = () => {
  confirmModal.value.visible = false
}

// 新建书架弹窗状态
const showAddShelfDialog = ref(false)
const newShelfName = ref('')

const handleAddShelf = () => {
  newShelfName.value = ''
  showAddShelfDialog.value = true
}

const handleConfirmAddShelf = () => {
  if (newShelfName.value.trim()) {
    store.createShelf(newShelfName.value.trim())
    showAddShelfDialog.value = false
    newShelfName.value = ''
  }
}

const handleCancelAddShelf = () => {
  showAddShelfDialog.value = false
  newShelfName.value = ''
}

// 重命名书架弹窗状态
const showEditShelfDialog = ref(false)
const editShelfName = ref('')
const editShelfId = ref('')

const handleEditShelf = (id: string, oldName: string) => {
  editShelfId.value = id
  editShelfName.value = oldName
  showEditShelfDialog.value = true
}

const handleConfirmEditShelf = () => {
  if (editShelfName.value.trim() && editShelfId.value) {
    store.renameShelf(editShelfId.value, editShelfName.value.trim())
  }
  showEditShelfDialog.value = false
  editShelfName.value = ''
  editShelfId.value = ''
}

const handleCancelEditShelf = () => {
  showEditShelfDialog.value = false
  editShelfName.value = ''
  editShelfId.value = ''
}

// 拖拽状态
const draggedIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

// 拖拽开始
const handleDragStart = (event: DragEvent, index: number) => {
  draggedIndex.value = index
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', index.toString())
  }
}

// 拖拽进入
const handleDragEnter = (event: DragEvent, index: number) => {
  event.preventDefault()
  if (draggedIndex.value !== null && draggedIndex.value !== index) {
    dragOverIndex.value = index
  }
}

// 拖拽悬停
const handleDragOver = (event: DragEvent) => {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
}

// 拖拽离开
const handleDragLeave = () => {
  dragOverIndex.value = null
}

// 拖拽结束
const handleDragEnd = () => {
  draggedIndex.value = null
  dragOverIndex.value = null
}

// 点击书架
const handleShelfClick = (shelfId: string, shelfName: string) => {
  store.setActiveShelf(shelfId)
  emit('open-shelf', shelfId, shelfName)
}

// 放置
const handleDrop = (event: DragEvent, dropIndex: number) => {
  event.preventDefault()
  
  if (draggedIndex.value === null || draggedIndex.value === dropIndex) {
    handleDragEnd()
    return
  }
  
  // 复制当前书架列表
  const newOrder = [...store.shelves]
  const [removed] = newOrder.splice(draggedIndex.value, 1)
  newOrder.splice(dropIndex, 0, removed)
  
  // 保存新顺序
  store.reorderShelves(newOrder)
  
  handleDragEnd()
}

const handleDeleteShelf = (id: string) => {
  showConfirmModal(
    '确认删除',
    '确定要删除这个书架吗？',
    'warning',
    { onConfirm: () => {
      store.deleteShelf(id)
    }}
  )
}
</script>

<template>
  <div class="shelf-wrapper">
    <div class="logo-area" style="--wails-draggable: drag;">
      <h1>📚 EPUB Reader</h1>
    </div>

    <div class="shelves-area">
      <div class="shelves-header">
        <span>我的书架</span>
        <button class="icon-btn" @click="handleAddShelf" title="添加书架">+</button>
      </div>
      <ul class="shelf-list">
        <li 
          v-for="(shelf, index) in store.shelves" 
          :key="shelf.id"
          :class="[
            'shelf-item', 
            { 
              active: store.activeShelfId === shelf.id,
              dragging: draggedIndex === index,
              'drag-over': dragOverIndex === index
            }
          ]"
          draggable="true"
          @dragstart="handleDragStart($event, index)"
          @dragenter="handleDragEnter($event, index)"
          @dragover="handleDragOver"
          @dragleave="handleDragLeave"
          @dragend="handleDragEnd"
          @drop="handleDrop($event, index)"
          @click="handleShelfClick(shelf.id, shelf.name)"
        >
          <span class="drag-handle" title="拖拽排序">⋮⋮</span>
          <span class="shelf-name">{{ shelf.name }}</span>
          <div class="shelf-actions">
            <button class="action-btn edit" @click.stop="handleEditShelf(shelf.id, shelf.name)">✎</button>
            <button class="action-btn delete" @click.stop="handleDeleteShelf(shelf.id)">✖</button>
          </div>
        </li>
      </ul>
    </div>

    <!-- 确认弹窗 -->
    <CustomModal
      :visible="confirmModal.visible"
      :title="confirmModal.title"
      :message="confirmModal.message"
      :type="confirmModal.type"
      :showCancel="confirmModal.showCancel"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    />
    
    <!-- 新建书架弹窗 -->
    <CustomModalEx
      :visible="showAddShelfDialog"
      title="新建书架"
      confirmText="创建"
      cancelText="取消"
      :confirmDisabled="!newShelfName.trim()"
      @confirm="handleConfirmAddShelf"
      @cancel="handleCancelAddShelf"
    >
      <input 
        v-model="newShelfName" 
        type="text" 
        class="modal-input" 
        placeholder="请输入书架名称"
        @keyup.enter="handleConfirmAddShelf"
        autofocus
      />
    </CustomModalEx>
    
    <!-- 重命名书架弹窗 -->
    <CustomModalEx
      :visible="showEditShelfDialog"
      title="修改书架名称"
      confirmText="确定"
      cancelText="取消"
      :confirmDisabled="!editShelfName.trim()"
      @confirm="handleConfirmEditShelf"
      @cancel="handleCancelEditShelf"
    >
      <input 
        v-model="editShelfName" 
        type="text" 
        class="modal-input" 
        placeholder="请输入书架名称"
        @keyup.enter="handleConfirmEditShelf"
        autofocus
      />
    </CustomModalEx>
  </div>
</template>

<style scoped>
.shelf-wrapper {
  width: 100%;
  height: 100%;
  background-color: var(--sidebar-bg);
  color: var(--text-color);
  display: flex;
  flex-direction: column;
  user-select: none;
  border-right: 0px solid var(--border-color);
  position: relative;
}

/* Logo区域 - 现代化渐变背景 */
.logo-area {
  padding: 25.8px 20px;
  text-align: center;
  border-bottom: 1px solid var(--border-color);
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.08) 0%, transparent 100%);
  position: relative;
  overflow: hidden;
}

.logo-area::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 60px;
  height: 60px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.15) 0%, transparent 70%);
  border-radius: 50%;
}

.logo-area h1 {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: -0.03em;
  position: relative;
  z-index: 1;
}

/* 书架区域 */
.shelves-area {
  flex: 1;
  padding: 20px 10px;
  overflow-y: auto;
}

.shelves-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 8px 16px;
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

/* 添加按钮 */
.icon-btn {
  background: var(--primary-light);
  border: 1px solid transparent;
  color: var(--primary-color);
  cursor: pointer;
  font-size: 1.1rem;
  font-weight: 500;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  transition: all var(--transition-normal);
  padding-bottom: 4px;
}

.icon-btn:hover {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  transform: scale(1.05);
}

/* 书架列表 */
.shelf-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

/* 书架项 - 现代化卡片样式 */
.shelf-item {
  padding: 12px 14px;
  margin-bottom: 6px;
  border-radius: var(--radius-lg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: all var(--transition-normal);
  position: relative;
  background: transparent;
}

.shelf-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 0;
  background: var(--primary-color);
  border-radius: 0 3px 3px 0;
  transition: height var(--transition-normal);
}

.shelf-item:hover::before {
  height: 60%;
}

.shelf-item:hover {
  background: var(--primary-light);
  transform: translateX(6px);
}

.shelf-item.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  box-shadow: var(--shadow-lg);
  transform: translateX(4px);
}

.shelf-item.active::before {
  height: 100%;
  background: rgba(255, 255, 255, 0.4);
}

.shelf-item.dragging {
  opacity: 0.5;
  transform: scale(0.96) rotate(1deg);
  box-shadow: var(--shadow-xl);
}

.shelf-item.drag-over {
  border-top: 3px solid var(--primary-color);
  margin-top: -3px;
  background: var(--primary-light);
}

.shelf-item:active {
}

/* 拖拽手柄 */
.drag-handle {
  color: var(--text-muted);
  opacity: 0;
  padding: 4px 8px;
  font-size: 0.7rem;
  transition: all var(--transition-fast);
  border-radius: var(--radius-sm);
  margin-right: 8px;
}

.shelf-item:hover .drag-handle {
  opacity: 0.5;
}

.shelf-item:hover .drag-handle:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.08);
}

.shelf-item.active .drag-handle {
  color: rgba(255, 255, 255, 0.7);
}

.shelf-item.active .drag-handle:hover {
  background: rgba(255, 255, 255, 0.15);
}

/* 书架名称 */
.shelf-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  font-size: 0.9rem;
  font-weight: 500;
  letter-spacing: -0.01em;
}

/* 操作按钮组 */
.shelf-actions {
  display: none;
  gap: 4px;
}

.shelf-item:hover .shelf-actions {
  display: flex;
}

.action-btn {
  background: rgba(0, 0, 0, 0.05);
  border: none;
  cursor: pointer;
  color: inherit;
  font-size: 0.75rem;
  opacity: 0.6;
  width: 26px;
  height: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.action-btn:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.12);
}

.shelf-item.active .action-btn {
  background: rgba(255, 255, 255, 0.1);
}

.shelf-item.active .action-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* 滚动条美化 */
.shelves-area::-webkit-scrollbar {
  width: 4px;
}

.shelves-area::-webkit-scrollbar-track {
  background: transparent;
}

.shelves-area::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.shelves-area::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>
