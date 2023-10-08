<script lang="ts" setup>
import Chart from "chart.js/auto";
import type { ChartItem } from "chart.js/auto";
import { onMounted } from "vue";
import { ChartConfiguration } from "chart.js";
import chroma from "chroma-js";

let dateList = [
  "09-10-2022",
  "08-10-2022",
  "09-10-2022",
  "10-10-2022",
  "11-10-2022",
  "12-10-2022",
  "13-10-2022",
  "14-10-2022",
  "15-10-2022",
  "16-10-2022",
  "17-10-2022",
  "18-10-2022",
  "19-10-2022",
  "20-10-2022",
  "21-10-2022",
];

const generateDate = (userNum: number) => {
  let date = [];
  let color = chroma.scale(["#A6BC09", "#01415B"]).mode("lch").colors(userNum);

  for (let i = 1; i <= userNum; i++) {
    date.push({
      label: `User ${i}`,
      data: Array.from({ length: dateList.length }, () =>
        Math.floor(Math.random() * 10)
      ),
      backgroundColor: color[i - 1],
    });
  }

  return date;
};

// commit date
const commitChartData = {
  type: "bar",
  data: {
    labels: dateList,
    datasets: generateDate(8),
  },
  options: {
    plugins: {
      title: {
        display: false,
      },
    },
    responsive: true,
    scales: {
      x: {
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
  <div style="width: 100%; display: flex; height: 100%; align-items: center">
    <canvas id="commit-chart"></canvas>
  </div>
</template>
