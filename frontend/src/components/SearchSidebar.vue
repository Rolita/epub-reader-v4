<script lang="ts">
export default {
  name: 'SearchSidebar'
};
</script>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import SearchIcon from './icons/SearchIcon.vue';
import XIcon from './icons/XIcon.vue';
import { getSearchHistory, saveSearchHistory, clearSearchHistory, type SearchHistoryItem } from '../composables/useReaderProgress';

const props = defineProps<{
  hasActiveBook: boolean;
  bookTitle?: string;
  filePath?: string;
  searchInBook?: (keyword: string) => Promise<Array<{ chapter: string; snippet: string; href: string; cfi: string; page: number }>>;
  highlightSearchKeyword?: (keyword: string) => void;
  clearSearchHighlight?: () => void;
}>();

interface SearchResult {
  id: string;
  chapter: string;
  snippet: string;
  page: number;
  href: string;
  cfi: string;
}

const searchQuery = ref('');
const searchResults = ref<SearchResult[]>([]);
const isSearching = ref(false);
const searchHistory = ref<SearchHistoryItem[]>([]);

interface GroupedResults {
  chapter: string;
  results: SearchResult[];
}

const groupedResults = computed<GroupedResults[]>(() => {
  const groups: { [key: string]: SearchResult[] } = {};
  
  searchResults.value.forEach((result) => {
    const chapter = result.chapter || '未知章节';
    if (!groups[chapter]) {
      groups[chapter] = [];
    }
    groups[chapter].push(result);
  });
  
  return Object.entries(groups).map(([chapter, results]) => ({
    chapter,
    results
  }));
});

const emit = defineEmits<{
  (e: 'jump', payload: { href: string; cfi: string }): void;
}>();

let searchTimeout: ReturnType<typeof setTimeout> | null = null;

const loadSearchHistory = async () => {
  if (!props.filePath) return;
  try {
    const result = await getSearchHistory(props.filePath);
    if (result) {
      searchHistory.value = result;
    }
  } catch (error) {
    console.error('加载搜索历史失败:', error);
  }
};

const handleSaveSearchHistory = async (keyword: string) => {
  if (!props.filePath) return;
  try {
    await saveSearchHistory(props.filePath, keyword);
    await loadSearchHistory();
  } catch (error) {
    console.error('保存搜索历史失败:', error);
  }
};

const handleClearSearchHistory = async () => {
  if (!props.filePath) return;
  try {
    await clearSearchHistory(props.filePath);
    searchHistory.value = [];
  } catch (error) {
    console.error('清除搜索历史失败:', error);
  }
};

const handleSearch = async () => {
  if (!props.hasActiveBook || !props.searchInBook || !searchQuery.value.trim()) {
    searchResults.value = [];
    props.clearSearchHighlight?.();
    return;
  }
  
  isSearching.value = true;
  try {
    const results = await props.searchInBook(searchQuery.value.trim());
    searchResults.value = results.map((r, index) => ({
      id: `result-${index}`,
      chapter: r.chapter,
      snippet: r.snippet,
      page: r.page,
      href: r.href,
      cfi: r.cfi
    }));
    
    if (results.length > 0) {
      props.highlightSearchKeyword?.(searchQuery.value.trim());
      await handleSaveSearchHistory(searchQuery.value.trim());
    }
  } catch (error) {
    console.error('搜索失败:', error);
    searchResults.value = [];
  } finally {
    isSearching.value = false;
  }
};

onMounted(() => {
  loadSearchHistory();
});

const handleInput = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  searchTimeout = setTimeout(handleSearch, 300);
};

const handleClear = () => {
  searchQuery.value = '';
  searchResults.value = [];
  props.clearSearchHighlight?.();
};

const handleResultClick = (result: SearchResult) => {
  emit('jump', { href: result.href, cfi: result.cfi });
  setTimeout(() => {
    emit('jump', { href: result.href, cfi: result.cfi });
  }, 50);
};

const handleHistoryClick = (keyword: string) => {
  searchQuery.value = keyword;
  handleSearch();
};

const formatTime = (timestamp: number) => {
  const now = Date.now();
  const diff = now - timestamp;
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);
  
  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  if (hours < 24) return `${hours}小时前`;
  if (days < 7) return `${days}天前`;
  return new Date(timestamp).toLocaleDateString();
};
</script>

