<script setup lang="ts">
import { computed, ref, onMounted, watch, onUnmounted } from 'vue'
import { useLibraryStore } from '../stores/library'
import { useSettingsStore } from '../stores/settings'
import { importBook, importBooks, importBooksFromFolder, type ImportResult } from '../utils/bookImporter'
import BookIcon from './icons/BookIcon.vue'
import CloudIcon from './icons/CloudIcon.vue'
import DownloadIcon from './icons/DownloadIcon.vue'
import FileIcon from './icons/FileIcon.vue'
import { naturalCompare } from '../composables/useNaturalSort'
import ListIcon from './icons/ListIcon.vue'
import XIcon from './icons/XIcon.vue'
import FolderIcon from './icons/FolderIcon.vue'
import PlusIcon from './icons/PlusIcon.vue'
import CustomModal from './CustomModal.vue'
import CustomModalEx from './CustomModalEx.vue'

const props = defineProps<{ groupId?: string }>()

const emit = defineEmits<{
  (e: 'open-book', book: any): void
  (e: 'toggle-webdav'): void
  (e: 'open-book-detail', book: any): void
  (e: 'back'): void
  (e: 'close-tab-menu'): void
}>()

const store = useLibraryStore()
const settingsStore = useSettingsStore()

// WebDAV侧边栏状态
const isWebdavOpen = ref(false)

// 弹窗状态
const showModal = ref(false)
const modalTitle = ref('提示')
const modalMessage = ref('')
const modalType = ref<'info' | 'warning' | 'success' | 'error'>('info')
const modalShowCancel = ref(true)
const modalConfirmAction = ref<(() => void) | null>(null)

// 显示确认弹窗
const showConfirmModal = (
  title: string,
  message: string,
  type: 'info' | 'warning' | 'success' | 'error' = 'info',
  options?: { showCancel?: boolean; onConfirm?: () => void }
) => {
  modalTitle.value = title
  modalMessage.value = message
  modalType.value = type
  modalShowCancel.value = options?.showCancel ?? true
  modalConfirmAction.value = options?.onConfirm || null
  showModal.value = true
}

// 确认弹窗回调
const handleModalConfirm = () => {
  showModal.value = false
  if (modalConfirmAction.value) {
    modalConfirmAction.value()
  }
}

const handleModalCancel = () => {
  showModal.value = false
}

// 显示提示弹窗
const showAlert = (title: string, message: string, type: 'info' | 'warning' | 'success' | 'error' = 'info') => {
  modalTitle.value = title
  modalMessage.value = message
  modalType.value = type
  modalShowCancel.value = false
  modalConfirmAction.value = null
  showModal.value = true
}

// 当前分组（优先使用传入的 groupId）
const currentGroup = computed(() => {
  if (props.groupId) {
    return store.currentGroups.find(g => g.id === props.groupId) || null
  }
  return store.activeGroup
})

// 当前分组的书籍列表
const books = computed(() => {
  const gid = props.groupId || store.activeGroupId
  if (!gid) return []
  return store.currentBooks.filter(b => b.groupId === gid)
})

const sortBooks = (items: any[]) => {
  const sortBy = settingsStore.sortBy
  if (sortBy === 'default') return items
  
  return [...items].sort((a, b) => {
    const aTitle = a.title || ''
    const bTitle = b.title || ''
    const aAuthor = a.author || ''
    const bAuthor = b.author || ''
    
    let result = 0
    if (sortBy === 'title-asc') result = naturalCompare(aTitle, bTitle)
    else if (sortBy === 'title-desc') result = naturalCompare(bTitle, aTitle)
    else if (sortBy === 'author-asc') result = naturalCompare(aAuthor, bAuthor)
    else if (sortBy === 'author-desc') result = naturalCompare(bAuthor, aAuthor)
    return result
  })
}

const filteredBooks = computed(() => {
  if (!searchKeyword.value.trim()) {
    return sortBooks(books.value)
  }
  
  const keyword = searchKeyword.value.toLowerCase().trim()
  const results = books.value.filter(book => {
    const title = (book.title || '').toLowerCase()
    const author = (book.author || '').toLowerCase()
    return title.includes(keyword) || author.includes(keyword)
  })
  return sortBooks(results)
})

// 正在下载的书籍 ID 列表
const downloadingBooks = ref<Set<string>>(new Set())

// 书籍下载状态缓存 (bookKey -> isDownloaded)
const bookDownloadedStatus = ref<Map<string, boolean>>(new Map())

// 导入下拉菜单
const showImportMenu = ref(false)
const importMenuRef = ref<HTMLDivElement | null>(null)
const isImporting = ref(false)

// 选择模式
const isSelectMode = ref(false)
const selectedBooks = ref<Set<string>>(new Set())

// 显示移动到分组对话框
const showMoveToGroupDialog = ref(false)
const selectedTargetGroupId = ref<string | null>(null)

// 创建分组对话框
const showCreateGroupDialog = ref(false)
const newGroupName = ref('')
const createGroupReturnToMove = ref(false)  // 创建后是否返回移动面板

// 从移动到分组面板中创建分组
const handleCreateFromMove = () => {
  showMoveToGroupDialog.value = false
  newGroupName.value = ''
  createGroupReturnToMove.value = true
  showCreateGroupDialog.value = true
}

// 创建分组
const handleCreateGroup = async () => {
  if (!newGroupName.value.trim()) {
    return
  }
  
  await store.createGroup(newGroupName.value.trim())
  showCreateGroupDialog.value = false
  newGroupName.value = ''
  
  // 如果是从移动面板触发的，创建后返回移动面板
  if (createGroupReturnToMove.value) {
    createGroupReturnToMove.value = false
    showMoveToGroupDialog.value = true
  }
}

// 取消创建分组
const handleCancelCreateGroup = () => {
  showCreateGroupDialog.value = false
  newGroupName.value = ''
  
  // 如果是从移动面板触发的，取消后返回移动面板
  if (createGroupReturnToMove.value) {
    createGroupReturnToMove.value = false
    showMoveToGroupDialog.value = true
  }
}

