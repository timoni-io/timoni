<script setup lang="ts">
// imports
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useI18n } from "vue-i18n";
import { renderIcon } from "@/utils/renderIcon";
import { DropdownOption } from "naive-ui";
import { useRoute } from "vue-router";
import { useEnv } from "@/store/envStore";
import { api } from "@/zodios/api";
// import ContainersList from "@/components/containers/ContainersList.vue";
import { ElementMapRespExtended } from "@/zodios/schemas/elements";
import { z } from "zod";
import { useMessage } from "naive-ui";
// import ElementScale from "@/components/elements/ElementScale.vue";
import { Input } from "../../zodios/schemas/inputs";
// import { EnvModRequest } from "../../zodios/schemas/envMod";
// import ContainersList from "@/components/containers/ContainersList.vue";
import { filterXSS } from "xss";
import { envIcon } from "@/utils/iconFactory";
// import { containerResponse } from "@/zodios/schemas/containers";
import type { EnvElement } from "@/store/envStore";
import { getContainerStatus } from "@/utils/getContainerStatus";
import router from "@/router";
import { useUserStore } from "@/store/userStore";

export type ElementInput = z.infer<typeof Input>;
export interface ElementInputs {
  [k: string]: ElementInput;
}

const userStore = useUserStore();
defineProps<{ logsSize: number }>();

// type containerRes = z.infer<typeof containerResponse>[number];
const message = useMessage();

const { t } = useI18n();
const route = useRoute();
// const envID = route.params.id as string;
// const envStore = useEnvStore();
const env = useEnv(computed(() => z.string().parse(route.params.id)));

type Warning = {
  Message: string;
  ExitCode: number;
  Reason: string;
  RestartCount: number;
};

let warningsShow = $ref(false);
let warnings = $ref({} as Record<string, Warning[]>);
let warningsName = $ref("");
let actionsShow = $ref(false);
let elementName = $ref("");
let actionName = $ref("");
let editShow = $ref(false);
let fromScratchTouched = $ref(false);
let autoUpdate = $ref<boolean>(false);
let elementToDelete = $ref("");
let reload = $ref<string>();
let noBranchModalShow = $ref<boolean>(false);

watch(
  () => env.value,
  (_, actual) => {
    if (actual?.EnvInfo?.Env.Schedule.OnCrons !== reload) {
      reload = actual?.EnvInfo?.Env.Schedule.OnCrons?.toString();
    }
  }
);
// columns
const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: "",
    template: "status",
    width: "2%",
    minWidth: "3rem",
  },
  {
    title: t("fields.name"),
    template: "name",
    width: "20%",
  },
  {
    title: t("fields.type"),
    template: "type",
    width: "9%",
    minWidth: "8rem",
  },
  // {
  //   title: t("fields.project"),
  //   key: "SourceGit.ProjectName",
  //   width: "10%",
  // },
  {
    title: t("fields.version"),
    template: "version",
    width: "21%",
    // render(row) {
    //   return h("p", "asd");
    // },
  },
  // {
  //   title: t("fields.message"),
  //   template: "message",
  //   key: "StatusMessage",
  //   width: "17%",
  // },
  {
    title: t("fields.updateMode"),
    template: "update",
    width: "6%",
    minWidth: "8rem",
  },
  {
    title: t("fields.cpu"),
    template: "cpu",
    width: "6%",
    className: "cpu",
    minWidth: "6rem",
  },
  {
    title: "RAM [MB]",
    template: "ram",
    width: "6%",
    className: "ram",
    minWidth: "6rem",
  },
  {
    title: t("objects.pod", 2),
    template: "containers",
    width: "8%",
    className: "ram",
    minWidth: "8rem",
  },
  {
    title: "",
    template: "actions",
    width: "20%",
  },
]);

// type ElementData = {
//   Name: string;
//   Status: number;
//   Type: string;
//   Version: string;
// };

