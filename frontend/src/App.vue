<script setup lang="ts">
import MainView from './views/MainView.vue'
</script>

<template>
  <MainView />
</template>

<style>
body, html {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  background-color: var(--bg-color);
  overflow: hidden;
}
*, *::before, *::after {
  box-sizing: inherit;
}
</style>