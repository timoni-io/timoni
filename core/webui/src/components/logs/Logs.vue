<script setup lang="ts">
import { useApi } from "@/next-api";
import { ILog, useLogs } from "@/store/logsStore";
// import { format } from "date-fns";
import { useRoute } from "vue-router";
// import moment from "moment";
import { useTimeFormatter } from "@/utils/formatTime";
// @ts-ignore
import InfiniteLoading from "v3-infinite-loading";
import "v3-infinite-loading/lib/style.css";
import iconFactory from "@/utils/iconFactory";
import { useEnv } from "@/store/envStore";
import type { EnvElement } from "@/store/envStore";
// import { useEnvStore } from "@/store/envStore";

import { useUserStore } from "@/store/userStore";
const userStore = useUserStore();

const { dateFranekFormat } = useTimeFormatter();
// import { format } from "date-fns";

// const envStore = useEnvStore();
const env = useEnv(computed(() => z.string().parse(route.params.id)));

let elements = $ref<[EnvElement[string]] | []>([]);

watch(env, (env, oldEnv) => {
  if (!elements.length) {
    elements = Object.values(env?.EnvElements || {}) as [EnvElement[string]];
    return;
  }
  if (
    JSON.stringify(
      Object.values(env.EnvElements || {}).map((element) => {
        if (element.Info.Type === "pod") return element.Info.Build?.ImageID;
      })
    ).localeCompare(
      JSON.stringify(
        Object.values(oldEnv.EnvElements || {}).map((element) => {
          if (element.Info.Type === "pod") return element.Info.Build?.ImageID;
        })
      )
    )
  ) {
    refresh();
    elements = Object.values(env?.EnvElements || {}) as [EnvElement[string]];
  }
  // if (
  //   JSON.stringify(oldEnv.EnvElements) !== JSON.stringify(env.EnvElements) ||
  //   !elements.length
  // ) {
  // }
});

interface Log {
  time: string;
  id: string;
  level: string;
  message: string;
  element: string;
  pod: string;
  env: string;
  version: string;
  user: string;
  project: string;
  details_string: string;
  details_number: number;
}

interface LogMode {
  value: string;
  label: string;
  disabled?: boolean;
  icon?: string;
}
useApi();

const { t } = useI18n();
const route = useRoute();

const props = defineProps<{ fullscreen: boolean; backgroundColor: string }>();
const emit = defineEmits<{
  (type: "update:fullscreen", event: boolean): void;
  (type: "minimize"): void;
}>();

let renderKey = $ref(0);
watch(
  () => props.backgroundColor,
  () => {
    renderKey += 1;
  }
);

let fullscreen = $(useVModel(props, "fullscreen", emit));

// onMounted(() => {
//   logs.liveArgs = {
//     EnvID: route.params.id as string,
//     Limit: 50,
//   };
// });
let tabValue = $ref<"runtime" | "build" | "all" | "events">(
  (() => {
    let arr = [
      { tab: "events", perm: "Env_ViewLogsEvents" },
      { tab: "runtime", perm: "Env_ViewLogsRuntime" },
      { tab: "build", perm: "Env_ViewLogsBuild" },
    ];
    if (
      !arr.reduce((acc, cur) => {
        if (!userStore.havePermission(cur.perm as any)) {
          acc.push(cur);
          return acc;
        }
        return acc;
      }, [] as any).length
    )
      return "all";
    if (localStorage.getItem("logs-default")) {
      if (localStorage.getItem("logs-default") === "all") {
        return "all";
      }
      if (
        userStore.havePermission(
          arr.find((el) => el.tab === localStorage.getItem("logs-default"))!
            .perm as any
        )
      ) {
        return localStorage.getItem("logs-default") as
          | "runtime"
          | "build"
          | "all"
          | "events";
      } else {
        return arr.find((el) => {
          return userStore.havePermission(el.perm as any);
        })!.tab as "runtime" | "build" | "all" | "events";
      }
    }
    return arr.find((el) => {
      return userStore.havePermission(el.perm as any);
    })!.tab as "runtime" | "build" | "all" | "events";
  })()
);
watch($$(tabValue), () => {
  localStorage.setItem("logs-default", tabValue);
});

