<script setup lang="ts">
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { SystemInfo } from "../../zodios/schemas/system";
import { useTimeFormatter } from "@/utils/formatTime";
import { useI18n } from "vue-i18n";
import { z } from "zod";

const { t } = useI18n();
const { relativeOrDistanceToNow, formatTime } = useTimeFormatter();

type SytemInfo = z.infer<typeof SystemInfo>;

let systemInfo = $ref<SytemInfo | undefined>(undefined);

api.get("/system-info").then((res) => {
  systemInfo = res;
});

// images builder queue
const columnsBuildingQueue: ComputedRef<Column[]> = computed(() => [
  // {
  //   title: t("fields.IP"),
  //   key: "IP",
  //   render: (_, index) => {
  //     return `${index + 1}`;
  //   },
  // },
  {
    title: t("fields.ImageID"),
    key: "ID",
  },
  {
    title: t("adminPanel.status"),
    key: "Status",
  },
  {
    title: t("fields.createdTimeStamp"),
    template: "createdTimeStamp",
  },
  {
    title: t("fields.startTime"),
    template: "startTime",
  },
]);

// images building status
const columnsBuildingStatus: ComputedRef<Column[]> = computed(() => [
  {
    title: "ID",
    template: "ID",
  },
  {
    title: "Nazwa Poda",
    key: "PodName",
  },
  {
    title: t("fields.IP"),
    key: "IP",
  },
  {
    title: t("fields.buildStatus"),
    template: "buildStatus",
  },
  {
    title: "Status HTTP Alive",
    template: "statusHTTPAlive",
  },
  {
    title: "Status Pod Exist",
    template: "statusPodExist",
  },
  {
    title: "Blueprint",
    template: "blueprint",
  },
]);

let showBlueprintDialog = $ref(false);
let blueprintToShow = $ref<string | undefined>(undefined);
</script>

<template>
  <div>
    <AdminTabs />
    <PageLayout>
      <div class="admin-cards">
        <n-card
          :title="systemInfo ? t('adminPanel.imagesBuildingQueue') : ''"
          size="small"
        >
          <Spinner :data="systemInfo">
            <data-table
              v-if="
                systemInfo &&
                systemInfo.ImageBuilderQueue &&
                systemInfo.ImageBuilderQueue.length > 0
              "
              :columns="columnsBuildingQueue"
              :data="
                systemInfo.ImageBuilderQueue
                  ? systemInfo.ImageBuilderQueue
                  : undefined
              "
            >
              <template #createdTimeStamp="row">
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <n-tag type="default" :bordered="false" size="tiny">
                      {{
                        relativeOrDistanceToNow(
                          new Date(row.TimeCreation * 1000)
                        )
                      }}
                    </n-tag>
                  </template>
                  {{ formatTime(new Date(row.TimeCreation * 1000)) }}
                </n-tooltip>
              </template>
              <template #startTime="row">
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <n-tag type="default" :bordered="false" size="tiny">
                      {{
                        relativeOrDistanceToNow(new Date(row.TimeBegin * 1000))
                      }}
                    </n-tag>
                  </template>
                  {{ formatTime(new Date(row.TimeBegin * 1000)) }}
                </n-tooltip>
              </template>
            </data-table>
            <div v-else>
              <div class="base-alert">
                {{ t("messages.imageQueueEmpty") }}
              </div>
            </div>
          </Spinner>
        </n-card>
        <n-card
          :title="systemInfo ? t('adminPanel.imagesBuildingStatus') : ''"
          size="small"
        >
          <Spinner :data="systemInfo">
            <data-table
              v-if="
                systemInfo &&
                systemInfo.ImageBuilderStatus &&
                Object.keys(systemInfo.ImageBuilderStatus).length > 0
              "
              :columns="columnsBuildingStatus"
              :data="Object.values(systemInfo.ImageBuilderStatus)"
            >
              <template #ID="row">
                {{
                  Object.keys(systemInfo.ImageBuilderStatus).indexOf(
                    row.PodName
                  ) + 1
                }}
              </template>
              <template #buildStatus="{ StatusBuilding }">
                <n-tag
                  :type="StatusBuilding ? 'success' : 'error'"
                  :bordered="false"
                  size="tiny"
                >
                  {{ StatusBuilding }}
                </n-tag>
              </template>
              <template #statusHTTPAlive="{ StatusHTTPAlive }">
                <n-tag
                  :type="StatusHTTPAlive ? 'success' : 'error'"
                  :bordered="false"
                  size="tiny"
                >
                  {{ StatusHTTPAlive }}
                </n-tag>
              </template>
              <template #statusPodExist="{ StatusPodExist }">
                <n-tag
                  :type="StatusPodExist ? 'success' : 'error'"
                  :bordered="false"
                  size="tiny"
                >
                  {{ StatusPodExist }}
                </n-tag>
              </template>
              <template #blueprint="{ Blueprint }">
                <div style="display: flex; align-items: center">
                  <n-button
                    :disabled="!Blueprint"
                    secondary
                    type="primary"
                    size="tiny"
                    @click="
                      () => {
                        showBlueprintDialog = true;
                        blueprintToShow = Blueprint;
                      }
                    "
                  >
                    <n-icon size="14px"
                      ><mdi :path="mdiNoteTextOutline"
                    /></n-icon>
                  </n-button>
                </div>
              </template>
            </data-table>
            <div v-else>
              <div class="base-alert">
                {{ t("messages.noImagesStatus") }}
              </div>
            </div>
          </Spinner>
        </n-card>
      </div>
    </PageLayout>

    <Modal
      :title="'Blueprint'"
      v-model:show="showBlueprintDialog"
      style="width: 40rem"
      :touched="false"
    >
      <pre>{{ blueprintToShow }}</pre>
    </Modal>
  </div>
</template>

<style scoped>
pre {
  white-space: pre-wrap;
  white-space: -moz-pre-wrap;
  white-space: -pre-wrap;
  white-space: -o-pre-wrap;
  word-wrap: break-word;
}
</style>
