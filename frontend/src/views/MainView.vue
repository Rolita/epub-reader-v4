<script setup lang="ts">
import { ref, onMounted, watch, computed, onUnmounted } from 'vue'
import SidebarContainer from '../components/SidebarContainer.vue'
import Bookshelf from '../components/Bookshelf.vue'
import SettingsTab from '../components/SettingsTab.vue'
import ReaderContent from '../components/ReaderContent.vue'
import AddThemeTab from '../components/AddThemeTab.vue'
import BookDetailTab from '../components/BookDetailTab.vue'
import GroupDetail from '../components/GroupDetail.vue'
import CustomModal from '../components/CustomModal.vue'
import EmptyState from '../components/EmptyState.vue'
import { useLibraryStore } from '../stores/library'
import { useBookStore } from '../stores/book'
import { useThemeStore } from '../stores/theme'
import BookIcon from '../components/icons/BookIcon.vue'
import BookSVG from '../components/icons/BookSVG.vue'
import ListIcon from '../components/icons/ListIcon.vue'
import PaletteIcon from '../components/icons/PaletteIcon.vue'
import SparklesIcon from '../components/icons/SparklesIcon.vue'
import SettingsIcon from '../components/icons/SettingsIcon.vue'
import RefreshIcon from '../components/icons/RefreshIcon.vue'
import FolderIcon from '../components/icons/FolderIcon.vue'
import XIcon from '../components/icons/XIcon.vue'
import MinusIcon from '../components/icons/MinusIcon.vue'
import SquareIcon from '../components/icons/SquareIcon.vue'
import CloudIcon from '../components/icons/CloudIcon.vue'
import DownloadIcon from '../components/icons/DownloadIcon.vue'
import BookmarkIcon from '../components/icons/BookmarkIcon.vue'
import SaveIcon from '../components/icons/SaveIcon.vue'
import LayoutGridIcon from '../components/icons/LayoutGridIcon.vue'
import { handleRestoreProgress as restoreProgressAction, handleSaveProgress as saveProgressAction } from '../composables/useProgressButtons'

interface Tab {
  id: string
  name: string
  shelfId?: string
  groupId?: string
  type: 'bookshelf' | 'settings' | 'reader' | 'add-theme' | 'book-detail' | 'group-detail'
  bookId?: string
  filePath?: string
  editThemeId?: string
  bookData?: any
  groupData?: any
  icon?: 'book' | 'folder' | 'settings' | 'sparkles' | 'reader'
}

const store = useLibraryStore()
const bookStore = useBookStore()
const themeStore = useThemeStore()
const tabs = ref<Tab[]>([])

// 根据图标类型返回对应的图标组件
const getTabIcon = (icon?: string) => {
  switch (icon) {
    case 'book':
      return BookIcon
    case 'folder':
      return FolderIcon
    case 'settings':
      return SettingsIcon
    case 'sparkles':
      return SparklesIcon
    case 'reader':
      return BookSVG
    default:
      return BookIcon
  }
}
const activeTabId = ref<string>('')
const sidebarRef = ref<InstanceType<typeof SidebarContainer> | null>(null)
const readerRefs = new Map<string, InstanceType<typeof ReaderContent>>()
const activeSidebar = ref<string>('shelf')
const tabRefreshKeys = ref<Record<string, number>>({})

// 获取当前激活的阅读器组件
const getActiveReader = () => {
  if (!activeTabId.value) return null
  return readerRefs.get(activeTabId.value)
}

// 监听 tab 切换，当从阅读器切换走时自动保存进度，切换到阅读器时自动恢复进度
watch(activeTabId, (newTabId, oldTabId) => {
  // 从阅读器切换走时保存进度
  if (oldTabId) {
    const oldTab = tabs.value.find(t => t.id === oldTabId)
    if (oldTab?.type === 'reader') {
      const oldReader = readerRefs.get(oldTabId)
      oldReader?.saveProgress()
    }
  }
  
  // 切换到阅读器时恢复进度并更新目录侧边栏
  if (newTabId) {
    const newTab = tabs.value.find(t => t.id === newTabId)
    if (newTab?.type === 'reader') {
      const newReader = readerRefs.get(newTabId)
      newReader?.updateBookStore()
      handleRestoreProgress()
    }
  }
})

