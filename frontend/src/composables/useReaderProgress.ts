export interface ProgressData {
  cfi: string
  percentage: number
  timestamp: number
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
 * 恢复到指定 CFI 位置
 * @param rendition epub.js rendition 实例
 * @param cfi EPUB CFI 位置标识
 * @param onFocus 可选的聚焦回调
 */
export async function restoreReaderProgress(
  rendition: any,
  cfi: string,
  onFocus?: () => void
): Promise<void> {
  if (!rendition || !cfi) return
  try {
    await rendition.display(cfi)
    onFocus?.()
    console.log('已恢复到进度:', cfi)
  } catch (err) {
    console.error('恢复进度失败:', err)
  }
}
