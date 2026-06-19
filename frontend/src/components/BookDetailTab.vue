<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import type { Book } from '../stores/library'
import { useLibraryStore } from '../stores/library'
import BookIcon from './icons/BookIcon.vue'
import EditIcon from './icons/EditIcon.vue'
import CheckIcon from './icons/CheckIcon.vue'
import XIcon from './icons/XIcon.vue'
import CustomModal from './CustomModal.vue'

const store = useLibraryStore()

// 弹窗状态
const showModal = ref(false)
const modalTitle = ref('')
const modalMessage = ref('')
const modalType = ref<'info' | 'warning' | 'success' | 'error'>('info')

const showAlert = (title: string, message: string, type: 'info' | 'warning' | 'success' | 'error' = 'info') => {
  modalTitle.value = title
  modalMessage.value = message
  modalType.value = type
  showModal.value = true
}

const handleModalConfirm = () => {
  showModal.value = false
}

interface Props {
  book: Book
}

const props = defineProps<Props>()

// 书籍详细信息
const bookDetails = ref({
  title: '',
  author: '未知',
  publisher: '未知',
  publishDate: '未知',
  updateDate: '未知',
  addDate: '未知',
  language: '未知',
  subjects: [] as string[],
  format: 'EPUB',
  fileSize: '0 KB',
  md5: '',
  description: '',
  localPath: ''
})