<template>
  <div class="search-wrapper">
    <div class="sidebar-header" style="--wails-draggable: drag;">
      <h2><SearchIcon :size="22" style="margin-right: 8px;" />搜索</h2>
    </div>
    
    <div class="search-content">
      <div class="search-input-wrapper">
        <SearchIcon :size="16" class="search-input-icon" />
        <input 
          v-model="searchQuery"
          type="text" 
          class="search-input"
          :placeholder="hasActiveBook ? '搜索书籍内容...' : '请先打开书籍'"
          :disabled="!hasActiveBook"
          @input="handleInput"
          @keyup.enter="handleSearch"
        />
        <button v-if="searchQuery" class="search-clear-btn" @click="handleClear">
          <XIcon :size="14" />
        </button>
      </div>

      <div v-if="hasActiveBook" class="current-book-info">
        <span class="book-name">{{ bookTitle || '未知书籍' }}</span>
      </div>

      <div v-if="!hasActiveBook" class="no-book-state">
        <BookIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
        <span>请先打开书籍</span>
        <p style="font-size: 0.8rem; color: var(--text-muted); margin-top: 8px;">打开书籍后可搜索书中内容</p>
      </div>

      <template v-else>

        <div v-if="isSearching" class="search-loading">
          <span class="loading-dots">搜索中...</span>
        </div>

        <div v-else-if="searchQuery && searchResults.length === 0" class="empty-state">
          <SearchIcon :size="48" style="opacity: 0.3; margin-bottom: 12px;" />
          <span>未找到相关结果</span>
          <p style="font-size: 0.8rem; color: var(--text-muted); margin-top: 8px;">尝试其他关键词</p>
        </div>

        <div v-else-if="searchResults.length > 0" class="search-results">
          <div class="results-header">
            <span>找到 {{ searchResults.length }} 个结果</span>
          </div>
          <div class="grouped-results">
            <div v-for="group in groupedResults" :key="group.chapter" class="chapter-group">
              <div class="chapter-header">{{ group.chapter }}</div>
              <ul class="results-list">
                <li 
                  v-for="result in group.results" 
                  :key="result.id"
                  class="result-item"
                  @click="handleResultClick(result)"
                >
                  <div class="result-content">
                    <span class="result-snippet">{{ result.snippet }}</span>
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div v-else-if="searchHistory.length > 0" class="search-history">
          <div class="history-header">
            <span>搜索历史</span>
            <button class="history-clear-btn" @click="handleClearSearchHistory" title="清除历史">
              <XIcon :size="14" />
            </button>
          </div>
          <ul class="history-list">
            <li 
              v-for="(item, index) in searchHistory" 
              :key="index"
              class="history-item"
              @click="handleHistoryClick(item.keyword)"
            >
              <SearchIcon :size="14" class="history-icon" />
              <span class="history-keyword">{{ item.keyword }}</span>
              <span class="history-time">{{ formatTime(item.timestamp) }}</span>
            </li>
          </ul>
        </div>

        <div v-else class="search-tips">
          <p>输入关键词搜索当前书籍内容</p>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.search-wrapper {
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

.search-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.no-book-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

.search-input-wrapper {
  position: relative;
  margin-bottom: 12px;
}

.search-input {
  width: 100%;
  padding: 10px 36px 10px 32px;
  border: none;
  border-radius: var(--radius-md);
  background: rgba(var(--primary-rgb), 0.06);
  color: var(--text-primary);
  font-size: 0.85rem;
  transition: all var(--transition-fast);
  box-sizing: border-box;
}

.search-input:hover {
  background: rgba(var(--primary-rgb), 0.1);
}

.search-input:focus {
  background: rgba(var(--primary-rgb), 0.12);
  box-shadow: 0 0 0 2px rgba(var(--primary-rgb), 0.2);
  outline: none;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-input:disabled {
  background: var(--border-color);
  cursor: not-allowed;
}

.current-book-info {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 16px;
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  background: var(--border-color);
}

.book-name {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-input-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.search-clear-btn {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.search-clear-btn:hover {
  color: var(--text-secondary);
  background: var(--border-color);
}

.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
}

.loading-dots {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  color: var(--text-secondary);
}

.search-tips {
  padding: 20px 0;
  text-align: center;
}

.search-tips p {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.search-results {
  margin-top: 4px;
}

.results-header {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.results-header::before {
  content: '';
  width: 4px;
  height: 14px;
  background: var(--primary-color);
  border-radius: 2px;
}

.chapter-group {
  margin-bottom: 20px;
}

.chapter-header {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-secondary);
  padding: 8px 0;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.chapter-header::after {
  content: '';
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, var(--border-color) 0%, transparent 100%);
}

.results-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.result-item {
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-bottom: 4px;
  background: transparent;
}

.result-item:hover {
  background: var(--primary-light);
}

.result-item:active {
  transform: scale(0.98);
}

.result-content {
  min-width: 0;
}

.result-snippet {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: 0.8rem;
  color: var(--text-secondary);
  line-height: 1.6;
  padding-left: 12px;
  border-left: 2px solid var(--border-color);
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-history {
  margin-top: 4px;
}

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 12px;
}

.history-clear-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.history-clear-btn:hover {
  color: var(--text-secondary);
  background: var(--border-color);
}

.history-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-bottom: 4px;
}

.history-item:hover {
  background: var(--primary-light);
}

.history-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.history-keyword {
  flex: 1;
  font-size: 0.8rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-time {
  font-size: 0.7rem;
  color: var(--text-muted);
  flex-shrink: 0;
}

.search-content::-webkit-scrollbar {
  width: 4px;
}

.search-content::-webkit-scrollbar-track {
  background: transparent;
}

.search-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 2px;
}

.search-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}
</style>