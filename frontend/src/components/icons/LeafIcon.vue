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
    <path d="M11 20A7 7 0 0 1 9.8 6.1C15.5 5 17 4.48 19 2c1 2 2 4.27 1 7-1 3-3 5-4 6-2 2-3.09 2.49-4 3z"/>
    <path d="M7 21c-3-1-4-3.5-2-5 1-1.5 2-2.5 3-4-.1-2 1-4 1-4"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>
