<script setup lang="ts">
import iconFactory from "@/utils/iconFactory";

const props = defineProps<{ level: string; size?: number }>();

const details = computed(() => {
  return iconFactory(props.level);
});
</script>

<template>
  <n-button quaternary circle :size="size ? 'tiny' : 'medium'">
    <n-icon :size="size ?? 18">
      <Mdi :path="details.icon" :color="details.color" />
    </n-icon>
  </n-button>
</template>
