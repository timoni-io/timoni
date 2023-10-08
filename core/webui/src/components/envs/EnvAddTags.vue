<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { useRoute } from "vue-router";

const { t } = useI18n();
const message = useMessage();
const route = useRoute();
defineProps<{
  manage: boolean;
}>();
let newTag = $ref("");
let inputTag = $ref<HTMLInputElement>();
let addEnvTagOpen = $ref(true);

const addEnvtagDisplay = () => {
  newTag = "";
  watch($$(inputTag), () => {
    inputTag?.focus();
  });
};

const addEnvtag = () => {
  api
    .get("/env-tag-create", {
      queries: {
        env: route.params.id as string,
        name: newTag as string,
      },
    })
    .then((res) => {
      // if (res === "permission denied") {
      //   message.error(t("messages.permissionDenied"));
      //   return;
      // }
      if (res === "ok") {
        newTag = "";
        message.success(t("messages.tagAdded"));
        addEnvTagOpen = !addEnvTagOpen;
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
    :title="`${t('actions.addNewTag')}`"
    @positive-click="addEnvtag"
    @negative-click="() => {}"
    :show-footer="{
      positiveText: t('actions.confirm'),
      negativeText: t('actions.cancel'),
    }"
    :show="addEnvTagOpen"
    :touched="newTag.length > 0"
    :width="'20rem'"
  >
    <template #trigger>
      <n-button
        size="tiny"
        strong
        secondary
        type="primary"
        @click="newTag = ''"
        :disabled="!manage"
      >
        <n-icon class="icon" @click="addEnvtagDisplay" size="15px"
          ><mdi :path="mdiPlus"
        /></n-icon>
      </n-button>
    </template>
    <template #content>
      <div style="display: flex; flex-direction: column; gap: 0.5rem">
        <n-input
          v-model:value="newTag"
          ref="inputTag"
          @keyup.enter="addEnvtag"
          :placeholder="t('actions.addNewTag')"
        />
      </div>
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
  color: var(--primaryColor);
}
.alias-input {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 0.5rem;
  justify-items: start;
  align-items: center;
}
</style>