// options
const options = (
  element: z.infer<typeof ElementMapRespExtended>
): DropdownOption[] => {
  const type = element.Info.Type;
  // const name = element.Info.Name;
  const stopped = element.Info.Stopped;
  // const repoName = element.Info.SourceGit.RepoName;
  const Unschedulable = element.Info.Unschedulable;
  const pods = element.Status.PodCount;
  return [
    ...(type !== "domain" && type !== "config"
      ? [
        {
          label: t("elements.options.rebuild"),
          key: "rebuild",
          icon: renderIcon(mdiWrenchOutline),
        },
        ...(!stopped
          ? [
            {
              label: t("elements.options.restart"),
              key: "restart",
              icon: renderIcon(mdiRestart),
            },
          ]
          : []),
      ]
      : []),
    ...(type !== "config"
      ? [
        {
          label: t("objects.logs"),
          key: "logs",
          icon: renderIcon(mdiHistory),
        },
      ]
      : []),
    ...(stopped
      ? [
        ...(type !== "config"
          ? [
            {
              label: "start",
              key: "start",
              icon: renderIcon(mdiPlay),
            },
          ]
          : []),
      ]
      : [
        ...(type === "domain" || pods
          ? [
            {
              label: "stop",
              key: "stop",
              icon: renderIcon(mdiPause),
            },
          ]
          : []),
      ]),
    ...(Unschedulable
      ? [
        ...(type !== "config" &&
          env.value.EnvInfo?.Env.Schedule.OnCrons?.length
          ? [
            {
              label: t("actions.enableSchedule"),
              key: "enable",
              icon: renderIcon(mdiLockOpenVariant),
            },
          ]
          : []),
      ]
      : [
        ...(type !== "config" &&
          env.value.EnvInfo?.Env.Schedule.OnCrons?.length
          ? [
            {
              label: t("actions.disableSchedule"),
              key: "disable",
              icon: renderIcon(mdiLock),
            },
          ]
          : []),
      ]),
    {
      label: t("elements.options.delete"),
      key: "delete",
      icon: renderIcon(mdiDelete),
    },
  ];
};

type OptionKey =
  | "git"
  | "update"
  | "rebuild"
  | "logs"
  | "restart"
  | "version"
  | "history"
  | "manifest"
  | "containers"
  | "docker-file"
  | "createDebugElement"
  | "scale"
  | "delete"
  | "start"
  | "stop"
  | "enable"
  | "disable";

type tagStatus =
  | "default"
  | "error"
  | "primary"
  | "info"
  | "success"
  | "warning"
  | undefined;

const tagStatus = (status: number): tagStatus => {
  switch (status) {
    case 0:
    case 1:
    case 5:
      return "warning";
    case 3:
      return "success";
    case 2:
    case 7:
      return "error";
    case 4:
      return "info";
    case 6:
    default:
      return "default";
  }
};

const rowClassName = (row: any): string => {
  let cpuV = row.Info.CPUUsageAvgCores;
  let classV = "";
  if (cpuV >= 50 && cpuV < 75) classV += "warning-cpu";
  else if (cpuV >= 75) classV += "alert-cpu";

  return classV;
};

const fromScratchEdited = () => {
  editShow = false;
  env.value.refetch();
}

// select action
let currentDialogComponent = $ref<OptionKey | null>(null);
let currentElement = $ref<EnvElement[string]>({} as EnvElement[string]);
let currentElementName = $ref("");
let openDialog = $ref<boolean>(false);
let result = $ref("");
let cleanResult = $ref("");
// let elementInputs = $ref<ElementInputs>({});

