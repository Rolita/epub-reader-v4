import { defineStore } from 'pinia';
import { ref, computed } from 'vue';

const defaultThemes = [
  // --- 您之前的主题 ---
  {
    id: 'dawn-light',
    name: '清晨阳光',
    icon: 'sun',
    bg: '#F0F4F8',
    text: '#2C3E50',
    sidebar: '#E8EEF4',
    functionBar: '#FFFFFF',
    border: '#DDE6ED',
    primary: '#3199df',
    accent: '#47aaec',
  },
  {
    id: 'ink-black',
    name: '墨韵龙鳞',
    icon: 'moon',
    bg: '#161616ff',
    text: '#E5E5E5',
    sidebar: '#161616ff',
    functionBar: '#0D0D0D',
    border: '#1A1A1A',
    primary: '#1E4D4B',
    accent: '#6ee28bff'
  },

  // --- 专业阅读器系列 ---

  {
    id: 'parchment',
    name: '复古羊皮',
    icon: 'cloud',
    bg: '#F4ECD8',
    text: '#5B4636',
    sidebar: '#E8DEC0',
    functionBar: '#F4ECD8',
    border: '#D8CDB3',
    primary: '#8C6D46',
    accent: '#A68B6A'
  },
  {
    id: 'matcha',
    name: '护眼抹茶',
    icon: 'cloud',
    bg: '#F0F4E8',
    text: '#3D4F3D',
    sidebar: '#E2E8D5',
    functionBar: '#F0F4E8',
    border: '#D1D9C2',
    primary: '#4E6B4E',
    accent: '#6B8E6B'
  },
  {
    id: 'nordic',
    name: '北欧极简',
    icon: 'cloud',
    bg: '#E5E9F0',
    text: '#2E3440',
    sidebar: '#D8DEE9',
    functionBar: '#E5E9F0',
    border: '#CCD3DE',
    primary: '#5E81AC',
    accent: '#81A1C1'
  },
  {
    id: 'soft-gray-pro',
    name: '象牙柔灰',
    icon: 'cloud',
    bg: '#F0F1F2',
    text: '#212529',
    sidebar: '#F0F1F2',
    functionBar: '#F8F9FA',
    border: '#DDE6ED',
    primary: '#4B5563',
    accent: '#6366F1'
  }
];

