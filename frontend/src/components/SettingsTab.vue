<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useSettingsStore } from '../stores/settings'
import SettingsIcon from './icons/SettingsIcon.vue'
import FolderOpenIcon from './icons/FolderIcon.vue'
import WindowIcon from './icons/LayoutGridIcon.vue'
import FileTextIcon from './icons/FileIcon.vue'
import CustomModal from './CustomModal.vue'

const settingsStore = useSettingsStore()

const logs = ref<any[]>([])

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

interface LogEntry {
  timestamp: string
  action: string
  result: string
  type: string
}

const activeTab = ref<'all' | 'success' | 'skip' | 'error'>('all')

const filteredLogs = computed(() => {
  switch (activeTab.value) {
    case 'success':
      return logs.value.filter(log => log.type === 'SUCCESS')
    case 'skip':
      return logs.value.filter(log => log.type === 'SKIP')
    case 'error':
      return logs.value.filter(log => log.type === 'ERROR')
    default:
      return logs.value
  }
})

const successCount = computed(() => logs.value.filter(log => log.type === 'SUCCESS').length)
const skipCount = computed(() => logs.value.filter(log => log.type === 'SKIP').length)
const errorCount = computed(() => logs.value.filter(log => log.type === 'ERROR').length)

const loadLogs = async () => {
  try {
    // @ts-ignore
    logs.value = await window.go.main.App.GetWebDAVLogs()
  } catch (error) {
    console.error("加载日志失败:", error)
  }
}

const clearLogs = async () => {
  showConfirmModal(
    '确认清空',
    '确定要清空所有日志吗？',
    'warning',
    { onConfirm: async () => {
      try {
        // @ts-ignore
        await window.go.main.App.ClearWebDAVLogs()
        logs.value = []
      } catch (error) {
        console.error("清空日志失败:", error)
      }
    }}
  )
}

const handleOpenDataDir = async () => {
  try {
    // @ts-ignore
    await window.go.main.App.OpenConfigDir()
  } catch (error) {
    console.error("无法打开目录:", error)
    showConfirmModal('错误', '无法打开配置文件目录，请检查路径权限。', 'error', { showCancel: false })
  }
}

const handleLogUpdate = (newLog: LogEntry) => {
  logs.value = [newLog, ...logs.value].slice(0, 200)
}

const handleSaveWindowSize = async () => {
  try {
    // @ts-ignore
    await window.go.main.App.SaveWindowSize(settingsStore.windowWidth, settingsStore.windowHeight)
    showConfirmModal('成功', '窗口大小已保存，下次启动时生效。', 'success', { showCancel: false })
  } catch (error) {
    console.error("保存窗口大小失败:", error)
    showConfirmModal('错误', '保存窗口大小失败，请重试。', 'error', { showCancel: false })
  }
}

const resetWindowSize = () => {
  settingsStore.windowWidth = 1920
  settingsStore.windowHeight = 1080
}

const validateWindowWidth = () => {
  let value = String(settingsStore.windowWidth).replace(/\D/g, '')
  if (value === '') value = '800'
  let num = parseInt(value)
  if (num < 800) num = 800
  if (num > 4096) num = 4096
  settingsStore.windowWidth = num
}

const validateWindowHeight = () => {
  let value = String(settingsStore.windowHeight).replace(/\D/g, '')
  if (value === '') value = '600'
  let num = parseInt(value)
  if (num < 600) num = 600
  if (num > 2160) num = 2160
  settingsStore.windowHeight = num
}

watch([() => settingsStore.windowWidth, () => settingsStore.windowHeight], ([newWidth, newHeight]) => {
  if (newWidth < 800) settingsStore.windowWidth = 800
  if (newHeight < 600) settingsStore.windowHeight = 600
})

onMounted(async () => {
  await loadLogs()
  // @ts-ignore
  window.runtime?.EventsOn('webdav-log-updated', handleLogUpdate)
  // @ts-ignore
  window.go.main.App.GetWindowSize().then((result: string) => {
    try {
      const config = JSON.parse(result)
      settingsStore.windowWidth = config.width
      settingsStore.windowHeight = config.height
    } catch (e) {
      console.error("解析窗口大小失败:", e)
    }
  })
})