const escapeHtml = (unsafe: string) => {
  return unsafe
    .replaceAll("&amp;", "&")
    .replaceAll("&#38;", "&")
    .replaceAll("&lt;", "<")
    .replaceAll("&#60;", "<")
    .replaceAll("&gt;", ">")
    .replaceAll("&#62;", ">")
    .replaceAll("&quot;", '"')
    .replaceAll("&#34;", '"')
    .replaceAll("&apos;", "'")
    .replaceAll("&#39;", "'")
    .replaceAll("\\n", "\n");
};
const elementControl = async (
  element: z.infer<typeof ElementMapRespExtended>,
  control: number,
  text: string
) => {
  api
    .post("/env-element-run-control", {
      EnvID: route.params.id as string,
      Element: element.Info.Name,
      Control: control,
    })
    .then((res) => {
      if (res === "ok") message.success(text);
    });
};
const handleSelect = (
  key: OptionKey,
  element: z.infer<typeof ElementMapRespExtended>
) => {
  result = "";
  currentElementName = element.Info.Name;
  currentElement = element;
  currentDialogComponent = key;
  autoUpdate = element.Info.AutoUpdate as boolean;
  switch (currentDialogComponent) {
    case "logs":
      router.push(
        `/env/history/${route.params.id}?element=${currentElementName}`
      );
      break;
    case "start":
      elementControl(element, 1, t("messages.elementStarted"));
      break;
    case "stop":
      elementControl(element, 2, t("messages.elementStopped"));
      break;
    case "enable":
      elementControl(element, 3, t("messages.elementEnabled"));
      break;
    case "disable":
      elementControl(element, 4, t("messages.elementDisabled"));
      break;
    case "manifest":
    case "docker-file":
      api
        .post(
          currentDialogComponent === "manifest"
            ? "/env-element-manifest"
            : "/env-element-docker-file",
          {
            Env: route.params.id as string,
            Elements: [currentElementName],
          }
        )
        .then((res) => {
          result = res[currentElementName];
          cleanResult = escapeHtml(filterXSS(result));
          openDialog = true;
        })
        .catch(() => {
          message.success(currentDialogComponent + " loading failed");
          openDialog = false;
        });
      break;
    // case "input":
    //   api
    //     .post("/env-element-inputs", {
    //       Env: route.params.id as string,
    //       Elements: [currentElementName],
    //     })
    //     .then((res) => {
    //       elementInputs = res[currentElementName];
    //       openDialog = true;
    //     })
    //     .catch(() => {
    //       message.success(currentDialogComponent + " loading failed");
    //       openDialog = false;
    //     });
    //   break;
    default:
      openDialog = true;
  }
};

const handleRepoPush = (row: any) => {
  api.get("/git-repo-branch-list", {queries: {
    name: row.Info.SourceGit.RepoName,
    level: 1
  }}).then((res) => {
    if(res.includes(row.Info.SourceGit.BranchName)) {
      router.push(
        '/code/' +
        row.Info.SourceGit.RepoName +
        '/' +
        row.Info.SourceGit.BranchName +
        '/files' +
        row.Info.SourceGit.FilePath
      );
    } else {
      noBranchModalShow = true;
    }
  })
}

// is debug element
const isDebugElement = () => {
  return currentElement
    ? currentElement.Info.Name.indexOf("-debug") >= 0
    : null;
};

// const updateInput = (res: string[]) => {
//   elementInputs[res[0]].Value = "" + res[1];
// };