// 右键菜单
const contextMenu = ref({
  show: false,
  x: 0,
  y: 0,
  bookId: ''
})
const contextMenuRef = ref<HTMLDivElement | null>(null)

// 搜索功能
const searchKeyword = ref('')

// 书籍拖拽排序
const draggingBookId = ref<string | null>(null)
const dragOverBookId = ref<string | null>(null)

const handleBookDragStart = (event: DragEvent, bookId: string) => {
  if (isSelectMode.value) {
    event.preventDefault()
    return
  }
  draggingBookId.value = bookId
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', bookId)
  }
}

const handleBookDragOver = (event: DragEvent, bookId: string) => {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
  dragOverBookId.value = bookId
}

const handleBookDragLeave = () => {
  dragOverBookId.value = null
}

const handleBookDrop = (event: DragEvent, targetBookId: string) => {
  event.preventDefault()
  
  if (!draggingBookId.value || draggingBookId.value === targetBookId) {
    draggingBookId.value = null
    dragOverBookId.value = null
    return
  }
  
  const groupId = props.groupId || store.activeGroupId || undefined
  
  const filteredBooks = books.value.filter(b => b.groupId === groupId)
  const fromIndex = filteredBooks.findIndex(b => b.id === draggingBookId.value)
  const toIndex = filteredBooks.findIndex(b => b.id === targetBookId)
  
  if (fromIndex !== -1 && toIndex !== -1) {
    store.reorderBook(fromIndex, toIndex, groupId)
  }
  
  draggingBookId.value = null
  dragOverBookId.value = null
}

const handleBookDragEnd = () => {
  draggingBookId.value = null
  dragOverBookId.value = null
}

// ============ 快捷键功能 ============

