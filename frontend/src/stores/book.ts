import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface TocItem {
  id: string;
  label: string;
  href: string;
  level: number;
  parentId?: string;
  hasChildren?: boolean;
}

export const useBookStore = defineStore('book', () => {
  const activeBookPath = ref('');
  const toc = ref<TocItem[]>([]);

  const setActiveBook = (path: string, tocItems: any[]) => {
    activeBookPath.value = path;
    toc.value = tocItems;
  };

  const clearActiveBook = () => {
    activeBookPath.value = '';
    toc.value = [];
  };

  return {
    activeBookPath,
    toc,
    setActiveBook,
    clearActiveBook
  };
}, {
  persist: true
});