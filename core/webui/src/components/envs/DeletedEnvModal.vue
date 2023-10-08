<script setup lang="ts">
import useErrorMsg from "@/utils/errorMsg";
import { useRouter } from "vue-router";

const router = useRouter();
let { setEnvError } = useErrorMsg();

const redirect = () => {
  setEnvError(null);
  router.replace("/env");
};
</script>

<template>
  <Modal
    show
    :title="$t('messages.envDeleted')"
    style="width: 60rem; max-width: calc(100% - 2rem)"
    :show-footer="{
      positiveText: 'OK',
    }"
    unclose
    @positive-click="redirect"
  >
    <div style="display: flex; align-items: center; gap: 1rem">
      <n-icon :size="50" color="var(--errorColor)">
        <Mdi :path="mdiTrashCan" />
      </n-icon>
      {{ $t("messages.envDeleted") }}
    </div>
  </Modal>
</template>