// 检查当前焦点是否在输入框内
const isInputFocused = () => {
  const activeElement = document.activeElement
  return activeElement?.tagName === 'INPUT' || activeElement?.tagName === 'TEXTAREA'
}

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && isSelectMode.value) {
    e.preventDefault()
    exitSelectMode()
  }
  
  // Ctrl+A - 全选书籍，但如果焦点在输入框内则不处理
  if (e.ctrlKey && e.key === 'a' && !isInputFocused()) {
    e.preventDefault()
    if (!isSelectMode.value) {
      enterSelectMode()
    }
    selectAllBooks()
  }
  
  // Ctrl+D - 删除选中书籍（仅在选择模式下）
  if (e.ctrlKey && e.key === 'd' && isSelectMode.value && selectedBooks.value.size > 0) {
    e.preventDefault()
    handleBatchDelete()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

// ============ 选择模式 ============

const enterSelectMode = () => {
  isSelectMode.value = true
  selectedBooks.value.clear()
}

const toggleSelectMode = () => {
  if (isSelectMode.value) {
    exitSelectMode()
  } else {
    enterSelectMode()
  }
}

const exitSelectMode = () => {
  isSelectMode.value = false
  selectedBooks.value.clear()
}

const handleBookClick = (event: MouseEvent, book: any) => {
  if (event.ctrlKey && event.button === 0) {
    event.preventDefault()
    if (!isSelectMode.value) {
      enterSelectMode()
    }
    toggleBookSelection(book.id)
    return
  }
  
  if (isSelectMode.value) {
    toggleBookSelection(book.id)
  } else {
    openBook(book)
  }
}

const selectAllBooks = () => {
  filteredBooks.value.forEach(book => {
    selectedBooks.value.add(book.id)
  })
}

const deselectAllBooks = () => {
  selectedBooks.value.clear()
}

const toggleBookSelection = (bookId: string) => {
  if (selectedBooks.value.has(bookId)) {
    selectedBooks.value.delete(bookId)
  } else {
    selectedBooks.value.add(bookId)
  }
}

const handleBatchDelete = async () => {
  if (selectedBooks.value.size === 0) return
  
  showConfirmModal(
    '确认删除',
    `确定要删除 ${selectedBooks.value.size} 本书籍吗？`,
    'warning',
    { onConfirm: async () => {
      for (const bookId of selectedBooks.value) {
        await store.deleteBook(bookId)
      }
      exitSelectMode()
    }}
  )
}

// 计算批量操作的状态文本
const batchReadStatusText = computed(() => {
  let hasRead = false
  let hasUnread = false
  
  for (const bookId of selectedBooks.value) {
    const book = books.value.find((b: any) => b.id === bookId)
    if (book?.readStatus === 'read') {
      hasRead = true
    } else {
      hasUnread = true
    }
    
    if (hasRead && hasUnread) break
  }
  
  if (hasRead && !hasUnread) return '未读'
  return '已读'
})

// 批量切换阅读状态
const handleBatchToggleRead = async () => {
  if (selectedBooks.value.size === 0) return
  
  const targetStatus = batchReadStatusText.value === '已读' ? 'read' : 'unread'
  
  for (const bookId of selectedBooks.value) {
    const book = books.value.find((b: any) => b.id === bookId)
    if (book) {
      book.readStatus = targetStatus
      await store.updateBook(book)
    }
  }
}

// 批量移出分组
const handleBatchRemoveFromGroup = async () => {
  if (selectedBooks.value.size === 0) return
  
  showConfirmModal(
    '确认移出',
    `确定要将 ${selectedBooks.value.size} 本书籍移出分组吗？`,
    'info',
    { onConfirm: async () => {
      for (const bookId of selectedBooks.value) {
        await store.removeBookFromGroup(bookId)
      }
      exitSelectMode()
    }}
  )
}

// 获取分组内的书籍数量
const getGroupBookCount = (groupId: string) => {
  return store.currentBooks.filter(book => book.groupId === groupId).length
}

// 批量移动书籍到分组（点击按钮）
const handleMoveToGroup = async () => {
  if (selectedBooks.value.size === 0 || !selectedTargetGroupId.value) return
  
  for (const bookId of selectedBooks.value) {
    await store.addBookToGroup(bookId, selectedTargetGroupId.value)
  }
  
  showMoveToGroupDialog.value = false
  selectedTargetGroupId.value = null
  exitSelectMode()
}

// 批量移动书籍到分组（点击分组直接移动）
const handleMoveToGroupDirectly = async (groupId: string) => {
  if (selectedBooks.value.size === 0) return
  
  for (const bookId of selectedBooks.value) {
    await store.addBookToGroup(bookId, groupId)
  }
  
  showMoveToGroupDialog.value = false
  selectedTargetGroupId.value = null
  exitSelectMode()
}

// ============ 导入功能 ============

const toggleWebdav = () => {
  isWebdavOpen.value = !isWebdavOpen.value
  emit('toggle-webdav')
}

const toggleImportMenu = () => {
  showImportMenu.value = !showImportMenu.value
}

const closeImportMenu = () => {
  showImportMenu.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  if (importMenuRef.value && !importMenuRef.value.contains(event.target as Node)) {
    showImportMenu.value = false
  }
}

const processImportResult = async (result: ImportResult) => {
  if (!result.success) {
    console.error('导入书籍失败:', result.error)
    return
  }
  
  const shelfName = store.activeShelf?.name || '默认书架'
  const groupId = currentGroup.value?.id
  
  const added = await store.addBook(result.title, result.coverUrl, result.md5, result.filePath, result.author)
  
  if (added && groupId) {
    await store.addBookToGroup(Date.now().toString(), groupId)
    console.log('书籍导入成功并添加到分组:', result.title)
    const shelfNameForSync = store.activeShelf?.name || '默认书架'
    // @ts-ignore
    window.go.main.App.TriggerAutoSync(shelfNameForSync)
  } else if (added) {
    console.log('书籍导入成功:', result.title)
  } else {
    console.log('书籍已存在，跳过:', result.title)
  }
}

const handleSingleFileImport = async () => {
  closeImportMenu()
  
  try {
    // @ts-ignore
    const filePath = await window.go.main.App.SelectEpubFiles()
    
    if (!filePath) {
      return
    }

    isImporting.value = true
    const shelfName = store.activeShelf?.name || '默认书架'
    const result = await importBook(filePath, shelfName)
    
    if (result.success) {
      const added = await store.addBook(result.title, result.coverUrl, result.md5, result.filePath, result.author)
      if (added && currentGroup.value) {
        const newBook = store.currentBooks[store.currentBooks.length - 1]
        await store.addBookToGroup(newBook.id, currentGroup.value.id)
      }
    }
  } catch (e) {
    console.error('导入书籍失败', e)
    showConfirmModal('导入失败', '导入书籍失败，请重试', 'error', { showCancel: false })
  } finally {
    isImporting.value = false
  }
}

const handleMultipleFilesImport = async () => {
  closeImportMenu()
  
  try {
    // @ts-ignore
    const filePaths: string[] = await window.go.main.App.SelectMultipleEpubFiles()
    
    if (!filePaths || filePaths.length === 0) {
      return
    }

    isImporting.value = true
    const shelfName = store.activeShelf?.name || '默认书架'
    
    const results = await importBooks(filePaths, shelfName, (current, total, title) => {
      console.log(`导入进度: ${current}/${total} - ${title}`)
    })
    
    let successCount = 0
    let skipCount = 0
    let failCount = 0
    
    for (const result of results) {
      if (result.success) {
        const added = await store.addBook(result.title, result.coverUrl, result.md5, result.filePath, result.author)
        if (added && currentGroup.value) {
          const newBook = store.currentBooks[store.currentBooks.length - 1]
          await store.addBookToGroup(newBook.id, currentGroup.value.id)
          successCount++
        } else if (added) {
          successCount++
        } else {
          skipCount++
        }
      } else {
        failCount++
        console.error('导入失败:', result.error)
      }
    }
    
    const shelfNameForSync = store.activeShelf?.name || '默认书架'
    // @ts-ignore
    window.go.main.App.TriggerAutoSync(shelfNameForSync)
    
    showConfirmModal('导入完成', `成功: ${successCount}，跳过: ${skipCount}，失败: ${failCount}`, 
      failCount > 0 ? 'warning' : 'success', { showCancel: false })
  } catch (e) {
    console.error('批量导入书籍失败', e)
    showConfirmModal('导入失败', '批量导入书籍失败，请重试', 'error', { showCancel: false })
  } finally {
    isImporting.value = false
  }
}

const handleFolderImport = async () => {
  closeImportMenu()
  
  try {
    // @ts-ignore
    const folderPath = await window.go.main.App.SelectEpubFolder()
    
    if (!folderPath) {
      return
    }

    isImporting.value = true
    const shelfName = store.activeShelf?.name || '默认书架'
    
    const results = await importBooksFromFolder(folderPath, shelfName, (current, total, title) => {
      console.log(`导入进度: ${current}/${total} - ${title}`)
    })
    
    let successCount = 0
    let skipCount = 0
    let failCount = 0
    
    for (const result of results) {
      if (result.success) {
        const added = await store.addBook(result.title, result.coverUrl, result.md5, result.filePath, result.author)
        if (added && currentGroup.value) {
          const newBook = store.currentBooks[store.currentBooks.length - 1]
          await store.addBookToGroup(newBook.id, currentGroup.value.id)
          successCount++
        } else if (added) {
          successCount++
        } else {
          skipCount++
        }
      } else {
        failCount++
        console.error('导入失败:', result.error)
      }
    }
    
    const shelfNameForSync = store.activeShelf?.name || '默认书架'
    // @ts-ignore
    window.go.main.App.TriggerAutoSync(shelfNameForSync)
    
    showConfirmModal('导入完成', `成功: ${successCount}，跳过: ${skipCount}，失败: ${failCount}`, 
      failCount > 0 ? 'warning' : 'success', { showCancel: false })
  } catch (e) {
    console.error('文件夹导入书籍失败', e)
    showConfirmModal('导入失败', '文件夹导入书籍失败，请重试', 'error', { showCancel: false })
  } finally {
    isImporting.value = false
  }
}

// ============ 下载和打开书籍 ============

const getBookKey = (book: any): string => {
  return `${store.activeShelf?.name || ''}:${book.md5}`
}

const isBookDownloaded = async (book: any): Promise<boolean> => {
  if (!book.md5 || !store.activeShelf) return false
  try {
    // @ts-ignore
    const localPath = await window.go.main.App.GetBookLocalPath(store.activeShelf.name, book.md5)
    // @ts-ignore
    return await window.go.main.App.FileExists(localPath)
  } catch {
    return false
  }
}

const checkAllBooksStatus = async () => {
  if (!store.activeShelf) return
  
  for (const book of filteredBooks.value) {
    const key = getBookKey(book)
    const exists = await isBookDownloaded(book)
    bookDownloadedStatus.value.set(key, exists)
  }
}

const updateBookStatus = (book: any, downloaded: boolean) => {
  const key = getBookKey(book)
  bookDownloadedStatus.value.set(key, downloaded)
}

const downloadBook = async (book: any) => {
  if (!book.md5 || !store.activeShelf) return
  
  const bookKey = getBookKey(book)
  if (downloadingBooks.value.has(bookKey)) {
    console.log('该书籍正在下载中')
    return
  }
  
  const exists = await isBookDownloaded(book)
  if (exists) {
    console.log('该书籍已存在，跳过下载:', book.title)
    updateBookStatus(book, true)
    return
  }
  
  downloadingBooks.value.add(bookKey)
  
  try {
    // @ts-ignore
    await window.go.main.App.DownloadSingleEpub(store.activeShelf.name, book.md5, '')
    console.log('书籍下载完成:', book.title)
    updateBookStatus(book, true)
  } catch (e) {
    console.error('下载失败:', e)
  } finally {
    downloadingBooks.value.delete(bookKey)
  }
}

const updateBookLocalPath = async (book: any) => {
  if (!book.md5 || !store.activeShelf) return
  
  try {
    // @ts-ignore
    const localPath = await window.go.main.App.GetBookLocalPath(store.activeShelf.name, book.md5)
    if (localPath) {
      book.filePath = localPath
      await store.saveBooks()
    }
  } catch (error) {
    console.error('更新书籍路径失败:', error)
  }
}

const openBook = async (book: any) => {
  if (!book.md5 || !store.activeShelf) return
  
  const bookKey = getBookKey(book)
  
  const exists = bookDownloadedStatus.value.get(bookKey) || await isBookDownloaded(book)
  
  if (!exists) {
    if (downloadingBooks.value.has(bookKey)) {
      console.log('该书籍正在下载中')
      return
    }
    
    downloadingBooks.value.add(bookKey)
    
    try {
      // @ts-ignore
      await window.go.main.App.DownloadSingleEpub(store.activeShelf.name, book.md5, '')
      console.log('书籍下载完成，准备打开:', book.title)
      updateBookStatus(book, true)
      await updateBookLocalPath(book)
    } catch (e) {
      console.error('下载失败:', e)
      downloadingBooks.value.delete(bookKey)
      return
    } finally {
      downloadingBooks.value.delete(bookKey)
    }
  }
  
  await store.moveBookToFront(book.id)
  
  emit('open-book', book)
}

watch([store.activeShelf, books], async () => {
  await checkAllBooksStatus()
}, { immediate: true })

onMounted(async () => {
  await checkAllBooksStatus()
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('click', handleClickOutsideContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('click', handleClickOutsideContextMenu)
})

const handleDeleteBook = (bookId: string) => {
  showConfirmModal(
    '确认删除',
    '确定要删除这本书吗？',
    'warning',
    { onConfirm: () => {
      store.deleteBook(bookId)
    }}
  )
}

const showContextMenu = (event: MouseEvent, bookId: string) => {
  event.preventDefault()
  event.stopPropagation()
  
  if (isSelectMode.value) return
  
  emit('close-tab-menu')
  
  contextMenu.value = {
    show: true,
    x: event.clientX,
    y: event.clientY,
    bookId
  }
}

const closeContextMenu = () => {
  contextMenu.value.show = false
}

// 获取当前书籍的阅读状态文本
const getReadStatusText = () => {
  if (!contextMenu.value.bookId) return '标记状态'
  const book = books.value.find((b: any) => b.id === contextMenu.value.bookId)
  const isRead = book?.readStatus === 'read' || (book?.readingProgress && book.readingProgress >= 100)
  return isRead ? '标记为未读' : '标记为已读'
}

// 切换阅读状态
const handleToggleReadStatus = async () => {
  if (!contextMenu.value.bookId) return
  
  const book = books.value.find((b: any) => b.id === contextMenu.value.bookId)
  if (!book) {
    closeContextMenu()
    return
  }
  
  const wasRead = book.readStatus === 'read'
  book.readStatus = wasRead ? 'unread' : 'read'
  
  if (wasRead && (book.readingProgress === undefined || book.readingProgress >= 100)) {
    if (book.filePath) {
      try {
        // @ts-ignore
        await window.go.main.App.ClearProgress(book.filePath)
        book.readingProgress = 0
      } catch (e) {
        console.warn('清除进度失败:', e)
      }
    }
  }
  
  await store.updateBook(book)
  closeContextMenu()
}

const handleContextMenuDelete = async () => {
  if (!contextMenu.value.bookId) return
  
  closeContextMenu()
  
  showConfirmModal(
    '确认删除',
    '确定要删除这本书吗？',
    'warning',
    { onConfirm: async () => {
      await store.deleteBook(contextMenu.value.bookId)
    }}
  )
}

const handleContextMenuRemoveFromGroup = async () => {
  if (!contextMenu.value.bookId) return
  
  await store.removeBookFromGroup(contextMenu.value.bookId)
  
  closeContextMenu()
}

const handleContextMenuMoveToGroup = () => {
  if (!contextMenu.value.bookId) return
  
  selectedBooks.value.clear()
  selectedBooks.value.add(contextMenu.value.bookId)
  showMoveToGroupDialog.value = true
  
  closeContextMenu()
}

const handleContextMenuDetail = () => {
  if (!contextMenu.value.bookId) return
  
  const book = filteredBooks.value.find(b => b.id === contextMenu.value.bookId)
  if (book) {
    emit('open-book-detail', book)
  }
  
  closeContextMenu()
}

// 打开文件所在位置
const handleOpenFileLocation = async () => {
  if (!contextMenu.value.bookId) return
  
  const book = filteredBooks.value.find(b => b.id === contextMenu.value.bookId)
  if (!book || !book.filePath) {
    closeContextMenu()
    return
  }
  
  closeContextMenu()
  
  try {
    // @ts-ignore
    await window.go.main.App.OpenFileLocation(book.filePath)
  } catch (e) {
    console.error('打开文件位置失败:', e)
  }
}

// 重命名相关状态
const isRenaming = ref(false)
const renamingBookId = ref('')
const renamingTitle = ref('')

// 开始重命名书籍
const startRenameBook = async () => {
  if (!contextMenu.value.bookId) return
  
  const book = filteredBooks.value.find((b: any) => b.id === contextMenu.value.bookId)
  if (book) {
    renamingBookId.value = book.id
    renamingTitle.value = book.title
    isRenaming.value = true
  }
  
  closeContextMenu()
}

// 保存重命名
const saveBookRename = async () => {
  if (!renamingBookId.value || !renamingTitle.value.trim()) {
    showAlert('提示', '书名不能为空', 'warning')
    return
  }
  
  try {
    const newTitle = renamingTitle.value.trim()
    const book = filteredBooks.value.find((b: any) => b.id === renamingBookId.value)
    
    if (book && book.filePath) {
      // 保存到 config.json
      const lastSlash = Math.max(
        book.filePath.lastIndexOf('/'),
        book.filePath.lastIndexOf('\\')
      )
      const bookDir = book.filePath.substring(0, lastSlash)
      const configPath = `${bookDir}/config.json`
      
      // @ts-ignore
      const configExists = await window.go.main.App.FileExists(configPath)
      
      if (configExists) {
        // @ts-ignore
        const configContent = await window.go.main.App.ReadFile(configPath)
        const config = JSON.parse(configContent)
        config.title = newTitle
        const configJson = JSON.stringify(config, null, 2)
        const configBytes = Array.from(new TextEncoder().encode(configJson))
        // @ts-ignore
        await window.go.main.App.SaveFile(bookDir, 'config.json', configBytes)
      }
    }
    
    // 更新书架列表中的书名
    await store.updateBookTitle(renamingBookId.value, newTitle)
    
    isRenaming.value = false
    renamingBookId.value = ''
    renamingTitle.value = ''
  } catch (error) {
    console.error('保存书名失败:', error)
    showAlert('错误', '保存书名失败，请重试', 'error')
  }
}

// 取消重命名
const cancelRenameBook = () => {
  isRenaming.value = false
  renamingBookId.value = ''
  renamingTitle.value = ''
}

const handleClickOutsideContextMenu = (event: MouseEvent) => {
  if (contextMenuRef.value && !contextMenuRef.value.contains(event.target as Node)) {
    closeContextMenu()
  }
}
</script>

<template>
  <main class="group-detail-container">
    <header class="topbar">
      <div class="topbar-left">
        <input 
          type="text" 
          class="search-input" 
          v-model="searchKeyword"
          :placeholder="`在 ${books.length} 本书籍中搜索...`"
        />
      </div>
      <div class="topbar-actions">
        <button 
          class="btn primary import-btn" 
          :class="{ active: isSelectMode }"
          @click="toggleSelectMode"
        >
          <BookIcon :size="16" />{{ isSelectMode ? '取消选择' : '选择书籍' }}
        </button>
        <button 
          class="btn primary import-btn" 
          :class="{ active: isWebdavOpen }"
          @click="toggleWebdav"
        >
          <CloudIcon :size="16" />{{ isWebdavOpen ? 'WebDAV' : 'WebDAV' }}
        </button>
        <div class="import-dropdown" ref="importMenuRef">
          <button 
            class="btn primary import-btn" 
            @click.stop="toggleImportMenu"
            :disabled="isImporting"
          >
            <FileIcon :size="16" />{{ isImporting ? '导入中...' : '导入书籍' }}
          </button>
          <div v-show="showImportMenu" class="dropdown-menu">
            <div class="dropdown-item" @click="handleSingleFileImport">
              <FileIcon :size="18" class="item-icon" />
              <div class="item-content">
                <div class="item-title">单文件导入</div>
                <div class="item-desc">选择一个 EPUB 文件</div>
              </div>
            </div>
            <div class="dropdown-item" @click="handleMultipleFilesImport">
              <ListIcon :size="18" class="item-icon" />
              <div class="item-content">
                <div class="item-title">多文件导入</div>
                <div class="item-desc">选择多个 EPUB 文件</div>
              </div>
            </div>
            <div class="dropdown-divider"></div>
            <div class="dropdown-item" @click="handleFolderImport">
              <FileIcon :size="18" class="item-icon" />
              <div class="item-content">
                <div class="item-title">文件夹导入</div>
                <div class="item-desc">扫描文件夹中的所有 EPUB</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>

    <div class="content-area">
      <div v-if="filteredBooks.length === 0" class="empty-state">
        <div class="empty-text">
          {{ searchKeyword ? '没有找到匹配的书籍' : '分组中还没有书籍' }}
        </div>
        <button v-if="!searchKeyword" class="btn primary" @click="handleSingleFileImport" :disabled="isImporting">
          {{ isImporting ? '导入中...' : '导入你的第一本书' }}
        </button>
        <button v-else class="btn secondary" @click="searchKeyword = ''">
          清除搜索
        </button>
      </div>
      
      <div v-else class="book-grid" :style="{ gridTemplateColumns: `repeat(${settingsStore.bookshelfColumns}, minmax(0, 1fr))`, gap: `${settingsStore.coverGap}px` }">
        <div 
          v-for="book in filteredBooks" 
          :key="book.id" 
          draggable="true"
          :class="['book-card', { selected: selectedBooks.has(book.id), 'select-mode': isSelectMode, 'dragging': draggingBookId === book.id, 'drag-over': dragOverBookId === book.id }]"
          @click="handleBookClick($event, book)"
          @contextmenu="showContextMenu($event, book.id)"
          @dragstart="handleBookDragStart($event, book.id)"
          @dragover="handleBookDragOver($event, book.id)"
          @dragleave="handleBookDragLeave"
          @drop="handleBookDrop($event, book.id)"
          @dragend="handleBookDragEnd"
        >
          <div class="cover-wrapper">
            <div class="cover">
              <img v-if="book.coverUrl" :src="book.coverUrl" alt="封面" class="cover-image" />
              <div v-else class="cover-placeholder"><BookIcon :size="32" /></div>
              
              <div v-if="isSelectMode" class="select-checkbox">
                <span v-if="selectedBooks.has(book.id)">✓</span>
              </div>
              
              <div v-if="book.readStatus === 'read' || (book.readingProgress && book.readingProgress >= 100)" class="read-status">已读</div>
              <div v-else-if="book.readingProgress && book.readingProgress > 0 && book.readingProgress < 100" class="reading-progress">{{ book.readingProgress }}%</div>
              
              <div v-if="!isSelectMode && downloadingBooks.has(getBookKey(book))" class="download-status downloading">
                  <span class="spinner">⏳</span>
                </div>
                <div 
                  v-else-if="!isSelectMode && bookDownloadedStatus.get(getBookKey(book)) === false" 
                  class="download-status not-downloaded"
                  @click.stop="downloadBook(book)"
                >
                  <DownloadIcon :size="16" title="下载书籍" />
                </div>
            </div>
          </div>
          <div class="title">{{ book.title }}</div>
        </div>
      </div>
    </div>
    
    <div 
      v-show="contextMenu.show" 
      class="context-menu"
      ref="contextMenuRef"
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      @click.stop
    >
      <button class="context-menu-item" @click="handleContextMenuDetail">
        书籍详情
      </button>
      
      <button class="context-menu-item" @click="startRenameBook">
        重命名
      </button>

      <button class="context-menu-item" @click="handleContextMenuMoveToGroup">
        书籍分组
      </button>

      <button class="context-menu-item" @click="handleContextMenuRemoveFromGroup">
        移出分组
      </button>
      
      <button class="context-menu-item" @click="handleOpenFileLocation">
        打开当前位置
      </button>
      
      <button class="context-menu-item" @click="handleToggleReadStatus">
        {{ getReadStatusText() }}
      </button>
      
      <button class="context-menu-item danger" @click="handleContextMenuDelete">
        删除书籍
      </button>
    </div>
    
    <div v-if="isSelectMode" class="bottom-action-bar">
      <div class="action-bar-content">
        <div class="action-bar-info">
          已选择 {{ selectedBooks.size }}/{{ filteredBooks.length }} 本书籍
        </div>
        <div class="action-bar-buttons">
          <button 
            v-if="selectedBooks.size < books.length" 
            class="btn secondary" 
            @click="selectAllBooks"
          >
            全选
          </button>
          <button 
            v-else 
            class="btn secondary" 
            @click="deselectAllBooks"
          >
            取消全选
          </button>
          <button class="btn secondary" @click="handleBatchRemoveFromGroup" :disabled="selectedBooks.size === 0">
            移出分组
          </button>
          <button 
            v-if="store.currentGroups.filter(g => g.id !== currentGroup?.id).length > 0"
            class="btn secondary" 
            @click="showMoveToGroupDialog = true"
            :disabled="selectedBooks.size === 0"
          >
            移动到分组
          </button>
          <button class="btn secondary" @click="handleBatchToggleRead" :disabled="selectedBooks.size === 0">
            {{ batchReadStatusText }}
          </button>
          <button class="btn danger" @click="handleBatchDelete" :disabled="selectedBooks.size === 0">
            删除
          </button>
          <button class="btn secondary" @click="exitSelectMode">
            取消
          </button>
        </div>
      </div>
    </div>
    
    <!-- 重命名书籍对话框 -->
    <CustomModalEx
      :visible="isRenaming"
      title="重命名书籍"
      confirmText="确定"
      cancelText="取消"
      :confirmDisabled="!renamingTitle.trim()"
      @confirm="saveBookRename"
      @cancel="cancelRenameBook"
    >
      <input 
        type="text" 
        v-model="renamingTitle"
        class="modal-input"
        placeholder="请输入书籍名称"
        @keyup.enter="saveBookRename"
        autofocus
      />
    </CustomModalEx>
    
    <!-- 移动到分组对话框 -->
    <CustomModalEx
      :visible="showMoveToGroupDialog"
      title="选择目标分组"
      confirmText="移动"
      cancelText="取消"
      :confirmDisabled="!selectedTargetGroupId"
      @confirm="handleMoveToGroup"
      @cancel="showMoveToGroupDialog = false"
    >
      <div class="group-list">
        <div 
          v-for="group in store.currentGroups.filter(g => g.id !== currentGroup?.id)" 
          :key="group.id"
          class="group-option"
          :class="{ active: selectedTargetGroupId === group.id }"
          @click="selectedTargetGroupId = group.id"
          @dblclick="handleMoveToGroupDirectly(group.id)"
        >
          <FolderIcon :size="18" />
          <span class="group-name">{{ group.name }}</span>
          <span class="group-count">({{ getGroupBookCount(group.id) }} 本书)</span>
        </div>
        <div 
          v-if="store.currentGroups.filter(g => g.id !== currentGroup?.id).length === 0"
          class="group-empty"
        >
          暂无其他分组
        </div>
      </div>
      <div class="group-add-row" @click="handleCreateFromMove">
        <PlusIcon :size="16" />
        <span>创建新分组</span>
      </div>
    </CustomModalEx>

    <!-- 创建分组对话框 -->
    <CustomModalEx
      :visible="showCreateGroupDialog"
      title="创建分组"
      confirmText="创建"
      cancelText="取消"
      @confirm="handleCreateGroup"
      @cancel="handleCancelCreateGroup"
    >
      <input 
        type="text" 
        v-model="newGroupName"
        class="modal-input"
        placeholder="请输入分组名称"
        @keyup.enter="handleCreateGroup"
        autofocus
      />
    </CustomModalEx>
  </main>

  <!-- 自定义确认弹窗 -->
  <CustomModal
    :visible="showModal"
    :title="modalTitle"
    :message="modalMessage"
    :type="modalType"
    :showCancel="modalShowCancel"
    @confirm="handleModalConfirm"
    @cancel="handleModalCancel"
  />
</template>

<style scoped>
.group-detail-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color);
  min-height: 0;
  position: relative;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px 36px;
  border-bottom: 0px solid var(--border-color);
  user-select: none;
  box-shadow: var(--shadow-sm);
  position: relative;
}

