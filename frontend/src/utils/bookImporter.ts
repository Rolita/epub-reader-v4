export interface ImportResult {
  success: boolean
  title: string
  author?: string
  coverUrl?: string
  md5: string
  filePath: string
  error?: string
}

/**
 * 导入书籍的完整流程
 * @param filePath 文件路径
 * @param shelfName 书架名称
 * @returns 导入结果
 */
export async function importBook(filePath: string, shelfName: string): Promise<ImportResult> {
  try {
    // 调用 Go 后端进行处理
    // @ts-ignore
    const result: ImportResult = await window.go.main.App.ProcessAndImportEpub(filePath, shelfName)
    return result
  } catch (error) {
    console.error('导入书籍失败:', error)
    return {
      success: false,
      title: '',
      md5: '',
      filePath: '',
      error: error instanceof Error ? error.message : '未知错误'
    }
  }
}

/**
 * 批量导入书籍
 * @param filePaths 文件路径列表
 * @param shelfName 书架名称
 * @param onProgress 进度回调
 * @returns 导入结果列表
 */
export async function importBooks(
  filePaths: string[],
  shelfName: string,
  onProgress?: (current: number, total: number, title: string) => void
): Promise<ImportResult[]> {
  const results: ImportResult[] = []
  
  for (let i = 0; i < filePaths.length; i++) {
    const filePath = filePaths[i]
    
    if (onProgress) {
      onProgress(i + 1, filePaths.length, filePath.split(/[\\/]/).pop() || '')
    }
    
    const result = await importBook(filePath, shelfName)
    results.push(result)
  }
  
  return results
}

/**
 * 从文件夹导入书籍
 * 先扫描文件夹中的所有 EPUB 文件，然后批量导入
 * @param folderPath 文件夹路径
 * @param shelfName 书架名称
 * @param onProgress 进度回调
 * @returns 导入结果列表
 */
export async function importBooksFromFolder(
  folderPath: string,
  shelfName: string,
  onProgress?: (current: number, total: number, title: string) => void
): Promise<ImportResult[]> {
  // 扫描文件夹中的 EPUB 文件
  // @ts-ignore
  const filePaths: string[] = await window.go.main.App.ScanEpubFilesInFolder(folderPath)
  
  if (filePaths.length === 0) {
    return [{
      success: false,
      title: '',
      md5: '',
      filePath: '',
      error: '未找到 EPUB 文件'
    }]
  }
  
  return await importBooks(filePaths, shelfName, onProgress)
}