// positive click
const positiveClick = () => {
  switch (currentDialogComponent) {
    case "delete":
      api
        .get(
          "/env-element-delete",
          {
            queries: {
              env: route.params.id as string,
              element: currentElementName as string,
            },
          }
          //   DeleteElements: [currentElementName],
          //   Elements: {},
          //   Apply: true,
          // },
          // {
          //   queries: {
          //     env: route.params.id as string,
          //   },
          // }
        )
        .then((res) => {
          openDialog = false;
          if (res === "ok") {
            message.info(t("messages.willBeDeleted"));
            elementToDelete = currentElementName;
          } else message.error(res);
        })
        .catch(() => {
          openDialog = false;
        });
      break;
    case "createDebugElement":
      api
        .get("/env-element-debug-update", {
          queries: {
            env: route.params.id as string,
            element: currentElementName as string,
            state: isDebugElement() ? "off" : "on",
          },
        })
        .then(() => {
          openDialog = false;
          message.success(t("messages.debugElementCreated"));
        })
        .catch(() => {
          openDialog = false;
        });
      break;
    case "rebuild":
      api
        .get("/image-rebuild", {
          queries: {
            envID: route.params.id as string,
            imageID:
              currentElement &&
                currentElement.Info &&
                "Build" in currentElement.Info
                ? currentElement.Info.Build.ImageID
                : "",
          },
        })
        .then((res) => {
          openDialog = false;
          if (res.Message) message.success(t("messages.elementRebuilded"));
          else message.error(res);
        });
      break;
    case "restart":
      api
        .get("/env-element-restart-pods", {
          queries: {
            env: route.params.id as string,
            element: currentElement.Info!.Name,
          },
        })
        .then((res) => {
          if (res === "ok") message.success(t("messages.elementRestarted"));
          else message.error(res);
          openDialog = false;
        });
      break;
    case "update":
      api
        .get("/env-element-update-mode-set", {
          queries: {
            env: route.params.id as string,
            element: currentElement.Info!.Name,
            mode: autoUpdate ? "auto" : "manual",
          },
        })
        .then((res) => {
          if (res === "ok") message.success(t("messages.updateModeChanged"));
          else message.error(res);
          openDialog = false;
        });
      break;
  }
};

const runAction = (element: string, action: string) => {
  api
    .get("/env-element-actions-run", {
      queries: {
        env: route.params.id as string,
        element: element,
        action: action,
      },
    })
    .then((res) => {
      if (res === null) message.success(t("fields.started"));
      else message.error(res);
      actionsShow = false;
    });
};

const elements = computed(() => {
  return Object.values(env.value?.EnvElements || {});
});

const openWeb = (url: string) => {
  window.open("http://" + url, "_blank");
};
// let messageWidth = $ref(0);
// let timeout: NodeJS.Timeout;

// const setMessageWidth = () => {
//   if (timeout) clearTimeout(timeout);
//   timeout = setTimeout(() => {
//     messageWidth =
//       Math.floor(document.documentElement.clientWidth / 150) * 2 + 0;
//   });
// };

// onBeforeMount(() => {
//   setMessageWidth();
//   window.addEventListener("resize", setMessageWidth);
// });

// onBeforeUnmount(() => {
//   window.removeEventListener("resize", setMessageWidth);
// });
const dataTableContainer = ref(null);
const { height } = useElementSize(dataTableContainer);

watch(
  () => env.value.EnvElements,
  () => {
    if (
      elementToDelete &&
      !Object.keys(env.value.EnvElements || {}).includes(elementToDelete)
    ) {
      message.success(t("messages.elementDeleted") + ": " + elementToDelete);
      elementToDelete = "";
    }
  }
);

// window width
let screenWidth = $ref<number>(0);
// let screenTimeout: NodeJS.Timeout;

onBeforeMount(() => {
  screenWidth = window.innerWidth;
});

onMounted(() => {
  window.addEventListener("resize", () => {
    screenWidth = window.innerWidth;
    // if (screenTimeout) clearTimeout(screenTimeout);
    // screenTimeout = setTimeout(() => {
    //   screenWidth = window.innerWidth;
    // }, 100);
  });
});
const openTerminal = (el: z.infer<typeof ElementMapRespExtended>) => {
  window.open(
    `/cli?env=${route.params.id}&element=${el.Info.Name}`,
    "_blank"
  );
};
</script>

