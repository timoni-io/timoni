<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";

const { t } = useI18n();
const message = useMessage();

const props = defineProps<{
  email: string;
}>();
const emit = defineEmits<{
  (e: "refreshUsersList"): void;
}>();

let newAliasEmail = $ref(props.email);
let newAlias = $ref("");
let addAliasOpen = $ref(true);
const newAliasRef = $ref(null as HTMLInputElement | null);

const addAliasDisplay = () => {
  newAlias = "";
  newAliasRef?.focus();
};

const addAlias = () => {
  api
    .get("/user/alias-add", {
      queries: {
        email: newAliasEmail as string,
        alias: newAlias as string,
      },
    })
    .then((res) => {
      if (res === "ok") {
        message.success(t("messages.aliasAdded"));
        addAliasOpen = false;
        newAlias = "";
      } else message.info(res);
      emit("refreshUsersList");
    })
    .catch((error) => {
      message.error(error);
      emit("refreshUsersList");
    });
};
</script>

<template>
  <PopModal
    :title="`${t('actions.addNewAlias')}`"
    @positive-click="addAlias"
    @negative-click="() => {}"
    :show-footer="{
      positiveText: t('actions.confirm'),
      negativeText: t('actions.cancel'),
    }"
    :show="addAliasOpen"
    :touched="newAlias.length > 0"
    :width="'20rem'"
  >
    <template #trigger>
      <n-icon class="icon" color="#18a058" @click="addAliasDisplay"
        ><mdi :path="mdiPlus"
      /></n-icon>
    </template>
    <template #content>
      <div style="display: flex; flex-direction: column; gap: 0.5rem">
        <n-input
          v-model:value="newAlias"
          ref="newAliasRef"
          autofocus
          @keyup.enter="addAlias"
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
}
.alias-input {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 0.5rem;
  justify-items: start;
  align-items: center;
}
</style>
