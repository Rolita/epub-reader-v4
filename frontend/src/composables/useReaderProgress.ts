export interface ProgressData {
  cfi: string
  percentage: number
  timestamp: number
}

export interface Bookmark {
  cfi: string
  percentage: number
  timestamp: number
  chapterTitle?: string
  snippet?: string
}

/**
 * 手动保存当前阅读进度
 * @param rendition epub.js rendition 实例
 * @param filePath 书籍文件路径
 * @returns 保存的进度数据，失败返回 null
 */
export async function saveReaderProgress(
  rendition: any,
  filePath: string
): Promise<ProgressData | null> {
  if (!rendition) return null
  try {
    const location = rendition.currentLocation()
    if (!location || !location.start) return null
    const progressData: ProgressData = {
      cfi: location.start.cfi,
      percentage: location.start.percentage || 0,
      timestamp: Date.now()
    }
    // @ts-ignore
    await window.go.main.App.SaveProgress(filePath, JSON.stringify(progressData))
    console.log('进度已手动保存:', progressData)
    return progressData
  } catch (err) {
    console.error('保存进度失败:', err)
    return null
  }
}

/**
 * 恢复到指定 CFI 位置，或从存储中自动恢复
 * @param rendition epub.js rendition 实例
 * @param cfi EPUB CFI 位置标识（可选，不提供则自动从存储恢复）
 * @param onFocus 可选的聚焦回调
 * @param filePath 书籍文件路径（自动恢复时需要）
 */
export async function restoreReaderProgress(
  rendition: any,
  cfi?: string | null,
  onFocus?: () => void,
  filePath?: string
): Promise<void> {
  if (!rendition) return

  try {
    let targetCfi = cfi
    
    // 如果没有提供 cfi，尝试从存储中获取
    if (!targetCfi && filePath) {
      // @ts-ignore
      const progressJson = await window.go.main.App.GetProgress(filePath)
      if (progressJson) {
        const progress = JSON.parse(progressJson)
        targetCfi = progress.cfi
      }
    }
    
    if (targetCfi) {
      await rendition.display(targetCfi)
      onFocus?.()
      console.log('已恢复到进度:', targetCfi)
    }
  } catch (err) {
    console.error('恢复进度失败:', err)
  }
}

/**
 * 根据CFI获取对应段落文字
 * @param rendition epub.js rendition 实例
 * @param cfi CFI字符串
 * @returns 段落文本内容
 */
function getParagraphTextByCfi(rendition: any, cfi: string): string {
  try {
    const range = rendition.getRange ? rendition.getRange(cfi) : null
    if (!range) {
      console.warn('未获取到页面DOM Range，等待渲染完成')
      return ''
    }

    let paraEl = range.startContainer
    let deep = 0
    while (paraEl && deep < 10) {
      const tag = paraEl.nodeName?.toLowerCase() || ''
      if (["p", "div", "section", "h1", "h2", "h3", "h4", "blockquote"].includes(tag)) {
        break
      }
      paraEl = paraEl.parentNode
      deep++
    }

    const rawText = paraEl?.textContent || ''
    const previewText = rawText.replace(/\s+/g, ' ').trim()
    return previewText
  } catch (err) {
    console.error('提取段落失败:', err)
    return ''
  }
}

/**
 * 获取当前章节标题
 * @param rendition epub.js rendition 实例
 * @param location epub.js currentLocation 返回值
 * @returns 章节标题
 */
function getCurrentChapterTitle(rendition: any, location: any): string {
  try {
    if (location && location.start && location.start.section?.title) {
      return location.start.section.title.trim()
    }

    const contents = rendition.getContents ? rendition.getContents() : []
    if (contents.length === 0) return ''

    for (const content of contents) {
      const doc = content.document
      if (!doc) continue

      const headings = doc.querySelectorAll('h1, h2, h3, h4, h5, h6')
      if (headings.length > 0) {
        return headings[0].textContent?.trim() || ''
      }

      const titleEl = doc.querySelector('title')
      if (titleEl) {
        return titleEl.textContent?.trim() || ''
      }
    }

    return ''
  } catch (err) {
    console.error('获取章节标题失败:', err)
    return ''
  }
}

/**
 * 手动保存当前阅读位置为书签
 * @param rendition epub.js rendition 实例
 * @param filePath 书籍文件路径
 * @returns 保存的书签数据，失败返回 null
 */
export async function saveBookmark(
  rendition: any,
  filePath: string
): Promise<Bookmark | null> {
  if (!rendition) return null
  try {
    const location = rendition.currentLocation()
    if (!location || !location.start) return null

    const cfi = location.start.cfi
    const chapterTitle = getCurrentChapterTitle(rendition, location)
    const snippet = getParagraphTextByCfi(rendition, cfi)

    const bookmark: Bookmark = {
      cfi: cfi,
      percentage: location.start.percentage || 0,
      timestamp: Date.now(),
      chapterTitle,
      snippet
    }

    // @ts-ignore
    await window.go.main.App.SaveBookmark(filePath, JSON.stringify(bookmark))
    console.log('书签已保存:', bookmark)
    return bookmark
  } catch (err) {
    console.error('保存书签失败:', err)
    return null
  }
}

