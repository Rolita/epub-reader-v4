<template>
  <div 
    class="reader-view" 
    :class="{ 'fullscreen': isFullscreen }"
    @mousemove="handleMouseMove"
  >
    <!-- 加载动画 -->
    <LoadingOverlay v-if="isLoading" />
    
    <!-- 页眉栏 -->
    <div class="reader-header" v-if="settingsStore.showHeader && currentChapterTitle">
      <span class="chapter-title">{{ currentChapterTitle }}</span>
    </div>
    
    <div ref="viewerContainer" class="viewer-container" :class="{ 'hidden': isLoading }"></div>
    
    <!-- 全屏模式顶部拖拽区域 -->
    <div v-if="isFullscreen" class="fullscreen-drag-area"></div>
    
    <div class="reader-controls" :class="{ 'hidden': isLoading }">
      <button @click="prevPage" class="nav-btn left">〈</button>
      <button @click="nextPage" class="nav-btn right">〉</button>
    </div>
    
    <!-- 阅读进度条 -->
    <ReaderProgress 
      :progress="progress" 
      :is-hidden="isLoading" 
      :visible="showProgress"
      @click="handleProgressClick"
      @mouseenter="handleProgressMouseEnter"
      @mouseleave="handleProgressMouseLeave"
    />



    <ReaderFunctionMenu
      :visible="showFunctionMenu"
      :is-loading="isLoading"
      :is-fullscreen="isFullscreen"
      :file-path="props.filePath"
      :rendition="rendition"
      @enter-fullscreen="enterFullscreen"
      @exit-fullscreen="exitFullscreen"
      @mouseleave="showFunctionMenu = false"
      @bookmark-saved="handleBookmarkSaved"
      @copy-selected="copySelectedText"
      @save-note="saveSelectedTextAsNote"
    />

    <!-- 图片预览组件 -->
    <ImagePreview 
      :visible="imagePreviewVisible" 
      :src="previewImageSrc"
      :alt="previewImageAlt"
      :book-md5="props.bookMd5"
      :shelf-name="props.shelfName"
      @close="closeImagePreview"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import Epub from 'epubjs';
import ImagePreview from './ImagePreview.vue';
import LoadingOverlay from './LoadingOverlay.vue';
import ReaderProgress from './ReaderProgress.vue';
import { useBookStore, TocItem } from '../stores/book';
import { useSettingsStore } from '../stores/settings';
import { useThemeStore } from '../stores/theme';
import { useLibraryStore } from '../stores/library';
import { saveReaderProgress, restoreReaderProgress, saveBookmark, saveNote, getNotes, type Note } from '../composables/useReaderProgress';
import { eventBus } from '../composables/useEventBus';
import FullscreenIcon from './icons/FullscreenIcon.vue';
import ReaderFunctionMenu from './ReaderFunctionMenu.vue';

const props = defineProps<{ filePath: string; isSplitMode?: boolean; isActive?: boolean; tabId?: string; bookMd5?: string; shelfName?: string }>();
const emit = defineEmits<{ 
  (e: 'click'): void;
  (e: 'scroll'): void;
  (e: 'ready'): void;
  (e: 'bookmark-saved'): void;
  (e: 'note-saved'): void;
}>();
const viewerContainer = ref<HTMLElement | null>(null);

// 图片预览相关状态
const imagePreviewVisible = ref(false)
const previewImageSrc = ref('')
const previewImageAlt = ref('')

// 全屏状态
const isFullscreen = ref(false)

// 进度条显示状态
const showProgress = ref(false)

// 鼠标是否在进度条上
const isMouseOnProgress = ref(false)

// 离开进度条后的冷却期
const justLeftProgress = ref(false)
let leaveProgressTimer: any = null

// 当前章节标题
const currentChapterTitle = ref('')
// 当前书籍的目录（组件内部维护，不依赖全局 store）
const bookToc = ref<any[]>([])
// 当前搜索关键词（用于翻页后重新高亮）
const currentSearchKeyword = ref('')

// 图片预览方法
const openImagePreview = (src: string, alt: string = '') => {
  previewImageSrc.value = src
  previewImageAlt.value = alt
  imagePreviewVisible.value = true
}

const closeImagePreview = () => {
  imagePreviewVisible.value = false
}

const libraryStore = useLibraryStore();

const saveProgressAndUpdateLibrary = async () => {
  const result = await saveReaderProgress(rendition, props.filePath)
  if (result && result.percentage !== undefined && props.bookMd5) {
    const progressPercent = Math.round(result.percentage * 100)
    libraryStore.updateBookProgress(props.bookMd5, progressPercent)
  }
  return result
}

let book: any = null;
let rendition: any = null;
let isScrolling = false; // 防抖标志位
let autoSaveTimer: any = null; // 自动保存进度定时器

// 加载状态
const isLoading = ref(true)
// 标记是否正在初始化阅读器
let isInitializing = false

// 阅读进度
const progress = ref(0)
const progressText = ref('0%')

const bookStore = useBookStore();
const settingsStore = useSettingsStore();
const themeStore = useThemeStore();

// 应用排版设置
const applyTypography = () => {
  if (!rendition || !rendition.themes) {
    console.warn('applyTypography: rendition 或 rendition.themes 未就绪');
    return;
  }

  const colors = themeStore.themeColors;

  // 清空 epub.js 内部累积的 injected 样式，避免重复 style 标签
  if (rendition.themes._injected) {
    rendition.themes._injected = [];
  }

  // 使用 default 方法直接作用于基础渲染层
  rendition.themes.default({
    body: {
      'font-family': settingsStore.fontFamily + ' !important',
      'font-size': `${settingsStore.fontSize}px !important`,
      'text-align': settingsStore.textAlign + ' !important',
      'line-height': settingsStore.lineHeight + ' !important',
      'letter-spacing': `${settingsStore.letterSpacing}px !important`,
      'color': colors.text + ' !important',
      'background-color': colors.bg + ' !important'
    },
    p: {
      'text-indent': `${settingsStore.indent}em !important`,
      'margin-bottom': `${settingsStore.paragraphGap}px !important`,
      'font-size': `${settingsStore.fontSize}px !important`,
      'line-height': settingsStore.lineHeight + ' !important',
      'font-family': settingsStore.fontFamily + ' !important',
      'text-align': settingsStore.textAlign + ' !important',
      'letter-spacing': `${settingsStore.letterSpacing}px !important`
    }
  });

  // 注入强力的全局 CSS 覆盖，打破内联样式限制
  injectPowerfulStyles();
};

