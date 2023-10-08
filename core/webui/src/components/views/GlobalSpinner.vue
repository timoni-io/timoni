<script lang="ts" setup>
import { useSpinner } from "@/store/spinner";
import useErrorMsg from "@/utils/errorMsg";
import { useMessage } from "naive-ui";

const { t } = useI18n();

const message = useMessage();
const spinner = useSpinner();
let showSpinner = $ref(false);
const props = defineProps<{
  data: any;
}>();
onMounted(() => {
  setTimeout(() => {
    if (!props.data) showSpinner = true;
  }, 200);
});
watch(
  () => spinner.spinner,
  () => {
    setTimeout(() => {
      if (spinner.spinner) showSpinner = true;
    }, 200);
  }
);

watchEffect(() => {
  if (props.data) {
    showSpinner = false;
    spinner.spinner = false;
  }
});
let { errorComunicate, setErrorComunicate } = useErrorMsg();
watch(errorComunicate, () => {
  if (errorComunicate.value) message.error(t(errorComunicate.value));
  setErrorComunicate("");
});
</script>

<template>
  <n-space
    vertical
    :style="
      !data
        ? 'display: flex; justify-content: center; align-items: center;min-height:400px;'
        : ''
    "
  >
    <!-- {{ spinner }} {{ showSpinner }} -->
    <n-spin
      :show="showSpinner"
      :size="100"
      stroke="#ffff70"
      :stroke-width="10"
      style="height: 90vh"
    >
      <slot v-if="data"></slot>
    </n-spin>
  </n-space>
</template>
