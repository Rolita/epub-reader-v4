<script setup lang="ts">
import { useSettingsStore } from '../stores/settings';
import PaletteIcon from './icons/PaletteIcon.vue';

const emit = defineEmits<{
  (e: 'switch-view', viewName: string): void
}>();

const settingsStore = useSettingsStore();

const fontFamilies = [
  { value: 'Source Han Serif SC Light', label: '思源宋体 Light' },
  { value: 'Source Han Serif SC Medium', label: '思源宋体 Medium' },
  { value: 'LXGW WenKai Medium', label: '霞鹜文楷 Medium' },
];

const textAligns = [
  { value: 'left', label: '左对齐' },
  { value: 'justify', label: '两端对齐' },
];
</script>

<template>
  <div class="layout-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><PaletteIcon :size="22" style="margin-right: 8px;" />阅读排版</h2>
    </div>
    
    <div class="layout-content">
      <!-- 字号 -->
      <div class="setting-item">
        <label class="setting-label">字号: {{ settingsStore.fontSize }}px</label>
        <div class="setting-control">
          <input 
            type="range" 
            v-model.number="settingsStore.fontSize" 
            min="20" 
            max="32" 
            step="1"
            class="setting-slider"
          />
        </div>
      </div>

      <!-- 字体 -->
      <div class="setting-item">
        <label class="setting-label">字体</label>
        <select v-model="settingsStore.fontFamily" class="setting-select">
          <option v-for="font in fontFamilies" :key="font.value" :value="font.value">
            {{ font.label }}
          </option>
        </select>
      </div>

      <!-- 行间距 -->
      <div class="setting-item">
        <label class="setting-label">行间距: {{ settingsStore.lineHeight }}</label>
        <div class="setting-control">
          <input 
            type="range" 
            v-model.number="settingsStore.lineHeight" 
            min="1.0" 
            max="3.0" 
            step="0.1"
            class="setting-slider"
          />
        </div>
      </div>

      <!-- 段间距 -->
      <div class="setting-item">
        <label class="setting-label">段间距: {{ settingsStore.paragraphGap }}px</label>
        <div class="setting-control">
          <input 
            type="range" 
            v-model.number="settingsStore.paragraphGap" 
            min="0" 
            max="60" 
            step="1"
            class="setting-slider"
          />
        </div>
      </div>

      <!-- 首行缩进 -->
      <div class="setting-item">
        <label class="setting-label">首行缩进: {{ settingsStore.indent }}em</label>
        <div class="setting-control">
          <input 
            type="range" 
            v-model.number="settingsStore.indent" 
            min="0" 
            max="5" 
            step="0.5"
            class="setting-slider"
          />
        </div>
      </div>

      <!-- 字间距 -->
      <div class="setting-item">
        <label class="setting-label">字间距: {{ settingsStore.letterSpacing }}px</label>
        <div class="setting-control">
          <input 
            type="range" 
            v-model.number="settingsStore.letterSpacing" 
            min="-2" 
            max="10" 
            step="0.5"
            class="setting-slider"
          />
        </div>
      </div>

      <!-- 文本对齐 -->
      <div class="setting-item">
        <label class="setting-label">文本对齐</label>
        <div class="align-buttons">
          <button 
            v-for="align in textAligns" 
            :key="align.value"
            :class="['align-btn', { active: settingsStore.textAlign === align.value }]"
            @click="settingsStore.textAlign = align.value"
          >
            {{ align.label }}
          </button>
        </div>
      </div>

      
    </div>
  </div>
</template>

<style scoped>
/* 排版设置容器 */
.layout-wrapper {
  width: 100%;
  height: 100%;
  background: var(--sidebar-bg);
  color: var(--text-color);
  display: flex;
  flex-direction: column;
  user-select: none;
}

/* 侧边栏头部 */
.sidebar-header {
  padding: 28px 20px;
  border-bottom: 1px solid var(--border-color);
  background: linear-gradient(180deg, var(--primary-light) 0%, transparent 100%);
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  display: flex;
  align-items: center;
}

/* 设置内容区 */
.layout-content {
  flex: 1;
  padding: 24px 20px;
  overflow-y: auto;
}

/* 设置项 */
.setting-item {
  margin-bottom: 26px;
}

/* 设置标签 */
.setting-label {
  display: block;
  font-size: 0.88rem;
  color: var(--text-secondary);
  margin-bottom: 10px;
  font-weight: 500;
}

/* 设置控制区 */
.setting-control {
  display: flex;
  align-items: center;
  gap: 14px;
}

/* 滑块样式 */
.setting-slider {
  width: 100%;
  height: 6px;
  border-radius: 3px;
  background: var(--border-color);
  appearance: none;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.setting-slider:hover {
  background: rgba(99, 102, 241, 0.2);
}

.setting-slider::-webkit-slider-thumb {
  appearance: none;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  cursor: pointer;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-fast);
}

.setting-slider::-webkit-slider-thumb:hover {
  transform: scale(1.15);
  box-shadow: var(--shadow-lg);
}

.setting-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  cursor: pointer;
  border: none;
  box-shadow: var(--shadow-md);
}

/* 下拉选择框 */
.setting-select {
  width: 100%;
  padding: 12px 14px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 0.9rem;
  cursor: pointer;
  transition: all var(--transition-fast);
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='20' height='20' viewBox='0 0 24 24' fill='none' stroke='%236B7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
}

.setting-select:hover {
  border-color: var(--primary-color);
}

.setting-select:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
}

/* 对齐按钮组 */
.align-buttons {
  display: flex;
  gap: 10px;
}

.align-btn {
  flex: 1;
  padding: 12px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-secondary);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-weight: 500;
}

.align-btn:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.align-btn.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  border-color: var(--primary-color);
  box-shadow: var(--shadow-md);
}

/* 滚动条美化 */
.layout-content::-webkit-scrollbar {
  width: 4px;
}

.layout-content::-webkit-scrollbar-track {
  background: transparent;
}

.layout-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.layout-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>