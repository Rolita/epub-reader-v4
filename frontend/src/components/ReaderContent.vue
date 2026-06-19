<template>
  <div 
    class="reader-view" 
    :class="{ 'fullscreen': isFullscreen }"
    @mousemove="handleMouseMove"
  >
    <!-- 加载动画 -->
    <LoadingOverlay v-if="isLoading" />
    
    <div ref="viewerContainer" class="viewer-container" :class="{ 'hidden': isLoading }"></div>
    
    <div class="reader-controls" :class="{ 'hidden': isLoading }">
      <button @click="prevPage" class="nav-btn left">〈</button>
      <button @click="nextPage" class="nav-btn right">〉</button>
    </div>

    <!-- 全屏控制栏 -->
    <div 
      v-if="!isLoading && isFullscreen" 
      class="fullscreen-controls" 
      :class="{ 'hidden': !showFullscreenControls }"
      @mouseenter="showControls"
    >
      <button @click="exitFullscreen" class="fullscreen-btn exit">
        <span class="fullscreen-icon">⟲</span>
        <span class="fullscreen-text">退出全屏</span>
      </button>
      <div class="fullscreen-hint">按 Esc 或 F11 退出全屏</div>
    </div>

    <!-- 非全屏时的全屏按钮 -->
    <button 
      v-if="!isLoading && !isFullscreen" 
      @click="enterFullscreen" 
      class="fullscreen-toggle-btn"
      title="全屏阅读 (F11)"
    >
      ➤
    </button>

    <!-- 图片预览组件 -->
    <ImagePreview 
      :visible="imagePreviewVisible" 
      :src="previewImageSrc"
      :alt="previewImageAlt"
      @close="closeImagePreview"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import Epub from 'epubjs';
import ImagePreview from './ImagePreview.vue';
import LoadingOverlay from './LoadingOverlay.vue';
import { useBookStore, TocItem } from '../stores/book';
import { useSettingsStore } from '../stores/settings';
import { useThemeStore } from '../stores/theme';

const props = defineProps<{ filePath: string }>();
const viewerContainer = ref<HTMLElement | null>(null);

// 图片预览相关状态
const imagePreviewVisible = ref(false)
const previewImageSrc = ref('')
const previewImageAlt = ref('')

// 全屏状态
const isFullscreen = ref(false)
const showFullscreenControls = ref(true)
let hideControlsTimer: any = null

// 图片预览方法
const openImagePreview = (src: string, alt: string = '') => {
  previewImageSrc.value = src
  previewImageAlt.value = alt
  imagePreviewVisible.value = true
}

const closeImagePreview = () => {
  imagePreviewVisible.value = false
}

// 将图片点击事件注入到 epub 渲染内容中
const injectImageClickHandler = () => {
  if (!rendition) return

  // 使用 hooks.content 在 iframe 内容加载时绑定事件
  rendition.hooks.content.register((contents: any) => {
    const doc = contents.document
    if (!doc) return

    const bindEventsToImages = () => {
      // 1. 绑定普通 <img>
      const images = doc.querySelectorAll('img')
      images.forEach((img: HTMLImageElement) => {
        if (img.dataset.previewBound) return
        img.dataset.previewBound = 'true'
        img.style.cursor = 'zoom-in'
        bindImageEvents(img)
      })

      // 2. 绑定 SVG 内的 <image> 标签
      const svgImages = doc.querySelectorAll('svg image')
      svgImages.forEach((svgImg: any) => {
        if (svgImg.dataset.previewBound) return
        svgImg.dataset.previewBound = 'true'
        svgImg.style.cursor = 'zoom-in'
        // SVG image 需要通过 xlink:href 或 href 获取 src
        let src = svgImg.getAttribute('xlink:href') || svgImg.getAttribute('href')
        if (src) {
          // 相对路径需要转换为完整 URL
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
          svgImg.addEventListener('contextmenu', (e: Event) => {
            e.preventDefault()
            previewImageSrc.value = src
            previewImageAlt.value = svgImg.getAttribute('alt') || ''
          })
        }
      })
    }

    const bindImageEvents = (img: HTMLImageElement) => {
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

      img.addEventListener('contextmenu', (e) => {
        e.preventDefault()
        let src = img.src
        if (!src || src === 'about:blank') {
          const currentSrc = (img as any).currentSrc || img.getAttribute('src')
          if (currentSrc) src = currentSrc
        }
        previewImageSrc.value = src
        previewImageAlt.value = img.alt || ''
      })
    }

    // 立即绑定一次
    bindEventsToImages()

    // 监听动态加载的图片
    const observer = new MutationObserver(() => {
      bindEventsToImages()
    })
    observer.observe(doc.body || doc.documentElement, {
      childList: true,
      subtree: true
    })
  })

  // 同时监听 rendered 确保首次渲染也能触发
  rendition.on('rendered', () => {
    // rendered 时再次触发绑定，防止 hooks.content 未触发
    setTimeout(() => {
      const doc = rendition?.manager?.container?.contentDocument
      if (doc) {
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
      }
    }, 100)
  })
}

