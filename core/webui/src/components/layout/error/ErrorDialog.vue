<script lang="ts" setup>
import useErrorMsg from "@/utils/errorMsg";
const { t } = useI18n();

defineProps<{ msg?: string }>();
let errorMsg = useErrorMsg().errorMsg;
const setErrorMsg = useErrorMsg().setErrorMsg;
let translation = useErrorMsg().errorTranslation.value;

let showErrorModal = $ref(true);

watch(
  () => showErrorModal,
  () => {
    if (!showErrorModal) {
      setErrorMsg(null, false);
    }
  }
);
</script>

<template>
  <Modal
    v-model:show="showErrorModal"
    style="width: 40rem"
    title="Alert"
    :errorModal="true"
    :unclose="true"
  >
    <div>
      {{ msg ? msg : translation ? t(errorMsg as string) : errorMsg }}
    </div>
  </Modal>
</template>