export const useThemeStore = defineStore('theme', () => {
  const currentTheme = ref('kindle-white');
  
  // 存储主题ID的顺序（用于持久化）
  const themeOrder = ref<string[]>(defaultThemes.map(t => t.id));
  
  // 存储自定义主题
  const customThemes = ref<any[]>([]);
  
  // 所有主题（默认 + 自定义）
  const allThemes = computed(() => {
    return [...defaultThemes, ...customThemes.value];
  });
  
  // 根据顺序动态排序的主题列表
  const themes = computed(() => {
    return themeOrder.value
      .map(id => allThemes.value.find(t => t.id === id))
      .filter((t): t is any => !!t); // 只保留真实存在的主题，过滤掉过时的 ID
  });
  
  const themeColors = computed(() => {
    const theme = themes.value.find(t => t.id === currentTheme.value);
    return theme || themes.value[0];
  });
  
  const setTheme = (themeId: string) => {
    currentTheme.value = themeId;
    const theme = themes.value.find(t => t.id === themeId);
    if (theme) {
      document.documentElement.style.setProperty('--bg-color', theme.bg);
      document.documentElement.style.setProperty('--text-color', theme.text);
      document.documentElement.style.setProperty('--sidebar-bg', theme.sidebar);
      document.documentElement.style.setProperty('--function-bar-bg', theme.functionBar);
      document.documentElement.style.setProperty('--border-color', theme.border);
      document.documentElement.style.setProperty('--primary-color', theme.primary);
      document.documentElement.style.setProperty('--accent-color', theme.accent);
      
      // 根据主题设置文字层级颜色
      document.documentElement.style.setProperty('--text-primary', theme.text);
      
      // 判断是否为深色主题（通过背景色亮度判断）
      const bgColor = theme.bg;
      const r = parseInt(bgColor.slice(1, 3), 16);
      const g = parseInt(bgColor.slice(3, 5), 16);
      const b = parseInt(bgColor.slice(5, 7), 16);
      const brightness = (r * 299 + g * 587 + b * 114) / 1000;
      
      if (brightness < 128) {
        // 深色主题：文字颜色应该更亮
        document.documentElement.style.setProperty('--text-secondary', '#9CA3AF');
        document.documentElement.style.setProperty('--text-muted', '#6B7280');
        document.documentElement.style.setProperty('--slider-bg', '#374151');
      } else {
        // 浅色主题：文字颜色应该更深
        document.documentElement.style.setProperty('--text-secondary', '#6B7280');
        document.documentElement.style.setProperty('--text-muted', '#9CA3AF');
        document.documentElement.style.setProperty('--slider-bg', '#DDE6ED');
      }
    }
  };
  
  const initTheme = () => {
    // 获取当前代码中定义的所有有效 ID
    const allIds = new Set(allThemes.value.map(t => t.id));
    
    // 1. 过滤掉 themeOrder 中已经不存在的 ID（清理残留）
    const validOrder = themeOrder.value.filter(id => allIds.has(id));
    
    // 2. 检查是否有新增的默认主题不在 validOrder 中
    const missingIds = allThemes.value
      .map(t => t.id)
      .filter(id => !validOrder.includes(id));
    
    // 如果顺序发生了变化（有残留或有新增），则更新顺序
    if (validOrder.length !== themeOrder.value.length || missingIds.length > 0) {
      themeOrder.value = [...validOrder, ...missingIds];
    }
    
    // 3. 确保当前主题是有效的，否则回退到第一个
    if (!allIds.has(currentTheme.value)) {
      currentTheme.value = defaultThemes[0].id;
    }
    
    setTheme(currentTheme.value);
  };
  
  // 重新排序主题列表（会保存顺序）
  const reorderThemes = (newThemes: typeof defaultThemes) => {
    themeOrder.value = newThemes.map(t => t.id);
  };
  
  // 添加自定义主题
  const addTheme = (theme: any) => {
    customThemes.value.push(theme);
    themeOrder.value.push(theme.id);
    setTheme(theme.id);
  };
  
  // 更新自定义主题
  const updateTheme = (themeId: string, updatedData: any) => {
    const idx = customThemes.value.findIndex(t => t.id === themeId);
    if (idx !== -1) {
      // 创建新的数组以触发响应式更新
      const newCustomThemes = [...customThemes.value];
      newCustomThemes[idx] = { ...newCustomThemes[idx], ...updatedData };
      customThemes.value = newCustomThemes;
      
      // 如果正在使用该主题，重新应用
      if (currentTheme.value === themeId) {
        setTheme(themeId);
      }
    }
  };
  
  // 从默认主题创建自定义副本（用于编辑默认主题）
  const duplicateDefaultTheme = (themeId: string): string | null => {
    const defaultTheme = defaultThemes.find(t => t.id === themeId);
    if (!defaultTheme) return null;
    
    // 生成新的 ID 和名称
    const newId = `custom-${Date.now()}`;
    const newName = `${defaultTheme.name}（副本）`;
    
    // 创建副本并添加到自定义主题
    const newTheme = {
      id: newId,
      name: newName,
      icon: defaultTheme.icon,
      primary: defaultTheme.primary,
      accent: defaultTheme.accent,
      bg: defaultTheme.bg,
      text: defaultTheme.text,
      sidebar: defaultTheme.sidebar,
      functionBar: defaultTheme.functionBar,
      border: defaultTheme.border
    };
    
    customThemes.value.push(newTheme);
    themeOrder.value.push(newId);
    
    return newId;
  };
  
  // 判断是否为默认主题
  const isDefaultTheme = (themeId: string) => {
    return defaultThemes.some(t => t.id === themeId);
  };
  
  // 删除自定义主题
  const deleteTheme = (themeId: string) => {
    customThemes.value = customThemes.value.filter(t => t.id !== themeId);
    themeOrder.value = themeOrder.value.filter(id => id !== themeId);
    // 如果删除的是当前使用的主题，切换到第一个
    if (currentTheme.value === themeId) {
      setTheme(themes.value[0]?.id || 'ivory');
    }
  };
  
  return {
    currentTheme,
    themes,
    themeColors,
    themeOrder,
    customThemes,
    setTheme,
    initTheme,
    reorderThemes,
    addTheme,
    updateTheme,
    duplicateDefaultTheme,
    isDefaultTheme,
    deleteTheme
  };
}, {
  persist: true
});