const envId = $computed(() => {
  return {
    runtime: [{ id: route.params.id as string }],
    build: [
      {
        id: "image_builder",
        imageIdList: elements
          .map((el) => {
            if ("Build" in el.Info) {
              return el.Info?.Build?.ImageID;
            }
          })
          .filter((el) => el),
      },
    ],
    events: [{ id: route.params.id as string, events: true }],
    all: [
      { id: route.params.id as string },
      {
        id: "image_builder",
        imageIdList: elements
          .map((el) => {
            if ("Build" in el.Info) {
              return el.Info?.Build?.ImageID;
            }
          })
          .filter((el) => el),
      },
      {
        id: route.params.id as string,
        events: true,
      },
    ],
  }[tabValue as string] as { id: string; imageIdList?: string[] }[];
});

let {
  logs,
  startLive,
  loadMore,
  isLive,
  incomingCount,
  isLoading,
  refresh,
  noMoreLogs,
  clear,
} = useLogs($$(envId));

// watch($$(elements), () => {}, { deep: true });

let scrollbar = $ref<HTMLElement | null>(null);

let hover = $ref(false);
let showDetails = $ref(false);
let currentLog = $ref<Log>();
let modes = $ref<LogMode[]>([
  {
    value: "all",
    label: t("fields.all"),
    disabled:
      !userStore.havePermission("Env_ViewLogsBuild") ||
      !userStore.havePermission("Env_ViewLogsEvents") ||
      !userStore.havePermission("Env_ViewLogsRuntime"),
  },
  {
    value: "events",
    label: "Events",
    icon: mdiMessageBadge,
    disabled: !userStore.havePermission("Env_ViewLogsEvents"),
  },
  {
    value: "runtime",
    label: t("fields.runtime"),
    icon: mdiFlash,
    disabled: !userStore.havePermission("Env_ViewLogsRuntime"),
  },
  {
    value: "build",
    label: t("fields.build"),
    icon: mdiWrench,
    disabled: !userStore.havePermission("Env_ViewLogsBuild"),
  },
]);
// userStore.havePermission("Env_ViewLogsBuild");
const detailsShow = (log: Log) => {
  currentLog = log;
  showDetails = true;
};

// const relativeOrDistanceToNowLukaszFormatted = (time: any) => {
//   const timeFormatted = relativeOrDistanceToNow(
//     new Date(Math.floor(time / 1000000))
//   );
//   if (
//     ["mniej niż minuta temu", "less than a minute ago"].includes(timeFormatted)
//   )
//     return t("time.now").toLowerCase();
//   return timeFormatted;
// };

// useIntervalFn(() => {
//   logsStore.fetchLogs();
// }, 2000);

// watchEffect(() => {
//   logs.LogsList = [];
//   logs.liveArgs = {
//     EnvID: [route.params.id as string],
//     // EnvID: ["traefik"],
//     Limit: 50,
//   };
// });

// logs.$subscribe((_, { LogsList }) => {
//   if (!hover) {
//     updateLogList();
//   } else {
//     incomingLogs = LogsList.length - localLogList.length;
//   }
// });

// watch($$(hover), () => {
//   if (!hover) updateLogList();
// });

onMounted(() => {
  startLive();

  window.addEventListener("show-log-details", (e) => {
    const logTime = (e as CustomEvent).detail as string;
    detailsShow(logs.value.find((log) => log.time === logTime)!);
  });
});

watch(
  () => isLive.value,
  () => {
    setTimeout(
      () => scrollbar?.scrollBy({ top: 1000000, behavior: "smooth" }),
      150
    );
  }
);

let logsEl = $ref(null as HTMLElement | null);
const logsElSize = useElementSize($$(logsEl));

const loadOlderLogs = async () => {
  if (!isLive.value) {
    const olderLogs = await loadMore(20);
    nextTick(() => {
      scrollbar?.scrollBy({ top: 21 * olderLogs.length });
    });
  }
};

// onKeyStroke(" ", () => {
//   isLive.value = !isLive.value;
// });

const iconHtml = (level: string) => {
  const i = iconFactory(level);
  return `<svg viewBox="0 0 24 24" style="width: 14px; color: ${i.color}">
  <path fill="currentColor" d="${i.icon}" />
</svg>`;
};

const logHtml = (log: ILog & { Event: boolean }) => {
  return `<tr
  class="${log.level.toLowerCase()}"
  style="height: 21px"
>
  <th style="width: 1rem">
    <button
      class="log-icon-btn"
      onclick="window.dispatchEvent(new CustomEvent('show-log-details', { detail: '${
        log.time
      }' }));"
    >
    ${iconHtml(log.level)}
    </button>
  </th>
  <th style="width: 5rem; white-space: nowrap; word-break: keep-all">
    <time> ${dateFranekFormat(log.time)} </time>
  </th>
  <th style="width: 14rem">
    <span style="white-space: nowrap; word-break: keep-all">
      ${
        log.env_id === "image_builder"
          ? log.tags_string.imageid.split(".")[0]
          : log.Event
          ? "Events"
          : t("fields.runtime")
      }

    </span>
  </th>
  <th style="width: 8rem">
    <span style="color: #6ee7b7; white-space: nowrap; word-break: keep-all"
      >${log.element}</span
    >
  </th>
  <th style="white-space: nowrap; width: 70%">${log.message}</th>
</tr>
`;
};
let logList = $ref<HTMLElement | null>(null);
let incomingCountButFaster = $ref(0);

