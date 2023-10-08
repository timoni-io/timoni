<script lang="ts" setup>
import useErrorMsgEnv from "@/utils/errorMsgEnv";

let errorMsgEnv = useErrorMsgEnv().errorMsgEnv;
const setErrorMsgEnv = useErrorMsgEnv().setErrorMsgEnv;

let showErrorModal = $ref(true);

watch(
  () => showErrorModal,
  () => {
    if (!showErrorModal) {
      setErrorMsgEnv(null);
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
      {{ errorMsgEnv }}
    </div>
  </Modal>
</template>
