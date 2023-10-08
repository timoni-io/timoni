<script setup lang="ts">
import { useEnv } from "@/store/envStore";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import EnvChangeName from "@/components/envs/EnvChangeName.vue";

import { useMessage } from "naive-ui";
// import useErrorMsg from "@/utils/errorMsg";
import { useUserStore } from "@/store/userStore";
import { EnvInfo } from "@/zodios/schemas/env";
import { z } from "zod";

const userStore = useUserStore();
// let { setErrorMsg } = useErrorMsg();

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const message = useMessage();
const env = useEnv(computed(() => route.params.id as string));
let showRemoveModal = $ref(false);
let showRerenderModal = $ref(false);
let timeMetrics = $ref(false);
let localAlerts = $ref<z.infer<typeof EnvInfo>["Alerts"]>(env.value.Alerts);

watch(
  () => env.value.Alerts,
  (actual, prev) => {
    if (JSON.stringify(actual).localeCompare(JSON.stringify(prev))) {
      localAlerts = actual;
    }
  }
);
let time = $ref<number>(
  parseInt(localStorage.getItem("timeSelected") as string) || 15
);
let timeUnit = $ref<string>(localStorage.getItem("timeUnitSelected") || "m");
let timeSelected = $ref<number>(
  parseInt(localStorage.getItem("timeSelected") as string) || 15
);
let timeUnitSelected = $ref<string>(
  localStorage.getItem("timeUnitSelected") || "m"
);
let timeMode = $ref<string>(localStorage.getItem("timeMode") || "relative");
let timerange = $ref<[number, number]>([Date.now(), Date.now()]);
let timerangeSelected = $ref<[number, number]>([
  parseInt(localStorage.getItem("timerangeSelectedFrom") as string) ||
    Date.now(),
  parseInt(localStorage.getItem("timerangeSelectedTo") as string) || Date.now(),
]);
const options = $ref([
  {
    label: "Minutes",
    value: "m",
  },
  {
    label: "Hours",
    value: "h",
  },
]);

// URL - open web
const openWeb = (url: string) => {
  window.open(url, "_blank");
};
// watchEffect(() => {
//   if (env.value?.EnvInfo?.Status === 4) {
//     // router.push("/");
//     setErrorMsg(t("messages.environmentNotFound"));
//   }
// });
const rerenderEnv = () => {
  api
    .get("/env-rerender", {
      queries: {
        env: route.params.id as string,
      },
    })
    .then((res) => {
      if (res === "ok") {
        message.success(t("messages.envRerendered"));
      } else {
        message.error(res);
      }
      showRerenderModal = false;
    });
};

const deleteEnv = () => {
  api
    .get("/env-delete", {
      queries: {
        id: route.params.id as string,
      },
    })
    .then((res) => {
      // if (res === "permission denied") {
      //   message.error(t("messages.permissionDenied"));
      //   return;
      // }
      if (res === "ok") {
        message.info(t("messages.envDeleting"));
        router.push("/");
      } else {
        message.error(res);
      }
    });
};

const timeApply = () => {
  timeSelected = time;
  timeUnitSelected = timeUnit;
  timeMode = "relative";
  localStorage.setItem("timeMode", timeMode);
  localStorage.setItem("timeSelected", "" + timeSelected);
  localStorage.setItem("timeUnitSelected", timeUnitSelected);
  timeMetrics = false;
};

const timerangeApply = () => {
  timerangeSelected = timerange;
  timeMode = "absolute";
  localStorage.setItem("timeMode", timeMode);
  localStorage.setItem("timerangeSelectedFrom", "" + timerangeSelected[0]);
  localStorage.setItem("timerangeSelectedTo", "" + timerangeSelected[1]);
  timeMetrics = false;
};

const dateDisabled = (ts: number) => {
  return new Date(ts).getTime() > Date.now();
};

const elementsCounter = computed(() => {
  return Object.keys(env.value.EnvInfo?.Env.Elements || {}).length;
});