// WebDAV 下载状态
const isDownloading = ref(false)
const downloadStatus = ref('')
const downloadStatusType = ref<'success' | 'error' | ''>('')

// 全局弹窗状态（用于监听 store 发出的事件）
const globalModal = ref({
  visible: false,
  title: '',
  message: '',
  type: 'info' as 'info' | 'warning' | 'success' | 'error'
})

const handleGlobalModalConfirm = () => {
  globalModal.value.visible = false
}

// 监听全局 alert 事件
const handleShowAlert = (e: CustomEvent) => {
  globalModal.value = {
    visible: true,
    title: e.detail.title,
    message: e.detail.message,
    type: e.detail.type
  }
}

// 拖拽相关
const draggedTabId = ref<string | null>(null)
const dragOverTabId = ref<string | null>(null)

// 右键菜单相关
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuTabId = ref<string | null>(null)

// 处理目录跳转
const handleJump = (href: string) => {
  getActiveReader()?.jumpTo(href);
};

// 监听侧边栏选中变化
watch(() => store.activeShelfId, (newShelfId) => {
  if (newShelfId) {
    const shelf = store.shelves.find(s => s.id === newShelfId)
    if (shelf) {
      openBookshelfTab(newShelfId, shelf.name)
    }
  }
})

const switchSidebar = (viewName: string) => {
  activeSidebar.value = viewName
  sidebarRef.value?.switchView(viewName)
}

const openWebDav = () => {
  switchSidebar('webdav')
}

const toggleWebDav = () => {
  if (activeSidebar.value === 'webdav') {
    switchSidebar('shelf')
  } else {
    switchSidebar('webdav')
  }
}

// 主题保存后关闭标签并打开主题侧边栏
const handleThemeSaved = async (tabId: string) => {
  await closeTab(tabId)
  switchSidebar('theme')
}

// 刷新当前 tab 页面
const handleRefresh = () => {
  if (!activeTabId.value) return
  tabRefreshKeys.value[activeTabId.value] = (tabRefreshKeys.value[activeTabId.value] || 0) + 1
}

// 显示气泡提示（来自 WebDAV 侧边栏）
const handleShowToast = (message: string, type: 'success' | 'error') => {
  downloadStatus.value = message
  downloadStatusType.value = type
  setTimeout(() => {
    downloadStatus.value = ''
  }, 3000)
}

// 从云端下载（WebDAV）
const handleDownloadFromCloud = async () => {
  if (!store.activeShelfId) {
    downloadStatus.value = '请先选择一个书架'
    downloadStatusType.value = 'error'
    setTimeout(() => {
      downloadStatus.value = ''
    }, 3000)
    return
  }
  
  isDownloading.value = true
  downloadStatus.value = ''
  try {
    // @ts-ignore
    const res = await window.go.main.App.DownloadShelf(store.activeShelfId)
    downloadStatus.value = '下载成功: ' + res
    downloadStatusType.value = 'success'
    
    // 下载成功后刷新
    store.scanShelves()
    if (store.activeShelfId) {
      await store.loadShelfBooks(store.activeShelfId)
    }
    handleRefresh()
    
    // 1秒后自动隐藏气泡
    setTimeout(() => {
      downloadStatus.value = ''
    }, 1500)
  } catch (e) {
    downloadStatus.value = '下载失败: ' + (e as Error).message
    downloadStatusType.value = 'error'
    
    // 1秒后自动隐藏气泡
    setTimeout(() => {
      downloadStatus.value = ''
    }, 1500)
  } finally {
    isDownloading.value = false
  }
}

// 气泡通知辅助函数
const showToast = (message: string, type: 'success' | 'error') => {
  downloadStatus.value = message
  downloadStatusType.value = type
  const duration = type === 'success' ? 1500 : 3000
  setTimeout(() => { downloadStatus.value = '' }, duration)
}

// 恢复阅读进度
const handleRestoreProgress = async () => {
  await restoreProgressAction(activeTab.value, getActiveReader() || null, showToast)
}

