<script setup lang="ts">
import { useRoute } from "vue-router";
// import { useEnv } from "@/store/envStore";
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useI18n } from "vue-i18n";
import { useTimeFormatter } from "@/utils/formatTime";
import { renderIcon } from "@/utils/renderIcon";
import { useMessage } from "naive-ui";
import {
  getContainerStatus,
  actionStatusToNumber,
} from "@/utils/getContainerStatus";

const { relativeOrDistanceToNow } = useTimeFormatter();

const { t } = useI18n();
const route = useRoute();
const message = useMessage();

// const env = useEnv(computed(() => route.params.id as string));

type Warning = {
  Message: string;
  ExitCode: number;
  Reason: string;
  RestartCount: number;
};

type OptionKey = "restart" | "delete";

type Action = {
  Name: string;
  Status: string;
  Error: string;
  Warnings: { [k: string]: string[] | Warning[] };
  GitRepoName: string;
  Branch: string;
  Commit: string;
  FileName: string;
  Tags: string[];
  ParentName: string;
  ActionName: string;
  TimeBegin: number;
  TimeEnd: number;
};

let warningsShow = $ref(false);
let warnings = $ref({} as Record<string, Warning[]>);
let warningsName = $ref("");

let openDialog = $ref<boolean>(false);
let currentDialogComponent = $ref<OptionKey | null>(null);
let currentAction = $ref<Action | null>(null);
let currentActionName = $ref("");

const showWarnings = (el: any) => {
  warningsName = el.Name;
  warnings = toRaw(el.Warnings);
  warningsShow = true;
};

const handleSelect = (key: OptionKey, action: any) => {
  currentActionName = action.Name;
  currentAction = action;
  currentDialogComponent = key;
  switch (currentDialogComponent) {
    case "restart":
      api
        .get("/env-element-restart-pods", {
          queries: {
            env: route.params.id as string,
            element: currentAction!.Name,
          },
        })
        .then(() => {
          message.success(t("messages.elementRestarted"));
          openDialog = false;
        });
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
          openDialog = false;
          if (res[currentActionName].Status === 0)
            message.success(t("messages.actionDeleted"));
          else message.error(res[currentActionName].Message);
        })
        .catch(() => {
          openDialog = false;
        });
      break;
    default:
      openDialog = true;
  }
};

const options = $ref([
  {
    label: t("elements.options.restart"),
    key: "restart",
    icon: renderIcon(mdiRestart),
  },
  {
    label: t("elements.options.delete"),
    key: "delete",
    icon: renderIcon(mdiDelete),
  },
]);

const positiveClick = () => {
  switch (currentDialogComponent) {
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
          openDialog = false;
          if (res[currentActionName].Status === 0)
            message.success(t("messages.elementDeleted"));
          else message.error(res[currentActionName].Message);
        })
        .catch(() => {
          openDialog = false;
        });
      break;
    case "restart":
      api
        .get("/env-element-restart-pods", {
          queries: {
            env: route.params.id as string,
            element: currentAction!.Name,
          },
        })
        .then(() => {
          message.success(t("messages.actionRestarted"));
          openDialog = false;
        });
      break;
  }
};

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: "",
    template: "state",
    width: "3%",
  },
  {
    title: t("fields.name"),
    key: "Name",
    width: "35%",
  },
  // {
  //   title: t("fields.project"),
  //   key: "ProjectName",
  //   width: "10%",
  // },
  // {
  //   title: t("fields.branch"),
  //   key: "Branch",
  //   width: "10%",
  // },
  // {
  //   title: t("fields.version"),
  //   template: "version",
  //   width: "15%",
  // },
  {
    title: t("fields.created"),
    template: "time",
    width: "15%",
  },
  {
    title: t("objects.warnings"),
    template: "message",
  },
  {
    title: "",
    template: "actions",
    width: "20%",
  },
]);
</script>

<template>
  <data-table :columns="columns" :data="undefined">
    <template #version="row">
      {{ row.Commit.substring(0, 8) }}
    </template>
    <template #time="row">
      {{ relativeOrDistanceToNow(new Date(row.TimeBegin)) }}
    </template>
    <template #message="row">
      <div style="display: flex; align-items: center">
        <n-button
          v-if="Object.values(toRaw(row.Warnings || [])).flat().length"
          strong
          secondary
          circle
          type="error"
          size="tiny"
          style="margin-right: 5px"
          @click="showWarnings(row)"
        >
          <template #icon>
            <n-icon><Mdi :path="mdiExclamation" /></n-icon>
          </template>
        </n-button>
        <p>{{ row.StatusMessage }}</p>
      </div>
    </template>
    <template #actions="row">
      <div
        style="
          width: 100%;
          height: 100%;
          display: flex;
          justify-content: flex-end;
        "
      >
        <n-button
          secondary
          type="primary"
          size="small"
          style="margin-right: 0.5em"
        >
          <n-icon style="cursor: pointer; margin-right: 5px">
            <mdi :path="mdiConsole" width="15" />
          </n-icon>
          CLI
        </n-button>
        <n-dropdown
          trigger="click"
          :options="options"
          placement="bottom-end"
          @select="handleSelect($event, row)"
        >
          <div class="dropdown-menu">
            <n-button secondary type="primary" size="small" class="dots-button">
              <n-icon size="22px"><mdi :path="mdiDotsHorizontal" /></n-icon>
            </n-button>
          </div>
        </n-dropdown>
      </div>
    </template>
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
                getContainerStatus(actionStatusToNumber(container.Status)).color
              };width:10px;height:10px;`"
            ></div>
          </template>
          <span>{{ container.Status.toLowerCase() }}</span>
        </n-tooltip>
      </div>
    </template>
  </data-table>
  <Modal
    v-model:show="warningsShow"
    :title="t('objects.warnings') + ': ' + warningsName"
    :touched="false"
    style="width: 1000px"
  >
    <n-card
      v-for="key in Object.keys(warnings)"
      :key="key"
      :title="key"
      size="small"
    >
      <n-table :bordered="false" :single-line="false">
        <thead>
          <tr>
            <th>{{ t("fields.message") }}</th>
            <th>{{ t("fields.reason") }}</th>
            <th>{{ t("fields.restartCount") }}</th>
            <th>{{ t("fields.exitCode") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in warnings[key]" :key="row.Message">
            <td>{{ row.Message }}</td>
            <td>{{ row.Reason }}</td>
            <td>{{ row.RestartCount }}</td>
            <td>{{ row.ExitCode }}</td>
          </tr>
        </tbody>
      </n-table>
    </n-card>
  </Modal>
  <Modal
    v-model:show="openDialog"
    style="width: 20rem"
    :title="
      currentDialogComponent
        ? t(`elements.options.${currentDialogComponent}`)
        : ''
    "
    :touched="false"
    :showFooter="true"
    @negative-click="openDialog = false"
    @positive-click="positiveClick"
  >
    <div>
      {{ t("questions.sure") }}
    </div>
  </Modal>
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
