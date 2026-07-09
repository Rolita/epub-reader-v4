<script setup lang="ts">
import { ref, onMounted, watch, computed, onUnmounted, nextTick } from 'vue'
import SidebarContainer from '../components/SidebarContainer.vue'
import NotesSidebar from '../components/NotesSidebar.vue'
import Bookshelf from '../components/Bookshelf.vue'
import SettingsTab from '../components/SettingsTab.vue'
import ReaderContent from '../components/ReaderContent.vue'
import AddThemeTab from '../components/AddThemeTab.vue'
import BookDetailTab from '../components/BookDetailTab.vue'
import GroupDetail from '../components/GroupDetail.vue'
import CustomModal from '../components/CustomModal.vue'
import EmptyState from '../components/EmptyState.vue'
import ShelfSelector from '../components/ShelfSelector.vue'
import { importBook } from '../utils/bookImporter'
import { useLibraryStore } from '../stores/library'
import { useBookStore } from '../stores/book'
import { useThemeStore } from '../stores/theme'
import { useSettingsStore } from '../stores/settings'
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
import NoteIcon from '../components/icons/NoteIcon.vue'
import SaveIcon from '../components/icons/SaveIcon.vue'
import LayoutGridIcon from '../components/icons/LayoutGridIcon.vue'
import IllustrationIcon from '../components/icons/IllustrationIcon.vue'
import TranslateIcon from '../components/icons/TranslateIcon.vue'
import SearchIcon from '../components/icons/SearchIcon.vue'
import SunIcon from '../components/icons/SunIcon.vue'
import MoonIcon from '../components/icons/MoonIcon.vue'
import { handleRestoreProgress as restoreProgressAction, handleSaveProgress as saveProgressAction } from '../composables/useProgressButtons'

interface Tab {
  id: string
  name: string
  shelfId?: string
  groupId?: string
  type: 'bookshelf' | 'settings' | 'reader' | 'add-theme' | 'book-detail' | 'group-detail'
  bookId?: string
  filePath?: string
  bookMd5?: string
  shelfName?: string
  editThemeId?: string
  bookData?: any
  groupData?: any
  icon?: 'book' | 'folder' | 'settings' | 'sparkles' | 'reader'
}

interface Pane {
  id: string
  activeTabId: string
  tabIds: string[]
}

const store = useLibraryStore()
const settingsStore = useSettingsStore()
const bookStore = useBookStore()
const themeStore = useThemeStore()

const isDarkTheme = computed(() => {
  const bgColor = themeStore.themeColors.bg;
  const r = parseInt(bgColor.slice(1, 3), 16);
  const g = parseInt(bgColor.slice(3, 5), 16);
  const b = parseInt(bgColor.slice(5, 7), 16);
  const brightness = (r * 299 + g * 587 + b * 114) / 1000;
  return brightness < 128;
})

const tabs = ref<Tab[]>([])
let checkPendingEpubInterval: number | null = null

// ===== 分屏布局状态 =====
const layout = ref<{
  mode: 'single' | 'split'
  panes: Pane[]
}>({
  mode: 'single',
  panes: [
    { id: 'pane-1', activeTabId: '', tabIds: [] },
    { id: 'pane-2', activeTabId: '', tabIds: [] }
  ]
})

// 获取 tab 所在的 pane
const getPaneByTabId = (tabId: string): Pane | null => {
  for (const pane of layout.value.panes) {
    if (pane.tabIds.includes(tabId)) return pane
  }
  return null
}

// 获取 pane 中的 tab 对象
const getPaneTabs = (pane: Pane): Tab[] => {
  return pane.tabIds.map(id => tabs.value.find(t => t.id === id)).filter(Boolean) as Tab[]
}

// 获取当前焦点的 pane（包含 activeTabId 的 pane）
const getFocusedPane = (): Pane | null => {
  return getPaneByTabId(activeTabId.value)
}

// 将 tab 添加到指定 pane
const addTabToPane = (tabId: string, paneId: string) => {
  const pane = layout.value.panes.find(p => p.id === paneId)
  if (!pane) return
  if (!pane.tabIds.includes(tabId)) {
    pane.tabIds.push(tabId)
  }
  pane.activeTabId = tabId
}

// 从 pane 中移除 tab
const removeTabFromPane = (tabId: string) => {
  const pane = getPaneByTabId(tabId)
  if (!pane) return
  const idx = pane.tabIds.indexOf(tabId)
  if (idx !== -1) {
    pane.tabIds.splice(idx, 1)
    // 如果移除的是 activeTab，切换到相邻 tab
    if (pane.activeTabId === tabId) {
      const newIdx = idx < pane.tabIds.length ? idx : idx - 1
      pane.activeTabId = newIdx >= 0 ? pane.tabIds[newIdx] : ''
    }
  }
}

// 移动 tab 到另一个 pane
const moveTabToPane = (tabId: string, targetPaneId: string) => {
  removeTabFromPane(tabId)
  addTabToPane(tabId, targetPaneId)
  // 如果分屏模式下某个 pane 变空了，自动退出分屏
  if (layout.value.mode === 'split') {
    const emptyPane = layout.value.panes.find(p => p.tabIds.length === 0)
    if (emptyPane) {
      exitSplitMode()
    }
  }
}

// 切换到分屏模式
const enterSplitMode = () => {
  if (layout.value.mode === 'split') return
  layout.value.mode = 'split'
  // 将当前 activeTabId 以外的 tabs 保留在 pane-1
  const currentPane = layout.value.panes[0]
  currentPane.tabIds = tabs.value.map(t => t.id)
  currentPane.activeTabId = activeTabId.value
}

