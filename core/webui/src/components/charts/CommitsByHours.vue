<script lang="ts" setup>
import Chart from "chart.js/auto";
import type { ChartItem } from "chart.js/auto";
import { onMounted } from "vue";
import { ChartConfiguration } from "chart.js";

const { t } = useI18n();

let dateList = [
  0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
  22, 23,
];

// commit date
const commitChartData = {
  type: "bar",
  data: {
    labels: dateList,
    datasets: [
      {
        label: t("repositoryCharts.commitsByHour", 2),
        data: [
          42, 34, 22, 12, 10, 12, 13, 26, 37, 45, 70, 78, 72, 68, 77, 86, 72,
          58, 52, 42, 37, 30, 20, 20,
        ],
        backgroundColor: "#18a058",
      },
    ],
  },
  options: {
    maintainAspectRatio: false,
    plugins: {
      title: {
        display: false,
      },
      legend: {
        display: false,
      },
      tooltip: {
        callbacks: {
          title: () => {
            return "";
          },
        },
      },
    },
    responsive: true,
    animation: false,
  },
};

// add chart to DOM
const addChartToDOM = () => {
  const canvas = document.getElementById("commit-chart");
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
    <canvas id="commit-chart"></canvas>
  </div>
</template>
