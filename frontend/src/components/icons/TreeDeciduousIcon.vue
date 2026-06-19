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
    <path d="M12 22c2-3 2.5-5.5 2.5-7.5A2.5 2.5 0 0 0 12 7a2.5 2.5 0 0 0-2.5 7c0 2 .5 4.5 2.5 7.5z"/>
    <path d="M12 22V11"/>
    <path d="M8 11h8"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>
