<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { useRoute } from "vue-router";

const { t } = useI18n();
const message = useMessage();
const route = useRoute();

const props = defineProps<{
  name: string;
  manage: boolean;
}>();
let inputRef = $ref(null as HTMLInputElement | null);
let newName = $ref(props.name);
watch(
  () => props.name,
  () => {
    newName = props.name;
  }
);
watch($$(inputRef), () => {
  inputRef?.focus();
});
let changeNameOpen = $ref(true);

let envNameError = $ref("");

const changeNameDisplay = () => {
  newName = props.name;
  envNameError = "";
};

const changeName = () => {
  if (!newName.trim()) {
    envNameError = t("messages.nameRequired");
    return;
  }

  if (newName.trim().length >= 31) {
    return;
  }

  api
    .get("/env-rename", {
      queries: {
        env: route.params.id as string,
        newName: newName as string,
      },
    })
    .then((res) => {
      // if (res === "permission denied") {
      //   message.error(t("messages.permissionDenied"));
      //   return;
      // }
      if (res === "ok") {
        message.success("Environment name changed");
        changeNameOpen = !changeNameOpen;
      } else {
        message.error(res);
      }
    })
    .catch((error) => {
      message.error(error);
    });
};

watch(
  () => newName,
  () => {
    if (newName.trim().length >= 31) envNameError = t("messages.tooLongName");
    else envNameError = "";
  }
);
</script>

<template>
  <PopModal
    :title="`${t('actions.changeName')}`"
    @positive-click="changeName"
    @negative-click="() => {}"
    :show-footer="{
      positiveText: t('actions.confirm'),
      negativeText: t('actions.cancel'),
    }"
    :show="changeNameOpen"
    :touched="newName !== name"
    :width="'20rem'"
  >
    <template #trigger>
      <n-button
        size="tiny"
        strong
        secondary
        type="primary"
        @click="newName = name"
        :disabled="!manage"
      >
        {{ name }}
        <template #icon>
          <n-icon class="icon" @click="changeNameDisplay"
            ><mdi :path="mdiPencil"
          /></n-icon>
        </template>
      </n-button>
    </template>
    <template #content>
      <div
        style="display: flex; flex-direction: column; gap: 0.5rem"
        :style="envNameError ? 'padding-bottom: 0.65rem' : ''"
      >
        <Input
          :placeholder="t('fields.envName')"
          :errorMessage="envNameError"
          :focus="true"
          :removeWhiteSpace="true"
          :name="newName"
          @keyup.enter="changeName"
          @update:value="
                          (v: string) => {
                            newName = v;
                          }
                        "
        />
        <!-- <n-input
          v-model:value="newName"
          @keyup.enter="changeName"
          ref="inputRef"
        /> -->
      </div>
    </template>
  </PopModal>
</template>

<style scoped lang="scss">
.icon {
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  color: var(--n-color-target);

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
