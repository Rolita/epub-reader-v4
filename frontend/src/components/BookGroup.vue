<script setup lang="ts">
import { computed, ref } from 'vue'
import { useLibraryStore, type Group } from '../stores/library'
import BookIcon from './icons/BookIcon.vue'
import FolderIcon from './icons/FolderIcon.vue'
import DefaultCover from './DefaultCover.vue'

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

// 获取分组内书籍（最多4个，用于2x2显示）
const groupBooksForCover = computed(() => {
  return groupBooks.value
    .slice(0, 4)
})

// 获取分组内的书籍数量
const bookCount = computed(() => {
  return groupBooks.value.length
})

const handleClick = () => {
  emit('click', props.group)
}

const coverFailed = ref<Set<number>>(new Set())

const handleContextMenu = (event: MouseEvent) => {
  event.preventDefault()
  event.stopPropagation()
  emit('contextmenu', event, props.group)
}

const handleCoverError = (event: Event, index: number) => {
  coverFailed.value.add(index)
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
        <div v-if="groupBooksForCover.length > 0" class="covers-grid">
          <div 
            v-for="(book, index) in groupBooksForCover" 
            :key="index"
            class="cover-item"
          >
            <img 
              v-if="book.coverUrl && !coverFailed.has(index)" 
              :src="book.coverUrl" 
              :alt="`封面 ${index + 1}`" 
              class="cover-thumbnail"
              @error="handleCoverError($event, index)"
            />
            <DefaultCover v-else :title="book.title" :author="book.author" :compact="true" />
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
  user-select: none;
  transition: all 0.2s ease;
}

.group-card:active {
  cursor: pointer;
}

.group-card.dragging {
  opacity: 0.3;
}

.group-card.dragging .cover-wrapper::before {
  opacity: 0;
}

.group-card.drag-over .cover {
  border: 2px solid var(--primary-color);
  background: rgba(79, 70, 229, 0.1);
}

.group-card.drag-over {
  transform: scale(1.02);
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
  font-size: 0.87rem;
  color: var(--text-color);
  text-align: center;
  font-weight: 500;
  letter-spacing: 0.01em;
  padding: 0px 8px 0px;
  line-height: 1.45;
  overflow: hidden;
  min-height: calc(0.87rem * 1.45 * 2);

  /* WebKit 内核（Chrome/Edge/Safari） */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;

  /* W3C 标准属性（Firefox 等，消除编辑器警告） */
  display: box;
  line-clamp: 2;
  box-orient: vertical;
}
</style>