.topbar::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent 0%, var(--primary-color) 50%, transparent 100%);
  opacity: 0.3;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.back-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background: transparent;
  color: var(--text-color);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-size: 0.9rem;
}

.back-button:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.group-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-primary);
}

.topbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  align-items: center;
}

.search-input {
  padding: 10px 18px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  outline: none;
  background: var(--bg-color);
  color: var(--text-color);
  width: 260px;
  font-size: 0.9rem;
  transition: all var(--transition-normal);
  font-weight: 400;
  letter-spacing: -0.01em;
}

.search-input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px var(--primary-light);
  width: 320px;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.btn {
  padding: 11px 20px;
  border: none;
  border-radius: var(--radius-lg);
  cursor: pointer;
  font-weight: 500;
  font-size: 0.88rem;
  transition: all var(--transition-normal);
  display: inline-flex;
  align-items: center;
  gap: 7px;
  letter-spacing: -0.01em;
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

.btn.primary {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  box-shadow: var(--shadow-md);
}

.btn.primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.btn.primary:active:not(:disabled) {
  transform: translateY(-1px);
}

.btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.btn.secondary {
  background: transparent;
  color: var(--text-color);
  border: 1.5px solid var(--border-color);
}

.btn.secondary:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
}

.btn.secondary.active {
  background: var(--primary-light);
  border-color: var(--primary-color);
}

