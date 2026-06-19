export interface ReaderTab {
  type: string
  filePath?: string
}

export interface ReaderRef {
  saveProgress?: () => Promise<any>
  restoreProgress?: (cfi: string) => Promise<void>
}

export type NotifyType = 'success' | 'error'

/**
 * 恢复阅读进度（按钮点击处理）
 * @param activeTab 当前活跃的 tab
 * @param readerRef 阅读器组件引用
 * @param onNotify 通知回调
 */
export async function handleRestoreProgress(
  activeTab: ReaderTab | undefined,
  readerRef: ReaderRef | null,
  onNotify: (message: string, type: NotifyType) => void
): Promise<void> {
  if (!activeTab || activeTab.type !== 'reader' || !activeTab.filePath) {
    onNotify('请先打开一本书', 'error')
    return
  }

  try {
    // @ts-ignore
    const progressJSON = await window.go.main.App.GetProgress(activeTab.filePath)
    if (!progressJSON) {
      onNotify('没有找到阅读进度', 'error')
      return
    }

    const progress = JSON.parse(progressJSON)
    if (!progress.cfi) {
      onNotify('进度数据无效', 'error')
      return
    }

    await readerRef?.restoreProgress?.(progress.cfi)
    await readerRef?.restoreProgress?.(progress.cfi)
    onNotify('已恢复阅读进度', 'success')
  } catch (e) {
    onNotify('恢复进度失败: ' + (e as Error).message, 'error')
  }
}

/**
 * 保存阅读进度（按钮点击处理）
 * @param activeTab 当前活跃的 tab
 * @param readerRef 阅读器组件引用
 * @param onNotify 通知回调
 */
export async function handleSaveProgress(
  activeTab: ReaderTab | undefined,
  readerRef: ReaderRef | null,
  onNotify: (message: string, type: NotifyType) => void
): Promise<void> {
  if (!activeTab || activeTab.type !== 'reader' || !activeTab.filePath) {
    onNotify('请先打开一本书', 'error')
    return
  }

  try {
    const result = await readerRef?.saveProgress?.()
    if (result) {
      onNotify('阅读进度已保存', 'success')
    } else {
      onNotify('保存进度失败', 'error')
    }
  } catch (e) {
    onNotify('保存进度失败: ' + (e as Error).message, 'error')
  }
}