const elementsPercentage = (el: any) => {
  const elements = el.ElementNotReady + el.ElementReady;
  if (elements === 0) return 0;
  return (el.ElementReady / elements) * 100;
};
const podsPercentage = (el: any) => {
  const pods = el.PodNotReady + el.PodReady;
  if (pods === 0) return 0;
  return (el.PodReady / pods) * 100;
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

// const gg1 = `grid-template-areas: "chart-1 chart-2 chart-3 chart-7" "chart-4 chart-4 chart-4 chart-6" "chart-5 chart-5 chart-5 chart-6" "chart-5 chart-5 chart-5 chart-6";`;
const gg2 = `grid-template-areas: "chart-1 chart-1 chart-1 chart-6" "chart-4 chart-4 chart-4 chart-6" "chart-5 chart-5 chart-5 chart-6" "chart-5 chart-5 chart-5 chart-6";`;
</script>

<template>
  <div>
    <EnvTab />
    <div v-if="userStore.havePermission('Env_View')">
      <n-scrollbar
        style="
          height: 100%;
          max-height: calc(100vh - 2.5rem);
          transform: translateY(-1rem);
        "
      >
        <div
          style="padding: 1rem 0"
          class="env-charts-container"
          :style="Object.keys(localAlerts).length ? gg2 : gg2"
        >
          <!-- <n-card
            :title="'Alerts'"
            v-if="Object.keys(localAlerts).length"
            size="small"
            style="grid-area: chart-7; position: relative"
          >
            <n-scrollbar
              style="
                position: absolute;
                margin: 18px;
                margin-top: 35px;
                inset: 0;
                height: 80%;
                width: 95%;
              "
            >
              <n-table
                :bordered="false"
                :single-line="false"
                style="width: 95%"
              >
                <thead>
                  <tr>
                    <th>Element</th>
                    <th>{{ t("fields.message") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(alert, key) in localAlerts" :key="key">
                    <td>{{ key }}</td>
                    <td>{{ alert }}</td>
                  </tr>
                </tbody>
              </n-table>
            </n-scrollbar>
          </n-card> -->

          <div
            style="
              grid-area: chart-1;
              display: flex;
              justify-content: space-between;
              gap: 1rem;
            "
          >
            <n-card :title="t('objects.environment')" size="small">
              <template #header-extra>
                <div style="display: flex; gap: 0.5rem">
                  <EnvClone
                    :name="env?.EnvInfo?.Env.Name || ''"
                    :manage="
                      userStore.havePermission('Glob_CreateAndDeleteEnvs')
                    "
                  />
                  <n-button
                    strong
                    secondary
                    type="error"
                    size="tiny"
                    @click="showRemoveModal = true"
                    :disabled="
                      !userStore.havePermission('Glob_CreateAndDeleteEnvs')
                    "
                  >
                    <template #icon>
                      <n-icon>
                        <Mdi :path="mdiTrashCan" />
                      </n-icon>
                    </template>
                    {{ t("elements.options.delete") }}
                  </n-button>
                </div>
              </template>
              <div style="display: flex">
                <div class="env-settings">
                  <div>
                    <div class="env-set" style="align-items: center">
                      <p style="min-width: 4em">{{ t("fields.name") }}:</p>
                      <EnvChangeName
                        :name="env?.EnvInfo?.Env.Name || ''"
                        :manage="userStore.havePermission('Env_Rename')"
                      />
                    </div>
                    <!-- <div class="env-set">
                      <p style="min-width: 4em">{{ t("objects.schedule") }}:</p>
                      <Cron />
                    </div> -->

                    <div class="env-set">
                      <p style="min-width: 4em">{{ t("objects.cluster") }}:</p>
                      <n-button
                        size="tiny"
                        strong
                        secondary
                        type="primary"
                        :disabled="
                          !userStore.havePermission('Env_ManageCluster')
                        "
                      >
                        local
                        <template #icon>
                          <n-icon class="icon">
                            <mdi :path="mdiPencil" />
                          </n-icon>
                        </template>
                      </n-button>
                    </div>

                    <div class="env-set">
                      <p style="min-width: 4em">GitOps:</p>
                      <Management
                        :manage="userStore.havePermission('Env_ManageGitOPS')"
                      />
                    </div>

                    <div class="env-set">
                      <p style="min-width: 4em">{{ t("objects.tags", 2) }}:</p>
                      <Tags
                        :tags="env?.EnvInfo?.Env.Tags"
                        :manage="userStore.havePermission('Env_ManageTags')"
                      />
                      <!-- <EnvAddTags /> -->
                    </div>
                  </div>
                </div>
              </div>
            </n-card>
            <n-card
              :title="`${t('objects.element', 2)} & ${t('objects.pod', 2)}`"
              size="small"
            >
              <div
                style="
                  display: flex;
                  justify-content: center;
                  align-items: center;
                  gap: 2rem;
                  height: 100%;
                "
              >
                <n-progress
                  v-if="env?.EnvInfo?.Env.Readiness"
                  style="width: 150px"
                  type="multiple-circle"
                  :percentage="[
                    elementsPercentage(env?.EnvInfo?.Env.Readiness),
                    podsPercentage(env?.EnvInfo?.Env.Readiness),
                  ]"
                  :show-indicator="false"
                  :color="['#18A058', '#C97C10']"
                  :stroke-width="14"
                  :rail-color="['#18A05850', '#C97C1050']"
                />
                <div
                  style="
                    display: flex;
                    flex-flow: column;
                    justify-content: center;
                    gap: 0.5rem;
                  "
                >
                  <div class="legend-container">
                    <div class="legend-dot" style="background: #18a058">
                      <p class="legend-value">
                        {{ env?.EnvInfo?.Env.Readiness.ElementReady }}
                      </p>
                    </div>
                    <p>{{ t("elements.charts.elementsReady") }}</p>
                  </div>
                  <div
                    class="legend-container"
                    v-if="env?.EnvInfo?.Env.Readiness.ElementNotReady"
                  >
                    <div class="legend-dot" style="background: #18a05850">
                      <p class="legend-value">
                        {{ env?.EnvInfo?.Env.Readiness.ElementNotReady }}
                      </p>
                    </div>
                    <p>{{ t("elements.charts.elementsNotReady") }}</p>
                  </div>
                  <div class="legend-container">
                    <div class="legend-dot" style="background: #c97c10">
                      <p class="legend-value">
                        {{ env?.EnvInfo?.Env.Readiness.PodReady }}
                      </p>
                    </div>
                    <p>{{ t("elements.charts.podsReady") }}</p>
                  </div>
                  <div
                    class="legend-container"
                    v-if="env?.EnvInfo?.Env.Readiness.PodNotReady"
                  >
                    <div class="legend-dot" style="background: #c97c1050">
                      <p class="legend-value">
                        {{ env?.EnvInfo?.Env.Readiness.PodNotReady }}
                      </p>
                    </div>
                    <p>{{ t("elements.charts.podsNotReady") }}</p>
                  </div>
                </div>
              </div>
            </n-card>
          </div>
          <!-- wykorzystanie zasobów -->
          <n-card
            :title="t('others.resourceUtilization')"
            size="small"
            style="grid-area: chart-4"
            v-if="env?.Metrics.Enabled"
          >
            <div style="display: flex; justify-content: space-between">
              <div style="display: flex; justify-content: center; width: 100%">
                <div
                  style="
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    flex-direction: column;
                  "
                >
                  <n-icon size="30px" color="#888">
                    <img
                      src="@/assets/icons/cpu.svg"
                      alt="CPU icon"
                      height="30"
                      width="30"
                    />
                  </n-icon>
                  <p
                    style="margin: 0 1em 0 1em; white-space: nowrap"
                    :style="screenWidth <= 1400 ? 'font-size: 12px' : ''"
                  >
                    CPU
                  </p>
                </div>
                <div style="display: grid">
                  <div style="display: flex; align-items: center">
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <n-progress
                          type="line"
                          style="width: 300px"
                          :height="18"
                          :fill-border-radius="0"
                          :border-radius="4"
                          :percentage="
                            Math.round(
                              ((env.EnvInfo?.Env.Resources.CPUUsageAvgCores ||
                                0) /
                                (env.EnvInfo?.Env.Resources.CPUMaxCores || 0)) *
                                100
                            ) || 0
                          "
                          color="var(--successColor)"
                        />
                      </template>
                      {{
                        (env.EnvInfo?.Env?.Resources.CPUUsageAvgCores || 0) /
                        100
                      }}/{{
                        (env.EnvInfo?.Env?.Resources.CPUMaxCores || 0) / 100
                      }}
                      {{ t("objects.core", 3) }}
                    </n-tooltip>
                    <p
                      style="width: 35%; text-align: left; margin-left: 10px"
                      :style="screenWidth <= 1400 ? 'font-size: 12px' : ''"
                    >
                      {{ t("others.used") }}
                    </p>
                  </div>
                  <div style="display: flex; align-items: center">
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <n-progress
                          type="line"
                          style="width: 300px"
                          :height="18"
                          :fill-border-radius="0"
                          :border-radius="4"
                          :percentage="
                            Math.round(
                              ((env.EnvInfo?.Env.Resources.CPUReservedCores ||
                                0) /
                                (env.EnvInfo?.Env.Resources.CPUMaxCores || 0)) *
                                100
                            ) || 0
                          "
                          color="var(--warningColorPressed)"
                        />
                      </template>
                      {{
                        (env.EnvInfo?.Env?.Resources.CPUReservedCores || 0) /
                        100
                      }}/{{
                        (env.EnvInfo?.Env?.Resources.CPUMaxCores || 0) / 100
                      }}
                      {{ t("objects.core", 3) }}
                    </n-tooltip>
                    <p
                      style="width: 35%; text-align: left; margin-left: 10px"
                      :style="screenWidth <= 1400 ? 'font-size: 12px' : ''"
                    >
                      {{ t("others.reserved") }}
                    </p>
                  </div>
                </div>
              </div>
              <n-divider vertical style="height: auto" />
              <div style="display: flex; justify-content: center; width: 100%">
                <div
                  style="
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    flex-direction: column;
                  "
                >
                  <n-icon size="30px" color="#888">
                    <img
                      src="@/assets/icons/ram.svg"
                      alt="CPU icon"
                      height="30"
                      width="30"
                    />
                  </n-icon>
                  <p
                    style="margin: 0 1em 0 1em; white-space: nowrap"
                    :style="screenWidth <= 1400 ? 'font-size: 12px' : ''"
                  >
                    RAM
                  </p>
                </div>
                <div style="display: grid">
                  <div style="display: flex; align-items: center">
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <n-progress
                          type="line"
                          style="width: 300px"
                          :height="18"
                          :fill-border-radius="0"
                          :border-radius="4"
                          :percentage="
                            Math.round(
                              ((env.EnvInfo?.Env.Resources.RAMUsageAvgMB || 0) /
                                (env.EnvInfo?.Env.Resources.RAMMaxMB || 0)) *
                                100
                            ) || 0
                          "
                          color="var(--successColor)"
                        />
                      </template>
                      {{ env.EnvInfo?.Env?.Resources.RAMUsageAvgMB }}/{{
                        env.EnvInfo?.Env?.Resources.RAMMaxMB
                      }}
                      MB
                    </n-tooltip>
                    <p
                      style="width: 35%; text-align: left; margin-left: 10px"
                      :style="screenWidth <= 1400 ? 'font-size: 12px' : ''"
                    >
                      {{ t("others.used") }}
                    </p>
                  </div>
                  <div style="display: flex; align-items: center">
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <n-progress
                          type="line"
                          style="width: 300px"
                          :height="18"
                          :fill-border-radius="0"
                          :border-radius="4"
                          :percentage="
                            Math.round(
                              ((env.EnvInfo?.Env.Resources.RAMReservedMB || 0) /
                                (env.EnvInfo?.Env.Resources.RAMMaxMB || 0)) *
                                100
                            ) || 0
                          "
                          color="var(--warningColorPressed)"
                        />
                      </template>
                      {{ env.EnvInfo?.Env?.Resources.RAMReservedMB }}/{{
                        env.EnvInfo?.Env?.Resources.RAMMaxMB
                      }}
                      MB
                    </n-tooltip>
                    <p
                      style="width: 35%; text-align: left; margin-left: 10px"
                      :style="screenWidth <= 1400 ? 'font-size: 12px' : ''"
                    >
                      {{ t("others.reserved") }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </n-card>
          <!-- <n-card
        v-if="env?.URLs && env?.URLs.length > 0"
        :title="t('objects.url', 2)"
        size="small"
      >
        <div class="urls-container">
          <n-tag
            v-for="url in env?.URLs"
            :key="url"
            type="info"
            :bordered="false"
            @click="openWeb(url)"
            class="url"
            >{{ url }}</n-tag
          >
        </div>
        <div v-else>
          <n-alert
            class="no-urls-msg"
            :title="t('messages.noUrls')"
            :show-icon="false"
          />
        </div>
      </n-card> -->
          <!-- <n-card :title="t('objects.statistics', 2)" size="small">
        <div style="display: flex; gap: 2rem">
          <div>
            <n-progress
              type="multiple-circle"
              :percentage="[107, 20]"
              :circle-gap="2"
              :color="['#13bd76', '#e08b0a']"
              style="width: 120px; margin: 0 8px 12px 0"
              />
              <p>Scamowanie: 107%</p>
            </div>
            <div style="text-align: center">
              <n-progress
              type="circle"
              :percentage="69"
              style="margin: 0 8px 12px"
              />
            <p>Etyka pracy: 69%</p>
          </div>
        </div>
      </n-card> -->
          <!-- <div class="env-charts-container"> -->
          <!-- <n-card
          size="small"
          class="chart-card"
          :_title="'Pods Top CPU'"
          > -->
          <n-card
            size="small"
            style="grid-area: chart-6"
            :title="t('elements.options.history')"
            v-if="userStore.havePermission('Env_ViewLogsEvents')"
          >
            <template #header-extra>
              <n-button
                strong
                secondary
                type="primary"
                size="tiny"
                @click="
                  router.push('/env/history/' + route.params.id + '?tab=events')
                "
              >
                {{ t("actions.more") }}
              </n-button>
            </template>
            <Activities />
          </n-card>
          <!-- </div> -->
          <n-card
            :title="t('objects.metrics', 2)"
            size="small"
            style="grid-area: chart-5"
            v-if="userStore.havePermission('Env_ViewMetrics') && env?.Metrics.Enabled"
          >
            <template #header-extra>
              <PopModal
                v-if="elementsCounter"
                title="Set time"
                :width="'25rem'"
                :show="timeMetrics"
              >
                <template #trigger>
                  <n-button type="primary" secondary size="tiny">
                    <template #icon>
                      <n-icon>
                        <Mdi :path="mdiCalendar" />
                      </n-icon>
                    </template>
                    <span v-if="timeMode === 'relative'">
                      {{ t("fields.last") }} {{ timeSelected }}
                      {{ timeUnitSelected === "m" ? "min" : "h" }}
                    </span>
                    <span v-else style="display: flex">
                      <n-time
                        :time="timerangeSelected[0]"
                        format="MM-dd hh:mm"
                      />
                      <n-icon>
                        <Mdi :path="mdiArrowRightThin" />
                      </n-icon>
                      <n-time
                        :time="timerangeSelected[1]"
                        format="MM-dd hh:mm"
                      />
                    </span>
                  </n-button>
                </template>
                <template #content>
                  <n-tabs type="segment" :default-value="timeMode">
                    <n-tab-pane name="relative" tab="Relative">
                      <div
                        style="display: flex; justify-content: space-between"
                      >
                        <n-input-number
                          v-model:value="time"
                          clearable
                          style="margin-right: 5px"
                          :min="1"
                        />
                        <n-select v-model:value="timeUnit" :options="options" />
                      </div>
                      <div
                        style="
                          display: flex;
                          justify-content: flex-end;
                          margin-top: 1em;
                        "
                      >
                        <n-button
                          secondary
                          type="primary"
                          size="small"
                          @click="timeApply"
                        >
                          {{ t("actions.apply") }}
                        </n-button>
                      </div>
                    </n-tab-pane>
                    <n-tab-pane name="absolute" tab="Absolute">
                      <n-date-picker
                        v-model:value="timerange"
                        type="datetimerange"
                        :is-date-disabled="dateDisabled"
                        clearable
                      />
                      <div
                        style="
                          display: flex;
                          justify-content: flex-end;
                          margin-top: 1em;
                        "
                      >
                        <n-button
                          secondary
                          type="primary"
                          size="small"
                          @click="timerangeApply"
                        >
                          {{ t("actions.apply") }}
                        </n-button>
                      </div>
                    </n-tab-pane>
                  </n-tabs>
                </template>
              </PopModal>
              <n-button
                strong
                secondary
                type="primary"
                size="tiny"
                style="margin-left: 0.5em"
                @click="openWeb('/grafana/')"
              >
                <!-- to trzeba zrobić lepiej ^ -->
                {{ t("actions.more") }}
              </n-button>
            </template>
            <!-- <div
          style="display: flex; justify-content: flex-end; margin-bottom: 1em"
        >
          
        </div> -->
            <n-grid v-if="elementsCounter" :x-gap="12" :y-gap="8" :cols="2">
              <n-grid-item>
                <MetricsIFrame
                  src="/grafana/d-solo/8PxwYSNVz/elements_cpu_max?orgId=1&panelId=2"
                  style="grid-area: chart-5"
                  :refreshRate="5"
                  :mode="'disable'"
                  :timeWindow="timeSelected"
                  :timeUnit="timeUnitSelected"
                  :timerange="timerangeSelected"
                />
              </n-grid-item>
              <n-grid-item>
                <MetricsIFrame
                  src="/grafana/d-solo/QxwUl5N4z/elements_ram_max?orgId=1&panelId=2"
                  style="grid-area: chart-5"
                  :refreshRate="5"
                  :mode="'disable'"
                  :timeWindow="timeSelected"
                  :timeUnit="timeUnitSelected"
                  :timerange="timerangeSelected"
                />
              </n-grid-item>
              <!-- <n-grid-item>
            <MetricsIFrame
              src="/grafana/d-solo/AAOMjeHmk/kubernetes-pod-and-cluster-monitoring-via-prometheus?orgId=1&panelId=8"
              :refreshRate="5"
              style="grid-area: chart-5"
              :mode="timeMode"
              :timeWindow="timeSelected"
              :timeUnit="timeUnitSelected"
              :timerange="timerangeSelected"
            />
          </n-grid-item> -->
              <n-grid-item>
                <MetricsIFrame
                  src="/grafana/d-solo/rukFEu4Vk/cpu_env?orgId=1&panelId=2"
                  :refreshRate="5"
                  :mode="timeMode"
                  :timeWindow="timeSelected"
                  :timeUnit="timeUnitSelected"
                  :timerange="timerangeSelected"
                />
              </n-grid-item>
              <n-grid-item>
                <MetricsIFrame
                  src="/grafana/d-solo/rukFEurss/rss_env?orgId=1&panelId=2"
                  :refreshRate="5"
                  :mode="timeMode"
                  :timeWindow="timeSelected"
                  :timeUnit="timeUnitSelected"
                  :timerange="timerangeSelected"
                />
              </n-grid-item>
              <!-- <n-grid-item>
            <MetricsIFrame
              src="/grafana/d-solo/AAOMjeHmk/kubernetes-pod-and-cluster-monitoring-via-prometheus?orgId=1&panelId=9"
              :refreshRate="5"
              :mode="timeMode"
              :timeWindow="timeSelected"
              :timeUnit="timeUnitSelected"
              :timerange="timerangeSelected"
            />
          </n-grid-item> -->
            </n-grid>
            <div
              v-else
              style="
                display: flex;
                justify-content: center;
                align-items: center;
                height: 10em;
              "
            >
              <p style="font-size: 20px">No data</p>
            </div>
          </n-card>
        </div>
      </n-scrollbar>
      <Modal
        v-model:show="showRemoveModal"
        :title="`${t('actions.deleteEnv')}  '${env?.EnvInfo?.Env.Name}'`"
        :show-icon="false"
        :showFooter="true"
        @positive-click="deleteEnv"
        style="width: 20rem"
      >
        {{ t("questions.sure") }}
      </Modal>
      <Modal
        v-model:show="showRerenderModal"
        :title="`${t('actions.rerender')} '${env?.EnvInfo?.Env.Name}'`"
        :show-icon="false"
        :showFooter="true"
        @positive-click="rerenderEnv"
        style="width: 20rem"
      >
        {{ t("questions.sure") }}
      </Modal>

      <DeletedEnvModal v-if="env?.EnvInfo?.Env.ToDelete" />
    </div>
    <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
  </div>
</template>
<style scoped lang="scss">
.legend-container {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
.legend-value {
  font-weight: 700 !important;
  color: rgba(255, 255, 255, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding-left: 4px;
  padding-right: 4px;
}
.legend-dot {
  min-width: 22px;
  height: 22px;
  border-radius: 10px;
}
</style>
<style lang="scss">
.env-cards {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin: 0 20px 0 20px;

  & .n-card-header__main {
    font-size: 1rem;
  }
}

.urls-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.url {
  cursor: pointer;
  transition: 0.2s ease-in-out;

  &:hover {
    filter: brightness(1.15);
  }
}

.no-urls-msg {
  & .n-alert-body__title {
    font-size: 0.9rem !important;
  }
}

.env-set {
  display: flex;
  gap: 0.5rem;
  align-items: center;

  margin-top: 0.5rem;
  margin-bottom: 1rem;
}

.env-charts-container {
  display: grid;
  margin: 0 20px;
  grid-template-columns: repeat(4, 1fr);
  grid-template-rows: 2fr 1fr 3fr;
  gap: 1rem;
  grid-template-areas: "chart-1 chart-2 chart-3 chart-6" "chart-4 chart-4 chart-4 chart-6" "chart-5 chart-5 chart-5 chart-6" "chart-5 chart-5 chart-5 chart-6";
}
.alerts {
  grid-template-areas: "chart-1 chart-2 chart-3 chart-7" "chart-4 chart-4 chart-4 chart-6" "chart-5 chart-5 chart-5 chart-6" "chart-5 chart-5 chart-5 chart-6";
}

.chart-card {
  min-height: 10rem;
}

.hidden-logs-indicator {
  position: absolute;
  bottom: 0;
  background-color: black;
  padding: 10px;
}
</style>
