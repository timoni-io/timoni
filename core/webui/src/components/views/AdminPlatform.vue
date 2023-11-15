<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const { t } = useI18n();

let platformData = $ref<any>(undefined);

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("adminPanel.domainCluster"),
    key: "clusterDomain",
    width: "30%",
  },
  {
    title: t("adminPanel.releaseHash"),
    key: "gitTag",
    width: "40%",
  },
]);

onMounted(() => {
  api.get("/system-version").then((res) => {
    platformData = [res];
  });
});
</script>

<template>
  <div>
    <AdminTabs />
    <PageLayout>
      <div
        class="admin-cards"
        v-if="userStore.havePermission('Glob_AccessToAdminZone')"
      >
        <n-card
          :title="platformData ? t('adminPanel.platform') : ''"
          class="data-table-card"
          size="small"
        >
          <Spinner :data="platformData">
            <data-table
              :columns="columns"
              :data="platformData"
              class="data-table"
            >
            </data-table>
          </Spinner>
        </n-card>
      </div>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
  </div>
</template>