<template>
  <div ref="dataTableContainer" style="height: calc(100% - 2rem); position: relative">
    <!-- <Spinner :data="elements"> -->
    <data-table :columns="columns" :data="elements || undefined" :row-class-name="rowClassName"
      :max-height="`calc(${height}px - 1rem)`" style="height: 100%; max-height: 100%">

      <template #projectBranch="row">
        {{ row.Info.SourceGit.RepoName }} / {{ row.Info.SourceGit.BranchName }}
      </template>

      <template #cpu="row">
        <p v-if="row.Status.PodCount">
          {{ (row.Info.CPUUsageAvgCores / 100).toFixed(2) }}
        </p>
      </template>

      <template #ram="row">
        <p v-if="row.Status.PodCount">
          {{ Math.ceil(row.Info.RAMUsageAvgMB) }}
        </p>
      </template>

      <template #name="row">
        <div style="align-items: center; gap: 0.4rem" :class="getContainerStatus(row.Status.State).label">
          
            {{ row.Info.Name }}<br>

            <div v-if="row.Status.Next && !row.Info.ToDelete" style="font-size: 0.8em">
              {{ row.Status.Next.StepCurrent }}/{{ row.Status.Next.StepCount }}
              {{ row.Status.Next.Message }}
            </div>

        </div>
      </template>

      <template #status="element">
        <n-space vertical>
          <div style="
                            display: flex;
                            align-items: center;
                            position: relative;
                            z-index: 9;
                          ">
            <n-tooltip v-if="
              element.Status.Alerts &&
              element.Status.Alerts.length &&
              element.Status.State !== 1
            ">
              <template #trigger>
                <!-- <Mdi
                                :path="mdiAlertCircle"
                                width="20"
                                class="pulse"
                                style="color: tomato"
                              /> -->
                <ElementStatusIcon :cron="
                  !element.Info.Unschedulable &&
                    (element.Info.Type === 'domain' ||
                      element.Info.Type === 'pod') &&
                    env.EnvInfo?.Env.Schedule.OnCrons?.length
                    ? true
                    : false
                " :status="element.Status.State" :toDelete="element.Info.ToDelete" :iconFunc="envIcon" class="pulse" />
              </template>

              <ul style="list-style-type: none">
                <li v-for="el in element.Status.Alerts" :key="el">
                  {{ el }}
                </li>
              </ul>
            </n-tooltip>
            <ElementStatusIcon v-else :cron="
              !element.Info.Unschedulable &&
                (element.Info.Type === 'pod' || element.Info.Type === 'domain') &&
                env.EnvInfo?.Env.Schedule.OnCrons?.length
                ? true
                : false
            " :status="element.Status.State" :toDelete="element.Info.ToDelete" :iconFunc="envIcon" />
          </div>
          <!-- <div v-if="element.Status.Next && !element.Info.ToDelete" style="
                            display: flex;
                            align-items: center;
                            position: relative;
                            z-index: 9;
                          ">
              <ElementStatusIcon :status="element.Status.Next.State" :iconFunc="envIcon" />
            </div> -->
        </n-space>
      </template>
      <template #version="row">
        <div style="width: 100%; display: flex; align-items: center">
          <n-button v-if="row.Info.SourceGit.RepoName" :disabled="
            !userStore.havePermission('Env_ElementFullManage') ||
            !userStore.havePermission('Env_ElementVersionChangeOnly')
          " class="version" secondary size="tiny" type="primary" style="width: 7em"
            @click="() => handleSelect('version', row)">
            <n-icon size="14px">
              <mdi :path="mdiSwapVerticalBold" />
            </n-icon>
            <span>{{ row.Info.SourceGit.CommitHash.substring(0, 8) }}</span>
          </n-button>

          <n-button v-if="row.Info.SourceGit.RepoName" :disabled="!userStore.havePermission('Repo_View')" secondary
            size="tiny" type="primary" @click="handleRepoPush(row)" style="margin-left: 0.5em; width: 100px; overflow: hidden">
            <n-icon size="14px" style="margin-right: 4px">
              <mdi :path="mdiGit" />
            </n-icon>
            <p style="text-align: left">
              {{ row.Info.SourceGit.RepoName }}<br />
              {{ row.Info.SourceGit.BranchName }}
            </p>
          </n-button>

          <n-icon v-if="row.Status.NewerVersion" size="14px" style="margin-right: 4px">
            <mdi :path="mdiArrowUpBoldBox" />
          </n-icon>

        </div>
        <!-- <router-link
                            v-if="row.SourceGit.RepoName"
                            :to="
                              '/code/' +
                              row.SourceGit.RepoName +
                              '/' +
                              row.SourceGit.BranchName +
                              '/commits'
                            "
                            >{{ row.SourceGit.RepoName }} / {{ row.SourceGit.BranchName }} /
                            {{ row.SourceGit.CommitHash.substring(0, 8) }}</router-link
                          > -->
      </template>
      <template #update="row">
        <n-button secondary type="primary" size="tiny" :disabled="
          !userStore.havePermission('Env_ElementFullManage') ||
          !row.Info.SourceGit.RepoName
        " @click="
  () => {
    handleSelect('update', row);
  }