.btn.danger {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  box-shadow: var(--shadow-md);
}

.btn.danger:hover:not(:disabled) {
  background: linear-gradient(135deg, var(--accent-color) 0%, var(--primary-color) 100%);
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.btn.danger:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
}

.import-dropdown {
  position: relative;
  display: inline-block;
}

.import-btn {
  display: flex;
  align-items: center;
  gap: 7px;
  white-space: nowrap;
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  min-width: 260px;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-2xl);
  z-index: 100;
  padding: 10px;
  animation: dropdownIn var(--transition-normal);
  overflow: hidden;
}

@keyframes dropdownIn {
  from {
    opacity: 0;
    transform: translateY(-10px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dropdown-item:hover {
  background: var(--primary-light);
  transform: translateX(6px);
}

.item-icon {
  font-size: 1.6rem;
}

.item-content {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.item-title {
  font-size: 0.92rem;
  font-weight: 500;
  color: var(--text-primary);
}

.item-desc {
  font-size: 0.78rem;
  color: var(--text-secondary);
}

.dropdown-divider {
  height: 1px;
  background: var(--border-color);
  margin: 8px 0;
}

.content-area {
  flex: 1;
  padding: 36px;
  overflow-y: auto;
}

.empty-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  user-select: none;
}

.empty-text {
  color: var(--text-secondary);
  font-size: 1.25rem;
  font-weight: 500;
}

.book-grid {
  display: grid;
  gap: 28px 24px;
}

.book-card {
  display: flex;
  flex-direction: column;
  cursor: pointer;
  position: relative;
  border-radius: var(--radius-xl);
  user-select: none;
  transition: all 0.2s ease;
}

.book-card:active {
  cursor: pointer;
}

.book-card.dragging {
  opacity: 0.3;
}

.book-card.dragging .cover-wrapper::before {
  opacity: 0;
}

.book-card.drag-over .cover {
  border: 2px solid var(--primary-color);
  background: rgba(79, 70, 229, 0.05);
}

.cover-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 2/3;
  margin-bottom: 10px;
  border-radius: 10px;
}

.cover-wrapper::before {
  content: '';
  position: absolute;
  inset: -5px;
  border-radius: 14px;
  background: conic-gradient(
    from var(--angle, 0deg),
    transparent 0deg,
    var(--primary-color) 60deg,
    var(--accent-color) 120deg,
    transparent 180deg,
    transparent 360deg
  );
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 0;
  pointer-events: none;
  filter: blur(4px);
}

.book-card:hover .cover-wrapper::before {
  opacity: 1;
  animation: rotateBorder 2s linear infinite;
}

@keyframes rotateBorder {
  from { --angle: 0deg; }
  to   { --angle: 360deg; }
}

.cover {
  position: absolute;
  inset: 0;
  background: var(--bg-color);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-placeholder {
  font-size: 3rem;
  opacity: 0.4;
}

.title {
  font-size: 0.82rem;
  color: var(--text-secondary);
  text-align: center;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-weight: 400;
  letter-spacing: 0;
  padding: 0 2px;
  line-height: 1.4;
}

.book-card.select-mode {
  cursor: pointer;
}

.book-card.selected .cover {
  border: 3px solid hsla(from var(--accent-color) h s l / 0.7);
  box-shadow: var(--shadow-lg);
}

.select-checkbox {
  position: absolute;
  top: 12px;
  left: 12px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.95);
  border: 2px solid var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
  color: var(--primary-color);
  font-weight: 600;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-fast);
}

.book-card.selected .select-checkbox {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  transform: scale(1.15);
}

.read-status {
  position: absolute;
  bottom: 12px;
  left: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  background: color-mix(in srgb, var(--primary-color) 50%, var(--accent-color) 50%);
  font-size: 0.75rem;
  color: white;
  font-weight: 500;
}

.reading-progress {
  position: absolute;
  bottom: 12px;
  left: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  font-size: 0.75rem;
  color: white;
  font-weight: 500;
}

.download-status {
  position: absolute;
  bottom: 12px;
  right: 12px;
  width: 34px;
  height: 34px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(8px);
}

.download-status.not-downloaded {
  cursor: pointer;
  transition: all var(--transition-fast);
}

.download-status.not-downloaded:hover {
  background: var(--primary-color);
  transform: scale(1.15);
}

.download-status.downloading {
  animation: pulse 1.5s ease-in-out infinite;
}

.spinner {
  font-size: 1.1rem;
}

.bottom-action-bar {
  position: fixed;
  bottom: 24px;
  left: 240px;
  right: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 200;
  pointer-events: none;
}

.bottom-action-bar .action-bar-content {
  position: relative;
  width: fit-content;
  min-width: 400px;
  max-width: 900px;
  background: var(--sidebar-bg);
  border-radius: var(--radius-xl);
  padding: 14px 28px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12), 0 2px 8px rgba(0, 0, 0, 0.06);
  animation: slideUp var(--transition-normal);
  backdrop-filter: blur(20px) saturate(180%);
  pointer-events: auto;
}

