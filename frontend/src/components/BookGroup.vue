<script setup lang="ts">
import { computed } from 'vue'
import { useLibraryStore, type Group } from '../stores/library'
import BookIcon from './icons/BookIcon.vue'
import FolderIcon from './icons/FolderIcon.vue'

interface Props {
  group: Group
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'click', group: Group): void
  (e: 'contextmenu', event: MouseEvent, group: Group): void
}>()

const store = useLibraryStore()

// 获取分组内的书籍
const groupBooks = computed(() => {
  return store.currentBooks.filter(b => b.groupId === props.group.id)
})

// 获取分组内书籍的封面（最多4个，用于2x2显示）
const groupCovers = computed(() => {
  const covers = groupBooks.value
    .filter(book => book.coverUrl)
    .slice(0, 4)
    .map(book => book.coverUrl)
  
  return covers
})

// 获取分组内的书籍数量
const bookCount = computed(() => {
  return groupBooks.value.length
})

const handleClick = () => {
  emit('click', props.group)
}

const handleContextMenu = (event: MouseEvent) => {
  event.preventDefault()
  event.stopPropagation()
  emit('contextmenu', event, props.group)
}
</script>

<template>
  <div 
    class="group-card" 
    @click="handleClick"
    @contextmenu="handleContextMenu"
  >
    <div class="cover-wrapper">
      <div class="cover">
        <!-- 2x2 封面网格 -->
        <div v-if="groupCovers.length > 0" class="covers-grid">
          <div 
            v-for="(cover, index) in groupCovers" 
            :key="index"
            class="cover-item"
          >
            <img 
              :src="cover" 
              :alt="`封面 ${index + 1}`" 
              class="cover-thumbnail"
            />
          </div>
        </div>
        
        <!-- 没有封面时的显示 -->
        <div v-else class="no-covers">
          <FolderIcon :size="40" />
        </div>
        

      </div>
    </div>
    <div class="title">{{ group.name }}</div>
  </div>
</template>

<style scoped>
.group-card {
  display: flex;
  flex-direction: column;
  cursor: pointer;
  position: relative;
  border-radius: var(--radius-xl);
}

.cover-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 2/3;
  margin-bottom: 10px;
  border-radius: 10px;
}

.cover-wrapper::before {
  content: '';
  position: absolute;
  inset: -5px;
  border-radius: 14px;
  background: conic-gradient(
    from var(--angle, 0deg),
    transparent 0deg,
    var(--primary-color) 60deg,
    var(--accent-color) 120deg,
    transparent 180deg,
    transparent 360deg
  );
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 0;
  pointer-events: none;
  filter: blur(4px);
}

.group-card:hover .cover-wrapper::before {
  opacity: 1;
  animation: rotateBorder 2s linear infinite;
}

@keyframes rotateBorder {
  from { --angle: 0deg; }
  to   { --angle: 360deg; }
}

.cover {
  position: absolute;
  inset: 0;
  background: var(--bg-color);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 2x2 封面网格 */
.covers-grid {
  width: 100%;
  height: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 4px;
  padding: 6px;
}


.cover-item {
  border-radius:2px;  /*分组封面书籍圆角*/
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.cover-thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-covers {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.4;
}



.title {
  font-size: 0.82rem;
  color: var(--text-secondary);
  text-align: center;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-weight: 400;
  letter-spacing: 0;
  padding: 0 2px;
  line-height: 1.4;
}
</style>
