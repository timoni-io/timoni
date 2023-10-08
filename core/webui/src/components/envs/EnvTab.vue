<script setup lang="ts">
import { useEnv } from "@/store/envStore";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import NavbarTabs from "@/components/layout/nav/NavbarTabs.vue";
// import { envIcon } from "@/utils/iconFactory";
import { useSpinner } from "@/store/spinner";
import useErrorMsg from "@/utils/errorMsg";
import { useUserStore } from "@/store/userStore";
const userStore = useUserStore();
let { envError } = useErrorMsg();

const spinner = useSpinner();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const env = useEnv(computed(() => route.params.id as string));

const currentTab = computed(() => {
  if (route.path.split("/")[2] === route.params.id) return "/";
  return route.path.split("/")[2];
});

const envTabs = computed(() => [
  {
    name: env.value?.EnvInfo?.Env?.Name,
    route: "/",
    path: "/",
    icon: {
      icon: mdiCloud,
      color: "",
      iconSize: undefined,
    },
    disabled: !userStore.havePermission("Env_View"),
  },
  {
    name: t("envTabs.elements"),
    route: "elements",
    path: "/elements/",
    icon: {
      icon: mdiApps,
      color: "",
      iconSize: undefined,
    },
    disabled: !userStore.havePermission("Env_View"),
  },
  {
    name: t("envTabs.variables"),
    route: "variables",
    path: "/variables/",
    icon: {
      icon: mdiFileReplaceOutline,
      color: "",
      iconSize: undefined,
    },
    disabled: !userStore.havePermission("Env_View"),
  },
  {
    name: t("objects.member", 2),
    route: "members",
    path: "/members/",
    icon: {
      icon: mdiAccountGroup,
      color: "",
      iconSize: undefined,
    },
    disabled: !userStore.havePermission("Env_ViewMembers"),
  },
  {
    name: t("envTabs.history"),
    route: "history",
    path: "/history/",
    icon: {
      icon: mdiHistory,
      color: "",
      iconSize: undefined,
    },
    disabled:
      !userStore.havePermission("Env_ViewLogsBuild") &&
      !userStore.havePermission("Env_ViewLogsEvents") &&
      !userStore.havePermission("Env_ViewLogsRuntime"),
  },
  // {
  //   name: t("envTabs.infra"),
  //   route: "infrastructure",
  //   path: "/infrastructure/",
  //   icon: {
  //     icon: mdiServerNetwork,
  //     color: "",
  //   },
  // },
  // {
  //   name: t("envTabs.topology"),
  //   route: "topology",
  //   path: "/topology/",
  //   icon: {
  //     icon: mdiLan,
  //     color: "",
  //   },
  // },
  {
    name: t("envTabs.pods"),
    route: "pods",
    path: "/pods/",
    icon: {
      icon: mdiHexagonMultiple,
      color: "",
      iconSize: undefined,
    },
    disabled: !userStore.havePermission("Env_View"),
  },
  {
    name: t("objects.statistics"),
    route: "statistics",
    path: "/statistics/",
    icon: {
      icon: mdiPoll,
      color: "",
      iconSize: undefined,
    },
    disabled: !userStore.havePermission("Env_View"),
  },
]);

let showRemoveModal = $ref(false);
const push = (path: string) => {
  if (route.path !== path) spinner.spinner = true;
  router.push(path);
};
</script>
<template>
  <NavbarTabs :activeTab="currentTab">
    <template #tabs>
      <NavTab
        v-for="tab in envTabs"
        :key="tab.name"
        :name="tab.route"
        @click="push(`/env${tab.path}${$router.currentRoute.value.params.id}`)"
        :disabled="tab.disabled"
      >
        <div style="display: flex; align-items: center">
          <n-icon style="margin-right: 3px; padding-bottom: 15px">
            <Mdi :path="tab.icon!.icon" :color="tab.icon!.color" />
          </n-icon>
          <p>{{ tab.name }}</p>
        </div>
      </NavTab>
    </template>
    <!-- <template #suffix>
      <n-button
        strong
        circle
        type="error"
        style="margin-right: 0.5rem"
        @click="showRemoveModal = true"
      >
        <template #icon>
          <n-icon><Mdi :path="mdiTrashCan" /></n-icon>
        </template>
      </n-button>
    </template> -->
  </NavbarTabs>
  <Modal
    v-model:show="showRemoveModal"
    :title="`${t('actions.deleteEnv')}  '${env?.EnvInfo?.Env?.Name}'`"
    :show-icon="false"
    style="width: 50%"
  >
    Komentuj nie usuwaj!
  </Modal>
  <DeletedEnvModal v-if="env?.EnvInfo?.Env.ToDelete || envError" />
</template>
