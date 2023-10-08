<script lang="ts" setup>
import Chart from "chart.js/auto";
import type { ChartItem } from "chart.js/auto";
import { onMounted } from "vue";
import { ChartConfiguration } from "chart.js";

let dateList = [
  "01-01-2022",
  "01-02-2022",
  "01-03-2022",
  "01-04-2022",
  "01-05-2022",
  "01-06-2022",
  "01-07-2022",
  "01-08-2022",
  "01-09-2022",
  "01-10-2022",
  "01-11-2022",
];

// commit date
const commitChartData = {
  type: "bar",
  data: {
    labels: dateList,
    datasets: [
      {
        label: "Removed lines",
        data: [278, 300, 504, 623, 914, 918, 982, 994, 1000, 1015, 1234],
        backgroundColor: "#d03050",
        cubicInterpolationMode: "monotone",
        tension: 0.4,
      },
      {
        label: "Changed lines",
        data: [100, 221, 702, 723, 814, 902, 929, 938, 940, 1000, 1024],
        backgroundColor: "#f0a020",
        cubicInterpolationMode: "monotone",
        tension: 0.4,
      },
      {
        label: "Added lines",
        data: [578, 682, 702, 723, 814, 902, 929, 938, 940, 1000, 930],
        backgroundColor: "#18a058",
        cubicInterpolationMode: "monotone",
        tension: 0.4,
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
    },
    interaction: {
      mode: "nearest",
      axis: "x",
      intersect: false,
    },
    responsive: true,
    animation: false,
    scales: {
      x: {
        display: false,
        stacked: true,
      },
      y: {
        stacked: true,
      },
    },
  },
};

// add chart to DOM
const addChartToDOM = () => {
  const canvas = document.getElementById("repo-activity-chart");
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
    <canvas id="repo-activity-chart"></canvas>
  </div>
</template>
