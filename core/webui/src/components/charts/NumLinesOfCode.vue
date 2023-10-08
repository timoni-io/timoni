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
  type: "line",
  data: {
    labels: dateList,
    datasets: [
      {
        label: "Vue",
        data: [578, 682, 702, 723, 814, 902, 929, 938, 940, 1000, 930],
        backgroundColor: "#18a05899",
        borderColor: "#18a058",
        fill: true,
        cubicInterpolationMode: "monotone",
        tension: 0.4,
      },
      {
        label: "JS",
        data: [100, 221, 702, 723, 814, 902, 929, 938, 940, 1000, 1024],
        backgroundColor: "#f0a02099",
        borderColor: "#f0a020",
        fill: true,
        cubicInterpolationMode: "monotone",
        tension: 0.4,
      },
      {
        label: "Go",
        data: [278, 300, 504, 623, 914, 918, 982, 994, 1000, 1015, 1234],
        backgroundColor: "#1ba3fd99",
        borderColor: "#1ba3fd",
        fill: true,
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
        position: "right",
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
      },
      y: {
        stacked: true,
      },
    },
  },
};

// add chart to DOM
const addChartToDOM = () => {
  const canvas = document.getElementById("lines-of-code-chart");
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
    <canvas id="lines-of-code-chart"></canvas>
  </div>
</template>