// 注入强力的全局样式覆盖
const injectPowerfulStyles = () => {
  if (!rendition || !rendition.themes) return;

  const css = `
   /* === 专业排版修正公式 === */
     /* 1. 强制重置所有段落的边距，交给段间距逻辑处理 */
     p {
       margin-top: 0 !important;       /* 严禁使用上边距，防止叠加 */
       margin-bottom: ${settingsStore.paragraphGap}px !important; /* 段间距：只用下边距，统一间距感 */
       /* 2. 行间距：使用纯倍数，保持阅读舒适度 */
       line-height: ${settingsStore.lineHeight} !important;
       /* 3. 中文排版精髓：首行缩进 */
       text-indent: ${settingsStore.indent}em !important;
       /* 4. 强制覆盖段落字体，防止 EPUB 内部样式干扰 */
       font-family: ${settingsStore.fontFamily} !important;
       font-size: ${settingsStore.fontSize}px !important;
       text-align: ${settingsStore.textAlign} !important;
       letter-spacing: ${settingsStore.letterSpacing}px !important;
     }
    /* 4. 强力覆盖所有文本元素的字号和字体 */
    body, div, span, h1, h2, h3, h4, h5, h6,
    li, td, th, blockquote, pre, a, strong, em,
    b, i, u, s, strike, sub, sup, code, var,
    article, section, header, footer, nav {
      font-size: ${settingsStore.fontSize}px !important;
      line-height: ${settingsStore.lineHeight} !important;
      font-family: ${settingsStore.fontFamily} !important;
      text-align: ${settingsStore.textAlign} !important;
      letter-spacing: ${settingsStore.letterSpacing}px !important;
    }
    /* 5. 确保根元素也被覆盖 */
    html {
      font-size: ${settingsStore.fontSize}px !important;
    }
    /* 6. 图片和表格不应该被影响 */
    img, svg, canvas, table, tr, td, th {
      font-size: initial !important;
      line-height: initial !important;
    }
    /* 7. 标题样式优化 */
    h1, h2, h3, h4, h5, h6 {
      margin-top: 1.5em !important;
      margin-bottom: 0.5em !important;
      font-weight: bold !important;
    }
    /* 8. 列表样式优化 */
    ul, ol {
      margin-top: 0 !important;
      margin-bottom: ${settingsStore.paragraphGap}px !important;
      padding-left: 2em !important;
    }
    li {
      margin-top: 0 !important;
      margin-bottom: 0.3em !important;
    }
    /* 9. 搜索高亮样式 */
    .readest-search-highlight {
      background-color: #FFEB3B !important;
      color: #000000 !important;
      border-radius: 2px !important;
      padding: 0 2px !important;
    }
    /* 10. 笔记高亮样式 */
    .readest-note-highlight {
      --note-color: 255, 205, 210;
      background-color: rgba(var(--note-color), 0.8) !important;
      color: #000000 !important;
      border-radius: 2px !important;
      padding: 0 2px !important;
    }
  `;

  // 直接注入到 iframe DOM，避免 epub.js append 累积 style 标签
  const contents = rendition.getContents ? rendition.getContents() : [];
  contents.forEach((content: any) => {
    const doc = content.document;
    if (!doc) return;
    let style = doc.getElementById('readest-typography-style');
    if (!style) {
      style = doc.createElement('style');
      style.id = 'readest-typography-style';
      doc.head.appendChild(style);
    }
    style.textContent = css;
  });
};

// 笔记缓存，用于优化性能
let notesCache: Map<string, any[]> = new Map();
let lastAppliedNotesHash = '';

// 简单的哈希函数，用于检测笔记是否变化
function getNotesHash(notes: any[]): string {
  return notes.map(n => `${n.cfi}${n.content}${n.color}`).join('|');
}

const clearNoteHighlights = (targetCfi?: string) => {
  if (!rendition) return;
  const contents = rendition.getContents ? rendition.getContents() : [];
  contents.forEach((content: any) => {
    const doc = content.document;
    if (!doc) return;
    
    let highlights: NodeListOf<HTMLElement>;
    if (targetCfi) {
      // 只清除特定 CFI 的高亮
      highlights = doc.querySelectorAll(`.readest-note-highlight[data-note-cfi="${targetCfi}"]`);
    } else {
      // 清除所有高亮
      highlights = doc.querySelectorAll('.readest-note-highlight');
    }
    
    highlights.forEach((span: HTMLElement) => {
      const parent = span.parentNode as HTMLElement;
      if (parent) {
        const text = span.textContent || '';
        const textNode = doc.createTextNode(text);
        parent.replaceChild(textNode, span);
      }
    });
    
    // 只在清除所有高亮时才合并相邻文本节点
    if (!targetCfi) {
      const mergeAdjacentTextNodes = (node: Node) => {
        if (node.nodeType === Node.TEXT_NODE) {
          while (node.nextSibling && node.nextSibling.nodeType === Node.TEXT_NODE) {
            node.textContent = (node.textContent || '') + (node.nextSibling.textContent || '');
            node.parentNode?.removeChild(node.nextSibling);
          }
        } else {
          for (let i = 0; i < node.childNodes.length; i++) {
            mergeAdjacentTextNodes(node.childNodes[i]);
          }
        }
      };
      mergeAdjacentTextNodes(doc.body);
    }
  });
};

// 辅助函数：将十六进制颜色转换为 rgb 值（不带透明度）
const hexToRgb = (hex: string) => {
  let r = 0, g = 0, b = 0;
  
  // 3位缩写
  if (hex.length === 4) {
    r = parseInt(hex[1] + hex[1], 16);
    g = parseInt(hex[2] + hex[2], 16);
    b = parseInt(hex[3] + hex[3], 16);
  } 
  // 6位完整
  else if (hex.length === 7) {
    r = parseInt(hex[1] + hex[2], 16);
    g = parseInt(hex[3] + hex[4], 16);
    b = parseInt(hex[5] + hex[6], 16);
  }
  
  return `${r}, ${g}, ${b}`;
};

// 使用 CFI 范围来精确高亮文本
const highlightNoteByCfi = (content: any, note: any) => {
  if (!content.cfiFromRange || !content.rangeFromCfi) {
    console.warn('epub.js 内容不支持 CFI 范围操作，降级到文本匹配');
    return false;
  }
  
  try {
    const doc = content.document;
    if (!doc) return false;
    
    // 尝试通过 CFI 直接获取范围
    let range: Range | null = null;
    try {
      range = content.rangeFromCfi(note.cfi);
    } catch (e) {
      // CFI 可能失效，降级到文本匹配
      console.warn('CFI 范围获取失败，使用文本匹配:', note.cfi);
      return false;
    }
    
    if (!range) return false;
    
    // 使用范围创建高亮
    const highlightRange = (rangeToHighlight: Range) => {
      const span = doc.createElement('mark');
      span.className = 'readest-note-highlight';
      span.setAttribute('data-note-cfi', note.cfi);
      if (note.color) {
        span.style.setProperty('--note-color', hexToRgb(note.color), 'important');
      }
      
      try {
        rangeToHighlight.surroundContents(span);
        return true;
      } catch (e) {
        // 如果不能直接 surround（跨节点），使用更简单的方法
        const fragment = rangeToHighlight.extractContents();
        span.appendChild(fragment);
        rangeToHighlight.insertNode(span);
        return true;
      }
    };
    
    return highlightRange(range);
  } catch (err) {
    console.warn('CFI 高亮失败:', err);
    return false;
  }
};

// 优化的文本匹配高亮（作为降级方案）
const highlightNoteByText = (content: any, note: any) => {
  if (!note.selectedText || !note.selectedText.trim()) return false;
  
  const doc = content.document;
  if (!doc) return false;
  
  const searchText = note.selectedText.trim();
  const escapedText = searchText.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  
  // 使用更高效的 TreeWalker 而不是完整遍历
  const walker = doc.createTreeWalker(
    doc.body,
    NodeFilter.SHOW_TEXT,
    {
      acceptNode: (node: Node) => {
        // 跳过已高亮的节点
        const parent = node.parentNode as HTMLElement;
        if (parent?.classList?.contains('readest-note-highlight') || 
            parent?.classList?.contains('readest-search-highlight')) {
          return NodeFilter.FILTER_REJECT;
        }
        const text = node.textContent || '';
        return text.length >= searchText.length ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_SKIP;
      }
    }
  );
  
  let found = false;
  const node: Node | null = walker.currentNode;
  
  while (walker.nextNode()) {
    const textNode = walker.currentNode as Text;
    const text = textNode.textContent || '';
    const index = text.indexOf(searchText);
    
    if (index !== -1) {
      // 找到匹配，进行高亮
      const before = text.substring(0, index);
      const matchText = text.substring(index, index + searchText.length);
      const after = text.substring(index + searchText.length);
      
      const span = doc.createElement('mark');
      span.className = 'readest-note-highlight';
      span.setAttribute('data-note-cfi', note.cfi);
      span.textContent = matchText;
      if (note.color) {
        span.style.setProperty('--note-color', hexToRgb(note.color), 'important');
      }
      
      const parent = textNode.parentNode as HTMLElement;
      if (before) {
        textNode.textContent = before;
        parent.insertBefore(span, textNode.nextSibling);
      } else {
        parent.replaceChild(span, textNode);
      }
      
      if (after) {
        const afterNode = doc.createTextNode(after);
        parent.insertBefore(afterNode, span.nextSibling);
      }
      
      found = true;
      break; // 只高亮第一个匹配
    }
  }
  
  return found;
};

