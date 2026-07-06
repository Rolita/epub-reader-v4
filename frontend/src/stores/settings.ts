import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface IllustrationItem {
  src: string
  alt: string
  index: number
  href: string
  chapterHref: string
  chapterTitle: string
  cfi: string
  bookId?: string
  bookTitle?: string
}

export const useSettingsStore = defineStore('settings', () => {
  const fontSize = ref(16);            // 字号 (px)
  const fontFamily = ref('Microsoft YaHei'); // 字体
  const paragraphGap = ref(10);        // 段间距 (px)
  const lineHeight = ref(1.5);         // 行间距
  const letterSpacing = ref(0);        // 字间距 (px)
  const indent = ref(2);               // 首行缩进 (em)
  const textAlign = ref('justify');     // 两端对齐
  const sidebarWidth = ref(240);       // 侧边栏宽度 (px)
  const bookshelfColumns = ref(6);     // 书架列数 (3-8)
  const coverGap = ref(12);           // 封面间距 (8/12/20)
  const sortBy = ref('default');      // 排序方式: default/title-asc/title-desc/author-asc/author-desc
  const showHeader = ref(true);        // 是否显示页眉
  const showIllustrationSidebar = ref(false); // 是否显示插画侧边栏
  const illustrations = ref<IllustrationItem[]>([]); // 当前书籍的插图列表
  const windowWidth = ref(1920);       // 启动窗口宽度
  const windowHeight = ref(1080);      // 启动窗口高度

  const setIllustrations = (items: IllustrationItem[]) => {
    illustrations.value = items;
  };

  const clearIllustrations = () => {
    illustrations.value = [];
  };

  return {
    fontSize,
    fontFamily,
    paragraphGap,
    lineHeight,
    letterSpacing,
    indent,
    textAlign,
    sidebarWidth,
    bookshelfColumns,
    coverGap,
    sortBy,
    showHeader,
    showIllustrationSidebar,
    illustrations,
    windowWidth,
    windowHeight,
    setIllustrations,
    clearIllustrations
  };
}, {
  persist: true
});