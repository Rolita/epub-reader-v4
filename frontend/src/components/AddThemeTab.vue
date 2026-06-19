<template>
  <main class="add-theme-container">
    <div class="content-area">
      <div class="add-theme-content">
        <h1>{{ isEditing ? '编辑主题' : '添加新主题' }}</h1>
        
        <div class="theme-form">
          <!-- 主题名称 -->
          <div class="form-group">
            <label>主题名称</label>
            <input 
              type="text" 
              v-model="themeName" 
              placeholder="例如：墨韵龙鳞"
              class="form-input"
            />
          </div>
          
          <!-- 主题图标 -->
          <div class="form-group">
            <label>主题图标</label>
            <div class="icon-selector">
              <button 
                v-for="icon in availableIcons" 
                :key="icon"
                :class="['icon-btn', { active: selectedIcon === icon }]"
                @click="selectedIcon = icon"
              >
                <component :is="iconComponents[icon]" :size="20" />
              </button>
            </div>
          </div>

          <!-- 颜色配置 + 预览 并排 -->
          <div class="color-preview-row">
            <div class="color-grid">
              <div class="color-item">
                <label>主色调</label>
                <div class="color-picker-wrapper">
                  <input type="color" v-model="primaryColor" class="color-picker" />
                  <span class="color-value">{{ primaryColor }}</span>
                </div>
              </div>
              <div class="color-item">
                <label>强调色</label>
                <div class="color-picker-wrapper">
                  <input type="color" v-model="accentColor" class="color-picker" />
                  <span class="color-value">{{ accentColor }}</span>
                </div>
              </div>
              <div class="color-item">
                <label>背景色</label>
                <div class="color-picker-wrapper">
                  <input type="color" v-model="bgColor" class="color-picker" />
                  <span class="color-value">{{ bgColor }}</span>
                </div>
              </div>
              <div class="color-item">
                <label>侧边栏色</label>
                <div class="color-picker-wrapper">
                  <input type="color" v-model="sidebarColor" class="color-picker" />
                  <span class="color-value">{{ sidebarColor }}</span>
                </div>
              </div>
              <div class="color-item">
                <label>功能栏色</label>
                <div class="color-picker-wrapper">
                  <input type="color" v-model="functionBarColor" class="color-picker" />
                  <span class="color-value">{{ functionBarColor }}</span>
                </div>
              </div>
              <div class="color-item">
                <label>边框色</label>
                <div class="color-picker-wrapper">
                  <input type="color" v-model="borderColorManual" class="color-picker" />
                  <span class="color-value">{{ borderColorManual }}</span>
                </div>
              </div>
            </div>

            <!-- 预览 -->
            <div class="theme-preview-container">
              <label class="preview-label">主题预览</label>
              <div class="preview-wrapper" :style="{ background: bgColor, borderColor: borderColorManual }">
                <div class="preview-header" :style="{ background: functionBarColor }">
                  <div class="preview-title-row">
                    <span 
                      class="preview-title" 
                      :style="{ 
                        background: `linear-gradient(135deg, ${primaryColor} 0%, ${accentColor} 100%)`,
                        WebkitBackgroundClip: 'text',
                        WebkitTextFillColor: 'transparent',
                        backgroundClip: 'text'
                      }"
                    >EPUB Reader</span>
                  </div>
                </div>
                <div class="preview-body" :style="{ background: bgColor }">
                  <div class="preview-card" :style="{ borderColor: borderColorManual }">
                    <p :style="{ color: textColor }">这是一段示例文字，用于预览主题效果。</p>
                  </div>
                </div>
                <div class="preview-actions">
                  <button 
                    class="preview-action-btn save-btn" 
                    :style="{ 
                      background: `linear-gradient(135deg, ${primaryColor} 0%, ${accentColor} 100%)`,
                      color: '#fff'
                    }"
                    @click="handleSave"
                  >
                    {{ isEditing ? '保存' : '添加' }}
                  </button>
                  <button 
                    class="preview-action-btn reset-btn" 
                    :style="{ 
                      background: bgColor,
                      color: textColor,
                      borderColor: borderColorManual
                    }"
                    @click="handleReset"
                  >
                    重置
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useThemeStore } from '../stores/theme'
import BookOpenIcon from './icons/BookOpenIcon.vue'
import CloudIcon from './icons/CloudIcon.vue'
import MoonIcon from './icons/MoonIcon.vue'
import SunIcon from './icons/SunIcon.vue'
import StarIcon from './icons/StarIcon.vue'
import BookIcon from './icons/BookIcon.vue'