// 保存阅读进度
const handleSaveProgress = async () => {
  await saveProgressAction(activeTab.value, getActiveReader() || null, showToast)
}

// 显示创建书架提示
const showCreateShelfHint = () => {
  // 切换到书架侧边栏，让用户看到创建按钮
  switchSidebar('shelf')
}

const openBookshelfTab = (shelfId: string, shelfName: string) => {
  // 检查是否已存在该 tab
  const existingTab = tabs.value.find(t => t.shelfId === shelfId)
  
  if (existingTab) {
    // 如果存在，激活该 tab
    activeTabId.value = existingTab.id
  } else {
    // 如果不存在，创建新 tab
    const newTab: Tab = {
      id: shelfId,
      name: shelfName,
      shelfId: shelfId,
      type: 'bookshelf',
      icon: 'book'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
  }
  // 切换到书架侧边栏
  switchSidebar('shelf')
}

// 处理侧边栏点击书架事件
const handleOpenShelf = (shelfId: string, shelfName: string) => {
  openBookshelfTab(shelfId, shelfName)
}

const openSettingsTab = () => {
  const existingTab = tabs.value.find(t => t.type === 'settings')
  
  if (existingTab) {
    activeTabId.value = existingTab.id
  } else {
    const newTab: Tab = {
      id: 'settings',
      name: '设置',
      type: 'settings',
      icon: 'settings'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
  }
}

const openAddThemeTab = () => {
  const existingTab = tabs.value.find(t => t.id === 'add-theme')
  
  if (existingTab) {
    activeTabId.value = existingTab.id
  } else {
    const newTab: Tab = {
      id: 'add-theme',
      name: '添加主题',
      type: 'add-theme',
      icon: 'sparkles'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
  }
  // 切换到主题侧边栏
  switchSidebar('theme')
}

const openEditThemeTab = (themeId: string) => {
  const tabId = 'edit-theme-' + themeId
  const existingTab = tabs.value.find(t => t.id === tabId)

  if (existingTab) {
    activeTabId.value = existingTab.id
  } else {
    const theme = themeStore.themes.find(t => t.id === themeId)
    const newTab: Tab = {
      id: tabId,
      name: theme ? `编辑: ${theme.name}` : '编辑主题',
      type: 'add-theme',
      editThemeId: themeId,
      icon: 'sparkles'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
  }
  switchSidebar('theme')
}

const openBookDetailTab = (book: any) => {
  const tabId = 'book-detail-' + book.id
  const existingTab = tabs.value.find(t => t.id === tabId)
  
  if (existingTab) {
    activeTabId.value = existingTab.id
  } else {
    const newTab: Tab = {
      id: tabId,
      name: book.title || '书籍详情',
      type: 'book-detail',
      bookId: book.id,
      bookData: book,
      icon: 'reader'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
  }
  // 切换到书架侧边栏
  switchSidebar('shelf')
}

const openGroupTab = (group: any) => {
  const tabId = 'group-' + group.id
  const existingTab = tabs.value.find(t => t.id === tabId)
  
  if (existingTab) {
    activeTabId.value = existingTab.id
  } else {
    const newTab: Tab = {
      id: tabId,
      name: group.name || '分组',
      type: 'group-detail',
      groupId: group.id,
      groupData: group,
      icon: 'folder'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
  }
  // 激活当前分组
  store.setActiveGroup(group.id)
}

const openReaderTab = async (book: any) => {
  const existingTab = tabs.value.find(t => t.type === 'reader' && t.bookId === book.id)
  
  if (existingTab) {
    activeTabId.value = existingTab.id
  } else {
    let filePath = book.filePath
    
    if (!filePath && book.md5) {
      const shelfDir = book.shelfId || '默认书架'
      // @ts-ignore
      const booksDir = await window.go.main.App.GetBooksDir()
      filePath = `${booksDir}/${shelfDir}/${book.md5}/${book.title}.epub`
    }
    
    const newTab: Tab = {
      id: `reader-${book.id}`,
      name: book.title,
      type: 'reader',
      bookId: book.id,
      filePath: filePath,
      icon: 'reader'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
  }
  
  // 自动切换到目录侧边栏
  switchSidebar('toc')
}

const closeTab = async (tabId: string) => {
  const index = tabs.value.findIndex(t => t.id === tabId)
  if (index !== -1) {
    const closingTab = tabs.value[index]
    const isActive = activeTabId.value === tabId
    // 关闭当前活跃的阅读器标签前保存进度
    if (isActive && closingTab.type === 'reader') {
      const reader = readerRefs.get(tabId)
      await reader?.saveProgress?.()
    }
    tabs.value.splice(index, 1)
    
    // 清理阅读器组件引用
    if (closingTab.type === 'reader') {
      readerRefs.delete(tabId)
    }
    
    // 如果关闭的是阅读器标签，清空目录
    if (closingTab.type === 'reader') {
      bookStore.clearActiveBook()
    }
    
    // 如果关闭的是分组标签，清空激活的分组
    if (closingTab.type === 'group-detail') {
      store.setActiveGroup(null)
    }
    
    // 如果关闭的是当前激活的 tab
    if (isActive) {
      // 选择相邻的 tab
      const newIndex = index < tabs.value.length ? index : index - 1
      if (newIndex >= 0 && tabs.value[newIndex]) {
        activeTabId.value = tabs.value[newIndex].id
        const newTab = tabs.value[newIndex]
        if (newTab.type === 'bookshelf' && newTab.shelfId) {
          store.setActiveShelf(newTab.shelfId)
          // 清空激活的分组
          store.setActiveGroup(null)
          // 自动切换到书架侧边栏
          switchSidebar('shelf')
        } else if (newTab.type === 'group-detail' && newTab.groupId) {
          store.setActiveGroup(newTab.groupId)
        }
      } else {
        activeTabId.value = ''
        store.setActiveGroup(null)
      }
    }
  }
}

const switchTab = (tabId: string) => {
  // 获取当前激活的tab（切换前）
  const currentTab = tabs.value.find(t => t.id === activeTabId.value)
  // 获取目标tab
  const targetTab = tabs.value.find(t => t.id === tabId)
  
  // 如果点击的是当前已激活的tab，不切换侧边栏
  if (activeTabId.value === tabId) {
    return
  }
  
  activeTabId.value = tabId
  
  if (targetTab) {
    if (targetTab.type === 'bookshelf' && targetTab.shelfId) {
      store.setActiveShelf(targetTab.shelfId)
      store.setActiveGroup(null)
    } else if (targetTab.type === 'group-detail' && targetTab.groupId) {
      store.setActiveGroup(targetTab.groupId)
    }
    
    // 如果是在两个书籍tab之间切换，保留当前侧边栏
    if (currentTab?.type === 'reader' && targetTab.type === 'reader') {
      return
    }
    
    // 根据tab类型切换对应的侧边栏
    switch (targetTab.type) {
      case 'bookshelf':
      case 'group-detail':
        switchSidebar('shelf')
        break
      case 'settings':
        switchSidebar('theme')
        break
      case 'add-theme':
        switchSidebar('theme')
        break
      case 'reader':
        switchSidebar('toc')
        break
      default:
        switchSidebar('shelf')
    }
  }
}

// ========== 快捷键功能 ==========

// 处理键盘事件
const handleKeyDown = async (e: KeyboardEvent) => {
  // Ctrl+W - 关闭当前标签页
  if (e.ctrlKey && e.key === 'w') {
    e.preventDefault()
    if (activeTabId.value) {
      await closeTab(activeTabId.value)
      console.log('快捷键: Ctrl+W - 关闭当前标签页')
    }
  }
}

// 监听键盘事件
onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
  window.addEventListener('show-alert', handleShowAlert as EventListener)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
  window.removeEventListener('show-alert', handleShowAlert as EventListener)
})

const activeTab = computed(() => {
  return tabs.value.find(t => t.id === activeTabId.value)
})

// 拖拽开始
const handleDragStart = (tabId: string) => {
  draggedTabId.value = tabId
}

// 拖拽经过
const handleDragOver = (e: DragEvent, tabId: string) => {
  e.preventDefault()
  if (tabId !== draggedTabId.value) {
    dragOverTabId.value = tabId
  }
}

// 拖拽离开
const handleDragLeave = () => {
  dragOverTabId.value = null
}

// 放置
const handleDrop = (e: DragEvent, targetTabId: string) => {
  e.preventDefault()
  if (draggedTabId.value && draggedTabId.value !== targetTabId) {
    const draggedIndex = tabs.value.findIndex(t => t.id === draggedTabId.value)
    const targetIndex = tabs.value.findIndex(t => t.id === targetTabId)
    
    if (draggedIndex !== -1 && targetIndex !== -1) {
      const [draggedTab] = tabs.value.splice(draggedIndex, 1)
      tabs.value.splice(targetIndex, 0, draggedTab)
    }
  }
  draggedTabId.value = null
  dragOverTabId.value = null
}

// 拖拽结束
const handleDragEnd = () => {
  draggedTabId.value = null
  dragOverTabId.value = null
}

// 右键菜单函数
const showContextMenu = (e: MouseEvent, tabId: string) => {
  e.preventDefault()
  contextMenuX.value = e.clientX
  contextMenuY.value = e.clientY
  contextMenuTabId.value = tabId
  contextMenuVisible.value = true
}

const hideContextMenu = () => {
  contextMenuVisible.value = false
  contextMenuTabId.value = null
}

const handleCloseTabWithContext = async () => {
  if (contextMenuTabId.value) {
    await closeTab(contextMenuTabId.value)
  }
  hideContextMenu()
}

const handleCloseOtherTabsWithContext = () => {
  if (!contextMenuTabId.value) return
  
  const currentTab = tabs.value.find(t => t.id === contextMenuTabId.value)
  if (currentTab) {
    tabs.value = [currentTab]
    activeTabId.value = currentTab.id
    
    // 如果是书架标签，切换到书架侧边栏
    if (currentTab.type === 'bookshelf' && currentTab.shelfId) {
      store.setActiveShelf(currentTab.shelfId)
      switchSidebar('shelf')
    }
  }
  hideContextMenu()
}

const handleCloseRightTabsWithContext = () => {
  if (!contextMenuTabId.value) return
  
  const currentIndex = tabs.value.findIndex(t => t.id === contextMenuTabId.value)
  if (currentIndex !== -1 && currentIndex < tabs.value.length - 1) {
    const closingTabs = tabs.value.slice(currentIndex + 1)
    
    // 检查是否要关闭阅读器或分组标签
    closingTabs.forEach(tab => {
      if (tab.type === 'reader') {
        bookStore.clearActiveBook()
      }
      if (tab.type === 'group-detail') {
        store.setActiveGroup(null)
      }
    })
    
    tabs.value = tabs.value.slice(0, currentIndex + 1)
    
    // 如果关闭的包含当前激活的tab，切换到当前tab
    if (closingTabs.some(t => t.id === activeTabId.value)) {
      activeTabId.value = contextMenuTabId.value
    }
  }
  hideContextMenu()
}

// 点击页面其他地方关闭右键菜单
const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (!target.closest('.tab-item') && !target.closest('.context-menu')) {
    hideContextMenu()
  }
}

// 窗口控制函数
const minimizeWindow = () => {
  const w = window as any
  if (w.runtime && w.runtime.WindowMinimise) {
    w.runtime.WindowMinimise()
  }
}

const toggleMaximize = () => {
  const w = window as any
  if (w.runtime && w.runtime.WindowToggleMaximise) {
    w.runtime.WindowToggleMaximise()
  }
}

const closeWindow = async () => {
  // 关闭前保存当前阅读进度
  const reader = getActiveReader()
  await reader?.saveProgress?.()
  const w = window as any
  // 使用 Quit() 代替 WindowClose()，这是最稳定的关闭方式
  if (w.runtime && w.runtime.Quit) {
    w.runtime.Quit()
  } else if (w.runtime && w.runtime.WindowClose) {
    w.runtime.WindowClose()
  }
}

onMounted(() => {
  store.scanShelves()
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('contextmenu', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('contextmenu', handleClickOutside)
})
</script>

<template>
  <div class="layout-wrapper">
    <!-- 功能栏 -->
    <aside class="function-bar">
      <div class="function-items">
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'shelf' }"
          @click="switchSidebar('shelf')" 
          title="书架"
        ><BookIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'toc' }"
          @click="switchSidebar('toc')" 
          title="目录"
        ><ListIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'layout' }"
          @click="switchSidebar('layout')" 
          title="排版"
        ><PaletteIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'theme' }"
          @click="switchSidebar('theme')" 
          title="主题"
        ><SparklesIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'bookshelf-layout' }"
          @click="switchSidebar('bookshelf-layout')" 
          title="书架布局"
        ><LayoutGridIcon :size="22" /></button>
        <button 
          class="func-btn" 
          @click="handleRefresh" 
          title="刷新"
        ><RefreshIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: isDownloading }"
          @click="handleDownloadFromCloud" 
          :disabled="isDownloading"
          title="从云端下载"
        ><DownloadIcon :size="22" /></button>
        <button 
          class="func-btn" 
          @click="handleRestoreProgress" 
          title="恢复阅读进度"
        ><BookmarkIcon :size="22" /></button>
        <button 
          class="func-btn" 
          @click="handleSaveProgress" 
          title="保存阅读进度"
        ><SaveIcon :size="22" /></button>
      </div>
      
      <div class="function-bottom">
        <button class="func-btn settings" @click="openSettingsTab" title="设置"><SettingsIcon :size="22" /></button>
      </div>
    </aside>
    
    <!-- 侧边栏容器 -->
    <SidebarContainer ref="sidebarRef" @jump="handleJump" @open-shelf="handleOpenShelf" @add-theme="openAddThemeTab" @edit-theme="openEditThemeTab" @sync-complete="handleRefresh" @show-toast="handleShowToast" />
    
    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 顶部栏（包含 tabs 和窗口控制） -->
      <header class="top-header" @dblclick.self="toggleMaximize">
        <!-- Tabs 容器 -->
        <div v-if="tabs.length > 0" class="tabs-container" @dblclick.self="toggleMaximize">
          <div 
            v-for="tab in tabs" 
            :key="tab.id"
            :class="['tab-item', { active: activeTabId === tab.id, dragging: draggedTabId === tab.id, 'drag-over': dragOverTabId === tab.id }]"
            :title="tab.name"
            @click="switchTab(tab.id)"
            @dblclick="closeTab(tab.id)"
            @contextmenu="(e) => showContextMenu(e, tab.id)"
            draggable="true"
            @dragstart="handleDragStart(tab.id)"
            @dragover="(e) => handleDragOver(e, tab.id)"
            @dragleave="handleDragLeave"
            @drop="(e) => handleDrop(e, tab.id)"
            @dragend="handleDragEnd"
          >
            <component 
              :is="getTabIcon(tab.icon)" 
              :size="14" 
              class="tab-icon"
            />
            <span class="tab-name">{{ tab.name }}</span>
            <button class="tab-close" @click.stop="closeTab(tab.id)"><XIcon :size="14" /></button>
          </div>
        </div>
        
        <!-- 窗口控制按钮（独立容器，始终固定在右侧） -->
        <div class="window-controls">
          <button class="window-btn minimize-btn" @click="minimizeWindow" title="最小化">
            <MinusIcon :size="14" />
          </button>
          <button class="window-btn maximize-btn" @click="toggleMaximize" title="最大化">
            <SquareIcon :size="14" />
          </button>
          <button class="window-btn close-btn" @click="closeWindow" title="关闭">
            <XIcon :size="14" />
          </button>
        </div>
        
        <!-- 右键菜单 -->
        <div 
          v-if="contextMenuVisible" 
          class="context-menu"
          :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
          @click.stop
        >
          <button class="context-menu-item" @click="handleCloseTabWithContext">关闭</button>
          <button class="context-menu-item" @click="handleCloseOtherTabsWithContext">关闭其他标签页</button>
          <button class="context-menu-item" @click="handleCloseRightTabsWithContext">关闭右侧标签页</button>
        </div>
      </header>
      
      <!-- 内容区域 -->
      <div v-if="activeTab" class="content-container">
        <!-- 书架内容 -->
        <Bookshelf 
          v-if="activeTab.type === 'bookshelf' && activeTab.shelfId"
          :shelf-id="activeTab.shelfId" 
          :key="'bookshelf-' + activeTab.id + '-' + (tabRefreshKeys[activeTab.id] || 0)"
          @open-book="openReaderTab"
          @toggle-webdav="toggleWebDav"
          @open-book-detail="openBookDetailTab"
          @open-group="openGroupTab"
          @close-tab-menu="hideContextMenu"
        />
        
        <!-- 分组详情内容 -->
        <GroupDetail 
          v-else-if="activeTab.type === 'group-detail' && activeTab.groupId"
          :key="'group-' + activeTab.id + '-' + (tabRefreshKeys[activeTab.id] || 0)"
          @open-book="openReaderTab"
          @toggle-webdav="toggleWebDav"
          @open-book-detail="openBookDetailTab"
          @close-tab-menu="hideContextMenu"
        />
        
        <!-- 设置内容 -->
        <SettingsTab 
          v-else-if="activeTab.type === 'settings'"
          :key="'settings-' + activeTab.id + '-' + (tabRefreshKeys[activeTab.id] || 0)"
        />
        
        <!-- 添加/编辑主题内容 -->
        <AddThemeTab 
          v-else-if="activeTab.type === 'add-theme'"
          :key="'add-theme-' + (activeTab.editThemeId || '') + '-' + (tabRefreshKeys[activeTab.id] || 0)"
          :edit-theme-id="activeTab.editThemeId"
          @saved="handleThemeSaved(activeTab.id)"
        />
        
        <!-- 书籍详情内容 -->
        <BookDetailTab
          v-else-if="activeTab.type === 'book-detail' && activeTab.bookData"
          :book="activeTab.bookData"
          :key="'book-detail-' + activeTab.id + '-' + (tabRefreshKeys[activeTab.id] || 0)"
        />
        
        <!-- 阅读器内容 - 使用独立容器，v-show 保持组件存活 -->
        <div v-show="activeTab?.type === 'reader'" class="reader-container">
          <template v-for="tab in tabs" :key="'reader-container-' + tab.id">
            <ReaderContent 
              v-if="tab.type === 'reader' && tab.filePath"
              v-show="activeTabId === tab.id"
              :file-path="tab.filePath"
              :ref="(el: any) => { 
                if (el) readerRefs.set(tab.id, el)
                else readerRefs.delete(tab.id)
              }"
            />
          </template>
        </div>
      </div>
      
      <!-- 下载状态气泡提示 -->
      <div 
        v-if="downloadStatus" 
        :class="['download-toast', downloadStatusType]"
      >
        {{ downloadStatus }}
      </div>
      
      <EmptyState v-if="!activeTab" @action="showCreateShelfHint" />
    </main>

    <!-- 全局提示弹窗 -->
    <CustomModal
      :visible="globalModal.visible"
      :title="globalModal.title"
      :message="globalModal.message"
      :type="globalModal.type"
      :showCancel="false"
      @confirm="handleGlobalModalConfirm"
    />
  </div>
