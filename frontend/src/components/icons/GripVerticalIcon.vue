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
    <path d="M7 20h2M7 16h2M7 12h2M7 8h2M7 4h2M15 20h2M15 16h2M15 12h2M15 8h2M15 4h2"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>