// 退出分屏模式
const exitSplitMode = () => {
  if (layout.value.mode === 'single') return
  layout.value.mode = 'single'
  // 合并所有 tabs 到 pane-1
  const allTabIds = new Set<string>()
  layout.value.panes.forEach(p => p.tabIds.forEach(id => allTabIds.add(id)))
  layout.value.panes[0].tabIds = Array.from(allTabIds)
  layout.value.panes[0].activeTabId = activeTabId.value
  layout.value.panes[1].tabIds = []
  layout.value.panes[1].activeTabId = ''
  // 恢复当前激活tab对应的store状态
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab?.type === 'group-detail' && activeTab.groupId) {
    store.setActiveGroup(activeTab.groupId)
  } else if (activeTab?.type === 'bookshelf' && activeTab.shelfId) {
    store.setActiveShelf(activeTab.shelfId)
    store.setActiveGroup(null)
  }
  // 隐藏右键菜单
  hideContextMenu()
}

// 切换分屏/单屏
const toggleSplitMode = () => {
  if (layout.value.mode === 'single') {
    enterSplitMode()
  } else {
    exitSplitMode()
  }
}

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

const getActiveReaderFilePath = () => {
  return activeTab.value?.type === 'reader' ? activeTab.value?.filePath : undefined
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
      // 使用 nextTick 确保 ReaderContent 组件已挂载
      nextTick(() => {
        const newReader = readerRefs.get(newTabId)
        newReader?.updateBookStore()
        handleRestoreProgress()
        
        // 如果插图侧边栏正在显示，重新收集当前书籍的插图
        if (activeSidebar.value === 'illustration') {
          collectIllustrationsFromReader()
        }
      })
    }
  }
})

// WebDAV 下载状态
const isDownloading = ref(false)

interface ToastItem {
  id: number
  message: string
  type: 'success' | 'error'
}

const toastList = ref<ToastItem[]>([])
let toastIdCounter = 0

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
const dragOverPaneId = ref<string | null>(null)

// 右键菜单相关
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuTabId = ref<string | null>(null)

// 处理目录跳转
const handleJump = (payload: { href: string; cfi: string } | string) => {
  if (typeof payload === 'string') {
    getActiveReader()?.jumpTo(payload);
  } else {
    getActiveReader()?.jumpTo(payload.href, payload.cfi);
  }
};

