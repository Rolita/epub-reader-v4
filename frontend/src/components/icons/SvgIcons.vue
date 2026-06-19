<template>
  <svg 
    :width="size || 20" 
    :height="size || 20" 
    :viewBox="viewBox || '0 0 24 24'" 
    fill="none" 
    stroke="currentColor" 
    stroke-width="2" 
    stroke-linecap="round" 
    stroke-linejoin="round"
    :class="className"
  >
    <slot />
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  viewBox?: string
  className?: string
}>()
</script>