const applyNoteHighlights = async (incrementalNoteCfi?: string) => {
  if (!rendition || !props.filePath) return;
  
  try {
    const notes = await getNotes(props.filePath);
    const currentNotes = notes || [];
    
    // 检查笔记是否真的变化了（除非是强制刷新）
    const currentHash = getNotesHash(currentNotes);
    const isForcedRefresh = lastAppliedNotesHash === '';
    if (!incrementalNoteCfi && !isForcedRefresh && currentHash === lastAppliedNotesHash) {
      console.log('笔记未变化，跳过高亮更新');
      return;
    }
    
    const contents = rendition.getContents ? rendition.getContents() : [];
    if (contents.length === 0) return;
    
    if (incrementalNoteCfi) {
      // 增量更新：只处理变化的笔记
      const noteToUpdate = currentNotes.find(n => n.cfi === incrementalNoteCfi);
      
      // 先清除旧的高亮
      clearNoteHighlights(incrementalNoteCfi);
      
      if (noteToUpdate) {
        // 应用新高亮
        for (const content of contents) {
          let highlighted = highlightNoteByCfi(content, noteToUpdate);
          if (!highlighted) {
            highlightNoteByText(content, noteToUpdate);
          }
        }
      }
      
      console.log('增量更新笔记高亮:', incrementalNoteCfi);
    } else {
      // 完整更新：清除并重新应用所有
      clearNoteHighlights();
      
      for (const note of currentNotes) {
        let highlighted = false;
        
        for (const content of contents) {
          highlighted = highlightNoteByCfi(content, note);
          if (highlighted) break;
        }
        
        if (!highlighted) {
          for (const content of contents) {
            highlightNoteByText(content, note);
          }
        }
      }
      
      lastAppliedNotesHash = currentHash;
      console.log('完整应用笔记高亮:', currentNotes.length, '个笔记');
    }
  } catch (err) {
    console.error('应用笔记高亮失败:', err);
  }
  
  // 确保 lastAppliedNotesHash 被重置，这样即使笔记没变化，下次强制刷新也会生效
  // 但是如果是增量更新，我们不需要重置
  if (!incrementalNoteCfi) {
    // 这里不重置，保持正常逻辑
  }
};

// 应用主题到 EPUB 内容
const applyTheme = () => {
  if (!rendition || !rendition.themes) {
    console.warn('applyTheme: rendition 或 rendition.themes 未就绪');
    return;
  }

  const colors = themeStore.themeColors;

  // 清空 epub.js 内部累积的 injected 样式，避免重复 style 标签
  if (rendition.themes._injected) {
    rendition.themes._injected = [];
  }

  // 强制设置默认主题色
  rendition.themes.default({
    body: {
      'color': colors.text + ' !important',
      'background-color': colors.bg + ' !important'
    },
    '*': {
      'color': colors.text + ' !important'
    }
  });

  // 直接注入到 iframe DOM，避免 epub.js append 累积 style 标签
  const themeCss = `
    * {
      color: ${colors.text} !important;
      background-color: transparent !important;
      text-shadow: none !important;
    }
    html, body {
      background-color: ${colors.bg} !important;
      color: ${colors.text} !important;
    }
    table, tr, td, div, p, span, h1, h2, h3, h4, h5, h6 {
      background-color: transparent !important;
      color: ${colors.text} !important;
    }
    a {
      color: ${colors.text} !important;
    }
  `;

  const contents = rendition.getContents ? rendition.getContents() : [];
  contents.forEach((content: any) => {
    const doc = content.document;
    if (!doc) return;
    let style = doc.getElementById('readest-theme-style');
    if (!style) {
      style = doc.createElement('style');
      style.id = 'readest-theme-style';
      doc.head.appendChild(style);
    }
    style.textContent = themeCss;
  });
};

// 获取 href 中的文件名部分
const getFileName = (href: string) => {
  if (!href) return '';
  const url = new URL(href, 'http://example.com');
  return url.pathname.split('/').pop() || '';
};

// 更新阅读进度
const updateProgress = () => {
  if (!rendition || !book) return
  
  const location = rendition.location
  if (!location) return
  
  let progressPercent = 0
  
  if (book.locations && book.locations.total > 0 && location.start.cfi) {
    try {
      const percentage = book.locations.percentageFromCfi(location.start.cfi)
      progressPercent = Math.round(percentage * 100)
    } catch (e) {
      console.warn('使用 locations 计算进度失败:', e)
    }
  }
  
  if (progressPercent === 0 && book.spine) {
    const sections: any[] = []
    if (typeof book.spine.each === 'function') {
      book.spine.each((section: any) => sections.push(section))
    } else if (Array.isArray(book.spine)) {
      sections.push(...book.spine)
    } else if (book.spine.items) {
      sections.push(...book.spine.items)
    }
    
    const currentSectionIndex = sections.findIndex(s => 
      location.start.href && s.href && location.start.href.includes(s.href)
    )
    
    if (currentSectionIndex >= 0 && sections.length > 0) {
      progressPercent = Math.round((currentSectionIndex / (sections.length - 1)) * 100)
    }
  }
  
  if (location.start && location.end) {
    const indexProgress = Math.round((location.start.index / location.end.index) * 100)
    if (indexProgress > 0 && indexProgress < 100) {
      progressPercent = indexProgress
    }
  }
  
  progress.value = Math.min(Math.max(progressPercent, 0), 100)
  progressText.value = `${progress.value}%`
}

// 点击进度条跳转
const handleProgressClick = (percent: number) => {
  if (!rendition || !book) return
  
  if (book.locations && typeof book.locations.cfiFromPercentage === 'function') {
    const cfi = book.locations.cfiFromPercentage(percent)
    rendition.display(cfi)
    setTimeout(() => {
      rendition.display(cfi)
    }, 50)
  } else if (book.spine) {
    const sections: any[] = []
    if (typeof book.spine.each === 'function') {
      book.spine.each((section: any) => sections.push(section))
    } else if (Array.isArray(book.spine)) {
      sections.push(...book.spine)
    } else if (book.spine.items) {
      sections.push(...book.spine.items)
    }
    
    const targetIndex = Math.min(Math.floor(percent * sections.length), sections.length - 1)
    const targetSection = sections[targetIndex]
    if (targetSection) {
      rendition.display(targetSection.href)
      setTimeout(() => {
        rendition.display(targetSection.href)
      }, 50)
    }
  }
}

// 更新当前章节标题
const updateChapterTitle = (section?: any) => {
  const toc = bookToc.value;
  if (!toc || toc.length === 0) return;
  
  let currentHref = '';
  
  if (section && section.href) {
    currentHref = section.href;
  } else if (rendition?.location?.start?.href) {
    currentHref = rendition.location.start.href;
  } else if (section && section.cfi) {
    const startPos = section.cfi.indexOf('/') + 1;
    const endPos = section.cfi.indexOf('/', startPos);
    if (endPos > startPos) {
      currentHref = section.cfi.substring(startPos, endPos);
    }
  } else {
    return;
  }
  
  const currentFileName = getFileName(currentHref);
  
  let currentChapter = toc.find(item => {
    const itemFileName = getFileName(item.href);
    return currentFileName === itemFileName;
  });
  
  if (!currentChapter) {
    currentChapter = toc.find(item => {
      return currentHref.includes(item.href);
    });
  }
  
  currentChapterTitle.value = currentChapter?.label || '';
};