onUnmounted(() => {
  // @ts-ignore
  window.runtime?.EventsOff('webdav-log-updated', handleLogUpdate)
})
</script>

<template>
  <main class="settings-container">
    <div class="content-area">
      <div class="settings-content">
        <header class="settings-header">
          <div class="header-left">
            <SettingsIcon :size="28" />
            <div class="header-title-group">
              <h1>系统设置</h1>
            </div>
          </div>
        </header>

        <div class="settings-grid">
          <div class="settings-card">
            <div class="card-header">
              <WindowIcon :size="20" />
              <h2>启动窗口大小</h2>
            </div>
            <div class="card-content">
              <div class="size-input-group">
                <div class="size-input-wrapper">
                  <label class="input-label">宽度</label>
                  <input 
                    type="text" 
                    v-model="settingsStore.windowWidth" 
                    class="size-input"
                    @input="validateWindowWidth"
                  />
                  <span class="input-unit">px</span>
                </div>
                <span class="size-separator">×</span>
                <div class="size-input-wrapper">
                  <label class="input-label">高度</label>
                  <input 
                    type="text" 
                    v-model="settingsStore.windowHeight" 
                    class="size-input"
                    @input="validateWindowHeight"
                  />
                  <span class="input-unit">px</span>
                </div>
              </div>
              <p class="size-hint">当前设置: {{ settingsStore.windowWidth }} × {{ settingsStore.windowHeight }}</p>
              <div class="card-actions">
                <button class="btn secondary" @click="resetWindowSize">恢复默认</button>
                <button class="btn primary" @click="handleSaveWindowSize">保存设置</button>
              </div>
            </div>
          </div>

          <div class="settings-card">
            <div class="card-header">
              <FolderOpenIcon :size="20" />
              <h2>本地数据存储</h2>
            </div>
            <div class="card-content">
              <p class="card-desc">打开配置文件和书籍数据的存放位置</p>
              <button class="btn primary full-width" @click="handleOpenDataDir">
                <FolderOpenIcon :size="16" />
                打开目录
              </button>
            </div>
          </div>

          <div class="settings-card logs-card">
            <div class="card-header">
              <FileTextIcon :size="20" />
              <h2>WebDAV 交互日志</h2>
              <div class="card-actions-header">
                <span class="log-count">{{ logs.length }}/200</span>
                <button class="btn secondary small" @click="clearLogs">清空日志</button>
              </div>
            </div>
            <div class="card-content">
              <div class="logs-header">
                <div class="logs-tab-bar">
                  <button 
                    class="tab-chip" 
                    :class="{ active: activeTab === 'all' }"
                    @click="activeTab = 'all'"
                  >
                    <span class="tab-indicator"></span>
                    <span class="tab-label">全部</span>
                    <span class="tab-badge">{{ logs.length }}</span>
                  </button>
                  <button 
                    class="tab-chip success" 
                    :class="{ active: activeTab === 'success' }"
                    @click="activeTab = 'success'"
                  >
                    <span class="tab-indicator"></span>
                    <span class="tab-label">成功</span>
                    <span class="tab-badge">{{ successCount }}</span>
                  </button>
                  <button 
                    class="tab-chip skip" 
                    :class="{ active: activeTab === 'skip' }"
                    @click="activeTab = 'skip'"
                  >
                    <span class="tab-indicator"></span>
                    <span class="tab-label">跳过</span>
                    <span class="tab-badge">{{ skipCount }}</span>
                  </button>
                  <button 
                    class="tab-chip error" 
                    :class="{ active: activeTab === 'error' }"
                    @click="activeTab = 'error'"
                  >
                    <span class="tab-indicator"></span>
                    <span class="tab-label">失败</span>
                    <span class="tab-badge">{{ errorCount }}</span>
                  </button>
                </div>
              </div>
              
              <div class="logs-content">
                <div v-if="filteredLogs.length === 0" class="empty-logs">
                  <div class="empty-icon">📝</div>
                  <p>{{ activeTab === 'all' ? '暂无日志记录' : `暂无${activeTab === 'success' ? '成功' : activeTab === 'skip' ? '跳过' : '失败'}记录` }}</p>
                  <p class="empty-hint">执行同步操作后，日志将显示在这里</p>
                </div>
                <div v-else class="logs-list">
                  <div 
                    v-for="(log, index) in filteredLogs" 
                    :key="index" 
                    class="log-item"
                    :class="[`log-${log.type.toLowerCase()}`]"
                  >
                    <div class="log-indicator"></div>
                    <div class="log-icon">
                      <span v-if="log.type === 'SUCCESS'">✓</span>
                      <span v-else-if="log.type === 'SKIP'">→</span>
                      <span v-else>✗</span>
                    </div>
                    <div class="log-body">
                      <div class="log-action">{{ log.action }}</div>
                      <div class="log-meta">
                        <span class="log-timestamp">{{ log.timestamp }}</span>
                        <span class="log-result">{{ log.result }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <CustomModal
      :visible="confirmModal.visible"
      :title="confirmModal.title"
      :message="confirmModal.message"
      :type="confirmModal.type"
      :showCancel="confirmModal.showCancel"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    />
  </main>
</template>

<style scoped>
.settings-container { 
  flex: 1; 
  display: flex; 
  flex-direction: column; 
  background-color: var(--bg-color); 
  overflow: hidden;
}

.content-area { 
  flex: 1; 
  padding: 32px; 
  overflow-y: auto; 
  min-height: 0;
}

.settings-content { 
  max-width: 900px; 
  margin: 0 auto; 
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-title-group {
  display: flex;
  flex-direction: column;
}

.header-title-group h1 {
  margin: 0;
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text-color);
  letter-spacing: -0.02em;
}

.header-desc {
  margin: 6px 0 0 0;
  font-size: 0.95rem;
  color: var(--text-secondary);
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
}

.settings-card {
  background-color: var(--sidebar-bg);
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  border: 1px solid var(--border-color);
  transition: box-shadow var(--transition-fast);
}

.settings-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.03) 0%, transparent 100%);
}

