<template>
  <Teleport to="body" :disabled="false">
    <Transition name="modal" :appear="false">
      <div v-if="visible" class="modal-overlay" @click.self="handleOverlayClick">
        <div class="modal-dialog" :style="dialogStyle">
          <h3 v-if="title" class="modal-title">{{ title }}</h3>
          <div class="modal-content">
            <slot></slot>
          </div>
          <div v-if="showFooter" class="modal-actions">
            <button v-if="showCancel" class="btn-cancel" @click="handleCancel">
              {{ cancelText }}
            </button>
            <button class="btn-confirm" :disabled="confirmDisabled" @click="handleConfirm">
              {{ confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onBeforeUnmount, computed } from 'vue'

const props = withDefaults(defineProps<{
  visible: boolean
  title?: string
  confirmText?: string
  cancelText?: string
  showCancel?: boolean
  showFooter?: boolean
  confirmDisabled?: boolean
  maxWidth?: string
}>(), {
  title: '提示',
  confirmText: '确定',
  cancelText: '取消',
  showCancel: true,
  showFooter: true,
  confirmDisabled: false,
  maxWidth: '420px'
})

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const handleOverlayClick = () => {
  if (props.showCancel) emit('cancel')
}

const handleConfirm = () => emit('confirm')
const handleCancel = () => emit('cancel')

// 提前计算 style 对象，避免在 Teleport 渲染时动态计算导致 reflow
const dialogStyle = computed(() => ({
  maxWidth: props.maxWidth
}))

let savedOverflow = ''

watch(() => props.visible, (val) => {
  if (val) {
    savedOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = savedOverflow
  }
})

onBeforeUnmount(() => {
  if (document.body.style.overflow === 'hidden') {
    document.body.style.overflow = savedOverflow
  }
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
}

.modal-dialog {
  min-width: 280px;
  max-width: 90vw;
  min-height: 160px;  /* 稳定最小高度，避免内容变化时抖动 */
  background: var(--sidebar-bg, #fafafa);
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.modal-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color, #303133);
  margin: 0 0 16px 0;
}

.modal-content {
  font-size: 13px;
  color: var(--text-color, #303133);
  line-height: 1.5;
  flex: 1;
}

/* 模态框标签 */
.modal-label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary, #606266);
  margin-bottom: 8px;
  font-weight: 500;
}

/* 模态框输入框 */
.modal-input {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid var(--border-color, #dcdfe6);
  border-radius: 8px;
  background: var(--bg-color, #fff);
  color: var(--text-color, #303133);
  font-size: 14px;
  box-sizing: border-box;
  outline: none;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.modal-input:focus {
  border-color: var(--primary-color, #4A90D9);
  box-shadow: 0 0 0 3px rgba(74, 144, 217, 0.1);
}

.modal-input::placeholder {
  color: var(--text-muted, #909399);
}

/* 分组列表 */
.group-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  /* 固定滚动条空间，避免出现/消失导致抖动 */
  max-height: 200px;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.group-list::-webkit-scrollbar {
  width: 6px;
}

.group-list::-webkit-scrollbar-track {
  background: transparent;
}

.group-list::-webkit-scrollbar-thumb {
  background: var(--border-color, #dcdfe6);
  border-radius: 3px;
}

.group-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
  flex-shrink: 0;
}

.group-option:hover {
  background: var(--primary-light, rgba(74, 144, 217, 0.08));
}

.group-option.active {
  background: var(--primary-color, #4A90D9);
  color: #fff;
}

.group-option .group-name {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-option .group-count {
  font-size: 12px;
  opacity: 0.6;
}

/* 分组创建按钮 */
.group-add-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 12px;
  padding: 10px 16px;
  border: 1px dashed var(--border-color, #dcdfe6);
  border-radius: 8px;
  cursor: pointer;
  color: var(--primary-color, #4A90D9);
  font-size: 13px;
  font-weight: 500;
  transition: all 0.15s ease;
  user-select: none;
}

.group-add-row:hover {
  background: var(--primary-light, rgba(74, 144, 217, 0.08));
  border-color: var(--primary-color, #4A90D9);
  border-style: solid;
}

.group-empty {
  text-align: center;
  color: var(--text-muted, #909399);
  padding: 16px;
  font-size: 13px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 20px;
}

.btn-cancel,
.btn-confirm {
  padding: 8px 20px;
  font-size: 13px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-cancel {
  background: transparent;
  color: var(--text-color, #606266);
  border: 1px solid var(--border-color, #dcdfe6);
}

.btn-cancel:hover {
  border-color: var(--primary-color, #4A90D9);
  color: var(--primary-color, #4A90D9);
}

.btn-confirm {
  background: linear-gradient(135deg, var(--primary-color, #4A90D9), var(--accent-color, #67B8DE));
  color: #fff;
  border: none;
}

.btn-confirm:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-confirm:active:not(:disabled) {
  transform: scale(0.98);
}

.btn-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 动画 - 仅淡入淡出，不涉及位置/尺寸变化，彻底避免抖动 */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.15s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>