.bottom-action-bar .action-bar-content::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: calc(var(--radius-xl) + 2px);
  background: conic-gradient(
    from var(--angle, 0deg),
    transparent 0deg,
    var(--primary-color) 60deg,
    var(--accent-color) 120deg,
    transparent 180deg,
    transparent 360deg
  );
  opacity: 0.6;
  z-index: -1;
  pointer-events: none;
  filter: blur(3px);
  animation: rotateBorder 3s linear infinite;
}

.bottom-action-bar .action-bar-content::after {
  content: '';
  position: absolute;
  inset: 2px;
  border-radius: var(--radius-xl);
  background: var(--sidebar-bg);
  z-index: -1;
  pointer-events: none;
}

.action-bar-info {
  color: var(--text-primary);
  font-size: 0.95rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
  letter-spacing: -0.01em;
  flex-shrink: 0;
  min-width: 160px;
  white-space: nowrap;
}

.action-bar-info::before {
  content: '';
  width: 4px;
  height: 20px;
  background: linear-gradient(180deg, var(--primary-color) 0%, var(--accent-color) 100%);
  border-radius: 2px;
}

.action-bar-buttons {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.action-bar-buttons .btn {
  padding: 8px 18px;
  border-radius: var(--radius-lg);
  font-weight: 600;
  font-size: 0.85rem;
  transition: all var(--transition-normal);
  letter-spacing: -0.01em;
  position: relative;
  overflow: hidden;
}

.action-bar-buttons .btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.3) 50%, transparent 100%);
  transition: left var(--transition-slow);
}