// 处理图片预览
const handlePreview = (payload: { src: string; alt: string }) => {
  getActiveReader()?.openImagePreview(payload.src, payload.alt);
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

const switchSidebar = async (viewName: string) => {
  activeSidebar.value = viewName
  
  if (viewName === 'illustration') {
    await collectIllustrationsFromReader()
  }
  
  sidebarRef.value?.switchView(viewName)
}

const searchInBook = async (keyword: string) => {
  const reader = getActiveReader()
  if (!reader || !reader.searchInBook) return []
  return await reader.searchInBook(keyword)
}

const highlightSearchKeyword = (keyword: string) => {
  const reader = getActiveReader()
  if (!reader || !reader.highlightSearchKeyword) return
  reader.highlightSearchKeyword(keyword)
}

const clearSearchHighlight = () => {
  const reader = getActiveReader()
  if (!reader || !reader.clearHighlight) return
  reader.clearHighlight()
}

// 从阅读器收集插图
const collectIllustrationsFromReader = async () => {
  console.log('[插图收集] activeTabId:', activeTabId.value)
  console.log('[插图收集] readerRefs 大小:', readerRefs.size)
  
  let reader = getActiveReader()
  
  if (!reader) {
    // 尝试从所有阅读器中获取第一个
    for (const [tabId, ref] of readerRefs) {
      const tab = tabs.value.find(t => t.id === tabId)
      if (tab?.type === 'reader') {
        reader = ref
        console.log('[插图收集] 从所有阅读器中找到:', tabId)
        break
      }
    }
  }
  
  if (!reader) {
    console.warn('[插图收集] 没有找到激活的阅读器')
    return
  }
  
  try {
    const illustrations = await reader.getIllustrations?.()
    console.log('[插图收集] 从阅读器获取到', illustrations?.length || 0, '张图片')
    
    settingsStore.setIllustrations(illustrations || [])
  } catch (err) {
    console.error('[插图收集] 获取插图失败:', err)
  }
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
const handleRefresh = async () => {
  await handleSaveProgress()
  
  if (layout.value.mode === 'single') {
    if (!activeTabId.value) return
    const tab = tabs.value.find(t => t.id === activeTabId.value)
    if (tab?.type === 'reader') {
      const reader = readerRefs.get(activeTabId.value)
      if (reader) reader.refresh()
    } else {
      tabRefreshKeys.value[activeTabId.value] = (tabRefreshKeys.value[activeTabId.value] || 0) + 1
    }
  } else {
    layout.value.panes.forEach(pane => {
      if (pane.activeTabId) {
        const tab = tabs.value.find(t => t.id === pane.activeTabId)
        if (tab?.type === 'reader') {
          const reader = readerRefs.get(pane.activeTabId)
          if (reader) reader.refresh()
        } else {
          tabRefreshKeys.value[pane.activeTabId] = (tabRefreshKeys.value[pane.activeTabId] || 0) + 1
        }
      }
    })
  }
}

// 显示气泡提示（来自 WebDAV 侧边栏）
const handleShowToast = (message: string, type: 'success' | 'error') => {
  showToast(message, type)
}

// 从云端下载（WebDAV）
const handleDownloadFromCloud = async () => {
  if (!store.activeShelfId) {
    showToast('请先选择一个书架', 'error')
    return
  }
  
  isDownloading.value = true
  try {
    // @ts-ignore
    const res = await window.go.main.App.DownloadShelf(store.activeShelfId)
    showToast('下载成功: ' + res, 'success')
    
    // 下载成功后刷新
    store.scanShelves()
    if (store.activeShelfId) {
      await store.loadShelfBooks(store.activeShelfId)
    }
    handleRefresh()
  } catch (e) {
    showToast('下载失败: ' + (e as Error).message, 'error')
  } finally {
    isDownloading.value = false
  }
}

// 气泡通知辅助函数
const showToast = (message: string, type: 'success' | 'error') => {
  const id = ++toastIdCounter
  const toastItem: ToastItem = { id, message, type }
  
  toastList.value.push(toastItem)
  
  const duration = type === 'success' ? 5000 : 5000
  setTimeout(() => {
    const index = toastList.value.findIndex(t => t.id === id)
    if (index !== -1) {
      toastList.value.splice(index, 1)
    }
  }, duration)
}

// 恢复阅读进度
const handleRestoreProgress = async () => {
  await restoreProgressAction(activeTab.value, getActiveReader() || null, showToast)
}

// 书籍初始化完成后延时恢复进度
const handleReaderReady = () => {
  setTimeout(() => {
    handleRestoreProgress()
  }, 500)
}

const handleBookmarkSaved = () => {
  sidebarRef.value?.refreshBookmarks()
}

// 保存阅读进度
const handleSaveProgress = async () => {
  if (layout.value.mode === 'single') {
    if (!activeTab.value || activeTab.value.type !== 'reader') return
    const reader = getActiveReader()
    if (!reader) return
    await saveProgressAction(activeTab.value, reader, showToast)
  } else {
    let savedCount = 0
    for (const pane of layout.value.panes) {
      if (pane.activeTabId) {
        const tab = tabs.value.find(t => t.id === pane.activeTabId)
        const reader = readerRefs.get(pane.activeTabId)
        if (tab?.type === 'reader' && reader) {
          try {
            const result = await reader.saveProgress?.()
            if (result) savedCount++
          } catch (e) {
            console.error('保存进度失败:', e)
          }
        }
      }
    }
    if (savedCount > 0) {
      showToast(`已保存 ${savedCount} 个阅读进度`, 'success')
    }
  }
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
    // 确保在正确的 pane 中
    const pane = getPaneByTabId(existingTab.id) || layout.value.panes[0]
    pane.activeTabId = existingTab.id
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
    // 添加到当前焦点的 pane（单屏时就是 pane-1）
    const focusedPane = getFocusedPane() || layout.value.panes[0]
    addTabToPane(newTab.id, focusedPane.id)
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
    const pane = getPaneByTabId(existingTab.id) || layout.value.panes[0]
    pane.activeTabId = existingTab.id
  } else {
    const newTab: Tab = {
      id: 'settings',
      name: '设置',
      type: 'settings',
      icon: 'settings'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
    const focusedPane = getFocusedPane() || layout.value.panes[0]
    addTabToPane(newTab.id, focusedPane.id)
  }
}

const openAddThemeTab = () => {
  const existingTab = tabs.value.find(t => t.id === 'add-theme')
  
  if (existingTab) {
    activeTabId.value = existingTab.id
    const pane = getPaneByTabId(existingTab.id) || layout.value.panes[0]
    pane.activeTabId = existingTab.id
  } else {
    const newTab: Tab = {
      id: 'add-theme',
      name: '添加主题',
      type: 'add-theme',
      icon: 'sparkles'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
    const focusedPane = getFocusedPane() || layout.value.panes[0]
    addTabToPane(newTab.id, focusedPane.id)
  }
  // 切换到主题侧边栏
  switchSidebar('theme')
}

const openEditThemeTab = (themeId: string) => {
  const tabId = 'edit-theme-' + themeId
  const existingTab = tabs.value.find(t => t.id === tabId)

  if (existingTab) {
    activeTabId.value = existingTab.id
    const pane = getPaneByTabId(existingTab.id) || layout.value.panes[0]
    pane.activeTabId = existingTab.id
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
    const focusedPane = getFocusedPane() || layout.value.panes[0]
    addTabToPane(newTab.id, focusedPane.id)
  }
  switchSidebar('theme')
}

const openBookDetailTab = (book: any) => {
  const tabId = 'book-detail-' + book.id
  const existingTab = tabs.value.find(t => t.id === tabId)
  
  if (existingTab) {
    activeTabId.value = existingTab.id
    const pane = getPaneByTabId(existingTab.id) || layout.value.panes[0]
    pane.activeTabId = existingTab.id
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
    const focusedPane = getFocusedPane() || layout.value.panes[0]
    addTabToPane(newTab.id, focusedPane.id)
  }
  // 切换到书架侧边栏
  switchSidebar('shelf')
}

const openGroupTab = (group: any, paneId?: string) => {
  const tabId = 'group-' + group.id
  const existingTab = tabs.value.find(t => t.id === tabId)
  
  if (existingTab) {
    activeTabId.value = existingTab.id
    const pane = getPaneByTabId(existingTab.id) || layout.value.panes[0]
    pane.activeTabId = existingTab.id
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
    const targetPaneId = paneId || getFocusedPane()?.id || layout.value.panes[0].id
    addTabToPane(newTab.id, targetPaneId)
  }
  // 激活当前分组
  store.setActiveGroup(group.id)
}

const openReaderTab = async (book: any, targetPaneId?: string) => {
  // 更新书籍最后阅读时间
  store.updateBookLastReadTime(book.id)
  
  const existingTab = tabs.value.find(t => t.type === 'reader' && t.bookId === book.id)
  
  if (existingTab) {
    activeTabId.value = existingTab.id
    const pane = getPaneByTabId(existingTab.id) || layout.value.panes[0]
    pane.activeTabId = existingTab.id
 
  } else {
    let filePath = book.filePath
    
    if (book.md5 && book.shelfId) {
      try {
        // @ts-ignore
        const localPath = await window.go.main.App.GetBookLocalPath(book.shelfId, book.md5)
        if (localPath) {
          filePath = localPath
        }
      } catch (error) {
        console.error('获取书籍本地路径失败:', error)
      }
    } else if (!filePath && book.md5) {
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
      bookMd5: book.md5,
      shelfName: book.shelfId,
      icon: 'reader'
    }
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
    const paneId = targetPaneId || getFocusedPane()?.id || layout.value.panes[0].id
    addTabToPane(newTab.id, paneId)
    
 
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
    
    // 从 pane 中移除
    removeTabFromPane(tabId)
    
    // 清理阅读器组件引用
    if (closingTab.type === 'reader') {
      readerRefs.delete(tabId)
    }
    
    // 如果关闭的是阅读器标签，注销 EPUB 文件并释放文件句柄
    if (closingTab.type === 'reader') {
      // @ts-ignore
      await window.go.main.App.UnregisterEpubTab(tabId)
      bookStore.clearActiveBook()
      settingsStore.clearIllustrations()
      await store.loadBooksProgress()
    }
    
    // 如果关闭的是分组标签，清空激活的分组
    if (closingTab.type === 'group-detail') {
      store.setActiveGroup(null)
    }
    
    // 如果关闭的是当前激活的 tab
    if (isActive) {
      // 从当前 pane 中选择相邻的 tab
      const focusedPane = getFocusedPane()
      if (focusedPane && focusedPane.activeTabId) {
        activeTabId.value = focusedPane.activeTabId
        const newTab = tabs.value.find(t => t.id === focusedPane.activeTabId)
        if (newTab) {
          if (newTab.type === 'bookshelf' && newTab.shelfId) {
            store.setActiveShelf(newTab.shelfId)
            store.setActiveGroup(null)
            switchSidebar('shelf')
          } else if (newTab.type === 'group-detail' && newTab.groupId) {
            store.setActiveGroup(newTab.groupId)
          }
        }
      } else {
        // 如果当前 pane 没有 tab 了，尝试从另一个 pane 找
        const otherPane = layout.value.panes.find(p => p.id !== focusedPane?.id && p.activeTabId)
        if (otherPane) {
          activeTabId.value = otherPane.activeTabId
          const newTab = tabs.value.find(t => t.id === otherPane.activeTabId)
          if (newTab) {
            if (newTab.type === 'bookshelf' && newTab.shelfId) {
              store.setActiveShelf(newTab.shelfId)
              store.setActiveGroup(null)
              switchSidebar('shelf')
            } else if (newTab.type === 'group-detail' && newTab.groupId) {
              store.setActiveGroup(newTab.groupId)
            }
          }
        } else {
          activeTabId.value = ''
          store.setActiveGroup(null)
        }
      }
    }
    
    // 如果分屏模式下某个 pane 空了，退出分屏
    if (layout.value.mode === 'split') {
      const emptyPane = layout.value.panes.find(p => p.tabIds.length === 0)
      if (emptyPane) {
        exitSplitMode()
      }
    }
  }
}

const switchTab = (tabId: string, paneId?: string) => {
  // 获取当前激活的tab（切换前）
  const currentTab = tabs.value.find(t => t.id === activeTabId.value)
  // 获取目标tab
  const targetTab = tabs.value.find(t => t.id === tabId)
  
  // 如果点击的是当前已激活的tab，只更新 pane 的 activeTabId
  if (activeTabId.value === tabId) {
    if (paneId) {
      const pane = layout.value.panes.find(p => p.id === paneId)
      if (pane) pane.activeTabId = tabId
    }
    return
  }
  
  activeTabId.value = tabId
  
  // 更新 pane 的 activeTabId
  if (paneId) {
    const pane = layout.value.panes.find(p => p.id === paneId)
    if (pane) pane.activeTabId = tabId
  } else {
    const pane = getPaneByTabId(tabId)
    if (pane) pane.activeTabId = tabId
  }
  
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

const handlePaneClick = (paneId: string) => {
  const pane = layout.value.panes.find(p => p.id === paneId)
  if (!pane || !pane.activeTabId) return
  
  const tab = tabs.value.find(t => t.id === pane.activeTabId)
  if (tab && tab.type !== 'reader') {
    switchTab(tab.id, paneId)
  }
}

// ========== 快捷键功能 ==========

// 处理窗口大小改变
let resizeTimer: ReturnType<typeof setTimeout> | null = null
const handleResize = () => {
  if (resizeTimer) clearTimeout(resizeTimer)
  resizeTimer = setTimeout(() => {
    handleRefresh()
  }, 300)
}

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
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
  window.removeEventListener('show-alert', handleShowAlert as EventListener)
  window.removeEventListener('resize', handleResize)
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
const handleDrop = (e: DragEvent, targetTabId: string, paneId?: string) => {
  e.preventDefault()
  e.stopPropagation()
  if (draggedTabId.value && draggedTabId.value !== targetTabId) {
    const draggedPane = getPaneByTabId(draggedTabId.value)
    const targetPane = getPaneByTabId(targetTabId)

    if (draggedPane && targetPane && draggedPane.id === targetPane.id) {
      // 同一个 pane 内拖拽：调整 pane.tabIds 顺序
      const draggedIdx = draggedPane.tabIds.indexOf(draggedTabId.value)
      const targetIdx = targetPane.tabIds.indexOf(targetTabId)
      if (draggedIdx !== -1 && targetIdx !== -1) {
        const [id] = draggedPane.tabIds.splice(draggedIdx, 1)
        draggedPane.tabIds.splice(targetIdx, 0, id)
      }
    } else if (draggedPane && targetPane) {
      // 跨 pane 拖拽：移动 tab 到目标 pane
      moveTabToPane(draggedTabId.value, targetPane.id)
    }
  }
  draggedTabId.value = null
  dragOverTabId.value = null
  dragOverPaneId.value = null
}

// 拖拽经过 pane 容器
const handlePaneDragOver = (e: DragEvent, paneId: string) => {
  e.preventDefault()
  dragOverPaneId.value = paneId
}

// 拖拽离开 pane 容器
const handlePaneDragLeave = () => {
  dragOverPaneId.value = null
}

// 放置到 pane 容器
const handlePaneDrop = (e: DragEvent, paneId: string) => {
  e.preventDefault()
  if (draggedTabId.value) {
    const draggedPane = getPaneByTabId(draggedTabId.value)
    if (draggedPane && draggedPane.id !== paneId) {
      // 跨 pane 拖拽：移动 tab 到目标 pane
      moveTabToPane(draggedTabId.value, paneId)
    }
  }
  draggedTabId.value = null
  dragOverTabId.value = null
  dragOverPaneId.value = null
}

// 拖拽结束
const handleDragEnd = () => {
  draggedTabId.value = null
  dragOverTabId.value = null
  dragOverPaneId.value = null
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
  
  const currentPane = getPaneByTabId(contextMenuTabId.value)
  if (!currentPane) return
  
  // 只关闭当前 pane 中的其他 tabs
  const tabsToClose = currentPane.tabIds.filter(id => id !== contextMenuTabId.value)
  tabsToClose.forEach(id => {
    const tab = tabs.value.find(t => t.id === id)
    if (tab) {
      if (tab.type === 'reader') {
        readerRefs.delete(id)
        bookStore.clearActiveBook()
      }
      if (tab.type === 'group-detail') {
        store.setActiveGroup(null)
      }
    }
    const globalIdx = tabs.value.findIndex(t => t.id === id)
    if (globalIdx !== -1) tabs.value.splice(globalIdx, 1)
  })
  
  currentPane.tabIds = [contextMenuTabId.value]
  currentPane.activeTabId = contextMenuTabId.value
  activeTabId.value = contextMenuTabId.value
  
  const currentTab = tabs.value.find(t => t.id === contextMenuTabId.value)
  if (currentTab?.type === 'bookshelf' && currentTab.shelfId) {
    store.setActiveShelf(currentTab.shelfId)
    switchSidebar('shelf')
  }
  hideContextMenu()
}

const handleCloseRightTabsWithContext = () => {
  if (!contextMenuTabId.value) return
  
  const currentPane = getPaneByTabId(contextMenuTabId.value)
  if (!currentPane) return
  
  const currentIndex = currentPane.tabIds.indexOf(contextMenuTabId.value)
  if (currentIndex !== -1 && currentIndex < currentPane.tabIds.length - 1) {
    const closingIds = currentPane.tabIds.slice(currentIndex + 1)
    
    closingIds.forEach(id => {
      const tab = tabs.value.find(t => t.id === id)
      if (tab) {
        if (tab.type === 'reader') {
          readerRefs.delete(id)
          bookStore.clearActiveBook()
        }
        if (tab.type === 'group-detail') {
          store.setActiveGroup(null)
        }
      }
      const globalIdx = tabs.value.findIndex(t => t.id === id)
      if (globalIdx !== -1) tabs.value.splice(globalIdx, 1)
    })
    
    currentPane.tabIds = currentPane.tabIds.slice(0, currentIndex + 1)
    
    if (closingIds.some(id => id === activeTabId.value)) {
      activeTabId.value = contextMenuTabId.value
      currentPane.activeTabId = contextMenuTabId.value
    }
  }
  hideContextMenu()
}

// 在分屏中打开
const handleOpenInSplit = () => {
  if (!contextMenuTabId.value) return
  
  const tabId = contextMenuTabId.value
  const currentPane = getPaneByTabId(tabId)
  if (!currentPane) return
  
  // 如果当前是单屏，先进入分屏模式
  if (layout.value.mode === 'single') {
    enterSplitMode()
  }
  
  // 找到另一个 pane
  const otherPane = layout.value.panes.find(p => p.id !== currentPane.id)
  if (!otherPane) return
  
  // 移动 tab 到另一个 pane
  moveTabToPane(tabId, otherPane.id)
  activeTabId.value = tabId
  otherPane.activeTabId = tabId
  
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
  if (layout.value.mode === 'split') {
    setTimeout(() => {
      handleRefresh()
    }, 300)
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

const shelfSelectorVisible = ref(false)
const pendingEpubPath = ref('')

const handleOpenPendingEpub = async () => {
  await store.scanShelves()
  
  const path = pendingEpubPath.value
  if (!path) return
  
  if (store.shelves.length === 1) {
    await importAndOpenBook(path, store.shelves[0].id)
    return
  }
  
  if (store.shelves.length > 1) {
    const bookMd5 = await getFileMd5(path)
    if (bookMd5) {
      const bookInActiveShelf = store.currentBooks.find((b: any) => b.md5 === bookMd5)
      if (bookInActiveShelf) {
        openReaderTab(bookInActiveShelf)
        showToast('书籍已存在，直接打开', 'success')
        pendingEpubPath.value = ''
        return
      }
    }
    
    shelfSelectorVisible.value = true
  }
}

const getFileMd5 = async (filePath: string): Promise<string> => {
  try {
    // @ts-ignore
    const md5 = await window.go.main.App.CalculateFileMD5(filePath)
    return md5 || ''
  } catch (e) {
    console.error('获取文件MD5失败:', e)
  }
  return ''
}

const handleShelfSelect = async (shelfId: string, epubPath: string) => {
  shelfSelectorVisible.value = false
  await importAndOpenBook(epubPath, shelfId)
}

const handleShelfSelectorClose = () => {
  shelfSelectorVisible.value = false
  pendingEpubPath.value = ''
}

const handleImportOnly = async (shelfId: string, epubPath: string) => {
  shelfSelectorVisible.value = false
  try {
    showToast('正在导入书籍...', 'success')
    
    const result = await importBook(epubPath, shelfId)
    
    if (result.success) {
      await store.addBook(
        result.title,
        result.coverUrl,
        result.md5,
        result.filePath,
        result.author
      )
      showToast('书籍导入成功', 'success')
    } else {
      showToast('导入失败: ' + (result.error || '未知错误'), 'error')
    }
  } catch (error) {
    showToast('导入失败: ' + (error as Error).message, 'error')
  } finally {
    pendingEpubPath.value = ''
  }
}

const importAndOpenBook = async (filePath: string, shelfId: string) => {
  try {
    showToast('正在导入书籍...', 'success')
    
    await store.setActiveShelf(shelfId)
    
    const result = await importBook(filePath, shelfId)
    
    if (result.success) {
      const isAdded = await store.addBook(
        result.title,
        result.coverUrl,
        result.md5,
        result.filePath,
        result.author
      )
      
      if (isAdded) {
        const book = store.currentBooks.find((b: any) => b.md5 === result.md5)
        if (book) {
          openReaderTab(book)
          showToast('书籍导入成功', 'success')
        }
      } else {
        const book = store.currentBooks.find((b: any) => b.md5 === result.md5)
        if (book) {
          openReaderTab(book)
          showToast('书籍已存在，直接打开', 'success')
        }
      }
    } else {
      showToast('导入失败: ' + (result.error || '未知错误'), 'error')
    }
  } catch (error) {
    showToast('导入失败: ' + (error as Error).message, 'error')
  } finally {
    pendingEpubPath.value = ''
  }
}

onMounted(() => {
  store.scanShelves()
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('contextmenu', handleClickOutside)
  
  ;(async () => {
    await nextTick()
    try {
      // @ts-ignore
      const path = await window.go.main.App.GetPendingEpubPath()
      if (path) {
        pendingEpubPath.value = path
        await handleOpenPendingEpub()
      }
    } catch (e) {
      console.error('获取待打开 EPUB 路径失败:', e)
    }
  })()
  
  checkPendingEpubInterval = window.setInterval(async () => {
    try {
      // @ts-ignore
      const path = await window.go.main.App.GetPendingEpubPath()
      if (path) {
        pendingEpubPath.value = path
        await handleOpenPendingEpub()
      }
    } catch (e) {
      // ignore
    }
  }, 1000)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('contextmenu', handleClickOutside)
  if (checkPendingEpubInterval !== null) {
    clearInterval(checkPendingEpubInterval)
  }
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
        >
          <SunIcon v-if="!isDarkTheme" :size="22" />
          <MoonIcon v-else :size="22" />
        </button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'bookshelf-layout' }"
          @click="switchSidebar('bookshelf-layout')" 
          title="书架布局"
        ><LayoutGridIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'illustration' }"
          @click="switchSidebar('illustration')" 
          title="插画"
        ><IllustrationIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'bookmarks' }"
          @click="switchSidebar('bookmarks')" 
          title="书签"
        ><BookmarkIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'translate' }"
          @click="switchSidebar('translate')" 
          title="翻译"
        ><TranslateIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'search' }"
          @click="switchSidebar('search')" 
          title="搜索"
        ><SearchIcon :size="22" /></button>
        <button 
          class="func-btn" 
          :class="{ active: activeSidebar === 'notes' }"
          @click="switchSidebar('notes')" 
          title="笔记"
        ><NoteIcon :size="22" /></button>
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
          v-if="false"
          class="func-btn" 
          @click="handleRestoreProgress" 
          title="恢复阅读进度"
        ><BookmarkIcon :size="22" /></button>
        <button 
          v-if="false"
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
    <SidebarContainer
      ref="sidebarRef"
      :has-active-book="activeTab?.type === 'reader'"
      :book-title="activeTab?.type === 'reader' ? activeTab?.name : ''"
      :file-path="activeTab?.type === 'reader' ? getActiveReaderFilePath() : undefined"
      :search-in-book="searchInBook"
      :highlight-search-keyword="highlightSearchKeyword"
      :clear-search-highlight="clearSearchHighlight"
      @jump="handleJump"
      @preview="handlePreview"
      @open-shelf="handleOpenShelf"
      @add-theme="openAddThemeTab"
      @edit-theme="openEditThemeTab"
      @sync-complete="handleRefresh"
      @show-toast="handleShowToast"
    />
    
    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 顶部栏（包含 tabs 和窗口控制） -->
      <header class="top-header" @dblclick.self="toggleMaximize">
        <!-- 单屏模式：一个 Tabs 容器 -->
        <template v-if="layout.mode === 'single'">
          <div v-if="layout.panes[0].tabIds.length > 0" class="tabs-container" @dblclick.self="toggleMaximize">
            <div 
              v-for="tabId in layout.panes[0].tabIds" 
              :key="tabId"
              :class="['tab-item', { active: activeTabId === tabId, dragging: draggedTabId === tabId, 'drag-over': dragOverTabId === tabId }]"
              :title="tabs.find(t => t.id === tabId)?.name"
              @click="switchTab(tabId)"
              @dblclick="closeTab(tabId)"
              @contextmenu="(e) => showContextMenu(e, tabId)"
              draggable="true"
              @dragstart="handleDragStart(tabId)"
              @dragover="(e) => handleDragOver(e, tabId)"
              @dragleave="handleDragLeave"
              @drop="(e) => handleDrop(e, tabId)"
              @dragend="handleDragEnd"
            >
              <component 
                :is="getTabIcon(tabs.find(t => t.id === tabId)?.icon)" 
                :size="14" 
                class="tab-icon"
              />
              <span class="tab-name">{{ tabs.find(t => t.id === tabId)?.name }}</span>
              <button class="tab-close" @click.stop="closeTab(tabId)"><XIcon :size="14" /></button>
            </div>
          </div>
        </template>
        
        <!-- 分屏模式：两个 Tabs 容器 -->
        <template v-else>
          <div class="tabs-split-container">
            <!-- 左 Pane tabs -->
            <div
              class="tabs-container pane-tabs left-tabs"
              :class="{ 'drag-over-pane': dragOverPaneId === 'pane-1' }"
              @dragover="(e) => handlePaneDragOver(e, 'pane-1')"
              @dragleave="handlePaneDragLeave"
              @drop="(e) => handlePaneDrop(e, 'pane-1')"
              @dblclick.self="toggleMaximize"
            >
              <div
                v-for="tabId in layout.panes[0].tabIds"
                :key="tabId"
                :class="['tab-item', { active: activeTabId === tabId, dragging: draggedTabId === tabId, 'drag-over': dragOverTabId === tabId }]"
                :title="tabs.find(t => t.id === tabId)?.name"
                @click="switchTab(tabId, 'pane-1')"
                @dblclick="closeTab(tabId)"
                @contextmenu="(e) => showContextMenu(e, tabId)"
                draggable="true"
                @dragstart="handleDragStart(tabId)"
                @dragover="(e) => handleDragOver(e, tabId)"
                @dragleave="handleDragLeave"
                @drop="(e) => handleDrop(e, tabId, 'pane-1')"
                @dragend="handleDragEnd"
              >
                <component
                  :is="getTabIcon(tabs.find(t => t.id === tabId)?.icon)"
                  :size="14"
                  class="tab-icon"
                />
                <span class="tab-name">{{ tabs.find(t => t.id === tabId)?.name }}</span>
                <button class="tab-close" @click.stop="closeTab(tabId)"><XIcon :size="14" /></button>
              </div>
            </div>

            <!-- 右 Pane tabs -->
            <div
              class="tabs-container pane-tabs right-tabs"
              :class="{ 'drag-over-pane': dragOverPaneId === 'pane-2' }"
              @dragover="(e) => handlePaneDragOver(e, 'pane-2')"
              @dragleave="handlePaneDragLeave"
              @drop="(e) => handlePaneDrop(e, 'pane-2')"
              @dblclick.self="toggleMaximize"
            >
              <div
                v-for="tabId in layout.panes[1].tabIds"
                :key="tabId"
                :class="['tab-item', { active: activeTabId === tabId, dragging: draggedTabId === tabId, 'drag-over': dragOverTabId === tabId }]"
                :title="tabs.find(t => t.id === tabId)?.name"
                @click="switchTab(tabId, 'pane-2')"
                @dblclick="closeTab(tabId)"
                @contextmenu="(e) => showContextMenu(e, tabId)"
                draggable="true"
                @dragstart="handleDragStart(tabId)"
                @dragover="(e) => handleDragOver(e, tabId)"
                @dragleave="handleDragLeave"
                @drop="(e) => handleDrop(e, tabId, 'pane-2')"
                @dragend="handleDragEnd"
              >
                <component
                  :is="getTabIcon(tabs.find(t => t.id === tabId)?.icon)"
                  :size="14"
                  class="tab-icon"
                />
                <span class="tab-name">{{ tabs.find(t => t.id === tabId)?.name }}</span>
                <button class="tab-close" @click.stop="closeTab(tabId)"><XIcon :size="14" /></button>
              </div>
            </div>
          </div>
        </template>
        
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
          <button v-if="layout.mode !== 'split'" class="context-menu-item" @click="handleOpenInSplit">在分屏中打开</button>
          <button v-if="layout.mode === 'split'" class="context-menu-item" @click="exitSplitMode">退出分屏</button>
        </div>
      </header>
      
      <!-- 内容区域 -->
      <template v-if="layout.mode === 'single'">
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
            :group-id="activeTab.groupId"
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
                :tab-id="tab.id"
                :book-md5="tab.bookMd5"
                :shelf-name="tab.shelfName"
                :is-active="activeTabId === tab.id"
                :ref="(el: any) => { 
                  if (el) readerRefs.set(tab.id, el)
                  else readerRefs.delete(tab.id)
                }"
                @ready="handleReaderReady"
                @bookmark-saved="handleBookmarkSaved"
              />
            </template>
          </div>
        </div>
      </template>
      <template v-else>
        <div class="split-content-container">
          <!-- 左 Pane -->
          <div class="split-pane">
            <div 
              v-if="tabs.find(t => t.id === layout.panes[0].activeTabId)" 
              class="content-container"
              @click="handlePaneClick('pane-1')"
            >
              <template v-for="pane in [layout.panes[0]]" :key="pane.id">
                <Bookshelf 
                  v-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'bookshelf'"
                  :shelf-id="tabs.find(t => t.id === pane.activeTabId)?.shelfId" 
                  :key="'bookshelf-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                  @open-book="(book: any) => openReaderTab(book, 'pane-1')"
                  @toggle-webdav="toggleWebDav"
                  @open-book-detail="openBookDetailTab"
                  @open-group="(group: any) => openGroupTab(group, 'pane-1')"
                  @close-tab-menu="hideContextMenu"
                />
                <GroupDetail 
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'group-detail'"
                  :key="'group-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                  :group-id="tabs.find(t => t.id === pane.activeTabId)?.groupId"
                  @open-book="(book: any) => openReaderTab(book, 'pane-1')"
                  @toggle-webdav="toggleWebDav"
                  @open-book-detail="openBookDetailTab"
                  @close-tab-menu="hideContextMenu"
                />
                <SettingsTab 
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'settings'"
                  :key="'settings-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                />
                <AddThemeTab 
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'add-theme'"
                  :key="'add-theme-' + (tabs.find(t => t.id === pane.activeTabId)?.editThemeId || '') + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                  :edit-theme-id="tabs.find(t => t.id === pane.activeTabId)?.editThemeId"
                  @saved="handleThemeSaved(pane.activeTabId)"
                />
                <BookDetailTab
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'book-detail'"
                  :book="tabs.find(t => t.id === pane.activeTabId)?.bookData"
                  :key="'book-detail-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                />
                <div v-show="tabs.find(t => t.id === pane.activeTabId)?.type === 'reader'" class="reader-container">
                  <template v-for="tab in tabs.filter(t => t.type === 'reader' && pane.tabIds.includes(t.id))" :key="'reader-' + pane.id + '-' + tab.id">
                    <ReaderContent 
                      v-if="tab.filePath"
                      v-show="pane.activeTabId === tab.id"
                      :file-path="tab.filePath"
                      :tab-id="tab.id"
                      :book-md5="tab.bookMd5"
                      :shelf-name="tab.shelfName"
                      :is-split-mode="true"
                      :is-active="tab.id === activeTabId"
                      @click="switchTab(tab.id, pane.id)"
                      @scroll="switchTab(tab.id, pane.id)"
                      @ready="handleReaderReady"
                      @bookmark-saved="handleBookmarkSaved"
                      :ref="(el: any) => { 
                        if (el) readerRefs.set(tab.id, el)
                        else readerRefs.delete(tab.id)
                      }"
                    />
                  </template>
                </div>
              </template>
            </div>
          </div>
          
          <div class="split-divider"></div>
          
          <!-- 右 Pane -->
          <div class="split-pane">
            <div 
              v-if="tabs.find(t => t.id === layout.panes[1].activeTabId)" 
              class="content-container"
              @click="handlePaneClick('pane-2')"
            >
              <template v-for="pane in [layout.panes[1]]" :key="pane.id">
                <Bookshelf 
                  v-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'bookshelf'"
                  :shelf-id="tabs.find(t => t.id === pane.activeTabId)?.shelfId" 
                  :key="'bookshelf-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                  @open-book="(book: any) => openReaderTab(book, 'pane-2')"
                  @toggle-webdav="toggleWebDav"
                  @open-book-detail="openBookDetailTab"
                  @open-group="(group: any) => openGroupTab(group, 'pane-2')"
                  @close-tab-menu="hideContextMenu"
                />
                <GroupDetail 
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'group-detail'"
                  :key="'group-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                  :group-id="tabs.find(t => t.id === pane.activeTabId)?.groupId"
                  @open-book="(book: any) => openReaderTab(book, 'pane-2')"
                  @toggle-webdav="toggleWebDav"
                  @open-book-detail="openBookDetailTab"
                  @close-tab-menu="hideContextMenu"
                />
                <SettingsTab 
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'settings'"
                  :key="'settings-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                />
                <AddThemeTab 
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'add-theme'"
                  :key="'add-theme-' + (tabs.find(t => t.id === pane.activeTabId)?.editThemeId || '') + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                  :edit-theme-id="tabs.find(t => t.id === pane.activeTabId)?.editThemeId"
                  @saved="handleThemeSaved(pane.activeTabId)"
                />
                <BookDetailTab
                  v-else-if="tabs.find(t => t.id === pane.activeTabId)?.type === 'book-detail'"
                  :book="tabs.find(t => t.id === pane.activeTabId)?.bookData"
                  :key="'book-detail-' + pane.activeTabId + '-' + (tabRefreshKeys[pane.activeTabId] || 0)"
                />
                <div v-show="tabs.find(t => t.id === pane.activeTabId)?.type === 'reader'" class="reader-container">
                  <template v-for="tab in tabs.filter(t => t.type === 'reader' && pane.tabIds.includes(t.id))" :key="'reader-' + pane.id + '-' + tab.id">
                    <ReaderContent 
                      v-if="tab.filePath"
                      v-show="pane.activeTabId === tab.id"
                      :file-path="tab.filePath"
                      :tab-id="tab.id"
                      :book-md5="tab.bookMd5"
                      :shelf-name="tab.shelfName"
                      :is-split-mode="true"
                      :is-active="tab.id === activeTabId"
                      @click="switchTab(tab.id, pane.id)"
                      @scroll="switchTab(tab.id, pane.id)"
                      @ready="handleReaderReady"
                      @bookmark-saved="handleBookmarkSaved"
                      :ref="(el: any) => { 
                        if (el) readerRefs.set(tab.id, el)
                        else readerRefs.delete(tab.id)
                      }"
                    />
                  </template>
                </div>
              </template>
            </div>
          </div>
        </div>
      </template>
      
      <!-- 下载状态气泡提示 -->
      <div class="toast-container">
        <div 
          v-for="toast in toastList" 
          :key="toast.id"
          :class="['download-toast', toast.type]"
        >
          {{ toast.message }}
        </div>
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
    
    <!-- 书架选择弹窗 -->
    <ShelfSelector
      :visible="shelfSelectorVisible"
      :epub-path="pendingEpubPath"
      @close="handleShelfSelectorClose"
      @select="handleShelfSelect"
      @import-only="handleImportOnly"
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