">
          <span>
            {{
              row.Info.AutoUpdate
              ? t("mode.auto").toLowerCase()
              : t("mode.manual").toLowerCase()
            }}
          </span>
        </n-button>
      </template>
      <template #type="row">
        {{ row.Info.Type }}
      </template>
      <template #actions="row">
        <div v-if="!row.Info.ToDelete" style="display: flex; justify-content: end">
          <PopModal v-if="row.Info.Actions" :title="t('fields.actions')" width="auto">
            <template #trigger>
              <n-button secondary type="primary" size="tiny" class="change-version" style="margin-right: 0.5em">
                <n-icon size="13px">
                  <mdi :path="mdiCog" />
                </n-icon>
                {{ t("fields.actions") }}
              </n-button>
            </template>
            <template #content>
              <div v-for="action in Object.keys(row.Info.Actions)" :key="action" style="
                                  display: flex;
                                  justify-content: space-between;
                                  align-items: center;
                                  margin-top: 5px;
                                ">
                <n-tag v-if="row.Info.Actions[action]" size="tiny" round type="warning" style="margin: 0 5px 0 5px">
                  {{ row.Info.Actions[action].join(" ") }}
                </n-tag>
                <n-button strong secondary circle size="tiny" type="primary" @click="
                  () => {
                    actionsShow = true;
                    actionName = action;
                    elementName = row.Info.Name;
                  }
                ">
                  <template #icon>
                    <n-icon size="22px">
                      <mdi :path="mdiMenuRight" />
                    </n-icon>
                  </template>
                </n-button>
              </div>
            </template>
          </PopModal>

          <n-tooltip v-if="row.Info.Domain" trigger="hover" :disabled="screenWidth > 1400">
            <template #trigger>
              <n-button secondary type="primary" size="tiny" style="margin-right: 0.5em"
                @click="openWeb(row.Info.Domain)">
                <n-icon style="cursor: pointer">
                  <mdi :path="mdiDomain" width="15" />
                </n-icon>

                <span style="margin-left: 5px" v-if="screenWidth > 1400">{{ row.Info.Domain.slice(0, 25)
                }}{{ row.Info.Domain.length > 25 ? "..." : "" }}</span>
              </n-button>
            </template>
            {{ row.Info.Domain.slice(0, 25)
            }}{{ row.Info.Domain.length > 25 ? "..." : "" }}
          </n-tooltip>

          <n-tooltip v-if="row.Status.PodCount" trigger="hover" :disabled="screenWidth > 1400">
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
                <span style="margin-left: 5px" v-if="screenWidth > 1400">Terminal</span>
              </n-button>
            </template>
            Terminal
          </n-tooltip>
          <n-button v-if="
            !row.Info.Name.includes('-debug') &&
            row.Info.Name !== 'auto-domain'
          " secondary type="primary" size="tiny" class="change-version"
            :disabled="!userStore.havePermission('Env_ElementFullManage')" @click="
              () => {
                currentElement = row;
                editShow = true;
              }
            ">
            <!-- dodać edycję jak z dodawania from scratch -->
            <n-icon size="14px">
              <mdi :path="mdiPencil" />
            </n-icon>
            {{ t("actions.edit") }}
          </n-button>
          <n-dropdown trigger="click" :options="options(row)"
            :disabled="!userStore.havePermission('Env_ElementFullManage')" placement="bottom-end"
            @select="handleSelect($event, row)" :key="reload">
            <div class="dropdown-menu">
              <n-button :disabled="!userStore.havePermission('Env_ElementFullManage')" secondary type="primary"
                size="tiny" class="change-version">
                <n-icon size="22px">
                  <mdi :path="mdiDotsHorizontal" />
                </n-icon>
              </n-button>
            </div>
          </n-dropdown>
        </div>
      </template>
      <template #containers="row">
        <n-button secondary type="primary" size="tiny" v-if="row.Info.Type === 'pod' && !row.Info.Name.includes('-debug')"
          :disabled="
            !userStore.havePermission('Env_ElementFullManage') ||
            row.Info.Stopped ||
            row.Status.State === 2
          " @click="
  () => {
    handleSelect('scale', row);
  }
