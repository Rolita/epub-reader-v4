<template>
  <!-- 图片预览遮罩 -->
  <Teleport to="body">
    <div 
      v-if="visible" 
      class="image-preview-overlay"
      @click="handleOverlayClick"
      @mousedown="handleMouseDown"
      @mousemove="handleMouseMove"
      @mouseup="handleMouseUp"
      @mouseleave="handleMouseUp"
      @wheel.prevent="handleWheel"
      @contextmenu.prevent="rotateLeft"
    >
      <!-- 关闭按钮 -->
      <button class="preview-close" @click="close">×</button>
      
      <!-- 保存按钮 -->
      <button class="preview-save" @click="saveImage" title="保存图片">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
      </button>
      
      <!-- 复制按钮 -->
      <button class="preview-copy" @click="copyImage" title="复制图片">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
        </svg>
      </button>
      
      <!-- 复制成功提示 -->
      <div v-if="showCopySuccess" class="copy-success-tip">已复制</div>
      
      <!-- 左旋转按钮 -->
      <button class="preview-rotate preview-rotate-left" @click="rotateLeft" title="向左旋转">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="1 4 1 10 7 10"/>
          <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
        </svg>
      </button>
      
      <!-- 右旋转按钮 -->
      <button class="preview-rotate preview-rotate-right" @click="rotateRight" title="向右旋转">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"/>
          <path d="M20.49 15a9 9 0 1 1-2.13-9.36L23 10"/>
        </svg>
      </button>

      <!-- 大图预览 -->
      <div class="preview-image-wrapper" :style="wrapperStyle">
        <img 
          ref="previewImg"
          :src="currentSrc" 
          :alt="alt"
          class="preview-image"
          :style="imageStyle"
          draggable="false"
          @mousedown.stop="handleImageMouseDown"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  src: string
  alt?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

// 状态
const scale = ref(1)
const positionX = ref(0)
const positionY = ref(0)
const rotation = ref(0)
const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const startPosX = ref(0)
const startPosY = ref(0)

// 复制成功提示
const showCopySuccess = ref(false)
let copySuccessTimer: number | null = null

// 当前图片源
const currentSrc = computed(() => props.src)

// 预览图片引用
const previewImg = ref<HTMLImageElement | null>(null)

// 样式计算
const wrapperStyle = computed(() => ({
  cursor: isDragging.value ? 'grabbing' : 'grab'
}))

const imageStyle = computed(() => ({
  transform: `translate(${positionX.value}px, ${positionY.value}px) scale(${scale.value}) rotate(${rotation.value}deg)`,
  transition: isDragging.value ? 'none' : 'transform 0.1s ease-out'
}))

// 关闭预览
const close = () => {
  emit('close')
  // 重置状态
  scale.value = 1
  positionX.value = 0
  positionY.value = 0
  rotation.value = 0
}

// 向左旋转
const rotateLeft = () => {
  rotation.value -= 90
}

// 向右旋转
const rotateRight = () => {
  rotation.value += 90
}

// 点击遮罩关闭
const handleOverlayClick = (e: MouseEvent) => {
  if (e.target === e.currentTarget) {
    close()
  }
}

// 鼠标按下（准备拖拽）
const handleMouseDown = (e: MouseEvent) => {
  if (e.button !== 0) return // 只响应左键
  isDragging.value = true
  dragStartX.value = e.clientX
  dragStartY.value = e.clientY
  startPosX.value = positionX.value
  startPosY.value = positionY.value
}

// 拖拽中
const handleMouseMove = (e: MouseEvent) => {
  if (!isDragging.value) return
  const deltaX = e.clientX - dragStartX.value
  const deltaY = e.clientY - dragStartY.value
  positionX.value = startPosX.value + deltaX
  positionY.value = startPosY.value + deltaY
}

// 鼠标释放
const handleMouseUp = () => {
  isDragging.value = false
}

// 图片拖拽
const handleImageMouseDown = (e: MouseEvent) => {
  e.stopPropagation()
  handleMouseDown(e)
}

// 滚轮缩放
const handleWheel = (e: WheelEvent) => {
  e.preventDefault()
  const delta = e.deltaY > 0 ? 0.9 : 1.1
  const newScale = Math.min(Math.max(scale.value * delta, 0.2), 5)
  scale.value = newScale
}

// 双击重置
const handleDoubleClick = () => {
  scale.value = 1
  positionX.value = 0
  positionY.value = 0
}

// 保存图片
const saveImage = async () => {
  try {
    // 创建一个 <a> 标签下载
    const link = document.createElement('a')
    link.href = currentSrc.value
    // 从 URL 中提取文件名，或生成默认名
    const urlParts = currentSrc.value.split('/')
    let fileName = urlParts[urlParts.length - 1] || 'image'
    // 确保有扩展名
    if (!/\.(jpg|jpeg|png|gif|svg|webp)$/i.test(fileName)) {
      fileName += '.png'
    }
    link.download = fileName
    link.target = '_blank'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (err) {
    console.error('保存图片失败:', err)
  }
}

// 复制图片
const copyImage = async () => {
  try {
    // @ts-ignore
    await window.go.main.App.CopyImageToClipboard(currentSrc.value)
    
    // 显示复制成功提示
    showCopySuccess.value = true
    if (copySuccessTimer) clearTimeout(copySuccessTimer)
    copySuccessTimer = window.setTimeout(() => {
      showCopySuccess.value = false
    }, 2000)
    
    console.log('复制成功')
  } catch (err) {
    console.error('复制图片失败:', err)
    alert('复制失败，请尝试右键保存图片')
  }
}

// 监听打开/关闭
watch(() => props.visible, (newVal) => {
  if (newVal) {
    // 打开时重置状态
    scale.value = 1
    positionX.value = 0
    positionY.value = 0
    rotation.value = 0
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.9);
  z-index: 99999;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: grab;
}

.preview-close {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 44px;
  height: 44px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  font-size: 28px;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  z-index: 10;
}

.preview-close:hover {
  background: rgba(255, 255, 255, 0.2);
}

.preview-save {
  position: absolute;
  top: 20px;
  right: 74px;
  width: 44px;
  height: 44px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  z-index: 10;
}

.preview-save:hover {
  background: rgba(255, 255, 255, 0.2);
}

.preview-copy {
  position: absolute;
  top: 20px;
  right: 140px;
  width: 44px;
  height: 44px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  z-index: 10;
}

.preview-copy:hover {
  background: rgba(255, 255, 255, 0.2);
}

.preview-rotate {
  position: absolute;
  top: 20px;
  width: 44px;
  height: 44px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  z-index: 10;
}

.preview-rotate:hover {
  background: rgba(255, 255, 255, 0.2);
}

.preview-rotate-left {
  right: 200px;
}

.preview-rotate-right {
  right: 260px;
}

.copy-success-tip {
  position: absolute;
  top: 80px;
  right: 140px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 14px;
  pointer-events: none;
  animation: fadeInOut 2s ease-in-out;
  z-index: 20;
}

@keyframes fadeInOut {
  0% {
    opacity: 0;
    transform: translateY(-10px);
  }
  20% {
    opacity: 1;
    transform: translateY(0);
  }
  80% {
    opacity: 1;
    transform: translateY(0);
  }
  100% {
    opacity: 0;
    transform: translateY(-10px);
  }
}

.preview-image-wrapper {
  max-width: 90vw;
  max-height: 90vh;
  overflow: visible;
}

.preview-image {
  max-width: 90vw;
  max-height: 90vh;
  object-fit: contain;
  user-select: none;
  pointer-events: auto;
}
</style>
