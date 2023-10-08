<script lang="ts" setup>
import chroma from "chroma-js";
import { useI18n } from "vue-i18n";

let colors = chroma.scale(["#fafa6e", "#18a058", "#444"]).mode("lch").colors(7);
const { t } = useI18n();

const repoData = [
  { time: 1663451999, commitNumber: 32 },
  { time: 1663538399, commitNumber: 23 },
  { time: 1663624799, commitNumber: 16 },
  { time: 1663711199, commitNumber: 12 },
  { time: 1663797599, commitNumber: 5 },
  { time: 1663883999, commitNumber: 0 },
  { time: 1663970399, commitNumber: 2 },
  { time: 1664056799, commitNumber: 4 },
  { time: 1664143199, commitNumber: 14 },
  { time: 1664229599, commitNumber: 26 },
  { time: 1664315999, commitNumber: 18 },
  { time: 1664402399, commitNumber: 12 },
  { time: 1664488799, commitNumber: 11 },
  { time: 1664575199, commitNumber: 6 },
  { time: 1664661599, commitNumber: 0 },
  { time: 1664747999, commitNumber: 0 },
  { time: 1664834399, commitNumber: 6 },
  { time: 1664920799, commitNumber: 21 },
  { time: 1665007199, commitNumber: 6 },
  { time: 1665093599, commitNumber: 5 },
  { time: 1665179999, commitNumber: 26 },
  { time: 1665266399, commitNumber: 3 },
  { time: 1665352799, commitNumber: 4 },
  { time: 1665439199, commitNumber: 3 },
  { time: 1665525599, commitNumber: 3 },
  { time: 1665611999, commitNumber: 12 },
  { time: 1665698399, commitNumber: 26 },
  { time: 1665784799, commitNumber: 27 },
  { time: 1665871199, commitNumber: 26 },
  { time: 1665957599, commitNumber: 38 },
  { time: 1666043999, commitNumber: 32 },
  { time: 1666130399, commitNumber: 33 },
  { time: 1666216799, commitNumber: 3 },
  { time: 1666303199, commitNumber: 6 },
];

const week = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const monthNames = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

const getWeekDays = () => {
  let daysList = [];
  for (let i = 0; i < 7; i++) {
    daysList.push(week[(new Date(repoData[0].time * 1000).getDay() + i) % 7]);
  }
  return daysList;
};

const weekDays = getWeekDays();

const largestCommitNumber = Math.max(
  ...repoData.map((data) => data.commitNumber)
);

const getDate = (unixTime: number) => {
  const date = new Date(unixTime * 1000);
  return `${t(
    `time.monthsShort.${monthNames[date.getMonth()]}`
  )} ${date.getDate()}, ${date.getFullYear()}`;
};

const getColor = (commitNumber: number) => {
  let color = "";
  switch (true) {
    case commitNumber >= largestCommitNumber:
      color = colors[0];
      break;
    case commitNumber >= (largestCommitNumber / 5) * 4:
      color = colors[1];
      break;
    case commitNumber >= (largestCommitNumber / 5) * 3:
      color = colors[2];
      break;
    case commitNumber >= (largestCommitNumber / 5) * 2:
      color = colors[3];
      break;
    case commitNumber >= largestCommitNumber / 5:
      color = colors[4];
      break;
    case commitNumber > 0:
      color = colors[5];
      break;
    default:
      color = colors[6];
      break;
  }
  return color;
};
</script>

<template>
  <div style="display: flex">
    <div
      style="
        display: grid;
        font-size: 0.7rem;
        padding-right: 0.5rem;
        padding-top: 1.5rem;
        width: 2rem;
      "
    >
      <p>{{ t(`time.days.${weekDays[0]}`) }}</p>
      <p style="display: flex; align-items: center">
        {{ t(`time.days.${weekDays[3]}`) }}
      </p>
      <p style="display: flex; align-items: flex-end">
        {{ t(`time.days.${weekDays[6]}`) }}
      </p>
    </div>
    <div class="heat-map--columns">
      <div v-for="(data, index) in repoData" :key="data.time">
        <div
          style="
            width: 1rem;
            white-space: nowrap;
            font-size: 0.7rem;
            height: 1.5rem;
          "
          v-if="index % 7 === 0"
        >
          <p v-if="index % 21 === 0">
            {{
              t(
                `time.monthsShort.${
                  monthNames[new Date(data.time * 1000).getMonth()]
                }`
              )
            }}
          </p>
        </div>
        <n-tooltip trigger="hover">
          <template #trigger>
            <div
              class="heat-map--square"
              :style="{ backgroundColor: getColor(data.commitNumber) }"
            ></div>
          </template>
          {{
            `${data.commitNumber} ${
              data.commitNumber === 1
                ? t("objects.commitOn").toLowerCase()
                : t("objects.commitOn", 2).toLowerCase()
            } ${getDate(data.time)}`
          }}
        </n-tooltip>
      </div>
    </div>
  </div>
</template>

<style>
.heat-map--columns {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  height: 11.5rem;
  flex-wrap: wrap;
}

.heat-map--square {
  height: 1rem;
  width: 1rem;
  border-radius: 0.15rem;
}
</style>
