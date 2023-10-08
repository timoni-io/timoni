<script setup lang="ts">
import { useI18n } from "vue-i18n";

interface MostActiveElement {
  Name: string;
  Successes: number;
  Failures: number;
  InProgress: number;
  TotalChanges: number;
}

// props
let props = defineProps<{
  timeChangeTrigger: number;
  id: string;
  MostChangedElements: Array<MostActiveElement>;
}>();

const { t } = useI18n();

// elements data

// max total changes
const maxTotalChanges = computed(() => {
  return props.MostChangedElements[0].TotalChanges;
});

// get percentage
const getSuccessPerc = (element: MostActiveElement) => {
  if (maxTotalChanges.value > 0)
    return (element.Successes / maxTotalChanges.value) * 100;
  return 0;
};

const getErrorPerc = (element: MostActiveElement) => {
  if (maxTotalChanges.value > 0)
    return (element.Failures / maxTotalChanges.value) * 100;
  return 0;
};

const getBuildingPerc = (element: MostActiveElement) => {
  if (maxTotalChanges.value > 0)
    return (element.InProgress / maxTotalChanges.value) * 100;
  return 0;
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
let localMostChangedElements = computed(() => {
  const tmp = props.MostChangedElements;
  return tmp.sort(
    (a: any, b: any) =>
      b.TotalChanges - a.TotalChanges || a.Name.localeCompare(b.Name)
  );
});
</script>

<template>
  <div>
    <div v-if="localMostChangedElements && localMostChangedElements.length > 0">
      <div
        v-for="element in localMostChangedElements.slice(0, 20)"
        :key="element.Name"
      >
        <div>
          {{ element.Name }}
        </div>
        <div class="most-changed-elements--bar-number">
          <div class="most-changed-elements--bar">
            <div
              v-if="element.Successes > 0"
              style="height: 100%; background: var(--successColor)"
              :style="`width: ${getSuccessPerc(element)}%`"
            ></div>
            <div
              v-if="element.Failures > 0"
              style="height: 100%; background: var(--errorColor)"
              :style="`width: ${getErrorPerc(element)}%`"
            ></div>
            <div
              v-if="element.InProgress > 0"
              style="height: 100%; background: var(--warningColor)"
              :style="`width: ${getBuildingPerc(element)}%`"
            ></div>
          </div>
          <div style="font-weight: bold; white-space: nowrap">
            {{ element.TotalChanges }}
          </div>
        </div>
      </div>
    </div>
    <div v-else class="no-data-card">
      <n-icon :size="16">
        <Mdi :path="mdiInformationOutline" />
      </n-icon>
      <p>{{ t("fields.noData") }}</p>
    </div>
  </div>
</template>

<style scoped>
.most-changed-elements--bar-number {
  display: grid;
  grid-template-columns: auto 2rem;
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
  border-top-left-radius: 0.15rem;
  border-bottom-left-radius: 0.15rem;
}

.most-changed-elements--bar > div:last-of-type {
  border-top-right-radius: 0.15rem;
  border-bottom-right-radius: 0.15rem;
}

.no-data-card {
  background-color: rgba(0, 0, 0, 0.2);
  padding: 0.5rem 1rem;
  border-radius: 0.2rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
</style>
