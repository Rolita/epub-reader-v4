import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useSettingsStore = defineStore('settings', () => {
  const fontSize = ref(16);            // 字号 (px)
  const fontFamily = ref('Microsoft YaHei'); // 字体
  const paragraphGap = ref(10);        // 段间距 (px)
  const lineHeight = ref(1.5);         // 行间距
  const letterSpacing = ref(0);        // 字间距 (px)
  const indent = ref(2);               // 首行缩进 (em)
  const textAlign = ref('justify');     // 两端对齐
  const sidebarWidth = ref(240);       // 侧边栏宽度 (px)

  return {
    fontSize,
    fontFamily,
    paragraphGap,
    lineHeight,
    letterSpacing,
    indent,
    textAlign,
    sidebarWidth
  };
}, {
  persist: true
});