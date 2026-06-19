<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="modal-overlay" @click.self="handleOverlayClick">
        <div class="modal-dialog">
          <h3 class="modal-title">{{ title }}</h3>
          <p class="modal-message">{{ message }}</p>
          <div class="modal-actions">
            <button v-if="showCancel" class="btn-cancel" @click="handleCancel">
              {{ cancelText }}
            </button>
            <button class="btn-confirm" @click="handleConfirm">
              {{ confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch } from 'vue'

const props = withDefaults(defineProps<{
  visible: boolean
  title?: string
  message: string
  type?: 'info' | 'warning' | 'success' | 'error'
  confirmText?: string
  cancelText?: string
  showCancel?: boolean
}>(), {
  title: '提示',
  type: 'info',
  confirmText: '确定',
  cancelText: '取消',
  showCancel: true
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

watch(() => props.visible, (val) => {
  document.body.style.overflow = val ? 'hidden' : ''
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
  max-width: 360px;
  background: var(--sidebar-bg, #fafafa);
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
}

.modal-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color, #303133);
  margin: 0 0 8px 0;
}

.modal-message {
  font-size: 13px;
  color: var(--text-color, #303133);
  opacity: 0.7;
  margin: 0 0 16px 0;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
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

.btn-confirm:hover {
  filter: brightness(1.1);
}

.btn-confirm:active {
  transform: scale(0.98);
}

/* 动画 */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal-dialog,
.modal-leave-active .modal-dialog {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-dialog,
.modal-leave-to .modal-dialog {
  opacity: 0;
  transform: scale(0.95);
}
</style>