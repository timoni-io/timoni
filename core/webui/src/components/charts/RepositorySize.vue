<script lang="ts" setup>
import Chart from "chart.js/auto";
import type { ChartItem } from "chart.js/auto";
import { onMounted } from "vue";
import { ChartConfiguration } from "chart.js";

const { t } = useI18n();

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

// let dateListUnix = [
//   1641211978, 1643890378, 1646309578, 1648984378, 1651576378, 1654254778,
//   1656846778, 1659525178, 1662203578, 1664795578, 1667477578,
// ];

// const dateListTemp = computed(() => {
//   return dateListUnix.map((date) => {
//     return new Date(date * 1000);
//   });
// });

// commit date
const commitChartData = {
  type: "line",
  data: {
    labels: dateList,
    datasets: [
      {
        label: t("repositoryCharts.commitsByDayOfWeek", 2),
        data: [578, 621, 654, 723, 814, 902, 929, 938, 940, 1000, 1024],
        backgroundColor: "#18a05899",
        borderColor: "#18a058",
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
        display: false,
      },
      tooltip: {
        callbacks: {
          label: (label: any) => {
            return `${t("repositoryCharts.commitsByDayOfWeek", 2)}: ${
              label.raw
            }MB`;
          },
        },
      },
    },
    responsive: true,
    animation: false,
    scales: {
      x: {
        display: false,
      },
    },
  },
};

// add chart to DOM
const addChartToDOM = () => {
  const canvas = document.getElementById("repository-size-chart");
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
    <canvas id="repository-size-chart"></canvas>
  </div>
</template>
