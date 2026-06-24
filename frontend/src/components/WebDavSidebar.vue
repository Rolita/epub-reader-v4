<script setup lang="ts">
import { ref, watch } from 'vue';
import { useLibraryStore } from '../stores/library';
import CloudIcon from './icons/CloudIcon.vue';

const emit = defineEmits<{
  (e: 'switch-view', viewName: string): void
  (e: 'sync-complete'): void
  (e: 'show-toast', message: string, type: 'success' | 'error'): void
}>();

// 显示消息提示（通知父组件）
const showToast = (message: string, type: 'success' | 'error') => {
  statusMessage.value = message;
  statusType.value = type;
  emit('show-toast', message, type);
};

const store = useLibraryStore();

// 配置表单
const config = ref({
  base_url: '',
  remote_path: '',
  username: '',
  password: ''
});

const connecting = ref(false);
const uploading = ref(false);
const downloading = ref(false);
const statusMessage = ref('');
const statusType = ref<'success' | 'error' | ''>('');

// 获取当前激活的书架名称
const currentShelfName = ref<string | null>(null);

// ============ 先定义所有函数 ============

// 加载配置
const loadConfig = async (shelfName: string) => {
  if (!shelfName) return;
  try {
    // @ts-ignore
    const data = await window.go.main.App.LoadWebDavConfig(shelfName);
    if (data) {
      const cfg = JSON.parse(data);
      config.value = {
        base_url: cfg.base_url || '',
        remote_path: cfg.remote_path || '',
        username: cfg.username || '',
        password: cfg.password || ''
      };
    }
  } catch (e) {
    console.log('加载配置失败，使用默认值:', e);
  }
};

// 保存配置
const saveConfig = async () => {
  if (!currentShelfName.value) {
    showToast('请先选择一个书架', 'error');
    return;
  }
  
  try {
    // @ts-ignore
    await window.go.main.App.SaveWebDavConfig(currentShelfName.value, JSON.stringify(config.value));
    showToast('配置已保存', 'success');
  } catch (e) {
    showToast('保存失败: ' + (e as Error).message, 'error');
  }
};

// 测试连接
const handleTestConnection = async () => {
  if (!currentShelfName.value) {
    showToast('请先选择一个书架', 'error');
    return;
  }
  
  // 先保存配置再测试
  await saveConfig();
  
  connecting.value = true;
  try {
    // @ts-ignore
    const res = await window.go.main.App.TestShelfWebDav(currentShelfName.value);
    showToast('连接成功: ' + res, 'success');
  } catch (e) {
    showToast('连接失败: ' + (e as Error).message, 'error');
  } finally {
    connecting.value = false;
  }
};

// 上传到云端
const handleUpload = async () => {
  if (!currentShelfName.value) {
    showToast('请先选择一个书架', 'error');
    return;
  }
  
  // 先保存配置
  await saveConfig();
  
  uploading.value = true;
  try {
    // @ts-ignore
    const res = await window.go.main.App.UploadShelf(currentShelfName.value);
    showToast('上传成功: ' + res, 'success');
  } catch (e) {
    showToast('上传失败: ' + (e as Error).message, 'error');
  } finally {
    uploading.value = false;
  }
};

// 从云端下载
const handleDownload = async () => {
  if (!currentShelfName.value) {
    showToast('请先选择一个书架', 'error');
    return;
  }
  
  downloading.value = true;
  try {
    // @ts-ignore
    const res = await window.go.main.App.DownloadShelf(currentShelfName.value);
    showToast('下载成功: ' + res, 'success');
    
    // 下载成功后刷新书架界面并通知父组件
    store.scanShelves();
    if (currentShelfName.value) {
      await store.loadShelfBooks(currentShelfName.value);
    }
    emit('sync-complete');
  } catch (e) {
    showToast('下载失败: ' + (e as Error).message, 'error');
  } finally {
    downloading.value = false;
  }
};

// 返回书架列表
const goBack = () => {
  emit('switch-view', 'shelf');
};

// ============ 后定义 watch ============

watch(() => store.activeShelfId, (newId) => {
  currentShelfName.value = newId;
  // 加载现有配置
  if (newId) {
    loadConfig(newId);
  }
}, { immediate: true });
</script>

