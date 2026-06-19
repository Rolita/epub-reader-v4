<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import SettingsIcon from './icons/SettingsIcon.vue'
import CustomModal from './CustomModal.vue'

const logs = ref<any[]>([])

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

interface LogEntry {
  timestamp: string
  action: string
  result: string
  type: string // "SUCCESS", "ERROR", "SKIP"
}

// 当前选中的标签
const activeTab = ref<'all' | 'success' | 'skip' | 'error'>('all')

// 按类型过滤日志
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

// 统计各类型日志数量
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

// 监听日志更新事件
const handleLogUpdate = (newLog: LogEntry) => {
  logs.value = [newLog, ...logs.value].slice(0, 50)
}

onMounted(async () => {
  await loadLogs()
  // @ts-ignore
  window.runtime?.EventsOn('webdav-log-updated', handleLogUpdate)
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
        <h1 style="display: flex; align-items: center;"><SettingsIcon :size="22" style="margin-right: 10px;" />系统设置</h1>
        
        <div class="setting-section">
          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-title">本地数据存储</span>
              <span class="setting-desc">打开配置文件的存放位置</span>
            </div>
            <button class="btn secondary" @click="handleOpenDataDir">打开目录</button>
          </div>
        </div>

        <!-- WebDAV 交互日志面板 -->
        <div class="setting-section">
          <div class="section-header">
            <div class="setting-info">
              <span class="setting-title">WebDAV 交互日志</span>
              <span class="setting-desc">最近 200 条操作记录</span>
            </div>
            <div class="section-actions">
              <button class="btn secondary small" @click="clearLogs">清空日志</button>
            </div>
          </div>

          <div class="logs-container">
            <!-- 标签切换按钮 -->
            <div class="logs-tab-buttons">
              <button 
                class="tab-btn" 
                :class="{ active: activeTab === 'all' }"
                @click="activeTab = 'all'"
              >
                全部
                <span class="tab-count">{{ logs.length }}</span>
              </button>
              <button 
                class="tab-btn success" 
                :class="{ active: activeTab === 'success' }"
                @click="activeTab = 'success'"
              >
                ✓ 成功
                <span class="tab-count">{{ successCount }}</span>
              </button>
              <button 
                class="tab-btn skip" 
                :class="{ active: activeTab === 'skip' }"
                @click="activeTab = 'skip'"
              >
                → 跳过
                <span class="tab-count">{{ skipCount }}</span>
              </button>
              <button 
                class="tab-btn error" 
                :class="{ active: activeTab === 'error' }"
                @click="activeTab = 'error'"
              >
                ✗ 失败
                <span class="tab-count">{{ errorCount }}</span>
              </button>
            </div>
            
            <!-- 日志内容区域 -->
            <div class="logs-content">
              <div v-if="filteredLogs.length === 0" class="empty-logs">
                {{ activeTab === 'all' ? '暂无日志记录' : `暂无${activeTab === 'success' ? '成功' : activeTab === 'skip' ? '跳过' : '失败'}记录` }}
              </div>
              <div v-else class="logs-list">
                <div 
                  v-for="(log, index) in filteredLogs" 
                  :key="index" 
                  class="log-item"
                  :class="[`log-${log.type.toLowerCase()}`]"
                >
                  <span class="log-timestamp">{{ log.timestamp }}</span>
                  <span class="log-action">{{ log.action }}</span>
                  <span class="log-result">{{ log.result }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
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
  </main>
</template>

<style scoped>
.settings-container { flex: 1; display: flex; flex-direction: column; background-color: var(--bg-color); }
.content-area { flex: 1; padding: 30px; overflow-y: auto; display: flex; justify-content: center; }
.settings-content { width: 100%; max-width: 800px; }
.settings-content h1 { font-size: 1.5rem; color: var(--text-color); margin-bottom: 30px; }

.setting-section { background-color: var(--sidebar-bg); border-radius: 12px; padding: 20px; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1); margin-bottom: 20px; }
.setting-item { display: flex; justify-content: space-between; align-items: center; padding: 15px 0; border-bottom: 1px solid var(--border-color); }
.setting-item:last-child { border-bottom: none; }
.setting-info { display: flex; flex-direction: column; }
.setting-title { font-weight: bold; color: var(--text-color); font-size: 1rem; }
.setting-desc { font-size: 0.85rem; color: var(--text-color); margin-top: 4px; opacity: 0.6; }

.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; padding-bottom: 15px; border-bottom: 1px solid var(--border-color); }
.section-actions { display: flex; align-items: center; gap: 12px; }
.logs-container { max-height: 450px; display: flex; flex-direction: column; }

/* 标签切换按钮 */
.logs-tab-buttons {
  display: flex;
  gap: 8px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 12px;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: transparent;
  color: var(--text-secondary);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.tab-btn.active {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
  color: #fff;
}

.tab-btn.active .tab-count {
  background-color: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.tab-btn.success.active {
  background-color: #10b981;
  border-color: #10b981;
}

.tab-btn.skip.active {
  background-color: #f59e0b;
  border-color: #f59e0b;
}

.tab-btn.error.active {
  background-color: #ef4444;
  border-color: #ef4444;
}

.tab-count {
  background-color: var(--border-color);
  color: var(--text-secondary);
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.7rem;
  min-width: 20px;
  text-align: center;
}

/* 日志内容区域 */
.logs-content {
  flex: 1;
  overflow-y: auto;
}

.empty-logs { text-align: center; padding: 40px; color: var(--text-color); opacity: 0.5; font-size: 0.9rem; }

.logs-list { font-family: 'Monaco', 'Menlo', monospace; font-size: 0.85rem; }
.log-item { display: flex; align-items: center; padding: 8px 12px; border-radius: 6px; margin-bottom: 4px; gap: 12px; border-left: 3px solid transparent; }
.log-item:hover { background-color: rgba(0, 0, 0, 0.05); }

.log-timestamp { color: #6b7280; min-width: 60px; }
.log-action { flex: 1; color: var(--text-color); }
.log-result { font-weight: bold; }

/* SUCCESS - 绿色边框，灰色字体 */
.log-success { border-left-color: #10b981; }
.log-success .log-action { color: #6b7280; }
.log-success .log-result { color: #10b981; }

/* SKIP - 黄色边框，淡灰色字体 */
.log-skip { border-left-color: #f59e0b; }
.log-skip .log-action { color: #9ca3af; }
.log-skip .log-result { color: #d97706; }

/* ERROR - 红色边框，黑色字体 */
.log-error { border-left-color: #ef4444; background-color: rgba(239, 68, 68, 0.05); }
.log-error .log-action { color: var(--text-color); }
.log-error .log-result { color: #ef4444; }

.btn.secondary.small {
    padding: 8px 16px;
    font-size: 1rem;
}

.filter-checkbox input{
padding: 6px 6px;
}
</style>
