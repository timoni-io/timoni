<script setup lang="ts">
// import { useI18n } from "vue-i18n";
import { ContainersInfo } from "@/zodios/schemas/env";
import { z } from "zod";
// import { computed, ComputedRef } from "vue";
// import { Column } from "@/components/utils/DataTable.vue";
// import { getContainerStatus } from "@/utils/getContainerStatus";
// import moment from "moment";

// const { t } = useI18n();

defineProps<{ containers: z.TypeOf<typeof ContainersInfo> }>();

// let displayContainers = $ref(false);

// watchEffect(() => {
//   localContainers = props.containers?.sort((a, b) =>
//     ("" + a.PodName).localeCompare(b.PodName)
//   );
// });
// columns
// const columns: ComputedRef<Column[]> = computed(() => [
//   {
//     title: "",
//     template: "state",
//     width: "3%",
//   },
//   {
//     title: t("fields.name"),
//     key: "PodName",
//   },
//   {
//     title: t("fields.created"),
//     template: "timeBegin",
//   },
//   {
//     title: t("fields.restartCount"),
//     key: "RestartCount",
//   },
//   {
//     title: t("fields.nodeName"),
//     key: "NodeName",
//   },
// ]);
</script>

<template>
  <div
    style="display: flex; align-items: center; gap: 0.3rem"
    v-if="containers"
  >
    <p style="width: 2em; margin-right: 0.5em; text-align: center">
      {{ containers.length }}
    </p>
    <!-- <n-button secondary type="primary" size="small" style="margin-right: 0.5em">
      <n-icon style="cursor: pointer; margin-right: 5px">
        <mdi :path="mdiConsole" width="15" />
      </n-icon>
      CLI
    </n-button> -->
    <!-- <n-tooltip
      v-for="container in containers.slice(0, 3)"
      :key="container.PodName"
      placement="right"
      trigger="hover"
    >
      <template #trigger>
        <n-icon :color="getContainerStatus(container.Status).color" size="14px"
          ><mdi :path="mdiCheckboxBlank"
        /></n-icon>
      </template>
      <div>
        <p v-for="col in columns" :key="col.title">
          <span style="font-weight: 700">{{ col.title }}</span>
          :
          {{
            col.key === "Status"
              ? getContainerStatus(container[col.key]).label
              : col.key != undefined
              ? // @ts-ignore
                container[col.key]
              : null
          }}
        </p>
      </div>
    </n-tooltip> -->
    <!-- <n-button
      v-if="containers.length > 3"
      secondary
      type="primary"
      size="small"
      class="change-version"
      @click="displayContainers = true"
    >
      <n-icon size="14px"><mdi :path="mdiAppsBox" /></n-icon>
      {{ t("actions.displayMore") }}
    </n-button>
    <Modal
      v-model:show="displayContainers"
      :title="t('objects.container', 2)"
      :touched="false"
      style="width: 60rem"
    >
      <data-table :columns="columns" :data="containers">
        <template #state="container">
          <div
            style="
              display: flex;
              gap: 10px;
              align-items: center;
              justify-content: flex-end;
            "
          >
            <n-tooltip placement="top" trigger="hover">
              <template #trigger>
                <div
                  :style="`background: ${
                    getContainerStatus(container.Status).color
                  };width:10px;height:10px;`"
                ></div>
              </template>
              <span>{{ getContainerStatus(container.Status).label }}</span>
            </n-tooltip>
          </div>
        </template>

        <template #status="row">
          <div style="display: flex; align-items: center; gap: 0.3rem">
            <n-icon :color="getContainerStatus(row.Status).color" size="14px"
              ><mdi :path="mdiCheckboxBlank"
            /></n-icon>
            {{ getContainerStatus(row.Status).label }}
          </div>
        </template>
        <template #timeBegin="container">
          <div v-if="container.StartTime">
            {{ moment(container.StartTime * 1000).fromNow() }}
          </div>
        </template>
      </data-table>
    </Modal> -->
  </div>
</template>
