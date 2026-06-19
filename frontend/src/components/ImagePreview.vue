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
      @contextmenu.prevent="showContextMenu"
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

      <!-- 右键菜单 -->
      <div 
        v-if="contextMenuVisible" 
        class="preview-context-menu"
        :style="{ left: contextMenuPos.x + 'px', top: contextMenuPos.y + 'px' }"
        @click.stop
      >
        <button @click="saveImage">保存图片</button>
        <button @click="copyImage">复制图片</button>
        <button @click="contextMenuVisible = false">取消</button>
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
const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const startPosX = ref(0)
const startPosY = ref(0)

// 右键菜单
const contextMenuVisible = ref(false)
const contextMenuPos = ref({ x: 0, y: 0 })

// 当前图片源
const currentSrc = computed(() => props.src)

// 预览图片引用
const previewImg = ref<HTMLImageElement | null>(null)

// 样式计算
const wrapperStyle = computed(() => ({
  cursor: isDragging.value ? 'grabbing' : 'grab'
}))

const imageStyle = computed(() => ({
  transform: `translate(${positionX.value}px, ${positionY.value}px) scale(${scale.value})`,
  transition: isDragging.value ? 'none' : 'transform 0.1s ease-out'
}))

// 关闭预览
const close = () => {
  emit('close')
  // 重置状态
  scale.value = 1
  positionX.value = 0
  positionY.value = 0
  contextMenuVisible.value = false
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

// 显示右键菜单
const showContextMenu = (e: MouseEvent) => {
  e.preventDefault()
  contextMenuPos.value = { x: e.clientX, y: e.clientY }
  contextMenuVisible.value = true
}

// 保存图片
const saveImage = async () => {
  contextMenuVisible.value = false
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
  contextMenuVisible.value = false
  try {
    const response = await fetch(currentSrc.value)
    const blob = await response.blob()
    await navigator.clipboard.write([
      new ClipboardItem({ [blob.type]: blob })
    ])
  } catch (err) {
    console.error('复制图片失败:', err)
  }
}

// 监听打开/关闭
watch(() => props.visible, (newVal) => {
  if (newVal) {
    // 打开时重置状态
    scale.value = 1
    positionX.value = 0
    positionY.value = 0
    contextMenuVisible.value = false
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
  right: 80px;
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

.preview-context-menu {
  position: fixed;
  background: #2a2a2a;
  border: 1px solid #444;
  border-radius: 8px;
  padding: 6px 0;
  min-width: 140px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  z-index: 100;
}

.preview-context-menu button {
  display: block;
  width: 100%;
  padding: 10px 20px;
  border: none;
  background: none;
  color: white;
  text-align: left;
  cursor: pointer;
  font-size: 14px;
}

.preview-context-menu button:hover {
  background: #3a3a3a;
}
</style>