// 获取文件大小（字节转可读格式）
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化日期
const formatDate = (timestamp: number): string => {
  if (!timestamp) return '未知'
  const date = new Date(timestamp)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

// 编辑状态
const isEditing = ref(false)
const editingTitle = ref('')
const inputWidth = ref('auto') // 输入框宽度

// 开始编辑书名
const startEditTitle = () => {
  editingTitle.value = bookDetails.value.title
  isEditing.value = true
  
  // 计算书名文字的像素宽度
  nextTick(() => {
    const tempSpan = document.createElement('span')
    tempSpan.style.cssText = `
      font-size: 1.5rem;
      font-weight: 550;
      font-family: inherit;
      visibility: hidden;
      position: absolute;
      white-space: nowrap;
    `
    tempSpan.textContent = bookDetails.value.title
    document.body.appendChild(tempSpan)
    
    const textWidth = tempSpan.offsetWidth
    document.body.removeChild(tempSpan)
    
    // 输入框宽度 = 文字宽度 + 10px（按钮空间）
    inputWidth.value = `${textWidth + 10}px`
  })
}

// 取消编辑
const cancelEditTitle = () => {
  isEditing.value = false
  editingTitle.value = ''
}

// 保存书名
const saveTitle = async () => {
  if (!editingTitle.value.trim()) {
    showAlert('提示', '书名不能为空', 'warning')
    return
  }
  
  try {
    const newTitle = editingTitle.value.trim()
    
    // 1. 更新显示的书名
    bookDetails.value.title = newTitle
    
    // 2. 保存到 config.json
    if (props.book.filePath) {
      const lastSlash = Math.max(
        props.book.filePath.lastIndexOf('/'),
        props.book.filePath.lastIndexOf('\\')
      )
      const bookDir = props.book.filePath.substring(0, lastSlash)
      const configPath = `${bookDir}/config.json`
      
      // @ts-ignore
      const configExists = await window.go.main.App.FileExists(configPath)
      
      if (configExists) {
        // @ts-ignore
        const configContent = await window.go.main.App.ReadFile(configPath)
        const config = JSON.parse(configContent)
        
        console.log('保存前 config:', config)
        
        // 更新 title
        config.title = newTitle
        
        console.log('保存后 config:', config)
        
        // 保存回文件
        const configJson = JSON.stringify(config, null, 2)
        const configBytes = Array.from(new TextEncoder().encode(configJson))
        // @ts-ignore
        await window.go.main.App.SaveFile(bookDir, 'config.json', configBytes)
        
        console.log('书名已保存到 config.json:', newTitle)
        
        // 【关键】确保 description 没有丢失（从内存中的 config 对象读取）
        if (config.description) {
          bookDetails.value.description = config.description
          console.log('保持 description:', bookDetails.value.description)
        }
      }
    }
    
    // 3. 【关键】更新书架列表中的书名
    await store.updateBookTitle(props.book.id, newTitle)
    console.log('书架列表已更新')
    
    isEditing.value = false
    editingTitle.value = ''
  } catch (error) {
    console.error('保存书名失败:', error)
    showAlert('错误', '保存书名失败，请重试', 'error')
  }
}
onMounted(async () => {
  // 基本信息 - 优先使用 props.book 中的数据
  bookDetails.value.title = props.book.title || '未知书名'
  bookDetails.value.author = props.book.author || '未知'
  bookDetails.value.md5 = props.book.md5 || ''
  bookDetails.value.localPath = props.book.filePath || '未知路径'
  
  // 【关键】从 config.json 读取所有元数据（包括简介）
  try {
    if (props.book.filePath) {
      const lastSlash = Math.max(
        props.book.filePath.lastIndexOf('/'),
        props.book.filePath.lastIndexOf('\\')
      )
      const bookDir = props.book.filePath.substring(0, lastSlash)
      const configPath = `${bookDir}/config.json`
      
      // @ts-ignore
      const configExists = await window.go.main.App.FileExists(configPath)
      
      if (configExists) {
        // @ts-ignore
        const configContent = await window.go.main.App.ReadFile(configPath)
        const config = JSON.parse(configContent)
        
        // 加载简介
        bookDetails.value.description = config.description || ''
        console.log('从 config.json 加载 description:', bookDetails.value.description)
        
        // 加载其他元数据
        if (!props.book.author) {
          bookDetails.value.author = config.author || '未知'
        }
        bookDetails.value.publisher = config.publisher || '未知'
        bookDetails.value.publishDate = config.publishDate || '未知'
        bookDetails.value.language = config.language || '未知'
        bookDetails.value.subjects = config.subjects || []
        bookDetails.value.addDate = formatDate(config.createdAt)
      } else {
        bookDetails.value.addDate = formatDate(Date.now())
      }
    }
  } catch (error) {
    console.error('加载书籍详情失败:', error)
  }
  
  // 尝试获取文件大小（无论是否有 config.json）
  if (props.book.filePath) {
    try {
      // @ts-ignore
      const fileInfo = await window.go.main.App.GetFileInfo(props.book.filePath)
      if (fileInfo && fileInfo.size) {
        bookDetails.value.fileSize = formatFileSize(fileInfo.size)
      }
      if (fileInfo && fileInfo.modTime) {
        bookDetails.value.updateDate = formatDate(fileInfo.modTime)
      }
    } catch (error) {
      console.error('获取文件信息失败:', error)
    }
  }
})
</script>

<template>
  <div class="book-detail-container">
    <div class="detail-content">
      <!-- 主信息卡片：封面 + 书名 + 作者 -->
      <div class="main-info-card">
        <div class="cover-wrapper">
          <img 
            v-if="book.coverUrl" 
            :src="book.coverUrl" 
            alt="书籍封面" 
            class="cover-image"
          />
          <div v-else class="cover-placeholder">
            <BookIcon :size="48" />
          </div>
        </div>
        
        <div class="book-info">
          <div class="title-row">
            <h2 v-if="!isEditing" class="book-title">{{ bookDetails.title }}</h2>
            <input 
              v-else 
              v-model="editingTitle" 
              type="text" 
              class="title-input"
              :style="{ width: inputWidth }"
              @keyup.enter="saveTitle"
              @keyup.esc="cancelEditTitle"
              autofocus
            />
            <button 
              v-if="!isEditing" 
              class="edit-btn" 
              @click="startEditTitle"
              title="编辑书名"
            >
              <EditIcon :size="18" />
            </button>
            <div v-else class="edit-actions">
              <button class="action-btn save" @click="saveTitle" title="保存">
                <CheckIcon :size="18" />
              </button>
              <button class="action-btn cancel" @click="cancelEditTitle" title="取消">
                <XIcon :size="18" />
              </button>
            </div>
          </div>
          <div class="book-author">{{ bookDetails.author }}</div>
        </div>
      </div>
      
      <!-- 其他详细信息卡片 -->
      <div class="info-section">
        <div class="info-item">
          <span class="label">出版商</span>
          <span class="value">{{ bookDetails.publisher }}</span>
        </div>
        
        <div class="info-item">
          <span class="label">出版日期</span>
          <span class="value">{{ bookDetails.publishDate }}</span>
        </div>
        
        <div class="info-item">
          <span class="label">更新日期</span>
          <span class="value">{{ bookDetails.updateDate }}</span>
        </div>
        
        <div class="info-item">
          <span class="label">添加日期</span>
          <span class="value">{{ bookDetails.addDate }}</span>
        </div>
        
        <div class="info-item">
          <span class="label">语言</span>
          <span class="value">{{ bookDetails.language }}</span>
        </div>
        
        <div class="info-item" v-if="bookDetails.subjects.length > 0">
          <span class="label">主题</span>
          <span class="value">{{ bookDetails.subjects.join(', ') }}</span>
        </div>
        
        <div class="info-item">
          <span class="label">格式</span>
          <span class="value">{{ bookDetails.format }}</span>
        </div>
        
        <div class="info-item">
          <span class="label">文件大小</span>
          <span class="value">{{ bookDetails.fileSize }}</span>
        </div>
        
        <div class="info-item">
          <span class="label">MD5 标识</span>
          <span class="value md5">{{ bookDetails.md5 }}</span>
        </div>
        
        <div class="info-item full-row">
          <span class="label">本地路径</span>
          <span class="value path">{{ bookDetails.localPath }}</span>
        </div>
      </div>
      
      <!-- 简介 -->
      <div class="description-section" v-if="bookDetails.description && bookDetails.description.length > 0">
        <h3 class="section-title">简介</h3>
        <p class="description-text">{{ bookDetails.description }}</p>
      </div>
    </div>
  </div>

  <!-- 提示弹窗 -->
  <CustomModal
    :visible="showModal"
    :title="modalTitle"
    :message="modalMessage"
    :type="modalType"
    :showCancel="false"
    @confirm="handleModalConfirm"
  />
</template>

<style scoped>
/* 全局容器 */
.book-detail-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color);
  overflow-y: auto;
  padding: 24px;
  box-sizing: border-box;
}

