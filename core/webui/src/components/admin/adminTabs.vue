<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import NavbarTabs from "@/components/layout/nav/NavbarTabs.vue";
import { useSpinner } from "@/store/spinner";

const spinner = useSpinner();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const currentTab = computed(() => {
  if (route.path.split("/")[2] === route.params.id) return "/";
  return route.path.split("/")[2];
});
const envTabs = computed(() => [
  {
    name: t("adminPanel.platform"),
    route: "platform",
    path: "/admin/platform",
    icon: mdiApplicationCogOutline,
  },
  // {
  //   name: t("adminPanel.operators"),
  //   route: "users-list",
  //   path: "/admin/users-list",
  //   icon: mdiAccountGroup,
  // },
  {
    name: t("adminPanel.containerImages"),
    route: "image-list",
    path: "/admin/image-list",
    icon: mdiPackageVariantClosedCheck,
  },
  {
    name: t("adminPanel.ElementVersions"),
    route: "element-versions",
    path: "/admin/element-versions",
    icon: mdiPackageVariantClosedCheck,
  },
  // {
  //   name: t("adminPanel.systemLogs"),
  //   route: "logs",
  //   path: "/admin/logs",
  // },
  // {
  //   name: t("adminPanel.logsAnalysis"),
  //   route: "histogram",
  //   path: "/admin/histogram",
  // },
  {
    name: t("adminPanel.status"),
    route: "status",
    path: "/admin/status",
    icon: mdiListStatus,
  },
  // {
  //   name: t("adminPanel.certificates"),
  //   route: "certificates",
  //   path: "/admin/certificates",
  //   icon: mdiCertificateOutline,
  // }, commented after removal of '/api/system-certs'
  {
    name: t("adminPanel.buildingImages"),
    route: "building-images",
    path: "/admin/building-images",
    icon: mdiPackageVariantClosed,
  },
  {
    name: t("adminPanel.resources"),
    route: "resources",
    path: "/admin/resources",
    icon: mdiChartBar,
  },
]);

const push = (path: string) => {
  if (route.path !== path) spinner.spinner = true;
  router.push(path);
};
</script>
<template>
  <NavbarTabs :activeTab="currentTab">
    <template #tabs>
      <NavTab v-for="tab in envTabs" :key="tab.name" :name="tab.route" @click="push(`${tab.path}`)">
        <div style="display: flex; align-items: center">
          <n-icon-wrapper :size="17" :border-radius="24" :color="route.path === tab.path ? 'white' : 'black'"
            style="margin-right: 4px; background: rgba(0, 0, 0, 0)">
            <n-icon :color="route.path === tab.path ? 'white' : 'black'">
              <Mdi :path="tab.icon" />
            </n-icon>
          </n-icon-wrapper>
          {{ tab.name }}
        </div>
      </NavTab>
    </template>
  </NavbarTabs>
</template>