.action-bar-buttons .btn:hover::before {
  left: 100%;
}

.action-bar-buttons .btn.danger {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: #FFFFFF;
  border: none;
  box-shadow: 0 4px 14px rgba(99, 102, 241, 0.35);
  position: relative;
  z-index: 1;
}

.action-bar-buttons .btn.danger:hover:not(:disabled) {
  background: linear-gradient(135deg, var(--accent-color) 0%, var(--primary-color) 100%);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(99, 102, 241, 0.45);
}

.action-bar-buttons .btn.danger:active:not(:disabled) {
  transform: translateY(-1px);
}

.action-bar-buttons .btn.danger:disabled {
  background: linear-gradient(135deg, #E5E7EB 0%, #D1D5DB 100%);
  color: #9CA3AF;
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}

.action-bar-buttons .btn.secondary {
  background: transparent;
  color: var(--text-primary);
  border: 2px solid var(--border-color);
  position: relative;
  z-index: 1;
}

.action-bar-buttons .btn.secondary:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.action-bar-buttons .btn.secondary:active {
  transform: translateY(-1px);
}

/* 状态菜单 */
.status-menu-container {
  position: relative;
}

.status-menu {
  position: absolute;
  bottom: 100%;
  left: 0;
  margin-bottom: 8px;
  background: var(--sidebar-bg);
  border-radius: var(--radius-lg);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  z-index: 100;
  border: 1px solid var(--border-color);
}

.status-menu-item {
  padding: 8px 20px;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 0.85rem;
  font-weight: 500;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: left;
  white-space: nowrap;
}

.status-menu-item:hover {
  background: var(--primary-light);
  color: var(--primary-color);
}

.content-area::-webkit-scrollbar {
  width: 6px;
}

.content-area::-webkit-scrollbar-track {
  background: transparent;
}

.content-area::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.content-area::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

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
  transition: all 0.15s;
}

.context-menu-item:hover {
  background: var(--primary-light);
  color: var(--primary-color);
}

.context-menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger-color, #EF4444);
}

.context-menu-divider {
  height: 1px;
  background: var(--border-color);
  margin: 4px 0;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.dialog {
  background: var(--bg-color);
  border-radius: var(--radius-xl);
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
  min-width: 400px;
  max-width: 90vw;
  animation: dialogIn 0.2s ease-out;
}

@keyframes dialogIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.dialog-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.dialog-content {
  padding: 24px;
}

.dialog-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.group-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 2px solid transparent;
}

.group-option:hover {
  background: var(--primary-light);
}

.group-option.active {
  border-color: var(--primary-color);
  background: rgba(79, 70, 229, 0.1);
}

.group-name {
  flex: 1;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-count {
  color: var(--text-muted);
  font-size: 0.85rem;
}

/* 分组创建按钮 */
.group-add-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 12px;
  padding: 8px 16px;
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-lg);
  cursor: pointer;
  color: var(--primary-color);
  font-size: 14px;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.group-add-row:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  border-style: solid;
}

.group-empty {
  text-align: center;
  color: var(--text-muted);
  padding: 20px;
  font-size: 13px;
}
</style>