.card-header h2 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-color);
}

.card-actions-header {
  margin-left: auto;
}

.card-content {
  padding: 24px;
}

.card-desc {
  margin: 0 0 16px 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.5;
}

.btn.full-width {
  width: 100%;
  justify-content: center;
  gap: 8px;
}

.btn.primary {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  border: none;
  box-shadow: var(--shadow-md);
}

.btn.primary:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg);
}

.btn.secondary {
  background: var(--bg-color);
  border: 1.5px solid var(--border-color);
  color: var(--text-secondary);
}

.btn.secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.btn.small {
  padding: 6px 14px;
  font-size: 0.85rem;
}

.size-input-group {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.size-input-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.input-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.size-input-wrapper {
  position: relative;
}

.size-input {
  width: 100%;
  padding: 14px 16px 14px 48px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-color);
  font-size: 1rem;
  font-weight: 600;
  transition: all var(--transition-fast);
  box-sizing: border-box;
}

.size-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.input-unit {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 0.85rem;
  color: var(--text-secondary);
  pointer-events: none;
}

.size-separator {
  font-size: 1.5rem;
  font-weight: 300;
  color: var(--border-color);
}

.size-hint {
  margin: 0 0 20px 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
  opacity: 0.8;
}

.card-actions {
  display: flex;
  gap: 12px;
}

.card-actions .btn {
  flex: 1;
  padding: 12px;
  font-size: 0.9rem;
  font-weight: 500;
}

.logs-card {
  grid-column: 1 / -1;
}

.log-count {
  font-size: 0.75rem;
  color: var(--text-secondary);
  opacity: 0.7;
  margin-right: 12px;
}

.logs-header {
  margin-bottom: 0;
}

.logs-header .card-desc {
  margin-bottom: 14px;
}