/**
 * 获取书籍的所有书签
 * @param filePath 书籍文件路径
 * @returns 书签列表，失败返回 null
 */
export async function getBookmarks(filePath: string): Promise<Bookmark[] | null> {
  try {
    // @ts-ignore
    const result = await window.go.main.App.GetBookmarks(filePath)
    if (result) {
      return JSON.parse(result)
    }
    return null
  } catch (err) {
    console.error('获取书签失败:', err)
    return null
  }
}

/**
 * 删除指定书签
 * @param filePath 书籍文件路径
 * @param cfi 书签的 CFI
 */
export async function deleteBookmark(filePath: string, cfi: string): Promise<void> {
  try {
    // @ts-ignore
    await window.go.main.App.DeleteBookmark(filePath, cfi)
    console.log('书签已删除:', cfi)
  } catch (err) {
    console.error('删除书签失败:', err)
    throw err
  }
}

export interface Note {
  cfi: string;
  percentage: number;
  timestamp: number;
  chapterTitle?: string;
  content: string;
  selectedText?: string;
  color?: string;
}

export interface SearchHistoryItem {
  keyword: string;
  timestamp: number;
}

export async function getSearchHistory(filePath: string): Promise<SearchHistoryItem[] | null> {
  try {
    // @ts-ignore
    const result = await window.go.main.App.GetSearchHistory(filePath)
    if (result) {
      return JSON.parse(result)
    }
    return null
  } catch (err) {
    console.error('获取搜索历史失败:', err)
    return null
  }
}

export async function saveSearchHistory(filePath: string, keyword: string): Promise<void> {
  try {
    // @ts-ignore
    await window.go.main.App.SaveSearchHistory(filePath, keyword)
    console.log('搜索历史已保存:', keyword)
  } catch (err) {
    console.error('保存搜索历史失败:', err)
    throw err
  }
}

export async function clearSearchHistory(filePath: string): Promise<void> {
  try {
    // @ts-ignore
    await window.go.main.App.ClearSearchHistory(filePath)
    console.log('搜索历史已清除')
  } catch (err) {
    console.error('清除搜索历史失败:', err)
    throw err
  }
}

/**
 * 获取选中文本的 CFI 位置
 * @param rendition epub.js rendition 实例
 * @returns CFI 字符串，失败返回空字符串
 */
function getSelectionCfi(rendition: any): string {
  try {
    const contents = rendition.getContents ? rendition.getContents() : []
    for (const content of contents) {
      const selection = content.window.getSelection()
      if (selection && !selection.isCollapsed) {
        const range = selection.getRangeAt(0)
        return content.cfiFromRange(range)
      }
    }
    return ''
  } catch (err) {
    console.error('获取选中文本CFI失败:', err)
    return ''
  }
}

/**
 * 保存笔记
 * @param rendition epub.js rendition 实例
 * @param filePath 书籍文件路径
 * @param selectedText 选中的文本内容
 * @param content 用户笔记内容（可选）
 * @param color 高亮颜色（可选）
 * @returns 保存的笔记数据，失败返回 null
 */
export async function saveNote(
  rendition: any,
  filePath: string,
  selectedText: string,
  content?: string,
  color?: string
): Promise<Note | null> {
  if (!rendition || !selectedText.trim()) return null
  try {
    const cfi = getSelectionCfi(rendition)
    if (!cfi) {
      console.error('无法获取选中文本的CFI位置')
      return null
    }

    const location = rendition.currentLocation()
    const chapterTitle = location && location.start ? getCurrentChapterTitle(rendition, location) : ''
    const percentage = location && location.start ? location.start.percentage || 0 : 0

    const note: Note = {
      cfi: cfi,
      percentage: percentage,
      timestamp: Date.now(),
      chapterTitle,
      content: content || '',
      selectedText: selectedText.trim(),
      color: color || '#FFCDD2'
    }

    // @ts-ignore
    await window.go.main.App.SaveNote(filePath, JSON.stringify(note))
    console.log('笔记已保存:', note)
    return note
  } catch (err) {
    console.error('保存笔记失败:', err)
    return null
  }
}

/**
 * 获取书籍的所有笔记
 * @param filePath 书籍文件路径
 * @returns 笔记列表，失败返回 null
 */
export async function getNotes(filePath: string): Promise<Note[] | null> {
  try {
    // @ts-ignore
    const result = await window.go.main.App.GetNotes(filePath)
    if (result) {
      return JSON.parse(result)
    }
    return null
  } catch (err) {
    console.error('获取笔记失败:', err)
    return null
  }
}

/**
 * 删除指定笔记
 * @param filePath 书籍文件路径
 * @param cfi 笔记的 CFI
 */
export async function deleteNote(filePath: string, cfi: string): Promise<void> {
  try {
    // @ts-ignore
    await window.go.main.App.DeleteNote(filePath, cfi)
    console.log('笔记已删除:', cfi)
  } catch (err) {
    console.error('删除笔记失败:', err)
    throw err
  }
}
