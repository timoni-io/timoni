<script lang="ts" setup>
import Chart from "chart.js/auto";
import type { ChartItem } from "chart.js/auto";
import { onMounted } from "vue";
import { ChartConfiguration } from "chart.js";

const image = new Image();
image.src = "https://www.chartjs.org/img/chartjs-logo.svg";

let dataPoints = [14, 12, 11, 9, 9, 6, 3, 2];

const barAvatar = {
  id: "barAvatar",
  afterDatasetDraw(chart: any) {
    const {
      ctx,
      scales: { x, y },
    } = chart;
    ctx.save();
    let imgSize = 30;
    for (let i = 0; i < 8; i++) {
      ctx.drawImage(
        image,
        x.getPixelForValue(i) - imgSize / 2,
        y.getPixelForValue(dataPoints[i]) - imgSize - 5,
        imgSize,
        imgSize
      );
    }
  },
};

const generateData = (userNum: number) => {
  let labels = [];

  for (let i = 0; i < userNum; i++) {
    labels.push(`User ${i + 1}`);
  }

  return {
    labels: labels,
    datasets: [
      {
        label: "Datasets",
        data: dataPoints,
        backgroundColor: "#A6BC09",
      },
    ],
  };
};

// commit date
const commitChartData = {
  type: "bar",
  data: generateData(8),
  options: {
    layout: {
      padding: 40,
    },
    plugins: {
      title: {
        display: false,
      },
      legend: {
        display: false,
      },
    },
    responsive: true,
  },
  plugins: [barAvatar],
};

// add chart to DOM
const addChartToDOM = () => {
  const canvas = document.getElementById("users-chart");
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
    <canvas id="users-chart"></canvas>
  </div>
</template>