</template>

<style scoped>
.layout-wrapper { display: flex; height: 100vh; width: 100vw; overflow: hidden; background-color: var(--bg-color); }

/* 功能栏 */
.function-bar {
  width: 60px;
  background: var(--sidebar-bg);
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: center;
  padding: 20px 0;
  border-right: 1px solid var(--border-color);
  position: relative;
  overflow: hidden;
}

.function-bar::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 80px;
  height: 80px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.06) 0%, transparent 70%);
  pointer-events: none;
}

.function-bar::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.08) 0%, transparent 100%);
  pointer-events: none;
}

.function-items {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.function-bottom {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.func-btn {
  width: 44px;
  height: 44px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 1.3rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  box-shadow: none;
  outline: none;
}

.func-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
  color: var(--text-primary);
}

.func-btn.settings:hover {
  background-color: rgba(0, 0, 0, 0.05);
  color: var(--text-primary);
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}

.main-content::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 120px;
  height: 80px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.06) 0%, transparent 70%);
  pointer-events: none;
}

.main-content::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.06) 0%, transparent 100%);
  pointer-events: none;
}

/* 下载状态气泡 */
.download-toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  padding: 14px 20px;
  border-radius: 10px;
  font-size: 0.9rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  max-width: 400px;
  animation: slideUp 0.3s ease;
}

