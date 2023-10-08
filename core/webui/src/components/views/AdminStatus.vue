<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useTimeFormatter } from "@/utils/formatTime";
import { Check } from "../../zodios/schemas/system";
import { z } from "zod";

const { t } = useI18n();
const { relativeOrDistanceToNow, formatTime } = useTimeFormatter();

const columnsSystemStatus: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
  },
  {
    title: t("fields.status"),
    template: "status",
    // key: "Status",
  },
  {
    title: t("fields.message"),
    key: "Message",
  },
  {
    title: t("fields.lastCheckTime"),
    template: "lastCheckTime",
  },
]);

let StatusByModules = $ref<z.infer<typeof Check>[]>([]);
// let notificationOn = $ref<boolean>(false);

api.get("/system-info").then((res) => {
  StatusByModules = res.StatusByModules.sort((a, b) => a.Name.localeCompare(b.Name));
  // notificationOn = res.NotificationsSend;
});

// const updateNotification = () => {
//   api.get("/system/notification-update", {
//     queries: { value: notificationOn },
//   });
// };
</script>

<template>
  <div>
    <AdminTabs />
  <PageLayout>
    <div class="admin-cards">
      <n-card :title="t('adminPanel.systemStatus')" size="small">
        <Spinner :data="StatusByModules">
          <data-table :columns="columnsSystemStatus" :data="StatusByModules" v-if="StatusByModules?.length">

            <template #lastCheckTime="{ LastCheckTime }">
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-tag type="default" :bordered="false" size="tiny">
                    {{
                      relativeOrDistanceToNow(
                        new Date(LastCheckTime * 1000)
                      )
                    }}
                  </n-tag>
                </template>
                {{ formatTime(new Date(LastCheckTime * 1000)) }}
                </n-tooltip>
              </template>

              <template #status="{ LastUpdateTime, Status }">



                <n-tooltip trigger="hover">
                  <template #trigger>

                    <n-tag :type="Status == 'ready' ? 'success' : 'error'" :bordered="false" size="tiny">
                      {{ Status }} &nbsp; &nbsp; {{
                        relativeOrDistanceToNow(new Date(LastUpdateTime * 1000))
                      }}
                    </n-tag>
                  </template>
                  {{ formatTime(new Date(LastUpdateTime * 1000)) }}
                </n-tooltip>

              </template>

            </data-table>
            <div class="base-alert" v-else>
              {{ t("messages.systemStatusEmpty") }}
            </div>
          </Spinner>
        </n-card>
        <!-- <n-card
                    :title="systemInfo ? t('adminPanel.notifications') : ''"
                    size="small"
                  >
                    <Spinner :data="systemInfo">
                      <div style="display: flex; align-items: center; gap: 1rem">
                        <n-switch
                          v-model:value="notificationOn"
                          @update:value="updateNotification"
                        >
                          <template #checked>On</template>
                          <template #unchecked>Off</template>
                        </n-switch>
                        <p style="font-size: medium">
                          {{ t("adminPanel.notificationsForSystem") }}
                        </p>
                      </div>
                    </Spinner>
                  </n-card> -->
      </div>
    </PageLayout>
  </div>
</template>
