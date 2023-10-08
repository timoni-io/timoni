<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

const router = useRouter();
const { t } = useI18n();

const activeTab = computed(() =>
  router.currentRoute.value.path.split("/")[1] === "admin"
    ? ""
    : "/" + router.currentRoute.value.path.split("/")[1]
);

// admin options
let adminOptions = $ref([
  "platform",
  // "operators",
  "containerImages",
  "systemLogs",
  "logsAnalysis",
]);
</script>

<template>
  <n-tabs
    :value="activeTab"
    :key="activeTab"
    animated
    size="small"
    style="height: 2.4rem"
    ><n-tab
      v-for="option in adminOptions"
      :key="option"
      :name="`/${option}`"
      style="height: 2.4rem"
      >{{ t(`adminPanel.${option}`) }}</n-tab
    ></n-tabs
  >
</template>
