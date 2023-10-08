<script setup lang="ts">
import stc from "string-to-color";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

// props
defineProps<{ timeChangeTrigger: number; id: string }>();

// elements data
interface MostActiveUser {
  Name: string;
  LastName: string;
  AddLines: number;
  ChangedLines: number;
  RemoveLines: number;
}

let mostActiveElements: MostActiveUser[] = [
  {
    Name: "Matthew",
    LastName: "Sherman",
    AddLines: 304,
    ChangedLines: 270,
    RemoveLines: 122,
  },
  {
    Name: "Michael",
    LastName: "Bright",
    AddLines: 293,
    ChangedLines: 262,
    RemoveLines: 108,
  },
  {
    Name: "Pam",
    LastName: "Simmons",
    AddLines: 290,
    ChangedLines: 232,
    RemoveLines: 88,
  },
  {
    Name: "Norma",
    LastName: "Hughes",
    AddLines: 302,
    ChangedLines: 201,
    RemoveLines: 0,
  },
  {
    Name: "Angela",
    LastName: "Owen",
    AddLines: 150,
    ChangedLines: 120,
    RemoveLines: 92,
  },
];

// max changes
const maxChanges = computed(() => {
  return (
    mostActiveElements[0].AddLines +
    mostActiveElements[0].ChangedLines +
    mostActiveElements[0].RemoveLines
  );
});

// get percentage
const getAddLinesPerc = (element: MostActiveUser) => {
  return (element.AddLines / maxChanges.value) * 100;
};

const getChangedLinesPerc = (element: MostActiveUser) => {
  return (element.ChangedLines / maxChanges.value) * 100;
};

const getRemoveLinesPerc = (element: MostActiveUser) => {
  return (element.RemoveLines / maxChanges.value) * 100;
};

// time
// let timeMode = $ref<string>();
// let timeSelected = $ref<number>();
// let timeUnitSelected = $ref<string>();
// let timerangeSelected = $ref<[number, number]>();

// const getTimeSetting = () => {
//   timeMode = localStorage.getItem(`timeMode${props.id}`) || "relative";
//   timeSelected =
//     parseInt(localStorage.getItem(`timeSelected${props.id}`) as string) || 15;
//   timeUnitSelected = localStorage.getItem(`timeUnitSelected${props.id}`) || "m";
//   timerangeSelected = [
//     parseInt(
//       localStorage.getItem(`timerangeSelectedFrom${props.id}`) as string
//     ) || Date.now(),
//     parseInt(
//       localStorage.getItem(`timerangeSelectedTo${props.id}`) as string
//     ) || Date.now(),
//   ];
// };

// watch(
//   () => props.timeChangeTrigger,
//   () => {
//     getTimeSetting();
//   }
// );

// onBeforeMount(() => {
//   getTimeSetting();
// });

// screen height
let screenHeight = $ref<number>(800);

onBeforeMount(() => {
  screenHeight = window.innerHeight;
});

onMounted(() => {
  window.addEventListener("resize", () => {
    screenHeight = window.innerHeight;
  });
});

const getMostActiveElements = $computed(() => {
  if (screenHeight > 740) return mostActiveElements.slice(0, 5);
  else return mostActiveElements.slice(0, 4);
});
</script>

<template>
  <div style="display: flex; flex-direction: column; gap: 0.4rem; height: 18vh">
    <div v-for="element in getMostActiveElements" :key="element.Name">
      <div class="most-changed-elements--bar-number">
        <n-tooltip trigger="hover" class="most-active-user--tooltip">
          <template #trigger>
            <n-avatar
              round
              size="small"
              :style="{
                color: stc(element.LastName),
                backgroundColor: stc(element.LastName) + '4D',
              }"
              style="height: 20px; width: 20px; font-size: 0.7rem"
            >
              {{
                `${element.Name.slice(
                  0,
                  1
                ).toUpperCase()}${element.LastName.slice(0, 1).toUpperCase()}`
              }}
            </n-avatar>
          </template>
          <div
            style="font-size: 0.8rem; font-family: Arial, Helvetica, sans-serif"
          >
            {{ `${element.Name} ${element.LastName}` }}
          </div>
        </n-tooltip>
        <n-tooltip trigger="hover" class="most-active-user--tooltip">
          <template #trigger>
            <div class="most-changed-elements--bar">
              <div
                v-if="element.AddLines > 0"
                style="height: 100%; background: var(--successColor)"
                :style="`width: ${getAddLinesPerc(element)}%`"
              ></div>
              <div
                v-if="element.ChangedLines > 0"
                style="height: 100%; background: var(--warningColor)"
                :style="`width: ${getChangedLinesPerc(element)}%`"
              ></div>
              <div
                v-if="element.RemoveLines > 0"
                style="height: 100%; background: var(--errorColor)"
                :style="`width: ${getRemoveLinesPerc(element)}%`"
              ></div>
            </div>
          </template>
          <div
            style="font-size: 0.8rem; font-family: Arial, Helvetica, sans-serif"
          >
            {{
              `${t("repositoryCharts.addedLines")}: ${element.AddLines}, ${t(
                "repositoryCharts.changedLines"
              ).toLocaleLowerCase()}: ${element.ChangedLines} , ${t(
                "repositoryCharts.removedLines"
              ).toLocaleLowerCase()}: ${element.RemoveLines}`
            }}
          </div>
        </n-tooltip>
        <div style="font-weight: bold; white-space: nowrap">
          {{ element.AddLines + element.RemoveLines + element.ChangedLines }}
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.most-changed-elements--bar-number {
  display: grid;
  grid-template-columns: 1.5rem auto 2rem;
  align-items: center;
  gap: 1rem;
}

.most-changed-elements--bar {
  height: 0.75rem;
  display: flex;
  overflow: hidden;
  justify-content: flex-start;
}
.most-changed-elements--bar > div:first-of-type {
  border-top-left-radius: 0rem;
  border-bottom-left-radius: 0rem;
}

.most-changed-elements--bar > div:last-of-type {
  border-top-right-radius: 0rem;
  border-bottom-right-radius: 0rem;
}

.most-active-user--tooltip.n-popover:not(.n-popover--raw) {
  background-color: rgba(0, 0, 0, 0.7);
}

.most-active-user--tooltip.n-popover-shared
  .n-popover-arrow-wrapper
  .n-popover-arrow {
  background-color: rgba(0, 0, 0, 0.7);
}
</style>