.logs-tab-bar {
  display: flex;
  gap: 8px;
  padding: 4px;
  background-color: var(--bg-color);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.tab-chip {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  position: relative;
  overflow: hidden;
  border: none;
}

.tab-chip:hover {
  background-color: rgba(0, 0, 0, 0.04);
  color: var(--text-color);
}

.tab-chip.active {
  background-color: var(--primary-color);
  color: #fff;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
}

.tab-chip.active .tab-badge {
  background-color: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.tab-chip.success.active {
  background-color: #10b981;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.tab-chip.skip.active {
  background-color: #f59e0b;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.3);
}

.tab-chip.error.active {
  background-color: #ef4444;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
}

.tab-indicator {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--primary-color);
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.tab-chip.active .tab-indicator {
  opacity: 1;
}

.tab-chip.success.active .tab-indicator {
  background: #10b981;
}

.tab-chip.skip.active .tab-indicator {
  background: #f59e0b;
}

.tab-chip.error.active .tab-indicator {
  background: #ef4444;
}

.tab-label {
  position: relative;
  z-index: 1;
}

.tab-badge {
  position: relative;
  z-index: 1;
  background-color: rgba(0, 0, 0, 0.06);
  color: var(--text-secondary);
  padding: 1px 9px;
  border-radius: 10px;
  font-size: 0.7rem;
  min-width: 20px;
  text-align: center;
  font-weight: 600;
}

.logs-content {
  max-height: 400px;
  overflow-y: auto;
  margin-top: 16px;
  padding-right: 4px;
}

.logs-content::-webkit-scrollbar {
  width: 6px;
}

.logs-content::-webkit-scrollbar-track {
  background: transparent;
}

.logs-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.logs-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

.empty-logs { 
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px; 
  color: var(--text-color); 
}

.empty-icon {
  font-size: 2.5rem;
  margin-bottom: 12px;
  opacity: 0.5;
}

.empty-logs p {
  margin: 0;
  font-size: 0.9rem;
  opacity: 0.6;
}

.empty-hint {
  margin-top: 6px !important;
  font-size: 0.8rem !important;
  opacity: 0.4 !important;
}

.logs-list { 
  font-family: 'Monaco', 'Menlo', monospace; 
  font-size: 0.82rem; 
}

.log-item { 
  display: flex; 
  align-items: stretch; 
  padding: 12px 16px; 
  border-radius: var(--radius-md); 
  margin-bottom: 6px; 
  transition: all var(--transition-fast);
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
}

.log-item:hover { 
  background-color: rgba(99, 102, 241, 0.03);
  border-color: rgba(99, 102, 241, 0.2);
  transform: translateX(2px);
}

.log-indicator {
  width: 4px;
  border-radius: 2px;
  margin-right: 12px;
  flex-shrink: 0;
}

.log-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
  flex-shrink: 0;
  margin-right: 12px;
}

.log-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.log-action { 
  color: var(--text-color); 
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.75rem;
}

.log-timestamp { 
  color: #6b7280; 
}

.log-result { 
  font-weight: 600; 
  white-space: nowrap;
}

.log-success .log-indicator { 
  background-color: #10b981; 
}

.log-success .log-icon { 
  background-color: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.log-success .log-action { 
  color: #374151; 
}

.log-success .log-result { 
  color: #10b981; 
}

.log-skip .log-indicator { 
  background-color: #f59e0b; 
}

.log-skip .log-icon { 
  background-color: rgba(245, 158, 11, 0.1);
  color: #d97706;
}

.log-skip .log-action { 
  color: #6b7280; 
}

.log-skip .log-result { 
  color: #d97706; 
}

.log-error .log-indicator { 
  background-color: #ef4444; 
}

.log-error .log-icon { 
  background-color: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.log-error { 
  background-color: rgba(239, 68, 68, 0.03);
  border-color: rgba(239, 68, 68, 0.15);
}

.log-error .log-action { 
  color: var(--text-color); 
}

.log-error .log-result { 
  color: #ef4444; 
}

@media (max-width: 900px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
}
</style>