<script lang="ts" setup>
import { useI18n } from "vue-i18n";

const { t } = useI18n();

let props = defineProps<{
  id: string;
  size: "tiny" | "small" | "medium" | "large";
  timeUnitTo: "m" | "h" | "d" | "mo"; // "Minutes" | "Hours" | "Days" | "Months";
  responsiveVersion?: boolean;
  daysMonths?: boolean;
  defaultTimeProps?: { time: number; timeUnit: string };
}>();

const emit = defineEmits<{
  (e: "timeSettingChanged"): void;
}>();

let defaultTime = $computed(() => {
  if (props.defaultTimeProps) return props.defaultTimeProps;
  else return { time: 15, timeUnit: "m" };
});

let timeMetrics = $ref(true);
let time = $ref<number>(
  parseInt(localStorage.getItem(`timeSelected${props.id}`) as string) ||
    defaultTime.time
);
let timeUnit = $ref<string>(
  localStorage.getItem(`timeUnitSelected${props.id}`) || defaultTime.timeUnit
);
const getTimeUnit = () => {
  timeUnit =
    localStorage.getItem(`timeUnitSelected${props.id}`) || defaultTime.timeUnit;
};
let timeSelected = $ref<number>(
  parseInt(localStorage.getItem(`timeSelected${props.id}`) as string) ||
    defaultTime.time
);
let timeUnitSelected = $ref<string>(
  localStorage.getItem(`timeUnitSelected${props.id}`) || defaultTime.timeUnit
);
let timeMode = $ref<string>(
  localStorage.getItem(`timeMode${props.id}`) || "relative"
);
let timerange = $ref<[number, number]>([Date.now(), Date.now()]);
let timerangeSelected = $ref<[number, number]>([
  parseInt(
    localStorage.getItem(`timerangeSelectedFrom${props.id}`) as string
  ) || Date.now(),
  parseInt(localStorage.getItem(`timerangeSelectedTo${props.id}`) as string) ||
    Date.now(),
]);

const options = computed(() => {
  const timeUnitOptions = [
    {
      label: "Minutes",
      value: "m",
    },
    {
      label: "Hours",
      value: "h",
    },
    {
      label: "Days",
      value: "d",
    },
    {
      label: "Months",
      value: "mo",
    },
  ];
  if (props.daysMonths) return timeUnitOptions.slice(1);
  if (props.timeUnitTo === "m") return timeUnitOptions.slice(0, 1);
  if (props.timeUnitTo === "h") return timeUnitOptions.slice(0, 2);
  if (props.timeUnitTo === "d") return timeUnitOptions.slice(0, 3);
  if (props.timeUnitTo === "mo") return timeUnitOptions;
});

const timeApply = () => {
  timeSelected = time;
  timeUnitSelected = timeUnit;
  timeMode = "relative";
  localStorage.setItem(`timeMode${props.id}`, timeMode);
  localStorage.setItem(`timeSelected${props.id}`, "" + timeSelected);
  localStorage.setItem(`timeUnitSelected${props.id}`, timeUnitSelected);
  timeMetrics = false;
  emit("timeSettingChanged");
  setTimeout(() => {
    timeMetrics = true;
  }, 10);
};
const timerangeApply = () => {
  timerangeSelected = timerange;
  timeMode = "absolute";
  localStorage.setItem(`timeMode${props.id}`, timeMode);
  localStorage.setItem(
    `timerangeSelectedFrom${props.id}`,
    "" + timerangeSelected[0]
  );
  localStorage.setItem(
    `timerangeSelectedTo${props.id}`,
    "" + timerangeSelected[1]
  );
  timeMetrics = false;
  emit("timeSettingChanged");
  setTimeout(() => {
    timeMetrics = true;
  }, 10);
};

const dateDisabled = (ts: number) => {
  return new Date(ts).getTime() > Date.now();
};

const timeUnitShortcut = () => {
  let unit = "";
  switch (timeUnitSelected) {
    case "m":
      unit = "min";
      break;
    case "h":
      unit = "h";
      break;
    case "d":
      unit = "d";
      break;
    case "mo":
      unit = "mo";
      break;
    default:
      unit = "";
  }
  timeUnitSelected === "m" ? "min" : "h";
  return unit;
};
</script>

<template>
  <PopModal title="Set time" :width="'25rem'" :show="timeMetrics">
    <template #trigger>
      <n-tooltip
        trigger="hover"
        :disabled="!responsiveVersion"
        :z-index="20000"
      >
        <template #trigger>
          <n-button type="primary" secondary :size="size" @click="getTimeUnit">
            <template #icon>
              <n-icon><Mdi :path="mdiCalendar" /></n-icon>
            </template>
            <div v-if="!responsiveVersion">
              <span v-if="timeMode === 'relative'">
                {{ t("fields.last") }} {{ timeSelected }}
                {{ timeUnitShortcut() }}
              </span>
              <span v-else style="display: flex">
                <n-time :time="timerangeSelected[0]" format="MM-dd hh:mm" />
                <n-icon><Mdi :path="mdiArrowRightThin" /></n-icon>
                <n-time :time="timerangeSelected[1]" format="MM-dd hh:mm" />
              </span>
            </div>
          </n-button>
        </template>
        <div>
          <span v-if="timeMode === 'relative'">
            {{ t("fields.last") }} {{ timeSelected }}
            {{ timeUnitShortcut() }}
          </span>
          <span v-else style="display: flex">
            <n-time :time="timerangeSelected[0]" format="MM-dd hh:mm" />
            <n-icon><Mdi :path="mdiArrowRightThin" /></n-icon>
            <n-time :time="timerangeSelected[1]" format="MM-dd hh:mm" />
          </span>
        </div>
      </n-tooltip>
    </template>
    <template #content>
      <n-tabs type="segment" :default-value="timeMode">
        <n-tab-pane name="relative" tab="Relative">
          <div style="display: flex; justify-content: space-between">
            <n-input-number
              v-model:value="time"
              clearable
              style="margin-right: 5px"
              :min="1"
            />
            <n-select v-model:value="timeUnit" :options="options" />
          </div>
          <div
            style="display: flex; justify-content: flex-end; margin-top: 1em"
          >
            <n-button secondary type="primary" size="small" @click="timeApply">
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
            style="display: flex; justify-content: flex-end; margin-top: 1em"
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
</template>
