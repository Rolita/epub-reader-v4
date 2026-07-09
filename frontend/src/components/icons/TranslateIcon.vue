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
    <polyline points="12 3 20 7 12 11"/>
    <polyline points="12 13 20 17 12 21"/>
    <line x1="4" y1="7" x2="4" y2="17"/>
    <line x1="12" y1="3" x2="12" y2="5"/>
    <line x1="12" y1="19" x2="12" y2="21"/>
    <line x1="20" y1="7" x2="18" y2="5"/>
    <line x1="20" y1="17" x2="18" y2="19"/>
    <line x1="4" y1="7" x2="6" y2="5"/>
    <line x1="4" y1="17" x2="6" y2="19"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
  className?: string
}>()
</script>