// 清理内联样式并重新应用排版
const clearInlineStyles = () => {
  if (!rendition) return;

  rendition.on('rendered', (section: any) => {
    const doc = section.document;
    if (!doc) return;

    updateChapterTitle(section);

    const elements = doc.querySelectorAll('*');
    elements.forEach((el: any) => {
      if (el.classList?.contains?.('readest-search-highlight')) return;
      el.style.removeProperty('color');
      el.style.removeProperty('text-shadow');
      el.style.removeProperty('font-size');
      el.style.removeProperty('line-height');
      el.style.removeProperty('font-family');
      el.style.removeProperty('font-weight');
      el.style.removeProperty('text-align');
    });

    requestAnimationFrame(() => {
      if (rendition && rendition.themes) {
        injectPowerfulStyles();
        applyTheme();
      }
      
      if (currentSearchKeyword.value) {
        highlightSearchKeyword(currentSearchKeyword.value);
      }
      
      // 重新应用笔记高亮 - 清除旧的hash强制刷新
      lastAppliedNotesHash = '';
      // 稍微延迟一下确保样式应用完成
      setTimeout(() => {
        applyNoteHighlights();
      }, 100);
    });
  });
};

// 展平目录结构
const flattenToc = (items: any[], level: number = 1, parentId?: string): TocItem[] => {
  const result: TocItem[] = [];
  for (const item of items) {
    const id = `toc-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    result.push({
      id,
      label: item.label,
      href: item.href,
      level,
      parentId,
      hasChildren: item.subitems && item.subitems.length > 0
    });
    if (item.subitems && item.subitems.length > 0) {
      result.push(...flattenToc(item.subitems, level + 1, id));
    }
  }
  return result;
};

const base64ToBuffer = (base64: string) => {
  const bin = window.atob(base64);
  const buffer = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) buffer[i] = bin.charCodeAt(i);
  return buffer.buffer;
};

const prevPage = () => {
  if (rendition) {
    rendition.prev()
    requestAnimationFrame(() => {
      updateProgress()
    })
  }
};
const nextPage = () => {
  if (rendition) {
    rendition.next()
    requestAnimationFrame(() => {
      updateProgress()
    })
  }
};

const initReader = async () => {
  if (isInitializing) {
    return;
  }
  try {
    isInitializing = true;
    isLoading.value = true
		
		if (rendition) {
			rendition.destroy();
			rendition = null;
		}
		book = null;

		// 1. 构建本地文件服务 URL
		const encodedPath = encodeURIComponent(props.filePath);
		const epubUrl = `/local-file/${encodedPath}`;

		// 2. 初始化 epub.js - 直接使用 URL，避免 Base64 编码开销
		book = Epub(epubUrl);

    // 3. 加载书籍导航信息（目录）
    await book.ready;
    const navigation = book.navigation;
    const toc = navigation ? flattenToc(navigation.toc) : [];
    
    // 4. 更新 Pinia store
    bookStore.setActiveBook(props.filePath, toc);
    // 5. 保存到组件内部，供分屏模式使用
    bookToc.value = toc;
    
    // 6. 生成 locations 索引，支持进度条跳转
    if (book.locations) {
      await book.locations.generate()
    }

    // 6. 注册 tabId 和 EPUB 文件路径的映射
    if (props.tabId) {
      // @ts-ignore
      await window.go.main.App.RegisterEpubTab(props.tabId, props.filePath)
    }

    if (viewerContainer.value) {
      rendition = book.renderTo(viewerContainer.value, {
        width: '100%',
        height: '100%',
        flow: 'paginated',
        spread: props.isSplitMode ? 'none' : 'always',
      });

      // 5. 注入滚轮监听和点击焦点钩子
      rendition.hooks.content.register((contents: any) => {
        // 监听 iframe 内部的鼠标事件，用于隐藏功能菜单
        const iframeDoc = contents.window.document.documentElement || contents.window.document.body;
        if (iframeDoc) {
          iframeDoc.addEventListener('mouseenter', () => {
            showFunctionMenu.value = false;
          });
          // Note: mouseleave from iframe will naturally be caught by parent's mousemove if it re-enters the parent region
          // Or, if we want to be explicit, we could add a mouseleave here that does nothing, or relies on parent logic.
          // For now, just handling mouseenter to hide. The parent's mousemove on .reader-view will handle showing.
        }

        contents.window.addEventListener('wheel', (event: WheelEvent) => {
          if (isScrolling) return;

          event.preventDefault(); // 阻止默认的页面滚动
          isScrolling = true;

          if (event.deltaY > 0) {
            nextPage();
          } else {
            prevPage();
          }

          // 400ms 防抖，防止连翻
          setTimeout(() => { isScrolling = false; }, 50);

          // 触发 scroll 事件，用于更新焦点
          emit('scroll')
        }, { passive: false });

        const doc = contents.document

        // 绑定图片点击预览事件
        const bindImageEvents = () => {
          const images = doc.querySelectorAll('img')
          images.forEach((img: HTMLImageElement) => {
            if (img.dataset.previewBound) return
            img.dataset.previewBound = 'true'
            img.style.cursor = 'zoom-in'
            img.addEventListener('click', (e) => {
              e.preventDefault()
              e.stopPropagation()
              let src = img.src
              if (!src || src === 'about:blank') {
                const currentSrc = (img as any).currentSrc || img.getAttribute('src')
                if (currentSrc) src = currentSrc
              }
              openImagePreview(src, img.alt || '')
            })
          })

          const svgImages = doc.querySelectorAll('svg image')
          svgImages.forEach((svgImg: any) => {
            if (svgImg.dataset.previewBound) return
            svgImg.dataset.previewBound = 'true'
            svgImg.style.cursor = 'zoom-in'
            let src = svgImg.getAttribute('xlink:href') || svgImg.getAttribute('href')
            if (src) {
              if (!src.startsWith('data:') && !src.startsWith('http') && !src.startsWith('blob:')) {
                const baseUrl = doc.baseURI || window.location.href
                const imgPath = src.startsWith('/') ? src : new URL(src, baseUrl).href
                src = imgPath
              }
              svgImg.addEventListener('click', (e: Event) => {
                e.preventDefault()
                e.stopPropagation()
                openImagePreview(src, svgImg.getAttribute('alt') || '')
              })
            }
          })
        }

        bindImageEvents()

        const observer = new MutationObserver(() => {
          bindImageEvents()
        })
        observer.observe(doc.body || doc.documentElement, {
          childList: true,
          subtree: true
        })

        // 点击事件只在非图片元素上触发
        doc.addEventListener('click', (e: Event) => {
          const target = e.target as HTMLElement
          if (target.tagName === 'IMG' || target.closest('img')) {
            return
          }
          emit('click')
        });

        // 监听文本选择变化，通过 postMessage 传递给主窗口
        contents.window.addEventListener('selectionchange', () => {
          const selection = contents.window.getSelection()
          if (!selection || selection.isCollapsed || !selection.toString().trim()) {
            window.parent.postMessage({
              type: 'epub-selection-change',
              hasSelection: false
            }, '*')
            return
          }

          const range = selection.getRangeAt(0)
          const rect = range.getBoundingClientRect()
          const selectedText = selection.toString()

          window.parent.postMessage({
            type: 'epub-selection-change',
            hasSelection: true,
            text: selectedText,
            rect: {
              left: rect.left,
              top: rect.top,
              width: rect.width,
              height: rect.height,
              bottom: rect.bottom
            }
          }, '*')
        });

      });

      // 7. 注册内联样式清理（在每章渲染后执行）
      clearInlineStyles();

      // 9. 使用渲染钩子确保主题在内容渲染后应用
      rendition.hooks.render.register(() => {
        console.log('Rendition 渲染钩子触发');
        // 确保 themes 对象已就绪后再应用
        if (rendition && rendition.themes) {
          applyTypography();
          applyTheme();
        }
      });

      // 9. 先获取阅读进度，直接加载到上次阅读位置
      let startCfi: string | undefined
      try {
        // @ts-ignore
        const progressJSON = await window.go.main.App.GetProgress(props.filePath)
        if (progressJSON) {
          const progress = JSON.parse(progressJSON)
          if (progress.cfi) {
            startCfi = progress.cfi
            console.log('将从进度位置开始:', progress.cfi)
          }
        }
      } catch (e) {
        console.error('获取阅读进度失败:', e)
      }

      // 9.1 初始化渲染（从上次阅读位置或书籍开头开始）
      await rendition.display(startCfi || undefined);
      if (startCfi) {
        viewerContainer.value?.focus()
      }

      // 10. 在 display 完成后手动再应用一次（兜底）
      if (rendition && rendition.themes) {
        console.log('display 后手动应用主题');
        applyTypography();
        applyTheme();
      }
      
      // 11. 应用笔记高亮
      setTimeout(() => {
        applyNoteHighlights();
      }, 300);

      console.log("阅读器初始化完成");
      
      // 更新初始进度
      setTimeout(() => {
        updateProgress()
      }, 100)
      
      // 监听翻页事件更新进度
      rendition.on('rendered', () => {
        setTimeout(() => {
          updateProgress()
        }, 50)
      })
      
      // 监听 locationChanged 事件
      if (rendition.on) {
        rendition.on('locationChanged', () => {
          updateProgress()
        })
      }
      
      // 通知父组件书籍已加载完成
      emit('ready')
      
      // 加载完成，隐藏加载动画
      isLoading.value = false
      
      // 如果当前 tab 是激活状态，确保内容正确显示
      if (props.isActive && rendition) {
        setTimeout(async () => {
          rendition.resize('100%', '100%');
          applyTypography();
          applyTheme();
          // 延迟 500ms 后触发恢复阅读进度
          setTimeout(async () => {
            await restoreReaderProgress(rendition, null, () => viewerContainer.value?.focus(), props.filePath);
          }, 500);
        }, 50);
      }
    }
  } catch (err) {
    console.error("阅读器启动失败:", err);
    // 即使失败也要隐藏加载动画
    isLoading.value = false
  } finally {
    isInitializing = false;
  }
};

const handleKey = (e: KeyboardEvent) => {
  if (props.isSplitMode && !props.isActive) {
    return
  }
  
  if (e.key === 'ArrowLeft') prevPage();
  if (e.key === 'ArrowRight') nextPage();

  // 空格键翻页
  if (e.key === ' ') {
    e.preventDefault(); // 阻止默认的空格键滚动行为
    if (e.ctrlKey) {
      prevPage(); // Ctrl + Space 向上翻页
    } else {
      nextPage(); // Space 向下翻页
    }
  }
  
  // F11 键切换全屏
  if (e.key === 'F11') {
    e.preventDefault();
    if (isFullscreen.value) {
      exitFullscreen();
    } else {
      enterFullscreen();
    }
  }
  
  // ESC 键退出全屏
  if (e.key === 'Escape' && isFullscreen.value) {
    exitFullscreen();
  }
};

// 进入全屏
const enterFullscreen = () => {
  const container = document.documentElement;
  const requestFullscreen = () => {
    if (container.requestFullscreen) {
      return container.requestFullscreen();
    } else if ((container as any).webkitRequestFullscreen) {
      return (container as any).webkitRequestFullscreen();
    } else if ((container as any).mozRequestFullScreen) {
      return (container as any).mozRequestFullScreen();
    } else if ((container as any).msRequestFullscreen) {
      return (container as any).msRequestFullscreen();
    }
    return Promise.resolve();
  };

  requestFullscreen().then(() => {
    isFullscreen.value = true;
    
    setTimeout(() => {
      if (rendition) {
        rendition.resize('100%', '100%');
        applyTypography();
        applyTheme();
      }
    }, 100);
  }).catch((err: Error) => {
    console.error('进入全屏失败:', err);
  });
};

// 右侧功能菜单显示状态
const showFunctionMenu = ref(false)

// 鼠标移动处理
const handleMouseMove = (e: MouseEvent) => {
  if (isMouseOnProgress.value || justLeftProgress.value) {
    showFunctionMenu.value = false // 如果鼠标在进度条上，隐藏功能菜单
    return
  }

  const target = e.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()

  // 判断是否在底部区域显示进度条
  const bottomThreshold = 100
  showProgress.value = rect.height - e.clientY < bottomThreshold

  // 判断是否在右侧区域显示功能菜单
  const rightThreshold = 150; // 右侧 150px 区域
  showFunctionMenu.value = rect.width - e.clientX < rightThreshold;
};

const handleProgressMouseEnter = () => {
  isMouseOnProgress.value = true
  justLeftProgress.value = false
  if (leaveProgressTimer) {
    clearTimeout(leaveProgressTimer)
    leaveProgressTimer = null
  }
};

const handleProgressMouseLeave = () => {
  isMouseOnProgress.value = false
  showProgress.value = false
  justLeftProgress.value = true
  if (leaveProgressTimer) {
    clearTimeout(leaveProgressTimer)
  }
  leaveProgressTimer = setTimeout(() => {
    justLeftProgress.value = false
    leaveProgressTimer = null
  }, 100)
};

// 窗口大小变化处理
let resizeTimer: any = null;
const handleResize = () => {
  clearTimeout(resizeTimer);
  resizeTimer = setTimeout(() => {
    if (rendition) {
      rendition.resize('100%', '100%');
      applyTypography();
      applyTheme();
      console.log('窗口大小变化，重新渲染完成');
    }
  }, 200);
};

const handleBookmarkSaved = () => {
  console.log('书签保存成功');
  emit('bookmark-saved');
  eventBus.emit('bookmark-saved');
};

// 获取 iframe 中选中的文本
const getSelectedTextFromIframe = (): string => {
  if (!rendition || !rendition.getContents) return ''
  
  const contents = rendition.getContents()
  for (const content of contents) {
    const selection = content.window.getSelection()
    if (selection && !selection.isCollapsed) {
      return selection.toString()
    }
  }
  
  return ''
}

// 复制选中的文本
const copySelectedText = async () => {
  const text = getSelectedTextFromIframe()
  if (!text) {
    console.log('未选中任何文本')
    return
  }
  
  try {
    await navigator.clipboard.writeText(text)
    console.log('复制成功:', text)
  } catch (err) {
    console.error('复制失败:', err)
  }
};

// 保存当前操作的笔记 CFI，用于增量更新
let lastSavedNoteCfi: string | null = null;

const handleNoteSaved = async (noteCfi?: string) => {
  console.log('笔记保存成功:', noteCfi || '全部刷新');
  emit('note-saved');
  
  // 只有当我们自己不是发送者时才再次发送事件（避免循环）
  // 使用增量更新
  setTimeout(() => {
    applyNoteHighlights(noteCfi || (lastSavedNoteCfi ?? undefined));
    lastSavedNoteCfi = null;
  }, 50);
};

// 保存选中的文本为笔记
const saveSelectedTextAsNote = async () => {
  const text = getSelectedTextFromIframe()
  if (!text) {
    console.log('未选中任何文本')
    return
  }
  
  console.log('准备保存笔记，选中文本:', text)
  
  try {
    const result = await saveNote(rendition, props.filePath, text)
    if (result) {
      lastSavedNoteCfi = result.cfi;
      await handleNoteSaved(result.cfi)
    } else {
      console.error('笔记保存失败：saveNote 返回 null')
    }
  } catch (err) {
    console.error('保存笔记失败:', err)
  }
};

// 退出全屏
const exitFullscreen = () => {
  const exitFullscreenFn = () => {
    if (document.exitFullscreen) {
      return document.exitFullscreen();
    } else if ((document as any).webkitExitFullscreen) {
      return (document as any).webkitExitFullscreen();
    } else if ((document as any).mozCancelFullScreen) {
      return (document as any).mozCancelFullScreen();
    } else if ((document as any).msExitFullscreen) {
      return (document as any).msExitFullscreen();
    }
    return Promise.resolve();
  };

  exitFullscreenFn().then(() => {
    isFullscreen.value = false;
    
    // 延迟重新渲染，确保退出全屏完成
    setTimeout(() => {
      if (rendition) {
        rendition.resize('100%', '100%');
        applyTypography();
        applyTheme();
        console.log('退出全屏，重新渲染完成');
      }
    }, 300);
  }).catch((err: Error) => {
    console.error('退出全屏失败:', err);
    isFullscreen.value = false;
  });
};

// 监听全屏状态变化
const handleFullscreenChange = () => {
  const wasFullscreen = isFullscreen.value;
  isFullscreen.value = !!(document.fullscreenElement || 
    (document as any).webkitFullscreenElement || 
    (document as any).mozFullScreenElement || 
    (document as any).msFullscreenElement);
  
  // 如果状态发生变化，重新渲染
  if (wasFullscreen !== isFullscreen.value) {
    setTimeout(() => {
      if (rendition) {
        rendition.resize('100%', '100%');
        applyTypography();
        applyTheme();
        console.log('全屏状态变化，重新渲染完成');
      }
    }, 300);
  }
};

const clearHighlight = () => {
  currentSearchKeyword.value = '';
  const contents = rendition.getContents ? rendition.getContents() : [];
  contents.forEach((content: any) => {
    const doc = content.document;
    if (!doc) return;
    const highlights = doc.querySelectorAll('.readest-search-highlight');
    highlights.forEach((span: HTMLElement) => {
      const parent = span.parentNode as HTMLElement;
      if (parent) {
        const text = span.textContent || '';
        const textNode = doc.createTextNode(text);
        parent.replaceChild(textNode, span);
      }
    });
    mergeAdjacentTextNodes(doc.body);
  });
};

const mergeAdjacentTextNodes = (element: HTMLElement) => {
  let child = element.firstChild;
  while (child) {
    const next = child.nextSibling;
    if (child.nodeType === Node.TEXT_NODE && next && next.nodeType === Node.TEXT_NODE) {
      child.textContent = (child.textContent || '') + (next.textContent || '');
      element.removeChild(next);
    } else if (child.nodeType === Node.ELEMENT_NODE) {
      mergeAdjacentTextNodes(child as HTMLElement);
    }
    child = next;
  }
};

const highlightSearchKeyword = (keyword: string) => {
  if (!keyword.trim()) return;
  
  currentSearchKeyword.value = keyword;
  
  const contents = rendition.getContents ? rendition.getContents() : [];
  const escapedKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  
  contents.forEach((content: any) => {
    const doc = content.document;
    if (!doc) return;
    
    const highlights = doc.querySelectorAll('.readest-search-highlight');
    highlights.forEach((span: HTMLElement) => {
      const parent = span.parentNode as HTMLElement;
      if (parent) {
        const text = span.textContent || '';
        const textNode = doc.createTextNode(text);
        parent.replaceChild(textNode, span);
      }
    });
    
    mergeAdjacentTextNodes(doc.body);
    
    const getAdjacentTextNodes = (startNode: Node): { nodes: Node[], text: string, nodeIndices: number[] } => {
      const nodes: Node[] = [];
      const nodeIndices: number[] = [];
      let text = '';
      let currentNode: Node | null = startNode;
      
      while (currentNode) {
        if (currentNode.nodeType === Node.TEXT_NODE) {
          const nodeText = currentNode.textContent || '';
          nodes.push(currentNode);
          for (let i = 0; i < nodeText.length; i++) {
            nodeIndices.push(nodes.length - 1);
          }
          text += nodeText;
        } else if (currentNode.nodeType === Node.ELEMENT_NODE) {
          const element = currentNode as HTMLElement;
          if (element.tagName === 'BR' || element.tagName === 'HR') {
            break;
          }
          if (element.classList?.contains?.('readest-search-highlight')) {
            break;
          }
          const elementText = currentNode.textContent || '';
          nodes.push(currentNode);
          for (let i = 0; i < elementText.length; i++) {
            nodeIndices.push(nodes.length - 1);
          }
          text += elementText;
        }
        
        currentNode = currentNode.nextSibling;
      }
      
      return { nodes, text, nodeIndices };
    };
    
    const highlightAcrossNodes = (startNode: Node, keyword: string) => {
      const { nodes, text, nodeIndices } = getAdjacentTextNodes(startNode);
      const regex = new RegExp(escapedKeyword, 'gi');
      let match;
      
      while ((match = regex.exec(text)) !== null) {
        const matchStart = match.index;
        const matchEnd = matchStart + match[0].length;
        
        const startNodeIndex = nodeIndices[matchStart];
        const endNodeIndex = nodeIndices[matchEnd - 1];
        
        if (startNodeIndex === endNodeIndex) {
          const targetNode = nodes[startNodeIndex];
          if (targetNode.nodeType === Node.TEXT_NODE) {
            const nodeText = targetNode.textContent || '';
            const relativeStart = nodeIndices.indexOf(startNodeIndex, matchStart);
            const relativeEnd = relativeStart + match[0].length;
            
            const before = nodeText.substring(0, relativeStart);
            const matchText = nodeText.substring(relativeStart, relativeEnd);
            const after = nodeText.substring(relativeEnd);
            
            const parent = targetNode.parentNode as HTMLElement;
            if (!parent || parent.classList?.contains?.('readest-search-highlight')) continue;
            
            const fragment = doc.createDocumentFragment();
            if (before) fragment.appendChild(doc.createTextNode(before));
            
            const span = doc.createElement('span');
            span.className = 'readest-search-highlight';
            span.textContent = matchText;
            fragment.appendChild(span);
            
            if (after) fragment.appendChild(doc.createTextNode(after));
            
            parent.replaceChild(fragment, targetNode);
          }
        } else {
          for (let i = startNodeIndex; i <= endNodeIndex; i++) {
            const targetNode = nodes[i];
            
            if (targetNode.nodeType === Node.TEXT_NODE) {
              const nodeText = targetNode.textContent || '';
              let startOffset = 0;
              let endOffset = nodeText.length;
              
              if (i === startNodeIndex) {
                startOffset = nodeIndices.indexOf(startNodeIndex, matchStart);
              }
              if (i === endNodeIndex) {
                const endPos = matchEnd - 1;
                const relativeEnd = nodeIndices.lastIndexOf(endNodeIndex, endPos) + 1;
                endOffset = relativeEnd;
              }
              
              const before = nodeText.substring(0, startOffset);
              const matchPart = nodeText.substring(startOffset, endOffset);
              const after = nodeText.substring(endOffset);
              
              const parent = targetNode.parentNode as HTMLElement;
              if (!parent || parent.classList?.contains?.('readest-search-highlight')) continue;
              
              const fragment = doc.createDocumentFragment();
              if (before) fragment.appendChild(doc.createTextNode(before));
              
              const span = doc.createElement('span');
              span.className = 'readest-search-highlight';
              span.textContent = matchPart;
              fragment.appendChild(span);
              
              if (after) fragment.appendChild(doc.createTextNode(after));
              
              parent.replaceChild(fragment, targetNode);
            } else if (targetNode.nodeType === Node.ELEMENT_NODE) {
              const element = targetNode as HTMLElement;
              if (element.classList?.contains?.('readest-search-highlight')) continue;
              
              const span = doc.createElement('span');
              span.className = 'readest-search-highlight';
              span.textContent = element.textContent || '';
              element.parentNode?.replaceChild(span, element);
            }
          }
        }
      }
    };
    
    const processElement = (element: HTMLElement) => {
      if (element.classList?.contains?.('readest-search-highlight')) return;
      
      let child = element.firstChild;
      while (child) {
        const next = child.nextSibling;
        
        if (child.nodeType === Node.TEXT_NODE) {
          const text = child.textContent || '';
          if (text.length > 0) {
            highlightAcrossNodes(child, keyword);
          }
        } else if (child.nodeType === Node.ELEMENT_NODE) {
          processElement(child as HTMLElement);
        }
        
        child = next;
      }
    };
    
    processElement(doc.body);
  });
};

// 跳转到指定章节
const jumpTo = (href: string, cfi?: string) => {
  if (rendition) {
    const target = cfi || href
    console.log('[跳转] 跳转到:', target)
    
    clearNoteHighlights();
    // 强制刷新笔记高亮
    lastAppliedNotesHash = '';
    
    rendition.display(target).then(() => {
      viewerContainer.value?.focus();
      // 立即应用高亮，不延迟
      applyNoteHighlights();
    });
  }
};

// 更新书籍 store（用于 tab 切换时同步目录侧边栏）
const updateBookStore = () => {
  if (book) {
    const navigation = book.navigation;
    const toc = navigation ? flattenToc(navigation.toc) : [];
    bookStore.setActiveBook(props.filePath, toc);
  }
};

// 从 epubjs book 对象收集所有插图
const collectIllustrationsFromBook = async (): Promise<{ src: string; alt: string; index: number; href: string; chapterHref: string; chapterTitle: string; cfi: string }[]> => {
  const illustrations: { src: string; alt: string; index: number; href: string; chapterHref: string; chapterTitle: string; cfi: string }[] = []
  if (!book) return illustrations
  
  console.log('[插图收集] 开始从 epubjs book 收集插图')
  
  // 获取目录信息用于显示章节名称
  const navigation = book.navigation
  const toc = navigation ? flattenToc(navigation.toc) : []
  
  try {
    const images: { src: string; alt: string; href: string }[] = []
    
    if (book.resources && book.resources.resources) {
      console.log('[插图收集] resources.resources 长度:', book.resources.resources.length)
      
      for (const resource of book.resources.resources) {
        if (!resource) continue
        
        const href = resource.href || resource.id || ''
        if (!href) continue
        
        if (isImageFile(href)) {
          console.log('[插图收集] 资源结构:', JSON.stringify(resource))
          
          const resPath = href
          const imageUrl = props.tabId ? `/epub-img/${props.tabId}/${resPath}` : resPath
          
          console.log('[插图收集] 找到图片:', href, '→', imageUrl)
          images.push({ src: imageUrl, alt: href.split('/').pop() || '', href: resPath })
        }
      }
    }
    
    if (images.length === 0 && book.resources && book.resources.assets) {
      console.log('[插图收集] resources.assets 长度:', book.resources.assets.length)
      
      for (const asset of book.resources.assets) {
        if (!asset) continue
        
        const href = asset.href || asset.id || ''
        if (!href) continue
        
        if (isImageFile(href)) {
          const resPath = href
          const imageUrl = props.tabId ? `/epub-img/${props.tabId}/${resPath}` : resPath
          
          console.log('[插图收集] 从 assets 找到图片:', href, '→', imageUrl)
          images.push({ src: imageUrl, alt: href.split('/').pop() || '', href: resPath })
        }
      }
    }
    
    // 为每张图片找到所属章节
    for (let i = 0; i < images.length; i++) {
      const result = await findChapterContainingImage(images[i].href)
      console.log('[插图收集] 图片', images[i].href, '所属章节:', result.href, 'CFI:', result.cfi)
      
      // 根据 chapterHref 查找章节名称
      const tocItem = toc.find(item => item.href === result.href || item.href.startsWith(result.href))
      const chapterTitle = tocItem?.label || ''
      
      illustrations.push({
        src: images[i].src,
        alt: images[i].alt,
        index: i,
        href: images[i].href,
        chapterHref: result.href,
        chapterTitle: chapterTitle,
        cfi: result.cfi
      })
    }
    
    console.log('[插图收集] 完成，共收集', illustrations.length, '张图片')
  } catch (err) {
    console.error('[插图收集] 收集失败:', err)
  }
  
  return illustrations
}

// 查找包含指定图片的章节，并返回 CFI 用于精确定位
const findChapterContainingImage = async (imageHref: string): Promise<{ href: string; cfi: string }> => {
  if (!book || !book.spine) return { href: '', cfi: '' }
  
  const sections: any[] = []
  if (typeof book.spine.each === 'function') {
    book.spine.each((section: any) => sections.push(section))
  } else if (Array.isArray(book.spine)) {
    sections.push(...book.spine)
  } else if (book.spine.items) {
    sections.push(...book.spine.items)
  } else if (book.spine.spineItems) {
    sections.push(...book.spine.spineItems)
  }
  
  // 先尝试用字符串搜索方式找到章节
  // 遍历章节但不加载完整内容，只检查 href 是否可能包含该图片
  for (const section of sections) {
    const sectionHref = section.href || ''
    // 如果章节 href 本身包含图片路径相关的内容，优先使用
    if (sectionHref.includes(imageHref.split('/').pop() || '')) {
      return { href: sectionHref, cfi: '' }
    }
  }
  
  // 尝试加载章节并查找图片元素
  for (const section of sections) {
    try {
      // 使用 book.load 来加载内容，这样路径会正确解析
      const doc = await (book as any).load(section.href)
      if (doc && typeof doc === 'object') {
        const docEl = doc.documentElement ? doc : doc.ownerDocument || doc
        const imgElements = docEl.getElementsByTagName ? docEl.getElementsByTagName('img') : []
        
        for (let j = 0; j < imgElements.length; j++) {
          const src = imgElements[j].getAttribute('src') || ''
          const srcSet = imgElements[j].getAttribute('srcset') || ''
          if (src.includes(imageHref) || srcSet.includes(imageHref)) {
            // 尝试生成 CFI
            let cfi = ''
            if (typeof section.cfiFromElement === 'function') {
              try {
                cfi = section.cfiFromElement(imgElements[j])
              } catch (e) {
                console.warn('[插图收集] 生成 CFI 失败')
              }
            }
            return { href: section.href || '', cfi }
          }
        }
      }
    } catch (err) {
      console.warn('[插图收集] 加载章节失败:', section.href)
    }
  }
  
  // 如果所有章节都加载失败，返回第一个章节
  return { href: sections[0]?.href || '', cfi: '' }
}

function isImageFile(filename: string): boolean {
  const ext = filename.toLowerCase().split('.').pop() || ''
  return ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp'].includes(ext)
}

function isImageResource(resource: any): boolean {
  if (!resource) return false
  const href = resource.href || resource.id || ''
  return isImageFile(href)
}

const searchInBook = async (keyword: string): Promise<Array<{ chapter: string; snippet: string; href: string; cfi: string; page: number }>> => {
  if (!rendition || !book) return [];
  
  const results: Array<{ chapter: string; snippet: string; href: string; cfi: string; page: number }> = [];
  
  const sections: any[] = []
  if (typeof book.spine.each === 'function') {
    book.spine.each((section: any) => sections.push(section))
  } else if (Array.isArray(book.spine)) {
    sections.push(...book.spine)
  } else if (book.spine.items) {
    sections.push(...book.spine.items)
  } else if (book.spine.spineItems) {
    sections.push(...book.spine.spineItems)
  }
  
  const navigation = book.navigation
  const toc = navigation ? flattenToc(navigation.toc) : []
  
  for (const section of sections) {
    try {
      const doc = await (book as any).load(section.href)
      if (!doc || typeof doc !== 'object') continue
      
      const docEl = doc.documentElement ? doc : doc.ownerDocument || doc
      const chapterHref = section.href || ''
      
      const tocItem = toc.find(item => item.href === chapterHref || item.href.startsWith(chapterHref))
      const chapterTitle = tocItem?.label || ''
      
      const paragraphs = docEl.querySelectorAll ? docEl.querySelectorAll('p, h1, h2, h3, h4, h5, h6') : []
      
      for (let i = 0; i < paragraphs.length; i++) {
        const para = paragraphs[i]
        const text = para.textContent || ''
        if (text.toLowerCase().includes(keyword.toLowerCase())) {
          const index = text.toLowerCase().indexOf(keyword.toLowerCase())
          const start = Math.max(0, index - 30)
          const end = Math.min(text.length, index + keyword.length + 30)
          const snippet = (start > 0 ? '...' : '') + text.substring(start, end) + (end < text.length ? '...' : '')
          
          let cfi = ''
          if (typeof section.cfiFromElement === 'function') {
            try {
              cfi = section.cfiFromElement(para)
            } catch (e) {
              console.warn('获取 CFI 失败:', e)
            }
          }
          
          results.push({
            chapter: chapterTitle,
            snippet: snippet,
            href: chapterHref,
            cfi: cfi,
            page: 0
          })
        }
      }
    } catch (err) {
      console.warn('加载章节失败:', section.href)
    }
  }
  
  return results;
};

// 暴露方法给父组件调用
defineExpose({
  jumpTo,
  openImagePreview,
  restoreProgress: (cfi: string) => restoreReaderProgress(rendition, cfi, () => viewerContainer.value?.focus()),
  saveProgress: () => saveProgressAndUpdateLibrary(),
  updateBookStore,
  refresh: () => initReader(),
  getIllustrations: () => collectIllustrationsFromBook(),
  searchInBook,
  highlightSearchKeyword,
  clearHighlight,
  filePath: props.filePath
});

// 程序关闭前保存进度
const handleBeforeUnload = () => {
  saveProgressAndUpdateLibrary()
}

// 当前选中的文本
const selectedText = ref('')

onMounted(async () => {
  await initReader();
  window.addEventListener('keydown', handleKey);
  window.addEventListener('resize', handleResize);
  window.addEventListener('beforeunload', handleBeforeUnload);
  document.addEventListener('fullscreenchange', handleFullscreenChange);
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
  document.addEventListener('mozfullscreenchange', handleFullscreenChange);
  document.addEventListener('MSFullscreenChange', handleFullscreenChange);
  eventBus.on('note-saved', handleNoteSaved);
  // 每 10 分钟自动保存阅读进度
  autoSaveTimer = setInterval(() => {
    saveProgressAndUpdateLibrary()
  },10 * 60 * 1000)
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKey);
  window.removeEventListener('resize', handleResize);
  window.removeEventListener('beforeunload', handleBeforeUnload);
  document.removeEventListener('fullscreenchange', handleFullscreenChange);
  document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
  document.removeEventListener('mozfullscreenchange', handleFullscreenChange);
  document.removeEventListener('MSFullscreenChange', handleFullscreenChange);
  eventBus.off('note-saved', handleNoteSaved);
  clearTimeout(resizeTimer);
  clearInterval(autoSaveTimer);
  saveProgressAndUpdateLibrary()
  if (rendition) rendition.destroy();
  
  // 注销 tabId 和 EPUB 文件路径的映射
  if (props.tabId) {
    // @ts-ignore
    window.go.main.App.UnregisterEpubTab(props.tabId)
  }
});

// 监听设置变化，实时应用排版
watch(() => [
  settingsStore.fontSize,
  settingsStore.fontFamily,
  settingsStore.paragraphGap,
  settingsStore.lineHeight,
  settingsStore.letterSpacing,
  settingsStore.indent,
  settingsStore.textAlign
], () => {
  applyTypography();
}, { deep: true });

// 监听主题变化，实时应用主题
watch(() => themeStore.currentTheme, () => {
  applyTheme();
});

// 监听侧边栏宽度变化，通知 epub.js 重新计算布局
watch(() => settingsStore.sidebarWidth, () => {
  if (rendition) {
    // 等待 CSS 布局更新后再 resize
    requestAnimationFrame(() => {
      rendition.resize('100%', '100%')
    })
  }
})

// 监听 tab 激活状态变化，确保切换回 tab 时阅读器内容正常显示
watch(() => props.isActive, (newIsActive) => {
  if (newIsActive) {
    // 延迟执行，确保 DOM 已完全渲染
    setTimeout(async () => {
      if (isLoading.value || isInitializing) {
        // 如果还在加载中，等待更长时间再检查
        setTimeout(async () => {
          if (rendition) {
            rendition.resize('100%', '100%');
            applyTypography();
            applyTheme();
            // 延迟 500ms 后触发恢复阅读进度
            setTimeout(async () => {
              await restoreReaderProgress(rendition, null, () => viewerContainer.value?.focus(), props.filePath);
            }, 500);
          }
        }, 300);
      } else if (rendition) {
        rendition.resize('100%', '100%');
        applyTypography();
        applyTheme();
        // 延迟 500ms 后触发恢复阅读进度
        setTimeout(async () => {
          await restoreReaderProgress(rendition, null, () => viewerContainer.value?.focus(), props.filePath);
          // 如果有位置信息，确保显示正确的位置
          if (book?.locations && book.locations.length > 0 && rendition?.location?.start?.cfi) {
            // 确保内容正确渲染
            rendition.display(rendition.location.start.cfi);
          }
        }, 500);
      }
    }, 100);
  }
});

watch(() => props.filePath, async () => {
  await initReader();
});
</script>

<style scoped>
/* 阅读器视图 */
.reader-view {
  position: relative;
  width: 100%;
  height: 100%;
  background-color: var(--bg-color);
  overflow: hidden;
}

/* 阅读内容容器 */
.viewer-container {
  width: 100%;
  height: 100%;
  padding: 60px 100px 60px;
  box-sizing: border-box;
}

/* 页眉栏 */
.reader-header {
  position: absolute;
  top: 20px;
  left: 20px;
  right: 20px;
  height: 32px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  z-index: 10;
}

.chapter-title {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 强制隐藏 iframe 的原生滚动条 */
.viewer-container :deep(iframe) {
  overflow: hidden !important;
}

/* 翻页控制层 */
.reader-controls {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  display: none;
}

/* 翻页按钮 */
.reader-controls .nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0, 0, 0, 0.05);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  width: 70px;
  height: 140px;
  cursor: pointer;
  font-size: 32px;
  color: var(--text-color);
  opacity: 0;
  pointer-events: auto;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-lg);
}

.reader-view:hover .nav-btn {
  opacity: 0.5;
}

.reader-controls .nav-btn:hover {
  opacity: 1;
  background: var(--primary-light);
  border-color: var(--primary-color);
  transform: translateY(-50%) scale(1.08);
}

.nav-btn.left {
  left: 20px;
  border-radius: 0 var(--radius-lg) var(--radius-lg) 0;
  padding-left: 10px;
}

.nav-btn.right {
  right: 20px;
  border-radius: var(--radius-lg) 0 0 var(--radius-lg);
  padding-right: 10px;
}

/* 阅读设置按钮 */
.reader-settings-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 44px;
  height: 44px;
  background: rgba(0, 0, 0, 0.05);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-md);
  cursor: pointer;
  opacity: 0;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
}

.reader-view:hover .reader-settings-btn {
  opacity: 0.7;
}

.reader-settings-btn:hover {
  opacity: 1;
  background: var(--primary-light);
  border-color: var(--primary-color);
}

/* 隐藏元素 */
.hidden {
  visibility: hidden;
  pointer-events: none;
}

/* 全屏模式 */
.reader-view.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background-color: var(--bg-color);
}

.reader-view.fullscreen .viewer-container {
  padding: 80px 150px;
}

.fullscreen-drag-area {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 50px;
  z-index: 9998;
  --wails-draggable: drag;
}



/* 全屏控制栏 */
.fullscreen-controls {
  position: fixed;
  top: 20px;
  right: 20px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
  z-index: 1000;
  opacity: 1;
  transition: opacity 0.3s ease;
  pointer-events: auto;
}

.fullscreen-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  cursor: pointer;
  color: #fff;
  font-size: 0.9rem;
  transition: all var(--transition-normal);
}

.fullscreen-btn:hover {
  background: rgba(0, 0, 0, 0.8);
  border-color: rgba(255, 255, 255, 0.2);
}

.fullscreen-icon {
  font-size: 1rem;
}

.fullscreen-text {
  font-weight: 500;
}

.fullscreen-hint {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.6);
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.fullscreen-controls.hidden {
  opacity: 0;
  pointer-events: none;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .viewer-container {
    padding: 40px 20px;
  }
  
  .nav-btn.left {
    left: 10px;
    width: 50px;
    height: 100px;
  }
  
  .nav-btn.right {
    right: 10px;
    width: 50px;
    height: 100px;
  }
  
  .reader-view.fullscreen .viewer-container {
    padding: 60px 20px;
  }
  
  .fullscreen-toggle-btn {
    top: 15px;
    right: 15px;
    width: 38px;
    height: 38px;
  }
}

</style>