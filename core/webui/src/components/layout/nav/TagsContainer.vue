<script setup lang="ts">
// fav filters list
let favFilters = $ref<{ [key: string]: { active: boolean } }>({
  Filtr1: { active: true },
  "Filtr nr 2": { active: false },
  "kolejny filtr": { active: true },
});

// filters actions
const toggleFilter = (name: unknown) => {
  if (typeof name === "string" && Object.keys(favFilters).includes(name))
    favFilters[name].active = !favFilters[name].active;
};

const removeFav = (name: unknown) => {
  if (typeof name === "string") delete favFilters[name];
};
</script>

<template>
  <div class="fav-tags-container">
    <n-tag
      v-for="(active, name) in favFilters"
      :key="name"
      type="primary"
      :color="(active as any).active ? { color: '#63e2b72b' } : undefined"
      @click="toggleFilter(name)"
      style="cursor: pointer"
    >
      <div style="display: flex; align-items: center; gap: 0.2rem">
        {{ name }}
        <p @click.prevent="removeFav(name)" class="star-icon">
          <n-icon class="star"><mdi :path="mdiStar" /></n-icon>
          <n-icon class="star-outline"><mdi :path="mdiStarOutline" /></n-icon>
        </p>
      </div>
    </n-tag>
  </div>
</template>

<style scoped>
.fav-tags-container {
  display: flex;
  gap: 0.5rem;
}
.star-icon:hover .star {
  display: none;
}
.star-outline {
  display: none;
  transform: translateY(-1px);
}
.star-icon:hover .star-outline {
  display: block;
}
</style>
