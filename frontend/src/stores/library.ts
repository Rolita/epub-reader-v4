import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface Shelf {
  id: string
  name: string
}

export interface Book {
  id: string
  shelfId: string
  groupId?: string
  title: string
  author?: string
  coverUrl?: string
  md5?: string
  filePath?: string
  order: number
}

export interface Group {
  id: string
  shelfId: string
  name: string
  coverUrl?: string
  order: number
}

export const useLibraryStore = defineStore('library', () => {
  // 书架列表
  const shelves = ref<Shelf[]>([])
  const activeShelfId = ref<string | null>(null)
  
  // 当前书架的书籍列表（按需加载）
  const currentBooks = ref<Book[]>([])
  
  // 当前书架的分组列表
  const currentGroups = ref<Group[]>([])
  
  // 当前激活的分组
  const activeGroupId = ref<string | null>(null)
  
  // 当前激活的书架
  const activeShelf = computed(() => {
    return shelves.value.find(s => s.id === activeShelfId.value) || null
  })
  
  // 当前激活的分组
  const activeGroup = computed(() => {
    return currentGroups.value.find(g => g.id === activeGroupId.value) || null
  })
  
  // 当前分组内的书籍
  const groupBooks = computed(() => {
    if (!activeGroupId.value) return []
    return currentBooks.value.filter(b => b.groupId === activeGroupId.value)
  })
  
  // 书架根级别的书籍（不在任何分组内）
  const rootBooks = computed(() => {
    return currentBooks.value.filter(b => !b.groupId)
  })
  
  // 书架根级别的所有项目（书籍 + 分组，混合在一起）
  const rootItems = computed(() => {
    const groupItems = currentGroups.value.map(g => ({ ...g, type: 'group' as const }))
    const bookItems = rootBooks.value.map(b => ({ ...b, type: 'book' as const }))
    return [...groupItems, ...bookItems]
  })

  // ============ 书架操作 ============

  // 扫描所有书架
  async function scanShelves() {
    try {
      // @ts-ignore
      const dataStr = await window.go.main.App.ScanShelves()
      const data = JSON.parse(dataStr)
      shelves.value = data
      
      // 如果没有书架，创建一个默认书架
      if (shelves.value.length === 0) {
        await createShelf('默认书架')
      }
      
      // 设置激活书架
      if (!activeShelfId.value && shelves.value.length > 0) {
        activeShelfId.value = shelves.value[0].id
        await loadShelfBooks(shelves.value[0].id)
      }
    } catch (e) {
      console.error('扫描书架失败', e)
    }
  }

  // 创建新书架
  async function createShelf(name: string) {
    try {
      // @ts-ignore
      await window.go.main.App.CreateShelf(name)
      // 重新扫描书架列表
      await scanShelves()
      // 自动切换到新创建的书架
      activeShelfId.value = name
      currentBooks.value = []
    } catch (e) {
      console.error('创建书架失败', e)
    }
  }

  // 删除书架
  async function deleteShelf(name: string) {
    try {
      // @ts-ignore
      await window.go.main.App.DeleteShelf(name)
      
      // 如果删除的是当前激活的书架
      if (activeShelfId.value === name) {
        // 重新扫描书架列表
        await scanShelves()
        // 切换到第一个书架
        if (shelves.value.length > 0) {
          activeShelfId.value = shelves.value[0].id
          await loadShelfBooks(shelves.value[0].id)
        } else {
          activeShelfId.value = null
          currentBooks.value = []
        }
      } else {
        await scanShelves()
      }
    } catch (e) {
      console.error('删除书架失败', e)
    }
  }

  // 重命名书架（使用原子性操作）
  async function renameShelf(oldName: string, newName: string) {
    if (oldName === newName) return
    
    try {
      // 检查新名称是否已存在
      const existingShelf = shelves.value.find(s => s.id === newName)
      if (existingShelf) {
        window.dispatchEvent(new CustomEvent('show-alert', { 
          detail: { title: '提示', message: '该书架名称已存在！', type: 'warning' }
        }))
        return
      }
      
      // 使用 Go 后端的原子性重命名方法
      // @ts-ignore
      const result = await window.go.main.App.RenameShelf(oldName, newName)
      
      // 重新扫描书架列表
      await scanShelves()
      
      // 如果当前激活的是旧书架，切换到新书架
      if (activeShelfId.value === oldName) {
        activeShelfId.value = newName
        await loadShelfBooks(newName)
      }
    } catch (e) {
      console.error('重命名书架失败', e)
      window.dispatchEvent(new CustomEvent('show-alert', { 
        detail: { title: '错误', message: '重命名失败，可能该名称已存在', type: 'error' }
      }))
    }
  }

  // 切换激活书架
  async function setActiveShelf(shelfId: string) {
    if (activeShelfId.value === shelfId) return
    
    activeShelfId.value = shelfId
    if (shelfId) {
      await loadShelfBooks(shelfId)
    }
  }

  // 重新排序书架
  async function reorderShelves(newOrder: Shelf[]) {
    try {
      // 更新本地状态
      shelves.value = newOrder
      
      // 保存到后端
      const order = newOrder.map(s => s.id)
      // @ts-ignore
      await window.go.main.App.SaveShelfOrder(order)
    } catch (e) {
      console.error('保存书架顺序失败', e)
    }
  }

  // ============ 书籍操作 ============

  // 加载指定书架的书籍和分组
  async function loadShelfBooks(shelfName: string) {
    try {
      const dataStr = await loadShelfData(shelfName)
      const data = JSON.parse(dataStr)
      
      let books = data.books || []
      books.forEach((b: any, index: number) => {
        if (b.order === undefined) {
          b.order = index
        }
      })
      books.sort((a: any, b: any) => a.order - b.order)
      currentBooks.value = books
      
      let groups = data.groups || []
      groups.forEach((g: any, index: number) => {
        if (g.order === undefined) {
          g.order = index
        }
      })
      groups.sort((a: any, b: any) => a.order - b.order)
      currentGroups.value = groups
    } catch (e) {
      console.error('加载书架书籍失败', e)
      currentBooks.value = []
      currentGroups.value = []
    }
  }
  
  // 保存当前书架的数据（书籍 + 分组）
  async function saveShelfDataFull() {
    if (!activeShelfId.value) return
    
    const data = JSON.stringify({ 
      books: currentBooks.value, 
      groups: currentGroups.value 
    })
    await saveShelfData(activeShelfId.value, data)
  }

  // 加载书架数据（内部使用）
  async function loadShelfData(shelfName: string): Promise<string> {
    // @ts-ignore
    return await window.go.main.App.LoadShelfData(shelfName)
  }

  // 保存书架数据（内部使用）
  async function saveShelfData(shelfName: string, data: string) {
    // @ts-ignore
    return await window.go.main.App.SaveShelfData(shelfName, data)
  }

  // 添加书籍到当前书架（带 MD5 去重）
  async function addBook(title: string, coverUrl?: string, md5?: string, filePath?: string, author?: string): Promise<boolean> {
    if (!activeShelfId.value) return false
    
    // 【关键点】循环比对 MD5，防止重复导入
    if (md5) {
      const existingBook = currentBooks.value.find(b => b.md5 === md5)
      if (existingBook) {
        console.log("检测到相同书籍内容，跳过导入:", title)
        return false
      }
    }
    
    const newBook: Book = {
      id: Date.now().toString(),
      shelfId: activeShelfId.value,
      title,
      author,
      coverUrl,
      md5,
      filePath,
      order: currentBooks.value.length
    }
    
    currentBooks.value.push(newBook)
    await saveBooks()
    return true
  }

  // 保存当前书架的书籍（为了向后兼容）
  async function saveBooks() {
    await saveShelfDataFull()
  }
  
  // ============ 分组操作 ============
  
  // 创建分组
  async function createGroup(name: string, coverUrl?: string) {
    if (!activeShelfId.value) return false
    
    const newGroup: Group = {
      id: Date.now().toString(),
      shelfId: activeShelfId.value,
      name,
      coverUrl,
      order: currentGroups.value.length
    }
    
    currentGroups.value.push(newGroup)
    await saveShelfDataFull()
    return true
  }
  
  // 删除分组
  async function deleteGroup(groupId: string) {
    // 将分组内的书籍移回到根级别
    currentBooks.value = currentBooks.value.map(b => {
      if (b.groupId === groupId) {
        return { ...b, groupId: undefined }
      }
      return b
    })
    
    // 删除分组
    currentGroups.value = currentGroups.value.filter(g => g.id !== groupId)
    
    // 重新计算 order 值
    currentGroups.value.forEach((g, index) => {
      g.order = index
    })
    
    // 如果删除的是当前激活的分组，清空激活状态
    if (activeGroupId.value === groupId) {
      activeGroupId.value = null
    }
    
    await saveShelfDataFull()
  }
  
  // 重新排序分组
  async function reorderGroup(fromIndex: number, toIndex: number) {
    if (fromIndex === toIndex) return
    
    const groups = currentGroups.value
    
    const [removed] = groups.splice(fromIndex, 1)
    groups.splice(toIndex, 0, removed)
    
    groups.forEach((g, index) => {
      g.order = index
    })
    
    await saveShelfDataFull()
  }
  
  // 重新排序书籍（在同一分组内或根级别）
  async function reorderBook(fromIndex: number, toIndex: number, groupId?: string) {
    if (fromIndex === toIndex) return
    
    const books = currentBooks.value.filter(b => b.groupId === groupId)
    
    if (books.length <= 1) return
    
    const allBooks = currentBooks.value
    
    const fromBook = books[fromIndex]
    const toBook = books[toIndex]
    
    const fromBookIndex = allBooks.findIndex(b => b.id === fromBook.id)
    const toBookIndex = allBooks.findIndex(b => b.id === toBook.id)
    
    if (fromBookIndex === -1 || toBookIndex === -1) return
    
    const [removed] = allBooks.splice(fromBookIndex, 1)
    allBooks.splice(toBookIndex, 0, removed)
    
    const targetBooks = allBooks.filter(b => b.groupId === groupId)
    targetBooks.forEach((b, index) => {
      b.order = index
    })
    
    await saveShelfDataFull()
  }
  
  // 重命名分组
  async function renameGroup(groupId: string, newName: string) {
    const group = currentGroups.value.find(g => g.id === groupId)
    if (group) {
      group.name = newName
      await saveShelfDataFull()
      return true
    }
    return false
  }
  
  // 切换激活分组
  async function setActiveGroup(groupId: string | null) {
    activeGroupId.value = groupId
  }
  
  // 将书籍添加到分组
  async function addBookToGroup(bookId: string, groupId: string) {
    const book = currentBooks.value.find(b => b.id === bookId)
    const group = currentGroups.value.find(g => g.id === groupId)
    if (book && group) {
      book.groupId = groupId
      await saveShelfDataFull()
      return true
    }
    return false
  }
  
  // 将书籍从分组移回到根级别
  async function removeBookFromGroup(bookId: string) {
    const book = currentBooks.value.find(b => b.id === bookId)
    if (book) {
      book.groupId = undefined
      await saveShelfDataFull()
      return true
    }
    return false
  }

  // 删除书籍
  async function deleteBook(bookId: string) {
    const book = currentBooks.value.find(b => b.id === bookId)
    if (book && book.md5 && activeShelfId.value) {
      try {
        // @ts-ignore
        await window.go.main.App.DeleteBook(activeShelfId.value, book.id, book.md5)
      } catch (e) {
        console.error('删除书籍失败', e)
      }
    }
    
    currentBooks.value = currentBooks.value.filter(b => b.id !== bookId)
    await saveBooks()
  }

  // 更新书名
  async function updateBookTitle(bookId: string, newTitle: string) {
    const book = currentBooks.value.find(b => b.id === bookId)
    if (book) {
      book.title = newTitle
      await saveBooks()
      return true
    }
    return false
  }

  // 将书籍移到书架最前面（最近使用排序）
  async function moveBookToFront(bookId: string) {
    const index = currentBooks.value.findIndex(b => b.id === bookId)
    
    if (index > 0) {
      const book = currentBooks.value[index]
      currentBooks.value.splice(index, 1)
      currentBooks.value.unshift(book)
      await saveBooks()
    }
  }

  return {
    // 状态
    shelves,
    activeShelfId,
    activeShelf,
    currentBooks,
    currentGroups,
    activeGroupId,
    activeGroup,
    groupBooks,
    rootBooks,
    rootItems,
    
    // 书架操作
    scanShelves,
    createShelf,
    deleteShelf,
    renameShelf,
    setActiveShelf,
    reorderShelves,
    
    // 书籍操作
    addBook,
    deleteBook,
    updateBookTitle,
    moveBookToFront,
    reorderBook,
    loadShelfBooks,
    saveBooks,
    
    // 分组操作
    createGroup,
    deleteGroup,
    renameGroup,
    reorderGroup,
    setActiveGroup,
    addBookToGroup,
    removeBookFromGroup
  }
})
