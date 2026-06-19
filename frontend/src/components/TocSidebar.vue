<script setup lang="ts">
import { ref } from 'vue';
import { useBookStore, type TocItem } from '../stores/book';
import ListIcon from './icons/ListIcon.vue';

const props = defineProps<{ currentHref: string }>();
const emit = defineEmits<{
  (e: 'jump', href: string): void
}>();

const bookStore = useBookStore();

// 维护一个已展开节点的 ID 集合
const expandedNodes = ref<Set<string>>(new Set());

// 切换展开状态
const toggleExpand = (item: TocItem) => {
  if (expandedNodes.value.has(item.id)) {
    expandedNodes.value.delete(item.id);
  } else {
    expandedNodes.value.add(item.id);
  }
};

// 判断是否应该显示该条目
const isVisible = (item: TocItem) => {
  // 如果是顶级（level 1），永远可见
  if (item.level === 1) return true;
  // 如果父级 ID 存在于 expandedNodes 中，则可见
  return item.parentId && expandedNodes.value.has(item.parentId);
};

// 处理目录项点击
const handleItemClick = (item: TocItem) => {
  // 如果有子节点，先切换展开状态
  if (item.hasChildren) {
    toggleExpand(item);
  }
  // 跳转
  emit('jump', item.href);
};
</script>

<template>
  <div class="toc-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><ListIcon :size="22" style="margin-right: 8px;" />目录</h2>
    </div>
    
    <div class="toc-content">
      <div v-if="bookStore.toc.length === 0" class="empty-state">
        <span>暂无目录</span>
      </div>
      
      <ul v-else class="toc-list">
        <li 
          v-for="item in bookStore.toc" 
          :key="item.id"
          v-show="isVisible(item)"
          :class="[
            'toc-item', 
            `level-${item.level}`, 
            { 'is-active': item.href === props.currentHref },
            { 'has-children': item.hasChildren }
          ]"
          @click="handleItemClick(item)"
        >
          <span 
            v-if="item.hasChildren" 
            class="toggle-icon"
          >
            {{ expandedNodes.has(item.id) ? '▼' : '▶' }}
          </span>
          <span v-else class="toggle-placeholder"></span>
          
          <span class="toc-title">{{ item.label }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
/* 目录容器 */
.toc-wrapper {
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
  position: relative;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  display: flex;
  align-items: center;
}

/* 目录内容区 */
.toc-content {
  flex: 1;
  padding: 16px 0;
  overflow-y: auto;
}

/* 目录列表 */
.toc-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

/* 目录项 */
.toc-item {
  position: relative;
  padding: 12px 16px;
  margin: 4px 14px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  text-align: left;
  background: transparent;
}

.toc-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 2px;
  height: 0;
  background: var(--primary-color);
  border-radius: 0 2px 2px 0;
  transition: height var(--transition-normal);
}

.toc-item:hover::before {
  height: 60%;
}

.toc-item:hover {
  background: var(--primary-light);
  transform: translateX(6px);
}

.toc-item.is-active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.18) 0%, rgba(99, 102, 241, 0.08) 100%);
  color: var(--primary-color);
  font-weight: 500;
  box-shadow: var(--shadow-md);
}

.toc-item.is-active::before {
  height: 100%;
}

/* 展开/折叠图标 */
.toggle-icon {
  width: 20px;
  font-size: 0.55rem;
  opacity: 0.4;
  margin-right: 6px;
  text-align: center;
  transition: all var(--transition-fast);
}

.toc-item:hover .toggle-icon {
  opacity: 0.7;
}

.toc-item.has-children:hover .toggle-icon {
  transform: scale(1.25);
}

.toggle-placeholder {
  width: 24px;
}

/* 目录标题 */
.toc-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.88rem;
  line-height: 1.5;
  letter-spacing: -0.01em;
}

/* 层级缩进 */
.toc-item.level-1 {
  padding-left: 16px;
}

.toc-item.level-2 {
  padding-left: 38px;
  opacity: 0.9;
}

.toc-item.level-3 {
  padding-left: 60px;
  opacity: 0.8;
  font-size: 0.84rem;
}

/* 空状态 */
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
  font-size: 0.95rem;
}

/* 滚动条美化 */
.toc-content::-webkit-scrollbar {
  width: 4px;
}

.toc-content::-webkit-scrollbar-track {
  background: transparent;
}

.toc-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.toc-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>