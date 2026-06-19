<template>
  <svg 
    :width="size || 20" 
    :height="size || 20" 
    viewBox="0 0 24 24" 
    fill="none" 
    stroke="currentColor" 
    stroke-width="2" 
    stroke-linecap="round" 
    stroke-linejoin="round"
    :class="className"
  >
    <path d="M17 7a4 4 0 0 0-8 0c0 1.5 2 3 2 5s-1.5 3.5-2 5a4 4 0 0 0 8 0c0-1.5-2-3-2-5s1.5-3.5 2-5"/>
    <path d="M7 7a4 4 0 0 1 8 0c0 1.5-2 3-2 5s1.5 3.5 2 5a4 4 0 0 1-8 0c0-1.5 2-3 2-5s-1.5-3.5-2-5"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>
