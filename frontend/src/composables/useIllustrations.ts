import { ref } from 'vue'
import { useBookStore } from '../stores/book'

// 插图接口
export interface Illustration {
  id: string
  src: string        // 后端图片 URL
  alt: string        // 描述/文件名
  bookId: string     // 所属书籍ID
  bookTitle: string  // 所属书籍标题
  index: number      // 在书籍中的序号
}

// composable 函数
export function useIllustrations() {
  const illustrations = ref<Illustration[]>([])
  const isLoading = ref(false)
  const selectedIllustration = ref<Illustration | null>(null)

  // 从当前阅读书籍收集插图
  const collectFromCurrentBook = async (bookInfo: { bookId: string; bookTitle: string }) => {
    const bookStore = useBookStore()

    const filePath = bookStore.activeBookPath
    if (!filePath) {
      console.warn('[插图收集] 当前没有打开的书籍')
      return
    }

    console.log('[插图收集] 开始收集插图，filePath:', filePath)

    isLoading.value = true
    illustrations.value = []

    try {
      const { bookId, bookTitle } = bookInfo

      console.log('[插图收集] bookId:', bookId, 'bookTitle:', bookTitle)

      // 调用后端接口获取插图信息
      // @ts-ignore
      const result = await window.go.main.App.GetEpubIllustrations(filePath, bookId, bookTitle)
      if (!result) {
        console.warn('[插图收集] 无法获取插图信息')
        return
      }

      const data = JSON.parse(result)
      console.log('[插图收集] 获取到插图信息:', data)

      // 构建插图列表
      illustrations.value = data.illustrations.map((item: any, i: number) => ({
        id: `${bookId}-${i}`,
        src: `${item.src}?epubPath=${encodeURIComponent(filePath)}`,
        alt: item.alt,
        bookId: data.bookId,
        bookTitle: data.bookTitle,
        index: item.index
      }))

      console.log('[插图收集] 完成，收集到', illustrations.value.length, '张图片')
    } catch (err) {
      console.error('[插图收集] 收集失败:', err)
    } finally {
      isLoading.value = false
    }
  }

  // 选择插图
  const selectIllustration = (illustration: Illustration | null) => {
    selectedIllustration.value = illustration
  }

  // 清除所有插图
  const clearIllustrations = () => {
    illustrations.value = []
    selectedIllustration.value = null
  }

  return {
    illustrations,
    isLoading,
    selectedIllustration,
    collectFromCurrentBook,
    selectIllustration,
    clearIllustrations
  }
}
