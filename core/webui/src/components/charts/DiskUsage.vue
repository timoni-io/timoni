<script lang="ts" setup>
import { useI18n } from "vue-i18n";
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import chroma from "chroma-js";

const { t } = useI18n();

interface diskUsage {
  name: string;
  value: number;
  color: string;
}

let diskUsage = $ref<diskUsage[] | undefined>(undefined);
let diskUsageFull = $ref<number | undefined>(undefined);

api.get("/system-info").then((res) => {
  diskUsage = [];
  if (res.DiskUsage) {
    let colors = chroma
      .scale(["#fafa6e", "#2A4858"])
      .mode("lch")
      .colors(Object.keys(res.DiskUsage).length);

    Object.keys(res.DiskUsage).map((k) => {
      let value = res.DiskUsage ? res.DiskUsage[k] : 0;
      diskUsage?.push({
        name: k,
        value: value,
        color: "",
      });
    });

    diskUsage
      .sort((a, b) => {
        if (a.value > b.value) return -1;
        else if (a.value < b.value) return 1;
        else return 0;
      })
      .map((usage, index) => {
        usage.color = colors[index];
      });

    diskUsageFull = diskUsage.reduce(
      (partialSum, a) => partialSum + a.value,
      0
    );
  }
});

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "name",
    width: "40%",
  },
  {
    title: t("fields.value"),
    key: "value",
    template: "value",
    width: "60%",
  },
]);

const getPercentage = (val: number) => {
  if (!val || !diskUsageFull) return 0;
  return parseFloat(((val / diskUsageFull) * 100).toFixed(2));
};
</script>

<template>
  <n-card :title="diskUsage ? t('adminPanel.diskUsage') : ''" size="small">
    <Spinner :data="diskUsage">
      <div
        style="
          display: grid;
          grid-template-columns: 4fr 1fr;
          gap: 3rem;
          align-items: center;
        "
        v-if="diskUsage?.length"
      >
        <data-table :columns="columns" :data="diskUsage"
          ><template #value="row">
            <n-progress
              v-if="diskUsageFull"
              type="line"
              indicator-placement="inside"
              :color="row.color"
              :rail-color="row.color + '4D'"
              :percentage="getPercentage(row.value)" /></template
        ></data-table>
        <div style="display: flex; justify-content: center">
          <n-progress
            v-if="diskUsage && diskUsageFull"
            type="multiple-circle"
            :stroke-width="6"
            :circle-gap="0.5"
            :percentage="
              diskUsage.map((usage) => {
                if (!diskUsageFull) return 0;
                return (usage.value / diskUsageFull) * 100;
              })
            "
            :color="diskUsage.map((usage) => usage.color)"
            :rail-style="
              diskUsage.map((usage) => {
                return { stroke: usage.color, opacity: 0.3 };
              })
            "
          >
          </n-progress>
        </div>
      </div>
      <div class="base-alert" v-else>
        {{ t("messages.diskUsageEmpty") }}
      </div>
    </Spinner>
  </n-card>
</template>