">
          <span>
            {{
              row.Info.Stopped
              ? t("fields.stopped")
              : row.Status.PodCount !== row.Info.Scale.NrOfPodsMin
                ? row.Status.PodCount + " -> " + row.Info.Scale.NrOfPodsMin
                : row.Info.Scale.NrOfPodsCurrent
            }}

            <span v-if="row.Info.Scale.NrOfPodsMin !== row.Info.Scale.NrOfPodsMax" style="font-size:.8em">
              ({{ row.Info.Scale.NrOfPodsMin }}-{{ row.Info.Scale.NrOfPodsMax }})
            </span>

          </span>
        </n-button>
      </template>
    </data-table>
    <!-- </Spinner> -->
    <Modal v-model:show="openDialog" :style="
      currentDialogComponent === 'version'
        ? 'width:1000px'
        : 'max-width:40rem'
    " :title="
  currentDialogComponent
    ? t(`elements.options.${currentDialogComponent}`)
    : ''
" :touched="false" :showFooter="
  ['delete', 'rebuild', 'restart', 'createDebugElement', 'update'].includes(currentDialogComponent as string)
" @negative-click="openDialog = false" @positive-click="positiveClick">
      <ElementChangeVersion
        v-if="currentDialogComponent === 'version'"
        :elementName="currentElementName"
        :element="currentElement"
        @closeDialog="openDialog = false"
      />
      <div v-if="currentDialogComponent === 'delete'">
        <p>
          {{
            t("messages.deletingElement") +
            ' "' +
            currentElementName +
            '". ' +
            t("questions.sure")
          }}
        </p>
      </div>
      <div v-if="
        ['rebuild', 'restart', 'createDebugElement'].includes(currentDialogComponent as string)
      ">
        {{ t("questions.sure") }}
      </div>
      <div v-if="currentDialogComponent === 'scale'">
        <ElementScale :scale="
          currentElement &&
            currentElement.Info &&
            'Scale' in currentElement.Info
            ? currentElement.Info.Scale
            : undefined
        " :disable="false" :elementName="currentElementName" @closeDialog="openDialog = false"
          style="display: flex; flex-direction: column" />
      </div>
      <div v-if="['manifest', 'docker-file'].includes(currentDialogComponent as string)">
        <!-- <pre
                          style="overflow-wrap: break-word; white-space: pre-line"
                          v-html="cleanResult"
                        ></pre> -->
        <div style="height: 30rem">
          <Monaco :value="cleanResult" :lang="
            currentDialogComponent === 'manifest' ? 'toml' : 'dockerfile'
          " read-only />
        </div>
      </div>
      <div v-if="currentDialogComponent === 'update'">
        <div class="update-grid">
          <div class="grid-left">
            <n-button class="update-mode-button" :type="autoUpdate ? 'primary' : 'default'" @click="
              () => {
                autoUpdate = true;
              }
            ">
              {{ t("mode.auto") }}
            </n-button>
            <div class="grid-right">{{ t("scratch.docs.auto") }}</div>
          </div>
        </div>
        <div class="update-grid">
          <div class="grid-left">
            <n-button class="update-mode-button" :type="!autoUpdate ? 'primary' : 'default'" @click="
              () => {
                autoUpdate = false;
              }
            ">
              {{ t("mode.manual") }}
            </n-button>
            <div class="grid-right">{{ t("scratch.docs.manual") }}</div>
          </div>
        </div>
      </div>
      <!-- <div v-if="currentDialogComponent === 'input'">
                        <ElementInput
                          v-for="(input, k) in elementInputs"
                          @updateInput="updateInput"
                          :key="k"
                          :name="(k as string)"
                          :input="(input as ElementInput)"
                          :edit="true"
                          style="margin-top: 1em"
                        />
                      </div> -->
    </Modal>
    <Modal v-model:show="warningsShow" :title="t('objects.warnings') + ': ' + warningsName" :touched="false"
      style="width: 1000px">
      <n-card v-for="key in Object.keys(warnings)" :key="key" :title="key" size="small">
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
    <Modal v-model:show="actionsShow" :title="t('fields.actions') + ': ' + actionName" :touched="false"
      style="width: 20rem" :show-footer="true" @positive-click="runAction(elementName, actionName)">
      {{
        t("questions.runAction1") +
        actionName +
        t("questions.runAction2") +
        elementName +
        "'?"
      }}
    </Modal>
    <Modal v-model:show="editShow" :title="
      t('actions.edit') +
      ': ' +
      currentElement?.Info?.Name +
      ' (' +
      currentElement?.Info?.Type +
      ')'
    " :touched="false" style="width: 80rem" :show-footer="false">
      <ElementAddFromScratch @fromScratchCreated="fromScratchEdited" @isTouched="fromScratchTouched = true"
        @detouch="fromScratchTouched = false" :selectedTab="
          currentElement.Info.Type === 'config' ? 'variables' : 'general'
        " :element="currentElement" />
    </Modal>
    <Modal
      v-model:show="noBranchModalShow"
      :title="t('messages.errorOccured')"
      :touched="false"
      style="width: auto; min-width: 30em;"
    >
      {{ t('messages.branchDeleted') }}
    </Modal>
  </div>
