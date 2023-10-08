<script setup lang="ts">
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useI18n } from "vue-i18n";
import { systemCert } from "../../zodios/schemas/system";
import { z } from "zod";

const { t } = useI18n();

let certData = $ref<z.infer<typeof systemCert>[] | undefined>(undefined);

api.get("/system-certs").then((res) => {
  certData = res;
});

const columnsCerts: ComputedRef<Column[]> = computed(() => [
  {
    title: t("objects.domain"),
    key: "Name",
    width: "30%",
  },
  {
    title: t("adminPanel.daysLeft"),
    template: "daysLeft",
    width: "30%",
  },
  {
    title: t("adminPanel.certificate"),
    template: "certificate",
    width: "30%",
  },
  {
    title: "",
    template: "more",
    width: "10%",
  },
]);

let certificateDetailsShow = $ref(false);
let certificateDetails = $ref<z.infer<typeof systemCert> | undefined>(
  undefined
);
const certificateDetailsColumn: {
  title: string;
  dataCol: keyof z.infer<typeof systemCert>;
}[] = [
  { title: "objects.domain", dataCol: "Name" },
  { title: "adminPanel.ingressName", dataCol: "IngressName" },
  { title: "adminPanel.ingressNamespace", dataCol: "IngressNamespace" },
  { title: "adminPanel.secretName", dataCol: "SecretName" },
  { title: "adminPanel.secretExist", dataCol: "SecretExist" },
  {
    title: "adminPanel.secretExpirationDaysLeft",
    dataCol: "SecretExpirationDaysLeft",
  },
  { title: "adminPanel.certName", dataCol: "CertName" },
  { title: "adminPanel.certNamespace", dataCol: "CertNamespace" },
  { title: "adminPanel.certExist", dataCol: "CertExist" },
  { title: "adminPanel.certReady", dataCol: "CertReady" },
  { title: "adminPanel.Timoni", dataCol: "Timoni" },
  { title: "adminPanel.usedIn", dataCol: "UsedIn" },
];
</script>

<template>
  <div>
    <AdminTabs />
    <PageLayout>
      <div class="admin-cards">
        <n-card
          :title="certData ? t('adminPanel.certificationsStatus') : ''"
          size="small"
        >
          <Spinner :data="certData">
            <data-table :columns="columnsCerts" :data="certData">
              <template #daysLeft="{ SecretExpirationDaysLeft }">
                <n-tag
                  :type="SecretExpirationDaysLeft >= 7 ? 'success' : 'error'"
                  :bordered="false"
                  size="tiny"
                >
                  {{ SecretExpirationDaysLeft }}
                </n-tag>
              </template>
              <template #certificate="{ CertReady }">
                <n-tag
                  :type="CertReady ? 'success' : 'error'"
                  :bordered="false"
                  size="tiny"
                >
                  {{ CertReady }}
                </n-tag>
              </template>
              <template #more="row">
                <div
                  @click="
                    () => {
                      certificateDetailsShow = true;
                      certificateDetails = row;
                    }
                  "
                  style="cursor: pointer"
                >
                  <n-icon><mdi :path="mdiMagnify" /></n-icon>
                </div>
              </template>
            </data-table>
          </Spinner>
        </n-card>
      </div>
    </PageLayout>

    <Modal
      v-model:show="certificateDetailsShow"
      :title="t('adminPanel.certificateDetails')"
      :touched="false"
      style="width: 80vw; max-width: 56rem"
    >
      <n-table v-if="certificateDetails" :single-line="false" :bordered="false">
        <tr v-for="col in certificateDetailsColumn" :key="col.dataCol">
          <th style="width: 50%; padding-left: 1rem">{{ t(col.title) }}</th>
          <td v-if="typeof certificateDetails[col.dataCol] === 'boolean'">
            <n-tag
              :type="certificateDetails[col.dataCol] ? 'success' : 'error'"
              :bordered="false"
            >
              {{ certificateDetails[col.dataCol] }}
            </n-tag>
          </td>
          <td v-else-if="col.dataCol === 'SecretExpirationDaysLeft'">
            <n-tag
              :type="certificateDetails[col.dataCol] >= 7 ? 'success' : 'error'"
              :bordered="false"
            >
              {{ certificateDetails[col.dataCol] }}
            </n-tag>
          </td>
          <td
            v-else-if="col.dataCol === 'UsedIn'"
            style="width: 50%; padding-left: 1rem"
          >
            {{
              certificateDetails[col.dataCol]
                ? certificateDetails[col.dataCol].join(", ")
                : null
            }}
          </td>
          <td v-else style="width: 50%; padding-left: 1rem">
            {{ certificateDetails[col.dataCol] }}
          </td>
        </tr>
      </n-table>
    </Modal>
  </div>
</template>
