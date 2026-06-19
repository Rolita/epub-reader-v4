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
    <line x1="5" y1="12" x2="19" y2="12"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>