.download-toast.success {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.95) 0%, rgba(6, 160, 110, 0.95) 100%);
  color: white;
}

.download-toast.error {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.95) 0%, rgba(200, 50, 50, 0.95) 100%);
  color: white;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 顶部栏（包含 tabs 和窗口控制） */
.top-header {
  display: flex;
  align-items: center;
  background: var(--bg-color);
  border-bottom: 1px solid var(--border-color);
  padding: 0 16px;
  min-height: 48px;
  /* Wails v2 官方拖拽 API - 整个顶部栏都是拖拽区 */
  --wails-draggable: drag;
}

/* Tabs 容器 */
.tabs-container {
  flex: 1; /* 占据剩余的所有空间 */
  display: flex;
  overflow: hidden;
  gap: 4px;
  min-width: 0; /* 允许收缩 */
}

/* 窗口控制按钮 */
.window-controls {
  display: flex;
  gap: 4px;
  margin-left: auto; /* 关键：自动推到最右端 */
  /* Wails v2 官方不可拖拽 API */
  --wails-draggable: no-drag;
}

.window-btn {
  width: 40px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 6px;
  transition: background-color 0.2s;
  color: var(--text-secondary);
}

.window-btn:hover {
  background-color: rgba(0, 0, 0, 0.1);
}

.close-btn:hover {
  background-color: #EF4444;
}

