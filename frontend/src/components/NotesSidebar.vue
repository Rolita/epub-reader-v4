<template>
  <div class="notes-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><NoteIcon :size="22" style="margin-right: 8px;" />笔记</h2>
    </div>
    <div class="notes-content">
      <div class="empty-state">
        <NoteIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>暂无笔记</span>
        <p style="font-size: 0.8rem; color: var(--text-muted); margin-top: 8px;">在阅读时添加笔记</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import NoteIcon from './icons/NoteIcon.vue';
// 笔记侧边栏组件，暂时无逻辑
</script>

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

.notes-content {
  flex: 1;
  padding: 16px;
  color: var(--text-secondary);
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