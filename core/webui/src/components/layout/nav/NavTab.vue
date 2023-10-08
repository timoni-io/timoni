<script setup lang="ts">
defineProps<{ name: string; disabled?: boolean }>();
let activeTab = inject("tab") as string;
</script>

<template>
  <button :class="{ active: activeTab === name }" :disabled="disabled">
    <slot />
  </button>
</template>

<style scoped>
button {
  border: none;
  display: block;
  padding: 0.5rem 0.8rem;
  font-size: 0.9rem;
  background-color: var(--accentColor);
  color: black;
  height: 2.5rem;
  transition: all 0.2s;
  cursor: pointer;
}

button:disabled {
  color: #00000033;
  cursor: not-allowed;
}

button:hover {
  background-color: var(--accentColorDarken);
}

button.active {
  background-color: transparent;

  color: white;
}
</style>
