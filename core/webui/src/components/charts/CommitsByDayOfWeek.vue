<script lang="ts" setup>
import Chart from "chart.js/auto";
import type { ChartItem } from "chart.js/auto";
import { onMounted } from "vue";
import { ChartConfiguration } from "chart.js";

const { t } = useI18n();

let dayList = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

let dayListTranslated = computed(() => {
  return dayList.map((day) => {
    return t(`time.daysFull.${day}`);
  });
});
// commit date
const commitChartData = {
  type: "bar",
  data: {
    labels: dayListTranslated.value,
    datasets: [
      {
        label: t("repositoryCharts.commitsByDayOfWeek", 2),
        data: [2500, 3100, 2900, 2400, 1900, 300, 500],
        backgroundColor: "#18a058",
      },
    ],
  },
  options: {
    indexAxis: "y",
    maintainAspectRatio: false,
    plugins: {
      title: {
        display: false,
      },
      legend: {
        display: false,
      },
    },
    responsive: true,
    animation: false,
  },
};

// add chart to DOM
const addChartToDOM = () => {
  const canvas = document.getElementById("commit-by-day-chart");
  if (canvas) {
    new Chart(canvas as ChartItem, commitChartData as ChartConfiguration);
  }
};

onMounted(() => {
  addChartToDOM();
});
</script>

<template>
  <div style="width: 100%; display: flex; height: 18vh; align-items: center">
    <canvas id="commit-by-day-chart"></canvas>
  </div>
</template>
