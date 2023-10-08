<script setup lang="ts">
import { useRoute } from "vue-router";
import moment from "moment";
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useI18n } from "vue-i18n";
// import { getContainerStatus } from "@/utils/getContainerStatus";
import { renderIcon } from "@/utils/renderIcon";
import { useMessage } from "naive-ui";
import { podIcon } from "@/utils/iconFactory";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const { t } = useI18n();

type OptionKey = "restart" | "delete";
type ContainerRes = ResType<"/env-pods">[string];
type ContainerResArray = Array<ContainerRes>;

const route = useRoute();
const message = useMessage();

let containers = $ref<ContainerResArray | undefined>();
let currentActionName = $ref("");
let currentDialogComponent = $ref<OptionKey | null>(null);
let restartModalConfirm = $ref(false);
let currentElementName = $ref("");
let currentPodName = $ref("");
const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: "",
    template: "state",
    width: "3%",
  },
  {
    title: t("fields.name"),
    key: "PodName",
    width: "35%",
  },
  {
    title: t("fields.created"),
    template: "timeBegin",
    width: "15%",
  },
  {
    title: t("fields.restartCount"),
    key: "RestartCount",
  },
  {
    title: t("fields.nodeName"),
    key: "NodeName",
  },
  {
    title: "",
    template: "terminal",
    width: "4%",
  },
]);

onMounted(() => {
  api
    .get("/env-pods", {
      queries: {
        env: route.params.id as string,
      },
    })
    .then((r) => {
      containers = Object.values(r);
    });
});

const options = $ref([
  {
    label: t("elements.options.restart"),
    key: "restart",
    icon: renderIcon(mdiRestart),
  },
  // {
  //   label: t("elements.options.delete"),
  //   key: "delete",
  //   icon: renderIcon(mdiDelete),
  // },
]);

const handleSelect = (key: OptionKey, action: any) => {
  currentActionName = action.Name;
  currentElementName = action.ElementName;
  currentPodName = action.PodName;
  currentDialogComponent = key;
  switch (currentDialogComponent) {
    case "restart":
      restartModalConfirm = true;
      break;
    case "delete":
      api
        .post(
          "/env-mod",
          {
            DeleteElements: [currentActionName],
            Elements: {},
            Apply: true,
          },
          {
            queries: {
              env: route.params.id as string,
            },
          }
        )
        .then((res) => {
          if (res[currentActionName].Status === 0)
            message.success(t("messages.actionDeleted"));
          else message.error(res[currentActionName].Message);
        });
      break;
  }
};

useIntervalFn(() => {
  api
    .get("/env-pods", {
      queries: {
        env: route.params.id as string,
      },
    })
    .then((r) => {
      containers = Object.values(r);
    });
}, 1000);

const restartPod = () => {
  api
    .get("/env-pod-restart", {
      queries: {
        env: route.params.id as string,
        element: currentElementName as string,
        pod: currentPodName as string,
      },
    })
    .then(() => {
      message.success(t("messages.elementRestarted"));
      restartModalConfirm = false;
    })
    .catch((err) => {
      message.error(err);
      restartModalConfirm = false;
    });
};

// window width
let screenWidth = $ref<number>(0);

onBeforeMount(() => {
  screenWidth = window.innerWidth;
});

onMounted(() => {
  window.addEventListener("resize", () => {
    screenWidth = window.innerWidth;
  });
});

// terminal
const openTerminal = (el: ContainerRes) => {
  window.open(
    `/cli?env=${route.params.id}&element=${el.ElementName}`,
    "_blank"
  );
};
</script>

<template>
  <div>
    <EnvTab />
    <PageLayout>
      <n-card
        v-if="userStore.havePermission('Env_View')"
        :title="t('objects.elementPod', 2)"
        style="margin-bottom: 1em; height: calc(100vh - 5rem)"
        size="small"
      >
        <data-table
          :columns="columns"
          :data="
            containers?.sort((a : any, b: any) =>
              ('' + a.PodName).localeCompare(b.PodName)
            )
          "
        >
          <template #state="container">
            <!-- <div
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
                <span>{{ container.CheckResult.Msg }}</span>
              </n-tooltip>
            </div> -->
            <ElementStatusIcon
              :status="container.Status"
              :toDelete="false"
              :iconFunc="podIcon"
            />
            <!-- tutaj trzeba dodać pole ToDelete, na razie z backu nie przychodzi -->
          </template>
          <template #timeBegin="container">
            <div v-if="container.CreationTime">
              {{ moment(container.CreationTime * 1000).fromNow() }}
            </div>
          </template>
          <template #terminal="row">
            <div
              style="
                width: 100%;
                height: 100%;
                display: flex;
                justify-content: center;
              "
            >
              <n-tooltip trigger="hover" :disabled="screenWidth > 1400">
                <template #trigger>
                  <n-button
                    secondary
                    type="primary"
                    size="tiny"
                    style="margin-right: 0.5em"
                    @click="() => openTerminal(row)"
                  >
                    <n-icon style="cursor: pointer">
                      <mdi :path="mdiConsole" width="15" />
                    </n-icon>
                    <span style="margin-left: 5px" v-if="screenWidth > 1400">
                      Terminal
                    </span>
                  </n-button>
                </template>
                Terminal
              </n-tooltip>

              <n-dropdown
                trigger="click"
                :options="options"
                placement="bottom-end"
                @select="handleSelect($event, row)"
              >
                <div class="dropdown-menu">
                  <n-button
                    secondary
                    type="primary"
                    size="tiny"
                    class="dots-button"
                  >
                    <n-icon size="22px"
                      ><mdi :path="mdiDotsHorizontal"
                    /></n-icon>
                  </n-button>
                </div>
              </n-dropdown>
            </div>
          </template>
        </data-table>
      </n-card>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />

      <!-- <n-card v-if="env?.Actions" :title="t('objects.actionPod', 2)" size="small">
      <EnvironmentActions />
    </n-card> -->
      <Modal
        v-model:show="restartModalConfirm"
        title="Restart pod"
        style="width: 25rem"
        :showFooter="true"
        @positive-click="restartPod"
        @negative-click="restartModalConfirm = false"
      >
        {{ t("questions.sure") }}
      </Modal>
    </PageLayout>
  </div>
</template>

<style scoped>
.dropdown-menu {
  display: flex;
  align-items: center;
  justify-content: right;
  cursor: pointer;
  opacity: 0.6;
  transition: 0.2s ease-in-out;
}
.dropdown-menu:hover {
  opacity: 1;
}
</style>
