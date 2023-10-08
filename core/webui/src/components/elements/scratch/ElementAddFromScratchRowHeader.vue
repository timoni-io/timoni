<script setup lang="ts">
const props = defineProps<{
  name: string;
  tip: string;
}>();
</script>

<template>
  <div class="element-from-scratch-row-header">
    <p>
      {{ props.name }}
    </p>
    <n-tooltip trigger="hover">
      <template #trigger>
        <n-button size="tiny" secondary type="primary" circle>
          <n-icon :size="20">
            <Mdi :path="mdiInformation" />
          </n-icon>
        </n-button>
      </template>
      <div>
        <p>
          {{ props.tip }}
        </p>
      </div>
    </n-tooltip>
  </div>
</template>

<style scoped>
.element-from-scratch-row-header {
  width: 20%;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding-right: 2em;
  text-align: right;
}

.element-from-scratch-row-header > p {
  margin-right: 0.5em;
}
</style>
