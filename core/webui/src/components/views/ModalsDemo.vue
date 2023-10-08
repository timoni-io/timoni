<script setup lang="ts">
let show = $ref(false);

// let modalInput = $ref("");
let popModalInput = $ref("");
</script>

<template>
  <PageLayout>
    <n-button @click="show = !show">Show Modal</n-button>
    <!-- <Modal v-model:show="show" title="dupa" :touched="!!modalInput">
      Dupa <n-input v-model:value="modalInput" />
    </Modal> -->
    <PopModal
      title="Dupa"
      :touched="!!popModalInput"
      :show-footer="{ positiveText: 'Confirm' }"
    >
      <template #trigger>
        <n-button>Show PopModal</n-button>
      </template>
      <template #content>
        <n-input v-model:value="popModalInput" />
      </template>
    </PopModal>
  </PageLayout>
</template>
