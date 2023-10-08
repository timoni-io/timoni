<script setup lang="ts">
const props = defineProps<{
  name: string;
  tip: string;
}>();
</script>

<template>
  <div style="display: flex; align-items: center">
    <p style="margin-right: 5px">{{ props.name }}</p>
    <n-tooltip trigger="hover">
      <template #trigger>
        <n-button size="tiny" secondary type="primary" circle>
          <n-icon :size="15">
            <Mdi :path="mdiInformation" />
          </n-icon>
        </n-button>
      </template>
      <div>
        <p>{{ props.tip }}</p>
      </div>
    </n-tooltip>
  </div>
</template>