/* 主内容流 */
.detail-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* --- 主信息卡片（按截图优化）--- */
.main-info-card {
  display: flex;
  align-items: flex-start;
  gap: 24px;
  background: var(--sidebar-bg);
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(var(--border-color-rgb), 0.3);
}

/* 封面尺寸缩小，和截图比例一致 */
.cover-wrapper {
  width: 140px;
  height: 200px;
  border-radius: 8px;
  overflow: hidden;
  background: var(--bg-color);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.35;
}

.book-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 8px;
}

/* 字体大小按截图优化 */
.book-title {
  font-size: 1.5rem;
  font-weight: 550;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
  letter-spacing: 0.3px;
}

.book-author {
  font-size: 1.1rem;
  color: var(--text-secondary);
  margin: 0;
}

/* 书名编辑区域 */
.title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-input {
  padding: 0;           /* 无内边距 */
  margin: 0;            /* 无外边距 */
  font-size: 1.5rem;    /* 与 h2 相同 */
  font-weight: 550;     /* 与 h2 相同 */
  font-family: inherit; /* 继承字体 */
  border: none;         /* 无边框 */
  border-bottom: 2px solid var(--primary-color); /* 只有底部线 */
  background: transparent; /* 透明背景 */
  line-height: 1.3;     /* 与 h2 相同 */
  letter-spacing: 0.3px;/* 与 h2 相同 */
  border-bottom-left-radius: 1px;
  border-bottom-right-radius: 1px;
}

.title-input:focus {
  border-color: var(--primary-color-dark, #007bff);
  box-shadow: 0 0 0 3px rgba(var(--primary-color-rgb, 0, 123, 255), 0);
}

.edit-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.edit-btn:hover {
  background: rgba(var(--primary-color-rgb, 0, 123, 255), 0.1);
  color: var(--primary-color);
}

.edit-actions {
  display: flex;
  gap: 6px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: none;
  border-radius: 6px;
  font-size: 1.2rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn.save {
  background: linear-gradient(135deg, var(--primary-color), var(--accent-color));
  color: white;
}

.action-btn.save:hover {
  filter: brightness(1.08);
}

.action-btn.cancel {
  background: var(--border-color);
  color: var(--text-secondary);
}

.action-btn.cancel:hover {
  background: var(--text-muted);
  color: var(--text-primary);
}

/* --- 信息卡片 --- */
.info-section {
  background: var(--sidebar-bg);
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px 20px;
  border: 1px solid rgba(var(--border-color-rgb), 0.3);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item.full-row {
  grid-column: 1 / -1;
}

.label {
  font-size: 0.8rem;
  color: var(--text-muted);
  font-weight: 500;
  letter-spacing: 0.2px;
}

.value {
  font-size: 0.95rem;
  color: var(--text-primary);
  word-break: break-word;
  line-height: 1.4;
}

/* MD5 等代码类文本 */
.value.md5 {
  font-family: 'JetBrains Mono', 'Consolas', monospace;
  font-size: 0.85rem;
  color: var(--primary-color);
  letter-spacing: 0.3px;
}

.value.path {
  font-family: 'JetBrains Mono', 'Consolas', monospace;
  font-size: 0.8rem;
  color: var(--text-secondary);
  word-break: break-all;
  padding: 6px 10px;
  background: rgba(var(--border-color-rgb), 0.2);
  border-radius: 6px;
  line-height: 1.5;
}

/* 简介卡片 */
.description-section {
  background: var(--sidebar-bg);
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(var(--border-color-rgb), 0.3);
}

.section-title {
  font-size: 1.1rem;
  font-weight: 550;
  color: var(--text-primary);
  margin: 0 0 12px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.description-text {
  font-size: 0.95rem;
  line-height: 1.8;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-wrap: break-word;
  text-align: justify;
}

/* 滚动条美化 */
.book-detail-container::-webkit-scrollbar {
  width: 6px;
}
.book-detail-container::-webkit-scrollbar-track {
  background: transparent;
}
.book-detail-container::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}
.book-detail-container::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

/* 响应式适配 */
@media (max-width: 768px) {
  .book-detail-container {
    padding: 16px;
  }

  .main-info-card {
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 16px;
  }

  .book-title {
    font-size: 1.4rem;
  }

  .book-author {
    font-size: 1.1rem;
  }

  .cover-wrapper {
    width: 120px;
    height: 170px;
  }

  .info-section {
    grid-template-columns: 1fr;
    padding: 16px;
    gap: 14px;
  }

  .description-section {
    padding: 16px;
  }
}
</style>