<template>
  <div class="webdav-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><CloudIcon :size="18" style="margin-right: 8px;" />WebDAV 同步</h2>
    </div>
    
    <div class="webdav-content">
      <!-- 当前书架 -->
      <div class="current-shelf">
        <span>当前书架:</span>
        <span class="shelf-name">{{ currentShelfName || '未选择' }}</span>
      </div>
      
      <!-- 配置表单 -->
      <div class="form-group">
        <label class="form-label">服务器地址 (Base URL)</label>
        <input 
          v-model="config.base_url" 
          type="text" 
          placeholder="https://dav.example.com/dav/"
          class="form-input"
        />
      
      </div>
      
      <div class="form-group">
        <label class="form-label">远程路径 (Remote Path)</label>
        <input 
          v-model="config.remote_path" 
          type="text" 
          placeholder="bookshelf_backup/my-shelf"
          class="form-input"
        />
  
      </div>
      
      <div class="form-group">
        <label class="form-label">用户名（User Name）
        </label>
        <input 
          v-model="config.username" 
          type="text" 
          placeholder="username"
          class="form-input"
        />
      </div>
      
      <div class="form-group">
        <label class="form-label">密码（Password）</label>
        <input 
          v-model="config.password" 
          type="password" 
          placeholder="password"
          class="form-input"
        />
      </div>
      
      <!-- 操作按钮 -->
      <div class="actions">
        <button 
          class="btn btn-secondary" 
          @click="saveConfig"
        >
          保存配置
        </button>
        <button 
          class="btn btn-primary" 
          @click="handleTestConnection"
          :disabled="connecting"
        >
          {{ connecting ? '连接中...' : '测试连接' }}
        </button>
      </div>
      
      <!-- 同步操作 -->
      <div class="sync-section">
        <h4>同步操作</h4>
        <div class="actions">
          <button 
            class="btn btn-upload" 
            @click="handleUpload"
            :disabled="uploading"
          >
            {{ uploading ? '上传中...' : '上传' }}
          </button>
          <button 
            class="btn btn-download" 
            @click="handleDownload"
            :disabled="downloading"
          >
            {{ downloading ? '下载中...' : '下载' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* WebDAV容器 */
.webdav-wrapper {
  width: 100%;
  height: 100%;
  background: var(--sidebar-bg);
  color: var(--text-color);
  display: flex;
  flex-direction: column;
  user-select: none;
}

/* 侧边栏头部 */
.sidebar-header {
  padding: 28px 20px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(180deg, var(--primary-light) 0%, transparent 100%);
}

/* 返回按钮 */
.back-btn {
  background: var(--primary-light);
  border: 1px solid transparent;
  color: var(--primary-color);
  cursor: pointer;
  font-size: 1.1rem;
  padding: 8px 12px;
  border-radius: var(--radius-md);
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-btn:hover {
  background: var(--primary-color);
  color: white;
  transform: translateX(-3px);
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  letter-spacing: -0.02em;
}

/* WebDAV内容区 */
.webdav-content {
  flex: 1;
  padding: 24px 20px;
  overflow-y: auto;
}

/* 当前书架显示 */
.current-shelf {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: var(--bg-color);
  border-radius: var(--radius-lg);
  margin-bottom: 24px;
  font-size: 0.92rem;
  border: 1.5px solid var(--border-color);
}

.current-shelf .shelf-name {
  font-weight: 600;
  color: var(--primary-color);
}

/* 表单组 */
.form-group {
  margin-bottom: 10px;
}

/* 表单标签 */
.form-label {
  display: block;
  font-size: 0.89rem;
  color: var(--text-secondary);
  margin-bottom: 3px;
  font-weight: 550;
}

/* 表单输入框 */
.form-input {
  width: 100%;
  padding: 13px 16px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-lg);
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 0.92rem;
  box-sizing: border-box;
  outline: none;
  transition: all var(--transition-fast);
}

.form-input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px var(--primary-light);
}

.form-input::placeholder {
  color: var(--text-muted);
}

/* 表单提示 */
.form-hint {
  display: block;
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-top: 8px;
}

/* 操作按钮组 */
.actions {
  display: flex;
  gap: 14px;
  margin-top: 24px;
}

.btn {
  flex: 1;
  padding: 13px;
  border: none;
  border-radius: var(--radius-lg);
  cursor: pointer;
  font-size: 0.92rem;
  font-weight: 500;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.2) 50%, transparent 100%);
  transition: left var(--transition-slow);
}

.btn:hover::before {
  left: 100%;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
}

.btn-primary {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  box-shadow: var(--shadow-md);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.btn-secondary {
  background: transparent;
  color: var(--text-color);
  border: 1.5px solid var(--border-color);
}

.btn-secondary:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
}

/* 同步操作区 */
.sync-section {
  margin-top: 28px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color);
}

.sync-section h4 {
  margin: 0 0 16px 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.btn-upload {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  box-shadow: var(--shadow-md);
}

.btn-upload:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.btn-download {
  background: linear-gradient(135deg, var(--accent-color) 0%, var(--primary-color) 100%);
  color: white;
  box-shadow: var(--shadow-md);
}

.btn-download:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

/* 滚动条美化 */
.webdav-content::-webkit-scrollbar {
  width: 4px;
}

.webdav-content::-webkit-scrollbar-track {
  background: transparent;
}

.webdav-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.webdav-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>