const themeStore = useThemeStore()

const props = defineProps<{
  editThemeId?: string
}>()

const emit = defineEmits<{
  (e: 'saved'): void
}>()

const isEditing = computed(() => !!props.editThemeId)

const themeName = ref('')
const selectedIcon = ref('star')
const primaryColor = ref('#4A90D9')
const accentColor = ref('#67B8DE')
const bgColor = ref('#FFFFFF')
const sidebarColor = ref('#FFFFFF')
const functionBarColor = ref('#FFFFFF')
const borderColorManual = ref('#E8E3DC')

// 编辑模式时回填数据
onMounted(() => {
  if (props.editThemeId) {
    const theme = themeStore.themes.find(t => t.id === props.editThemeId)
    if (theme) {
      // 如果是默认主题，自动添加“（副本）”后缀提示用户
      if (themeStore.isDefaultTheme(props.editThemeId)) {
        themeName.value = `${theme.name}（副本）`
      } else {
        themeName.value = theme.name
      }
      selectedIcon.value = theme.icon || 'star'
      primaryColor.value = theme.primary
      accentColor.value = theme.accent
      bgColor.value = theme.bg
      sidebarColor.value = theme.sidebar
      functionBarColor.value = theme.functionBar
      borderColorManual.value = theme.border || '#E8E3DC'
    }
  }
})

// 根据背景色自动计算文字色
const textColor = computed(() => {
  const bg = bgColor.value
  const r = parseInt(bg.slice(1, 3), 16)
  const g = parseInt(bg.slice(3, 5), 16)
  const b = parseInt(bg.slice(5, 7), 16)
  const brightness = (r * 299 + g * 587 + b * 114) / 1000
  return brightness > 128 ? '#2D2D2D' : '#E2E8F0'
})

// 根据背景色自动计算边框色
const borderColor = computed(() => {
  const bg = bgColor.value
  const r = parseInt(bg.slice(1, 3), 16)
  const g = parseInt(bg.slice(3, 5), 16)
  const b = parseInt(bg.slice(5, 7), 16)
  const brightness = (r * 299 + g * 587 + b * 114) / 1000
  return brightness > 128 ? '#E8E3DC' : '#334155'
})

const availableIcons = ['book-open', 'cloud', 'moon', 'sun', 'star']

const iconComponents: Record<string, any> = {
  'book-open': BookOpenIcon,
  'cloud': CloudIcon,
  'moon': MoonIcon,
  'sun': SunIcon,
  'star': StarIcon
}

// 判断是否为浅色主题（用于按钮文字颜色）
const isLightTheme = computed(() => {
  const bg = bgColor.value
  const r = parseInt(bg.slice(1, 3), 16)
  const g = parseInt(bg.slice(3, 5), 16)
  const b = parseInt(bg.slice(5, 7), 16)
  const brightness = (r * 299 + g * 587 + b * 114) / 1000
  return brightness > 128
})

// 生成默认主题名称
const generateDefaultName = () => {
  let idx = 1
  const existingNames = new Set(themeStore.themes.map(t => t.name))
  while (existingNames.has(`未命名主题${idx}`)) {
    idx++
  }
  return `未命名主题${idx}`
}

