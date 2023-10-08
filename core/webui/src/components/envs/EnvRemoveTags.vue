<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { useRoute } from "vue-router";

const { t } = useI18n();
const message = useMessage();
const route = useRoute();

let props = defineProps<{ tagToRemove: string }>();

let removeTagOpen = $ref(true);

const removeTag = () => {
  api
    .get("/env-tag-delete", {
      queries: {
        env: route.params.id as string,
        name: props.tagToRemove,
      },
    })
    .then((res) => {
      // if (res === "permission denied") {
      //   message.error(t("messages.permissionDenied"));
      //   return;
      // }
      if (res === "ok") {
        message.success(t("messages.tagRemoved"));
        removeTagOpen = !removeTagOpen;
      } else {
        message.error(res);
      }
    })
    .catch((error) => {
      message.error(error);
    });
};
</script>

<template>
  <PopModal
    :title="`${t('actions.removeTag')}`"
    @positive-click="removeTag"
    @negative-click="() => {}"
    :show-footer="{
      positiveText: t('actions.confirm'),
      negativeText: t('actions.cancel'),
    }"
    :show="removeTagOpen"
    :width="'20rem'"
  >
    <template #trigger>
      <n-icon class="icon" color="#18a058"><mdi :path="mdiClose" /></n-icon>
    </template>
    <template #content>
      {{ t("questions.sure") }}
    </template>
  </PopModal>
</template>

<style scoped lang="scss">
.icon {
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  &:hover {
    filter: brightness(1.4);
  }
}
.alias-input {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 0.5rem;
  justify-items: start;
  align-items: center;
}
</style>