.func-btn.active {
  background-color: rgba(color-mix(in srgb, var(--primary-color)50%, var(--accent-color) 50%), 0.15);
  color: color-mix(in srgb, var(--primary-color) 50%, var(--accent-color) 50%);
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

/* 气泡容器 */
.toast-container {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: flex-end;
}

/* 下载状态气泡 */
.download-toast {
  padding: 14px 20px;
  border-radius: 10px;
  font-size: 0.9rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
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
  background: color-mix(in srgb, var(--primary-color) 50%, var(--accent-color) 50%);
  border-radius: 3px 3px 0 0;
  transform: scaleX(0);
  transition: transform var(--transition-normal);
}

.tab-item:hover {
  background: var(--sidebar-bg);
  color: var(--text-primary);
}

.tab-item.active {
  color: color-mix(in srgb, var(--primary-color) 50%, var(--accent-color) 50%);
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

/* ===== 分屏布局样式 ===== */

/* 分屏内容容器 */
.split-content-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 分屏 pane */
.split-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

/* 分屏内容分隔线 */
.split-divider {
  width: 1px;
  background: var(--border-color);
  flex-shrink: 0;
}

/* Tab bar 分屏容器 */
.tabs-split-container {
  flex: 1;
  display: flex;
  align-items: stretch;
  overflow: hidden;
}

/* 分屏时的 tabs 容器 */
.pane-tabs {
  flex: 1;
  display: flex;
  overflow: hidden;
  align-items: center;
  min-width: 0; /* 允许收缩，使文本截断生效 */
  transition: background-color 0.2s ease;
}

.pane-tabs.drag-over-pane {
  background-color: var(--primary-light);
  border-radius: var(--radius-md);
}

/* 右侧 tabs 容器 - 添加左侧分隔线（贯穿整个 tab 栏高度） */
.right-tabs {
  border-left: 1px solid var(--border-color);
  padding-left: 8px;
  margin-right: -136px;
}

</style>
