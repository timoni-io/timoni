<script setup lang="ts">
import { useRoute } from "vue-router";

const route = useRoute();
const env = computed(() => route.params.id as string);
const props = defineProps<{
  src: string;
  refreshRate: number;
  mode: string;
  timeWindow: number;
  timeUnit: string;
  timerange: [number, number];
}>();

const src = computed(() => {
  if (props.mode === "relative")
    return (
      props.src +
      "&refresh=" +
      props.refreshRate +
      "s&var-env=" +
      env.value +
      "&from=now-" +
      props.timeWindow +
      props.timeUnit +
      "&to=now"
    );
  else if (props.mode === "disable")
    return (
      props.src +
      "&refresh=" +
      props.refreshRate +
      "s&var-env=" +
      env.value +
      "&from=now-30s&to=now"
    );
  else
    return (
      props.src +
      "s&var-env=" +
      env.value +
      "&from=" +
      props.timerange[0] +
      "&to=" +
      props.timerange[1]
    );
});
</script>

<template>
  <iframe :src="src" width="100%" height="100%" frameborder="0"></iframe>
</template>
