<script setup lang="ts">
defineProps<{ path: string; color?: string }>();
</script>

<template>
  <svg viewBox="0 0 24 24">
    <path :fill="color || 'currentColor'" :d="path" />
  </svg>
</template>
