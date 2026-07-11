<template>
  <svg
    :width="size || 24"
    :height="size || 24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <line x1="12" y1="2" x2="12" y2="14"></line>
    <line x1="8" y1="18" x2="16" y2="18"></line>
    <line x1="12" y1="14" x2="12" y2="22"></line>
    <line x1="12" y1="14" x2="8" y2="18"></line>
    <line x1="12" y1="14" x2="16" y2="18"></line>
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: number
}>()
</script>
