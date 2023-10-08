<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { systemResourcesObject, Node } from "../../zodios/schemas/system";
import { z } from "zod";
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useThemeVars } from "naive-ui";
import { changeColor } from "seemly";

const { t } = useI18n();
const themeVars = useThemeVars();

let nodesData = $ref<z.infer<typeof Node>[] | undefined>(undefined);

api.get("/system-info").then((res) => {
  if (res.Nodes) {
    nodesData = Object.values(res.Nodes);
  } else nodesData = [];
});

// nodes
const columnsNodeData: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
    width: "15%",
  },
  {
    title: "IP",
    key: "IP",
    width: "15%",
  },
  {
    title: "CPU",
    template: "CPU",
    width: "20%",
  },
  {
    title: "RAM",
    template: "RAM",
    width: "20%",
  },
  {
    title: "",
    template: "releaseTime",
    width: "15%",
    align: "center",
  },
  {
    title: "Status",
    template: "status",
    width: "15%",
  },
]);

const calcPercentage = (node: {
  Usage: number;
  Guaranteed: number;
  Max: number;
  Capacity: number;
}) => {
  return Math.round(((node.Capacity - node.Guaranteed) / node.Capacity) * 100);
};

const availableColor = (node: {
  Usage: number;
  Guaranteed: number;
  Max: number;
  Capacity: number;
}) => {
  let percentage = ((node.Capacity - node.Guaranteed) / node.Capacity) * 100;
  let color = "";

  switch (true) {
    case percentage >= 40:
      color = themeVars.value.successColor;
      break;
    case percentage >= 15:
      color = themeVars.value.warningColor;
      break;
    case percentage < 15:
      color = themeVars.value.errorColor;
      break;
  }

  return color;
};

// CPUAndRAMUsage - /system-resources
interface ResourceRow {
  Name: string;
  CPUGuarantee: number | undefined;
  CPUUsage: number | undefined;
  RAMGuarantee: number | undefined;
  RAMUsage: number | undefined;
}
const columnsCPUData: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
    width: "20%",
  },
  {
    title: t("adminPanel.CPUGuarantee"),
    key: "CPUGuarantee",
    template: "CPUGuarantee",
    width: "35%",
    sorter: (row1: ResourceRow, row2: ResourceRow) => {
      if (
        row1?.CPUGuarantee !== undefined &&
        row2?.CPUGuarantee !== undefined
      ) {
        return row1.CPUGuarantee - row2.CPUGuarantee;
      } else {
        return 0;
      }
    },
    sortOrder: "descend",
  },
  {
    title: t("adminPanel.CPUUsage"),
    key: "CPUUsage",
    template: "CPUUsage",
    width: "35%",
  },
  {
    title: "",
    template: "details",
    width: "10%",
  },
]);

const columnsRAMData: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
    width: "20%",
  },
  {
    title: t("adminPanel.RAMGuarantee"),
    key: "RAMGuarantee",
    template: "RAMGuarantee",
    width: "35%",
    sorter: (row1: ResourceRow, row2: ResourceRow) => {
      if (
        row1?.RAMGuarantee !== undefined &&
        row2?.RAMGuarantee !== undefined
      ) {
        return row1.RAMGuarantee - row2.RAMGuarantee;
      } else {
        return 0;
      }
    },
    sortOrder: "descend",
  },
  {
    title: t("adminPanel.RAMUsage"),
    key: "RAMUsage",
    template: "RAMUsage",
    width: "35%",
  },
  {
    title: "",
    template: "details",
    width: "10%",
  },
]);

type SytemResources = z.infer<typeof systemResourcesObject>;

let resourcesData = $ref<SytemResources | undefined>(undefined);
let resourcesDataName = $ref<ResourceRow[]>([]);

const getResourseVal = (
  name: string,
  type: "CPU" | "RAM",
  data: "Guaranteed" | "Usage"
) => {
  if (!resourcesData) return undefined;
  let capacity =
    type === "CPU" ? resourcesData.CPUCapacity : resourcesData.RAMCapacity;
  return Math.round(
    (resourcesData.Resources[name].Total[type][data] / capacity) * 100
  );
};

api.get("/system-resources").then((res) => {
  resourcesData = res;
  if (resourcesData.Resources) {
    Object.keys(resourcesData.Resources).map((r) => {
      resourcesDataName.push({
        Name: r,
        CPUGuarantee: getResourseVal(r, "CPU", "Guaranteed"),
        CPUUsage: getResourseVal(r, "CPU", "Usage"),
        RAMGuarantee: getResourseVal(r, "RAM", "Guaranteed"),
        RAMUsage: getResourseVal(r, "RAM", "Usage"),
      });
    });
  }
});

let detailsResorceShow = $ref(false);
let detailsResorceName = $ref<string>("");
let detailsResorceType = ref<"CPU" | "RAM">("CPU");
</script>