watchThrottled(
  incomingCount,
  () => {
    incomingCountButFaster = incomingCount.value;
  },
  { throttle: 500 }
);

watchThrottled(
  logs,
  () => {
    if (logList) {
      logList.innerHTML = logs.value.map(logHtml).join("");

      scrollbar?.scrollBy({ top: 1000000, behavior: "smooth" });
    } else {
      until($$(logList))
        .toBeTruthy()
        .then(() => {
          logList!.innerHTML = logs.value.map(logHtml).join("");

          scrollbar?.scrollBy({ top: 1000000, behavior: "smooth" });
        });
    }
  },
  { throttle: 100 }
);
</script>
<template>
  <div
    ref="logsEl"
    class="logs n-card"
    :style="`background-color: ${
      backgroundColor || 'transparent'
    } !important; '--h': ${logsElSize.height} + 'px'`"
    :key="renderKey"
  >
    <PanelHeader :title="t('objects.log', 2)">
      <template #trailing>
        <div style="display: flex; gap: 0.5rem">
          <!-- <n-button quaternary strong size="tiny" @click="isLive = !isLive">
            <template #icon>
              <mdi :path="isLive ? mdiPause : mdiPlay" />
            </template>
          </n-button> -->
          <div
            name="radiobuttongroup1"
            @click="startLive"
            style="max-height: 1.4rem; display: flex; gap: 0.3rem"
          >
            <n-button
              v-for="mode in modes"
              @click="
                tabValue = mode.value as typeof tabValue;
                clear();
              "
              :secondary="tabValue !== mode.value"
              type="info"
              round
              icon-placement="left"
              :key="mode.value"
              :value="mode.value"
              :label="mode.label"
              :disabled="mode.disabled"
              style="max-height: 1.4rem; border: none; outline: none !important"
            >
              <template v-if="mode.icon" #icon>
                <n-icon
                  :size="16"
                  style="transform: translateX(5px); margin-right: 0.5em"
                >
                  <Mdi :path="mode.icon!" />
                </n-icon>
              </template>
              {{ mode.label }}
            </n-button>
          </div>
          <n-button
            quaternary
            strong
            secondary
            type="warning"
            size="tiny"
            @click="emit('minimize')"
          >
            <template #icon>
              <mdi :path="mdiWindowMinimize" />
            </template>
          </n-button>
          <n-button
            quaternary
            strong
            secondary
            type="primary"
            size="tiny"
            @click="fullscreen = !fullscreen"
          >
            <template #icon>
              <mdi :path="fullscreen ? mdiMoveResizeVariant : mdiResize" />
            </template>
          </n-button>
        </div>
      </template>
    </PanelHeader>

    <!-- <n-tabs type="line" size="small" v-model:value="tabValue" animated>
      <n-tab-pane name="work" tab="Praca"> </n-tab-pane>
      <n-tab-pane name="build" tab="Budowanie"> </n-tab-pane>
      <n-tab-pane name="activities" tab="Aktywności"></n-tab-pane>
    </n-tabs> -->
    <!-- <div
      style="display: flex; justify-content: center; margin-bottom: 1em"
    ></div> -->
    <n-scrollbar
      v-if="!isLoading"
      ref="scrollbar"
      class="scrollbar"
      :class="{ live: isLive }"
      style="height: calc(100% - 2.2rem); position: relative"
      trigger="none"
    >
      <div v-if="isLive" class="live-overlay">
        <n-tag
          style="
            position: absolute;
            bottom: 12px;
            right: 4.5rem;
            color: white;
            font-weight: 500;
          "
          :bordered="false"
          round
        >
          {{ t("others.interact") }}
        </n-tag>
      </div>
      <div
        v-if="!isLive && !noMoreLogs"
        style="width: 100%; display: grid; place-content: center"
      >
        <InfiniteLoading @infinite="loadOlderLogs" />
      </div>
      <div
        style="
          width: 100%;
          display: grid;
          place-content: center;
          line-height: 30px;
        "
        v-else
      >
        <p v-if="elements.length && logs.length !== 0">
          {{ t("messages.noMoreLogs") }}
        </p>
      </div>

      <n-empty v-if="logs.length === 0" style="padding-top: 3rem" />

      <table @pointerenter="hover = true" @pointerleave="hover = false">
        <tbody ref="logList"></tbody>
        <div
          style="transition: height 0.2s"
          :style="{ height: isLive ? '0rem' : '3.5rem' }"
        ></div>
        <div class="incoming" :class="{ live: isLive }" v-if="logs.length">
          <n-button
            type="info"
            round
            @click="
              () => {
                noMoreLogs = false;
                isLive = !isLive;
              }
            "
            icon-placement="right"
            @keydown.prevent
          >
            <span v-if="!isLive" style="padding-right: 0.2rem">{{
              t("others.paused")
            }}</span>
            <div v-if="!isLive">
              <span style="padding-right: 0.5rem"
                >• {{ t("others.incomingLogs") }}</span
              >
              <n-badge
                color="black"
                :value="incomingCountButFaster"
                show-zero
              />
            </div>
            <template #icon>
              <n-icon :size="22">
                <Mdi :path="isLive ? mdiPause : mdiPlayCircle" />
              </n-icon>
            </template>
          </n-button>
        </div>
      </table>
    </n-scrollbar>
    <div
      v-else
      style="
        display: flex;
        justify-content: center;
        align-items: center;
        height: inherit;
      "
    >
      <Spinner :data="''" />
    </div>
    <!-- </n-card> -->
    <Modal
      v-model:show="showDetails"
      :title="t('fields.logDetails')"
      style="width: 1200px"
    >
      <LogDetails :log="currentLog" />
    </Modal>
  </div>
</template>

<style>
.scrollbar.live .n-scrollbar-rail {
  display: none;
}
</style>

<style scoped>
.n-card {
  /* background: var(--cardColor); */
  padding: 15px;
  border-radius: 5px;
}
.logs {
  /* display: grid; */
  grid-template-rows: 2rem 1fr;
  width: 100%;
  max-width: 100%;
  height: 100%;
}
:deep(.scrollbar) {
  overflow-y: hidden;
  background-color: #00000056;
  border-radius: 10px;
  margin-top: 0.7rem;
  margin-bottom: 0.5rem;
  box-shadow: inset 0px 11px 24px 0px rgba(32, 32, 32, 1);
}
table {
  text-align: left;
  width: 100%;
  border-spacing: 0 1px;
  /* jeśli komuś nie podoba się spacing w logach to niech zedytuje linijkę powyżej */
}

tbody {
  width: 100%;
}

:deep(tr) {
  font-family: Menlo, Monaco, Consolas, "Liberation Mono", "Courier New",
    monospace;
  color: white;
  font-size: 0.8rem;
}

:deep(tr:hover) {
  background-color: #2c2c2cff;
}

:deep(tr.spaceUnder > td) {
  padding-bottom: 10em;
}

:deep(th) {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

:deep(.error) {
  color: #f87272;
  /* background-color: #470000af; */
}

:deep(.fatal) {
  color: rgb(235, 6, 235);
  /*  background-color: rgba(78, 2, 78, 0.603); */
}

:deep(.warn) {
  color: #fbbd23;
  /* background-color: #382800b0; */
}

time {
  color: #fde68a;
  white-space: nowrap;
}

.incoming {
  display: flex;
  position: absolute;
  right: 0;
  bottom: 0px;
  left: 0;
  padding-bottom: 0.5rem;
  transition: transform 0.2s;
  justify-content: center;
}
.incoming.hide {
  transform: translateY(3rem);
}

.incoming.live {
  transform: translateX(calc(50% - 2rem));
}

:deep(.log-icon-btn) {
  background-color: transparent;
  border: none;
  display: grid;
  place-content: center;
  cursor: pointer;
  border-radius: 2rem;
  transform: scale(1.2);
}

:deep(.log-icon-btn:hover) {
  background-color: #ffffff56;
}

.live-overlay {
  position: absolute;
  inset: 0;
  transition: opacity 0.5s;
  opacity: 0;
}

.live-overlay:hover {
  opacity: 1;
}

/* :deep(.n-tabs-pane-wrapper > * > *) {
  max-height: calc(var(--h) - 2rem);
}

:deep(.n-tabs-nav) {
  padding-left: 0.5rem;
} */
</style>
<style>
.scrollbar.n-scrollbar {
  max-height: calc(var(--h) - 5.5rem);
}
</style>
