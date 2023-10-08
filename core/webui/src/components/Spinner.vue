<script lang="ts" setup>
let showSpinner = $ref(false);

const props = defineProps<{
  data: any;
  spinnerSize?: number;
  showSpinnerProp?: boolean;
}>();

setTimeout(() => {
  if (!props.data) {
    showSpinner = true;
  }
}, 300);
watch([() => props.data, () => props.showSpinnerProp], () => {
  if (!props.data || props.showSpinnerProp) {
    setTimeout(() => {
      if (!props.data || props.showSpinnerProp) {
        showSpinner = true;
      }
    }, 300);
  } else showSpinner = false;
});
</script>

<template>
  <!-- <n-space
    vertical
    :style="
      !data
        ? 'display: flex; justify-content: center; align-items: center; height: 100%'
        : ''
    "
  > -->
  <n-spin
    :show="showSpinner"
    :size="spinnerSize || 60"
    stroke="#1ba3fd"
    :stroke-width="10"
    style="height: 100%"
  >
    <slot v-if="data"></slot>
    <!-- <div v-if="!data && showSpinner" style="min-height: 60px"></div> -->
  </n-spin>
  <!-- </n-space> -->
</template>
