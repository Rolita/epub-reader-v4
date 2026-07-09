<script setup lang="ts">
import { ref } from 'vue';
import TranslateIcon from './icons/TranslateIcon.vue';

const translateServices = [
  { value: 'none', label: '不使用翻译' },
  { value: 'baidu', label: '百度翻译' },
  { value: 'google', label: 'Google 翻译' },
  { value: 'deepL', label: 'DeepL 翻译' },
];

const languages = [
  { value: 'auto', label: '自动检测' },
  { value: 'zh', label: '中文' },
  { value: 'en', label: 'English' },
  { value: 'ja', label: '日本語' },
  { value: 'ko', label: '한국어' },
  { value: 'fr', label: 'Français' },
  { value: 'de', label: 'Deutsch' },
  { value: 'es', label: 'Español' },
];

const selectedService = ref('none');
const apiKey = ref('');
const sourceLang = ref('auto');
const targetLang = ref('zh');
const enableTranslate = ref(false);

const handleSave = () => {
};
</script>

<template>
  <div class="translate-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><TranslateIcon :size="22" style="margin-right: 8px;" />翻译设置</h2>
    </div>
    
    <div class="translate-content">
      <div class="setting-item">
        <label class="setting-label">启用翻译</label>
        <button 
          :class="['toggle-btn', { active: enableTranslate }]"
          @click="enableTranslate = !enableTranslate"
        >
          <span class="toggle-slider"></span>
        </button>
      </div>

      <div class="setting-item">
        <label class="setting-label">翻译服务</label>
        <select v-model="selectedService" class="setting-select" :disabled="!enableTranslate">
          <option v-for="service in translateServices" :key="service.value" :value="service.value">
            {{ service.label }}
          </option>
        </select>
      </div>

      <div class="setting-item" v-if="selectedService !== 'none'">
        <label class="setting-label">API Key</label>
        <input 
          v-model="apiKey" 
          type="password" 
          class="setting-input"
          placeholder="请输入 API Key"
          :disabled="!enableTranslate"
        />
      </div>

      <div class="setting-item">
        <label class="setting-label">源语言</label>
        <select v-model="sourceLang" class="setting-select" :disabled="!enableTranslate">
          <option v-for="lang in languages" :key="lang.value" :value="lang.value">
            {{ lang.label }}
          </option>
        </select>
      </div>

      <div class="setting-item">
        <label class="setting-label">目标语言</label>
        <select v-model="targetLang" class="setting-select" :disabled="!enableTranslate">
          <option v-for="lang in languages" :key="lang.value" :value="lang.value">
            {{ lang.label }}
          </option>
        </select>
      </div>

      <div class="setting-item" v-if="selectedService !== 'none' && selectedService !== 'google'">
        <label class="setting-label">App ID</label>
        <input 
          v-model="apiKey" 
          type="text" 
          class="setting-input"
          placeholder="请输入 App ID"
          :disabled="!enableTranslate"
        />
      </div>

      <div class="setting-item">
        <label class="setting-label">翻译显示位置</label>
        <div class="align-buttons">
          <button 
            :class="['align-btn', { active: true }]"
            @click=""
            :disabled="!enableTranslate"
          >
            弹窗
          </button>
          <button 
            :class="['align-btn']"
            @click=""
            :disabled="!enableTranslate"
          >
            侧边栏
          </button>
          <button 
            :class="['align-btn']"
            @click=""
            :disabled="!enableTranslate"
          >
            悬停
          </button>
        </div>
      </div>

      <div class="setting-item">
        <label class="setting-label">翻译热键</label>
        <input 
          type="text" 
          class="setting-input"
          value="Ctrl + T"
          readonly
          placeholder="点击设置"
        />
      </div>

      <div class="save-section">
        <button class="save-btn" @click="handleSave" :disabled="!enableTranslate">
          保存设置
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.translate-wrapper {
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

.translate-content {
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

.setting-select:hover:not(:disabled) {
  border-color: var(--primary-color);
}

.setting-select:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.setting-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.setting-input {
  width: 100%;
  padding: 12px 14px;
  border: 1.5px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 0.9rem;
  transition: all var(--transition-fast);
  box-sizing: border-box;
}

.setting-input:hover:not(:disabled) {
  border-color: var(--primary-color);
}

.setting-input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
  outline: none;
}

.setting-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.setting-input::placeholder {
  color: var(--text-muted);
}

.toggle-btn {
  width: 50px;
  height: 28px;
  border-radius: 14px;
  background: var(--border-color);
  border: none;
  cursor: pointer;
  position: relative;
  transition: background var(--transition-fast);
}

.toggle-btn.active {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
}

.toggle-slider {
  position: absolute;
  top: 4px;
  left: 4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: white;
  transition: transform var(--transition-fast);
  box-shadow: var(--shadow-sm);
}

.toggle-btn.active .toggle-slider {
  transform: translateX(22px);
}

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

.align-btn:hover:not(:disabled) {
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

.align-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.save-section {
  margin-top: 32px;
  padding-top: 20px;
  border-top: 1px solid var(--border-color);
}

.save-btn {
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
  color: white;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
  box-shadow: var(--shadow-md);
}

.save-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.save-btn:active:not(:disabled) {
  transform: translateY(0);
}

.save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.translate-content::-webkit-scrollbar {
  width: 4px;
}

.translate-content::-webkit-scrollbar-track {
  background: transparent;
}

.translate-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.translate-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>