</template>
<style>
.pulse {
  animation: pulse-animation 1s infinite;
  border-radius: 50%;
  box-shadow: 0px 0px 1px 1px #0000001a;
}

@keyframes pulse-animation {
  0% {
    box-shadow: 0 0 0 0px rgba(201, 0, 0, 0.5);
  }

  100% {
    box-shadow: 0 0 0 5px rgba(184, 0, 0, 0.2);
  }
}

.change-version {
  margin-right: 0.5em;
  padding: 0 5px;
}

.change-version span {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
}
</style>
<style scoped lang="scss">
.dropdown-menu {
  display: flex;
  align-items: center;
  justify-content: right;
  /* margin-right: 9px; */
  cursor: pointer;
  opacity: 0.6;
  transition: 0.2s ease-in-out;
}

.dropdown-menu:hover {
  opacity: 1;
}

:deep(.warning-cpu .cpu) {
  color: var(--warningColor) !important;
}

:deep(.alert-cpu .cpu) {
  color: var(--errorColor) !important;
}

:deep(.warning-ram .ram) {
  color: var(--warningColor) !important;
}

:deep(.alert-ram .ram) {
  color: var(--errorColor) !important;
}

:deep(.ram) {
  text-align: center !important;
}

:deep(.cpu) {
  text-align: center !important;
}

.update-grid {
  display: flex;
  padding-left: 1em;
}

.grid-left {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 8em;
  margin-left: 1em;
}

.grid-right {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 8em;
  margin-left: 2em;
}

.update-mode-button {
  width: 9em;
  height: 6em;
}

.update-mode-button:hover {
  background-color: var(--n-color) !important;
}

.update-mode-button:focus {
  background-color: var(--n-color) !important;
}

.version {
  & :deep(.n-button__content) {
    justify-content: space-between;
    width: inherit;
  }
}
</style>
