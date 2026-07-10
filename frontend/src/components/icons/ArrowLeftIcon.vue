<script setup lang="ts">
interface Props {
  size?: number
  color?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 16,
  color: 'currentColor'
})
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <path d="M19 12H5"/>
    <path d="m12 19-7-7 7-7"/>
  </svg>
</template>
