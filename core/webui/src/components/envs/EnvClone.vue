<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { useRoute, useRouter } from "vue-router";

const { t } = useI18n();
const message = useMessage();
const route = useRoute();
const router = useRouter();

const props = defineProps<{
  name: string;
  manage: boolean;
}>();

let envNameError = $ref("");
let inputRef = $ref(null as HTMLInputElement | null);
let clonedEnv = $ref<string>(`${props.name}-clone`);
watch(
  () => props.name,
  () => {
    clonedEnv = `${props.name}-clone`;
  }
);
watch($$(inputRef), () => {
  inputRef?.focus();
});
let cloneEnvShow = $ref(true);

const changeNameDisplay = () => {
  clonedEnv = `${props.name}-clone`;
};

const cloneEnv = () => {
  api
    .get("/env-clone", {
      queries: {
        env: route.params.id as string,
        targetName: clonedEnv as string,
      },
    })
    .then((res) => {
      if (res.substring(0, 3) === "env") {
        message.info(t("messages.createdEnvironment"));
        router.push("/");
      } else {
        message.error(res);
      }
    });
};

const createEnvF = () => {
  if (clonedEnv.trim()) {
    if (clonedEnv.trim().length >= 31) {
      return;
    }
    cloneEnv();
  } else {
    envNameError = t("messages.nameRequired");
  }
};

watch(
  () => clonedEnv,
  () => {
    if (clonedEnv.trim().length >= 31) envNameError = t("messages.tooLongName");
    else envNameError = "";
  }
);
</script>

<template>
  <PopModal
    :title="`${t('actions.cloneEnv')}`"
    @positive-click="cloneEnv"
    @negative-click="() => {}"
    :show-footer="{
      positiveText: t('actions.confirm'),
      negativeText: t('actions.cancel'),
    }"
    :show="cloneEnvShow"
    :touched="clonedEnv !== name + '-clone'"
    :width="'20rem'"
  >
    <template #trigger>
      <n-button
        size="tiny"
        strong
        secondary
        type="primary"
        :disabled="!manage"
        @click="changeNameDisplay"
      >
        <template #icon>
          <n-icon>
            <Mdi :path="mdiCheckboxMultipleBlankOutline" />
          </n-icon>
        </template>
        {{ t("elements.options.clone") }}
      </n-button>
    </template>
    <template #content>
      <div style="display: flex; flex-direction: column; gap: 0.5rem">
        <div
          style="display: flex; justify-content: space-between"
          :style="envNameError ? 'padding-bottom: 0.65rem' : ''"
        >
          <span>{{ t("fields.name") }}</span>
          <Input
            style="width: 70%"
            :placeholder="t('fields.envName')"
            :errorMessage="envNameError"
            :focus="true"
            :removeWhiteSpace="true"
            :name="clonedEnv"
            @keyup.enter="createEnvF"
            @update:value="
                          (v: string) => {
                            clonedEnv = v;
                          }
                        "
          />
        </div>
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
