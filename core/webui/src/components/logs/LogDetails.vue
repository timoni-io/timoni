<script setup lang="ts">
import iconFactory from "@/utils/iconFactory";
import { useTimeFormatter } from "@/utils/formatTime";
import chroma from "chroma-js";
// import { useLogsStore } from "@/store/logsStore";
import { useRoute } from "vue-router";
import { useApi } from "@/next-api";
// import { invoke } from "@vueuse/core";
// import { useRoute } from "vue-router";
import { format } from "date-fns";

const route = useRoute();
const api = useApi();

let logsDetails = $ref<ReturnType<typeof api["getLogDetails"]> | null>(null);
const props = defineProps<{
  log: any;
}>();

const { relativeOrDistanceToNow } = useTimeFormatter();
watchEffect(async () => {
  logsDetails = await api.getLogDetails({
    envId:
      props.log.element === "image-builder"
        ? "image-builder"
        : props.log.env_id
        ? props.log.env_id
        : route.params.id,
    time: props.log.time,
  });
});

const details = computed(() => {
  return iconFactory(props.log.level);
});
</script>

<template>
  <div style="display: flex; align-items: center">
    <n-tag
      round
      :bordered="false"
      :color="{
        color: chroma(details.color).darken().hex(),
      }"
    >
      <template #icon>
        <n-icon size="18" style="margin: 0">
          <Mdi :path="details.icon" :color="details.color" />
        </n-icon>
      </template>
      {{ log.level }}
    </n-tag>
    <n-divider class="log-details-divider" vertical />
    <div style="display: flex; flex-flow: column">
      <n-tooltip>
        <template #trigger>
          <n-tag round :bordered="false" type="info">
            {{ relativeOrDistanceToNow(new Date(log.time / 1000000)) }}
          </n-tag>
        </template>
        {{ format(new Date(log.time / 1000000), "hh:mm:ss.SSS a") }}
      </n-tooltip>
    </div>

    <n-divider
      class="log-details-divider"
      vertical
      style="margin-right: 20px"
    />
    <div>
      <h5>project</h5>
      <p>{{ log.project }}</p>
    </div>
    <n-icon class="log-details-arrow" size="24">
      <Mdi :path="mdiChevronRight" color="var(--infoColorHover)" />
    </n-icon>
    <div>
      <h5>env</h5>
      <p>{{ log.env_id }}</p>
    </div>
    <n-icon class="log-details-arrow" size="24">
      <Mdi :path="mdiChevronRight" color="var(--infoColorHover)" />
    </n-icon>
    <div>
      <h5>version</h5>
      <p>{{ log.version }}</p>
    </div>
    <n-icon class="log-details-arrow" size="24">
      <Mdi :path="mdiChevronRight" color="var(--infoColorHover)" />
    </n-icon>
    <div>
      <h5>element</h5>
      <p>{{ log.element }}</p>
    </div>
    <n-icon class="log-details-arrow" size="24">
      <Mdi :path="mdiChevronRight" color="var(--infoColorHover)" />
    </n-icon>
    <div>
      <h5>pod</h5>
      <p>{{ log.pod }}</p>
    </div>
  </div>
  <div style="margin-top: 1em">
    <h4>{{ $t("fields.message") }}</h4>
    {{ log.message }}
    <!-- <pre>{{ JSON.parse(log.message) }}</pre> -->
  </div>
  <div style="margin-top: 1em">
    <h4>{{ $t("fields.details") }}</h4>
    <n-scrollbar x-scrollable style="max-height: 20rem">
      <pre>
      {{ logsDetails }}
    </pre
      >
    </n-scrollbar>
  </div>
</template>

<style scoped>
.log-details-divider {
  height: 30px;
}

.n-icon {
  margin-left: 10px;
  margin-right: 10px;
}
</style>
