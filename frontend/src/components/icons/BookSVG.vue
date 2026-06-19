<template>
  <svg 
    :width="size || 20" 
    :height="size || 20" 
    viewBox="0 0 24 24" 
    xmlns="http://www.w3.org/2000/svg"
    :class="className"
    fill="none"
    stroke="currentColor"
    stroke-width="1.5"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <path d="M5 2h14a2 2 0 0 1 2 2v16a2 2 0 0 1-2 2H5V2z"/>
    
    <path d="M7 2v20"/>
    
    <path d="M11 7h6M11 10h4M11 13h5"/>
    
    <path d="M8 2v4l1-0.8 1 0.8V2"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>