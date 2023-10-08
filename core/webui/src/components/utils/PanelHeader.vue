<script setup lang="ts">
defineProps<{ title: string }>();
const slots = useSlots();
</script>

<template>
  <div class="wrapper">
    <div class="header">
      <h4 class="panel-header">{{ title }}</h4>
      <slot name="trailing"></slot>
    </div>
    <div v-if="slots.bottom" class="bottom"><slot name="bottom"></slot></div>
  </div>
</template>

<style scoped>
.wrapper {
  position: sticky;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  /* background-color: var(--cardColor); */
  padding: 0 0.15rem;
  transform: translateY(-0.15rem);
  width: 100%;
}
.header h4 {
  font-weight: normal;
}
.bottom {
  padding: 0.2rem 0.5rem;
  background-color: rgba(31, 31, 36, 1);
}
</style>
