<script setup lang="ts">
import LayoutGridIcon from './icons/LayoutGridIcon.vue';
import { useSettingsStore } from '../stores/settings';

const emit = defineEmits<{
  (e: 'switch-view', viewName: string): void
}>();

const settingsStore = useSettingsStore();
</script>

<template>
  <div class="bookshelf-layout-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><LayoutGridIcon :size="22" style="margin-right: 8px;" />书架布局</h2>
    </div>
    
    <div class="layout-content">
      <div class="setting-item">
        <label class="setting-label">列数: {{ settingsStore.bookshelfColumns }}列</label>
        <div class="setting-control">
          <input 
            type="range" 
            v-model.number="settingsStore.bookshelfColumns" 
            min="3" 
            max="8" 
            step="1"
            class="setting-slider"
          />
        </div>
        <div class="columns-quick-select">
          <button 
            v-for="col in 6" 
            :key="col"
            :class="['col-btn', { active: settingsStore.bookshelfColumns === col + 2 }]"
            @click="settingsStore.bookshelfColumns = col + 2"
          >
            {{ col + 2 }}列
          </button>
        </div>
      </div>

      <div class="setting-item">
        <label class="setting-label">封面间距: {{ settingsStore.coverGap }}px</label>
        <div class="setting-control">
          <input 
            type="range" 
            v-model.number="settingsStore.coverGap" 
            min="8" 
            max="20" 
            step="4"
            class="setting-slider"
          />
        </div>
        <div class="gap-buttons">
          <button 
            :class="['gap-btn', { active: settingsStore.coverGap === 8 }]"
            @click="settingsStore.coverGap = 8"
          >
            紧凑
          </button>
          <button 
            :class="['gap-btn', { active: settingsStore.coverGap === 12 }]"
            @click="settingsStore.coverGap = 12"
          >
            适中
          </button>
          <button 
            :class="['gap-btn', { active: settingsStore.coverGap === 20 }]"
            @click="settingsStore.coverGap = 20"
          >
            宽松
          </button>
        </div>
      </div>

      <div class="setting-item">
        <label class="setting-label">排序</label>
        <div class="sort-buttons">
          <button 
            :class="['sort-btn', { active: settingsStore.sortBy === 'default' }]"
            @click="settingsStore.sortBy = 'default'"
          >
            默认
          </button>
          <button 
            :class="['sort-btn', { active: settingsStore.sortBy === 'title-asc' }]"
            @click="settingsStore.sortBy = 'title-asc'"
          >
            标题 A-Z
          </button>
          <button 
            :class="['sort-btn', { active: settingsStore.sortBy === 'title-desc' }]"
            @click="settingsStore.sortBy = 'title-desc'"
          >
            标题 Z-A
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bookshelf-layout-wrapper {
  width: 100%;
  height: 100%;
  background: var(--sidebar-bg);
  color: var(--text-color);
  display: flex;
  flex-direction: column;
  user-select: none;
}

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

.layout-content {
  flex: 1;
  padding: 24px 20px;
  overflow-y: auto;
}

.setting-item {
  margin-bottom: 26px;
}

.setting-label {
  display: block;
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 10px;
  font-weight: 600;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 14px;
}

.setting-slider {
  width: 100%;
  height: 6px;
  border-radius: 3px;
  background: var(--slider-bg, var(--border-color));
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

.columns-quick-select {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 16px;
}

.col-btn {
  padding: 10px 8px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-secondary);
  font-size: 0.82rem;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-weight: 500;
}

.col-btn:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.col-btn.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  border-color: var(--primary-color);
  box-shadow: var(--shadow-md);
}

.gap-buttons {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 16px;
}

.gap-btn {
  padding: 10px 8px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-secondary);
  font-size: 0.82rem;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-weight: 500;
}

.gap-btn:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.gap-btn.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  border-color: var(--primary-color);
  box-shadow: var(--shadow-md);
}

.sort-buttons {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 16px;
}

.sort-btn {
  padding: 10px 8px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-secondary);
  font-size: 0.82rem;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-weight: 500;
}

.sort-btn:hover {
  background: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.sort-btn.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  border-color: var(--primary-color);
  box-shadow: var(--shadow-md);
}
</style>