.close-btn:hover .window-icon {
  color: #FFFFFF;
}

/* Tab 项 */
.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 14px 16px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
  font-weight: 500;
  transition: all var(--transition-fast);
  white-space: nowrap;
  position: relative;
  user-select: none;
  /* Wails v2 官方不可拖拽 API - Tab 自身不参与窗口拖拽 */
  --wails-draggable: no-drag;
  /* 自适应收缩：优先根据内容宽度显示，空间不足时收缩 */
  flex-shrink: 1;
  flex-grow: 0;
  flex-basis: auto;
  min-width: 120px;
}

/* Tab 图标 */
.tab-icon {
  flex-shrink: 0;
  opacity: 0.7;
  transition: opacity var(--transition-fast);
}

.tab-item:hover .tab-icon,
.tab-item.active .tab-icon {
  opacity: 1;
}

.tab-item::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 15%;
  right: 15%;
  height: 3px;
  background: var(--primary-color);
  border-radius: 3px 3px 0 0;
  transform: scaleX(0);
  transition: transform var(--transition-normal);
}

.tab-item:hover {
  background: var(--sidebar-bg);
  color: var(--text-primary);
}

.tab-item.active {
  color: var(--primary-color);
  font-weight: 600;
}

.tab-item.active::after {
  transform: scaleX(1);
}

/* 拖拽状态 */
.tab-item.dragging {
  opacity: 0.4;
  cursor: grabbing;
  background: transparent;
}

.tab-item.drag-over::before {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--primary-color);
  border-radius: 3px 3px 0 0;
}

/* Tab 名称 */
.tab-name {
  letter-spacing: 0.01em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

/* Tab 关闭按钮 */
.tab-close {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  font-size: 0.85rem;
  padding: 3px;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  opacity: 0;
  transition: all var(--transition-fast);
  /* Wails v2 官方不可拖拽 API */
  --wails-draggable: no-drag;
}

.tab-item:hover .tab-close {
  opacity: 0.6;
}

.tab-close:hover {
  opacity: 1;
  background: rgba(239, 68, 68, 0.15);
  color: var(--danger-color);
}

.content-container {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.reader-container {
  flex: 1;
  width: 100%;
  height: 100%;
}

/* 右键菜单 */
.context-menu {
  position: fixed;
  z-index: 1000;
  min-width: 140px;
  padding: 4px 0;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.context-menu-item {
  display: block;
  width: 100%;
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 0.85rem;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
}

.context-menu-item:hover {
  background: var(--primary-light);
  color: var(--primary-color);
}

</style>