const handleSave = () => {
  const name = themeName.value || generateDefaultName()
  
  const themeData = {
    name,
    icon: selectedIcon.value,
    primary: primaryColor.value,
    accent: accentColor.value,
    bg: bgColor.value,
    text: textColor.value,
    sidebar: sidebarColor.value,
    functionBar: functionBarColor.value,
    border: borderColorManual.value  // 使用手动设置的边框色
  }
  
  if (isEditing.value && props.editThemeId) {
    // 编辑模式
    if (themeStore.isDefaultTheme(props.editThemeId)) {
      // 如果是默认主题，先创建副本
      const newThemeId = themeStore.duplicateDefaultTheme(props.editThemeId)
      if (newThemeId) {
        // 更新副本
        themeStore.updateTheme(newThemeId, themeData)
        // 切换到新创建的主题
        themeStore.setTheme(newThemeId)
      }
    } else {
      // 自定义主题直接更新
      themeStore.updateTheme(props.editThemeId, themeData)
    }
    emit('saved')
  } else {
    // 新建模式
    themeStore.addTheme({
      id: `custom-${Date.now()}`,
      ...themeData
    })
    emit('saved')
  }
}

const handleReset = () => {
  themeName.value = ''
  selectedIcon.value = 'star'
  primaryColor.value = '#4A90D9'
  accentColor.value = '#67B8DE'
  bgColor.value = '#FFFFFF'
  sidebarColor.value = '#FFFFFF'
  functionBarColor.value = '#FFFFFF'
  borderColorManual.value = '#E8E3DC'
}
</script>

<style scoped>
.add-theme-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-color);
}

.content-area {
  flex: 1;
  padding: 28px 36px;
  overflow-y: auto;
  background-color: var(--function-bar-bg);
}

.add-theme-content {
  max-width: 720px;
  margin: 0 auto;
}

.add-theme-content h1 {
  margin-bottom: 20px;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
}

/* 表单 */
.theme-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-primary);
}

.form-input {
  padding: 10px 14px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 0.9rem;
  transition: all var(--transition-normal);
}

.form-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
}

/* 图标选择器 */
.icon-selector {
  display: flex;
  gap: 8px;
}

.icon-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-primary);
  transition: all var(--transition-normal);
}

.icon-btn:hover {
  border-color: var(--primary-color);
  background: var(--primary-light);
}

.icon-btn.active {
  border-color: var(--primary-color);
  background: var(--primary-light);
  color: var(--primary-color);
}

/* 颜色网格 + 预览并排 */
.color-preview-row {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.color-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.color-item label {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-secondary);
}

/* 颜色选择器 */
.color-picker-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-picker {
  width: 36px;
  height: 36px;
  padding: 0;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  background: transparent;
}

.color-picker::-webkit-color-swatch-wrapper {
  padding: 3px;
}

.color-picker::-webkit-color-swatch {
  border-radius: var(--radius-sm);
  border: none;
}

.color-value {
  font-size: 0.78rem;
  color: var(--text-muted);
  font-family: monospace;
}

/* 颜色配置网格 */
.color-grid {
  flex: 1;
  min-width: 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 16px;
}

/* 预览区 */
.theme-preview-container {
  flex-shrink: 0;
  width: 280px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.preview-label {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-secondary);
}

/* 预览包装器 */
.preview-wrapper {
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  min-height: 200px;
  display: flex;
  flex-direction: column;
}

/* 预览顶部标题栏 */
.preview-header {
  padding: 24px 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.preview-logo {
  flex-shrink: 0;
}

.preview-title {
  font-size: 1.5rem;
  font-weight: 700;
}

/* 预览主体内容 */
.preview-body {
  flex: 1;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

/* 预览卡片 - 展示边框色效果 */
.preview-card {
  padding: 16px 20px;
  border: 2px solid;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(8px);
}

.preview-welcome {
  font-size: 1rem;
  text-align: center;
}

.preview-welcome p {
  margin: 0;
  opacity: 0.8;
}

/* 预览操作按钮 */
.preview-actions {
  padding: 20px 32px;
  display: flex;
  justify-content: center;
  gap: 16px;
}

.preview-action-btn {
  padding: 12px 32px;
  border-radius: var(--radius-md);
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  border: 1px solid;
}

.preview-action-btn.save-btn {
  border-color: transparent;
}

.preview-action-btn.save-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.preview-action-btn.save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.preview-action-btn.reset-btn {
  border-color: inherit;
}

.preview-action-btn.reset-btn:hover {
  opacity: 0.8;
}
</style>