let book: any = null;
let rendition: any = null;
let isScrolling = false; // 防抖标志位
let saveTimer: any = null; // 进度保存防抖计时器

// 加载状态
const isLoading = ref(true)

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
      'line-height': settingsStore.lineHeight + ' !important'
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
      margin-top: 0 !important;      /* 严禁使用上边距，防止叠加 */
      margin-bottom: ${settingsStore.paragraphGap}px !important; /* 段间距：只用下边距，统一间距感 */
      /* 2. 行间距：使用纯倍数，保持阅读舒适度 */
      line-height: ${settingsStore.lineHeight} !important;
      /* 3. 中文排版精髓：首行缩进 */
      text-indent: ${settingsStore.indent}em !important;
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
  `;

  rendition.themes.append(css);
};

// 应用主题到 EPUB 内容
const applyTheme = () => {
  if (!rendition || !rendition.themes) {
    console.warn('applyTheme: rendition 或 rendition.themes 未就绪');
    return;
  }

  const colors = themeStore.themeColors;
  
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

  // 追加更强制的 CSS 覆盖
  rendition.themes.append(`
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
  `);
};

// 清理内联样式
const clearInlineStyles = () => {
  if (!rendition) return;
  
  // 在每一章渲染后清理内联样式
  rendition.on('rendered', (section: any) => {
    const doc = section.document;
    if (!doc) return;
    
    const elements = doc.querySelectorAll('*');
    elements.forEach((el: any) => {
      // 移除内联的颜色和背景色
      el.style.removeProperty('color');
      el.style.removeProperty('background-color');
      el.style.removeProperty('background');
      el.style.removeProperty('text-shadow');
      
      // 移除内联的字号和行高（解决硬编码样式问题）
      el.style.removeProperty('font-size');
      el.style.removeProperty('line-height');
      el.style.removeProperty('font-family');
      el.style.removeProperty('font-weight');
      el.style.removeProperty('text-align');
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

const prevPage = () => rendition?.prev();
const nextPage = () => rendition?.next();

const initReader = async () => {
  try {
    // 开始加载
    isLoading.value = true
    
    if (rendition) {
      rendition.destroy();
      rendition = null;
    }
    book = null;

    // 1. 从 Go 获取文件内容
    // @ts-ignore
    const base64Data = await window.go.main.App.GetFileBytes(props.filePath);
    const buffer = base64ToBuffer(base64Data);

    // 2. 初始化 epub.js
    book = Epub(buffer);

    // 3. 加载书籍导航信息（目录）
    await book.ready;
    const navigation = book.navigation;
    const toc = navigation ? flattenToc(navigation.toc) : [];
    
    // 4. 更新 Pinia store
    bookStore.setActiveBook(props.filePath, toc);

    if (viewerContainer.value) {
      rendition = book.renderTo(viewerContainer.value, {
        width: '100%',
        height: '100%',
        flow: 'paginated',
        spread: 'always',
      });

      // 5. 从 Go 读取上一次的 CFI
      // @ts-ignore
      const savedCfi = await window.go.main.App.GetProgress(props.filePath);

      // 6. 注入滚轮监听钩子
      rendition.hooks.content.register((contents: any) => {
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
        }, { passive: false });
      });

      // 7. 注册内联样式清理（在每章渲染后执行）
      clearInlineStyles();

      // 8. 注入图片点击预览功能
      injectImageClickHandler();

      // 9. 使用渲染钩子确保主题在内容渲染后应用
      rendition.hooks.render.register(() => {
        console.log('Rendition 渲染钩子触发');
        // 确保 themes 对象已就绪后再应用
        if (rendition && rendition.themes) {
          applyTypography();
          applyTheme();
        }
      });

      // 9. 监听翻页事件并"防抖"保存进度
      rendition.on('relocated', (location: any) => {
        const cfi = location.start.cfi;
        const percentage = location.start.percentage || 0;
        
        // 防抖处理：停止翻页后延迟保存，避免磁盘频繁 IO
        clearTimeout(saveTimer);
        saveTimer = setTimeout(async () => {
          // 构建包含 CFI 和百分比的进度对象
          const progressData = {
            cfi: cfi,
            percentage: percentage,
            timestamp: Date.now()
          };
          // @ts-ignore
          await window.go.main.App.SaveProgress(props.filePath, JSON.stringify(progressData));
          console.log("进度已保存至 config.json:", progressData);
        }, 500);
      });

      // 9. 读取并恢复阅读进度（CFI + 百分比双重验证）
      let savedProgress: any = null;
      if (savedCfi) {
        try {
          savedProgress = JSON.parse(savedCfi);
          console.log("读取到保存的进度:", savedProgress);
        } catch (e) {
          console.error("解析进度数据失败:", e);
        }
      }
      
      if (savedProgress && savedProgress.cfi) {
        // 优先使用 CFI 跳转
        await rendition.display(savedProgress.cfi);
        
        // 验证跳转后的百分比是否匹配
        const currentLocation = rendition.currentLocation();
        const currentPercentage = currentLocation.start.percentage || 0;
        
        if (savedProgress.percentage && Math.abs(currentPercentage - savedProgress.percentage) > 0.05) {
          console.warn("CFI 漂移检测到！使用百分比兜底跳转");
          // 使用百分比生成新的 CFI 进行兜底
          if (book.locations) {
            const fallbackCfi = book.locations.cfiFromPercentage(savedProgress.percentage);
            if (fallbackCfi) {
              await rendition.display(fallbackCfi);
              console.log("百分比兜底跳转完成:", savedProgress.percentage);
            }
          }
        }
      } else {
        await rendition.display();
      }

      // 10. 在 display 完成后手动再应用一次（兜底）
      if (rendition && rendition.themes) {
        console.log('display 后手动应用主题');
        applyTypography();
        applyTheme();
      }

      console.log("阅读器初始化完成");
      
      // 加载完成，隐藏加载动画
      isLoading.value = false
    }
  } catch (err) {
    console.error("阅读器启动失败:", err);
    // 即使失败也要隐藏加载动画
    isLoading.value = false
  }
};

const handleKey = (e: KeyboardEvent) => {
  if (e.key === 'ArrowLeft') prevPage();
  if (e.key === 'ArrowRight') nextPage();
  
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
    showFullscreenControls.value = true;
    startHideControlsTimer();
    
    // 延迟重新渲染，确保全屏切换完成
    setTimeout(() => {
      if (rendition) {
        rendition.resize('100%', '100%');
        console.log('全屏模式下重新渲染完成');
      }
    }, 300);
  }).catch((err: Error) => {
    console.error('进入全屏失败:', err);
  });
};

// 显示控制栏
const showControls = () => {
  showFullscreenControls.value = true;
  startHideControlsTimer();
};

// 启动自动隐藏计时器
const startHideControlsTimer = () => {
  clearTimeout(hideControlsTimer);
  hideControlsTimer = setTimeout(() => {
    showFullscreenControls.value = false;
  }, 1000);
};

// 鼠标移动处理
const handleMouseMove = () => {
  if (isFullscreen.value) {
    showControls();
  }
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

// 跳转到指定章节
const jumpTo = (href: string) => {
  if (rendition) {
    rendition.display(href).then(() => {
      // 跳转后切回焦点，确保键盘翻页可用
      viewerContainer.value?.focus();
    });
  }
};

// 暴露方法给父组件调用
defineExpose({ jumpTo });

onMounted(async () => {
  await initReader();
  window.addEventListener('keydown', handleKey);
  window.addEventListener('resize', handleResize);
  document.addEventListener('fullscreenchange', handleFullscreenChange);
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
  document.addEventListener('mozfullscreenchange', handleFullscreenChange);
  document.addEventListener('MSFullscreenChange', handleFullscreenChange);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKey);
  window.removeEventListener('resize', handleResize);
  document.removeEventListener('fullscreenchange', handleFullscreenChange);
  document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
  document.removeEventListener('mozfullscreenchange', handleFullscreenChange);
  document.removeEventListener('MSFullscreenChange', handleFullscreenChange);
  clearTimeout(resizeTimer);
  clearTimeout(hideControlsTimer);
  if (rendition) rendition.destroy();
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
  padding: 60px 100px;
  box-sizing: border-box;
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

/* 阅读进度指示条 */
.reader-progress {
  position: absolute;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  width: 60%;
  height: 4px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 2px;
  cursor: pointer;
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.reader-view:hover .reader-progress {
  opacity: 1;
}

.reader-progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--primary-color) 0%, var(--accent-color) 100%);
  border-radius: 2px;
  transition: width var(--transition-fast);
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

/* 全屏切换按钮 */
.fullscreen-toggle-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 44px;
  height: 44px;
  background: rgba(0, 0, 0, 0.05);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-md);
  cursor: pointer;
  opacity: 0;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  color: var(--text-color);
}

.reader-view:hover .fullscreen-toggle-btn {
  opacity: 0.7;
}

.fullscreen-toggle-btn:hover {
  opacity: 1;
  background: var(--primary-light);
  border-color: var(--primary-color);
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