<template>
  <div>
    <AdminTabs />
    <n-scrollbar
      style="
        height: 100%;
        max-height: calc(100vh - 2.5rem);
        transform: translateY(-1rem);
      "
    >
      <PageLayout>
        <div class="admin-cards" style="margin: 1rem 0">
          <n-card :title="nodesData ? t('adminPanel.nodes') : ''" size="small">
            <Spinner :data="nodesData">
              <data-table :columns="columnsNodeData" :data="nodesData">
                <template #CPU="node">
                  <div class="resource-progress">
                    <n-progress
                      v-if="node"
                      type="line"
                      :show-indicator="false"
                      :color="availableColor(node.Resources.CPU)"
                      :rail-color="
                        changeColor(availableColor(node.Resources.CPU), {
                          alpha: 0.2,
                        })
                      "
                      :percentage="calcPercentage(node.Resources.CPU)"
                    />
                    <p>
                      available
                      {{
                        (node.Resources.CPU.Capacity -
                          node.Resources.CPU.Guaranteed) /
                        100
                      }}
                      vCore
                    </p>
                  </div>
                </template>
                <template #RAM="node">
                  <div class="resource-progress">
                    <n-progress
                      v-if="node"
                      type="line"
                      :show-indicator="false"
                      :color="availableColor(node.Resources.RAM)"
                      :rail-color="
                        changeColor(availableColor(node.Resources.RAM), {
                          alpha: 0.2,
                        })
                      "
                      :percentage="calcPercentage(node.Resources.RAM)"
                    />
                    <p>
                      available
                      {{
                        node.Resources.RAM.Capacity -
                        node.Resources.RAM.Guaranteed
                      }}
                      MB
                    </p>
                  </div>
                </template>
                <template #releaseTime="node">
                  <a
                    :href="'http://' + node.IP + ':10002/play'"
                    target="_blank"
                    class="clickhouse"
                  >
                    <n-tag
                      size="tiny"
                      :bordered="false"
                      :type="node.ClickHouseIsReady ? 'success' : 'error'"
                      style="cursor: pointer"
                    >
                      <n-icon color="yellow"
                        ><mdi :path="mdiDatabase"
                      /></n-icon>
                      clickhouse
                    </n-tag>
                  </a>
                </template>
                <template #status="{ Ready }">
                  <n-tag
                    size="tiny"
                    :bordered="false"
                    :type="Ready ? 'success' : 'error'"
                    >{{
                      Ready ? t("adminPanel.ready") : t("adminPanel.unready")
                    }}</n-tag
                  >
                </template>
              </data-table>
            </Spinner>
          </n-card>

          <DiskUsage />
          <n-card
            :title="resourcesData ? t('adminPanel.CPUAndRAMUsage') : ''"
            size="small"
          >
            <Spinner :data="resourcesData">
              <div style="display: grid; gap: 1rem">
                <data-table :columns="columnsCPUData" :data="resourcesDataName">
                  <template #CPUGuarantee="{ CPUGuarantee, Name }">
                    <n-tooltip trigger="hover" v-if="resourcesData">
                      <template #trigger>
                        <n-progress
                          type="line"
                          indicator-placement="inside"
                          :color="themeVars.successColor"
                          :rail-color="
                            changeColor(themeVars.successColor, { alpha: 0.2 })
                          "
                          :percentage="CPUGuarantee"
                        />
                      </template>
                      {{
                        `${resourcesData.Resources[Name].Total.CPU.Guaranteed} / ${resourcesData.CPUCapacity} vCores`
                      }}
                    </n-tooltip>
                  </template>
                  <template #CPUUsage="{ CPUUsage, Name }">
                    <n-tooltip trigger="hover" v-if="resourcesData">
                      <template #trigger>
                        <n-progress
                          type="line"
                          indicator-placement="inside"
                          :color="themeVars.successColor"
                          :rail-color="
                            changeColor(themeVars.successColor, { alpha: 0.2 })
                          "
                          :percentage="CPUUsage"
                        />
                      </template>
                      {{
                        `${resourcesData.Resources[Name].Total.CPU.Usage} / ${resourcesData.CPUCapacity} vCores`
                      }}
                    </n-tooltip>
                  </template>
                  <template #details="{ Name }"
                    ><div
                      class="magnify-resources"
                      @click="
                        () => {
                          detailsResorceShow = true;
                          detailsResorceName = Name;
                          detailsResorceType = 'CPU';
                        }
                      "
                    >
                      <n-icon><mdi :path="mdiMagnify" /></n-icon></div
                  ></template>
                </data-table>
                <data-table :columns="columnsRAMData" :data="resourcesDataName">
                  <template #RAMGuarantee="{ RAMGuarantee, Name }">
                    <n-tooltip trigger="hover" v-if="resourcesData">
                      <template #trigger>
                        <n-progress
                          type="line"
                          indicator-placement="inside"
                          :color="themeVars.successColor"
                          :rail-color="
                            changeColor(themeVars.successColor, { alpha: 0.2 })
                          "
                          :percentage="RAMGuarantee"
                        />
                      </template>
                      {{
                        `${resourcesData.Resources[Name].Total.RAM.Guaranteed} / ${resourcesData.RAMCapacity} MiB`
                      }}
                    </n-tooltip>
                  </template>
                  <template #RAMUsage="{ RAMUsage, Name }">
                    <n-tooltip trigger="hover" v-if="resourcesData">
                      <template #trigger>
                        <n-progress
                          type="line"
                          indicator-placement="inside"
                          :color="themeVars.successColor"
                          :rail-color="
                            changeColor(themeVars.successColor, { alpha: 0.2 })
                          "
                          :percentage="RAMUsage"
                        />
                      </template>
                      {{
                        `${resourcesData.Resources[Name].Total.RAM.Usage} / ${resourcesData.RAMCapacity} MiB`
                      }}
                    </n-tooltip>
                  </template>
                  <template #details="{ Name }"
                    ><div
                      class="magnify-resources"
                      @click="
                        () => {
                          detailsResorceShow = true;
                          detailsResorceName = Name;
                          detailsResorceType = 'RAM';
                        }
                      "
                    >
                      <n-icon><mdi :path="mdiMagnify" /></n-icon></div
                  ></template>
                </data-table>
              </div>
            </Spinner>
          </n-card>
        </div>
      </PageLayout>
    </n-scrollbar>

    <Modal
      v-model:show="detailsResorceShow"
      :title="`${detailsResorceName} - ${detailsResorceType} ${t(
        'fields.details'
      )}`"
      :touched="false"
      style="width: 80vw; max-width: 56rem"
    >
      <n-table :single-line="false" v-if="resourcesData" :bordered="false">
        <thead>
          <tr>
            <th>{{ t("fields.name") }}</th>
            <th>
              {{
                detailsResorceType === "CPU"
                  ? t("adminPanel.CPUGuarantee")
                  : t("adminPanel.RAMGuarantee")
              }}
            </th>
            <th>
              {{
                detailsResorceType === "CPU"
                  ? t("adminPanel.CPUUsage")
                  : t("adminPanel.RAMUsage")
              }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(value, name) in resourcesData.Resources[detailsResorceName]
              .Elements"
            :key="name"
          >
            <td style="width: 30%">{{ name }}</td>
            <td style="width: 35%">
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-progress
                    v-if="resourcesData"
                    type="line"
                    indicator-placement="inside"
                    :color="themeVars.successColor"
                    :rail-color="
                      changeColor(themeVars.successColor, { alpha: 0.2 })
                    "
                    :percentage="
                      Math.round(
                        (value[detailsResorceType].Guaranteed /
                          (detailsResorceType === 'CPU'
                            ? resourcesData.CPUCapacity
                            : resourcesData.RAMCapacity)) *
                          100
                      )
                    "
                  />
                </template>
                {{
                  `${value[detailsResorceType].Guaranteed} / ${resourcesData.CPUCapacity} vCores`
                }}
              </n-tooltip>
            </td>
            <td style="width: 35%">
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-progress
                    v-if="resourcesData"
                    type="line"
                    indicator-placement="inside"
                    :color="themeVars.successColor"
                    :rail-color="
                      changeColor(themeVars.successColor, { alpha: 0.2 })
                    "
                    :percentage="
                      Math.round(
                        (value[detailsResorceType].Usage /
                          (detailsResorceType === 'CPU'
                            ? resourcesData.CPUCapacity
                            : resourcesData.RAMCapacity)) *
                          100
                      )
                    "
                  />
                </template>
                {{
                  `${value[detailsResorceType].Usage} / ${resourcesData.RAMCapacity} MiB`
                }}
              </n-tooltip>
            </td>
          </tr>
        </tbody>
      </n-table>
    </Modal>
  </div>
</template>

<style lang="scss">
.resource-progress
  .n-progress
  .n-progress-graph
  .n-progress-graph-line
  .n-progress-graph-line-rail {
  height: 1.2rem;
}

.resource-progress {
  position: relative;
  padding: 0.25rem 0;

  p {
    position: absolute;
    top: 0;
    color: rgba(0, 0, 0, 0.9);
    font-size: 0.75rem;
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: 100%;
  }
}

.magnify-resources {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.n-modal .n-table td {
  padding-top: 4px !important;
  padding-bottom: 4px !important;
}

.n-modal .n-table tr,
.n-modal .n-table th,
.n-modal .n-table thead {
  background: none;
}

.n-modal .n-table th {
  padding-top: 4px !important;
  padding-bottom: 4px !important;
  font-size: var(--fontSizeMini);
}
</style>
