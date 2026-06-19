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
    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
    <path d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>
