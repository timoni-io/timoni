<script setup lang="ts">
import { useMessage } from "naive-ui";

const { isOnline, effectiveType, isSupported } = $(useNetwork());

const message = useMessage();
const { t, locale } = useI18n();

onMounted(() => {
  message.destroyAll();
  if (!isOnline) {
    message.error(t("network.offline"), { duration: 0 });
  } else if (
    isSupported &&
    effectiveType !== "4g" &&
    effectiveType !== "3g" &&
    !window.navigator.userAgent.includes("Firefox")
  ) {
    message.warning(t("network.slowNetwork"), {
      duration: 0,
      closable: true,
    });
  }
});
watch(
  () => [isOnline, effectiveType, locale.value],
  () => {
    message.destroyAll();
    if (!isOnline) {
      message.error(t("network.offline"), { duration: 0 });
    } else if (
      effectiveType !== "4g" &&
      effectiveType !== "3g" &&
      !window.navigator.userAgent.includes("Firefox")
    ) {
      message.warning(t("network.slowNetwork"), {
        duration: 0,
        closable: true,
      });
    }
  }
);
</script>
<template><div></div></template>
