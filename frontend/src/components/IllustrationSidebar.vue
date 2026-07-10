<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import IllustrationIcon from './icons/IllustrationIcon.vue'
import { useSettingsStore, type IllustrationItem } from '../stores/settings'

const settingsStore = useSettingsStore()

const expandedGroups = ref<Set<string>>(new Set())
const isFirstLoad = ref(true)

// 按章节分组
const groupedIllustrations = computed(() => {
  const groups: { chapterTitle: string; chapterHref: string; illustrations: IllustrationItem[] }[] = []
  const map = new Map<string, { chapterTitle: string; chapterHref: string; illustrations: IllustrationItem[] }>()
  
  for (const illustration of settingsStore.illustrations) {
    const key = illustration.chapterHref || 'unknown'
    if (!map.has(key)) {
      map.set(key, {
        chapterTitle: illustration.chapterTitle || '未分类',
        chapterHref: key,
        illustrations: []
      })
      groups.push(map.get(key)!)
    }
    map.get(key)!.illustrations.push(illustration)
  }
  
  return groups
})

// 切换分组展开/收缩
const toggleGroup = (chapterHref: string) => {
  if (expandedGroups.value.has(chapterHref)) {
    expandedGroups.value.delete(chapterHref)
  } else {
    expandedGroups.value.add(chapterHref)
  }
}

// 检查分组是否展开
const isGroupExpanded = (chapterHref: string) => {
  return expandedGroups.value.has(chapterHref)
}

// 监听插画变化，保持展开状态
watch(() => settingsStore.illustrations, (newIllustrations) => {
  if (isFirstLoad.value && newIllustrations.length > 0) {
    // 首次加载到数据时展开所有分组
    const groups = new Set<string>()
    for (const illustration of newIllustrations) {
      groups.add(illustration.chapterHref || 'unknown')
    }
    expandedGroups.value = groups
    isFirstLoad.value = false
  } else if (!isFirstLoad.value) {
    // 后续更新时，保持当前展开状态
    const currentGroups = new Set<string>()
    for (const illustration of newIllustrations) {
      currentGroups.add(illustration.chapterHref || 'unknown')
    }
    
    // 移除不存在的分组，保留已有的分组展开状态
    const newExpanded = new Set<string>()
    expandedGroups.value.forEach(group => {
      if (currentGroups.has(group)) {
        newExpanded.add(group)
      }
    })
    expandedGroups.value = newExpanded
  }
}, { immediate: true })

const emit = defineEmits<{
  jump: [payload: { href: string; cfi: string }]
  preview: [payload: { src: string; alt: string }]
}>()

const handleImageClick = (illustration: IllustrationItem) => {
  if (illustration.chapterHref) {
    const payload = { href: illustration.chapterHref, cfi: illustration.cfi || '' }
    // 单击一次，模拟双击效果（两次单击）
    emit('jump', payload)
    // 延迟一小段时间后再次触发第二次单击
    setTimeout(() => {
      emit('jump', payload)
    }, 50)
    setTimeout(() => {
      emit('jump', payload)
    }, 50)
  }
}

// 右键预览大图
const handleImageContextMenu = (e: MouseEvent, illustration: IllustrationItem) => {
  e.preventDefault()
  emit('preview', { src: illustration.src, alt: illustration.alt })
}
</script>

<template>
  <div class="illustration-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><IllustrationIcon :size="22" style="margin-right: 8px; vertical-align: middle;" />插画</h2>
      <span class="count" v-if="settingsStore.illustrations.length > 0">{{ settingsStore.illustrations.length }} 张</span>
    </div>
    
    <div class="illustration-content">
      <!-- 空状态 -->
      <div v-if="settingsStore.illustrations.length === 0" class="empty-state">
        <IllustrationIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>暂无插画内容</span>
        <p style="font-size: 0.8rem; color: var(--text-muted); margin-top: 8px;">打开书籍后自动收集</p>
      </div>
      
      <!-- 按章节分组的插图 -->
      <div v-else class="illustration-groups">
        <div v-for="group in groupedIllustrations" :key="group.chapterHref" class="illustration-group">
          <div class="group-header" @click="toggleGroup(group.chapterHref)">
            <div class="group-header-left">
              <span class="expand-icon" :class="{ 'expanded': isGroupExpanded(group.chapterHref) }">
                ▶
              </span>
              <span class="group-title">{{ group.chapterTitle }}</span>
            </div>
            <span class="group-count">{{ group.illustrations.length }} 张</span>
          </div>
          <div class="group-list" v-show="isGroupExpanded(group.chapterHref)">
            <div 
              v-for="illustration in group.illustrations" 
              :key="illustration.index"
              class="illustration-item"
              @click="handleImageClick(illustration)"
              @contextmenu.prevent="handleImageContextMenu($event, illustration)"
              :title="`点击跳转，右键预览`"
            >
              <img :src="illustration.src" :alt="illustration.alt" loading="lazy" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.illustration-wrapper {
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
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  display: flex;
  align-items: center;
}

.count {
  font-size: 0.85rem;
  color: var(--text-muted);
  background: var(--bg-secondary);
  padding: 4px 10px;
  border-radius: 12px;
}

.illustration-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

/* 章节分组 */
.illustration-groups {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.illustration-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  cursor: pointer;
}

.group-header-left {
  display: flex;
  align-items: center;
  gap: 6px;
}

.expand-icon {
  font-size: 0.6rem;
  color: var(--text-muted);
  transition: transform 0.2s ease;
  transform: rotate(0deg);
}

.expand-icon.expanded {
  transform: rotate(90deg);
}

.group-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.group-count {
  font-size: 0.8rem;
  color: var(--text-muted);
  background: var(--bg-secondary);
  padding: 2px 8px;
  border-radius: 10px;
  margin-left: 8px;
}

/* 插图列表 */
.group-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.illustration-item {
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
  background: var(--bg-secondary);
  border: 2px solid transparent;
  transition: all 0.2s ease;
}

.illustration-item:hover {
  border-color: var(--primary-color);
}

.illustration-item img {
  display: block;
  width: 100%;
  max-height: 312px;
  object-fit: contain;
}
</style>
