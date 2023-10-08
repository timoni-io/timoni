<script setup lang="ts">
const props = defineProps<{
  activeTab?: string;
}>();

provide("tab", props.activeTab);
</script>

<template>
  <!-- <div></div> -->
  <Teleport to="#nav-tabs">
    <slot name="tabs"></slot>
  </Teleport>
  <!--
  <Teleport to="#nav-suffix">
    <slot name="suffix"></slot>
  </Teleport> -->
</template>

<style lang="scss">
.tabs {
  padding: 0;
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 10;
  top: 0;
  left: 0;
  width: 100%;
  margin-bottom: 1rem;
  // background-color: var(--accentColor);

  & .n-tabs-tab__label {
    padding-left: 1rem;
    padding-right: 1rem;
    height: 100%;
  }

  & .n-tabs-tab-wrapper:has(.n-tabs-tab--active) {
    background-color: #191919;
    --n-tab-text-color-active: white;
  }

  .n-tabs-tab-wrapper {
    border-left: 1px solid #ebebeb1a;
    --n-tab-text-color: black;
    --n-tab-text-color-hover: black;
    border-left: 1px solid #31313152;
  }
  .n-tabs-tab-wrapper:hover {
    background-color: var(--accentColorDarken);
  }

  .n-tabs-tab-wrapper:last-child {
    border-right: 1px solid #31313152;
  }

  &__dark {
    background-color: #191919;
  }

  &__light {
    background-color: rgba(245, 245, 245, 0.9);
  }
}
</style>
