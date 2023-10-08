<script lang="ts" setup>
import { useRoute } from "vue-router";
import { useMessage } from "naive-ui";

const message = useMessage();
const { t } = useI18n();
const route = useRoute();
defineProps<{
  tags: any;
  manage: boolean;
}>();
const deleteTag = (name: string) => {
  api
    .get("/env-tag-delete", {
      queries: {
        env: route.params.id as string,
        name,
      },
    })
    .then((res) => {
      // if (res === "permission denied") {
      //   message.error(t("messages.permissionDenied"));
      //   return;
      // }
      if (res === "ok") {
        message.success(t("messages.tagRemoved"));
      } else {
        message.error(res);
      }
    });
};
</script>

<template>
  <div v-if="tags" style="display: flex; flex-flow: row wrap; gap: 0.5rem">
    <PopModal
      v-for="tag in tags"
      :key="tag"
      :title="`${t('actions.removeTag')}`"
      @positive-click="deleteTag(tag)"
      @negative-click="() => {}"
      :show-footer="{
        positiveText: t('actions.confirm'),
        negativeText: t('actions.cancel'),
      }"
      :width="'20rem'"
    >
      <template #trigger>
        <n-button
          size="tiny"
          strong
          secondary
          type="primary"
          :disabled="!manage"
          :manage="manage"
        >
          {{ tag }}
          <n-icon
            color="var(--errorColor)"
            style="margin-left: 5px"
            size="15px"
          >
            <mdi :path="mdiClose" />
          </n-icon>
        </n-button>
      </template>
      <template #content>
        {{ t("questions.sure") }}
      </template>
    </PopModal>
    <EnvAddTags :manage="manage" />
  </div>
  <div v-else>
    <EnvAddTags :manage="manage" />
  </div>
</template>
