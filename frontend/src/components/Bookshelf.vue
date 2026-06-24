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
import PlusIcon from './icons/PlusIcon.vue'
import BookGroup from './BookGroup.vue'
import CustomModal from './CustomModal.vue'
import CustomModalEx from './CustomModalEx.vue'
import type { Group } from '../stores/library'

const emit = defineEmits<{
  (e: 'open-book', book: any): void
  (e: 'toggle-webdav'): void
  (e: 'open-book-detail', book: any): void
  (e: 'open-group', group: Group): void
  (e: 'close-tab-menu'): void
}>()

const store = useLibraryStore()
const settingsStore = useSettingsStore()

// WebDAV侧边栏状态
const isWebdavOpen = ref(false)

// WebDAV下载状态
const isDownloading = ref(false)
const downloadStatus = ref('')
const downloadStatusType = ref<'success' | 'error' | ''>('')

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

// 当前书架（从 store 获取）
const currentShelf = computed(() => store.activeShelf)

// 当前书架的书籍列表（按需加载）
const books = computed(() => store.currentBooks)

// 当前书架的分组列表
const groups = computed(() => store.currentGroups)

// 根级别的所有项目（书籍 + 分组）
const rootItems = computed(() => store.rootItems)

const sortItems = (items: any[]) => {
  const sortBy = settingsStore.sortBy
  if (sortBy === 'default') return items
  
  return [...items].sort((a, b) => {
    if (a.type === 'group' && b.type !== 'group') return -1
    if (a.type !== 'group' && b.type === 'group') return 1
    
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

const filteredItems = computed(() => {
  if (!searchKeyword.value.trim()) {
    return sortItems(rootItems.value)
  }
  
  const keyword = searchKeyword.value.toLowerCase().trim()
  const results: any[] = []
  const addedBookIds = new Set<string>()
  
  rootItems.value.forEach(item => {
    if (item.type === 'group') {
      const name = (item.name || '').toLowerCase()
      if (name.includes(keyword)) {
        results.push(item)
      }
      const booksInGroup = books.value.filter((b: any) => b.groupId === item.id)
      booksInGroup.forEach((book: any) => {
        const title = (book.title || '').toLowerCase()
        const author = (book.author || '').toLowerCase()
        if (title.includes(keyword) || author.includes(keyword)) {
          if (!addedBookIds.has(book.id)) {
            results.push({ ...book, type: 'book', inGroup: item.name })
            addedBookIds.add(book.id)
          }
        }
      })
    } else {
      const title = (item.title || '').toLowerCase()
      const author = (item.author || '').toLowerCase()
      if (title.includes(keyword) || author.includes(keyword)) {
        if (!addedBookIds.has(item.id)) {
          results.push(item)
          addedBookIds.add(item.id)
        }
      }
    }
  })
  
  return sortItems(results)
})

// 只获取当前的书籍（不包括分组）
const currentBooksOnly = computed(() => {
  return filteredItems.value.filter(item => item.type !== 'group')
})

// 计算书架的总书籍数（包括分组里的）
const totalBooksCount = computed(() => {
  return books.value.length
})

// 显示创建分组对话框
const showCreateGroupDialog = ref(false)
const newGroupName = ref('')
const createGroupReturnToMove = ref(false)  // 创建后是否返回移动面板

// 显示移动到分组对话框
const showMoveToGroupDialog = ref(false)

// 显示重命名分组对话框
const showRenameGroupDialog = ref(false)
const renameGroupName = ref('')
const renameGroupId = ref('')  // 当前要重命名的分组ID
const selectedTargetGroupId = ref<string | null>(null)

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

// ========== 快捷键功能 ==========

// 检查当前焦点是否在输入框内
const isInputFocused = () => {
  const activeElement = document.activeElement
  return activeElement?.tagName === 'INPUT' || activeElement?.tagName === 'TEXTAREA'
}

// 处理键盘事件
const handleKeyDown = (e: KeyboardEvent) => {
  // Esc - 退出选择模式
  if (e.key === 'Escape' && isSelectMode.value) {
    e.preventDefault()
    exitSelectMode()
    console.log('快捷键: Esc - 退出选择模式')
  }
  
  // Ctrl+A - 全选书籍（自动进入选择模式），但如果焦点在输入框内则不处理
  if (e.ctrlKey && e.key === 'a' && !isInputFocused()) {
    e.preventDefault()
    if (!isSelectMode.value) {
      enterSelectMode()
    }
    selectAllBooks()
    console.log('快捷键: Ctrl+A - 全选书籍')
  }
  
  // Ctrl+D - 删除选中书籍（仅在选择模式下）
  if (e.ctrlKey && e.key === 'd' && isSelectMode.value && selectedBooks.value.size > 0) {
    e.preventDefault()
    handleBatchDelete()
    console.log('快捷键: Ctrl+D - 删除选中书籍')
  }
}

// 监听键盘事件
onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

// 进入选择模式
const enterSelectMode = () => {
  isSelectMode.value = true
  selectedBooks.value.clear()
}

// 切换选择模式
const toggleSelectMode = () => {
  if (isSelectMode.value) {
    exitSelectMode()
  } else {
    enterSelectMode()
  }
}

// 退出选择模式
const exitSelectMode = () => {
  isSelectMode.value = false
  selectedBooks.value.clear()
}

// 处理书籍点击（支持 Ctrl+鼠标左键进入选择模式）
const handleBookClick = (event: MouseEvent, book: any) => {
  // Ctrl+鼠标左键 - 进入选择模式并选中当前书籍
  if (event.ctrlKey && event.button === 0) {
    event.preventDefault()
    if (!isSelectMode.value) {
      enterSelectMode()
    }
    toggleBookSelection(book.id)
    console.log('快捷键: Ctrl+鼠标左键 - 进入选择模式并选中')
    return
  }
  
  // 普通点击 - 根据模式执行不同操作
  if (isSelectMode.value) {
    toggleBookSelection(book.id)
  } else {
    openBook(book)
  }
}

// 全选所有书籍
const selectAllBooks = () => {
  filteredItems.value.forEach(item => {
    if (item.type !== 'group') {
      selectedBooks.value.add(item.id)
    }
  })
}

// 取消全选
const deselectAllBooks = () => {
  selectedBooks.value.clear()
}

// 切换书籍选中状态
const toggleBookSelection = (bookId: string) => {
  if (selectedBooks.value.has(bookId)) {
    selectedBooks.value.delete(bookId)
  } else {
    selectedBooks.value.add(bookId)
  }
}

// 批量删除选中书籍
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

// 切换WebDAV侧边栏
const toggleWebdav = () => {
  isWebdavOpen.value = !isWebdavOpen.value
  emit('toggle-webdav')
}

// 从云端下载（WebDAV）
const handleDownloadFromCloud = async () => {
  if (!store.activeShelfId) {
    downloadStatus.value = '请先选择一个书架'
    downloadStatusType.value = 'error'
    return
  }
  
  isDownloading.value = true
  downloadStatus.value = ''
  try {
    // @ts-ignore
    const res = await window.go.main.App.DownloadShelf(store.activeShelfId)
    downloadStatus.value = '下载成功: ' + res
    downloadStatusType.value = 'success'
    
    // 下载成功后刷新书架
    store.scanShelves()
    if (store.activeShelfId) {
      await store.loadShelfBooks(store.activeShelfId)
    }
  } catch (e) {
    downloadStatus.value = '下载失败: ' + (e as Error).message
    downloadStatusType.value = 'error'
  } finally {
    isDownloading.value = false
  }
}

// 切换导入菜单
const toggleImportMenu = () => {
  showImportMenu.value = !showImportMenu.value
}

// 关闭导入菜单
const closeImportMenu = () => {
  showImportMenu.value = false
}

// 点击外部关闭菜单
const handleClickOutside = (event: MouseEvent) => {
  if (importMenuRef.value && !importMenuRef.value.contains(event.target as Node)) {
    showImportMenu.value = false
  }
}

// 处理导入结果
const processImportResult = async (result: ImportResult) => {
  if (!result.success) {
    console.error('导入书籍失败:', result.error)
    return
  }
  
  const shelfName = currentShelf.value?.name || '默认书架'
  
  // 添加到当前书架
  const added = await store.addBook(result.title, result.coverUrl, result.md5, result.filePath, result.author)
  
  if (added) {
    console.log('书籍导入成功:', result.title)
    // 触发后台自动同步
    // @ts-ignore
    window.go.main.App.TriggerAutoSync(shelfName)
  } else {
    console.log('书籍已存在，跳过:', result.title)
  }
}

// 1. 单文件导入
const handleSingleFileImport = async () => {
  closeImportMenu()
  
  try {
    // @ts-ignore
    const filePath = await window.go.main.App.SelectEpubFiles()
    
    if (!filePath) {
      return
    }

    isImporting.value = true
    const shelfName = currentShelf.value?.name || '默认书架'
    const result = await importBook(filePath, shelfName)
    await processImportResult(result)
  } catch (error) {
    console.error('导入书籍失败:', error)
    showConfirmModal('导入失败', '导入书籍失败，请重试', 'error', { showCancel: false })
  } finally {
    isImporting.value = false
  }
}

// 2. 多文件导入
const handleMultipleFilesImport = async () => {
  closeImportMenu()
  
  try {
    // @ts-ignore
    const filePaths: string[] = await window.go.main.App.SelectMultipleEpubFiles()
    
    if (!filePaths || filePaths.length === 0) {
      return
    }

    isImporting.value = true
    const shelfName = currentShelf.value?.name || '默认书架'
    
    const results = await importBooks(filePaths, shelfName, (current, total, title) => {
      console.log(`导入进度: ${current}/${total} - ${title}`)
    })
    
    let successCount = 0
    let skipCount = 0
    let failCount = 0
    
    for (const result of results) {
      if (result.success) {
        const added = await store.addBook(result.title, result.coverUrl, result.md5, result.filePath, result.author)
        if (added) {
          successCount++
        } else {
          skipCount++
        }
      } else {
        failCount++
        console.error('导入失败:', result.error)
      }
    }
    
    // 触发同步
    // @ts-ignore
    window.go.main.App.TriggerAutoSync(shelfName)
    
    showConfirmModal('导入完成', `成功: ${successCount}，跳过: ${skipCount}，失败: ${failCount}`, 
      failCount > 0 ? 'warning' : 'success', { showCancel: false })
  } catch (error) {
    console.error('批量导入书籍失败:', error)
    showConfirmModal('导入失败', '批量导入书籍失败，请重试', 'error', { showCancel: false })
  } finally {
    isImporting.value = false
  }
}

// 3. 文件夹导入
const handleFolderImport = async () => {
  closeImportMenu()
  
  try {
    // @ts-ignore
    const folderPath = await window.go.main.App.SelectEpubFolder()
    
    if (!folderPath) {
      return
    }

    isImporting.value = true
    const shelfName = currentShelf.value?.name || '默认书架'
    
    const results = await importBooksFromFolder(folderPath, shelfName, (current, total, title) => {
      console.log(`导入进度: ${current}/${total} - ${title}`)
    })
    
    let successCount = 0
    let skipCount = 0
    let failCount = 0
    
    for (const result of results) {
      if (result.success) {
        const added = await store.addBook(result.title, result.coverUrl, result.md5, result.filePath, result.author)
        if (added) {
          successCount++
        } else {
          skipCount++
        }
      } else {
        failCount++
        console.error('导入失败:', result.error)
      }
    }
    
    // 触发同步
    // @ts-ignore
    window.go.main.App.TriggerAutoSync(shelfName)
    
    showConfirmModal('导入完成', `成功: ${successCount}，跳过: ${skipCount}，失败: ${failCount}`, 
      failCount > 0 ? 'warning' : 'success', { showCancel: false })
  } catch (error) {
    console.error('文件夹导入书籍失败:', error)
    showConfirmModal('导入失败', '文件夹导入书籍失败，请重试', 'error', { showCancel: false })
  } finally {
    isImporting.value = false
  }
}

// 获取书籍的唯一标识
const getBookKey = (book: any): string => {
  return `${currentShelf.value?.name || ''}:${book.md5}`
}

// 检查书籍是否已下载
const isBookDownloaded = async (book: any): Promise<boolean> => {
  if (!book.filePath) return false
  // @ts-ignore
  return await window.go.main.App.FileExists(book.filePath)
}

// 批量检查书籍下载状态
const checkAllBooksStatus = async () => {
  if (!currentShelf.value) return
  
  // 遍历所有项目（书籍和分组）
  for (const item of filteredItems.value) {
    // 只处理书籍
    if (item.type !== 'group') {
      const key = getBookKey(item)
      const exists = await isBookDownloaded(item)
      bookDownloadedStatus.value.set(key, exists)
    }
  }
}

// 更新单本书的下载状态
const updateBookStatus = (book: any, downloaded: boolean) => {
  const key = getBookKey(book)
  bookDownloadedStatus.value.set(key, downloaded)
}

// 下载书籍
const downloadBook = async (book: any) => {
  if (!book.md5 || !book.filePath || !currentShelf.value) return
  
  const bookKey = getBookKey(book)
  if (downloadingBooks.value.has(bookKey)) {
    console.log('该书籍正在下载中')
    return
  }
  
  // 检查本地是否已有该书籍文件
  const exists = await isBookDownloaded(book)
  if (exists) {
    console.log('该书籍已存在，跳过下载:', book.title)
    updateBookStatus(book, true)
    return
  }
  
  downloadingBooks.value.add(bookKey)
  
  try {
    const fileName = book.filePath.split('/').pop() || book.filePath.split('\\').pop()
    // @ts-ignore
    await window.go.main.App.DownloadSingleEpub(currentShelf.value.name, book.md5, fileName)
    console.log('书籍下载完成:', book.title)
    // 更新下载状态
    updateBookStatus(book, true)
  } catch (error) {
    console.error('下载失败:', error)
  } finally {
    downloadingBooks.value.delete(bookKey)
  }
}

// 打开书籍（如果未下载则先下载）
const openBook = async (book: any) => {
  if (!book.filePath || !currentShelf.value) return
  
  const bookKey = getBookKey(book)
  
  // 检查是否已下载
  const exists = bookDownloadedStatus.value.get(bookKey) || await isBookDownloaded(book)
  
  if (!exists) {
    // 先下载
    if (downloadingBooks.value.has(bookKey)) {
      console.log('该书籍正在下载中')
      return
    }
    
    downloadingBooks.value.add(bookKey)
    
    try {
      const fileName = book.filePath.split('/').pop() || book.filePath.split('\\').pop()
      // @ts-ignore
      await window.go.main.App.DownloadSingleEpub(currentShelf.value.name, book.md5, fileName)
      console.log('书籍下载完成，准备打开:', book.title)
      // 更新下载状态
      updateBookStatus(book, true)
    } catch (error) {
      console.error('下载失败:', error)
      downloadingBooks.value.delete(bookKey)
      return
    } finally {
      downloadingBooks.value.delete(bookKey)
    }
  }
  
  // 将书籍移到书架最前面（最近使用排序）
  await store.moveBookToFront(book.id)
  
  // 打开书籍
  emit('open-book', book)
}

// 监听书架变化，重新检查书籍状态
watch([currentShelf, books], async () => {
  await checkAllBooksStatus()
}, { immediate: true })

onMounted(async () => {
  await checkAllBooksStatus()
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('click', handleClickOutsideContextMenu)
  document.addEventListener('click', handleClickOutsideGroupContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('click', handleClickOutsideContextMenu)
  document.removeEventListener('click', handleClickOutsideGroupContextMenu)
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

// 显示右键菜单
const showContextMenu = (event: MouseEvent, bookId: string) => {
  event.preventDefault()
  event.stopPropagation()
  
  // 在选择模式下禁用右键菜单
  if (isSelectMode.value) return
  
  // 关闭其他右键菜单
  closeGroupContextMenu()
  closeContextMenu()
  emit('close-tab-menu')
  
  contextMenu.value = {
    show: true,
    x: event.clientX,
    y: event.clientY,
    bookId
  }
}

// 关闭右键菜单
const closeContextMenu = () => {
  contextMenu.value.show = false
}

// 处理删除
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

// 处理查看书籍详情
const handleContextMenuDetail = () => {
  if (!contextMenu.value.bookId) return
  
  // 找到对应的书籍对象
  const book = currentBooksOnly.value.find((b: any) => b.id === contextMenu.value.bookId)
  if (book) {
    emit('open-book-detail', book)
  }
  
  closeContextMenu()
}

// 打开文件所在位置
const handleOpenFileLocation = async () => {
  if (!contextMenu.value.bookId) return
  
  // 找到对应的书籍对象
  const book = currentBooksOnly.value.find((b: any) => b.id === contextMenu.value.bookId)
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

// 处理移动到分组
const handleContextMenuMoveToGroup = () => {
  if (!contextMenu.value.bookId) return
  
  // 清空之前的选择，只选中当前书籍
  selectedBooks.value.clear()
  selectedBooks.value.add(contextMenu.value.bookId)
  
  // 打开移动到分组面板
  showMoveToGroupDialog.value = true
  
  closeContextMenu()
}

// 重命名相关状态
const isRenaming = ref(false)
const renamingBookId = ref('')
const renamingTitle = ref('')

// 开始重命名书籍
const startRenameBook = async () => {
  if (!contextMenu.value.bookId) return
  
  const book = currentBooksOnly.value.find((b: any) => b.id === contextMenu.value.bookId)
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
    const book = currentBooksOnly.value.find((b: any) => b.id === renamingBookId.value)
    
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

// ============ 分组功能 ============

// 打开创建分组对话框
const handleCreateGroupClick = () => {
  newGroupName.value = ''
  createGroupReturnToMove.value = false
  showCreateGroupDialog.value = true
}

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
    showConfirmModal('提示', '请输入分组名称', 'warning', { showCancel: false })
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

// 点击分组
const handleGroupClick = (group: Group) => {
  emit('open-group', group)
}

// 分组右键菜单
const groupContextMenu = ref({
  show: false,
  x: 0,
  y: 0,
  groupId: ''
})

const showGroupContextMenu = (event: MouseEvent, groupId: string) => {
  event.preventDefault()
  event.stopPropagation()
  
  if (isSelectMode.value) return
  
  // 关闭其他右键菜单
  closeContextMenu()
  closeGroupContextMenu()
  emit('close-tab-menu')
  
  groupContextMenu.value = {
    show: true,
    x: event.clientX,
    y: event.clientY,
    groupId
  }
}

const closeGroupContextMenu = () => {
  groupContextMenu.value.show = false
}

const handleContextMenuRenameGroup = () => {
  if (!groupContextMenu.value.groupId) return
  
  const group = groups.value.find(g => g.id === groupContextMenu.value.groupId)
  if (!group) return
  
  // 打开重命名对话框
  renameGroupId.value = group.id
  renameGroupName.value = group.name
  showRenameGroupDialog.value = true
  
  closeGroupContextMenu()
}

const handleConfirmRenameGroup = () => {
  if (renameGroupName.value.trim() && renameGroupId.value) {
    store.renameGroup(renameGroupId.value, renameGroupName.value.trim())
  }
  showRenameGroupDialog.value = false
  renameGroupName.value = ''
  renameGroupId.value = ''
}

const handleCancelRenameGroup = () => {
  showRenameGroupDialog.value = false
  renameGroupName.value = ''
  renameGroupId.value = ''
}

const handleContextMenuDeleteGroup = async () => {
  if (!groupContextMenu.value.groupId) return
  
  const group = groups.value.find(g => g.id === groupContextMenu.value.groupId)
  if (!group) return
  
  closeGroupContextMenu()
  
  const bookCount = books.value.filter(b => b.groupId === group.id).length
  const message = bookCount > 0 
    ? `确定要删除分组「${group.name}」吗？分组内的 ${bookCount} 本书籍将移回根目录。`
    : `确定要删除分组「${group.name}」吗？`
    
  showConfirmModal(
    '确认删除分组',
    message,
    'warning',
    { onConfirm: async () => {
      await store.deleteGroup(group.id)
    }}
  )
}

const handleClickOutsideGroupContextMenu = (event: MouseEvent) => {
  if (contextMenuRef.value && !contextMenuRef.value.contains(event.target as Node)) {
    closeGroupContextMenu()
  }
}

// 点击外部关闭菜单
const handleClickOutsideContextMenu = (event: MouseEvent) => {
  if (contextMenuRef.value && !contextMenuRef.value.contains(event.target as Node)) {
    closeContextMenu()
  }
}
</script>

<template>
  <main class="bookshelf-container">
    <header class="topbar">
      <div class="topbar-left">
        <input 
          type="text" 
          class="search-input" 
          v-model="searchKeyword"
          :placeholder="`在 ${totalBooksCount} 本书籍中搜索...`"
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
          @click="handleCreateGroupClick"
        >
          <PlusIcon :size="16" />新建分组
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
              <FolderIcon :size="18" class="item-icon" />
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
      <div v-if="filteredItems.length === 0" class="empty-state">
        <div class="empty-text">
          {{ searchKeyword ? '没有找到匹配的项目' : '书架空空如也' }}
        </div>
        <div v-if="!searchKeyword" class="empty-buttons">
          <button class="btn primary" @click="handleSingleFileImport" :disabled="isImporting">
            {{ isImporting ? '导入中...' : '导入你的第一本书' }}
          </button>
        </div>
        <button v-else class="btn secondary" @click="searchKeyword = ''">
          清除搜索
        </button>
      </div>
      
      <div v-else class="book-grid" :style="{ gridTemplateColumns: `repeat(${settingsStore.bookshelfColumns}, minmax(0, 1fr))`, gap: `${settingsStore.coverGap}px` }">
        <template v-for="item in filteredItems" :key="item.id">
          <!-- 分组卡片 -->
          <BookGroup 
            v-if="item.type === 'group'" 
            :group="item"
            @click="!isSelectMode && handleGroupClick(item)"
            @contextmenu="!isSelectMode && showGroupContextMenu($event, item.id)"
            :class="{ 'opacity-50': isSelectMode }"
          />
          
          <!-- 书籍卡片 -->
          <div 
            v-else
            :class="['book-card', { selected: selectedBooks.has(item.id), 'select-mode': isSelectMode }]"
            @click="handleBookClick($event, item)"
            @contextmenu="showContextMenu($event, item.id)"
          >
            <div class="cover-wrapper">
              <div class="cover">
                <img v-if="item.coverUrl" :src="item.coverUrl" alt="封面" class="cover-image" />
                <div v-else class="cover-placeholder"><BookIcon :size="32" /></div>
                
                <div v-if="isSelectMode" class="select-checkbox">
                  <span v-if="selectedBooks.has(item.id)">✓</span>
                </div>
                
                <div v-if="!isSelectMode && downloadingBooks.has(getBookKey(item))" class="download-status downloading">
                  <span class="spinner">⏳</span>
                </div>
                <div 
                  v-else-if="!isSelectMode && bookDownloadedStatus.get(getBookKey(item)) === false" 
                  class="download-status not-downloaded"
                  @click.stop="downloadBook(item)"
                >
                  <DownloadIcon :size="16" title="下载书籍" />
                </div>
              </div>
            </div>
            <div class="title">{{ item.title }}</div>
            <div v-if="item.inGroup" class="group-indicator">
              在「{{ item.inGroup }}」中
            </div>
          </div>
        </template>
      </div>
    </div>
    
    <!-- 书籍右键菜单 -->
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
      
      <button class="context-menu-item" @click="handleOpenFileLocation">
        打开当前位置
      </button>
      
      <button class="context-menu-item danger" @click="handleContextMenuDelete">
        删除书籍
      </button>
    </div>
    
    <!-- 分组右键菜单 -->
    <div 
      v-show="groupContextMenu.show" 
      class="context-menu"
      ref="contextMenuRef"
      :style="{ left: groupContextMenu.x + 'px', top: groupContextMenu.y + 'px' }"
      @click.stop
    >
      <button class="context-menu-item" @click="handleContextMenuRenameGroup">
        重命名
      </button>
      
      <button class="context-menu-item danger" @click="handleContextMenuDeleteGroup">
        删除分组
      </button>
    </div>
    
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
    
    <!-- 重命名分组对话框 -->
    <CustomModalEx
      :visible="showRenameGroupDialog"
      title="重命名分组"
      confirmText="确定"
      cancelText="取消"
      :confirmDisabled="!renameGroupName.trim()"
      @confirm="handleConfirmRenameGroup"
      @cancel="handleCancelRenameGroup"
    >
      <input 
        type="text" 
        v-model="renameGroupName"
        class="modal-input"
        placeholder="请输入分组名称"
        @keyup.enter="handleConfirmRenameGroup"
        autofocus
      />
    </CustomModalEx>
    
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
      title="书籍分组"
      confirmText="移动"
      cancelText="取消"
      :confirmDisabled="!selectedTargetGroupId"
      @confirm="handleMoveToGroup"
      @cancel="showMoveToGroupDialog = false"
    >
      <div class="group-list">
        <div 
          v-for="group in store.currentGroups" 
          :key="group.id"
          class="group-option"
          :class="{ active: selectedTargetGroupId === group.id }"
          @click="selectedTargetGroupId = group.id"
        >
          <FolderIcon :size="18" />
          <span class="group-name">{{ group.name }}</span>
          <span class="group-count">({{ getGroupBookCount(group.id) }} 本书)</span>
        </div>
        <div 
          v-if="store.currentGroups.length === 0"
          class="group-empty"
        >
          暂无分组
        </div>
      </div>
      <div class="group-add-row" @click="handleCreateFromMove">
        <PlusIcon :size="16" />
        <span>新建分组</span>
      </div>
    </CustomModalEx>
    
    <!-- 底部悬浮操作栏 -->
    <div v-if="isSelectMode" class="bottom-action-bar">
      <div class="action-bar-content">
        <div class="action-bar-info">
          已选择 {{ selectedBooks.size }}/{{ currentBooksOnly.length }} 本书籍
        </div>
        <div class="action-bar-buttons">
          <button 
            v-if="selectedBooks.size < currentBooksOnly.length" 
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
          <button 
            v-if="store.currentGroups.length > 0"
            class="btn secondary" 
            @click="showMoveToGroupDialog = true"
            :disabled="selectedBooks.size === 0"
          >
            移动到分组 
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
/* 书架容器 */
.bookshelf-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color);
  min-height: 0;
  position: relative;
}

/* 顶部操作栏 */
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
  flex-shrink: 0;
}

.topbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  align-items: center;
}

/* 搜索框 */
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

/* 按钮基础样式 */
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

/* 导入下拉菜单 */
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

.dropdown-arrow {
  font-size: 0.6rem;
  transition: transform var(--transition-fast);
}

.import-btn:hover .dropdown-arrow,
.import-btn:focus .dropdown-arrow {
  transform: rotate(180deg);
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

/* 内容区域 */
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

.empty-buttons {
  display: flex;
  gap: 12px;
}

.context-menu-submenu-label {
  padding: 8px 16px;
  font-size: 0.75rem;
  color: var(--text-muted);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
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

.dialog-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background: var(--bg-color);
  color: var(--text-color);
  font-size: 1rem;
  outline: none;
  transition: all var(--transition-fast);
}

.dialog-input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 4px var(--primary-light);
}

.dialog-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 分组列表 */
.group-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding:8px 16px;
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
  justify-content: left;
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

/* 书籍网格 */
.book-grid {
  display: grid;
  gap: 28px 24px;
}

/* 书籍卡片 */
.book-card {
  display: flex;
  flex-direction: column;
  cursor: pointer;
  position: relative;
  border-radius: var(--radius-xl);
}

.book-card:hover .delete-btn {
  opacity: 1;
  transform: scale(1);
}

/* 封面包裹层（用于粒子动画） */
.cover-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 2/3;
  margin-bottom: 10px;
  border-radius: 10px;
}

/* 粒子旋转边框动画层 */
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

@property --angle {
  syntax: '<angle>';
  initial-value: 0deg;
  inherits: false;
}

@keyframes rotateBorder {
  from { --angle: 0deg; }
  to   { --angle: 360deg; }
}

/* 封面容器 */
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

/* 书名 */
.title {
  font-size: 0.82rem;
  color: var(--text-secondary);
  text-align: center;
  font-weight: 400;
  letter-spacing: 0;
  padding: 0 2px;
  line-height: 1.4;
  overflow: hidden;

  /* WebKit 内核（Chrome/Edge/Safari） */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;

  /* W3C 标准属性（Firefox 等，消除编辑器警告） */
  display: box;
  line-clamp: 2;
  box-orient: vertical;
}

/* 分组指示器 */
.group-indicator {
  font-size: 0.7rem;
  color: var(--primary-color);
  text-align: center;
  margin-top: 4px;
  opacity: 0.8;
  background: rgba(79, 70, 229, 0.1);
  padding: 2px 8px;
  border-radius: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 选择模式下分组的样式 */
.opacity-50 {
  opacity: 0.5;
  pointer-events: none;
}

/* 删除按钮 */
.delete-btn {
  position: absolute;
  top: -12px;
  right: -12px;
  background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%);
  color: white;
  border: none;
  border-radius: 50%;
  width: 32px;
  height: 32px;
  cursor: pointer;
  opacity: 0;
  transform: scale(0.7);
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  box-shadow: var(--shadow-md);
}

.delete-btn:hover {
  background: linear-gradient(135deg, #DC2626 0%, #B91C1C 100%);
  transform: scale(1.1);
}

/* 选择模式 */
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

/* 下载状态图标 */
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

/* 底部悬浮操作栏 */
.bottom-action-bar {
  position: fixed;
  bottom: 24px;
  left: 240px; /* 从侧边栏右侧开始 */
  right: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 200;
  pointer-events: none; /* 让容器不阻挡点击 */
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
  pointer-events: auto; /* 恢复内容区域的点击 */
}

/* 粒子旋转边框动画层 */
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

/* 内部遮罩，只显示边框效果 */
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

/* 滚动条美化 */
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

</style>
